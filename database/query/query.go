package query

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"time"

	"github.com/dracory/neat/contracts/database"
	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/contracts/log"
	"github.com/dracory/neat/database/association"
	"github.com/dracory/neat/database/cursor"
	"github.com/dracory/neat/database/db"
	"github.com/dracory/neat/database/driver"
	"github.com/dracory/neat/database/observer"
)

// Query implements the Query interface using native database/sql.
type Query struct {
	ctx        context.Context
	db         *sql.DB
	driver     driver.Driver
	connection string
	dbConfig   *db.DBConfig
	log        log.Log
	queryLog   []contractsorm.QueryLog
	enableLog  bool

	// Query state
	table        string
	model        any
	selects      []any
	wheres       []whereClause
	joins        []joinClause
	groups       []string
	havings      []havingClause
	orders       []orderClause
	limit        *int
	offset       *int
	distinct     bool
	aggregate    string
	aggregateCol string

	// Transaction state
	inTransaction bool
	tx            *sql.Tx

	// Observer state
	modelToObserver []contractsorm.ModelToObserver
	withoutEvents   bool
	dispatcher      *observer.Dispatcher

	// Soft delete state
	withTrashed    bool
	onlyTrashed    bool
	withoutTrashed bool

	// Eager loading state
	withRelations []string
}

type whereClause struct {
	_type string // "and", "or"
	query string
	args  []any
}

type joinClause struct {
	_type string // "join", "left join", "right join", "cross join"
	query string
	args  []any
}

type havingClause struct {
	query string
	args  []any
}

type orderClause struct {
	column    string
	direction string // "asc", "desc"
}

// NewQuery creates a new Query instance.
func NewQuery(ctx context.Context, db *sql.DB, drv driver.Driver, connection string, dbConfig *db.DBConfig, log log.Log) *Query {
	return &Query{
		ctx:             ctx,
		db:              db,
		driver:          drv,
		connection:      connection,
		dbConfig:        dbConfig,
		log:             log,
		enableLog:       false,
		queryLog:        make([]contractsorm.QueryLog, 0),
		modelToObserver: make([]contractsorm.ModelToObserver, 0),
		withoutEvents:   false,
		dispatcher:      observer.NewDispatcher(log),
	}
}

// Connection returns a new Query instance with the specified connection.
func (q *Query) Connection(name string) contractsorm.Query {
	// This will be implemented when the ORM is fully integrated
	return q
}

// Model sets the model for the query.
func (q *Query) Model(value any) contractsorm.Query {
	q.model = value
	return q
}

// Table sets the table for the query.
func (q *Query) Table(name string, args ...any) contractsorm.Query {
	q.table = name
	return q
}

// Select adds columns to the select clause.
func (q *Query) Select(query any, args ...any) contractsorm.Query {
	q.selects = append(q.selects, query)
	return q
}

// Where adds a where clause to the query.
func (q *Query) Where(query any, args ...any) contractsorm.Query {
	q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("%v", query), args: args})
	return q
}

// OrWhere adds an or where clause to the query.
func (q *Query) OrWhere(query any, args ...any) contractsorm.Query {
	q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("%v", query), args: args})
	return q
}

// Order adds an order by clause to the query.
func (q *Query) Order(value any) contractsorm.Query {
	q.orders = append(q.orders, orderClause{column: fmt.Sprintf("%v", value), direction: "asc"})
	return q
}

// OrderBy adds an order by clause with direction.
func (q *Query) OrderBy(column string, direction ...string) contractsorm.Query {
	dir := "asc"
	if len(direction) > 0 {
		dir = direction[0]
	}
	q.orders = append(q.orders, orderClause{column: column, direction: dir})
	return q
}

// OrderByDesc adds an order by clause with desc direction.
func (q *Query) OrderByDesc(column string) contractsorm.Query {
	q.orders = append(q.orders, orderClause{column: column, direction: "desc"})
	return q
}

// Limit adds a limit clause to the query.
func (q *Query) Limit(limit int) contractsorm.Query {
	q.limit = &limit
	return q
}

// Offset adds an offset clause to the query.
func (q *Query) Offset(offset int) contractsorm.Query {
	q.offset = &offset
	return q
}

// Distinct adds distinct to the query.
func (q *Query) Distinct(args ...any) contractsorm.Query {
	q.distinct = true
	return q
}

// DB returns the underlying database connection.
func (q *Query) DB() (*sql.DB, error) {
	if q.tx != nil {
		return nil, fmt.Errorf("cannot get DB during transaction, use transaction methods instead")
	}
	return q.db, nil
}

// InTransaction returns true if the query is in a transaction.
func (q *Query) InTransaction() bool {
	return q.inTransaction
}

// Driver returns the database driver.
func (q *Query) Driver() database.Driver {
	return database.Driver(q.driver.Dialect())
}

// EnableQueryLog enables query logging.
func (q *Query) EnableQueryLog() {
	q.enableLog = true
}

// DisableQueryLog disables query logging.
func (q *Query) DisableQueryLog() {
	q.enableLog = false
}

