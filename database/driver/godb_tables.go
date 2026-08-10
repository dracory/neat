package driver

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/dracory/neat/support/arraysource"
)

// Tables is a map of table names to data slices.
// Each value can be []map[string]any, []SomeStruct, or any slice.
//
// Usage in config:
//
//	Tables: godb.Tables{
//	    "blogs":      blogs.Blogs,
//	    "categories": blogs.Categories,
//	}
type Tables map[string]any

// Table is an alternative config style that preserves declaration order.
// Useful if table creation order matters (e.g., for foreign key constraints).
//
// Usage in config:
//
//	Tables: []godb.Table{
//	    {Name: "blogs",      Data: blogs.Blogs},
//	    {Name: "categories", Data: blogs.Categories},
//	}
type Table struct {
	Name string
	Data any
}

// MaxGODBRows limits the number of rows that can be populated in a single table.
const MaxGODBRows = 100000

// normalizeTables normalizes the Tables config (Style A map or Style B slice)
// into a slice of Table structures.
func normalizeTables(tablesField any) ([]Table, error) {
	if tablesField == nil {
		return nil, nil
	}

	switch val := tablesField.(type) {
	case Tables:
		// Sort keys for deterministic map ordering (Style A)
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		result := make([]Table, len(keys))
		for i, k := range keys {
			result[i] = Table{Name: k, Data: val[k]}
		}
		return result, nil

	case []Table:
		return val, nil

	default:
		return nil, fmt.Errorf("godb: Tables config field must be driver.Tables or []driver.Table, got %T", tablesField)
	}
}

// populateGODBTable converts a data slice to rows, infers schema,
// creates the table, and inserts all rows in a transaction.
func populateGODBTable(db *sql.DB, tableName string, data any) error {
	// 1. Validate table name (SQL injection prevention)
	if !isSimpleIdentifier(tableName) {
		return fmt.Errorf("invalid table name: %s", tableName)
	}

	// 2. Convert data to []map[string]any
	rows, err := arraysource.ConvertSliceToRows(data)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil // skip empty dataset silently
	}

	// 3. Check rows limit
	if len(rows) > MaxGODBRows {
		return fmt.Errorf("table %s has %d rows, exceeding the limit of %d", tableName, len(rows), MaxGODBRows)
	}

	// 4. Infer schema from Go native types
	columns, colTypes := inferMapSchema(rows)
	if len(columns) == 0 {
		return nil // 0 columns = skip table silently
	}

	// 5. Validate column names (simple identifiers, no case-insensitive dupes)
	seenColsLower := make(map[string]string) // lowerColName -> originalColName
	for _, col := range columns {
		if !isSimpleIdentifier(col) {
			return fmt.Errorf("invalid column name in table %s: %s", tableName, col)
		}
		lowerCol := strings.ToLower(col)
		if prevCol, exists := seenColsLower[lowerCol]; exists {
			return fmt.Errorf("duplicate column name (case-insensitive) in table %s: %s and %s", tableName, prevCol, col)
		}
		seenColsLower[lowerCol] = col
	}

	// 6. CREATE TABLE with inferred schema
	var colDefs []string
	colTypeMap := make(map[string]string) // colName -> sqlType
	for i, col := range columns {
		colDefs = append(colDefs, fmt.Sprintf("\"%s\" %s", col, colTypes[i]))
		colTypeMap[col] = colTypes[i]
	}
	createSQL := fmt.Sprintf("CREATE TABLE \"%s\" (%s)", tableName, strings.Join(colDefs, ", "))
	if _, err := db.Exec(createSQL); err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	// 7. BEGIN TRANSACTION
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	colCount := len(columns)
	batchSize := 500 / colCount
	if batchSize == 0 {
		batchSize = 1
	}

	// Build insert query placeholders
	var colNames []string
	var placeholders []string
	for _, col := range columns {
		colNames = append(colNames, fmt.Sprintf("\"%s\"", col))
		placeholders = append(placeholders, "?")
	}
	insertPrefix := fmt.Sprintf("INSERT INTO \"%s\" (%s) VALUES ", tableName, strings.Join(colNames, ", "))

	// 8. INSERT rows in batches
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
			for _, col := range columns {
				val, exists := row[col]
				if exists {
					values = append(values, convertGoValue(val, colTypeMap[col]))
				} else {
					values = append(values, nil)
				}
			}
		}

		insertSQL := insertPrefix + strings.Join(placeholderGroups, ", ")
		if _, err := tx.Exec(insertSQL, values...); err != nil {
			return fmt.Errorf("failed to insert rows: %w", err)
		}
	}

	// 9. COMMIT
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
