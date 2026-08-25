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

// XMLDB implements the Driver interface for XML-directory-backed storage.
// It embeds *SQLite for all standard Driver methods and overrides Open to
// scan a directory of XML files and populate an in-memory SQLite database.
//
// The directory path is passed as the DSN (from ConnectionConfig.Database).
// Each .xml file in the directory becomes a table named after the filename
// (without the .xml extension). The root element is the container; each
// direct child element is a row. Attributes and leaf sub-elements define
// the columns.
//
// The driver is stateless — all state lives in the in-memory SQLite database.
type XMLDB struct {
	*SQLite
	fs fs.FS
}

// NewXMLDB creates a new XMLDB driver.
func NewXMLDB() *XMLDB {
	return &XMLDB{SQLite: NewSQLite()}
}

// SetFS configures an embedded filesystem (embed.FS / fs.FS) before Open is called.
func (x *XMLDB) SetFS(sys fs.FS) {
	x.fs = sys
}

// Dialect returns "sqlite" so the query builder generates SQLite-compatible
// SQL and uses SQLite placeholders.
func (x *XMLDB) Dialect() string {
	return "sqlite"
}

// Open opens an in-memory SQLite database, scans the directory at dirPath
// for .xml files, and populates one table per file.
// The dirPath is the DSN, which comes from ConnectionConfig.Database via BuildDSN.
//
// If dirPath is empty or ":memory:", no directory is scanned and an empty
// in-memory SQLite database is returned (useful for testing).
func (x *XMLDB) Open(dirPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, fmt.Errorf("xmldb: failed to open in-memory SQLite: %w", err)
	}

	var entries []fs.DirEntry
	var cleanPath string

	if x.fs != nil {
		if dirPath == "" || dirPath == ":memory:" {
			dirPath = "."
		}
		cleanPath = path.Clean(dirPath)
		if cleanPath == "/" || cleanPath == "." {
			cleanPath = "."
		} else {
			cleanPath = strings.TrimPrefix(cleanPath, "/")
		}

		info, err := fs.Stat(x.fs, cleanPath)
		if err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("xmldb: cannot access directory %s: %w", dirPath, err)
		}
		if !info.IsDir() {
			_ = db.Close()
			return nil, fmt.Errorf("xmldb: %s is not a directory", dirPath)
		}

		entries, err = fs.ReadDir(x.fs, cleanPath)
		if err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("xmldb: cannot read directory %s: %w", dirPath, err)
		}
	} else {
		if dirPath == "" || dirPath == ":memory:" {
			return db, nil
		}

		info, err := os.Stat(dirPath)
		if err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("xmldb: cannot access directory %s: %w", dirPath, err)
		}
		if !info.IsDir() {
			_ = db.Close()
			return nil, fmt.Errorf("xmldb: %s is not a directory", dirPath)
		}

		entries, err = os.ReadDir(dirPath)
		if err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("xmldb: cannot read directory %s: %w", dirPath, err)
		}
	}

	seenTables := make(map[string]string) // lower(tableName) → original filename

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		var rawExt string
		if x.fs != nil {
			rawExt = path.Ext(entry.Name())
		} else {
			rawExt = filepath.Ext(entry.Name())
		}

		extLower := strings.ToLower(rawExt)
		if extLower != ".xml" {
			continue
		}

		var filePath string
		if x.fs != nil {
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
			return nil, fmt.Errorf("xmldb: table name collision: %s and %s produce the same table name (case-insensitive)", prevFile, entry.Name())
		}
		seenTables[lowerName] = entry.Name()

		if err := populateXMLFile(db, tableName, x.fs, filePath); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("xmldb: failed to populate table %s from %s: %w", tableName, filePath, err)
		}
	}

	return db, nil
}
