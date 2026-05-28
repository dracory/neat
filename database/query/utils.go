package query

import (
	"reflect"
	"strings"
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
	if v.Kind() == reflect.Ptr {
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
			return int64(field.Uint())
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
			if !(prev >= 'A' && prev <= 'Z') {
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

// structScanDests builds a scan-destination slice aligned to columns.
// Each element is either a pointer to the matching struct field or a *any placeholder.
// Non-addressable fields get a *T temporary that copyScanResults will copy back.
func structScanDests(v reflect.Value, columns []string) []any {
	colToPath := getColumnToIndexPath(v.Type())

	dests := make([]any, len(columns))
	for i, col := range columns {
		key := strings.ToLower(col)
		if path, ok := colToPath[key]; ok {
			field := v.FieldByIndex(path)
			if field.CanAddr() {
				dests[i] = field.Addr().Interface()
			} else {
				// allocate a temporary pointer of the field's type
				ptr := reflect.New(field.Type())
				dests[i] = ptr.Interface()
			}
		} else {
			var placeholder any
			dests[i] = &placeholder
		}
	}
	return dests
}

// copyScanResults copies values from non-addressable temporaries back into struct fields.
// For addressable fields the scan wrote directly into them; this is a no-op for those.
func copyScanResults(v reflect.Value, columns []string, dests []any) {
	colToPath := getColumnToIndexPath(v.Type())
	for i, col := range columns {
		key := strings.ToLower(col)
		path, ok := colToPath[key]
		if !ok {
			continue
		}
		field := v.FieldByIndex(path)
		if field.CanAddr() {
			continue // already written by Scan
		}
		// copy from the temporary pointer
		ptrVal := reflect.ValueOf(dests[i])
		if ptrVal.Kind() == reflect.Ptr && !ptrVal.IsNil() {
			val := ptrVal.Elem()
			if val.Type().AssignableTo(field.Type()) && field.CanSet() {
				field.Set(val)
			}
		}
	}
}

// setModelPrimaryKey sets the primary key field (ID or Id) on a struct model to the given value.
func setModelPrimaryKey(value any, id int64) {
	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Ptr || v.IsNil() {
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
			field.SetUint(uint64(id))
			return
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			field.SetInt(id)
			return
		}
	}
}
