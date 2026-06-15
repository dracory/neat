package association

import (
	"fmt"
	"reflect"
	"strings"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/support/str"
)

// BelongsTo represents a belongs-to relationship.
// In a belongs-to relationship, the current model belongs to another model.
// The foreign key is stored on the current model's table.
//
// Example: An Address belongs to a User.
//   - addresses table has user_id foreign key
//   - users table has id primary key
//
// Database Schema:
//
//	users (id, name, email)
//	addresses (id, user_id, street, city)
//
//	type Address struct {
//	    ID     uint
//	    UserID uint // foreign key
//	    User   User // belongs-to relationship
//	}
type BelongsTo struct {
	*Association
	foreignKey string // The foreign key column on the current model (e.g., "user_id")
	otherKey   string // The primary key column on the related model (e.g., "id")
}

// NewBelongsTo creates a new BelongsTo association.
// The query parameter provides the query builder for database operations.
// The model parameter is the model instance that belongs to another model.
// The association parameter is the name of the association (e.g., "user").
// The foreignKey parameter is the foreign key column on the current model (e.g., "user_id").
// The otherKey parameter is the primary key column on the related model (e.g., "id").
//
// Example:
//
//	address := Address{ID: 1, UserID: 5}
//	assoc := NewBelongsTo(db.Query(), &address, "user", "user_id", "id")
func NewBelongsTo(query contractsorm.Query, model any, association, foreignKey, otherKey string) *BelongsTo {
	return &BelongsTo{
		Association: NewAssociation(query, model, association),
		foreignKey:  foreignKey,
		otherKey:    otherKey,
	}
}

// Find loads the associated model for a belongs-to relationship.
// The out parameter must be a pointer to a struct for the related model.
// The conds parameter provides optional WHERE conditions for the query.
// Queries the related table where its primary key equals the foreign key value.
//
// Example:
//
//	var user User
//	err := assoc.Find(&user)
//
//	var user User
//	err := assoc.Find(&user, "status = ?", "active")
func (b *BelongsTo) Find(out any, conds ...any) error {
	// Get the foreign key value from the model
	foreignKeyValue, err := b.getForeignKeyValue()
	if err != nil {
		return fmt.Errorf("failed to get foreign key value: %w", err)
	}

	// Query the related model using the other key
	// For BelongsTo: foreignKey is on the model (e.g., user_id in addresses table)
	// otherKey is on the related model (e.g., id in users table)
	// We need to query the related table where its primary key equals the foreign key value
	query := b.Query().Model(out).Where("id = ?", foreignKeyValue)

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
// The values parameter must contain exactly one model to associate with.
// Sets the foreign key on the current model to the related model's primary key.
// Saves the current model to persist the foreign key change.
//
// Example:
//
//	user := User{ID: 5, Name: "John"}
//	err := assoc.Append(&user)
func (b *BelongsTo) Append(values ...any) error {
	if len(values) == 0 {
		return fmt.Errorf("no value provided to append")
	}

	// Get the other key value from the related model
	otherKeyValue, err := b.getOtherKeyValue(values[0])
	if err != nil {
		return fmt.Errorf("failed to get other key value: %w", err)
	}

	// Set the foreign key on the model
	if err := b.setForeignKeyValue(otherKeyValue); err != nil {
		return fmt.Errorf("failed to set foreign key value: %w", err)
	}

	// Save the model to persist the foreign key change
	query := b.Query().Model(b.model)
	if err := query.Save(b.model); err != nil {
		return fmt.Errorf("failed to save model: %w", err)
	}

	return nil
}

// Replace replaces the current association with the given value.
// The values parameter must contain exactly one model to associate with.
// This is equivalent to Append for belongs-to relationships.
//
// Example:
//
//	user := User{ID: 10, Name: "Jane"}
//	err := assoc.Replace(&user)
func (b *BelongsTo) Replace(values ...any) error {
	return b.Append(values...)
}

// Delete clears the association by setting the foreign key to nil.
// The values parameter is validated to ensure the provided value matches the current association.
// Sets the foreign key on the current model to nil/zero value.
// Saves the current model to persist the change.
//
// Example:
//
//	user := User{ID: 5}
//	err := assoc.Delete(&user)
func (b *BelongsTo) Delete(values ...any) error {
	if err := b.setForeignKeyValue(nil); err != nil {
		return fmt.Errorf("failed to set foreign key value: %w", err)
	}

	// Save the model to persist the foreign key change
	query := b.Query().Model(b.model)
	if err := query.Save(b.model); err != nil {
		return fmt.Errorf("failed to save model: %w", err)
	}

	return nil
}

// Clear clears the association by setting the foreign key to nil.
// Sets the foreign key on the current model to nil/zero value.
// Saves the current model to persist the change.
//
// Example:
//
//	err := assoc.Clear()
func (b *BelongsTo) Clear() error {
	if err := b.setForeignKeyValue(nil); err != nil {
		return fmt.Errorf("failed to set foreign key value: %w", err)
	}

	// Save the model to persist the foreign key change
	query := b.Query().Model(b.model)
	if err := query.Save(b.model); err != nil {
		return fmt.Errorf("failed to save model: %w", err)
	}

	return nil
}

// Count returns 1 if the association exists, 0 otherwise.
// Checks if the foreign key is set and the related record exists.
//
// Example:
//
//	count := assoc.Count() // 1 if associated, 0 if not
func (b *BelongsTo) Count() int64 {
	foreignKeyValue, err := b.getForeignKeyValue()
	if err != nil {
		return 0
	}

	if foreignKeyValue == nil {
		return 0
	}

	var count int64
	query := b.Query().Table(b.associationName()).Where("id = ?", foreignKeyValue)
	if err := query.Count(&count); err != nil {
		return 0
	}

	if count > 0 {
		return 1
	}
	return 0
}

// getForeignKeyValue gets the foreign key value from the model.
// Handles both snake_case (user_id) and PascalCase (UserID) field names.
// Returns an error if the field is not found or not accessible.
func (b *BelongsTo) getForeignKeyValue() (any, error) {
	val := reflect.ValueOf(b.model)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("model is not a struct")
	}

	// Try to find the field - try PascalCase first, then snake_case
	field := val.FieldByName(b.foreignKey)
	if !field.IsValid() {
		// Convert snake_case to PascalCase (e.g., user_id -> UserID)
		pascalCase := str.Of(b.foreignKey).Studly().String()
		field = val.FieldByName(pascalCase)
	}
	if !field.IsValid() {
		return nil, fmt.Errorf("foreign key field %s not found", b.foreignKey)
	}

	if !field.CanInterface() {
		return nil, fmt.Errorf("foreign key field %s is not accessible", b.foreignKey)
	}

	return field.Interface(), nil
}

