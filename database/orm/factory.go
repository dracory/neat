package orm

import (
	"fmt"
	"reflect"
	"strings"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
)

// Factory implements the Factory interface for creating test data.
type Factory struct {
	orm   *Orm
	count int
}

// NewFactory creates a new Factory instance.
func NewFactory(orm *Orm) *Factory {
	return &Factory{
		orm:   orm,
		count: 1,
	}
}

// Count sets the number of models that should be generated.
func (f *Factory) Count(count int) contractsorm.Factory {
	if count <= 0 {
		count = 1
	}
	f.count = count
	return f
}

// Create creates a model and persists it to the database.
func (f *Factory) Create(value any, attributes ...map[string]any) error {
	instances, err := f.buildInstances(value, attributes...)
	if err != nil {
		return err
	}

	query := f.orm.Query()
	if query == nil {
		return fmt.Errorf("query not initialized")
	}
	return query.Create(instances)
}

// CreateQuietly creates a model and persists it to the database without firing any model events.
func (f *Factory) CreateQuietly(value any, attributes ...map[string]any) error {
	instances, err := f.buildInstances(value, attributes...)
	if err != nil {
		return err
	}

	query := f.orm.Query()
	if query == nil {
		return fmt.Errorf("query not initialized")
	}
	query = query.WithoutEvents()
	return query.Create(instances)
}

// Make creates a model and returns it, but does not persist it to the database.
// Note: When using Count() > 1, the created instances are not returned due to API limitations.
// The input model is modified in place for single instances (count == 1).
func (f *Factory) Make(value any, attributes ...map[string]any) error {
	instances, err := f.buildInstances(value, attributes...)
	if err != nil {
		return err
	}

	// If the input is a single model (not a slice), update it in place
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		// For single models, copy the created instance back to the input
		instancesValue := reflect.ValueOf(instances)
		if instancesValue.Kind() == reflect.Ptr {
			instancesValue = instancesValue.Elem()
		}
		// Try to assign if types are compatible
		if v.CanSet() {
			if instancesValue.Type().AssignableTo(v.Type()) {
				v.Set(instancesValue)
			} else if instancesValue.Type().ConvertibleTo(v.Type()) {
				v.Set(instancesValue.Convert(v.Type()))
			}
		}
	}

	return nil
}

// buildInstances creates the specified number of model instances with optional attributes.
func (f *Factory) buildInstances(value any, attributes ...map[string]any) (any, error) {
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	var mergedAttrs map[string]any
	if len(attributes) > 0 && attributes[0] != nil {
		mergedAttrs = attributes[0]
	}

	// Handle slice/array for bulk creation
	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		sliceType := reflect.SliceOf(v.Type().Elem())
		result := reflect.MakeSlice(sliceType, 0, f.count)

		for i := 0; i < f.count; i++ {
			// Create a new instance for each element
			elemType := v.Type().Elem()
			if elemType.Kind() == reflect.Ptr {
				elemType = elemType.Elem()
			}
			newElem := reflect.New(elemType).Elem()

			// If the original slice has elements, copy from the first one as template
			if v.Len() > 0 {
				template := v.Index(0)
				if template.Kind() == reflect.Ptr {
					template = template.Elem()
				}
				newElem.Set(template)
			}

			// Apply attributes
			if mergedAttrs != nil {
				if err := applyAttributes(newElem.Addr().Interface(), mergedAttrs); err != nil {
					return nil, err
				}
			}

			// If the element type is a pointer, wrap it
			if v.Type().Elem().Kind() == reflect.Ptr {
				result.Set(reflect.Append(result, newElem.Addr()))
			} else {
				result.Set(reflect.Append(result, newElem))
			}
		}

		// Return as the same type as input (pointer to slice or slice)
		if reflect.TypeOf(value).Kind() == reflect.Ptr {
			return result.Addr().Interface(), nil
		}
		return result.Interface(), nil
	}

	// Handle single model
	valueType := reflect.TypeOf(value)
	isPtr := valueType.Kind() == reflect.Ptr

	// Get the underlying type if it's a pointer
	var elemType reflect.Type
	if isPtr {
		elemType = valueType.Elem()
	} else {
		elemType = valueType
	}

	resultType := reflect.SliceOf(valueType)
	result := reflect.MakeSlice(resultType, f.count, f.count)

	for i := 0; i < f.count; i++ {
		var newInstance reflect.Value
		if isPtr {
			// Create new pointer to struct
			newInstance = reflect.New(elemType)
			// Copy struct values from original input
			originalValue := reflect.ValueOf(value)
			if originalValue.Kind() == reflect.Ptr {
				newInstance.Elem().Set(originalValue.Elem())
			}
		} else {
			// Create new struct value
			newInstance = reflect.New(elemType).Elem()
			if v.Kind() == reflect.Struct {
				newInstance.Set(v)
			}
		}

		// Apply attributes
		if mergedAttrs != nil {
			var modelPtr any
			if isPtr {
				modelPtr = newInstance.Interface()
			} else {
				modelPtr = newInstance.Addr().Interface()
			}
			if err := applyAttributes(modelPtr, mergedAttrs); err != nil {
				return nil, err
			}
		}

		result.Index(i).Set(newInstance)
	}

	// If count is 1, return single instance, otherwise return slice
	if f.count == 1 {
		return result.Index(0).Interface(), nil
	}
	return result.Interface(), nil
}

// applyAttributes applies the given attributes to a model instance using reflection.
func applyAttributes(model any, attributes map[string]any) error {
	v := reflect.ValueOf(model)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("model must be a pointer")
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("model must point to a struct")
	}

	t := v.Type()
	for key, value := range attributes {
		field, found := t.FieldByName(key)
		if !found {
			// Try to find field by JSON tag (parse tag to extract field name before options)
			for i := 0; i < t.NumField(); i++ {
				jsonTag := t.Field(i).Tag.Get("json")
				if jsonTag != "" {
					// Extract field name before comma (e.g., "name,omitempty" -> "name")
					fieldName := jsonTag
					if commaIdx := strings.Index(jsonTag, ","); commaIdx > 0 {
						fieldName = jsonTag[:commaIdx]
					}
					if fieldName == key {
						field = t.Field(i)
						found = true
						break
					}
				}
			}
		}

		if !found {
			continue // Skip unknown fields
		}

		fieldValue := v.FieldByName(field.Name)
		if !fieldValue.CanSet() {
			continue
		}

		attrValue := reflect.ValueOf(value)
		if attrValue.Type().ConvertibleTo(fieldValue.Type()) {
			fieldValue.Set(attrValue.Convert(fieldValue.Type()))
		} else {
			return fmt.Errorf("cannot convert attribute %q from type %T to field type %T", key, value, fieldValue.Interface())
		}
	}

	return nil
}
