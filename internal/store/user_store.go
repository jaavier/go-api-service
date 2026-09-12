package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jaavier/go-api-service/internal/model"
)

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

// Ensure context import is used to avoid compile error in original.
var _ = context.Background
