package driver

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/sijms/go-ora/v2"
)

// Oracle implements the Driver interface for Oracle databases.
type Oracle struct{}

// NewOracle creates a new Oracle driver.
func NewOracle() *Oracle {
	return &Oracle{}
}

// Open opens a connection to the Oracle database.
func (o *Oracle) Open(dsn string) (*sql.DB, error) {
	return sql.Open("oracle", dsn)
}

// Close closes the Oracle database connection.
func (o *Oracle) Close(db *sql.DB) error {
	return db.Close()
}

// Ping checks if the Oracle database connection is alive.
func (o *Oracle) Ping(ctx context.Context, db *sql.DB) error {
	return db.PingContext(ctx)
}

// BeginTx starts an Oracle transaction with the given options.
func (o *Oracle) BeginTx(ctx context.Context, db *sql.DB, opts *sql.TxOptions) (*sql.Tx, error) {
	return db.BeginTx(ctx, opts)
}

// Placeholder returns Oracle-style placeholders (:1, :2, :3).
func (o *Oracle) Placeholder(n int) string {
	return oraclePlaceholder(n)
}

// Dialect returns the dialect name.
func (o *Oracle) Dialect() string {
	return "oracle"
}

// oraclePlaceholder returns Oracle-style placeholders (:1, :2, :3).
func oraclePlaceholder(n int) string {
	return fmt.Sprintf(":%d", n)
}
