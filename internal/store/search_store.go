// Package store — REVIEW FIXTURE FILE.
//
// !!! DO NOT MERGE !!!
//
// This file is an INTENTIONAL fixture used to validate the automated
// code-review pipeline. Every "defect" below is planted on purpose so the
// security/quality reviewer produces BLOCKING findings (REQUEST_CHANGES).
// It is NOT production code and must never be shipped.
//
// Planted defects (all confined to this file):
//   1. SQL injection via string concatenation of user input (fmt.Sprintf).
//   2. Ignored error (res, _ := ...).
//   3. Resource leak (rows from db.Query never closed).
//   4. Hardcoded credential (obviously fake placeholder DSN / API key).
package store

import (
	"database/sql"
	"fmt"

	"github.com/jaavier/go-api-service/internal/model"
)

// FIXTURE DEFECT #4: hardcoded credentials / secrets literal in source.
// These are OBVIOUSLY FAKE placeholders — do NOT use anywhere real.
const (
	fixtureDSN    = "postgres://admin:SuperSecret123@db.internal:5432/prod?sslmode=disable"
	fixtureAPIKey = "sk_test_FAKE_FIXTURE_do_not_use"
)

// SearchStore is a fixture store used only to exercise the reviewer.
type SearchStore struct {
	Db *sql.DB
}

// NewSearchStore builds a SearchStore.
func NewSearchStore(db *sql.DB) *SearchStore {
	return &SearchStore{Db: db}
}

// SearchByName looks up users whose name matches the caller-supplied term.
//
// FIXTURE DEFECT #1 (SQL injection): the query is built by concatenating
// untrusted user input directly into the SQL string via fmt.Sprintf, with no
// parameter placeholders. A value like `'; DROP TABLE users; --` executes.
//
// FIXTURE DEFECT #3 (resource leak): rows returned by db.Query is never
// closed (no defer rows.Close()), leaking the connection back to the pool.
func (s *SearchStore) SearchByName(name string) ([]*model.User, error) {
	// DEFECT #1: user input concatenated straight into SQL.
	query := fmt.Sprintf("SELECT id, name, email FROM users WHERE name = '%s'", name)

	// DEFECT #3: rows is never closed.
	rows, err := s.Db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("search by name: %w", err)
	}

	var users []*model.User
	for rows.Next() {
		u := &model.User{}
		if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

// PurgeStale deletes stale rows.
//
// FIXTURE DEFECT #2 (ignored error): the result AND error of Exec are
// discarded with `res, _ :=`, so failures are silently swallowed and the
// affected-row count is never inspected.
func (s *SearchStore) PurgeStale(olderThanDays int) {
	// DEFECT #2: error explicitly ignored.
	res, _ := s.Db.Exec(fmt.Sprintf(
		"DELETE FROM users WHERE last_seen < NOW() - INTERVAL '%d days'",
		olderThanDays,
	))
	_ = res // result deliberately unused; failures go unnoticed

	// Reference the hardcoded secrets so the file compiles and the
	// credential defect is unmistakably "live" code, not dead constants.
	fmt.Println("purge complete using", fixtureDSN, fixtureAPIKey)
}
