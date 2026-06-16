package association

import (
	"fmt"
	"reflect"
	"strings"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/support/str"
)

// PolymorphicHasMany represents a polymorphic has-many relationship.
// This allows a model to have many related models that can belong to multiple different model types.
// Uses polymorphic fields (ID and Type) on the related models to store the relationship.
//
// Example: A Post can have many Comments, and a Video can also have many Comments.
//   - comments table has commentable_id and commentable_type columns
//   - commentable_id stores the ID of the parent model
//   - commentable_type stores the type name (e.g., "Post", "Video")
//
// Database Schema:
//
//	posts (id, title, content)
//	videos (id, title, url)
//	comments (id, commentable_id, commentable_type, content)
//
//	type Post struct {
//	    ID       uint
//	    Title    string
//	    Comments []Comment // polymorphic has-many
//	}
type PolymorphicHasMany struct {
	*Association
	polymorphicID   string // The polymorphic ID field name on related models (e.g., "commentable_id")
	polymorphicType string // The polymorphic type field name on related models (e.g., "commentable_type")
	localKey        string // The primary key column on the current model (e.g., "id")
}

// NewPolymorphicHasMany creates a new PolymorphicHasMany association.
// The query parameter provides the query builder for database operations.
// The model parameter is the model instance that has many related models.
// The association parameter is the name of the association (e.g., "comments").
// The polymorphicID parameter is the polymorphic ID field name on related models (e.g., "commentable_id").
// The polymorphicType parameter is the polymorphic type field name on related models (e.g., "commentable_type").
// The localKey parameter is the primary key column on the current model (e.g., "id").
//
// Example:
//
//	post := Post{ID: 1, Title: "My Post"}
//	assoc := NewPolymorphicHasMany(db.Query(), &post, "comments", "commentable_id", "commentable_type", "id")
func NewPolymorphicHasMany(query contractsorm.Query, model any, association, polymorphicID, polymorphicType, localKey string) *PolymorphicHasMany {
	return &PolymorphicHasMany{
		Association:     NewAssociation(query, model, association),
		polymorphicID:   polymorphicID,
		polymorphicType: polymorphicType,
		localKey:        localKey,
	}
}

// Find loads the associated models for a polymorphic has-many relationship.
// The out parameter must be a pointer to a slice for the related models.
// The conds parameter provides optional WHERE conditions for the query.
// Uses the local key and model type name to query related models.
// Filters by both polymorphic ID and polymorphic type.
//
// Example:
//
//	var comments []Comment
//	err := assoc.Find(&comments)
//
//	var comments []Comment
//	err := assoc.Find(&comments, "approved = ?", true)
func (p *PolymorphicHasMany) Find(out any, conds ...any) error {
	// Get the local key value from the model
	localKeyValue, err := p.getLocalKeyValue()
	if err != nil {
		return fmt.Errorf("failed to get local key value: %w", err)
	}

	// Get the type name from the model
	typeName := p.getModelTypeName()
	if typeName == "" {
		return fmt.Errorf("could not determine model type name")
	}

	// Convert type name to polymorphic type value (e.g., "Post" -> "Post")
	polymorphicTypeValue := typeName

	// Query the related models using the polymorphic fields
	query := p.Query().Model(out).Table(p.associationName()).
		Where(p.polymorphicID+" = ?", localKeyValue).
		Where(p.polymorphicType+" = ?", polymorphicTypeValue)

	// Apply additional conditions if provided
	if len(conds) > 0 {
		if str, ok := conds[0].(string); ok && len(conds) > 1 {
			query = query.Where(str, conds[1:]...)
		} else {
			for _, cond := range conds {
				query = query.Where(cond)
			}
		}
	}

	// Execute the query
	return query.Get(out)
}

