package jsonsource

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dracory/neat/support/arraysource"
)

// NewJsonSource parses a JSON or JSONL string and returns an array-backed
// data source ready for querying with the array driver.
//
// The content must be either:
//   - A JSON array of objects: [{"id":1,"name":"Alice"},...]
//   - JSONL/NDJSON (one object per line): {"id":1,"name":"Alice"}\n{"id":2,...}
//
// Pass isJSONL=true for JSONL content, false for a JSON array. For JSONL,
// empty lines are skipped.
//
// JSON has native types, so no string-to-type inference is needed — values
// are already int, float64, bool, string, or nil. String values that match
// RFC3339 format are converted to time.Time so the array driver maps them
// to DATETIME columns. Nested objects and arrays are stored as JSON strings
// (queryable via SQLite JSON functions).
//
// The table name must be provided explicitly since there is no filename
// to derive it from. Override with the .Table() method on the returned
// model if needed.
//
//	database.Query().
//	    Model(neat.NewJsonSource(jsonString, "users", false)).
//	    Where("active = ?", true).
//	    Get(&users)
//
// Panics if the content cannot be parsed — this is a programmer error
// (malformed JSON), not a runtime condition.
// Panics if the content cannot be parsed or is an empty array — there is
// no header row like CSV to infer column names from, so an empty JSON array
// cannot produce a queryable table. This is a programmer error, not a
// runtime condition.
func NewJsonSource(jsonString string, tableName string, isJSONL bool) *arraysource.Model {
	rows, err := parseJSONString(jsonString, isJSONL)
	if err != nil {
		panic(fmt.Sprintf("jsonsource: failed to parse JSON string: %v", err))
	}

	mapRows := normalizeRows(rows)

	if len(mapRows) == 0 {
		panic("jsonsource: empty JSON array or no objects found — cannot infer schema without data rows")
	}

	return arraysource.New(mapRows).Table(tableName)
}

// NewJsonFileSource reads a JSON or JSONL file and returns an array-backed
// data source ready for querying with the array driver.
//
// The file must contain either:
//   - A JSON array of objects: [{"id":1,"name":"Alice"},...]
//   - JSONL/NDJSON (one object per line): {"id":1,"name":"Alice"}\n{"id":2,...}
//
// JSONL mode is auto-detected from the file extension (.jsonl or .ndjson).
// For .json files, the content is parsed as a single JSON array.
//
// JSON has native types, so no string-to-type inference is needed — values
// are already int, float64, bool, string, or nil. String values that match
// RFC3339 format are converted to time.Time so the array driver maps them
// to DATETIME columns. Nested objects and arrays are stored as JSON strings
// (queryable via SQLite JSON functions).
//
// The table name is derived from the filename (without the extension).
// For example, "data/users.json" → table name "users". Override with
// the .Table() method on the returned model if needed.
//
//	database.Query().
//	    Model(neat.NewJsonFileSource("data/users.json")).
//	    Where("active = ?", true).
//	    Get(&users)
//
// Panics if the file cannot be opened, parsed, or contains an empty array
// — there is no header row like CSV to infer column names from, so an empty
// JSON array cannot produce a queryable table. These are programmer errors
// (wrong path, malformed file), not runtime conditions.
func NewJsonFileSource(filePath string) *arraysource.Model {
	rows, err := parseJSONFile(filePath)
	if err != nil {
		panic(fmt.Sprintf("jsonsource: failed to parse %s: %v", filePath, err))
	}

	tableName := deriveTableName(filePath)
	mapRows := normalizeRows(rows)

	if len(mapRows) == 0 {
		panic(fmt.Sprintf("jsonsource: %s contains no data rows — cannot infer schema without data", filePath))
	}

	return arraysource.New(mapRows).Table(tableName)
}

// deriveTableName extracts the table name from the file path by taking
// the base filename and removing the extension.
// "data/users.json" → "users", "events.jsonl" → "events"
func deriveTableName(filePath string) string {
	base := filepath.Base(filePath)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}

// isJSONLFile checks if the file extension indicates JSONL/NDJSON format.
func isJSONLFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".jsonl" || ext == ".ndjson"
}

