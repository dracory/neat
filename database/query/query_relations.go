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
					if err := q.loadRelations(elem); err != nil {
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
			continue
		}

		// Handle has-many relationships (slice fields)
		if field.Kind() == reflect.Slice {
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
					fmt.Printf("DEBUG: No ID field found for relation %s, v.Type()=%v\n", relation, v.Type())
					continue
				}
			}
			parentID := idField.Interface()
			fmt.Printf("DEBUG: Loading has-many relation %s, parentID=%v, tableName will be based on %s\n", relation, parentID, relationType.Name())

			// Create a query to load the related models
			relatedQuery := NewQuery(q.ctx, q.db, q.driver, q.connection, q.dbConfig, q.log)
			// Use the relation type name to determine table name (singular form)
			tableName := str.Of(relationType.Name()).Snake().String()
			// Simple pluralization: add 's' if not already ending with 's'
			if !strings.HasSuffix(tableName, "s") {
				tableName = tableName + "s"
			}
			relatedQuery.table = tableName
			relatedQuery.model = nil
			relatedQuery.withRelations = nil
			fmt.Printf("DEBUG: relatedQuery.db is nil: %v\n", relatedQuery.db == nil)

			// Build the base query with the foreign key condition on the related table
			// e.g., for Post.Comments, query comments where post_id = ?
			// Try to get the struct name from the parent type
			parentTypeName := ""
			if v.Type().Name() != "" {
				parentTypeName = v.Type().Name()
			} else {
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
					parentTypeName = modelType.Name()
				}
			}
			foreignKeyColumn := str.Of(parentTypeName).Snake().String() + "_id"
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
			fmt.Printf("DEBUG: Executing query: %s with args: %v\n", querySQL, args)
			fmt.Printf("DEBUG: q.db == relatedQuery.db: %v\n", q.db == relatedQuery.db)

			// Verify table exists
			checkRows, checkErr := q.db.QueryContext(q.ctx, "SELECT name FROM sqlite_master WHERE type='table' AND name=?", tableName)
			if checkErr != nil {
				fmt.Printf("DEBUG: Error checking table: %v\n", checkErr)
			} else {
				defer checkRows.Close()
				if checkRows.Next() {
					var tableName string
					checkRows.Scan(&tableName)
					fmt.Printf("DEBUG: Table exists: %s\n", tableName)
				} else {
					fmt.Printf("DEBUG: Table does NOT exist: %s\n", tableName)
				}
			}

			var rows *sql.Rows
			var err error
			// Use the same connection as the parent query
			if q.readDB != nil {
				rows, err = q.readDB.QueryContext(q.ctx, querySQL, args...)
			} else if q.db != nil {
				rows, err = q.db.QueryContext(q.ctx, querySQL, args...)
			} else {
				fmt.Printf("DEBUG: No database connection available\n")
				continue
			}
			if err != nil {
				fmt.Printf("DEBUG: Query error: %v\n", err)
				continue
			}
			defer rows.Close()

			// Scan rows into slice
			columns, err := rows.Columns()
			if err != nil {
				continue
			}

			slice := reflect.MakeSlice(field.Type(), 0, 0)
			for rows.Next() {
				dest := reflect.New(relationType)
				destValue := dest.Elem()
				values := structScanDests(destValue, columns)
				if err := rows.Scan(values...); err != nil {
					continue
				}
				copyScanResults(destValue, columns, values)

				if field.Type().Elem().Kind() == reflect.Ptr {
					slice = reflect.Append(slice, dest)
				} else {
					slice = reflect.Append(slice, destValue)
				}
			}

			field.Set(slice)
			continue
		}

		// Handle has-one relationships (pointer fields)
		if field.Kind() == reflect.Ptr {
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
			relationType := field.Type().Elem()

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
				// No rows found, set field to nil
				field.Set(reflect.Zero(field.Type()))
				continue
			}

			destValue := dest.Elem()
			values := structScanDests(destValue, columns)
			if err := rows.Scan(values...); err != nil {
				continue
			}
			copyScanResults(destValue, columns, values)

			// Set the field value
			field.Set(dest)
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
