package driver

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
)

// Array implements the Driver interface for array-backed storage using SQLite.
type Array struct {
	*SQLite
	populated sync.Map // map[string]bool, key is "dbPointer-tableName"
	mu        sync.Mutex
}

// NewArray creates a new Array driver.
func NewArray() *Array {
	return &Array{
		SQLite: NewSQLite(),
	}
}

// Dialect returns the dialect name.
func (a *Array) Dialect() string {
	return "array"
}

// Populate populates the database with rows from the given ArraySource.
func (a *Array) Populate(ctx context.Context, db *sql.DB, source contractsorm.ArraySource) error {
	tableName := source.TableName()
	if tableName == "" {
		return fmt.Errorf("table name cannot be empty")
	}

	// Check if already populated for this connection and table
	if a.isPopulated(db, tableName) {
		return nil
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// Double-check after acquiring lock
	if a.isPopulated(db, tableName) {
		return nil
	}

	rows, err := source.Rows()
	if err != nil {
		return fmt.Errorf("failed to get rows from source: %w", err)
	}

	var schema map[string]string
	if s, ok := source.(contractsorm.ArraySchema); ok {
		schema = s.Schema()
	}

	if schema == nil {
		schema = a.inferSchema(rows)
	}

	if len(schema) == 0 {
		return fmt.Errorf("could not infer schema for table %s and no schema provided", tableName)
	}

	// Create table
	if err := a.createTable(ctx, db, tableName, schema); err != nil {
		return fmt.Errorf("failed to create table %s: %w", tableName, err)
	}

	// Insert rows
	if len(rows) > 0 {
		if err := a.insertRows(ctx, db, tableName, schema, rows); err != nil {
			return fmt.Errorf("failed to insert rows into %s: %w", tableName, err)
		}
	}

	a.markPopulated(db, tableName)
	return nil
}

func (a *Array) isPopulated(db *sql.DB, tableName string) bool {
	key := fmt.Sprintf("%p-%s", db, tableName)
	_, ok := a.populated.Load(key)
	return ok
}

func (a *Array) markPopulated(db *sql.DB, tableName string) {
	key := fmt.Sprintf("%p-%s", db, tableName)
	a.populated.Store(key, true)
}

func (a *Array) inferSchema(rows []map[string]any) map[string]string {
	schema := make(map[string]string)
	if len(rows) == 0 {
		return schema
	}

	// Track found types for each column
	foundTypes := make(map[string]string)

	for _, row := range rows {
		for col, val := range row {
			if _, exists := foundTypes[col]; exists {
				continue
			}

			if val == nil {
				continue
			}

			foundTypes[col] = a.goTypeToSQLite(val)
		}
	}

	// For columns that are nil in all rows, default to TEXT
	for _, row := range rows {
		for col := range row {
			if _, ok := foundTypes[col]; !ok {
				foundTypes[col] = "TEXT"
			}
		}
	}

	return foundTypes
}

func (a *Array) goTypeToSQLite(val any) string {
	switch val.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return "INTEGER"
	case float32, float64:
		return "REAL"
	case bool:
		return "INTEGER" // SQLite uses 0/1 for boolean
	case time.Time, *time.Time:
		return "DATETIME"
	case string:
		return "TEXT"
	default:
		return "TEXT"
	}
}

func (a *Array) createTable(ctx context.Context, db *sql.DB, tableName string, schema map[string]string) error {
	var columns []string

	// Sort columns for deterministic SQL
	keys := make([]string, 0, len(schema))
	for k := range schema {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, col := range keys {
		sqlType := schema[col]
		// Convert our internal type names to SQLite types if they aren't already
		switch strings.ToLower(sqlType) {
		case "int", "integer":
			sqlType = "INTEGER"
		case "float", "real", "double":
			sqlType = "REAL"
		case "bool", "boolean":
			sqlType = "INTEGER"
		case "string", "text":
			sqlType = "TEXT"
		case "time", "datetime", "timestamp":
			sqlType = "DATETIME"
		}
		columns = append(columns, fmt.Sprintf("\"%s\" %s", col, sqlType))
	}

	sql := fmt.Sprintf("CREATE TABLE IF NOT EXISTS \"%s\" (%s)", tableName, strings.Join(columns, ", "))
	_, err := db.ExecContext(ctx, sql)
	return err
}

func (a *Array) insertRows(ctx context.Context, db *sql.DB, tableName string, schema map[string]string, rows []map[string]any) error {
	if len(rows) == 0 {
		return nil
	}

	// Sort columns for deterministic SQL
	cols := make([]string, 0, len(schema))
	for col := range schema {
		cols = append(cols, col)
	}
	sort.Strings(cols)

	colNames := make([]string, len(cols))
	placeholders := make([]string, len(cols))
	for i, col := range cols {
		colNames[i] = fmt.Sprintf("\"%s\"", col)
		placeholders[i] = "?"
	}

	sqlPrefix := fmt.Sprintf("INSERT INTO \"%s\" (%s) VALUES ", tableName, strings.Join(colNames, ", "))

	// SQLite has a limit on the number of variables (parameters) in a single statement.
	// Default is usually 999.
	batchSize := 500 / len(cols)
	if batchSize == 0 {
		batchSize = 1
	}

	for i := 0; i < len(rows); i += batchSize {
		end := i + batchSize
		if end > len(rows) {
			end = len(rows)
		}

		batch := rows[i:end]
		var values []any
		var placeholderGroups []string

		for _, row := range batch {
			placeholderGroups = append(placeholderGroups, "("+strings.Join(placeholders, ", ")+")")
			for _, col := range cols {
				val := row[col]
				// Convert bool to 0/1 for SQLite
				if b, ok := val.(bool); ok {
					if b {
						val = 1
					} else {
						val = 0
					}
				}
				values = append(values, val)
			}
		}

		sql := sqlPrefix + strings.Join(placeholderGroups, ", ")
		_, err := db.ExecContext(ctx, sql, values...)
		if err != nil {
			return err
		}
	}

	return nil
}
