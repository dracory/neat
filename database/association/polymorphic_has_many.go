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
// Example: A Post can have many Comments, and a Video can also have many Comments.
type PolymorphicHasMany struct {
	*Association
	polymorphicID   string
	polymorphicType string
	localKey        string
}

// NewPolymorphicHasMany creates a new PolymorphicHasMany association.
func NewPolymorphicHasMany(query contractsorm.Query, model any, association, polymorphicID, polymorphicType, localKey string) *PolymorphicHasMany {
	return &PolymorphicHasMany{
		Association:     NewAssociation(query, model, association),
		polymorphicID:   polymorphicID,
		polymorphicType: polymorphicType,
		localKey:        localKey,
	}
}

// Find loads the associated models for a polymorphic has-many relationship.
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

	// Set the polymorphic fields on each related model
	for _, value := range values {
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
func (p *PolymorphicHasMany) Replace(values ...any) error {
	// First, clear the current association
	if err := p.Clear(); err != nil {
		return fmt.Errorf("failed to clear association: %w", err)
	}

	// Then, append the new values
	return p.Append(values...)
}

// Delete removes the given values from the association.
func (p *PolymorphicHasMany) Delete(values ...any) error {
	if len(values) == 0 {
		return fmt.Errorf("no values provided to delete")
	}

	// Set the polymorphic fields to null for each related model using direct SQL update
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
		_, err = query.Update(updates)
		if err != nil {
			return fmt.Errorf("failed to clear polymorphic fields on model: %w", err)
		}
	}

	return nil
}

// Clear clears the association by setting the polymorphic fields to null for all related models.
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
func (p *PolymorphicHasMany) getLocalKeyValue() (any, error) {
	val := reflect.ValueOf(p.model)
	if val.Kind() == reflect.Ptr {
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
func (p *PolymorphicHasMany) getModelTypeName() string {
	val := reflect.ValueOf(p.model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return ""
	}

	return val.Type().Name()
}

// setPolymorphicIDValue sets the polymorphic ID value on a related model.
func (p *PolymorphicHasMany) setPolymorphicIDValue(model any, value any) error {
	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("model is not a struct")
	}

	// Try to find the field by db tag first, then by name
	var field reflect.Value
	found := false

	for i := 0; i < val.NumField(); i++ {
		structField := val.Type().Field(i)
		dbTag := structField.Tag.Get("db")
		if dbTag == p.polymorphicID {
			field = val.Field(i)
			found = true
			break
		}
	}

	if !found {
		// Try PascalCase version
		pascalCase := str.Of(p.polymorphicID).Studly().String()
		field = val.FieldByName(pascalCase)
		// Also try with uppercase ID
		if !field.IsValid() && strings.HasSuffix(pascalCase, "Id") {
			pascalCase = strings.TrimSuffix(pascalCase, "Id") + "ID"
			field = val.FieldByName(pascalCase)
		}
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
func (p *PolymorphicHasMany) setPolymorphicTypeValue(model any, value any) error {
	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("model is not a struct")
	}

	// Try to find the field by db tag first, then by name
	var field reflect.Value
	found := false

	for i := 0; i < val.NumField(); i++ {
		structField := val.Type().Field(i)
		dbTag := structField.Tag.Get("db")
		if dbTag == p.polymorphicType {
			field = val.Field(i)
			found = true
			break
		}
	}

	if !found {
		// Try PascalCase version
		pascalCase := str.Of(p.polymorphicType).Studly().String()
		field = val.FieldByName(pascalCase)
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
func (p *PolymorphicHasMany) associationName() string {
	// Try to get the field type from the model to infer table name
	val := reflect.ValueOf(p.model)
	if val.Kind() == reflect.Ptr {
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
