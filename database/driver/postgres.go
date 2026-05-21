package driver

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
)

// PostgreSQL implements the Driver interface for PostgreSQL databases.
type PostgreSQL struct{}

// NewPostgreSQL creates a new PostgreSQL driver.
func NewPostgreSQL() *PostgreSQL {
	return &PostgreSQL{}
}

// Open opens a connection to the PostgreSQL database.
func (p *PostgreSQL) Open(dsn string) (*sql.DB, error) {
	return sql.Open("postgres", dsn)
}

// Close closes the PostgreSQL database connection.
func (p *PostgreSQL) Close(db *sql.DB) error {
	return db.Close()
}

// Ping checks if the PostgreSQL database connection is alive.
func (p *PostgreSQL) Ping(ctx context.Context, db *sql.DB) error {
	return db.PingContext(ctx)
}

// BeginTx starts a PostgreSQL transaction with the given options.
func (p *PostgreSQL) BeginTx(ctx context.Context, db *sql.DB, opts *sql.TxOptions) (*sql.Tx, error) {
	return db.BeginTx(ctx, opts)
}

// Placeholder returns PostgreSQL-style placeholders ($1, $2, $3).
func (p *PostgreSQL) Placeholder(n int) string {
	return postgresPlaceholder(n)
}

// Dialect returns the dialect name.
func (p *PostgreSQL) Dialect() string {
	return "postgres"
}
