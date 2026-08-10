package driver

import (
	"encoding/json"
	"sort"
	"time"
)

// inferMapSchema examines all rows (each a map[string]any) and returns:
//   - sorted column names (union of all keys across all rows)
//   - SQLite type for each column (inferred from native Go types)
//
// This is the shared implementation used by both the JSONDB and XMLDB drivers,
// which both operate on []map[string]any rows with native Go types
// (int64, float64, bool, time.Time, string).
func inferMapSchema(rows []map[string]any) ([]string, []string) {
	colMap := make(map[string]string) // columnName -> inferredType
	for _, row := range rows {
		for k, v := range row {
			if v == nil {
				if _, exists := colMap[k]; !exists {
					colMap[k] = "" // Initial empty type
				}
				continue
			}

			valType := inferGoValueType(v)
			if currentType, exists := colMap[k]; !exists || currentType == "" {
				colMap[k] = valType
			} else {
				colMap[k] = widenType(currentType, valType)
			}
		}
	}

	// Collect column names and sort them alphabetically for deterministic ordering
	var columns []string
	for k := range colMap {
		columns = append(columns, k)
	}
	sort.Strings(columns)

	// Build types slice corresponding to sorted columns
	types := make([]string, len(columns))
	for i, col := range columns {
		t := colMap[col]
		if t == "" {
			t = "TEXT" // Default to TEXT for columns with only nil values
		}
		types[i] = t
	}

	return columns, types
}

// inferGoValueType determines the SQLite type for a native Go value.
// Used by both JSONDB and XMLDB drivers, which produce native Go types
// (int64, float64, bool, time.Time, string) from their respective parsers.
// Consolidated and widened to support the full set of Go types for the GODB driver.
func inferGoValueType(v any) string {
	switch v.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return "INTEGER"
	case float32, float64:
		return "REAL"
	case bool:
		return "INTEGER"
	case time.Time, *time.Time:
		return "DATETIME"
	case []byte, json.RawMessage:
		return "BLOB"
	default:
		return "TEXT"
	}
}

// convertGoValue converts a native Go value into a form suitable for
// database/sql insertion based on the inferred SQLite type.
// Used by JSONDB, XMLDB, and GODB drivers.
func convertGoValue(val any, sqlType string) any {
	if val == nil {
		return nil
	}

	switch sqlType {
	case "INTEGER":
		switch v := val.(type) {
		case bool:
			if v {
				return int64(1)
			}
			return int64(0)
		case int:
			return int64(v)
		case int8:
			return int64(v)
		case int16:
			return int64(v)
		case int32:
			return int64(v)
		case int64:
			return v
		case uint:
			return int64(v)
		case uint8:
			return int64(v)
		case uint16:
			return int64(v)
		case uint32:
			return int64(v)
		case uint64:
			return int64(v)
		case float32:
			return int64(v)
		case float64:
			return int64(v)
		}
	case "REAL":
		switch v := val.(type) {
		case float32:
			return float64(v)
		case float64:
			return v
		case int:
			return float64(v)
		case int8:
			return float64(v)
		case int16:
			return float64(v)
		case int32:
			return float64(v)
		case int64:
			return float64(v)
		case uint:
			return float64(v)
		case uint8:
			return float64(v)
		case uint16:
			return float64(v)
		case uint32:
			return float64(v)
		case uint64:
			return float64(v)
		}
	case "DATETIME":
		if t, ok := val.(time.Time); ok {
			return t
		}
		if t, ok := val.(*time.Time); ok {
			if t != nil {
				return *t
			}
			return nil
		}
	}
	return val
}
