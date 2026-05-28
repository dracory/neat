package driver

import (
	"context"
	"database/sql"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

// Turso implements the Driver interface for Turso databases.
type Turso struct{}

// NewTurso creates a new Turso driver.
func NewTurso() *Turso {
	return &Turso{}
}

// Open opens a connection to the Turso database.
func (t *Turso) Open(dsn string) (*sql.DB, error) {
	return sql.Open("libsql", dsn)
}

// Close closes the Turso database connection.
func (t *Turso) Close(db *sql.DB) error {
	return db.Close()
}

// Ping checks if the Turso database connection is alive.
func (t *Turso) Ping(ctx context.Context, db *sql.DB) error {
	return db.PingContext(ctx)
}

// BeginTx starts a Turso transaction with the given options.
func (t *Turso) BeginTx(ctx context.Context, db *sql.DB, opts *sql.TxOptions) (*sql.Tx, error) {
	return db.BeginTx(ctx, opts)
}

// Placeholder returns SQLite-style placeholders (?) since Turso uses SQLite.
func (t *Turso) Placeholder(n int) string {
	return sqlitePlaceholder(n)
}

// Dialect returns the dialect name.
func (t *Turso) Dialect() string {
	return "turso"
}
