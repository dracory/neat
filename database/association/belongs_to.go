package association

import (
	"fmt"
	"reflect"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
)

// BelongsTo represents a belongs-to relationship.
type BelongsTo struct {
	*Association
	foreignKey string
	otherKey   string
}

// NewBelongsTo creates a new BelongsTo association.
func NewBelongsTo(query contractsorm.Query, model any, association, foreignKey, otherKey string) *BelongsTo {
	return &BelongsTo{
		Association: NewAssociation(query, model, association),
		foreignKey:  foreignKey,
		otherKey:    otherKey,
	}
}

// Find loads the associated model for a belongs-to relationship.
func (b *BelongsTo) Find(out any, conds ...any) error {
	// Get the foreign key value from the model
	foreignKeyValue, err := b.getForeignKeyValue()
	if err != nil {
		return fmt.Errorf("failed to get foreign key value: %w", err)
	}

	// Query the related model using the other key
	query := b.Query().Table(b.associationName()).Where(b.otherKey+" = ?", foreignKeyValue)

	// Apply additional conditions if provided
	for _, cond := range conds {
		query = query.Where(cond)
	}

	// Execute the query
	return query.First(out)
}

// Append sets the foreign key to associate the model.
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
	return b.setForeignKeyValue(otherKeyValue)
}

// Replace replaces the current association with the given value.
func (b *BelongsTo) Replace(values ...any) error {
	return b.Append(values...)
}

// Delete clears the association by setting the foreign key to nil.
func (b *BelongsTo) Delete(values ...any) error {
	return b.setForeignKeyValue(nil)
}

// Clear clears the association by setting the foreign key to nil.
func (b *BelongsTo) Clear() error {
	return b.setForeignKeyValue(nil)
}

// Count returns 1 if the association exists, 0 otherwise.
func (b *BelongsTo) Count() int64 {
	foreignKeyValue, err := b.getForeignKeyValue()
	if err != nil {
		return 0
	}

	if foreignKeyValue == nil {
		return 0
	}

	var count int64
	query := b.Query().Table(b.associationName()).Where(b.otherKey+" = ?", foreignKeyValue)
	if err := query.Count(&count); err != nil {
		return 0
	}

	if count > 0 {
		return 1
	}
	return 0
}

// getForeignKeyValue gets the foreign key value from the model.
func (b *BelongsTo) getForeignKeyValue() (any, error) {
	val := reflect.ValueOf(b.model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("model is not a struct")
	}

	field := val.FieldByName(b.foreignKey)
	if !field.IsValid() {
		return nil, fmt.Errorf("foreign key field %s not found", b.foreignKey)
	}

	if !field.CanInterface() {
		return nil, fmt.Errorf("foreign key field %s is not accessible", b.foreignKey)
	}

	return field.Interface(), nil
}

// getOtherKeyValue gets the other key value from the related model.
func (b *BelongsTo) getOtherKeyValue(model any) (any, error) {
	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("model is not a struct")
	}

	field := val.FieldByName(b.otherKey)
	if !field.IsValid() {
		return nil, fmt.Errorf("other key field %s not found", b.otherKey)
	}

	if !field.CanInterface() {
		return nil, fmt.Errorf("other key field %s is not accessible", b.otherKey)
	}

	return field.Interface(), nil
}

// setForeignKeyValue sets the foreign key value on the model.
func (b *BelongsTo) setForeignKeyValue(value any) error {
	val := reflect.ValueOf(b.model)
	if val.Kind() == reflect.Ptr {
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
func (b *BelongsTo) associationName() string {
	return b.association
}
