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

// ListPaginated fetches a single page of users ordered by id, together with
// the total number of users so callers can build pagination metadata.
//
// It validates the page (size > 0, offset >= 0) and fails fast with
// ErrInvalidPage before touching the database. It propagates the request
// context, defers rows.Close, checks rows.Err and wraps errors with %w so
// callers can inspect them with errors.Is/As.
func (s *UserStore) ListPaginated(ctx context.Context, page model.Page) (users []*model.User, total int, err error) {
	if page.Size < 1 || page.Offset() < 0 {
		return nil, 0, fmt.Errorf("%w: size=%d offset=%d", ErrInvalidPage, page.Size, page.Offset())
	}

	if err := s.Db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	rows, err := s.Db.QueryContext(ctx,
		"SELECT id, name, email FROM users ORDER BY id LIMIT $1 OFFSET $2",
		page.Size, page.Offset(),
	)
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
