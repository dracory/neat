package driver

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"io/fs"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// MaxCSVRows limits the number of data rows that can be loaded from a single
// CSV file to prevent unbounded memory/CPU consumption. This mirrors the
// Array driver's MaxArrayRows limit.
const MaxCSVRows = 100000

// parseCSV reads a CSV file from os or fs.FS and returns column names and rows.
func parseCSV(filePath string) (columns []string, rows [][]string, err error) {
	return parseCSVFS(nil, filePath)
}

// parseCSVFS reads a CSV file from an optional fs.FS (or os.Open if sys is nil) and returns column names and rows.
func parseCSVFS(sys fs.FS, filePath string) (columns []string, rows [][]string, err error) {
	var f io.ReadCloser
	if sys != nil {
		file, err := sys.Open(filePath)
		if err != nil {
			return nil, nil, err
		}
		f = file
	} else {
		file, err := os.Open(filePath)
		if err != nil {
			return nil, nil, err
		}
		f = file
	}
	defer func() { _ = f.Close() }()

	reader := csv.NewReader(f)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		if err == io.EOF {
			return nil, nil, fmt.Errorf("CSV file is empty")
		}
		return nil, nil, fmt.Errorf("CSV parse error: %w", err)
	}

	if len(header) > 0 {
		header[0] = strings.TrimPrefix(header[0], "\xEF\xBB\xBF")
	}

	for {
		record, readErr := reader.Read()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, nil, fmt.Errorf("CSV parse error: %w", readErr)
		}
		rows = append(rows, record)
		if len(rows) > MaxCSVRows {
			return nil, nil, fmt.Errorf("CSV file exceeds the limit of %d data rows", MaxCSVRows)
		}
	}

	return header, rows, nil
}

// inferColumnTypes examines all rows and infers the SQLite type for each
// column. For each column, it tries int → float → bool → time → string,
// widening to a more general type if any value doesn't fit.
func inferColumnTypes(columns []string, rows [][]string) []string {
	types := make([]string, len(columns))
	// Start with empty type; the first non-empty value sets the initial type.
	// This avoids the optimistic "INTEGER" initial value incorrectly widening
	// to TEXT when the first value is a DATETIME.

	for _, row := range rows {
		for i, val := range row {
			if i >= len(types) {
				break
			}
			if val == "" {
				continue // skip empty strings, don't affect type
			}

			valType := inferValueType(val)
			if types[i] == "" {
				types[i] = valType
				continue
			}
			types[i] = widenType(types[i], valType)
		}
	}

	// Default to TEXT for columns with no non-empty values
	for i := range types {
		if types[i] == "" {
			types[i] = "TEXT"
		}
	}

	return types
}

// inferValueType tries to determine the Go-native type of a string value.
func inferValueType(val string) string {
	if _, err := strconv.ParseInt(val, 10, 64); err == nil {
		return "INTEGER"
	}
	// Reject Inf/NaN — strconv.ParseFloat accepts them, but they're almost
	// always the literal text "Inf"/"NaN" in a CSV, not a numeric value.
	if f, err := strconv.ParseFloat(val, 64); err == nil && !math.IsNaN(f) && !math.IsInf(f, 0) {
		return "REAL"
	}
	if val == "true" || val == "false" || val == "True" || val == "False" {
		return "INTEGER" // SQLite stores bools as 0/1
	}
	if _, err := time.Parse(time.RFC3339, val); err == nil {
		return "DATETIME"
	}
	// Try common date formats
	for _, layout := range []string{"2006-01-02", "2006-01-02 15:04:05", "01/02/2006"} {
		if _, err := time.Parse(layout, val); err == nil {
			return "DATETIME"
		}
	}
	return "TEXT"
}

// widenType returns the wider of two SQLite types.
// INTEGER → REAL → TEXT is the widening chain.
// DATETIME is compatible only with itself and TEXT; any mix of DATETIME
// with INTEGER or REAL widens to TEXT.
// BLOB mixed with any other type widens to BLOB, because BLOB affinity
// stores values as-is without conversion — TEXT affinity would corrupt
// binary []byte values that happen to be valid UTF-8.
func widenType(current, new string) string {
	if current == new {
		return current
	}
	// BLOB mixed with any other type → BLOB (preserves binary data integrity)
	if current == "BLOB" || new == "BLOB" {
		return "BLOB"
	}
	// INTEGER and REAL are compatible — widen to REAL
	if current == "INTEGER" && new == "REAL" {
		return "REAL"
	}
	if current == "REAL" && new == "INTEGER" {
		return "REAL"
	}
	// DATETIME mixed with INTEGER or REAL → TEXT (incompatible numeric/date)
	if current == "DATETIME" || new == "DATETIME" {
		// DATETIME + TEXT → TEXT, DATETIME + INTEGER/REAL → TEXT
		return "TEXT"
	}
	// Any other incompatible mix → TEXT
	return "TEXT"
}

