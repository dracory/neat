package association

import (
	"fmt"
	"reflect"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
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
	query := h.Query().Table(h.associationName()).Where(h.foreignKey+" = ?", localKeyValue)

	// Apply additional conditions if provided
	for _, cond := range conds {
		query = query.Where(cond)
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

	// Set the foreign key to null for the related model
	if err := h.setForeignKeyValue(values[0], nil); err != nil {
		return fmt.Errorf("failed to clear foreign key on model: %w", err)
	}

	// Save the related model
	query := h.Query().Model(values[0])
	if err := query.Save(values[0]); err != nil {
		return fmt.Errorf("failed to save related model: %w", err)
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
func (h *HasOne) setForeignKeyValue(model any, value any) error {
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
func (h *HasOne) associationName() string {
	return h.association
}