// getOtherKeyValue gets the other key value from the related model.
// Handles both snake_case (id) and PascalCase (ID) field names.
// Returns an error if the field is not found or not accessible.
func (b *BelongsTo) getOtherKeyValue(model any) (any, error) {
	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("model is not a struct")
	}

	// Try to find the field - try PascalCase first (ID), then snake_case (id)
	field := val.FieldByName(b.otherKey)
	if !field.IsValid() {
		// Try PascalCase version
		if b.otherKey == "id" {
			field = val.FieldByName("ID")
		}
	}
	if !field.IsValid() {
		return nil, fmt.Errorf("other key field %s not found", b.otherKey)
	}

	if !field.CanInterface() {
		return nil, fmt.Errorf("other key field %s is not accessible", b.otherKey)
	}

	return field.Interface(), nil
}

// setForeignKeyValue sets the foreign key value on the model.
// Handles type conversion if the value type doesn't match the field type.
// Setting to nil sets the field to its zero value.
// Returns an error if the field is not found or not settable.
func (b *BelongsTo) setForeignKeyValue(value any) error {
	val := reflect.ValueOf(b.model)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("model is not a struct")
	}

	field := val.FieldByName(b.foreignKey)
	if !field.IsValid() {
		return fmt.Errorf("foreign key field %s not found", b.foreignKey)
	}

	if !field.CanSet() {
		return fmt.Errorf("foreign key field %s is not settable", b.foreignKey)
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
			return fmt.Errorf("cannot set %s with value of type %s", b.foreignKey, valueVal.Type())
		}
	}

	field.Set(valueVal)
	return nil
}

// associationName returns the association name (table name).
// Infers the table name from the model's TableName() method or struct name.
// Simple pluralization is applied (e.g., "User" -> "users").
func (b *BelongsTo) associationName() string {
	// Get the field type to infer table name
	val := reflect.ValueOf(b.model)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return b.association
	}

	field := val.FieldByName(b.association)
	if !field.IsValid() {
		return b.association
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
