package driver

import (
	"context"
	"database/sql"

	_ "modernc.org/sqlite"
)

// SQLite implements the Driver interface for SQLite databases.
type SQLite struct{}

// NewSQLite creates a new SQLite driver.
func NewSQLite() *SQLite {
	return &SQLite{}
}

// Open opens a connection to the SQLite database.
func (s *SQLite) Open(dsn string) (*sql.DB, error) {
	return sql.Open("sqlite", dsn)
}

// Close closes the SQLite database connection.
func (s *SQLite) Close(db *sql.DB) error {
	return db.Close()
}

// Ping checks if the SQLite database connection is alive.
func (s *SQLite) Ping(ctx context.Context, db *sql.DB) error {
	return db.PingContext(ctx)
}

// BeginTx starts a SQLite transaction with the given options.
func (s *SQLite) BeginTx(ctx context.Context, db *sql.DB, opts *sql.TxOptions) (*sql.Tx, error) {
	return db.BeginTx(ctx, opts)
}

// Placeholder returns SQLite-style placeholders (?).
func (s *SQLite) Placeholder(n int) string {
	return sqlitePlaceholder(n)
}

// Dialect returns the dialect name.
func (s *SQLite) Dialect() string {
	return "sqlite"
}