// Find retrieves records matching the given conditions.
func (q *Query) Find(dest any, conds ...any) error {
	// Add conditions to where clause
	for _, cond := range conds {
		q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("%v", cond), args: nil})
	}

	// Build SELECT query
	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// Execute query
	if q.tx != nil {
		// Use transaction
		rows, err := q.tx.QueryContext(q.ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
		defer rows.Close()

		// Log query if enabled
		if q.enableLog {
			q.queryLog = append(q.queryLog, contractsorm.QueryLog{
				Query:    sql,
				Bindings: args,
				Time:     0, // TODO: track duration
			})
		}

		// Scan results into dest
		return q.scanRows(rows, dest)
	}

	dbConn, err := q.DB()
	if err != nil {
		return err
	}

	rows, err := dbConn.QueryContext(q.ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Log query if enabled
	if q.enableLog {
		q.queryLog = append(q.queryLog, contractsorm.QueryLog{
			Query:    sql,
			Bindings: args,
			Time:     0, // TODO: track duration
		})
	}

	return q.scanRows(rows, dest)
}

// FindOrFail retrieves records matching the given conditions or returns an error if not found.
func (q *Query) FindOrFail(dest any, conds ...any) error {
	return fmt.Errorf("not implemented")
}

// First retrieves the first record matching the query.
func (q *Query) First(dest any) error {
	// Set limit to 1
	limit := 1
	q.limit = &limit

	// Build SELECT query
	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// Execute query
	var err error
	if q.tx != nil {
		rows, err := q.tx.QueryContext(q.ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
		defer rows.Close()

		if q.enableLog {
			q.queryLog = append(q.queryLog, contractsorm.QueryLog{
				Query:    sql,
				Bindings: args,
				Time:     0,
			})
		}

		return q.scanRows(rows, dest)
	}

	dbConn, err := q.DB()
	if err != nil {
		return err
	}

	rows, err := dbConn.QueryContext(q.ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	if q.enableLog {
		q.queryLog = append(q.queryLog, contractsorm.QueryLog{
			Query:    sql,
			Bindings: args,
			Time:     0,
		})
	}

	return q.scanRows(rows, dest)
}

// FirstOrFail retrieves the first record or returns an error if not found.
func (q *Query) FirstOrFail(dest any) error {
	return fmt.Errorf("not implemented")
}

// Get retrieves all records matching the query.
func (q *Query) Get(dest any) error {
	// Build SELECT query
	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// Execute query
	var err error
	if q.tx != nil {
		rows, err := q.tx.QueryContext(q.ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
		defer rows.Close()

		if q.enableLog {
			q.queryLog = append(q.queryLog, contractsorm.QueryLog{
				Query:    sql,
				Bindings: args,
				Time:     0,
			})
		}

		return q.scanRows(rows, dest)
	}

	dbConn, err := q.DB()
	if err != nil {
		return err
	}

	rows, err := dbConn.QueryContext(q.ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	if q.enableLog {
		q.queryLog = append(q.queryLog, contractsorm.QueryLog{
			Query:    sql,
			Bindings: args,
			Time:     0,
		})
	}

	return q.scanRows(rows, dest)
}

// Create inserts a new record into the database.
func (q *Query) Create(value any) error {
	// Fire Creating event if not disabled
	if !q.withoutEvents {
		attributes := observer.ExtractModelAttributes(value)
		if err := q.dispatcher.DispatchCreating(q.ctx, value, q.modelToObserver, nil, attributes, nil, q); err != nil {
			return fmt.Errorf("creating event error: %w", err)
		}
	}

	// Build INSERT query
	builder := NewBuilder(q)
	sql, args := builder.BuildInsert(value)
	if sql == "" {
		return fmt.Errorf("failed to build INSERT query")
	}

	// Execute query
	var err error
	if q.tx != nil {
		_, err = q.tx.ExecContext(q.ctx, sql, args...)
	} else {
		dbConn, err := q.DB()
		if err != nil {
			return err
		}
		_, err = dbConn.ExecContext(q.ctx, sql, args...)
	}

	if err != nil {
		return fmt.Errorf("failed to execute INSERT query: %w", err)
	}

	// Log query if enabled
	if q.enableLog {
		q.queryLog = append(q.queryLog, contractsorm.QueryLog{
			Query:    sql,
			Bindings: args,
			Time:     0,
		})
	}

	// Fire Created event if not disabled
	if !q.withoutEvents {
		attributes := observer.ExtractModelAttributes(value)
		if err := q.dispatcher.DispatchCreated(q.ctx, value, q.modelToObserver, nil, attributes, nil, q); err != nil {
			return fmt.Errorf("created event error: %w", err)
		}
	}

	return nil
}

// Save saves the model to the database.
func (q *Query) Save(value any) error {
	return fmt.Errorf("not implemented")
}

// SaveQuietly saves the model without firing events.
func (q *Query) SaveQuietly(value any) error {
	return fmt.Errorf("not implemented")
}

// Update updates records in the database.
func (q *Query) Update(column any, value ...any) (*contractsorm.Result, error) {
	// Fire Updating event if not disabled
	if !q.withoutEvents && q.model != nil {
		attributes := observer.ExtractModelAttributes(q.model)
		if err := q.dispatcher.DispatchUpdating(q.ctx, q.model, q.modelToObserver, nil, attributes, nil, q); err != nil {
			return nil, fmt.Errorf("updating event error: %w", err)
		}
	}

	// Build UPDATE query
	builder := NewBuilder(q)
	sql, args := builder.BuildUpdate(column, value...)
	if sql == "" {
		return nil, fmt.Errorf("failed to build UPDATE query")
	}

	// Execute query
	var err error
	var result interface{ RowsAffected() (int64, error) }
	if q.tx != nil {
		result, err = q.tx.ExecContext(q.ctx, sql, args...)
	} else {
		dbConn, err := q.DB()
		if err != nil {
			return nil, err
		}
		result, err = dbConn.ExecContext(q.ctx, sql, args...)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute UPDATE query: %w", err)
	}

	// Log query if enabled
	if q.enableLog {
		q.queryLog = append(q.queryLog, contractsorm.QueryLog{
			Query:    sql,
			Bindings: args,
			Time:     0,
		})
	}

	// Fire Updated event if not disabled
	if !q.withoutEvents && q.model != nil {
		attributes := observer.ExtractModelAttributes(q.model)
		if err := q.dispatcher.DispatchUpdated(q.ctx, q.model, q.modelToObserver, nil, attributes, nil, q); err != nil {
			return nil, fmt.Errorf("updated event error: %w", err)
		}
	}

	// Get affected rows
	rowsAffected, _ := result.RowsAffected()
	return &contractsorm.Result{
		RowsAffected: rowsAffected,
	}, nil
}

// hasSoftDeleteCapability checks if the model has soft delete capability.
func hasSoftDeleteCapability(model any) bool {
	if model == nil {
		return false
	}

	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return false
	}

	// Check for DeletedAt field
	deletedAtField := val.FieldByName("DeletedAt")
	return deletedAtField.IsValid() && deletedAtField.Type() == reflect.TypeOf(&time.Time{})
}

// Delete deletes records from the database.
func (q *Query) Delete(value ...any) (*contractsorm.Result, error) {
	// Fire Deleting event if not disabled
	if !q.withoutEvents && q.model != nil {
		attributes := observer.ExtractModelAttributes(q.model)
		if err := q.dispatcher.DispatchDeleting(q.ctx, q.model, q.modelToObserver, nil, attributes, nil, q); err != nil {
			return nil, fmt.Errorf("deleting event error: %w", err)
		}
	}

	// Check if model has soft delete capability
	useSoftDelete := hasSoftDeleteCapability(q.model)

	var sql string
	var args []any
	var err error

	if useSoftDelete {
		// Use UPDATE to set deleted_at instead of DELETE
		builder := NewBuilder(q)
		now := time.Now()
		sql, args = builder.BuildUpdate("deleted_at", now)
		if sql == "" {
			return nil, fmt.Errorf("failed to build SOFT DELETE query")
		}
	} else {
		// Build DELETE query
		builder := NewBuilder(q)
		sql, args = builder.BuildDelete()
		if sql == "" {
			return nil, fmt.Errorf("failed to build DELETE query")
		}
	}

	// Execute query
	var result interface{ RowsAffected() (int64, error) }
	if q.tx != nil {
		result, err = q.tx.ExecContext(q.ctx, sql, args...)
	} else {
		dbConn, err := q.DB()
		if err != nil {
			return nil, err
		}
		result, err = dbConn.ExecContext(q.ctx, sql, args...)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute DELETE query: %w", err)
	}

	// Log query if enabled
	if q.enableLog {
		q.queryLog = append(q.queryLog, contractsorm.QueryLog{
			Query:    sql,
			Bindings: args,
			Time:     0,
		})
	}

	// Fire Deleted event if not disabled
	if !q.withoutEvents && q.model != nil {
		attributes := observer.ExtractModelAttributes(q.model)
		if err := q.dispatcher.DispatchDeleted(q.ctx, q.model, q.modelToObserver, nil, attributes, nil, q); err != nil {
			return nil, fmt.Errorf("deleted event error: %w", err)
		}
	}

	// Get affected rows
	rowsAffected, _ := result.RowsAffected()
	return &contractsorm.Result{
		RowsAffected: rowsAffected,
	}, nil
}

// InsertGetId inserts a record and returns the ID.
func (q *Query) InsertGetId(values any) (uint, error) {
	// Build INSERT query
	builder := NewBuilder(q)
	sql, args := builder.BuildInsert(values)
	if sql == "" {
		return 0, fmt.Errorf("failed to build INSERT query")
	}

	// Execute query
	var err error
	var result interface{ LastInsertId() (int64, error) }
	if q.tx != nil {
		result, err = q.tx.ExecContext(q.ctx, sql, args...)
	} else {
		dbConn, err := q.DB()
		if err != nil {
			return 0, err
		}
		result, err = dbConn.ExecContext(q.ctx, sql, args...)
	}

	if err != nil {
		return 0, fmt.Errorf("failed to execute INSERT query: %w", err)
	}

	// Log query if enabled
	if q.enableLog {
		q.queryLog = append(q.queryLog, contractsorm.QueryLog{
			Query:    sql,
			Bindings: args,
			Time:     0,
		})
	}

	// Get last insert ID
	lastID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return uint(lastID), nil
}

// FlushQueryLog clears the query log.
func (q *Query) FlushQueryLog() {
	q.queryLog = make([]contractsorm.QueryLog, 0)
}

// GetQueryLog returns the query log.
func (q *Query) GetQueryLog() []contractsorm.QueryLog {
	return q.queryLog
}

// Transaction runs a callback wrapped in a database transaction.
func (q *Query) Transaction(txFunc func(tx contractsorm.Query) error, opts ...*sql.TxOptions) error {
	var txOpts *sql.TxOptions
	if len(opts) > 0 {
		txOpts = opts[0]
	}
	tx, err := q.db.BeginTx(q.ctx, txOpts)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := txFunc(q); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction error: %v, rollback error: %w", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

// WithContext returns a new Query instance with the specified context.
func (q *Query) WithContext(ctx context.Context) contractsorm.Query {
	newQuery := *q
	newQuery.ctx = ctx
	return &newQuery
}

// Observe registers an observer for the given model.
func (q *Query) Observe(model any, observer contractsorm.Observer) {
	q.modelToObserver = append(q.modelToObserver, contractsorm.ModelToObserver{
		Model:    model,
		Observer: observer,
	})
}

// WithoutEvents disables event firing for the query.
func (q *Query) WithoutEvents() contractsorm.Query {
	newQuery := *q
	newQuery.withoutEvents = true
	return &newQuery
}

// Placeholder methods for the Query interface
// These will be implemented in subsequent phases

func (q *Query) Join(query string, args ...any) contractsorm.Query {
	q.joins = append(q.joins, joinClause{_type: "JOIN", query: query, args: args})
	return q
}
func (q *Query) LeftJoin(query string, args ...any) contractsorm.Query {
	q.joins = append(q.joins, joinClause{_type: "LEFT JOIN", query: query, args: args})
	return q
}
func (q *Query) RightJoin(query string, args ...any) contractsorm.Query {
	q.joins = append(q.joins, joinClause{_type: "RIGHT JOIN", query: query, args: args})
	return q
}
func (q *Query) CrossJoin(query string, args ...any) contractsorm.Query {
	q.joins = append(q.joins, joinClause{_type: "CROSS JOIN", query: query, args: args})
	return q
}
func (q *Query) Group(name string) contractsorm.Query {
	q.groups = append(q.groups, name)
	return q
}
func (q *Query) Having(query any, args ...any) contractsorm.Query {
	q.havings = append(q.havings, havingClause{query: fmt.Sprintf("%v", query), args: args})
	return q
}

func (q *Query) WhereIn(column string, values []any) contractsorm.Query {
	q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("%s IN (?)", column), args: []any{values}})
	return q
}
func (q *Query) WhereNotIn(column string, values []any) contractsorm.Query {
	q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("%s NOT IN (?)", column), args: []any{values}})
	return q
}
func (q *Query) OrWhereIn(column string, values []any) contractsorm.Query {
	q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("%s IN (?)", column), args: []any{values}})
	return q
}
func (q *Query) OrWhereNotIn(column string, values []any) contractsorm.Query {
	q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("%s NOT IN (?)", column), args: []any{values}})
	return q
}
func (q *Query) WhereBetween(column string, x, y any) contractsorm.Query {
	q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("%s BETWEEN ? AND ?", column), args: []any{x, y}})
	return q
}
func (q *Query) WhereNotBetween(column string, x, y any) contractsorm.Query {
	q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("%s NOT BETWEEN ? AND ?", column), args: []any{x, y}})
	return q
}
func (q *Query) OrWhereBetween(column string, x, y any) contractsorm.Query {
	q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("%s BETWEEN ? AND ?", column), args: []any{x, y}})
	return q
}
func (q *Query) OrWhereNotBetween(column string, x, y any) contractsorm.Query {
	q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("%s NOT BETWEEN ? AND ?", column), args: []any{x, y}})
	return q
}
func (q *Query) WhereNull(column string) contractsorm.Query {
	q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("%s IS NULL", column), args: nil})
	return q
}
func (q *Query) WhereNotNull(column string) contractsorm.Query {
	q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("%s IS NOT NULL", column), args: nil})
	return q
}
func (q *Query) OrWhereNull(column string) contractsorm.Query {
	q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("%s IS NULL", column), args: nil})
	return q
}
func (q *Query) WhereColumn(first, operator, second string) contractsorm.Query {
	q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("%s %s %s", first, operator, second), args: nil})
	return q
}
func (q *Query) OrWhereColumn(first, operator, second string) contractsorm.Query {
	q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("%s %s %s", first, operator, second), args: nil})
	return q
}
func (q *Query) WhereExists(callback func(contractsorm.Query) contractsorm.Query) contractsorm.Query {
	// TODO: Implement subquery support
	return q
}
func (q *Query) WhereNot(query any, args ...any) contractsorm.Query {
	q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("NOT (%v)", query), args: args})
	return q
}
func (q *Query) OrWhereNot(query any, args ...any) contractsorm.Query {
	q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("NOT (%v)", query), args: args})
	return q
}
func (q *Query) WhereAny(columns []string, operator string, value any) contractsorm.Query {
	// TODO: Implement WhereAny
	return q
}
func (q *Query) WhereAll(columns []string, operator string, value any) contractsorm.Query {
	// TODO: Implement WhereAll
	return q
}
func (q *Query) WhereNone(columns []string, operator string, value any) contractsorm.Query {
	// TODO: Implement WhereNone
	return q
}