// convertValue converts a string value to the appropriate Go type for
// the given SQLite column type, for insertion via database/sql.
func convertValue(val string, sqlType string) any {
	if val == "" {
		return nil
	}

	switch sqlType {
	case "INTEGER":
		if val == "true" || val == "True" {
			return int64(1)
		}
		if val == "false" || val == "False" {
			return int64(0)
		}
		if n, err := strconv.ParseInt(val, 10, 64); err == nil {
			return n
		}
		// Fallback if inference was wrong for this specific row
		return val
	case "REAL":
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
		return val
	case "DATETIME":
		if t, err := time.Parse(time.RFC3339, val); err == nil {
			return t
		}
		for _, layout := range []string{"2006-01-02", "2006-01-02 15:04:05", "01/02/2006"} {
			if t, err := time.Parse(layout, val); err == nil {
				return t
			}
		}
		return val
	default:
		return val
	}
}

// populateCSVFile reads a CSV file, infers types, creates a table, and
// inserts all rows. This is the core logic called by CSVDB.Open for each
// .csv file in the directory.
func populateCSVFile(db *sql.DB, tableName string, sys fs.FS, filePath string) error {
	columns, rows, err := parseCSVFS(sys, filePath)
	if err != nil {
		return err
	}

	// Validate table and column names (SQL injection prevention)
	if !isSimpleIdentifier(tableName) {
		return fmt.Errorf("invalid table name derived from filename: %s", tableName)
	}
	seenCols := make(map[string]bool, len(columns))
	for _, col := range columns {
		if !isSimpleIdentifier(col) {
			return fmt.Errorf("invalid column name in CSV header: %s", col)
		}
		if seenCols[col] {
			return fmt.Errorf("duplicate column name in CSV header: %s", col)
		}
		seenCols[col] = true
	}

	// Validate that no data row has more fields than the header.
	// Rows with fewer fields are allowed (missing fields become NULL),
	// but extra fields indicate a malformed CSV and would silently lose data.
	colCount := len(columns)
	for i, row := range rows {
		if len(row) > colCount {
			return fmt.Errorf("row %d has %d fields but header has %d columns", i+1, len(row), colCount)
		}
	}

	// Infer column types
	colTypes := inferColumnTypes(columns, rows)

	// Sort columns for deterministic ordering
	// (not strictly necessary, but consistent with the array driver)
	indices := make([]int, len(columns))
	for i := range indices {
		indices[i] = i
	}
	sort.Slice(indices, func(a, b int) bool {
		return columns[indices[a]] < columns[indices[b]]
	})

	// CREATE TABLE
	var colDefs []string
	for _, idx := range indices {
		colDefs = append(colDefs, fmt.Sprintf("\"%s\" %s", columns[idx], colTypes[idx]))
	}
	createSQL := fmt.Sprintf("CREATE TABLE \"%s\" (%s)", tableName, strings.Join(colDefs, ", "))
	if _, err := db.Exec(createSQL); err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	if len(rows) == 0 {
		return nil
	}

	// Wrap all inserts in a transaction for atomicity and performance
	// (single commit vs. N commits).
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }() // safe to call after Commit (no-op)

	// INSERT rows in batches (SQLite parameter limit is ~999)
	batchSize := 500 / colCount
	if batchSize == 0 {
		batchSize = 1
	}

	// Build column list (sorted)
	var colNames []string
	var placeholders []string
	for _, idx := range indices {
		colNames = append(colNames, fmt.Sprintf("\"%s\"", columns[idx]))
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
			for _, idx := range indices {
				if idx < len(row) {
					values = append(values, convertValue(row[idx], colTypes[idx]))
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
//
// This is the shared package-level implementation used by both the Array and
// CSVDB drivers.
func isSimpleIdentifier(s string) bool {
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
