package driver

import (
	"context"
	"database/sql"
)

// Driver defines the interface for database driver implementations.
type Driver interface {
	// Open opens a connection to the database.
	Open(dsn string) (*sql.DB, error)
	// Close closes the database connection.
	Close(db *sql.DB) error
	// Ping checks if the database connection is alive.
	Ping(ctx context.Context, db *sql.DB) error
	// BeginTx starts a transaction with the given options.
	BeginTx(ctx context.Context, db *sql.DB, opts *sql.TxOptions) (*sql.Tx, error)
	// Placeholder returns the placeholder format for the driver.
	Placeholder(n int) string
	// Dialect returns the dialect name (mysql, postgres, sqlite, sqlserver, turso).
	Dialect() string
}