func (q *Query) WhereJsonContains(column string, value any) contractsorm.Query {
	q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("JSON_CONTAINS(%s, ?)", column), args: []any{value}})
	return q
}
func (q *Query) OrWhereJsonContains(column string, value any) contractsorm.Query {
	q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("JSON_CONTAINS(%s, ?)", column), args: []any{value}})
	return q
}
func (q *Query) WhereJsonDoesntContain(column string, value any) contractsorm.Query {
	q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("NOT JSON_CONTAINS(%s, ?)", column), args: []any{value}})
	return q
}
func (q *Query) OrWhereJsonDoesntContain(column string, value any) contractsorm.Query {
	q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("NOT JSON_CONTAINS(%s, ?)", column), args: []any{value}})
	return q
}
func (q *Query) WhereJsonContainsKey(column string) contractsorm.Query {
	q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("JSON_CONTAINS_PATH(%s, '$.%s')", column, column), args: nil})
	return q
}
func (q *Query) OrWhereJsonContainsKey(column string) contractsorm.Query {
	q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("JSON_CONTAINS_PATH(%s, '$.%s')", column, column), args: nil})
	return q
}
func (q *Query) WhereJsonDoesntContainKey(column string) contractsorm.Query {
	q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("NOT JSON_CONTAINS_PATH(%s, '$.%s')", column, column), args: nil})
	return q
}
func (q *Query) OrWhereJsonDoesntContainKey(column string) contractsorm.Query {
	q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("NOT JSON_CONTAINS_PATH(%s, '$.%s')", column, column), args: nil})
	return q
}
func (q *Query) WhereJsonLength(column string, operator string, value any) contractsorm.Query {
	q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("JSON_LENGTH(%s) %s ?", column, operator), args: []any{value}})
	return q
}

