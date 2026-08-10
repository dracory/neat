package arraysource

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync/atomic"
	"time"

	"github.com/dracory/neat/support/structref"
)

// NewArraySourceFrom creates an array-backed data source from a slice of
// structs or []map[string]any, with zero setup: no custom struct, no table
// name, no schema. This is the primary, day-to-day entry point for the
// array driver — the third constructor in the NewArraySource family.
//
//	statuses := []Status{{ID: 1, Name: "Pending"}, {ID: 2, Name: "Active"}}
//	database.Query().Model(neat.NewArraySourceFrom(statuses)).Get(&out)
//
// For struct slices, field names are resolved to column names using the same
// tag convention as the ORM (db > neat > gorm > snake_case). Embedded structs
// are flattened. Association fields (slices, struct pointers) are skipped.
// Nullable pointer fields (*string, *int) are dereferenced; nil pointers
// become nil (NULL in SQLite). time.Time fields are included as-is.
//
// For []map[string]any, rows are shallow-copied (snapshot at call time).
//
// Panics if T is not a struct, pointer-to-struct, or map[string]any — this
// is a programmer error, not a runtime condition.
// Panics if the slice is empty — use NewWithSchema for empty datasets.
func NewArraySourceFrom[T any](items []T) *Model {
	if len(items) == 0 {
		panic("arraysource: NewArraySourceFrom() requires non-empty items; use NewWithSchema() for empty datasets")
	}

	// Type assertion on the slice — avoids reflect.TypeOf(zero) which panics
	// on nil interface zero values
	if rows, ok := any(items).([]map[string]any); ok {
		return &Model{
			table: nextTableName[T](),
			data:  copyRows(rows),
		}
	}

	rows := structsToRows(items)
	if len(rows) == 0 {
		panic("arraysource: NewArraySourceFrom() produced no rows; check that the struct has exported fields")
	}

	return &Model{
		table: nextTableName[T](),
		data:  rows,
	}
}

// copyRows creates a shallow copy of each map so the source is a snapshot
// at call time — mutating the caller's maps after NewArraySourceFrom doesn't
// affect the array source. Both the struct path and map path now have the
// same "snapshot at call time" contract.
func copyRows(rows []map[string]any) []map[string]any {
	out := make([]map[string]any, len(rows))
	for i, row := range rows {
		m := make(map[string]any, len(row))
		for k, v := range row {
			m[k] = v
		}
		out[i] = m
	}
	return out
}

func structsToRows[T any](items []T) []map[string]any {
	rows := make([]map[string]any, 0, len(items))
	for i := range items {
		v := reflect.ValueOf(items[i])
		if v.Kind() == reflect.Pointer {
			if v.IsNil() {
				continue // skip nil elements
			}
			v = v.Elem()
		}
		if v.Kind() != reflect.Struct {
			panic(fmt.Sprintf("arraysource: NewArraySourceFrom requires struct or map[string]any, got %s", v.Kind()))
		}
		rows = append(rows, structToMap(v))
	}
	return rows
}

func structToMap(v reflect.Value) map[string]any {
	row := make(map[string]any)
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Flatten embedded structs
		if field.Anonymous && fieldValue.Kind() == reflect.Struct {
			for k, val := range structToMap(fieldValue) {
				row[k] = val
			}
			continue
		}

		// Skip unexported fields
		if !fieldValue.CanInterface() {
			continue
		}

		// Skip fields explicitly ignored with db:"-", neat:"-", gorm:"-"
		ignored := false
		for _, tag := range []string{"db", "neat", "gorm"} {
			if v := field.Tag.Get(tag); v == "-" {
				ignored = true
				break
			}
		}
		if ignored {
			continue
		}

		// Skip association fields — mirrors builder_extract.go logic:
		// - Slices (except []byte, json.RawMessage) → associations
		// - Structs (except time.Time) → associations
		// - Pointers to structs → associations
		// - Pointers to basic types (*string, *int) → KEEP, dereference
		if isAssociationField(fieldValue) {
			continue
		}

		col := structref.FieldColumnName(field)
		if col == "" || col == "-" {
			continue
		}

		// Dereference non-nil pointers to basic types
		// Nil pointers → nil (becomes NULL in SQLite)
		if fieldValue.Kind() == reflect.Pointer {
			if fieldValue.IsNil() {
				row[col] = nil
			} else {
				row[col] = fieldValue.Elem().Interface()
			}
			continue
		}

		row[col] = fieldValue.Interface()
	}
	return row
}

