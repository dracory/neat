package query

import (
	"database/sql"
	"reflect"
	"strings"
	"time"
)

// toCamelCase converts snake_case to CamelCase
func toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	for i := range parts {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + strings.ToLower(parts[i][1:])
		}
	}
	return strings.Join(parts, "")
}

// getPrimaryKeyValue returns the primary key value (ID/Id) of a struct as int64, 0 if absent or zero.
func getPrimaryKeyValue(value any) int64 {
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return 0
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return 0
	}
	for _, name := range []string{"ID", "Id"} {
		field := v.FieldByName(name)
		if !field.IsValid() {
			continue
		}
		switch field.Kind() {
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			u := field.Uint()
			if u > uint64(1<<63-1) {
				return 0 // Overflow, return 0
			}
			return int64(u)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return field.Int()
		}
	}
	return 0
}

// structFieldColumnName returns the column name for a struct field by checking
// db, neat, gorm tags (in that order), then falling back to a snake_case of the field name.
func structFieldColumnName(f reflect.StructField) string {
	for _, tag := range []string{"db", "neat", "gorm"} {
		if v := f.Tag.Get(tag); v != "" && v != "-" {
			// take the first semicolon-delimited part; for gorm it may be "column:name"
			parts := strings.SplitN(v, ";", 2)
			if len(parts) == 0 {
				continue
			}
			part := parts[0]
			if strings.HasPrefix(part, "column:") {
				return strings.TrimPrefix(part, "column:")
			}
			// db and neat tags use the value directly as the column name
			if tag == "db" || tag == "neat" {
				return part
			}
		}
	}
	// snake_case the Go field name
	return camelToSnake(f.Name)
}

// camelToSnake converts CamelCase to snake_case.
func camelToSnake(s string) string {
	var out []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			// Don't add underscore if previous char is also uppercase (acronym handling)
			prev := s[i-1]
			if prev < 'A' || prev > 'Z' {
				out = append(out, '_')
			}
		}
		out = append(out, []rune(strings.ToLower(string(r)))...)
	}
	return string(out)
}

func getColumnToIndexPath(t reflect.Type) map[string][]int {
	m := make(map[string][]int)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Anonymous && f.Type.Kind() == reflect.Struct {
			for k, path := range getColumnToIndexPath(f.Type) {
				if _, ok := m[k]; !ok {
					newPath := make([]int, len(path)+1)
					newPath[0] = i
					copy(newPath[1:], path)
					m[k] = newPath
				}
			}
			continue
		}
		col := strings.ToLower(structFieldColumnName(f))
		if _, ok := m[col]; !ok {
			m[col] = f.Index
		}
	}
	return m
}

// nullableScanDest returns a nullable scan destination for a field type,
// so that NULL values from LEFT JOINs don't cause "converting NULL to T" errors.
// Returns nil if no nullable wrapper is needed (field is already a pointer or interface).
func nullableScanDest(fieldType reflect.Type) any {
	switch fieldType.Kind() {
	case reflect.String:
		return new(sql.NullString)
	case reflect.Bool:
		return new(sql.NullBool)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return new(sql.NullInt64)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return new(sql.NullInt64)
	case reflect.Float32, reflect.Float64:
		return new(sql.NullFloat64)
	}
	// time.Time
	if fieldType == reflect.TypeOf(time.Time{}) {
		return new(sql.NullTime)
	}
	return nil
}

// structScanDests builds a scan-destination slice aligned to columns.
// Each element is either a pointer to the matching struct field or a *any placeholder.
// Non-addressable fields and non-pointer value types use nullable wrappers so that
// NULL values from LEFT JOINs don't cause scan errors.
func structScanDests(v reflect.Value, columns []string) []any {
	colToPath := getColumnToIndexPath(v.Type())

	dests := make([]any, len(columns))
	for i, col := range columns {
		key := strings.ToLower(col)
		if path, ok := colToPath[key]; ok {
			field := v.FieldByIndex(path)
			ft := field.Type()
			// Use nullable wrappers for non-pointer value types so NULL doesn't
			// cause "converting NULL to T is unsupported" scan errors.
			if ft.Kind() != reflect.Pointer && ft.Kind() != reflect.Interface {
				if nd := nullableScanDest(ft); nd != nil {
					dests[i] = nd
					continue
				}
			}
			if field.CanAddr() {
				dests[i] = field.Addr().Interface()
			} else {
				// allocate a temporary pointer of the field's type
				ptr := reflect.New(ft)
				dests[i] = ptr.Interface()
			}
		} else {
			var placeholder any
			dests[i] = &placeholder
		}
	}
	return dests
}

