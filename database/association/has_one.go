package association

import (
	"fmt"
	"reflect"
	"strings"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/support/str"
)

// HasOne represents a has-one relationship.
type HasOne struct {
	*Association
	foreignKey string
	localKey   string
}

// NewHasOne creates a new HasOne association.
func NewHasOne(query contractsorm.Query, model any, association, foreignKey, localKey string) *HasOne {
	return &HasOne{
		Association: NewAssociation(query, model, association),
		foreignKey:  foreignKey,
		localKey:    localKey,
	}
}

// Find loads the associated model for a has-one relationship.
func (h *HasOne) Find(out any, conds ...any) error {
	// Get the local key value from the model
	localKeyValue, err := h.getLocalKeyValue()
	if err != nil {
		return fmt.Errorf("failed to get local key value: %w", err)
	}

	// Query the related model using the foreign key
	query := h.Query().Model(out).Where(h.foreignKey+" = ?", localKeyValue)

	// Apply additional conditions if provided
	if len(conds) > 0 {
		// Handle conditions as (string, ...any) format
		if str, ok := conds[0].(string); ok && len(conds) > 1 {
			query = query.Where(str, conds[1:]...)
		} else {
			// Handle as individual conditions
			for _, cond := range conds {
				query = query.Where(cond)
			}
		}
	}

	// Execute the query
	return query.First(out)
}

// Append sets the foreign key to associate the model.
func (h *HasOne) Append(values ...any) error {
	if len(values) == 0 {
		return fmt.Errorf("no value provided to append")
	}

	// Get the local key value from the model
	localKeyValue, err := h.getLocalKeyValue()
	if err != nil {
		return fmt.Errorf("failed to get local key value: %w", err)
	}

	// Set the foreign key on the related model
	if err := h.setForeignKeyValue(values[0], localKeyValue); err != nil {
		return fmt.Errorf("failed to set foreign key on model: %w", err)
	}

	// Save the related model
	query := h.Query().Model(values[0])
	if err := query.Save(values[0]); err != nil {
		return fmt.Errorf("failed to save related model: %w", err)
	}

	return nil
}

// Replace replaces the current association with the given value.
func (h *HasOne) Replace(values ...any) error {
	// First, clear the current association by setting foreign key to null
	if err := h.Clear(); err != nil {
		return fmt.Errorf("failed to clear association: %w", err)
	}

	// Then, append the new value
	return h.Append(values...)
}

// Delete removes the given value from the association.
func (h *HasOne) Delete(values ...any) error {
	if len(values) == 0 {
		return fmt.Errorf("no value provided to delete")
	}

	// Set the foreign key to null for the related model using direct SQL update
	value := values[0]
	val := reflect.ValueOf(value)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	idField := val.FieldByName("ID")
	if !idField.IsValid() {
		return fmt.Errorf("could not find ID field on model")
	}
	modelID := idField.Interface()

	// Directly update the foreign key to NULL in the database
	query := h.Query().Table(h.associationName()).Where("id = ?", modelID)
	_, err := query.Update(h.foreignKey, nil)
	if err != nil {
		return fmt.Errorf("failed to clear foreign key on model: %w", err)
	}

	return nil
}

// Clear clears the association by setting the foreign key to null.
func (h *HasOne) Clear() error {
	// Get the local key value from the model
	localKeyValue, err := h.getLocalKeyValue()
	if err != nil {
		return fmt.Errorf("failed to get local key value: %w", err)
	}

	// Update the related model to set foreign key to null
	query := h.Query().Table(h.associationName()).Where(h.foreignKey+" = ?", localKeyValue)
	_, err = query.Update(h.foreignKey, nil)
	if err != nil {
		return fmt.Errorf("failed to clear association: %w", err)
	}

	return nil
}

// Count returns 1 if the association exists, 0 otherwise.
func (h *HasOne) Count() int64 {
	localKeyValue, err := h.getLocalKeyValue()
	if err != nil {
		return 0
	}

	var count int64
	query := h.Query().Table(h.associationName()).Where(h.foreignKey+" = ?", localKeyValue)
	if err := query.Count(&count); err != nil {
		return 0
	}

	if count > 0 {
		return 1
	}
	return 0
}

