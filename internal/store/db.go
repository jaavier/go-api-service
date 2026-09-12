package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

const (
	dbMaxOpenConns    = 25
	dbMaxIdleConns    = 5
	dbConnMaxLifetime = 5 * time.Minute
)

// NewDB opens and validates a Postgres connection pool.
//
// The DSN must be a valid libpq connection string, e.g.:
//
//	postgres://user:pass@localhost/dbname?sslmode=disable
//
// A Ping is performed to detect misconfigurations early.
func NewDB(ctx context.Context, dsn string) (*sql.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("store: dsn must not be empty")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("store: open db: %w", err)
	}

	db.SetMaxOpenConns(dbMaxOpenConns)
	db.SetMaxIdleConns(dbMaxIdleConns)
	db.SetConnMaxLifetime(dbConnMaxLifetime)

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("store: ping db: %w", err)
	}

	return db, nil
}