func (q *Query) Count(count *int64) error {
	// Set aggregate
	q.aggregate = "COUNT"
	q.aggregateCol = "*"

	// Build SELECT query
	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// Execute query
	var err error
	if q.tx != nil {
		err = q.tx.QueryRowContext(q.ctx, sql, args...).Scan(count)
	} else {
		databaseConn, err := q.DB()
		if err != nil {
			return err
		}
		err = databaseConn.QueryRowContext(q.ctx, sql, args...).Scan(count)
	}

	if err != nil {
		return fmt.Errorf("failed to execute COUNT query: %w", err)
	}

	// Log query if enabled
	if q.enableLog {
		q.queryLog = append(q.queryLog, contractsorm.QueryLog{
			Query:    sql,
			Bindings: args,
			Time:     0,
		})
	}

	return nil
}
func (q *Query) Sum(column string, dest any) error {
	// Set aggregate
	q.aggregate = "SUM"
	q.aggregateCol = column

	// Build SELECT query
	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// Execute query
	var err error
	if q.tx != nil {
		err = q.tx.QueryRowContext(q.ctx, sql, args...).Scan(dest)
	} else {
		databaseConn, err := q.DB()
		if err != nil {
			return err
		}
		err = databaseConn.QueryRowContext(q.ctx, sql, args...).Scan(dest)
	}

	if err != nil {
		return fmt.Errorf("failed to execute SUM query: %w", err)
	}

	// Log query if enabled
	if q.enableLog {
		q.queryLog = append(q.queryLog, contractsorm.QueryLog{
			Query:    sql,
			Bindings: args,
			Time:     0,
		})
	}

	return nil
}
func (q *Query) Avg(column string, dest any) error {
	// Set aggregate
	q.aggregate = "AVG"
	q.aggregateCol = column

	// Build SELECT query
	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// Execute query
	var err error
	if q.tx != nil {
		err = q.tx.QueryRowContext(q.ctx, sql, args...).Scan(dest)
	} else {
		databaseConn, err := q.DB()
		if err != nil {
			return err
		}
		err = databaseConn.QueryRowContext(q.ctx, sql, args...).Scan(dest)
	}

	if err != nil {
		return fmt.Errorf("failed to execute AVG query: %w", err)
	}

	// Log query if enabled
	if q.enableLog {
		q.queryLog = append(q.queryLog, contractsorm.QueryLog{
			Query:    sql,
			Bindings: args,
			Time:     0,
		})
	}

	return nil
}
func (q *Query) Min(column string, dest any) error {
	// Set aggregate
	q.aggregate = "MIN"
	q.aggregateCol = column

	// Build SELECT query
	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// Execute query
	var err error
	if q.tx != nil {
		err = q.tx.QueryRowContext(q.ctx, sql, args...).Scan(dest)
	} else {
		databaseConn, err := q.DB()
		if err != nil {
			return err
		}
		err = databaseConn.QueryRowContext(q.ctx, sql, args...).Scan(dest)
	}

	if err != nil {
		return fmt.Errorf("failed to execute MIN query: %w", err)
	}

	// Log query if enabled
	if q.enableLog {
		q.queryLog = append(q.queryLog, contractsorm.QueryLog{
			Query:    sql,
			Bindings: args,
			Time:     0,
		})
	}

	return nil
}
func (q *Query) Max(column string, dest any) error {
	// Set aggregate
	q.aggregate = "MAX"
	q.aggregateCol = column

	// Build SELECT query
	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// Execute query
	var err error
	if q.tx != nil {
		err = q.tx.QueryRowContext(q.ctx, sql, args...).Scan(dest)
	} else {
		databaseConn, err := q.DB()
		if err != nil {
			return err
		}
		err = databaseConn.QueryRowContext(q.ctx, sql, args...).Scan(dest)
	}

	if err != nil {
		return fmt.Errorf("failed to execute MAX query: %w", err)
	}

	// Log query if enabled
	if q.enableLog {
		q.queryLog = append(q.queryLog, contractsorm.QueryLog{
			Query:    sql,
			Bindings: args,
			Time:     0,
		})
	}

	return nil
}
func (q *Query) Exists(exists *bool) error {
	// Set aggregate for EXISTS check
	q.aggregate = "COUNT"
	q.aggregateCol = "1"
	limit := 1
	q.limit = &limit

	// Build SELECT query
	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// Execute query
	var count int64
	var err error
	if q.tx != nil {
		err = q.tx.QueryRowContext(q.ctx, sql, args...).Scan(&count)
	} else {
		databaseConn, err := q.DB()
		if err != nil {
			return err
		}
		err = databaseConn.QueryRowContext(q.ctx, sql, args...).Scan(&count)
	}

	if err != nil {
		return fmt.Errorf("failed to execute EXISTS query: %w", err)
	}

	*exists = count > 0

	// Log query if enabled
	if q.enableLog {
		q.queryLog = append(q.queryLog, contractsorm.QueryLog{
			Query:    sql,
			Bindings: args,
			Time:     0,
		})
	}

	return nil
}

