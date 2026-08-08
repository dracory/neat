package structref

import (
	"reflect"
	"strings"
)

// FieldColumnName returns the column name for a struct field by checking
// db, neat, gorm tags (in that order), then falling back to snake_case
// of the field name.
//
// This is the single source of truth for struct→column name resolution,
// shared by the query builder and the array source helper.
func FieldColumnName(f reflect.StructField) string {
	for _, tag := range []string{"db", "neat", "gorm"} {
		if v := f.Tag.Get(tag); v != "" && v != "-" {
			parts := strings.SplitN(v, ";", 2)
			if len(parts) == 0 {
				continue
			}
			part := parts[0]
			if strings.HasPrefix(part, "column:") {
				return strings.TrimPrefix(part, "column:")
			}
			if tag == "db" || tag == "neat" {
				return part
			}
		}
	}
	return CamelToSnake(f.Name)
}

// CamelToSnake converts CamelCase to snake_case.
func CamelToSnake(s string) string {
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
