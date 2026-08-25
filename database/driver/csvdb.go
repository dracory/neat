package driver

import (
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path"
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
	fs fs.FS
}

// NewCSVDB creates a new CSVDB driver.
func NewCSVDB() *CSVDB {
	return &CSVDB{SQLite: NewSQLite()}
}

// SetFS configures an embedded filesystem (embed.FS / fs.FS) before Open is called.
func (c *CSVDB) SetFS(sys fs.FS) {
	c.fs = sys
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

	var entries []fs.DirEntry
	var cleanPath string

	if c.fs != nil {
		if dirPath == "" || dirPath == ":memory:" {
			dirPath = "."
		}
		cleanPath = path.Clean(dirPath)
		if cleanPath == "/" || cleanPath == "." {
			cleanPath = "."
		} else {
			cleanPath = strings.TrimPrefix(cleanPath, "/")
		}

		info, err := fs.Stat(c.fs, cleanPath)
		if err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("csvdb: cannot access directory %s: %w", dirPath, err)
		}
		if !info.IsDir() {
			_ = db.Close()
			return nil, fmt.Errorf("csvdb: %s is not a directory", dirPath)
		}

		entries, err = fs.ReadDir(c.fs, cleanPath)
		if err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("csvdb: cannot read directory %s: %w", dirPath, err)
		}
	} else {
		if dirPath == "" || dirPath == ":memory:" {
			return db, nil
		}

		info, err := os.Stat(dirPath)
		if err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("csvdb: cannot access directory %s: %w", dirPath, err)
		}
		if !info.IsDir() {
			_ = db.Close()
			return nil, fmt.Errorf("csvdb: %s is not a directory", dirPath)
		}

		entries, err = os.ReadDir(dirPath)
		if err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("csvdb: cannot read directory %s: %w", dirPath, err)
		}
	}

	seenTables := make(map[string]string) // lower(tableName) → original filename

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		var rawExt string
		if c.fs != nil {
			rawExt = path.Ext(entry.Name())
		} else {
			rawExt = filepath.Ext(entry.Name())
		}

		if strings.ToLower(rawExt) != ".csv" {
			continue
		}

		var filePath string
		if c.fs != nil {
			if cleanPath == "." {
				filePath = entry.Name()
			} else {
				filePath = path.Join(cleanPath, entry.Name())
			}
		} else {
			filePath = filepath.Join(dirPath, entry.Name())
		}
		tableName := strings.TrimSuffix(entry.Name(), rawExt)

		lowerName := strings.ToLower(tableName)
		if prevFile, exists := seenTables[lowerName]; exists {
			_ = db.Close()
			return nil, fmt.Errorf("csvdb: table name collision: %s and %s produce the same table name (case-insensitive)", prevFile, entry.Name())
		}
		seenTables[lowerName] = entry.Name()

		if err := populateCSVFile(db, tableName, c.fs, filePath); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("csvdb: failed to populate table %s from %s: %w", tableName, filePath, err)
		}
	}

	return db, nil
}
