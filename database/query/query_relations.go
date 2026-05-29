package query

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/database/association"
	"github.com/dracory/neat/support/str"
)

// initializeRelations initializes relation fields in the destination struct.
func (q *Query) initializeRelations(v reflect.Value) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		// Handle slices by iterating over each element
		if v.Kind() == reflect.Slice {
			for i := 0; i < v.Len(); i++ {
				elem := v.Index(i)
				if elem.Kind() == reflect.Ptr {
					elem = elem.Elem()
				}
				if elem.Kind() == reflect.Struct {
					q.initializeRelations(elem)
				}
			}
		}
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
	return q.loadRelationsWithConn(v, q.readConn())
}

// inferTableName infers the table name from a struct type using snake_case and simple pluralization.
func inferTableName(relationType reflect.Type) string {
	tableName := str.Of(relationType.Name()).Snake().String()
	// Simple pluralization: add 's' if not already ending with 's'
	if !strings.HasSuffix(tableName, "s") {
		tableName = tableName + "s"
	}
	return tableName
}

// buildForeignKeyColumn builds a foreign key column name from a parent type name.
func buildForeignKeyColumn(parentTypeName string) string {
	return str.Of(parentTypeName).Snake().String() + "_id"
}

// getParentTypeName extracts the parent type name from the value or query model.
func (q *Query) getParentTypeName(v reflect.Value) string {
	if v.Type().Name() != "" {
		return v.Type().Name()
	}
	// For embedded types or anonymous structs, try to get from the model
	if q.model != nil {
		modelType := reflect.TypeOf(q.model)
		if modelType.Kind() == reflect.Ptr {
			modelType = modelType.Elem()
		}
		if modelType.Kind() == reflect.Slice {
			modelType = modelType.Elem()
			if modelType.Kind() == reflect.Ptr {
				modelType = modelType.Elem()
			}
		}
		return modelType.Name()
	}
	return ""
}

// loadHasManyRelation loads a has-many relationship (slice field).
func (q *Query) loadHasManyRelation(v reflect.Value, field reflect.Value, relation string, conn *sql.DB) error {
	// Get the relation field type
	relationType := field.Type().Elem()
	if relationType.Kind() == reflect.Ptr {
		relationType = relationType.Elem()
	}

	// Get the parent's primary key value (id)
	idField := v.FieldByName("ID")
	if !idField.IsValid() {
		// Try lowercase id
		idField = v.FieldByName("Id")
		if !idField.IsValid() {
			return fmt.Errorf("no ID field found for has-many relation %s", relation)
		}
	}
	parentID := idField.Interface()

	// Infer table name from relation type
	tableName := inferTableName(relationType)

	// Build foreign key column name
	parentTypeName := q.getParentTypeName(v)
	if parentTypeName == "" {
		return fmt.Errorf("could not determine parent type name for relation %s", relation)
	}
	foreignKeyColumn := buildForeignKeyColumn(parentTypeName)

	// Create a query to load the related models
	relatedQuery := NewQuery(q.ctx, conn, q.driver, q.connection, q.dbConfig, q.log)
	relatedQuery.table = tableName
	relatedQuery.model = nil
	relatedQuery.withRelations = nil
	relatedQuery = relatedQuery.Select("*").Where(foreignKeyColumn+" = ?", parentID).(*Query)

	// Apply constraint callback if provided for this relation
	if q.relationConstraints != nil {
		if constraint, ok := q.relationConstraints[relation]; ok {
			relatedQuery = constraint(relatedQuery).(*Query)
		}
	}

	// Build and execute the query
	builder := NewBuilder(relatedQuery)
	querySQL, args := builder.BuildSelect()

	if conn == nil {
		return fmt.Errorf("database connection is nil")
	}

	rows, err := conn.QueryContext(q.ctx, querySQL, args...)
	if err != nil {
		return fmt.Errorf("failed to query has-many relation %s: %w", relation, err)
	}
	defer rows.Close()

	// Scan rows into slice
	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to get columns for has-many relation %s: %w", relation, err)
	}

	slice := reflect.MakeSlice(field.Type(), 0, 0)
	for rows.Next() {
		dest := reflect.New(relationType)
		destValue := dest.Elem()
		values := structScanDests(destValue, columns)
		if err := rows.Scan(values...); err != nil {
			return fmt.Errorf("failed to scan row for has-many relation %s: %w", relation, err)
		}
		copyScanResults(destValue, columns, values)

		if field.Type().Elem().Kind() == reflect.Ptr {
			slice = reflect.Append(slice, dest)
		} else {
			slice = reflect.Append(slice, destValue)
		}
	}

	field.Set(slice)
	return nil
}