// isAssociationField returns true for fields that represent associations
// rather than scalar columns. Mirrors the logic in
// database/query/builder_extract.go:155-176.
func isAssociationField(fieldValue reflect.Value) bool {
	switch fieldValue.Kind() {
	case reflect.Slice:
		// []byte and json.RawMessage are scalar-like, keep them
		if fieldValue.Type() == reflect.TypeOf([]byte(nil)) ||
			fieldValue.Type() == reflect.TypeOf(json.RawMessage(nil)) {
			return false
		}
		return true
	case reflect.Struct:
		// time.Time is a scalar column type, keep it
		if fieldValue.Type() == reflect.TypeOf(time.Time{}) {
			return false
		}
		return true
	case reflect.Pointer:
		elemType := fieldValue.Type().Elem()
		if elemType == reflect.TypeOf(time.Time{}) {
			return false
		}
		if fieldValue.IsNil() {
			// Nil pointer: skip if it points to a struct (association),
			// keep if it points to a basic type (nullable column)
			return elemType.Kind() == reflect.Struct
		}
		// Non-nil pointer: skip if it points to a struct (association),
		// keep if it points to a basic type (nullable column)
		return fieldValue.Elem().Kind() == reflect.Struct
	default:
		return false
	}
}

// ConvertSliceToRows converts a slice of maps, structs, or pointer-to-structs
// to []map[string]any using reflection.
// If data is nil, or represents an empty slice, it returns nil, nil.
// If any element of the slice is not a struct, pointer-to-struct, or map[string]any,
// it returns an error.
func ConvertSliceToRows(data any) ([]map[string]any, error) {
	if data == nil {
		return nil, nil
	}

	// Direct type assertion for []map[string]any
	if rows, ok := data.([]map[string]any); ok {
		if len(rows) == 0 {
			return nil, nil
		}
		return copyRows(rows), nil
	}

	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil, nil
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Slice {
		return nil, fmt.Errorf("arraysource: expected slice, got %T", data)
	}

	l := v.Len()
	if l == 0 {
		return nil, nil
	}

	rows := make([]map[string]any, 0, l)
	for i := 0; i < l; i++ {
		elem := v.Index(i)

		if !elem.IsValid() {
			continue
		}

		// Unpack interface if the slice is e.g. []any
		if elem.Kind() == reflect.Interface {
			if elem.IsNil() {
				continue // skip nil elements
			}
			elem = elem.Elem()
		}

		if !elem.IsValid() {
			continue
		}

		// Handle pointer elements
		if elem.Kind() == reflect.Pointer {
			if elem.IsNil() {
				continue // skip nil elements
			}
			elem = elem.Elem()
		}

		if !elem.IsValid() {
			continue
		}

		// Handle map[string]any elements
		if m, ok := elem.Interface().(map[string]any); ok {
			// Copy map to satisfy snapshot semantics
			cp := make(map[string]any, len(m))
			for k, val := range m {
				cp[k] = val
			}
			rows = append(rows, cp)
			continue
		}

		// Check if it is a struct
		if elem.Kind() == reflect.Struct {
			rows = append(rows, structToMap(elem))
			continue
		}

		return nil, fmt.Errorf("arraysource: slice element must be a struct, pointer to struct, or map[string]any, got %s", elem.Kind())
	}

	return rows, nil
}

var arrayCounter uint64

func nextTableName[T any]() string {
	n := atomic.AddUint64(&arrayCounter, 1)
	var name string
	t := reflect.TypeFor[T]()
	// Handle pointer types: []*Status → "status", not ""
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	// Handle slice types (shouldn't happen in practice, but be safe)
	if t.Kind() == reflect.Slice {
		t = t.Elem()
		if t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
	}
	name = strings.ToLower(t.Name())
	if name == "" {
		name = "array"
	}
	return fmt.Sprintf("array_%s_%d", name, n)
}