func (q *Query) Pluck(column string, dest any) error {
	// Set select to only the specified column
	q.selects = []any{column}

	// Build SELECT query
	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// Execute query
	var err error
	if q.tx != nil {
		rows, err := q.tx.QueryContext(q.ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("failed to execute PLUCK query: %w", err)
		}
		defer rows.Close()

		return q.pluckRows(rows, dest)
	}

	databaseConn, err := q.DB()
	if err != nil {
		return err
	}

	rows, err := databaseConn.QueryContext(q.ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to execute PLUCK query: %w", err)
	}
	defer rows.Close()

	return q.pluckRows(rows, dest)
}
func (q *Query) Value(column string, dest any) error {
	// Set select to only the specified column and limit to 1
	q.selects = []any{column}
	limit := 1
	q.limit = &limit

	// Build SELECT query
	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// Execute query
	var err error
	if q.tx != nil {
		err = q.tx.QueryRowContext(q.ctx, sql, args...).Scan(dest)
	} else {
		databaseConn, err := q.DB()
		if err != nil {
			return err
		}
		err = databaseConn.QueryRowContext(q.ctx, sql, args...).Scan(dest)
	}

	if err != nil {
		return fmt.Errorf("failed to execute VALUE query: %w", err)
	}

	return nil
}
func (q *Query) Scan(dest any) error {
	// Build SELECT query
	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// Execute query
	var err error
	if q.tx != nil {
		rows, err := q.tx.QueryContext(q.ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("failed to execute SCAN query: %w", err)
		}
		defer rows.Close()

		return q.scanRows(rows, dest)
	}

	databaseConn, err := q.DB()
	if err != nil {
		return err
	}

	rows, err := databaseConn.QueryContext(q.ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to execute SCAN query: %w", err)
	}
	defer rows.Close()

	return q.scanRows(rows, dest)
}
func (q *Query) Chunk(size int, callback any) error {
	// Set limit to chunk size
	q.limit = &size

	// Build SELECT query
	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// Execute query
	var err error
	if q.tx != nil {
		rows, err := q.tx.QueryContext(q.ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("failed to execute CHUNK query: %w", err)
		}
		defer rows.Close()

		return q.chunkRows(rows, size, callback)
	}

	databaseConn, err := q.DB()
	if err != nil {
		return err
	}

	rows, err := databaseConn.QueryContext(q.ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to execute CHUNK query: %w", err)
	}
	defer rows.Close()

	return q.chunkRows(rows, size, callback)
}
func (q *Query) Paginate(page, limit int, dest any, total *int64) error {
	// Calculate offset
	offset := (page - 1) * limit
	q.offset = &offset
	q.limit = &limit

	// Get total count first
	countQuery := *q
	countQuery.limit = nil
	countQuery.offset = nil
	var count int64
	if err := countQuery.Count(&count); err != nil {
		return fmt.Errorf("failed to get total count: %w", err)
	}
	if total != nil {
		*total = count
	}

	// Build SELECT query for paginated results
	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// Execute query
	var err error
	if q.tx != nil {
		rows, err := q.tx.QueryContext(q.ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("failed to execute PAGINATE query: %w", err)
		}
		defer rows.Close()

		return q.scanRows(rows, dest)
	}

	databaseConn, err := q.DB()
	if err != nil {
		return err
	}

	rows, err := databaseConn.QueryContext(q.ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to execute PAGINATE query: %w", err)
	}
	defer rows.Close()

	return q.scanRows(rows, dest)
}