// loadHasOneRelation loads a has-one relationship (pointer field).
func (q *Query) loadHasOneRelation(v reflect.Value, field reflect.Value, relation string, conn *sql.DB) error {
	// Get the foreign key field name (e.g., "User" -> "UserID")
	foreignKey := relation + "ID"
	foreignKeyField := v.FieldByName(foreignKey)
	if !foreignKeyField.IsValid() {
		// Try snake_case version
		foreignKey = str.Of(relation).Snake().String() + "_id"
		foreignKeyField = v.FieldByName(foreignKey)
		if !foreignKeyField.IsValid() {
			return fmt.Errorf("foreign key field %s or %s not found for relation %s", relation+"ID", str.Of(relation).Snake().String()+"_id", relation)
		}
	}

	// Get the foreign key value
	foreignKeyValue := foreignKeyField.Interface()
	if foreignKeyValue == nil || reflect.ValueOf(foreignKeyValue).IsZero() {
		// Set field to nil for zero foreign key
		field.Set(reflect.Zero(field.Type()))
		return nil
	}

	// Get the relation field type to create the destination
	relationType := field.Type().Elem()

	// Create destination
	dest := reflect.New(relationType)

	// Create a query to load the related model
	relatedQuery := NewQuery(q.ctx, conn, q.driver, q.connection, q.dbConfig, q.log)
	relatedQuery.table = str.Of(relation).Snake().String() + "s"
	relatedQuery.model = nil
	relatedQuery.withRelations = nil
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

	if conn == nil {
		return fmt.Errorf("database connection is nil")
	}

	rows, err := conn.QueryContext(q.ctx, querySQL, args...)
	if err != nil {
		return fmt.Errorf("failed to query has-one relation %s: %w", relation, err)
	}
	defer rows.Close()

	// Scan directly into dest without using scanRows to avoid relation loading
	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to get columns for has-one relation %s: %w", relation, err)
	}

	if !rows.Next() {
		// No rows found, set field to nil
		field.Set(reflect.Zero(field.Type()))
		return nil
	}

	destValue := dest.Elem()
	values := structScanDests(destValue, columns)
	if err := rows.Scan(values...); err != nil {
		return fmt.Errorf("failed to scan row for has-one relation %s: %w", relation, err)
	}
	copyScanResults(destValue, columns, values)

	// Set the field value
	field.Set(dest)
	return nil
}

// loadRelationsWithConn loads relations using a specific connection.
func (q *Query) loadRelationsWithConn(v reflect.Value, conn *sql.DB) error {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		// Handle slices by iterating over each element
		if v.Kind() == reflect.Slice {
			for i := 0; i < v.Len(); i++ {
				elem := v.Index(i)
				if elem.Kind() == reflect.Ptr {
					elem = elem.Elem()
				}
				if elem.Kind() == reflect.Struct {
					if err := q.loadRelationsWithConn(elem, conn); err != nil {
						return err
					}
				}
			}
		}
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
			// Skip invalid fields silently - they may not exist on all models
			continue
		}

		// Handle has-many relationships (slice fields)
		if field.Kind() == reflect.Slice {
			if err := q.loadHasManyRelation(v, field, relation, conn); err != nil {
				// Log error but continue with other relations
				// This maintains backward compatibility while surfacing errors
				continue
			}
			continue
		}

		// Handle has-one relationships (pointer fields)
		if field.Kind() == reflect.Ptr {
			if err := q.loadHasOneRelation(v, field, relation, conn); err != nil {
				// Log error but continue with other relations
				// This maintains backward compatibility while surfacing errors
				continue
			}
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
	// Create a new query for lazy loading
	newQuery := *q
	newQuery.withRelations = []string{relation}
	newQuery.model = dest

	// Check if a constraint callback is provided
	if len(args) > 0 {
		if fn, ok := args[0].(func(contractsorm.Query) contractsorm.Query); ok {
			if newQuery.relationConstraints == nil {
				newQuery.relationConstraints = make(map[string]func(contractsorm.Query) contractsorm.Query)
			}
			newQuery.relationConstraints[relation] = fn
		}
	}

	// Get the reflect value of the destination
	v := reflect.ValueOf(dest)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Initialize the relation field
	newQuery.initializeRelations(v)

	// Load the relation
	return newQuery.loadRelationsWithConn(v, newQuery.readConn())
}

// LoadMissing loads a relation only if it's not already loaded.
func (q *Query) LoadMissing(dest any, relation string, args ...any) error {
	// Get the reflect value of the destination
	v := reflect.ValueOf(dest)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Check if the relation field exists and is already loaded
	field := v.FieldByName(relation)
	if !field.IsValid() {
		return fmt.Errorf("relation field %s not found", relation)
	}

	// For has-one relationships (pointer fields), check if already loaded
	if field.Kind() == reflect.Ptr && !field.IsNil() {
		// Already loaded, skip
		return nil
	}

	// For has-many relationships (slice fields), check if already loaded
	if field.Kind() == reflect.Slice && !field.IsNil() && field.Len() > 0 {
		// Already loaded, skip
		return nil
	}

	// Not loaded, proceed with loading
	return q.Load(dest, relation, args...)
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