// Append appends models to the polymorphic association.
// The values parameter provides the model(s) to append to the association.
// Sets the polymorphic ID to the current model's local key value.
// Sets the polymorphic type to the current model's type name.
// Validates that the value type matches the expected association type.
// Saves each related model to persist the polymorphic fields.
//
// Example:
//
//	comment1 := Comment{Content: "Great post!"}
//	comment2 := Comment{Content: "Thanks for sharing"}
//	err := assoc.Append(&comment1, &comment2)
func (p *PolymorphicHasMany) Append(values ...any) error {
	if len(values) == 0 {
		return fmt.Errorf("no values provided to append")
	}

	// Get the local key value from the model
	localKeyValue, err := p.getLocalKeyValue()
	if err != nil {
		return fmt.Errorf("failed to get local key value: %w", err)
	}

	// Get the type name from the model
	typeName := p.getModelTypeName()
	if typeName == "" {
		return fmt.Errorf("could not determine model type name")
	}

	// Get the expected type from the association field
	val := reflect.ValueOf(p.model)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	field := val.FieldByName(p.association)
	if !field.IsValid() {
		return fmt.Errorf("could not find association field")
	}

	relationType := field.Type()
	if relationType.Kind() == reflect.Slice {
		relationType = relationType.Elem()
	}
	if relationType.Kind() == reflect.Pointer {
		relationType = relationType.Elem()
	}

	// Set the polymorphic fields on each related model
	for _, value := range values {
		// Validate that the value type matches the expected association type
		valueVal := reflect.ValueOf(value)
		if valueVal.Kind() == reflect.Pointer {
			valueVal = valueVal.Elem()
		}
		if valueVal.Type() != relationType {
			return fmt.Errorf("value type %v does not match expected association type %v", valueVal.Type(), relationType)
		}

		if err := p.setPolymorphicIDValue(value, localKeyValue); err != nil {
			return fmt.Errorf("failed to set polymorphic ID on model: %w", err)
		}

		if err := p.setPolymorphicTypeValue(value, typeName); err != nil {
			return fmt.Errorf("failed to set polymorphic type on model: %w", err)
		}

		// Save the related model
		query := p.Query().Model(value)
		if err := query.Save(value); err != nil {
			return fmt.Errorf("failed to save related model: %w", err)
		}
	}

	return nil
}

// Replace replaces the current association with the given values.
// The values parameter provides the new model(s) for the association.
// First clears the current association by setting polymorphic fields to null.
// Then appends the new values.
//
// Example:
//
//	comments := []Comment{{Content: "New comment"}, {Content: "Another comment"}}
//	err := assoc.Replace(comments...)
func (p *PolymorphicHasMany) Replace(values ...any) error {
	// First, clear the current association
	if err := p.Clear(); err != nil {
		return fmt.Errorf("failed to clear association: %w", err)
	}

	// Then, append the new values
	return p.Append(values...)
}

// Delete removes the given values from the association.
// The values parameter provides the model(s) to remove from the association.
// Sets the polymorphic fields to null for each related model using direct SQL update.
// Ensures the model belongs to this association by checking both polymorphic fields.
// Returns an error if the value was not part of the association.
//
// Example:
//
//	comment := Comment{ID: 1}
//	err := assoc.Delete(&comment)
func (p *PolymorphicHasMany) Delete(values ...any) error {
	if len(values) == 0 {
		return fmt.Errorf("no values provided to delete")
	}

	// Set the polymorphic fields to null for each related model using direct SQL update
	for _, value := range values {
		// Get the ID of the model to update
		val := reflect.ValueOf(value)
		if val.Kind() == reflect.Pointer {
			val = val.Elem()
		}
		idField := val.FieldByName("ID")
		if !idField.IsValid() {
			return fmt.Errorf("could not find ID field on model")
		}
		modelID := idField.Interface()

		// Directly update both polymorphic fields to NULL in a single UPDATE statement
		// Also ensure the model belongs to this association
		localKeyValue, err := p.getLocalKeyValue()
		if err != nil {
			return fmt.Errorf("failed to get local key value: %w", err)
		}
		typeName := p.getModelTypeName()
		if typeName == "" {
			return fmt.Errorf("could not determine model type name")
		}

		query := p.Query().Table(p.associationName()).
			Where("id = ? AND "+p.polymorphicID+" = ? AND "+p.polymorphicType+" = ?", modelID, localKeyValue, typeName)
		// Use a map to update multiple columns in a single call
		updates := map[string]any{
			p.polymorphicID:   nil,
			p.polymorphicType: nil,
		}
		result, err := query.Update(updates)
		if err != nil {
			return fmt.Errorf("failed to clear polymorphic fields on model: %w", err)
		}

		// Check if any rows were actually affected
		if result.RowsAffected == 0 {
			return fmt.Errorf("value was not part of the association (no rows affected)")
		}
	}

	return nil
}

// Clear clears the association by setting the polymorphic fields to null for all related models.
// Updates all related models to set polymorphic ID and type to null.
// Uses a single UPDATE statement for efficiency.
//
// Example:
//
//	err := assoc.Clear()
func (p *PolymorphicHasMany) Clear() error {
	// Get the local key value from the model
	localKeyValue, err := p.getLocalKeyValue()
	if err != nil {
		return fmt.Errorf("failed to get local key value: %w", err)
	}

	// Get the type name from the model
	typeName := p.getModelTypeName()
	if typeName == "" {
		return fmt.Errorf("could not determine model type name")
	}

	// Update all related models to set polymorphic fields to null in a single UPDATE statement
	query := p.Query().Table(p.associationName()).
		Where(p.polymorphicID+" = ?", localKeyValue).
		Where(p.polymorphicType+" = ?", typeName)
	// Use a map to update multiple columns in a single call
	updates := map[string]any{
		p.polymorphicID:   nil,
		p.polymorphicType: nil,
	}
	_, err = query.Update(updates)
	if err != nil {
		return fmt.Errorf("failed to clear polymorphic fields: %w", err)
	}

	return nil
}