// copyScanResults copies values from nullable wrappers and non-addressable temporaries
// back into struct fields after scanning.
func copyScanResults(v reflect.Value, columns []string, dests []any) {
	colToPath := getColumnToIndexPath(v.Type())
	for i, col := range columns {
		key := strings.ToLower(col)
		path, ok := colToPath[key]
		if !ok {
			continue
		}
		field := v.FieldByIndex(path)
		if !field.CanSet() {
			continue
		}
		dest := dests[i]

		// Handle nullable wrappers produced by nullableScanDest
		switch d := dest.(type) {
		case *sql.NullString:
			if d.Valid {
				field.SetString(d.String)
			}
			continue
		case *sql.NullBool:
			if d.Valid {
				field.SetBool(d.Bool)
			}
			continue
		case *sql.NullInt64:
			if d.Valid {
				switch field.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					field.SetInt(d.Int64)
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					if d.Int64 >= 0 {
						field.SetUint(uint64(d.Int64))
					}
				}
			}
			continue
		case *sql.NullFloat64:
			if d.Valid {
				field.SetFloat(d.Float64)
			}
			continue
		case *sql.NullTime:
			if d.Valid {
				field.Set(reflect.ValueOf(d.Time))
			}
			continue
		}

		// For addressable fields that were scanned directly, nothing to do.
		if field.CanAddr() && reflect.ValueOf(dest) == reflect.ValueOf(field.Addr().Interface()) {
			continue
		}

		// copy from the temporary pointer
		ptrVal := reflect.ValueOf(dest)
		if ptrVal.Kind() == reflect.Pointer && !ptrVal.IsNil() {
			val := ptrVal.Elem()
			if val.Type().AssignableTo(field.Type()) {
				field.Set(val)
			}
		}
	}
}

// setModelPrimaryKey sets the primary key field (ID or Id) on a struct model to the given value.
// Supports int64 for integer PKs and string for short-ID PKs.
func setModelPrimaryKey(value any, id any) {
	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return
	}
	for _, name := range []string{"ID", "Id"} {
		field := v.FieldByName(name)
		if !field.IsValid() || !field.CanSet() {
			continue
		}
		switch field.Kind() {
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if i, ok := id.(int64); ok && i >= 0 {
				field.SetUint(uint64(i))
			}
			return
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if i, ok := id.(int64); ok {
				field.SetInt(i)
			}
			return
		case reflect.String:
			if s, ok := id.(string); ok {
				field.SetString(s)
			}
			return
		}
	}
}

// getPrimaryKeyValueAny returns the primary key value (ID/Id) of a struct as any.
// Returns the value and true if found, or nil and false if absent.
func getPrimaryKeyValueAny(value any) (any, bool) {
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil, false
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, false
	}
	for _, name := range []string{"ID", "Id"} {
		field := v.FieldByName(name)
		if !field.IsValid() {
			continue
		}
		switch field.Kind() {
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			u := field.Uint()
			if u > uint64(1<<63-1) {
				return int64(0), true
			}
			return int64(u), true
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return field.Int(), true
		case reflect.String:
			return field.String(), true
		}
	}
	return nil, false
}

// isPrimaryKeyZero reports whether the primary key is unset.
// For integers: zero value; for strings: empty string.
func isPrimaryKeyZero(value any) bool {
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return true
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return true
	}
	for _, name := range []string{"ID", "Id"} {
		field := v.FieldByName(name)
		if !field.IsValid() {
			continue
		}
		switch field.Kind() {
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return field.Uint() == 0
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return field.Int() == 0
		case reflect.String:
			return field.String() == ""
		}
	}
	return true
}

// isShortIDModel reports whether the value is a struct (or slice of structs)
// with a string ID field, indicating it uses client-generated short IDs.
func isShortIDModel(value any) bool {
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return false
		}
		v = v.Elem()
	}
	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		if v.Len() == 0 {
			return false
		}
		elem := v.Index(0)
		if elem.Kind() == reflect.Pointer {
			elem = elem.Elem()
		}
		v = elem
	}
	if v.Kind() != reflect.Struct {
		return false
	}
	for _, name := range []string{"ID", "Id"} {
		field := v.FieldByName(name)
		if !field.IsValid() {
			continue
		}
		if field.Kind() == reflect.String {
			return true
		}
	}
	return false
}

