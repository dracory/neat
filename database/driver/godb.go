package driver

import (
	"database/sql"
	"fmt"
	"strings"

	_ "modernc.org/sqlite"
)

// GODB implements the Driver interface for Go-data-backed storage.
// It embeds *SQLite for all standard Driver methods and overrides Open
// to populate an in-memory SQLite database from Go data slices passed
// via the config Tables field.
//
// The data is compiled into the binary at build time by the Go compiler.
// At runtime, the driver converts each data slice to rows and inserts
// them into SQLite. No file I/O, no parsing.
type GODB struct {
	*SQLite
	tables any
}

// NewGODB creates a new GODB driver.
func NewGODB() *GODB {
	return &GODB{SQLite: NewSQLite()}
}

// SetTables configures the data tables before Open is called.
// Called by the ORM during connection setup, reading from config.
// It accepts either Tables (map) or []Table (slice).
func (g *GODB) SetTables(tables any) {
	g.tables = tables
}

// Dialect returns "sqlite" so the query builder generates SQLite-compatible
// SQL and uses SQLite placeholders.
func (g *GODB) Dialect() string {
	return "sqlite"
}

// Open opens an in-memory SQLite database, normalizes the configured Tables,
// checks for case-insensitive table name collisions, and populates the database.
func (g *GODB) Open(_ string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, fmt.Errorf("godb: failed to open in-memory SQLite: %w", err)
	}

	// Normalize tables from config (Style A or Style B)
	normalized, err := normalizeTables(g.tables)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	seen := make(map[string]string) // lower(tableName) -> original tableName

	for _, tbl := range normalized {
		tableName := tbl.Name
		lowerName := strings.ToLower(tableName)

		if prev, exists := seen[lowerName]; exists {
			_ = db.Close()
			return nil, fmt.Errorf("godb: table name collision (case-insensitive): %s and %s produce the same table name", prev, tableName)
		}
		seen[lowerName] = tableName

		if err := populateGODBTable(db, tableName, tbl.Data); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("godb: failed to populate table %s: %w", tableName, err)
		}
	}

	return db, nil
}
