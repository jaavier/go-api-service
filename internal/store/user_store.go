package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jaavier/go-api-service/internal/model"
)

// ErrNotFound is returned when a requested resource does not exist.
var ErrNotFound = errors.New("store: not found")

// UserRepository defines the behaviour required by the handler layer.
// Programming to this interface instead of *UserStore makes it trivial to
// substitute a fake or mock in unit tests.
type UserRepository interface {
	List(ctx context.Context) ([]*model.User, error)
	GetByID(ctx context.Context, id int64) (*model.User, error)
	Create(ctx context.Context, u *model.User) error
	DeleteByID(ctx context.Context, id int64) error
}

// UserStore is the production implementation of UserRepository backed by
// a *sql.DB connection pool.
type UserStore struct {
	db *sql.DB // unexported: callers must use the interface
}

// NewUserStore constructs a UserStore.
func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

// List fetches all users.
func (s *UserStore) List(ctx context.Context) ([]*model.User, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, name, email FROM users")
	if err != nil {
		return nil, fmt.Errorf("store: list users: %w", err)
	}
	defer rows.Close() // always close, even on Scan error

	var users []*model.User
	for rows.Next() {
		u := &model.User{}
		if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
			return nil, fmt.Errorf("store: scan user: %w", err)
		}
		users = append(users, u)
	}
	// rows.Err() reports any error that occurred during iteration.
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: iterate users: %w", err)
	}

	return users, nil
}

// GetByID returns a single user, or ErrNotFound if no row matches.
func (s *UserStore) GetByID(ctx context.Context, id int64) (*model.User, error) {
	u := &model.User{}
	err := s.db.QueryRowContext(
		ctx,
		"SELECT id, name, email FROM users WHERE id=$1",
		id,
	).Scan(&u.ID, &u.Name, &u.Email)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, ErrNotFound
	case err != nil:
		return nil, fmt.Errorf("store: get user %d: %w", id, err)
	}

	return u, nil
}

// Create inserts a new user.
func (s *UserStore) Create(ctx context.Context, u *model.User) error {
	_, err := s.db.ExecContext(
		ctx,
		"INSERT INTO users (name, email) VALUES ($1, $2)",
		u.Name, u.Email,
	)
	if err != nil {
		return fmt.Errorf("store: create user: %w", err)
	}
	return nil
}

// DeleteByID removes a user by primary key.
// Returns ErrNotFound when no row was deleted.
func (s *UserStore) DeleteByID(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, "DELETE FROM users WHERE id=$1", id)
	if err != nil {
		return fmt.Errorf("store: delete user %d: %w", id, err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("store: rows affected: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}

	return nil
}

// Compile-time check: UserStore must satisfy UserRepository.
var _ UserRepository = (*UserStore)(nil)
