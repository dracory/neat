package query

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/support/str"
)

// Model sets the model for the query.
func (q *Query) Model(value any) orm.Query {
	q.model = value
	q.table = q.resolveTableName(value)

	// If driver is "array" and model implements ArraySource, populate the database
	if q.driver != nil && q.driver.Dialect() == "array" {
		if source, ok := value.(orm.ArraySource); ok {
			tableName := source.TableName()
			if q.populatedTables == nil {
				q.populatedTables = make(map[string]bool)
			}

			if !q.populatedTables[tableName] {
				if arrayDriver, ok := q.driver.(orm.ArrayPopulator); ok {
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

	// Load global scopes if model implements orm.GlobalScope
	q.globalScopes = nil
	q.withoutGlobalScopes = false
	q.disabledScopes = nil
	q.scopes = nil

	if gs, ok := value.(orm.GlobalScope); ok {
		q.globalScopes = gs.GlobalScopes()
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

// inferTableFromValue resolves the table name from the given value when no
// table was configured explicitly via .Model(), .Table() or raw SQL. The
// value may be a struct, a pointer to a struct, or a slice destination.
func (q *Query) inferTableFromValue(value any) {
	if q.table != "" || q.rawSQL != "" {
		return
	}
	table := q.resolveTableName(value)
	if table == "" {
		return
	}
	q.table = table
	q.adoptInferredModel(modelInstance(value))
}

// inferTableFromConds resolves the table name from the first struct-like
// condition argument (e.g. a model passed to Delete) when no table was
// configured explicitly.
func (q *Query) inferTableFromConds(conds []any) {
	if q.table != "" || q.rawSQL != "" {
		return
	}
	for _, c := range conds {
		if !isStructLike(c) {
			continue
		}
		table := q.resolveTableName(c)
		if table == "" {
			continue
		}
		q.table = table
		q.adoptInferredModel(modelInstance(c))
		return
	}
}

// adoptInferredModel sets q.model from an inferred value and loads the
// model's global scopes, mirroring the relevant parts of Model().
func (q *Query) adoptInferredModel(m any) {
	if q.model == nil {
		q.model = m
	}
	if gs, ok := q.model.(orm.GlobalScope); ok && len(q.globalScopes) == 0 {
		q.globalScopes = gs.GlobalScopes()
	}
}

// modelInstance returns a model value usable for capability checks (soft
// deletes, global scopes): the value itself when it is struct-like, otherwise
// a fresh instance of the slice/array element struct (e.g. *[]T -> *T).
func modelInstance(value any) any {
	if isStructLike(value) {
		return value
	}
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		t := v.Type().Elem()
		if t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
		if t.Kind() == reflect.Struct {
			return reflect.New(t).Interface()
		}
	}
	return nil
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

	// Check TableName() on the resolved struct element type. This covers
	// slice destinations such as *[]User and []User where the method check
	// above cannot see through the slice.
	if tn, ok := reflect.New(t).Interface().(interface{ TableName() string }); ok {
		return tn.TableName()
	}

	name := t.Name()
	snake := str.Of(name).Snake().String()

	// Simple pluralization
	if !strings.HasSuffix(snake, "s") {
		return snake + "s"
	}

	return snake
}
