package association

import (
	"fmt"
	"reflect"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
)

// HasMany represents a has-many relationship.
type HasMany struct {
	*Association
	foreignKey string
	localKey   string
}

// NewHasMany creates a new HasMany association.
func NewHasMany(query contractsorm.Query, model any, association, foreignKey, localKey string) *HasMany {
	return &HasMany{
		Association: NewAssociation(query, model, association),
		foreignKey:  foreignKey,
		localKey:    localKey,
	}
}

// Find loads the associated models for a has-many relationship.
func (h *HasMany) Find(out any, conds ...any) error {
	// Get the local key value from the model
	localKeyValue, err := h.getLocalKeyValue()
	if err != nil {
		return fmt.Errorf("failed to get local key value: %w", err)
	}

	// Query the related models using the foreign key
	query := h.Query().Table(h.associationName()).Where(h.foreignKey+" = ?", localKeyValue)

	// Apply additional conditions if provided
	for _, cond := range conds {
		query = query.Where(cond)
	}

	// Execute the query
	return query.Get(out)
}

// Append appends models to the association.
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
func (h *HasMany) Replace(values ...any) error {
	// First, clear the current association by setting foreign key to null
	if err := h.Clear(); err != nil {
		return fmt.Errorf("failed to clear association: %w", err)
	}

	// Then, append the new values
	return h.Append(values...)
}

// Delete removes the given values from the association.
func (h *HasMany) Delete(values ...any) error {
	if len(values) == 0 {
		return fmt.Errorf("no values provided to delete")
	}

	// Set the foreign key to null for each related model
	for _, value := range values {
		if err := h.setForeignKeyValue(value, nil); err != nil {
			return fmt.Errorf("failed to clear foreign key on model: %w", err)
		}

		// Save the related model
		query := h.Query().Model(value)
		if err := query.Save(value); err != nil {
			return fmt.Errorf("failed to save related model: %w", err)
		}
	}

	return nil
}

// Clear clears the association by setting the foreign key to null for all related models.
func (h *HasMany) Clear() error {
	// Get the local key value from the model
	localKeyValue, err := h.getLocalKeyValue()
	if err != nil {
		return fmt.Errorf("failed to get local key value: %w", err)
	}

	// Update all related models to set foreign key to null
	query := h.Query().Table(h.associationName()).Where(h.foreignKey+" = ?", localKeyValue)
	_, err = query.Update(h.foreignKey, nil)
	if err != nil {
		return fmt.Errorf("failed to clear association: %w", err)
	}

	return nil
}

// Count returns the number of records in the association.
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
func (h *HasMany) getLocalKeyValue() (any, error) {
	val := reflect.ValueOf(h.model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("model is not a struct")
	}

	field := val.FieldByName(h.localKey)
	if !field.IsValid() {
		return nil, fmt.Errorf("local key field %s not found", h.localKey)
	}

	if !field.CanInterface() {
		return nil, fmt.Errorf("local key field %s is not accessible", h.localKey)
	}

	return field.Interface(), nil
}

// setForeignKeyValue sets the foreign key value on a related model.
func (h *HasMany) setForeignKeyValue(model any, value any) error {
	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("model is not a struct")
	}

	field := val.FieldByName(h.foreignKey)
	if !field.IsValid() {
		return fmt.Errorf("foreign key field %s not found", h.foreignKey)
	}

	if !field.CanSet() {
		return fmt.Errorf("foreign key field %s is not settable", h.foreignKey)
	}

	valueVal := reflect.ValueOf(value)
	if valueVal.IsValid() && valueVal.Type() != field.Type() {
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
func (h *HasMany) associationName() string {
	return h.association
}
