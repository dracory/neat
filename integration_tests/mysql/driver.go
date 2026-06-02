package mysql_test

import (
	"context"
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

// MySQL implements the Driver interface for MySQL databases.
type MySQL struct{}

// NewMySQL creates a new MySQL driver.
func NewMySQL() *MySQL {
	return &MySQL{}
}

// Open opens a connection to the MySQL database.
func (m *MySQL) Open(dsn string) (*sql.DB, error) {
	return sql.Open("mysql", dsn)
}

// Close closes the MySQL database connection.
func (m *MySQL) Close(db *sql.DB) error {
	return db.Close()
}

// Ping checks if the MySQL database connection is alive.
func (m *MySQL) Ping(ctx context.Context, db *sql.DB) error {
	return db.PingContext(ctx)
}

// BeginTx starts a MySQL transaction with the given options.
func (m *MySQL) BeginTx(ctx context.Context, db *sql.DB, opts *sql.TxOptions) (*sql.Tx, error) {
	return db.BeginTx(ctx, opts)
}

// Placeholder returns MySQL-style placeholders (?, ?, ?).
func (m *MySQL) Placeholder(n int) string {
	return mysqlPlaceholder(n)
}

// Dialect returns the dialect name.
func (m *MySQL) Dialect() string {
	return "mysql"
}

// mysqlPlaceholder returns MySQL-style placeholders (?, ?, ?).
func mysqlPlaceholder(n int) string {
	return "?"
}
