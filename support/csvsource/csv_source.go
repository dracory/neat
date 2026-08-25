package csvsource

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dracory/neat/support/arraysource"
)

// NewCsvSource parses a CSV string and returns an array-backed data source
// ready for querying with the array driver.
//
// The first line must be a header row defining column names. Remaining
// lines are parsed with type inference: each value is tried as int, then
// float, then bool, then time (RFC3339 and common date formats), falling
// back to string. Type widening applies across rows — a column with mixed
// int and float values becomes float.
//
// The table name must be provided explicitly since there is no filename
// to derive it from. Override with the .Table() method on the returned
// model if needed.
//
//	database.Query().
//	    Model(neat.NewCsvSource(csvString, "users")).
//	    Where("active = ?", true).
//	    Get(&users)
//
// Panics if the CSV string is empty or has no header row — these are
// programmer errors, not runtime conditions.
func NewCsvSource(csvString string, tableName string) *arraysource.Model {
	columns, rows, err := parseCSVString(csvString)
	if err != nil {
		panic(fmt.Sprintf("csvsource: failed to parse CSV string: %v", err))
	}

	mapRows := convertRows(columns, rows)

	if len(mapRows) == 0 {
		// Header-only CSV: use NewWithSchema so the table is created with
		// the right columns even though there are no data rows.
		schema := inferSchema(columns, rows)
		return arraysource.NewWithSchema(mapRows, schema).Table(tableName)
	}

	return arraysource.New(mapRows).Table(tableName)
}

// NewCsvFSSource reads a CSV file from an embedded filesystem (embed.FS / fs.FS),
// infers column types, and returns an array-backed data source ready for querying.
func NewCsvFSSource(sys fs.FS, filePath string) *arraysource.Model {
	return NewCsvFSSourceWithDelimiter(sys, filePath, ',')
}

// NewCsvFSSourceWithDelimiter reads a CSV file from an embedded filesystem (embed.FS / fs.FS)
// with a custom field delimiter and returns an array-backed data source.
func NewCsvFSSourceWithDelimiter(sys fs.FS, filePath string, delimiter rune) *arraysource.Model {
	columns, rows, err := parseCSVFSFileWithDelimiter(sys, filePath, delimiter)
	if err != nil {
		panic(fmt.Sprintf("csvsource: failed to parse %s from fs: %v", filePath, err))
	}

	tableName := deriveTableName(filePath)
	mapRows := convertRows(columns, rows)

	if len(mapRows) == 0 {
		schema := inferSchema(columns, rows)
		return arraysource.NewWithSchema(mapRows, schema).Table(tableName)
	}

	return arraysource.New(mapRows).Table(tableName)
}

// NewCsvFileSource reads a CSV file, infers column types, and returns an
// array-backed data source ready for querying with the array driver.
//
// The first row of the file must be a header row defining column names.
// Remaining rows are parsed with type inference: each value is tried as
// int, then float, then bool, then time (RFC3339 and common date formats),
// falling back to string. Type widening applies across rows — a column
// with mixed int and float values becomes float.
//
// The table name is derived from the filename (without the extension).
// For example, "data/users.csv" → table name "users". Override with
// the .Table() method on the returned model if needed.
//
//	database.Query().
//	    Model(neat.NewCsvFileSource("data/users.csv")).
//	    Where("active = ?", true).
//	    Get(&users)
//
// Panics if the file cannot be opened, is empty, or has no header row —
// these are programmer errors (wrong path, malformed file), not runtime
// conditions.
func NewCsvFileSource(filePath string) *arraysource.Model {
	columns, rows, err := parseCSVFile(filePath)
	if err != nil {
		panic(fmt.Sprintf("csvsource: failed to parse %s: %v", filePath, err))
	}

	tableName := deriveTableName(filePath)
	mapRows := convertRows(columns, rows)

	if len(mapRows) == 0 {
		schema := inferSchema(columns, rows)
		return arraysource.NewWithSchema(mapRows, schema).Table(tableName)
	}

	return arraysource.New(mapRows).Table(tableName)
}

