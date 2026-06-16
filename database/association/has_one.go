package association

import (
	"fmt"
	"reflect"
	"strings"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/support/str"
)

// HasOne represents a has-one relationship.
// In a has-one relationship, the current model has one related model.
// The foreign key is stored on the related model's table.
// Similar to has-many, but only one related record is expected.
//
// Example: A User has one Profile.
//   - users table has id primary key
//   - profiles table has user_id foreign key (unique)
//
// Database Schema:
//
//	users (id, name, email)
//	profiles (id, user_id, bio, avatar)
//
//	type User struct {
//	    ID      uint
//	    Name    string
//	    Profile *Profile // has-one relationship
//	}
type HasOne struct {
	*Association
	foreignKey string // The foreign key column on the related model (e.g., "user_id")
	localKey   string // The primary key column on the current model (e.g., "id")
}

// NewHasOne creates a new HasOne association.
// The query parameter provides the query builder for database operations.
// The model parameter is the model instance that has one related model.
// The association parameter is the name of the association (e.g., "profile").
// The foreignKey parameter is the foreign key column on the related model (e.g., "user_id").
// The localKey parameter is the primary key column on the current model (e.g., "id").
//
// Example:
//
//	user := User{ID: 1, Name: "John"}
//	assoc := NewHasOne(db.Query(), &user, "profile", "user_id", "id")
func NewHasOne(query contractsorm.Query, model any, association, foreignKey, localKey string) *HasOne {
	return &HasOne{
		Association: NewAssociation(query, model, association),
		foreignKey:  foreignKey,
		localKey:    localKey,
	}
}

// Find loads the associated model for a has-one relationship.
// The out parameter must be a pointer to a struct for the related model.
// The conds parameter provides optional WHERE conditions for the query.
// Queries the related table where the foreign key equals the local key value.
// Only includes records where the foreign key is not null.
// Uses First() to retrieve a single record.
//
// Example:
//
//	var profile Profile
//	err := assoc.Find(&profile)
//
//	var profile Profile
//	err := assoc.Find(&profile, "active = ?", true)
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

// Append appends a model to the association.
// The values parameter must contain exactly one model to associate with.
// Sets the foreign key on the related model to the current model's local key value.
// Saves the related model to persist the foreign key change.
//
// Example:
//
//	profile := Profile{Bio: "Software Developer"}
//	err := assoc.Append(&profile)
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
// The values parameter must contain exactly one model to associate with.
// First clears the current association by setting the foreign key to null.
// Then appends the new value.
//
// Example:
//
//	profile := Profile{Bio: "Updated Bio"}
//	err := assoc.Replace(&profile)
func (h *HasOne) Replace(values ...any) error {
	// First, clear the current association by setting foreign key to null
	if err := h.Clear(); err != nil {
		return fmt.Errorf("failed to clear association: %w", err)
	}

	// Then, append the new value
	return h.Append(values...)
}

// Delete removes the given value from the association.
// The values parameter provides the model to remove from the association.
// Sets the foreign key to null for the related model using direct SQL update.
// Ensures the model belongs to this association by checking the foreign key.
//
// Example:
//
//	profile := Profile{ID: 1}
//	err := assoc.Delete(&profile)
func (h *HasOne) Delete(values ...any) error {
	if len(values) == 0 {
		return fmt.Errorf("no value provided to delete")
	}

	// Get the local key value to ensure we only delete records that belong to this association
	localKeyValue, err := h.getLocalKeyValue()
	if err != nil {
		return fmt.Errorf("failed to get local key value: %w", err)
	}

	// Set the foreign key to null for the related model using direct SQL update
	value := values[0]
	val := reflect.ValueOf(value)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	idField := val.FieldByName("ID")
	if !idField.IsValid() {
		return fmt.Errorf("could not find ID field on model")
	}
	modelID := idField.Interface()

	// Directly update the foreign key to NULL in the database
	// Also ensure the model belongs to this association by checking the foreign key
	// Note: h.foreignKey is from model definition, not user input, so concatenation is safe
	// Validate foreignKey is a valid SQL identifier to prevent SQL injection
	if !isValidIdentifier(h.foreignKey) {
		return fmt.Errorf("invalid foreign key identifier: %s", h.foreignKey)
	}
	query := h.Query().Table(h.associationName()).Where("id = ? AND "+h.foreignKey+" = ?", modelID, localKeyValue)
	_, err = query.Update(h.foreignKey, nil)
	if err != nil {
		return fmt.Errorf("failed to clear foreign key on model: %w", err)
	}

	return nil
}

// Clear clears the association by setting the foreign key to null.
// Updates the related model to set foreign key to null.
//
// Example:
//
//	err := assoc.Clear()
func (h *HasOne) Clear() error {
	// Get the local key value from the model
	localKeyValue, err := h.getLocalKeyValue()
	if err != nil {
		return fmt.Errorf("failed to get local key value: %w", err)
	}

	// Update the related model to set foreign key to null
	// Validate foreignKey is a valid SQL identifier to prevent SQL injection
	if !isValidIdentifier(h.foreignKey) {
		return fmt.Errorf("invalid foreign key identifier: %s", h.foreignKey)
	}
	query := h.Query().Table(h.associationName()).Where(h.foreignKey+" = ?", localKeyValue)
	_, err = query.Update(h.foreignKey, nil)
	if err != nil {
		return fmt.Errorf("failed to clear association: %w", err)
	}

	return nil
}

// Count returns 1 if the association exists, 0 otherwise.
// Counts records where the foreign key equals the local key value.
// Returns 1 if at least one record exists, 0 otherwise.
//
// Example:
//
//	count := assoc.Count() // 1 if associated, 0 if not
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
// Handles both snake_case (id) and PascalCase (ID) field names.
// Returns an error if the field is not found or not accessible.
func (h *HasOne) getLocalKeyValue() (any, error) {
	val := reflect.ValueOf(h.model)
	if val.Kind() == reflect.Pointer {
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
// Handles both snake_case (user_id) and PascalCase (UserID) field names.
// Handles type conversion if the value type doesn't match the field type.
// Setting to nil sets the field to its zero value.
// Returns an error if the field is not found or not settable.
func (h *HasOne) setForeignKeyValue(model any, value any) error {
	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Pointer {
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
// Infers the table name from the model's TableName() method or struct name.
// Simple pluralization is applied (e.g., "Profile" -> "profiles").
func (h *HasOne) associationName() string {
	// Get the field type to infer table name
	val := reflect.ValueOf(h.model)
	if val.Kind() == reflect.Pointer {
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
	if relationType.Kind() == reflect.Pointer {
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