// parseJSONString parses a JSON or JSONL string and returns raw rows.
func parseJSONString(content string, isJSONL bool) ([]map[string]any, error) {
	if isJSONL {
		return parseJSONLReader(strings.NewReader(content))
	}
	return parseJSONArrayReader(bytes.NewReader([]byte(content)))
}

// parseJSONFile reads a JSON or JSONL file and returns raw rows.
func parseJSONFile(filePath string) ([]map[string]any, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}
	defer f.Close()

	if isJSONLFile(filePath) {
		return parseJSONLReader(f)
	}
	return parseJSONArrayReader(f)
}

// parseJSONArrayReader parses a JSON array of objects from any io.Reader.
// Returns an error if there is trailing data after the closing bracket.
func parseJSONArrayReader(r io.Reader) ([]map[string]any, error) {
	dec := json.NewDecoder(r)

	// Read the opening bracket
	tok, err := dec.Token()
	if err != nil {
		return nil, fmt.Errorf("JSON parse error: %w", err)
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '[' {
		return nil, fmt.Errorf("expected JSON array, got %v", tok)
	}

	var rows []map[string]any
	for dec.More() {
		var row map[string]any
		if err := dec.Decode(&row); err != nil {
			return nil, fmt.Errorf("JSON decode error: %w", err)
		}
		rows = append(rows, row)
	}

	// Consume the closing bracket
	tok, err = dec.Token()
	if err != nil {
		return nil, fmt.Errorf("JSON parse error (expected closing bracket): %w", err)
	}
	if delim, ok := tok.(json.Delim); !ok || delim != ']' {
		return nil, fmt.Errorf("expected closing bracket, got %v", tok)
	}

	// Check for trailing data — reject malformed input like `][]`
	if dec.More() {
		return nil, fmt.Errorf("unexpected trailing data after JSON array")
	}

	return rows, nil
}

// parseJSONLReader parses JSONL (one JSON object per line) from any io.Reader.
// Empty lines are skipped.
func parseJSONLReader(r io.Reader) ([]map[string]any, error) {
	var rows []map[string]any
	scanner := bufio.NewScanner(r)
	// Increase buffer for long lines (default 64KB is too small for some data)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var row map[string]any
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			return nil, fmt.Errorf("JSONL decode error: %w", err)
		}
		rows = append(rows, row)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("JSONL read error: %w", err)
	}

	return rows, nil
}

// normalizeRows processes raw JSON rows to make them compatible with the
// array driver:
//   - String values matching RFC3339 are converted to time.Time
//   - Nested maps and arrays are stored as JSON strings
//   - Other values (int, float64, bool, nil, string) are kept as-is
func normalizeRows(rows []map[string]any) []map[string]any {
	if len(rows) == 0 {
		return nil
	}

	normalized := make([]map[string]any, len(rows))
	for i, row := range rows {
		m := make(map[string]any, len(row))
		for k, v := range row {
			m[k] = normalizeValue(v)
		}
		normalized[i] = m
	}
	return normalized
}

// normalizeValue converts a JSON-native value into a form the array driver
// can handle:
//   - float64 with integer value → int64 (JSON has no int type, but the
//     array driver infers INTEGER columns from int64 values)
//   - map[string]any → JSON string (via json.Marshal)
//   - []any → JSON string (via json.Marshal)
//   - string matching RFC3339 → time.Time
//   - everything else → unchanged
func normalizeValue(v any) any {
	switch val := v.(type) {
	case float64:
		// encoding/json decodes all numbers as float64. Convert whole
		// numbers to int64 so the array driver infers INTEGER columns
		// instead of REAL. This matches the CSV source's behavior and
		// avoids precision loss for large integers.
		if val == float64(int64(val)) {
			return int64(val)
		}
		return val
	case map[string]any:
		b, err := json.Marshal(val)
		if err != nil {
			return fmt.Sprintf("%v", val)
		}
		return string(b)
	case []any:
		b, err := json.Marshal(val)
		if err != nil {
			return fmt.Sprintf("%v", val)
		}
		return string(b)
	case string:
		// Try to parse as RFC3339 timestamp
		if t, err := time.Parse(time.RFC3339, val); err == nil {
			return t
		}
		return val
	default:
		return v
	}
}
