package association

import (
	"fmt"
	"reflect"
	"strings"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/support/str"
)

// PolymorphicBelongsTo represents a polymorphic belongs-to relationship.
// This allows a model to belong to multiple different model types.
// Example: A Comment can belong to a Post or a Video.
type PolymorphicBelongsTo struct {
	*Association
	polymorphicID   string
	polymorphicType string
}

// NewPolymorphicBelongsTo creates a new PolymorphicBelongsTo association.
func NewPolymorphicBelongsTo(query contractsorm.Query, model any, association, polymorphicID, polymorphicType string) *PolymorphicBelongsTo {
	return &PolymorphicBelongsTo{
		Association:     NewAssociation(query, model, association),
		polymorphicID:   polymorphicID,
		polymorphicType: polymorphicType,
	}
}

// Find loads the associated model for a polymorphic belongs-to relationship.
func (p *PolymorphicBelongsTo) Find(out any, conds ...any) error {
	// Get the polymorphic ID value
	polymorphicIDValue, err := p.getPolymorphicIDValue()
	if err != nil {
		return fmt.Errorf("failed to get polymorphic ID value: %w", err)
	}

	// Get the polymorphic type value
	polymorphicTypeValue, err := p.getPolymorphicTypeValue()
	if err != nil {
		return fmt.Errorf("failed to get polymorphic type value: %w", err)
	}

	if polymorphicIDValue == nil || polymorphicTypeValue == nil {
		return fmt.Errorf("polymorphic association is not set")
	}

	// Convert type string to table name (e.g., "Post" -> "posts")
	typeStr, ok := polymorphicTypeValue.(string)
	if !ok {
		return fmt.Errorf("polymorphic type value is not a string")
	}

	// Try to get table name from the model's TableName() method
	tableName := ""
	if tn, ok := out.(interface{ TableName() string }); ok {
		tableName = tn.TableName()
	}

	// Fallback to naive pluralization if TableName() not available
	if tableName == "" {
		tableName = str.Of(typeStr).Snake().String()
		if !strings.HasSuffix(tableName, "s") {
			tableName = tableName + "s"
		}
	}

	// Query the related model using the polymorphic ID
	query := p.Query().Model(out).Table(tableName).Where("id = ?", polymorphicIDValue)

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
	return query.First(out)
}

// Append sets the polymorphic fields to associate the model.
func (p *PolymorphicBelongsTo) Append(values ...any) error {
	if len(values) == 0 {
		return fmt.Errorf("no value provided to append")
	}

	if len(values) > 1 {
		return fmt.Errorf("BelongsTo relationship can only have one associated value, but %d values were provided", len(values))
	}

	// Get the ID from the related model
	relatedModel := values[0]
	val := reflect.ValueOf(relatedModel)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("related model is not a struct")
	}

	// Get the ID field
	idField := val.FieldByName("ID")
	if !idField.IsValid() {
		return fmt.Errorf("could not find ID field on related model")
	}
	relatedID := idField.Interface()

	// Validate that the ID is non-zero (model must be saved)
	if reflect.ValueOf(relatedID).IsZero() {
		return fmt.Errorf("related model must be saved before associating (ID is zero)")
	}

	// Get the type name from the related model
	typeName := val.Type().Name()
	if typeName == "" {
		return fmt.Errorf("could not determine type name of related model")
	}

	// Set the polymorphic ID
	if err := p.setPolymorphicIDValue(relatedID); err != nil {
		return fmt.Errorf("failed to set polymorphic ID value: %w", err)
	}

	// Set the polymorphic type
	if err := p.setPolymorphicTypeValue(typeName); err != nil {
		return fmt.Errorf("failed to set polymorphic type value: %w", err)
	}

	// Save the model to persist the polymorphic fields
	query := p.Query().Model(p.model)
	if err := query.Save(p.model); err != nil {
		return fmt.Errorf("failed to save model: %w", err)
	}

	return nil
}

// Replace replaces the current association with the given value.
func (p *PolymorphicBelongsTo) Replace(values ...any) error {
	return p.Append(values...)
}