func (q *Query) FirstOr(dest any, callback func() error) error {
	err := q.First(dest)
	if err != nil {
		return callback()
	}
	return nil
}
func (q *Query) FirstOrCreate(dest any, conds ...any) error {
	// Try to find the record first
	err := q.First(dest)
	if err == nil {
		return nil // Record exists
	}

	// Record doesn't exist, create it
	return q.Create(dest)
}
func (q *Query) FirstOrNew(dest any, attributes any, values ...any) error {
	// Try to find the record first
	err := q.First(dest)
	if err == nil {
		return nil // Record exists
	}

	// Record doesn't exist, prepare new instance (without saving)
	// This is a simplified implementation
	return nil
}
func (q *Query) UpdateOrCreate(dest any, attributes any, values any) error {
	// Try to find the record first
	err := q.First(dest)
	if err == nil {
		// Record exists, update it
		return q.Save(values)
	}

	// Record doesn't exist, create it
	return q.Create(values)
}
func (q *Query) UpdateOrInsert(attributes any, values any) error {
	// Try to find the record first
	count := int64(0)
	if err := q.Count(&count); err != nil {
		return err
	}

	if count > 0 {
		// Record exists, update it
		_, err := q.Update(values)
		return err
	}

	// Record doesn't exist, create it
	return q.Create(values)
}
func (q *Query) Increment(column string, amount ...any) (*contractsorm.Result, error) {
	incAmount := int64(1)
	if len(amount) > 0 {
		if val, ok := amount[0].(int64); ok {
			incAmount = val
		}
	}

	updateQuery := fmt.Sprintf("%s = %s + ?", column, column)
	return q.Update(updateQuery, incAmount)
}
func (q *Query) Decrement(column string, amount ...any) (*contractsorm.Result, error) {
	decAmount := int64(1)
	if len(amount) > 0 {
		if val, ok := amount[0].(int64); ok {
			decAmount = val
		}
	}

	updateQuery := fmt.Sprintf("%s = %s - ?", column, column)
	return q.Update(updateQuery, decAmount)
}
func (q *Query) InRandomOrder() contractsorm.Query {
	// Add random ordering
	q.orders = append(q.orders, orderClause{column: "RANDOM()", direction: ""})
	return q
}
func (q *Query) LockForUpdate() contractsorm.Query {
	// Add FOR UPDATE clause (simplified - would need dialect-specific handling)
	return q
}
func (q *Query) SharedLock() contractsorm.Query {
	// Add FOR SHARE clause (simplified - would need dialect-specific handling)
	return q
}
func (q *Query) Raw(sql string, values ...any) contractsorm.Query {
	// Execute raw SQL (simplified implementation)
	return q
}
func (q *Query) Exec(sql string, values ...any) (*contractsorm.Result, error) {
	// Execute raw SQL
	var err error
	var result interface{ RowsAffected() (int64, error) }
	if q.tx != nil {
		result, err = q.tx.ExecContext(q.ctx, sql, values...)
	} else {
		databaseConn, err := q.DB()
		if err != nil {
			return nil, err
		}
		result, err = databaseConn.ExecContext(q.ctx, sql, values...)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute raw SQL: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	return &contractsorm.Result{
		RowsAffected: rowsAffected,
	}, nil
}

func (q *Query) WithTrashed() contractsorm.Query {
	newQuery := *q
	newQuery.withTrashed = true
	newQuery.onlyTrashed = false
	newQuery.withoutTrashed = false
	return &newQuery
}

func (q *Query) OnlyTrashed() contractsorm.Query {
	newQuery := *q
	newQuery.withTrashed = false
	newQuery.onlyTrashed = true
	newQuery.withoutTrashed = false
	return &newQuery
}

func (q *Query) WithoutTrashed() contractsorm.Query {
	newQuery := *q
	newQuery.withTrashed = false
	newQuery.onlyTrashed = false
	newQuery.withoutTrashed = true
	return &newQuery
}

func (q *Query) Omit(columns ...string) contractsorm.Query { return q }

func (q *Query) Restore(model ...any) (*contractsorm.Result, error) {
	// Fire Restoring event if not disabled
	if !q.withoutEvents && len(model) > 0 {
		attributes := observer.ExtractModelAttributes(model[0])
		if err := q.dispatcher.DispatchRestoring(q.ctx, model[0], q.modelToObserver, nil, attributes, nil, q); err != nil {
			return nil, fmt.Errorf("restoring event error: %w", err)
		}
	}

	// Build UPDATE query to set deleted_at to NULL
	builder := NewBuilder(q)
	// This is a simplified implementation - in a real implementation, we'd need to
	// properly handle the update to set deleted_at = NULL
	sql, args := builder.BuildUpdate("deleted_at", nil)
	if sql == "" {
		return nil, fmt.Errorf("failed to build RESTORE query")
	}

	// Execute query
	var err error
	var result interface{ RowsAffected() (int64, error) }
	if q.tx != nil {
		result, err = q.tx.ExecContext(q.ctx, sql, args...)
	} else {
		dbConn, err := q.DB()
		if err != nil {
			return nil, err
		}
		result, err = dbConn.ExecContext(q.ctx, sql, args...)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute RESTORE query: %w", err)
	}

	// Log query if enabled
	if q.enableLog {
		q.queryLog = append(q.queryLog, contractsorm.QueryLog{
			Query:    sql,
			Bindings: args,
			Time:     0,
		})
	}

	// Fire Restored event if not disabled
	if !q.withoutEvents && len(model) > 0 {
		attributes := observer.ExtractModelAttributes(model[0])
		if err := q.dispatcher.DispatchRestored(q.ctx, model[0], q.modelToObserver, nil, attributes, nil, q); err != nil {
			return nil, fmt.Errorf("restored event error: %w", err)
		}
	}

	// Get affected rows
	rowsAffected, _ := result.RowsAffected()
	return &contractsorm.Result{
		RowsAffected: rowsAffected,
	}, nil
}

func (q *Query) ForceDelete(value ...any) (*contractsorm.Result, error) {
	// Fire ForceDeleting event if not disabled
	if !q.withoutEvents && q.model != nil {
		attributes := observer.ExtractModelAttributes(q.model)
		if err := q.dispatcher.DispatchForceDeleting(q.ctx, q.model, q.modelToObserver, nil, attributes, nil, q); err != nil {
			return nil, fmt.Errorf("force_deleting event error: %w", err)
		}
	}

	// Build DELETE query (permanent delete, not soft delete)
	builder := NewBuilder(q)
	sql, args := builder.BuildDelete()
	if sql == "" {
		return nil, fmt.Errorf("failed to build FORCE DELETE query")
	}

	// Execute query
	var err error
	var result interface{ RowsAffected() (int64, error) }
	if q.tx != nil {
		result, err = q.tx.ExecContext(q.ctx, sql, args...)
	} else {
		dbConn, err := q.DB()
		if err != nil {
			return nil, err
		}
		result, err = dbConn.ExecContext(q.ctx, sql, args...)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute FORCE DELETE query: %w", err)
	}

	// Log query if enabled
	if q.enableLog {
		q.queryLog = append(q.queryLog, contractsorm.QueryLog{
			Query:    sql,
			Bindings: args,
			Time:     0,
		})
	}

	// Fire ForceDeleted event if not disabled
	if !q.withoutEvents && q.model != nil {
		attributes := observer.ExtractModelAttributes(q.model)
		if err := q.dispatcher.DispatchForceDeleted(q.ctx, q.model, q.modelToObserver, nil, attributes, nil, q); err != nil {
			return nil, fmt.Errorf("force_deleted event error: %w", err)
		}
	}

	// Get affected rows
	rowsAffected, _ := result.RowsAffected()
	return &contractsorm.Result{
		RowsAffected: rowsAffected,
	}, nil
}

func (q *Query) With(query string, args ...any) contractsorm.Query {
	newQuery := *q
	newQuery.withRelations = append(newQuery.withRelations, query)
	return &newQuery
}

func (q *Query) Load(dest any, relation string, args ...any) error {
	// This is a simplified implementation - full lazy loading requires
	// additional work to detect relationships and load them properly
	return fmt.Errorf("lazy loading not fully implemented yet")
}

func (q *Query) LoadMissing(dest any, relation string, args ...any) error {
	// This is a simplified implementation - full lazy loading requires
	// additional work to detect relationships and load them properly
	return fmt.Errorf("lazy loading not fully implemented yet")
}
func (q *Query) Without(relations ...string) contractsorm.Query          { return q }
func (q *Query) WithCount(query string, args ...any) contractsorm.Query  { return q }
func (q *Query) WithExists(query string, args ...any) contractsorm.Query { return q }
func (q *Query) Association(assocName string) contractsorm.Association {
	// Return a base association - specific relationship types should be created
	// based on the relationship metadata from the model
	return association.NewAssociation(q, q.model, assocName)
}

func (q *Query) Begin(opts ...*sql.TxOptions) (contractsorm.Query, error) {
	var txOpts *sql.TxOptions
	if len(opts) > 0 {
		txOpts = opts[0]
	}

	var tx *sql.Tx
	var err error
	if q.tx != nil {
		// Already in a transaction, return current query
		return q, nil
	}

	dbConn, err := q.DB()
	if err != nil {
		return nil, err
	}

	tx, err = dbConn.BeginTx(q.ctx, txOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Create a new query instance with the transaction
	newQuery := *q
	newQuery.tx = tx
	newQuery.inTransaction = true
	return &newQuery, nil
}
func (q *Query) Commit() error {
	if !q.inTransaction || q.tx == nil {
		return fmt.Errorf("not in a transaction")
	}

	err := q.tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	q.inTransaction = false
	q.tx = nil
	return nil
}
func (q *Query) Rollback() error {
	if !q.inTransaction || q.tx == nil {
		return fmt.Errorf("not in a transaction")
	}

	err := q.tx.Rollback()
	if err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}

	q.inTransaction = false
	q.tx = nil
	return nil
}
func (q *Query) RollbackTo(level string) error {
	if !q.inTransaction || q.tx == nil {
		return fmt.Errorf("not in a transaction")
	}

	// Execute savepoint rollback (dialect-specific)
	_, err := q.tx.ExecContext(q.ctx, fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", level))
	if err != nil {
		return fmt.Errorf("failed to rollback to savepoint: %w", err)
	}

	return nil
}
func (q *Query) SavePoint(name string) error {
	if !q.inTransaction || q.tx == nil {
		return fmt.Errorf("not in a transaction")
	}

	// Execute savepoint creation (dialect-specific)
	_, err := q.tx.ExecContext(q.ctx, fmt.Sprintf("SAVEPOINT %s", name))
	if err != nil {
		return fmt.Errorf("failed to create savepoint: %w", err)
	}

	return nil
}
func (q *Query) Scopes(scopes ...func(contractsorm.Query) contractsorm.Query) contractsorm.Query {
	return q
}

func (q *Query) Cursor() (chan contractsorm.Cursor, error) {
	// Build SELECT query
	builder := NewBuilder(q)
	querySQL, args := builder.BuildSelect()

	// Execute query
	var err error
	var rows *sql.Rows
	if q.tx != nil {
		rows, err = q.tx.QueryContext(q.ctx, querySQL, args...)
	} else {
		databaseConn, err := q.DB()
		if err != nil {
			return nil, err
		}
		rows, err = databaseConn.QueryContext(q.ctx, querySQL, args...)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute CURSOR query: %w", err)
	}

	// Create channel for cursors
	cursorChan := make(chan contractsorm.Cursor, 1)

	// Create cursor and send to channel
	dbCursor := cursor.NewCursor(rows)
	cursorChan <- dbCursor

	// Close channel after sending
	close(cursorChan)

	return cursorChan, nil
}

// Transaction callback methods (stubs for now)
func (q *Query) BeforeCommit(callback func() error)   {}
func (q *Query) AfterCommit(callback func() error)    {}
func (q *Query) BeforeRollback(callback func() error) {}
func (q *Query) AfterRollback(callback func() error)  {}

// scanRows scans database rows into the destination.
func (q *Query) scanRows(rows *sql.Rows, dest any) error {
	// Use reflection to handle different destination types
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer")
	}

	destValue = destValue.Elem()

	// Handle slice destination
	if destValue.Kind() == reflect.Slice {
		sliceType := destValue.Type()
		elemType := sliceType.Elem()

		for rows.Next() {
			// Create new element
			elemPtr := reflect.New(elemType)
			elem := elemPtr.Elem()

			// Scan into element
			columns, err := rows.Columns()
			if err != nil {
				return fmt.Errorf("failed to get columns: %w", err)
			}

			values := make([]any, len(columns))
			for i := range values {
				// Get the field from the struct
				if elem.Kind() == reflect.Struct {
					if i < elem.NumField() {
						field := elem.Field(i)
						if field.CanAddr() {
							values[i] = field.Addr().Interface()
						} else {
							// Create a pointer to scan into
							ptr := reflect.New(field.Type())
							values[i] = ptr.Interface()
						}
					} else {
						// Use a placeholder for extra columns
						var placeholder any
						values[i] = &placeholder
					}
				} else {
					// For non-struct types, scan directly
					values[i] = elem.Addr().Interface()
				}
			}

			if err := rows.Scan(values...); err != nil {
				return fmt.Errorf("failed to scan row: %w", err)
			}

			// Append to slice
			destValue.Set(reflect.Append(destValue, elem))
		}

		return rows.Err()
	}

	// Handle single struct destination
	if destValue.Kind() == reflect.Struct {
		if !rows.Next() {
			return fmt.Errorf("no rows found")
		}

		columns, err := rows.Columns()
		if err != nil {
			return fmt.Errorf("failed to get columns: %w", err)
		}

		values := make([]any, len(columns))
		for i := range values {
			if i < destValue.NumField() {
				field := destValue.Field(i)
				if field.CanAddr() {
					values[i] = field.Addr().Interface()
				} else {
					var placeholder any
					values[i] = &placeholder
				}
			} else {
				var placeholder any
				values[i] = &placeholder
			}
		}

		if err := rows.Scan(values...); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		return rows.Err()
	}

	return fmt.Errorf("unsupported destination type: %T", dest)
}

// pluckRows scans a single column from database rows into the destination.
func (q *Query) pluckRows(rows *sql.Rows, dest any) error {
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer")
	}

	destValue = destValue.Elem()

	// Handle slice destination
	if destValue.Kind() == reflect.Slice {
		elemType := destValue.Type().Elem()

		for rows.Next() {
			// Create new element
			elemPtr := reflect.New(elemType)
			elem := elemPtr.Elem()

			if err := rows.Scan(elem.Addr().Interface()); err != nil {
				return fmt.Errorf("failed to scan row: %w", err)
			}

			// Append to slice
			destValue.Set(reflect.Append(destValue, elem))
		}

		return rows.Err()
	}

	return fmt.Errorf("unsupported destination type for PLUCK: %T", dest)
}

