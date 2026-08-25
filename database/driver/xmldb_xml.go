package driver

import (
	"database/sql"
	"fmt"
	"io/fs"
	"strings"

	"github.com/dracory/neat/support/xmlsource"
)

// populateXMLFile reads an XML file, infers schema, creates a table,
// and inserts all rows in a transaction. Validates table/column name safety,
// case-insensitive duplicates, and row limits.
func populateXMLFile(db *sql.DB, tableName string, sys fs.FS, filePath string) error {
	var rows []map[string]any
	var err error
	if sys != nil {
		rows, err = xmlsource.ParseXMLFSFile(sys, filePath)
	} else {
		rows, err = xmlsource.ParseXMLFile(filePath)
	}
	if err != nil {
		return err
	}

	// If empty array or no data, skip creating the table (no data = no table)
	if len(rows) == 0 {
		return nil
	}

	// Validate table name (SQL injection prevention)
	if !isSimpleIdentifier(tableName) {
		return fmt.Errorf("invalid table name derived from filename: %s", tableName)
	}

	// Infer columns and their types
	columns, colTypes := inferMapSchema(rows)

	// If no columns could be inferred (e.g. empty elements only), skip the table
	if len(columns) == 0 {
		return nil
	}

	// Validate columns for simple identifiers and case-insensitive duplicates
	seenColsLower := make(map[string]string) // lowerColName -> originalColName
	for _, col := range columns {
		if !isSimpleIdentifier(col) {
			return fmt.Errorf("invalid column name in XML keys: %s", col)
		}
		lowerCol := strings.ToLower(col)
		if prevCol, exists := seenColsLower[lowerCol]; exists {
			return fmt.Errorf("duplicate column name (case-insensitive): %s and %s", prevCol, col)
		}
		seenColsLower[lowerCol] = col
	}

	// Create CREATE TABLE SQL statement
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

	// Wrap all inserts in a transaction for atomicity and performance
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }() // safe to call after Commit (no-op)

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
	insertPrefix := fmt.Sprintf("INSERT INTO \"%s\" (%s) VALUES ",
		tableName, strings.Join(colNames, ", "))

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

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
