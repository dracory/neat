package driver

import (
	"context"
	"database/sql"

	_ "github.com/dracory/aztablessql"
)

// Aztables implements the Driver interface for Azure Table Storage via the
// aztablessql database/sql driver (github.com/dracory/aztablessql).
//
// The DSN is an Azure Storage connection string, e.g.
//
//	"DefaultEndpointsProtocol=https;AccountName=...;AccountKey=...;EndpointSuffix=core.windows.net"
//
// For local development with Azurite, use the well-known dev-store
// connection string pointing at the Azurite Table endpoint.
//
// Table Storage has no cross-partition transactions; BeginTx will fail
// with the aztablessql driver's "transactions not supported" error.
// Use the aztablessql batch API (Conn.Raw) for atomic multi-entity writes
// within a single partition.
type Aztables struct{}

// NewAztables creates a new Aztables driver.
func NewAztables() *Aztables {
	return &Aztables{}
}

// Open opens a connection to Azure Table Storage via the aztablessql driver.
func (a *Aztables) Open(dsn string) (*sql.DB, error) {
	return sql.Open("aztables", dsn)
}

// Close closes the Azure Table Storage connection.
func (a *Aztables) Close(db *sql.DB) error {
	return db.Close()
}

// Ping checks if the Azure Table Storage connection is alive.
func (a *Aztables) Ping(ctx context.Context, db *sql.DB) error {
	return db.PingContext(ctx)
}

// BeginTx starts a transaction. Azure Table Storage does not support
// cross-partition transactions, so this returns the aztablessql driver's
// "transactions not supported" error. Use the batch API for atomic
// single-partition writes.
func (a *Aztables) BeginTx(ctx context.Context, db *sql.DB, opts *sql.TxOptions) (*sql.Tx, error) {
	return db.BeginTx(ctx, opts)
}

// Placeholder returns Azure Table Storage-style placeholders (?).
// The aztablessql driver uses ? placeholders, same as SQLite/MySQL.
func (a *Aztables) Placeholder(n int) string {
	return sqlitePlaceholder(n)
}

// Dialect returns the dialect name.
func (a *Aztables) Dialect() string {
	return "aztables"
}
