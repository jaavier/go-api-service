package store

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// NewDB opens a Postgres connection.
// BUG: no SetMaxOpenConns / SetMaxIdleConns / SetConnMaxLifetime configured.
// BUG: dsn is never validated before use.
func NewDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err // error not wrapped
	}
	return db, nil
}

// execQuery runs a query and silently drops the error.
// BUG: returned error is discarded.
func execQuery(db *sql.DB, query string) {
	db.Exec(query) // error ignored
	fmt.Println("query executed")
}
