package query

import (
	"database/sql"
	"fmt"
	"reflect"

	"github.com/dracory/neat/database/association"
	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/support/str"
)

// initializeRelations initializes relation fields in the destination struct.
func (q *Query) initializeRelations(v reflect.Value) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return
	}

	for _, relation := range q.withRelations {
		field := v.FieldByName(relation)
		if !field.IsValid() {
			continue
		}

		if field.Kind() == reflect.Ptr && field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		} else if field.Kind() == reflect.Slice && field.IsNil() {
			field.Set(reflect.MakeSlice(field.Type(), 0, 0))
		}
	}
}

// loadRelations loads the actual data for relations requested via With.
func (q *Query) loadRelations(v reflect.Value) error {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}

	// Save and clear withRelations to prevent recursion during relation loading
	savedWithRelations := q.withRelations
	q.withRelations = nil

	// Disable observers during relation loading to prevent recursive events
	savedWithoutEvents := q.withoutEvents
	q.withoutEvents = true
	defer func() { q.withoutEvents = savedWithoutEvents }()

	for _, relation := range savedWithRelations {
		field := v.FieldByName(relation)
		if !field.IsValid() {
			continue
		}

		// Get the foreign key field name (e.g., "User" -> "UserID")
		foreignKey := relation + "ID"
		foreignKeyField := v.FieldByName(foreignKey)
		if !foreignKeyField.IsValid() {
			// Try snake_case version
			foreignKey = str.Of(relation).Snake().String() + "_id"
			foreignKeyField = v.FieldByName(foreignKey)
			if !foreignKeyField.IsValid() {
				continue
			}
		}

		// Get the foreign key value
		foreignKeyValue := foreignKeyField.Interface()
		if foreignKeyValue == nil || reflect.ValueOf(foreignKeyValue).IsZero() {
			continue
		}

		// Get the relation field type to create the destination
		relationType := field.Type()
		if relationType.Kind() == reflect.Ptr {
			relationType = relationType.Elem()
		}

		// Create destination
		dest := reflect.New(relationType)

		// Create a query to load the related model
		relatedQuery := NewQuery(q.ctx, q.db, q.driver, q.connection, q.dbConfig, q.log)
		relatedQuery.table = str.Of(relation).Snake().String() + "s"
		relatedQuery.model = nil
		relatedQuery.withRelations = nil // Prevent recursive relation loading

		// Build the base query with the foreign key condition
		relatedQuery = relatedQuery.Select("*").Where("id = ?", foreignKeyValue).(*Query)

		// Apply constraint callback if provided for this relation
		if q.relationConstraints != nil {
			if constraint, ok := q.relationConstraints[relation]; ok {
				relatedQuery = constraint(relatedQuery).(*Query)
			}
		}

		// Build and execute the query
		builder := NewBuilder(relatedQuery)
		querySQL, args := builder.BuildSelect()

		// Use a separate connection to avoid SQLite deadlock during relation loading
		var rows *sql.Rows
		var err error
		if q.readDB != nil {
			rows, err = q.readDB.QueryContext(q.ctx, querySQL, args...)
		} else {
			rows, err = q.db.QueryContext(q.ctx, querySQL, args...)
		}
		if err != nil {
			continue
		}
		defer rows.Close()

		// Scan directly into dest without using scanRows to avoid relation loading
		columns, err := rows.Columns()
		if err != nil {
			continue
		}

		if !rows.Next() {
			// No rows found, set field to nil if it's a pointer
			if field.Kind() == reflect.Ptr {
				field.Set(reflect.Zero(field.Type()))
			}
			continue
		}

		destValue := dest.Elem()
		values := structScanDests(destValue, columns)
		if err := rows.Scan(values...); err != nil {
			continue
		}
		copyScanResults(destValue, columns, values)

		// Set the field value
		if field.Kind() == reflect.Ptr {
			field.Set(dest)
		} else {
			field.Set(dest.Elem())
		}
	}

	// Restore withRelations after all relations are loaded
	q.withRelations = savedWithRelations

	return nil
}

// With specifies relations to eager load.
func (q *Query) With(query string, args ...any) contractsorm.Query {
	newQuery := *q
	newQuery.withRelations = append(newQuery.withRelations, query)

	// Check if a constraint callback is provided
	if len(args) > 0 {
		if fn, ok := args[0].(func(contractsorm.Query) contractsorm.Query); ok {
			if newQuery.relationConstraints == nil {
				newQuery.relationConstraints = make(map[string]func(contractsorm.Query) contractsorm.Query)
			}
			newQuery.relationConstraints[query] = fn
		}
	}

	return &newQuery
}

// Load loads a relation for the given model.
func (q *Query) Load(dest any, relation string, args ...any) error {
	// This is a simplified implementation - full lazy loading requires
	// additional work to detect relationships and load them properly
	return fmt.Errorf("lazy loading not fully implemented yet")
}

// LoadMissing loads a relation only if it's not already loaded.
func (q *Query) LoadMissing(dest any, relation string, args ...any) error {
	// This is a simplified implementation - full lazy loading requires
	// additional work to detect relationships and load them properly
	return fmt.Errorf("lazy loading not fully implemented yet")
}

// Without removes specified relations from eager loading.
func (q *Query) Without(relations ...string) contractsorm.Query {
	newQuery := *q
	// Remove specified relations from withRelations
	for _, relation := range relations {
		for i, r := range newQuery.withRelations {
			if r == relation {
				newQuery.withRelations = append(newQuery.withRelations[:i], newQuery.withRelations[i+1:]...)
				break
			}
		}
		// Also remove any constraint for this relation
		if newQuery.relationConstraints != nil {
			delete(newQuery.relationConstraints, relation)
		}
	}
	return &newQuery
}

// WithCount adds a count query to the relations (not yet implemented).
func (q *Query) WithCount(query string, args ...any) contractsorm.Query {
	return q
}

// WithExists adds an exists query to the relations (not yet implemented).
func (q *Query) WithExists(query string, args ...any) contractsorm.Query {
	return q
}

// Association returns an association for the given relationship name.
func (q *Query) Association(assocName string) contractsorm.Association {
	// Return a base association - specific relationship types should be created
	// based on the relationship metadata from the model
	return association.NewAssociation(q, q.model, assocName)
}
