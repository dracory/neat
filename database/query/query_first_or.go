package query

import (
	"reflect"
	"strings"
)

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
	if attrValue.Kind() == reflect.Ptr {
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
	if destValue.Kind() == reflect.Ptr {
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
	if attrValue.Kind() == reflect.Ptr {
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

// FirstOr retrieves the first record or executes a callback if not found.
func (q *Query) FirstOr(dest any, callback func() error) error {
	err := q.First(dest)
	if err != nil {
		return callback()
	}
	return nil
}

// FirstOrCreate retrieves the first record or creates it if not found.
func (q *Query) FirstOrCreate(dest any, conds ...any) error {
	// Try to find the record first
	err := q.First(dest)
	if err == nil {
		return nil // Record exists
	}

	// Record doesn't exist, create it
	return q.Create(dest)
}

// FirstOrNew retrieves the first record or prepares a new instance if not found.
func (q *Query) FirstOrNew(dest any, attributes any, values ...any) error {
	// Clone the query to avoid modifying the original
	query := q.Clone().(*Query)

	// Apply attributes as WHERE conditions
	if attributes != nil {
		if err := applyWhereConditions(query, attributes); err != nil {
			return err
		}
	}

	// Try to find the record first
	err := query.First(dest)
	if err == nil {
		return nil // Record exists
	}

	// Record doesn't exist, prepare new instance (without saving)
	// Apply attributes to the destination
	if attributes != nil {
		if err := applyAttributes(dest, attributes); err != nil {
			return err
		}
	}

	// Apply values if provided
	if len(values) > 0 && values[0] != nil {
		if err := applyAttributes(dest, values[0]); err != nil {
			return err
		}
	}

	return nil
}

// UpdateOrCreate updates a record if it exists, or creates it if it doesn't.
func (q *Query) UpdateOrCreate(dest any, attributes any, values any) error {
	// Try to find the record first
	err := q.First(dest)
	if err == nil {
		// Record exists, update it
		return q.Save(values)
	}

	// Record doesn't exist, create it
	return q.Create(values)
}
