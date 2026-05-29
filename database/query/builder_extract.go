package query

import (
	"fmt"
	"reflect"
	"time"
)

// extractColumnsAndValues extracts column names and values from a struct, map, or slice.
func (b *Builder) extractColumnsAndValues(value any) ([]string, []any, error) {
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Handle slice/array for bulk insert
	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		if v.Len() == 0 {
			return nil, []any{}, nil
		}

		var allColumns []string
		var allValues []any

		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i).Interface()
			cols, vals, err := b.extractSingleColumnsAndValues(elem)
			if err != nil {
				return nil, nil, err
			}

			if i == 0 {
				allColumns = cols
			}
			allValues = append(allValues, vals...)
		}

		if allValues == nil {
			allValues = []any{}
		}
		return allColumns, allValues, nil
	}

	return b.extractSingleColumnsAndValues(value)
}

// extractSingleColumnsAndValues extracts column names and values from a single struct or map.
func (b *Builder) extractSingleColumnsAndValues(value any) ([]string, []any, error) {
	var columns []string
	var values []any

	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Handle map[string]any
	if v.Kind() == reflect.Map {
		// Sort keys for deterministic SQL generation
		keys := v.MapKeys()
		sortedKeys := make([]reflect.Value, len(keys))
		copy(sortedKeys, keys)
		// Sort by string key name
		for i := 0; i < len(sortedKeys); i++ {
			for j := i + 1; j < len(sortedKeys); j++ {
				if sortedKeys[i].String() > sortedKeys[j].String() {
					sortedKeys[i], sortedKeys[j] = sortedKeys[j], sortedKeys[i]
				}
			}
		}
		for _, key := range sortedKeys {
			value := v.MapIndex(key).Interface()
			// Skip zero time.Time values to allow database DEFAULT values
			if t, ok := value.(time.Time); ok && t.IsZero() {
				continue
			}
			columns = append(columns, key.String())
			values = append(values, value)
		}
		if values == nil {
			values = []any{}
		}
		return columns, values, nil
	}

	// Handle struct using reflection
	if v.Kind() == reflect.Struct {
		cols, vals := b.extractStructColumnsAndValues(v)
		if vals == nil {
			vals = []any{}
		}
		return cols, vals, nil
	}

	return nil, []any{}, fmt.Errorf("unsupported value type for INSERT: %T", value)
}

func (b *Builder) extractStructColumnsAndValues(v reflect.Value) ([]string, []any) {
	var columns []string
	var values []any

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		if field.Anonymous && fieldValue.Kind() == reflect.Struct {
			cols, vals := b.extractStructColumnsAndValues(fieldValue)
			columns = append(columns, cols...)
			values = append(values, vals...)
			continue
		}

		// Skip unexported fields
		if !fieldValue.CanInterface() {
			continue
		}

		// Get column name from tag or field name
		columnName := structFieldColumnName(field)
		if columnName == "" {
			continue
		}

		// Skip slice/struct fields that are not handled as basic types
		if (fieldValue.Kind() == reflect.Slice || fieldValue.Kind() == reflect.Struct || fieldValue.Kind() == reflect.Ptr) &&
			fieldValue.Type() != reflect.TypeOf(time.Time{}) {
			// Special case: if it's a pointer to a basic type, we might want it, but for associations we skip
			if fieldValue.Kind() == reflect.Ptr && !fieldValue.IsNil() {
				elem := fieldValue.Elem()
				if elem.Kind() == reflect.Struct {
					continue
				}
			} else if fieldValue.Kind() == reflect.Slice || fieldValue.Kind() == reflect.Struct {
				continue
			} else if fieldValue.Kind() == reflect.Ptr {
				// Skip nil pointers except for deleted_at (soft delete)
				if columnName != "deleted_at" {
					continue
				}
			}
		}

		// Skip omitted columns
		omitted := false
		for _, omit := range b.query.omitColumns {
			if omit == columnName {
				omitted = true
				break
			}
		}
		if omitted {
			continue
		}

		// Skip zero values except for boolean and deleted_at (soft delete)
		// For deleted_at (nil pointer), we want to include it as NULL in INSERT
		// Skip zero time.Time values to allow database DEFAULT values
		if fieldValue.IsZero() && fieldValue.Kind() != reflect.Bool && !(columnName == "deleted_at" && fieldValue.Kind() == reflect.Ptr) {
			continue
		}

		columns = append(columns, columnName)
		values = append(values, fieldValue.Interface())
	}
	if values == nil {
		values = []any{}
	}
	return columns, values
}

// extractColumnNames extracts column names from a struct without checking values.
// This is used for SELECT clause generation where we want all columns regardless of their values.
func (b *Builder) extractColumnNames(value any) []string {
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	return b.extractStructColumnNames(v)
}

func (b *Builder) extractStructColumnNames(v reflect.Value) []string {
	var columns []string

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		if field.Anonymous && fieldValue.Kind() == reflect.Struct {
			columns = append(columns, b.extractStructColumnNames(fieldValue)...)
			continue
		}

		// Skip unexported fields
		if !fieldValue.CanInterface() {
			continue
		}

		// Get column name from tag or field name
		columnName := structFieldColumnName(field)
		if columnName == "" {
			continue
		}

		// Skip slice/struct fields that are not handled as basic types
		// But allow pointers to basic types or time.Time
		fieldType := field.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		if (fieldType.Kind() == reflect.Slice || fieldType.Kind() == reflect.Struct) &&
			fieldType != reflect.TypeOf(time.Time{}) {
			continue
		}

		columns = append(columns, columnName)
	}

	return columns
}
