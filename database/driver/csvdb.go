package driver

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

// CSVDB implements the Driver interface for CSV-directory-backed storage.
// It embeds *SQLite for all standard Driver methods and overrides Open to
// scan a directory of CSV files and populate an in-memory SQLite database.
//
// The directory path is passed as the DSN (from ConnectionConfig.Database).
// Each .csv file in the directory becomes a table named after the filename
// (without the .csv extension). The first row of each CSV is the header.
//
// The driver is stateless — all state lives in the in-memory SQLite database.
type CSVDB struct {
	*SQLite
}

// NewCSVDB creates a new CSVDB driver.
func NewCSVDB() *CSVDB {
	return &CSVDB{SQLite: NewSQLite()}
}

// Dialect returns "sqlite" so the query builder generates SQLite-compatible
// SQL and uses SQLite placeholders. The query builder's isSQLite() check
// returns true, and no ArraySource Model() hook fires (tables are already
// populated at Open time).
func (c *CSVDB) Dialect() string {
	return "sqlite"
}

// Open opens an in-memory SQLite database, scans the directory at dirPath
// for .csv files, and populates one table per file. The dirPath is the DSN,
// which comes from ConnectionConfig.Database via BuildDSN.
//
// If dirPath is empty or ":memory:", no directory is scanned and an empty
// in-memory SQLite database is returned (useful for testing).
func (c *CSVDB) Open(dirPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, fmt.Errorf("csvdb: failed to open in-memory SQLite: %w", err)
	}

	// If no directory path, return empty in-memory DB
	if dirPath == "" || dirPath == ":memory:" {
		return db, nil
	}

	// Verify directory exists
	info, err := os.Stat(dirPath)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("csvdb: cannot access directory %s: %w", dirPath, err)
	}
	if !info.IsDir() {
		db.Close()
		return nil, fmt.Errorf("csvdb: %s is not a directory", dirPath)
	}

	// Scan for .csv files and populate tables.
	// Table names are tracked case-insensitively to detect collisions on
	// case-insensitive filesystems (Windows, macOS default) where Users.csv
	// and users.csv would produce colliding SQLite table names.
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("csvdb: cannot read directory %s: %w", dirPath, err)
	}

	seenTables := make(map[string]string) // lower(tableName) → original filename

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(strings.ToLower(entry.Name()), ".csv") {
			continue
		}

		tableName := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		filePath := filepath.Join(dirPath, entry.Name())

		lowerName := strings.ToLower(tableName)
		if prevFile, exists := seenTables[lowerName]; exists {
			db.Close()
			return nil, fmt.Errorf("csvdb: table name collision: %s and %s produce the same table name (case-insensitive)", prevFile, entry.Name())
		}
		seenTables[lowerName] = entry.Name()

		if err := populateCSVFile(db, tableName, filePath); err != nil {
			db.Close()
			return nil, fmt.Errorf("csvdb: failed to populate table %s from %s: %w", tableName, filePath, err)
		}
	}

	return db, nil
}
