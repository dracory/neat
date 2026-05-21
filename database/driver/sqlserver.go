package driver

import (
	"context"
	"database/sql"

	_ "github.com/microsoft/go-mssqldb"
)

// SQLServer implements the Driver interface for SQL Server databases.
type SQLServer struct{}

// NewSQLServer creates a new SQL Server driver.
func NewSQLServer() *SQLServer {
	return &SQLServer{}
}

// Open opens a connection to the SQL Server database.
func (s *SQLServer) Open(dsn string) (*sql.DB, error) {
	return sql.Open("sqlserver", dsn)
}

// Close closes the SQL Server database connection.
func (s *SQLServer) Close(db *sql.DB) error {
	return db.Close()
}

// Ping checks if the SQL Server database connection is alive.
func (s *SQLServer) Ping(ctx context.Context, db *sql.DB) error {
	return db.PingContext(ctx)
}

// BeginTx starts a SQL Server transaction with the given options.
func (s *SQLServer) BeginTx(ctx context.Context, db *sql.DB, opts *sql.TxOptions) (*sql.Tx, error) {
	return db.BeginTx(ctx, opts)
}

// Placeholder returns SQL Server-style placeholders (@p1, @p2, @p3).
func (s *SQLServer) Placeholder(n int) string {
	return sqlserverPlaceholder(n)
}

// Dialect returns the dialect name.
func (s *SQLServer) Dialect() string {
	return "sqlserver"
}