// NewCsvFileSourceWithDelimiter is like NewCsvFileSource but allows
// specifying a custom field delimiter (e.g., '\t' for TSV files).
//
//	database.Query().
//	    Model(neat.NewCsvFileSourceWithDelimiter("data/events.tsv", '\t')).
//	    Get(&events)
func NewCsvFileSourceWithDelimiter(filePath string, delimiter rune) *arraysource.Model {
	columns, rows, err := parseCSVFileWithDelimiter(filePath, delimiter)
	if err != nil {
		panic(fmt.Sprintf("csvsource: failed to parse %s: %v", filePath, err))
	}

	tableName := deriveTableName(filePath)
	mapRows := convertRows(columns, rows)

	if len(mapRows) == 0 {
		schema := inferSchema(columns, rows)
		return arraysource.NewWithSchema(mapRows, schema).Table(tableName)
	}

	return arraysource.New(mapRows).Table(tableName)
}

// MaxCSVRows limits the number of data rows that can be loaded from a single
// CSV file to prevent unbounded memory/CPU consumption. This mirrors the
// CSVDB driver's MaxCSVRows limit.
const MaxCSVRows = 100000

// deriveTableName extracts the table name from the file path by taking
// the base filename and removing the extension.
// "data/users.csv" → "users", "events.tsv" → "events"
func deriveTableName(filePath string) string {
	base := filepath.Base(filePath)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}

// parseCSVString parses a CSV string and returns the header columns and
// data rows.
func parseCSVString(csvString string) ([]string, [][]string, error) {
	return parseCSVReader(strings.NewReader(csvString), ',')
}

// parseCSVFile reads a CSV file and returns the header columns and data rows.
func parseCSVFile(filePath string) ([]string, [][]string, error) {
	return parseCSVFileWithDelimiter(filePath, ',')
}

// parseCSVFileWithDelimiter reads a CSV file with a custom delimiter.
func parseCSVFileWithDelimiter(filePath string, delimiter rune) ([]string, [][]string, error) {
	return parseCSVFSFileWithDelimiter(nil, filePath, delimiter)
}

// parseCSVFSFileWithDelimiter reads a CSV file from fs.FS (or os.Open if sys is nil) with a custom delimiter.
func parseCSVFSFileWithDelimiter(sys fs.FS, filePath string, delimiter rune) ([]string, [][]string, error) {
	var f io.ReadCloser
	if sys != nil {
		file, err := sys.Open(filePath)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot open file: %w", err)
		}
		f = file
	} else {
		file, err := os.Open(filePath)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot open file: %w", err)
		}
		f = file
	}
	defer func() { _ = f.Close() }()
	return parseCSVReader(f, delimiter)
}

// parseCSVReader reads CSV records from any io.Reader and returns the
// header columns and data rows. The first record is treated as the header
// (column names); all remaining records are data rows.
//
// A leading UTF-8 BOM (\xEF\xBB\xBF), if present, is stripped from the first
// header field. Rows with fewer fields than the header are allowed (missing
// fields become NULL). Rows with more fields are also allowed (extra fields
// are retained). An error is returned if the file is empty or if the row
// count exceeds MaxCSVRows.
func parseCSVReader(r io.Reader, delimiter rune) ([]string, [][]string, error) {
	reader := csv.NewReader(r)
	reader.Comma = delimiter
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		if err == io.EOF {
			return nil, nil, fmt.Errorf("no records found")
		}
		return nil, nil, fmt.Errorf("CSV parse error: %w", err)
	}

	if len(header) > 0 {
		header[0] = strings.TrimPrefix(header[0], "\xEF\xBB\xBF")
	}

	var rows [][]string
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