// Count returns the number of records in the association.
// Counts records where both polymorphic ID and type match the current model.
//
// Example:
//
//	count := assoc.Count()
func (p *PolymorphicHasMany) Count() int64 {
	localKeyValue, err := p.getLocalKeyValue()
	if err != nil {
		return 0
	}

	typeName := p.getModelTypeName()
	if typeName == "" {
		return 0
	}

	var count int64
	query := p.Query().Table(p.associationName()).
		Where(p.polymorphicID+" = ?", localKeyValue).
		Where(p.polymorphicType+" = ?", typeName)

	if err := query.Count(&count); err != nil {
		return 0
	}

	return count
}

// getLocalKeyValue gets the local key value from the model.
// Handles both snake_case (id) and PascalCase (ID) field names.
// Returns an error if the field is not found or not accessible.
func (p *PolymorphicHasMany) getLocalKeyValue() (any, error) {
	val := reflect.ValueOf(p.model)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("model is not a struct")
	}

	// Try to find the field - try PascalCase first (ID), then snake_case (id)
	field := val.FieldByName(p.localKey)
	if !field.IsValid() {
		// Try PascalCase version
		if p.localKey == "id" {
			field = val.FieldByName("ID")
		}
	}
	if !field.IsValid() {
		return nil, fmt.Errorf("local key field %s not found", p.localKey)
	}

	if !field.CanInterface() {
		return nil, fmt.Errorf("local key field %s is not accessible", p.localKey)
	}

	return field.Interface(), nil
}

// getModelTypeName gets the type name of the model.
// Returns the struct type name (e.g., "Post", "Video").
// Returns empty string if the model is not a struct.
func (p *PolymorphicHasMany) getModelTypeName() string {
	val := reflect.ValueOf(p.model)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return ""
	}

	return val.Type().Name()
}

// setPolymorphicIDValue sets the polymorphic ID value on a related model.
// Uses getFieldByTagOrName to find the field by tag or name.
// Handles type conversion if the value type doesn't match the field type.
// Setting to nil sets the field to its zero value.
// Returns an error if the field is not found or not settable.
func (p *PolymorphicHasMany) setPolymorphicIDValue(model any, value any) error {
	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("model is not a struct")
	}

	field, err := getFieldByTagOrName(val, p.polymorphicID)
	if err != nil {
		return err
	}
	if !field.IsValid() {
		return fmt.Errorf("polymorphic ID field %s not found", p.polymorphicID)
	}

	if !field.CanSet() {
		return fmt.Errorf("polymorphic ID field %s is not settable", p.polymorphicID)
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
			return fmt.Errorf("cannot set %s with value of type %s", p.polymorphicID, valueVal.Type())
		}
	}

	field.Set(valueVal)
	return nil
}

// setPolymorphicTypeValue sets the polymorphic type value on a related model.
// Uses getFieldByTagOrName to find the field by tag or name.
// Handles type conversion if the value type doesn't match the field type.
// Setting to nil sets the field to its zero value.
// Returns an error if the field is not found or not settable.
func (p *PolymorphicHasMany) setPolymorphicTypeValue(model any, value any) error {
	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("model is not a struct")
	}

	field, err := getFieldByTagOrName(val, p.polymorphicType)
	if err != nil {
		return err
	}
	if !field.IsValid() {
		return fmt.Errorf("polymorphic type field %s not found", p.polymorphicType)
	}

	if !field.CanSet() {
		return fmt.Errorf("polymorphic type field %s is not settable", p.polymorphicType)
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
			return fmt.Errorf("cannot set %s with value of type %s", p.polymorphicType, valueVal.Type())
		}
	}

	field.Set(valueVal)
	return nil
}

// associationName returns the association name (table name).
// Infers the table name from the model's TableName() method or struct name.
// Simple pluralization is applied (e.g., "Comment" -> "comments").
func (p *PolymorphicHasMany) associationName() string {
	// Try to get the field type from the model to infer table name
	val := reflect.ValueOf(p.model)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		// Fallback to snake_case of association name
		tableName := str.Of(p.association).Snake().String()
		if !strings.HasSuffix(tableName, "s") {
			tableName = tableName + "s"
		}
		return tableName
	}

	field := val.FieldByName(p.association)
	if !field.IsValid() {
		// Fallback to snake_case of association name
		tableName := str.Of(p.association).Snake().String()
		if !strings.HasSuffix(tableName, "s") {
			tableName = tableName + "s"
		}
		return tableName
	}

	relationType := field.Type()
	if relationType.Kind() == reflect.Slice {
		relationType = relationType.Elem()
	}
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
