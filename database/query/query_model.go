package query

import (
	"fmt"
	"reflect"
	"strings"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/support/str"
)

// Model sets the model for the query.
func (q *Query) Model(value any) contractsorm.Query {
	q.model = value
	q.table = q.resolveTableName(value)

	// If driver is "array" and model implements ArraySource, populate the database
	if q.driver != nil && q.driver.Dialect() == "array" {
		if source, ok := value.(contractsorm.ArraySource); ok {
			tableName := source.TableName()
			if q.populatedTables == nil {
				q.populatedTables = make(map[string]bool)
			}

			if !q.populatedTables[tableName] {
				if arrayDriver, ok := q.driver.(contractsorm.ArrayPopulator); ok {
					// Use q.db directly instead of q.DB() because q.DB() returns
					// an error during transactions. The array driver needs *sql.DB
					// for DDL operations (CREATE TABLE / INSERT), which are not
					// transaction-scoped in SQLite for in-memory databases.
					if q.db == nil {
						q.buildError = fmt.Errorf("database connection is nil, cannot populate array source")
					} else if err := arrayDriver.Populate(q.ctx, q.db, source); err != nil {
						q.buildError = err
					} else {
						q.populatedTables[tableName] = true
					}
				}
			}
		}
	}

	// Reset query state to avoid pollution from previous queries
	q.selects = nil
	q.wheres = nil
	q.joins = nil
	q.groups = nil
	q.havings = nil
	q.orders = nil
	q.limit = nil
	q.offset = nil
	q.distinct = false
	q.distinctCols = nil
	q.aggregate = ""
	q.aggregateCol = ""
	q.rawSQL = ""
	q.rawArgs = nil
	q.lockForUpdate = false
	q.sharedLock = false
	q.omitColumns = nil
	// Don't reset soft delete state as it may be intentionally set
	// q.includeSoftDeleted = false
	// q.onlySoftDeleted = false
	// q.excludeSoftDeleted = false
	q.withRelations = nil
	q.relationConstraints = nil
	q.withCountQueries = nil
	q.withExistsQueries = nil
	return q
}

// resolveTableName resolves the table name from the model.
func (q *Query) resolveTableName(model any) string {
	if model == nil {
		return ""
	}

	v := reflect.ValueOf(model)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	// Check for TableName() string method
	if t, ok := model.(interface{ TableName() string }); ok {
		return t.TableName()
	}

	// Also check pointer receiver
	if v.CanAddr() {
		if t, ok := v.Addr().Interface().(interface{ TableName() string }); ok {
			return t.TableName()
		}
	}

	// Fallback to snake_case and pluralized struct name
	t := v.Type()
	if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		t = t.Elem()
		if t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
	}

	if t.Kind() != reflect.Struct {
		return ""
	}

	name := t.Name()
	snake := str.Of(name).Snake().String()

	// Simple pluralization
	if !strings.HasSuffix(snake, "s") {
		return snake + "s"
	}

	return snake
}
