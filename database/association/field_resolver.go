package association

import (
	"reflect"
	"strings"

	"github.com/dracory/neat/support/str"
)

// getFieldByTagOrName finds a field in a struct by db tag first, then by name.
// It tries multiple variations: db tag, PascalCase, snake_case, and handles "Id" -> "ID" suffix.
func getFieldByTagOrName(val reflect.Value, fieldName string) (reflect.Value, error) {
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return reflect.Value{}, nil
	}

	// Try to find the field by db tag first
	for i := 0; i < val.NumField(); i++ {
		structField := val.Type().Field(i)
		dbTag := structField.Tag.Get("db")
		if dbTag == fieldName {
			return val.Field(i), nil
		}
	}

	// Try PascalCase version
	field := val.FieldByName(fieldName)
	if field.IsValid() {
		return field, nil
	}

	// Try snake_case to PascalCase conversion
	pascalCase := str.Of(fieldName).Studly().String()
	field = val.FieldByName(pascalCase)
	if field.IsValid() {
		return field, nil
	}

	// Try with uppercase ID suffix (e.g., "userid" -> "UserID")
	if strings.HasSuffix(pascalCase, "Id") {
		pascalCase = strings.TrimSuffix(pascalCase, "Id") + "ID"
		field = val.FieldByName(pascalCase)
		if field.IsValid() {
			return field, nil
		}
	}

	return reflect.Value{}, nil
}