// getLocalKeyValue gets the local key value from the model.
func (h *HasOne) getLocalKeyValue() (any, error) {
	val := reflect.ValueOf(h.model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("model is not a struct")
	}

	// Try to find the field - try PascalCase first (ID), then snake_case (id)
	field := val.FieldByName(h.localKey)
	if !field.IsValid() {
		// Try PascalCase version
		if h.localKey == "id" {
			field = val.FieldByName("ID")
		}
	}
	if !field.IsValid() {
		return nil, fmt.Errorf("local key field %s not found", h.localKey)
	}

	if !field.CanInterface() {
		return nil, fmt.Errorf("local key field %s is not accessible", h.localKey)
	}

	return field.Interface(), nil
}

// setForeignKeyValue sets the foreign key value on a related model.
func (h *HasOne) setForeignKeyValue(model any, value any) error {
	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("model is not a struct")
	}

	// Try to find the field - try PascalCase first, then snake_case
	field := val.FieldByName(h.foreignKey)
	if !field.IsValid() {
		// Convert snake_case to PascalCase (e.g., user_id -> UserID)
		// First try Studly which gives UserId, then try with uppercase ID
		pascalCase := str.Of(h.foreignKey).Studly().String()
		field = val.FieldByName(pascalCase)
		// Also try with uppercase ID (e.g., UserID instead of UserId)
		if !field.IsValid() && strings.HasSuffix(pascalCase, "Id") {
			pascalCase = strings.TrimSuffix(pascalCase, "Id") + "ID"
			field = val.FieldByName(pascalCase)
		}
		// Also try direct conversion: split by underscore, capitalize each part
		if !field.IsValid() && strings.Contains(h.foreignKey, "_") {
			parts := strings.Split(h.foreignKey, "_")
			for i, part := range parts {
				if i == len(parts)-1 && part == "id" {
					parts[i] = "ID"
				} else if len(part) > 0 {
					// Capitalize first letter
					parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
				}
			}
			pascalCase = strings.Join(parts, "")
			field = val.FieldByName(pascalCase)
		}
	}
	if !field.IsValid() {
		return fmt.Errorf("foreign key field %s not found", h.foreignKey)
	}

	if !field.CanSet() {
		return fmt.Errorf("foreign key field %s is not settable", h.foreignKey)
	}

	valueVal := reflect.ValueOf(value)
	if !valueVal.IsValid() {
		// Setting to nil - set zero value for the field type
		field.Set(reflect.Zero(field.Type()))
		return nil
	}

	if valueVal.Type() != field.Type() {
		// Try to convert the value
		if valueVal.Type().ConvertibleTo(field.Type()) {
			valueVal = valueVal.Convert(field.Type())
		} else {
			return fmt.Errorf("cannot set %s with value of type %s", h.foreignKey, valueVal.Type())
		}
	}

	field.Set(valueVal)
	return nil
}

// associationName returns the association name (table name).
func (h *HasOne) associationName() string {
	// Get the field type to infer table name
	val := reflect.ValueOf(h.model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return h.association
	}

	field := val.FieldByName(h.association)
	if !field.IsValid() {
		return h.association
	}

	relationType := field.Type()
	if relationType.Kind() == reflect.Ptr {
		relationType = relationType.Elem()
	}

	// Check if the model has a TableName() method by creating a zero value
	if relationType.Kind() == reflect.Struct {
		zeroValue := reflect.New(relationType).Interface()
		if tabler, ok := zeroValue.(interface{ TableName() string }); ok {
			return tabler.TableName()
		}
	}

	// Infer table name from relation type
	tableName := str.Of(relationType.Name()).Snake().String()
	// Simple pluralization: add 's' if not already ending with 's'
	if !strings.HasSuffix(tableName, "s") {
		tableName = tableName + "s"
	}
	return tableName
}