// Delete clears the association by setting the polymorphic fields to nil.
// The values parameter is validated to ensure the provided value matches the current association.
func (p *PolymorphicBelongsTo) Delete(values ...any) error {
	// Validate that a value was provided
	if len(values) == 0 {
		return fmt.Errorf("no value provided to delete")
	}

	// Get current polymorphic ID and type values
	currentID, err := p.getPolymorphicIDValue()
	if err != nil {
		return fmt.Errorf("failed to get current polymorphic ID: %w", err)
	}

	currentType, err := p.getPolymorphicTypeValue()
	if err != nil {
		return fmt.Errorf("failed to get current polymorphic type: %w", err)
	}

	// If association is not set, nothing to delete
	if currentID == nil || currentType == nil {
		return nil
	}

	// Validate that the provided value matches the current association
	relatedModel := values[0]
	val := reflect.ValueOf(relatedModel)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("provided value is not a struct")
	}

	// Get the ID field from the provided value
	idField := val.FieldByName("ID")
	if !idField.IsValid() {
		return fmt.Errorf("could not find ID field on provided value")
	}
	providedID := idField.Interface()

	// Get the type name from the provided value
	providedType := val.Type().Name()

	// Convert current type to string for comparison
	currentTypeStr, ok := currentType.(string)
	if !ok {
		return fmt.Errorf("current polymorphic type is not a string")
	}

	// Validate that the provided value matches the current association
	if providedID != currentID || providedType != currentTypeStr {
		return fmt.Errorf("provided value does not match current association")
	}

	// Clear the association
	if err := p.setPolymorphicIDValue(nil); err != nil {
		return fmt.Errorf("failed to set polymorphic ID value: %w", err)
	}

	if err := p.setPolymorphicTypeValue(nil); err != nil {
		return fmt.Errorf("failed to set polymorphic type value: %w", err)
	}

	// Save the model to persist the changes
	query := p.Query().Model(p.model)
	if err := query.Save(p.model); err != nil {
		return fmt.Errorf("failed to save model: %w", err)
	}

	return nil
}

// Clear clears the association by setting the polymorphic fields to nil.
func (p *PolymorphicBelongsTo) Clear() error {
	if err := p.setPolymorphicIDValue(nil); err != nil {
		return fmt.Errorf("failed to set polymorphic ID value: %w", err)
	}

	if err := p.setPolymorphicTypeValue(nil); err != nil {
		return fmt.Errorf("failed to set polymorphic type value: %w", err)
	}

	// Save the model to persist the changes
	query := p.Query().Model(p.model)
	if err := query.Save(p.model); err != nil {
		return fmt.Errorf("failed to save model: %w", err)
	}

	return nil
}

// Count returns 1 if the association exists, 0 otherwise.
func (p *PolymorphicBelongsTo) Count() int64 {
	polymorphicIDValue, err := p.getPolymorphicIDValue()
	if err != nil {
		return 0
	}

	polymorphicTypeValue, err := p.getPolymorphicTypeValue()
	if err != nil {
		return 0
	}

	if polymorphicIDValue == nil || polymorphicTypeValue == nil {
		return 0
	}

	// Convert type string to table name
	typeStr, ok := polymorphicTypeValue.(string)
	if !ok {
		return 0
	}

	// Try to get table name from the model's TableName() method
	tableName := ""
	if tn, ok := p.model.(interface{ TableName() string }); ok {
		tableName = tn.TableName()
	}

	// Fallback to naive pluralization if TableName() not available
	if tableName == "" {
		tableName = str.Of(typeStr).Snake().String()
		if !strings.HasSuffix(tableName, "s") {
			tableName = tableName + "s"
		}
	}

	var count int64
	query := p.Query().Table(tableName).Where("id = ?", polymorphicIDValue)
	if err := query.Count(&count); err != nil {
		return 0
	}

	if count > 0 {
		return 1
	}
	return 0
}

// getPolymorphicIDValue gets the polymorphic ID value from the model.
func (p *PolymorphicBelongsTo) getPolymorphicIDValue() (any, error) {
	val := reflect.ValueOf(p.model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("model is not a struct")
	}

	field, err := getFieldByTagOrName(val, p.polymorphicID)
	if err != nil {
		return nil, err
	}
	if !field.IsValid() {
		return nil, fmt.Errorf("polymorphic ID field %s not found", p.polymorphicID)
	}

	if !field.CanInterface() {
		return nil, fmt.Errorf("polymorphic ID field %s is not accessible", p.polymorphicID)
	}

	return field.Interface(), nil
}

// getPolymorphicTypeValue gets the polymorphic type value from the model.
func (p *PolymorphicBelongsTo) getPolymorphicTypeValue() (any, error) {
	val := reflect.ValueOf(p.model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("model is not a struct")
	}

	field, err := getFieldByTagOrName(val, p.polymorphicType)
	if err != nil {
		return nil, err
	}
	if !field.IsValid() {
		return nil, fmt.Errorf("polymorphic type field %s not found", p.polymorphicType)
	}

	if !field.CanInterface() {
		return nil, fmt.Errorf("polymorphic type field %s is not accessible", p.polymorphicType)
	}

	return field.Interface(), nil
}

// setPolymorphicIDValue sets the polymorphic ID value on the model.
func (p *PolymorphicBelongsTo) setPolymorphicIDValue(value any) error {
	val := reflect.ValueOf(p.model)
	if val.Kind() == reflect.Ptr {
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

// setPolymorphicTypeValue sets the polymorphic type value on the model.
func (p *PolymorphicBelongsTo) setPolymorphicTypeValue(value any) error {
	val := reflect.ValueOf(p.model)
	if val.Kind() == reflect.Ptr {
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