// applyWhereConditions applies attributes as WHERE conditions to a query.
func applyWhereConditions(q *Query, attributes any) error {
	// Handle map[string]any attributes
	if attrMap, ok := attributes.(map[string]any); ok {
		for key, value := range attrMap {
			q.Where(key+" = ?", value)
		}
		return nil
	}

	// Handle struct attributes
	attrValue := reflect.ValueOf(attributes)
	if attrValue.Kind() == reflect.Pointer {
		attrValue = attrValue.Elem()
	}

	if attrValue.Kind() == reflect.Struct {
		attrType := attrValue.Type()
		for i := 0; i < attrValue.NumField(); i++ {
			field := attrValue.Field(i)
			fieldType := attrType.Field(i)

			// Skip unexported fields and zero values
			if !field.CanInterface() || field.IsZero() {
				continue
			}

			// Use field name as column name
			columnName := fieldType.Name
			q.Where(columnName+" = ?", field.Interface())
		}
	}

	return nil
}

// applyAttributes applies attributes from a map or struct to a destination struct.
func applyAttributes(dest any, attributes any) error {
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() == reflect.Pointer {
		destValue = destValue.Elem()
	}

	// Handle map[string]any attributes
	if attrMap, ok := attributes.(map[string]any); ok {
		if destValue.Kind() == reflect.Struct {
			destType := destValue.Type()
			for key, value := range attrMap {
				// Try to find field by name (case-insensitive)
				var field reflect.Value
				for i := 0; i < destValue.NumField(); i++ {
					fieldType := destType.Field(i)
					// Check db tag first, then field name
					dbTag := fieldType.Tag.Get("db")
					if dbTag == key {
						field = destValue.Field(i)
						break
					}
					// Case-insensitive field name matching
					if strings.EqualFold(fieldType.Name, key) {
						field = destValue.Field(i)
						break
					}
				}

				if field.IsValid() && field.CanSet() {
					// Handle type conversion
					val := reflect.ValueOf(value)
					if val.Type().ConvertibleTo(field.Type()) {
						field.Set(val.Convert(field.Type()))
					} else if val.Type() == field.Type() {
						field.Set(val)
					}
				}
			}
		}
		return nil
	}

	// Handle struct attributes
	attrValue := reflect.ValueOf(attributes)
	if attrValue.Kind() == reflect.Pointer {
		attrValue = attrValue.Elem()
	}

	if attrValue.Kind() == reflect.Struct && destValue.Kind() == reflect.Struct {
		attrType := attrValue.Type()
		for i := 0; i < attrValue.NumField(); i++ {
			field := attrValue.Field(i)
			fieldType := attrType.Field(i)

			// Skip unexported fields
			if !field.CanInterface() {
				continue
			}

			// Try to find matching field in destination by name
			destField := destValue.FieldByName(fieldType.Name)
			if destField.IsValid() && destField.CanSet() {
				// Handle type conversion
				if field.Type().ConvertibleTo(destField.Type()) {
					destField.Set(field.Convert(destField.Type()))
				} else if field.Type() == destField.Type() {
					destField.Set(field)
				}
			}
		}
	}

	return nil
}

// isSimpleIdentifier checks if a string is a simple column identifier
// that can be safely quoted. Returns false for:
// - Identifiers with dots (table.column)
// - Identifiers with parentheses (function calls)
// - Identifiers starting with numbers
// - Empty strings
// - SQL keywords (to prevent injection attempts)
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

	// Reject SQL keywords to prevent injection attempts
	// This is especially important for Oracle ID retrieval where table names
	// are extracted from SQL strings and used in subsequent queries
	upperS := strings.ToUpper(s)
	sqlKeywords := []string{
		"SELECT", "INSERT", "UPDATE", "DELETE", "DROP", "CREATE",
		"ALTER", "TRUNCATE", "REPLACE", "MERGE", "UNION", "EXCEPT",
		"INTERSECT", "WHERE", "FROM", "JOIN", "INNER", "OUTER",
		"LEFT", "RIGHT", "FULL", "CROSS", "ON", "USING", "AND",
		"OR", "NOT", "IN", "EXISTS", "BETWEEN", "LIKE", "IS",
		"NULL", "TRUE", "FALSE", "CASE", "WHEN", "THEN", "ELSE",
		"END", "GROUP", "HAVING", "ORDER", "BY", "LIMIT", "OFFSET",
		"DISTINCT", "ALL", "AS", "TABLE", "VIEW", "INDEX", "TRIGGER",
		"PROCEDURE", "FUNCTION", "DATABASE", "SCHEMA", "GRANT", "REVOKE",
		"EXEC", "EXECUTE", "DUAL", "SYSDATE", "SYSTIMESTAMP",
	}

	for _, keyword := range sqlKeywords {
		if upperS == keyword {
			return false
		}
	}

	return true
}