// convertRows converts string-based CSV rows into []map[string]any with
// inferred Go-native types. Each row becomes a map[string]any where keys
// are column names from the header and values are typed (int, float64,
// bool, time.Time, string, or nil for empty cells).
//
// Type inference is per-column: the type is determined by scanning all
// values in each column and applying widening (int → float → string).
// This means a column is consistently typed across all rows, not
// per-cell.
func convertRows(columns []string, rows [][]string) []map[string]any {
	if len(rows) == 0 {
		return nil
	}

	colTypes := inferColumnTypes(columns, rows)

	mapRows := make([]map[string]any, len(rows))
	for i, row := range rows {
		m := make(map[string]any, len(columns))
		for j, col := range columns {
			if j >= len(row) {
				m[col] = nil
				continue
			}
			m[col] = convertValue(row[j], colTypes[j])
		}
		mapRows[i] = m
	}
	return mapRows
}

// inferColumnTypes scans all rows and determines the SQLite type for each
// column. For each column, the type starts unknown ("") and is set by the
// first non-empty value. Subsequent values may widen the type:
// INTEGER → REAL → TEXT. DATETIME is detected separately and widens to
// TEXT if mixed with non-datetime types.
//
// Columns with no non-empty values default to TEXT.
func inferColumnTypes(columns []string, rows [][]string) []string {
	types := make([]string, len(columns))

	for _, row := range rows {
		for j := range columns {
			if j >= len(row) {
				continue
			}
			val := row[j]
			if val == "" {
				continue // skip empty cells, don't affect type
			}
			valType := inferValueType(val)
			if types[j] == "" {
				types[j] = valType
			} else {
				types[j] = widenType(types[j], valType)
			}
		}
	}

	// Columns with no non-empty values default to TEXT
	for i := range types {
		if types[i] == "" {
			types[i] = "TEXT"
		}
	}

	return types
}

// inferValueType tries to determine the type of a single string value.
func inferValueType(val string) string {
	if _, err := strconv.ParseInt(val, 10, 64); err == nil {
		return "INTEGER"
	}
	if _, err := strconv.ParseFloat(val, 64); err == nil {
		return "REAL"
	}
	if val == "true" || val == "false" || val == "True" || val == "False" {
		return "INTEGER" // SQLite stores bools as 0/1
	}
	if _, err := time.Parse(time.RFC3339, val); err == nil {
		return "DATETIME"
	}
	for _, layout := range []string{"2006-01-02", "2006-01-02 15:04:05", "01/02/2006"} {
		if _, err := time.Parse(layout, val); err == nil {
			return "DATETIME"
		}
	}
	return "TEXT"
}

// widenType returns the wider of two SQLite types.
// INTEGER → REAL → TEXT is the widening chain.
// DATETIME is compatible only with itself and widens to TEXT otherwise.
// An empty current type means "uninitialized" — the new type is adopted as-is.
func widenType(current, new string) string {
	if current == "" {
		return new
	}
	if current == new {
		return current
	}
	if current == "INTEGER" && new == "REAL" {
		return "REAL"
	}
	if current == "REAL" && new == "INTEGER" {
		return "REAL"
	}
	// Any incompatible mix → TEXT
	return "TEXT"
}

// convertValue converts a string value to the appropriate Go type for
// the given column type, for insertion via database/sql.
func convertValue(val string, colType string) any {
	if val == "" {
		return nil
	}

	switch colType {
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
		return val // fallback if inference was wrong for this specific value
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

// inferSchema builds a schema map from column names and inferred types,
// using the type names expected by the array driver ("int", "float",
// "bool", "time", "string").
func inferSchema(columns []string, rows [][]string) map[string]string {
	colTypes := inferColumnTypes(columns, rows)
	schema := make(map[string]string, len(columns))
	for i, col := range columns {
		schema[col] = sqliteTypeToSchemaType(colTypes[i])
	}
	return schema
}

// sqliteTypeToSchemaType converts internal SQLite type names to the
// schema type names expected by the array driver's ArraySchema interface.
func sqliteTypeToSchemaType(t string) string {
	switch t {
	case "INTEGER":
		return "int"
	case "REAL":
		return "float"
	case "DATETIME":
		return "time"
	default:
		return "string"
	}
}
