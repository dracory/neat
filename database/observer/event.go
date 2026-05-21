package observer

import (
	"context"
	"reflect"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
)

// Event implements the Event interface for model lifecycle events.
type Event struct {
	ctx        context.Context
	model      any
	original   map[string]any
	attributes map[string]any
	dirty      map[string]bool
	query      contractsorm.Query
	eventType  contractsorm.EventType
}

// NewEvent creates a new Event instance.
func NewEvent(
	ctx context.Context,
	model any,
	original map[string]any,
	attributes map[string]any,
	dirty map[string]bool,
	query contractsorm.Query,
	eventType contractsorm.EventType,
) *Event {
	return &Event{
		ctx:        ctx,
		model:      model,
		original:   original,
		attributes: attributes,
		dirty:      dirty,
		query:      query,
		eventType:  eventType,
	}
}

// Context returns the event context.
func (e *Event) Context() context.Context {
	return e.ctx
}

// GetAttribute returns the attribute value for the given key.
func (e *Event) GetAttribute(key string) any {
	if e.attributes == nil {
		return nil
	}
	return e.attributes[key]
}

// GetOriginal returns the original attribute value for the given key.
func (e *Event) GetOriginal(key string, def ...any) any {
	if e.original == nil {
		if len(def) > 0 {
			return def[0]
		}
		return nil
	}
	val, ok := e.original[key]
	if !ok && len(def) > 0 {
		return def[0]
	}
	return val
}

// IsClean returns true if the given column is clean (not dirty).
func (e *Event) IsClean(columns ...string) bool {
	if e.dirty == nil {
		return true
	}
	if len(columns) == 0 {
		return len(e.dirty) == 0
	}
	for _, col := range columns {
		if e.dirty[col] {
			return false
		}
	}
	return true
}

// IsDirty returns true if the given column is dirty (has been modified).
func (e *Event) IsDirty(columns ...string) bool {
	if e.dirty == nil {
		return false
	}
	if len(columns) == 0 {
		return len(e.dirty) > 0
	}
	for _, col := range columns {
		if e.dirty[col] {
			return true
		}
	}
	return false
}

// Query returns the query instance.
func (e *Event) Query() contractsorm.Query {
	return e.query
}

// SetAttribute sets the attribute value for the given key.
func (e *Event) SetAttribute(key string, value any) {
	if e.attributes == nil {
		e.attributes = make(map[string]any)
	}
	e.attributes[key] = value
}

// Model returns the model instance.
func (e *Event) Model() any {
	return e.model
}

// EventType returns the event type.
func (e *Event) EventType() contractsorm.EventType {
	return e.eventType
}

// ExtractModelAttributes extracts attributes from a model using reflection.
func ExtractModelAttributes(model any) map[string]any {
	attrs := make(map[string]any)
	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return attrs
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// Skip unexported fields
		if !fieldValue.CanInterface() {
			continue
		}

		// Get the field name (use json tag if available, otherwise use field name)
		name := field.Name
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			if jsonTag == "-" {
				continue
			}
			name = jsonTag
		}

		attrs[name] = fieldValue.Interface()
	}

	return attrs
}
