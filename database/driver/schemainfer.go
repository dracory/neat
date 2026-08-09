package driver

import (
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
func inferGoValueType(v any) string {
	switch v.(type) {
	case int64:
		return "INTEGER"
	case float64:
		return "REAL"
	case bool:
		return "INTEGER"
	case time.Time:
		return "DATETIME"
	default:
		return "TEXT"
	}
}

// convertGoValue converts a native Go value into a form suitable for
// database/sql insertion based on the inferred SQLite type.
// Used by both JSONDB and XMLDB drivers.
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
		case int64:
			return v
		case float64:
			return int64(v)
		}
	case "REAL":
		switch v := val.(type) {
		case float64:
			return v
		case int64:
			return float64(v)
		}
	case "DATETIME":
		if t, ok := val.(time.Time); ok {
			return t
		}
	}
	return val
}