// chunkRows processes rows in chunks and calls the callback for each chunk.
func (q *Query) chunkRows(rows *sql.Rows, size int, callback any) error {
	// Use reflection to call the callback
	callbackValue := reflect.ValueOf(callback)
	if callbackValue.Kind() != reflect.Func {
		return fmt.Errorf("callback must be a function")
	}

	// Process rows in chunks
	chunk := make([]any, 0, size)
	for rows.Next() {
		// Scan row into a map or struct (simplified for now)
		var row map[string]any
		columns, err := rows.Columns()
		if err != nil {
			return fmt.Errorf("failed to get columns: %w", err)
		}

		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		row = make(map[string]any)
		for i, col := range columns {
			row[col] = values[i]
		}

		chunk = append(chunk, row)

		// Call callback when chunk is full
		if len(chunk) >= size {
			results := callbackValue.Call([]reflect.Value{reflect.ValueOf(chunk)})
			if len(results) > 0 {
				if err, ok := results[0].Interface().(error); ok && err != nil {
					return err
				}
			}
			chunk = make([]any, 0, size)
		}
	}

	// Process remaining rows in the last chunk
	if len(chunk) > 0 {
		results := callbackValue.Call([]reflect.Value{reflect.ValueOf(chunk)})
		if len(results) > 0 {
			if err, ok := results[0].Interface().(error); ok && err != nil {
				return err
			}
		}
	}

	return rows.Err()
}
