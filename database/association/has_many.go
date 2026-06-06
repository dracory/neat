package association

import (
	"fmt"
	"reflect"
	"strings"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/support/str"
)

// HasMany represents a has-many relationship.
// In a has-many relationship, the current model has many related models.
// The foreign key is stored on the related model's table.
//
// Example: A User has many Posts.
//   - users table has id primary key
//   - posts table has user_id foreign key
//
// Database Schema:
//
//	users (id, name, email)
//	posts (id, user_id, title, content)
//
//	type User struct {
//	    ID    uint
//	    Name  string
//	    Posts []Post // has-many relationship
//	}
type HasMany struct {
	*Association
	foreignKey string // The foreign key column on the related model (e.g., "user_id")
	localKey   string // The primary key column on the current model (e.g., "id")
}

// NewHasMany creates a new HasMany association.
// The query parameter provides the query builder for database operations.
// The model parameter is the model instance that has many related models.
// The association parameter is the name of the association (e.g., "posts").
// The foreignKey parameter is the foreign key column on the related model (e.g., "user_id").
// The localKey parameter is the primary key column on the current model (e.g., "id").
//
// Example:
//
//	user := User{ID: 1, Name: "John"}
//	assoc := NewHasMany(db.Query(), &user, "posts", "user_id", "id")
func NewHasMany(query contractsorm.Query, model any, association, foreignKey, localKey string) *HasMany {
	return &HasMany{
		Association: NewAssociation(query, model, association),
		foreignKey:  foreignKey,
		localKey:    localKey,
	}
}

// Find loads the associated models for a has-many relationship.
// The out parameter must be a pointer to a slice for the related models.
// The conds parameter provides optional WHERE conditions for the query.
// Queries the related table where the foreign key equals the local key value.
// Only includes records where the foreign key is not null.
//
// Example:
//
//	var posts []Post
//	err := assoc.Find(&posts)
//
//	var posts []Post
//	err := assoc.Find(&posts, "published = ?", true)
func (h *HasMany) Find(out any, conds ...any) error {
	// Get the local key value from the model
	localKeyValue, err := h.getLocalKeyValue()
	if err != nil {
		return fmt.Errorf("failed to get local key value: %w", err)
	}

	// Query the related models using the foreign key
	// Only include records where the foreign key is not null
	query := h.Query().Model(out).Where(h.foreignKey+" = ?", localKeyValue).WhereNotNull(h.foreignKey)

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
	return query.Get(out)
}

// Append appends models to the association.
// The values parameter provides the model(s) to append to the association.
// Sets the foreign key on each related model to the current model's local key value.
// Saves each related model to persist the foreign key change.
//
// Example:
//
//	post1 := Post{Title: "First Post"}
//	post2 := Post{Title: "Second Post"}
//	err := assoc.Append(&post1, &post2)
func (h *HasMany) Append(values ...any) error {
	if len(values) == 0 {
		return fmt.Errorf("no values provided to append")
	}

	// Get the local key value from the model
	localKeyValue, err := h.getLocalKeyValue()
	if err != nil {
		return fmt.Errorf("failed to get local key value: %w", err)
	}

	// Set the foreign key on each related model
	for _, value := range values {
		if err := h.setForeignKeyValue(value, localKeyValue); err != nil {
			return fmt.Errorf("failed to set foreign key on model: %w", err)
		}

		// Save the related model
		query := h.Query().Model(value)
		if err := query.Save(value); err != nil {
			return fmt.Errorf("failed to save related model: %w", err)
		}
	}

	return nil
}

// Replace replaces the current association with the given values.
// The values parameter provides the new model(s) for the association.
// First clears the current association by setting foreign keys to null.
// Then appends the new values.
//
// Example:
//
//	posts := []Post{{Title: "New Post"}, {Title: "Another Post"}}
//	err := assoc.Replace(posts...)
func (h *HasMany) Replace(values ...any) error {
	// First, clear the current association by setting foreign key to null
	if err := h.Clear(); err != nil {
		return fmt.Errorf("failed to clear association: %w", err)
	}

	// Then, append the new values
	return h.Append(values...)
}

// Delete removes the given values from the association.
// The values parameter provides the model(s) to remove from the association.
// Sets the foreign key to null for each related model using direct SQL update.
// Ensures the model belongs to this association by checking the foreign key.
//
// Example:
//
//	post := Post{ID: 1}
//	err := assoc.Delete(&post)
func (h *HasMany) Delete(values ...any) error {
	if len(values) == 0 {
		return fmt.Errorf("no values provided to delete")
	}

	// Get the local key value to ensure we only delete records that belong to this association
	localKeyValue, err := h.getLocalKeyValue()
	if err != nil {
		return fmt.Errorf("failed to get local key value: %w", err)
	}

	// Set the foreign key to null for each related model using direct SQL update
	for _, value := range values {
		// Get the ID of the model to update
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
		// Also ensure the model belongs to this association by checking the foreign key
		// Note: h.foreignKey is from model definition, not user input, so concatenation is safe
		// Validate foreignKey is a valid SQL identifier to prevent SQL injection
		if !isValidIdentifier(h.foreignKey) {
			return fmt.Errorf("invalid foreign key identifier: %s", h.foreignKey)
		}
		query := h.Query().Table(h.associationName()).Where("id = ? AND "+h.foreignKey+" = ?", modelID, localKeyValue)
		_, err := query.Update(h.foreignKey, nil)
		if err != nil {
			return fmt.Errorf("failed to clear foreign key on model: %w", err)
		}
	}

	return nil
}

// Clear clears the association by setting the foreign key to null for all related models.
// Updates all related models to set foreign key to null.
//
// Example:
//
//	err := assoc.Clear()
func (h *HasMany) Clear() error {
	// Get the local key value from the model
	localKeyValue, err := h.getLocalKeyValue()
	if err != nil {
		return fmt.Errorf("failed to get local key value: %w", err)
	}

	// Update all related models to set foreign key to null
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

// Count returns the number of records in the association.
// Counts records where the foreign key equals the local key value.
//
// Example:
//
//	count := assoc.Count()
func (h *HasMany) Count() int64 {
	localKeyValue, err := h.getLocalKeyValue()
	if err != nil {
		return 0
	}

	var count int64
	query := h.Query().Table(h.associationName()).Where(h.foreignKey+" = ?", localKeyValue)
	if err := query.Count(&count); err != nil {
		return 0
	}

	return count
}

// getLocalKeyValue gets the local key value from the model.
// Handles both snake_case (id) and PascalCase (ID) field names.
// Returns an error if the field is not found or not accessible.
func (h *HasMany) getLocalKeyValue() (any, error) {
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
// Handles both snake_case (user_id) and PascalCase (UserID) field names.
// Handles type conversion if the value type doesn't match the field type.
// Setting to nil sets the field to its zero value.
// Returns an error if the field is not found or not settable.
func (h *HasMany) setForeignKeyValue(model any, value any) error {
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
// Infers the table name from the model's TableName() method or struct name.
// Simple pluralization is applied (e.g., "Post" -> "posts").
func (h *HasMany) associationName() string {
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
	if relationType.Kind() == reflect.Slice {
		relationType = relationType.Elem()
	}
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
