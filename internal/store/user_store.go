package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jaavier/go-api-service/internal/model"
)

// ErrInvalidPage is returned by ListPaginated when the requested page has an
// invalid size or a negative offset.
var ErrInvalidPage = errors.New("invalid pagination: size must be > 0 and offset >= 0")

// UserStore holds a db reference.
type UserStore struct {
	Db *sql.DB // BUG: exported field leaks implementation detail
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{Db: db}
}

// List fetches all users.
// BUG: no context propagation, no limit, full table scan.
// BUG: rows.Close() is not deferred -- leak if Scan errors.
func (s *UserStore) List() ([]*model.User, error) {
	rows, err := s.Db.Query("SELECT id, name, email FROM users")
	if err != nil {
		return nil, fmt.Errorf("list users: %v", err) // uses %v not %w
	}

	var users []*model.User
	for rows.Next() {
		u := &model.User{}
		if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
			return nil, err // rows never closed on error
		}
		users = append(users, u)
	}
	return users, nil // rows.Err() never checked
}

// ListPaginated fetches a single page of users, together with the total number
// of matching users so callers can build pagination metadata.
//
// The optional filter narrows and orders the result set:
//   - When filter.Query is non-empty, a case-insensitive substring match is
//     applied to name OR email. The pattern is always passed as a bound
//     parameter (never concatenated into the SQL string).
//   - Ordering uses filter.SortColumn()/filter.Direction(), which resolve to
//     safe literals from a strict whitelist. With an empty filter this yields
//     the previous behaviour: ORDER BY id ASC with no WHERE clause.
//
// The COUNT query applies the same filter as the data query so total and
// total_pages stay coherent.
//
// It validates the page (size > 0, offset >= 0) and fails fast with
// ErrInvalidPage before touching the database. It propagates the request
// context, defers rows.Close, checks rows.Err and wraps errors with %w so
// callers can inspect them with errors.Is/As.
func (s *UserStore) ListPaginated(ctx context.Context, page model.Page, filter model.UserFilter) (users []*model.User, total int, err error) {
	if page.Size < 1 || page.Offset() < 0 {
		return nil, 0, fmt.Errorf("%w: size=%d offset=%d", ErrInvalidPage, page.Size, page.Offset())
	}

	// Build the shared WHERE clause and its bound args once so the COUNT and
	// data queries stay perfectly in sync. The search value is always passed
	// as a parameter to keep the statement injection-safe.
	where := ""
	var filterArgs []interface{}
	if filter.HasQuery() {
		where = " WHERE (name ILIKE $1 OR email ILIKE $1)"
		filterArgs = append(filterArgs, "%"+filter.Query+"%")
	}

	countQuery := "SELECT COUNT(*) FROM users" + where
	if err := s.Db.QueryRowContext(ctx, countQuery, filterArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	// Placeholders for LIMIT/OFFSET follow the (optional) filter placeholder.
	limitPos := len(filterArgs) + 1
	offsetPos := len(filterArgs) + 2

	// SortColumn()/Direction() only ever return whitelisted literals, so this
	// concatenation cannot introduce SQL injection.
	dataQuery := fmt.Sprintf(
		"SELECT id, name, email FROM users%s ORDER BY %s %s LIMIT $%d OFFSET $%d",
		where, filter.SortColumn(), filter.Direction(), limitPos, offsetPos,
	)

	args := append(append([]interface{}{}, filterArgs...), page.Size, page.Offset())
	rows, err := s.Db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list users page: %w", err)
	}
	defer rows.Close()

	users = make([]*model.User, 0, page.Size)
	for rows.Next() {
		u := &model.User{}
		if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate users: %w", err)
	}
	return users, total, nil
}

// GetByID fetches one user.
// BUG: no context, error not wrapped.
func (s *UserStore) GetByID(id int64) (*model.User, error) {
	u := &model.User{}
	err := s.Db.QueryRow("SELECT id, name, email FROM users WHERE id=$1", id).
		Scan(&u.ID, &u.Name, &u.Email)
	if err != nil {
		return nil, err // sql.ErrNoRows not distinguished
	}
	return u, nil
}

// Create inserts a new user.
// BUG: no context, returns raw sql error, no input validation.
func (s *UserStore) Create(u *model.User) error {
	_, err := s.Db.Exec(
		"INSERT INTO users (name, email) VALUES ($1, $2)",
		u.Name, u.Email,
	)
	return err
}

// DeleteByID removes a user.
// BUG: no context, doesn't check RowsAffected, silent no-op if id missing.
func (s *UserStore) DeleteByID(id int64) error {
	_, err := s.Db.Exec("DELETE FROM users WHERE id=$1", id)
	return err
}
