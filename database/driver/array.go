package driver

import (
	"context"
	"database/sql"
	"encoding/json"
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
	locks     sync.Map // map[string]*sync.Mutex, key is "dbPointer-tableName"
	locksMu   sync.Mutex
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

// MaxArrayRows limits the number of rows that can be populated in a single
// Populate call to prevent unbounded memory/CPU consumption.
const MaxArrayRows = 100000

// Populate populates the database with rows from the given ArraySource.
func (a *Array) Populate(ctx context.Context, db *sql.DB, source contractsorm.ArraySource) error {
	tableName := source.TableName()
	if tableName == "" {
		return fmt.Errorf("table name cannot be empty")
	}

	// Validate table name to prevent SQL injection
	if !a.isSimpleIdentifier(tableName) {
		return fmt.Errorf("invalid table name: %s", tableName)
	}

	// Check if already populated for this connection and table
	if a.isPopulated(db, tableName) {
		return nil
	}

	// Get or create a per-table mutex
	lock := a.getTableMutex(db, tableName)
	lock.Lock()
	defer lock.Unlock()

	// Double-check after acquiring lock
	if a.isPopulated(db, tableName) {
		return nil
	}

	rows, err := source.Rows()
	if err != nil {
		return fmt.Errorf("failed to get rows from source: %w", err)
	}

	if len(rows) > MaxArrayRows {
		return fmt.Errorf("array source for table %s has %d rows, exceeding the limit of %d", tableName, len(rows), MaxArrayRows)
	}

	var schema map[string]string
	explicitSchema := false
	if s, ok := source.(contractsorm.ArraySchema); ok {
		schema = s.Schema()
		explicitSchema = true
	}

	if schema == nil {
		schema, err = a.inferSchema(rows)
		if err != nil {
			return err
		}
	}

	if len(schema) == 0 {
		return fmt.Errorf("could not infer schema for table %s and no schema provided", tableName)
	}

	// If explicit schema provided, validate that rows only contain declared keys
	if explicitSchema && len(rows) > 0 {
		for i, row := range rows {
			for key := range row {
				if _, ok := schema[key]; !ok {
					return fmt.Errorf("row %d contains key %q which is not in the explicit schema", i, key)
				}
			}
		}
	}

	// Get sorted column names to ensure deterministic ordering in CREATE and INSERT
	sortedCols := make([]string, 0, len(schema))
	for col := range schema {
		// Validate column name to prevent SQL injection
		if !a.isSimpleIdentifier(col) {
			return fmt.Errorf("invalid column name: %s", col)
		}
		sortedCols = append(sortedCols, col)
	}
	sort.Strings(sortedCols)

	// Check if the table already exists before creating it.
	// If it does, skip creation and insertion to avoid schema mismatch
	// (CREATE TABLE IF NOT EXISTS would silently succeed with a different schema).
	exists, err := a.tableExists(ctx, db, tableName)
	if err != nil {
		return fmt.Errorf("failed to check table existence for %s: %w", tableName, err)
	}
	if exists {
		a.markPopulated(db, tableName)
		return nil
	}

	// Create table
	if err := a.createTable(ctx, db, tableName, schema, sortedCols); err != nil {
		return fmt.Errorf("failed to create table %s: %w", tableName, err)
	}

	// Insert rows
	if len(rows) > 0 {
		if err := a.insertRows(ctx, db, tableName, sortedCols, rows); err != nil {
			return fmt.Errorf("failed to insert rows into %s: %w", tableName, err)
		}
	}

	a.markPopulated(db, tableName)
	return nil
}

func (a *Array) tableExists(ctx context.Context, db *sql.DB, tableName string) (bool, error) {
	var count int
	err := db.QueryRowContext(ctx,
		"SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?", tableName,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// isPopulated checks whether the given table has already been populated for
// the given *sql.DB connection.
//
// This method trusts the sync.Map without issuing a sqlite_master verification
// query on every call. The previous implementation queried sqlite_master each
// time to guard against *sql.DB pointer reuse (after Close + reopen), but this
// added a round-trip to every Model() call. Instead, we rely on Cleanup() being
// called when connections are closed (via Orm.Close() → Database.Close()) to
// remove stale entries. If a stale entry causes Populate to skip a missing
// table, the subsequent query will fail with a clear "no such table" error.
//
// Edge case: When a Database is created via NewFromSQLDB / BuildOrmFromDB, the
// caller retains ownership of the *sql.DB. If the caller closes it directly
// (without calling Database.Close()), Cleanup() is never invoked and stale
// entries remain in the sync.Map. A new *sql.DB allocated at the same memory
// address would inherit these stale entries and skip population. This is an
// acceptable trade-off for performance; the affected query will fail with a
// clear "no such table" error that is easy to diagnose.
func (a *Array) isPopulated(db *sql.DB, tableName string) bool {
	key := fmt.Sprintf("%p-%s", db, tableName)
	_, ok := a.populated.Load(key)
	return ok
}

func (a *Array) markPopulated(db *sql.DB, tableName string) {
	key := fmt.Sprintf("%p-%s", db, tableName)
	a.populated.Store(key, true)
}

func (a *Array) getTableMutex(db *sql.DB, tableName string) *sync.Mutex {
	key := fmt.Sprintf("%p-%s", db, tableName)
	if m, ok := a.locks.Load(key); ok {
		return m.(*sync.Mutex)
	}

	a.locksMu.Lock()
	defer a.locksMu.Unlock()

	// Double check
	if m, ok := a.locks.Load(key); ok {
		return m.(*sync.Mutex)
	}

	m := &sync.Mutex{}
	a.locks.Store(key, m)
	return m
}

func (a *Array) inferSchema(rows []map[string]any) (map[string]string, error) {
	schema := make(map[string]string)
	if len(rows) == 0 {
		return schema, nil
	}

	// First pass: find all columns across all rows
	for _, row := range rows {
		for col, val := range row {
			if val == nil {
				continue
			}

			currentType := schema[col]
			newType, err := a.goTypeToSQLite(val)
			if err != nil {
				return nil, err
			}

			if currentType == "" {
				schema[col] = newType
				continue
			}

			// Type widening logic
			if currentType != newType {
				if (currentType == "INTEGER" && newType == "REAL") || (currentType == "REAL" && newType == "INTEGER") {
					schema[col] = "REAL"
				} else {
					// Incompatible types, default to TEXT
					schema[col] = "TEXT"
				}
			}
		}
	}

	// For columns that are nil in all rows, default to TEXT
	for _, row := range rows {
		for col := range row {
			if _, ok := schema[col]; !ok {
				schema[col] = "TEXT"
			}
		}
	}

	return schema, nil
}

func (a *Array) goTypeToSQLite(val any) (string, error) {
	switch val.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return "INTEGER", nil
	case float32, float64:
		return "REAL", nil
	case bool:
		return "INTEGER", nil // SQLite uses 0/1 for boolean
	case time.Time, *time.Time:
		return "DATETIME", nil
	case string:
		return "TEXT", nil
	case []byte, json.RawMessage:
		return "BLOB", nil
	default:
		return "", fmt.Errorf("unsupported type %T in array source", val)
	}
}

func (a *Array) createTable(ctx context.Context, db *sql.DB, tableName string, schema map[string]string, sortedCols []string) error {
	var columns []string

	for _, col := range sortedCols {
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
		case "blob":
			sqlType = "BLOB"
		default:
			return fmt.Errorf("unsupported column type %q for column %q", sqlType, col)
		}
		columns = append(columns, fmt.Sprintf("\"%s\" %s", col, sqlType))
	}

	sql := fmt.Sprintf("CREATE TABLE IF NOT EXISTS \"%s\" (%s)", tableName, strings.Join(columns, ", "))
	_, err := db.ExecContext(ctx, sql)
	return err
}

func (a *Array) insertRows(ctx context.Context, db *sql.DB, tableName string, sortedCols []string, rows []map[string]any) error {
	if len(rows) == 0 {
		return nil
	}

	colNames := make([]string, len(sortedCols))
	placeholders := make([]string, len(sortedCols))
	for i, col := range sortedCols {
		colNames[i] = fmt.Sprintf("\"%s\"", col)
		placeholders[i] = "?"
	}

	sqlPrefix := fmt.Sprintf("INSERT INTO \"%s\" (%s) VALUES ", tableName, strings.Join(colNames, ", "))

	// SQLite has a limit on the number of variables (parameters) in a single statement.
	// Default is usually 999.
	batchSize := 500 / len(sortedCols)
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
			for _, col := range sortedCols {
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

// Cleanup removes all cached entries for the given *sql.DB from the populated
// and locks maps. This should be called when the connection is closed to prevent
// unbounded memory growth in long-running services.
func (a *Array) Cleanup(db *sql.DB) {
	prefix := fmt.Sprintf("%p-", db)
	a.populated.Range(func(key, _ any) bool {
		if k, ok := key.(string); ok && strings.HasPrefix(k, prefix) {
			a.populated.Delete(k)
		}
		return true
	})
	a.locks.Range(func(key, _ any) bool {
		if k, ok := key.(string); ok && strings.HasPrefix(k, prefix) {
			a.locks.Delete(k)
		}
		return true
	})
}

// isSimpleIdentifier checks if a string is a simple SQL identifier (table or
// column name) that can be safely embedded in a double-quoted SQL identifier.
//
// All identifiers are emitted with double quotes in the generated SQL (e.g.
// ""column_name""), which makes any valid identifier safe regardless of whether
// it coincides with a SQL keyword. Therefore, we do NOT reject SQL keywords
// here — doing so would prevent legitimate column names like "order", "group",
// "index", "level", "date", "key", "type", "user", etc.
//
// The check still rejects characters that could break out of the double-quote
// context: dots (table.column), parentheses (function calls), non-identifier
// characters, and identifiers starting with a digit.
func (a *Array) isSimpleIdentifier(s string) bool {
	if s == "" {
		return false
	}

	// Check for dots (table.column) or parentheses (function calls)
	if strings.Contains(s, ".") || strings.Contains(s, "(") || strings.Contains(s, ")") {
		return false
	}

	// Check if starts with a number
	if s[0] >= '0' && s[0] <= '9' {
		return false
	}

	// Check if contains only valid identifier characters
	for _, r := range s {
		isLetter := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
		isDigit := r >= '0' && r <= '9'
		isUnderscore := r == '_'
		if !isLetter && !isDigit && !isUnderscore {
			return false
		}
	}

	return true
}
