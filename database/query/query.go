package query

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/dracory/neat/contracts/database"
	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/contracts/log"
	"github.com/dracory/neat/database/association"
	"github.com/dracory/neat/database/cursor"
	"github.com/dracory/neat/database/db"
	"github.com/dracory/neat/database/driver"
	"github.com/dracory/neat/database/observer"
	"github.com/dracory/neat/support/str"
)

// Query implements the Query interface using native database/sql.
type Query struct {
	ctx        context.Context
	db         *sql.DB // primary (write) connection
	readDB     *sql.DB // optional read-replica connection; nil means use db
	writeDB    *sql.DB // optional explicit write connection; nil means use db
	driver     driver.Driver
	connection string
	dbConfig   *db.DBConfig
	log        log.Log
	queryLog   []contractsorm.QueryLog
	enableLog  bool

	// Query state
	table        string
	model        any
	selects      []selectClause
	wheres       []whereClause
	joins        []joinClause
	groups       []string
	havings      []havingClause
	orders       []orderClause
	limit        *int
	offset       *int
	distinct     bool
	distinctCols []string
	aggregate    string
	aggregateCol string

	// Transaction state
	inTransaction bool
	tx            *sql.Tx

	// Observer state
	modelToObserver []contractsorm.ModelToObserver
	withoutEvents   bool
	dispatcher      *observer.Dispatcher

	// Transaction lifecycle hooks
	beforeCommit   []func() error
	afterCommit    []func() error
	beforeRollback []func() error
	afterRollback  []func() error

	// Raw SQL state
	rawSQL  string
	rawArgs []any

	// Lock state
	lockForUpdate bool
	sharedLock    bool

	// Scopes
	scopes []func(contractsorm.Query) contractsorm.Query

	// Omit columns
	omitColumns []string

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

type selectClause struct {
	expr string
	args []any
}

// RawExpression represents a raw SQL expression that can be used as a value in maps
type RawExpression struct {
	SQL  string
	Args []any
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
		readDB:          nil,
		writeDB:         nil,
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

// NewQueryWithReplicas creates a Query with separate read and write sql.DB connections.
func NewQueryWithReplicas(ctx context.Context, writeConn, readConn *sql.DB, drv driver.Driver, connection string, dbConfig *db.DBConfig, lg log.Log) *Query {
	q := NewQuery(ctx, writeConn, drv, connection, dbConfig, lg)
	q.readDB = readConn
	q.writeDB = writeConn
	return q
}

// readConn returns the connection to use for read (SELECT) queries.
func (q *Query) readConn() *sql.DB {
	if q.readDB != nil {
		return q.readDB
	}
	return q.db
}

// writeConn returns the connection to use for write (INSERT/UPDATE/DELETE) queries.
func (q *Query) writeConn() *sql.DB {
	if q.writeDB != nil {
		return q.writeDB
	}
	return q.db
}

// Clone returns a new Query with shared connection state but empty query-builder state.
func (q *Query) Clone() contractsorm.Query {
	clone := NewQuery(q.ctx, q.db, q.driver, q.connection, q.dbConfig, q.log)
	clone.readDB = q.readDB
	clone.writeDB = q.writeDB
	clone.table = q.table
	clone.model = q.model
	clone.omitColumns = q.omitColumns
	clone.distinct = q.distinct
	clone.distinctCols = q.distinctCols
	clone.wheres = q.wheres
	return clone
}

// applyScopes applies registered scope functions and returns the modified query.
func (q *Query) applyScopes() *Query {
	if len(q.scopes) == 0 {
		return q
	}
	var result contractsorm.Query = q
	for _, fn := range q.scopes {
		result = fn(result)
	}
	if r, ok := result.(*Query); ok {
		return r
	}
	return q
}

// Connection returns a new Query instance scoped to the named connection.
func (q *Query) Connection(name string) contractsorm.Query {
	if name == "" || q.dbConfig == nil {
		return q
	}
	connCfg, ok := q.dbConfig.Connections[name]
	if !ok {
		return q
	}
	drv := newDriverForDialect(connCfg.Driver)
	dsn, err := db.NewConfigBuilder(connCfg).BuildDSN()
	if err != nil {
		return q
	}
	sqlDB, err := drv.Open(dsn)
	if err != nil {
		return q
	}
	return NewQuery(q.ctx, sqlDB, drv, name, q.dbConfig, q.log)
}

// newDriverForDialect returns a Driver for the given dialect name.
func newDriverForDialect(dialect string) driver.Driver {
	switch dialect {
	case "mysql":
		return driver.NewMySQL()
	case "postgres":
		return driver.NewPostgreSQL()
	case "sqlserver":
		return driver.NewSQLServer()
	case "turso":
		return driver.NewTurso()
	default:
		return driver.NewSQLite()
	}
}

// Model sets the model for the query.
func (q *Query) Model(value any) contractsorm.Query {
	q.model = value
	if q.table == "" {
		q.table = q.resolveTableName(value)
	}
	return q
}

// initializeRelations initializes association fields for relations requested via With.
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

// resolveTableName resolves the table name from the model.
func (q *Query) resolveTableName(model any) string {
	if model == nil {
		return ""
	}

	v := reflect.ValueOf(model)
	if v.Kind() == reflect.Ptr {
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
		if t.Kind() == reflect.Ptr {
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

// Table sets the table for the query.
func (q *Query) Table(name string, args ...any) contractsorm.Query {
	q.table = name
	return q
}

// Select adds columns to the select clause.
func (q *Query) Select(query any, args ...any) contractsorm.Query {
	// Process args to handle func(Query)Query callbacks
	processedArgs := make([]any, 0, len(args))
	for _, arg := range args {
		// Check if arg is a func(Query)Query callback
		if fn, ok := arg.(func(contractsorm.Query) contractsorm.Query); ok {
			// Invoke the callback with a clone of the current query
			subQuery := fn(q.Clone())
			// Build the subquery SQL
			builder := NewBuilder(subQuery.(*Query))
			subSQL, subArgs := builder.BuildSelect()
			// Replace the callback with the subquery SQL
			processedArgs = append(processedArgs, fmt.Sprintf("(%s)", subSQL))
			// Append subquery args
			processedArgs = append(processedArgs, subArgs...)
		} else {
			processedArgs = append(processedArgs, arg)
		}
	}
	q.selects = append(q.selects, selectClause{expr: fmt.Sprintf("%v", query), args: processedArgs})
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
	expr := fmt.Sprintf("%v", value)
	// Check if the expression already contains a direction keyword
	upperExpr := strings.ToUpper(expr)
	if strings.Contains(upperExpr, " DESC") {
		// Remove the direction from the expression and set it properly
		expr = strings.TrimSuffix(expr, " DESC")
		expr = strings.TrimSuffix(expr, " desc")
		q.orders = append(q.orders, orderClause{column: expr, direction: "desc"})
	} else if strings.Contains(upperExpr, " ASC") {
		// Remove the direction from the expression and set it properly
		expr = strings.TrimSuffix(expr, " ASC")
		expr = strings.TrimSuffix(expr, " asc")
		q.orders = append(q.orders, orderClause{column: expr, direction: "asc"})
	} else {
		// No direction specified, default to asc
		q.orders = append(q.orders, orderClause{column: expr, direction: "asc"})
	}
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
	if len(args) > 0 {
		q.distinctCols = make([]string, 0)
		for _, arg := range args {
			q.distinctCols = append(q.distinctCols, fmt.Sprintf("%v", arg))
		}
	}
	return q
}

// DB returns the write (primary) database connection.
func (q *Query) DB() (*sql.DB, error) {
	if q.tx != nil {
		return nil, fmt.Errorf("cannot get DB during transaction, use transaction methods instead")
	}
	return q.writeConn(), nil
}

// ReadDB returns the read-replica connection (falls back to primary if none configured).
func (q *Query) ReadDB() (*sql.DB, error) {
	if q.tx != nil {
		return nil, fmt.Errorf("cannot get ReadDB during transaction")
	}
	return q.readConn(), nil
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
	q = q.applyScopes()
	// Add conditions to where clause
	for _, cond := range conds {
		q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("%v", cond), args: nil})
	}

	// Build SELECT query
	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	start := time.Now()
	// Execute query
	if q.tx != nil {
		rows, err := q.tx.QueryContext(q.ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
		defer rows.Close()
		q.logQuery(sql, args, start)
		return q.scanRows(rows, dest)
	}

	dbConn, err := q.ReadDB()
	if err != nil {
		return err
	}

	rows, err := dbConn.QueryContext(q.ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()
	q.logQuery(sql, args, start)
	return q.scanRows(rows, dest)
}

// FindOrFail retrieves records matching the given conditions or returns an error if not found.
func (q *Query) FindOrFail(dest any, conds ...any) error {
	if err := q.Find(dest, conds...); err != nil {
		return err
	}
	// For slice destinations, empty result is a failure
	v := reflect.Indirect(reflect.ValueOf(dest))
	if v.Kind() == reflect.Slice && v.Len() == 0 {
		return fmt.Errorf("record not found")
	}
	return nil
}

// First retrieves the first record matching the query.
func (q *Query) First(dest any) error {
	q = q.applyScopes()
	// Set limit to 1
	limit := 1
	q.limit = &limit

	// Build SELECT query
	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	start := time.Now()
	if q.tx != nil {
		rows, err := q.tx.QueryContext(q.ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
		defer rows.Close()
		q.logQuery(sql, args, start)
		return q.scanRows(rows, dest)
	}

	dbConn, err := q.ReadDB()
	if err != nil {
		return err
	}

	rows, err := dbConn.QueryContext(q.ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()
	q.logQuery(sql, args, start)
	return q.scanRows(rows, dest)
}

// FirstOrFail retrieves the first record or returns an error if not found.
func (q *Query) FirstOrFail(dest any) error {
	if err := q.First(dest); err != nil {
		return err
	}
	// If the dest is a struct and still zero, nothing was found
	v := reflect.Indirect(reflect.ValueOf(dest))
	if v.Kind() == reflect.Struct && v.IsZero() {
		return fmt.Errorf("record not found")
	}
	return nil
}

// Get retrieves all records matching the query.
func (q *Query) Get(dest any) error {
	q = q.applyScopes()
	// Build SELECT query
	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	start := time.Now()
	if q.tx != nil {
		rows, err := q.tx.QueryContext(q.ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
		defer rows.Close()
		q.logQuery(sql, args, start)
		return q.scanRows(rows, dest)
	}

	dbConn, err := q.ReadDB()
	if err != nil {
		return err
	}

	rows, err := dbConn.QueryContext(q.ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()
	q.logQuery(sql, args, start)
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
	sqlStr, args := builder.BuildInsert(value)
	if sqlStr == "" {
		return fmt.Errorf("failed to build INSERT query")
	}

	// Execute query
	var result sql.Result
	var err error
	start := time.Now()
	if q.tx != nil {
		result, err = q.tx.ExecContext(q.ctx, sqlStr, args...)
	} else {
		var dbConn *sql.DB
		dbConn, err = q.DB()
		if err != nil {
			return err
		}
		result, err = dbConn.ExecContext(q.ctx, sqlStr, args...)
	}

	if err != nil {
		return fmt.Errorf("failed to execute INSERT query: %w", err)
	}
	q.logQuery(sqlStr, args, start)

	// Populate last insert ID back into the model's primary key field
	if lastID, err := result.LastInsertId(); err == nil && lastID > 0 {
		setModelPrimaryKey(value, lastID)
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

// Save saves the model to the database (INSERT if no primary key, UPDATE otherwise).
func (q *Query) Save(value any) error {
	// Fire Saving event
	if !q.withoutEvents {
		attributes := observer.ExtractModelAttributes(value)
		if err := q.dispatcher.DispatchSaving(q.ctx, value, q.modelToObserver, nil, attributes, nil, q); err != nil {
			return fmt.Errorf("saving event error: %w", err)
		}
	}

	id := getPrimaryKeyValue(value)
	var saveErr error
	if id != 0 {
		// UPDATE: set WHERE id = <id> on a clone, then call Update with the value
		clone := q.Clone().(*Query)
		clone.wheres = append(clone.wheres, whereClause{_type: "and", query: "id = ?", args: []any{id}})
		_, saveErr = clone.Update(value)
	} else {
		saveErr = q.Create(value)
	}

	if saveErr != nil {
		return saveErr
	}

	// Fire Saved event
	if !q.withoutEvents {
		attributes := observer.ExtractModelAttributes(value)
		if err := q.dispatcher.DispatchSaved(q.ctx, value, q.modelToObserver, nil, attributes, nil, q); err != nil {
			return fmt.Errorf("saved event error: %w", err)
		}
	}
	return nil
}

// SaveQuietly saves the model without firing events.
func (q *Query) SaveQuietly(value any) error {
	return q.WithoutEvents().(*Query).Save(value)
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
	sqlStr, args := builder.BuildUpdate(column, value...)
	if sqlStr == "" {
		return nil, fmt.Errorf("failed to build UPDATE query")
	}

	// Execute query
	var err error
	var result sql.Result
	start := time.Now()
	if q.tx != nil {
		result, err = q.tx.ExecContext(q.ctx, sqlStr, args...)
	} else {
		var dbConn *sql.DB
		dbConn, err = q.DB()
		if err != nil {
			return nil, err
		}
		result, err = dbConn.ExecContext(q.ctx, sqlStr, args...)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute UPDATE query: %w", err)
	}
	q.logQuery(sqlStr, args, start)

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

	var deleteSQL string
	var args []any
	var err error

	if useSoftDelete {
		// Use UPDATE to set deleted_at instead of DELETE
		builder := NewBuilder(q)
		now := time.Now()
		deleteSQL, args = builder.BuildUpdate("deleted_at", now)
		if deleteSQL == "" {
			return nil, fmt.Errorf("failed to build SOFT DELETE query")
		}
	} else {
		// Build DELETE query
		builder := NewBuilder(q)
		deleteSQL, args = builder.BuildDelete()
		if deleteSQL == "" {
			return nil, fmt.Errorf("failed to build DELETE query")
		}
	}

	// Execute query
	var result interface{ RowsAffected() (int64, error) }
	start := time.Now()
	if q.tx != nil {
		result, err = q.tx.ExecContext(q.ctx, deleteSQL, args...)
	} else {
		var dbConn *sql.DB
		dbConn, err = q.DB()
		if err != nil {
			return nil, err
		}
		result, err = dbConn.ExecContext(q.ctx, deleteSQL, args...)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute DELETE query: %w", err)
	}
	q.logQuery(deleteSQL, args, start)

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
	insertSQL, args := builder.BuildInsert(values)
	if insertSQL == "" {
		return 0, fmt.Errorf("failed to build INSERT query")
	}

	// Postgres: use RETURNING id to get inserted ID
	isPostgres := q.driver != nil && q.driver.Dialect() == "postgres"
	if isPostgres {
		insertSQL = insertSQL + " RETURNING id"
	}

	start := time.Now()
	var lastID int64

	if isPostgres {
		var row *sql.Row
		if q.tx != nil {
			row = q.tx.QueryRowContext(q.ctx, insertSQL, args...)
		} else {
			dbConn, err := q.DB()
			if err != nil {
				return 0, err
			}
			row = dbConn.QueryRowContext(q.ctx, insertSQL, args...)
		}
		if err := row.Scan(&lastID); err != nil {
			return 0, fmt.Errorf("failed to get inserted ID: %w", err)
		}
	} else {
		var err error
		var result sql.Result
		if q.tx != nil {
			result, err = q.tx.ExecContext(q.ctx, insertSQL, args...)
		} else {
			dbConn, err2 := q.DB()
			if err2 != nil {
				return 0, err2
			}
			result, err = dbConn.ExecContext(q.ctx, insertSQL, args...)
		}
		if err != nil {
			return 0, fmt.Errorf("failed to execute INSERT query: %w", err)
		}
		lastID, err = result.LastInsertId()
		if err != nil {
			return 0, fmt.Errorf("failed to get last insert ID: %w", err)
		}
	}

	q.logQuery(insertSQL, args, start)

	// Write the ID back to the struct if it's a pointer-to-struct
	setModelPrimaryKey(values, lastID)

	return uint(lastID), nil
}

// logQuery appends a QueryLog entry with the actual execution duration.
// It also emits a warning via the logger when SlowThreshold is configured and exceeded.
func (q *Query) logQuery(sql string, bindings []any, start time.Time) {
	elapsed := float64(time.Since(start).Milliseconds())
	if q.enableLog {
		q.queryLog = append(q.queryLog, contractsorm.QueryLog{
			Query:    sql,
			Bindings: bindings,
			Time:     elapsed,
		})
	}
	// Slow-query warning
	if q.dbConfig != nil && q.dbConfig.SlowThreshold > 0 && elapsed >= float64(q.dbConfig.SlowThreshold) {
		if q.log != nil {
			q.log.Warningf("[slow query %.1fms] %s %v", elapsed, sql, bindings)
		}
	}
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
	txQuery, err := q.Begin(opts...)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	txQ := txQuery.(*Query)

	defer func() {
		if p := recover(); p != nil {
			txQ.doRollback()
			panic(p)
		}
	}()

	if err := txFunc(txQuery); err != nil {
		if rbErr := txQ.doRollback(); rbErr != nil {
			return fmt.Errorf("transaction error: %v, rollback error: %w", err, rbErr)
		}
		return err
	}

	return txQ.doCommit()
}

// doCommit runs beforeCommit hooks, commits, then runs afterCommit hooks.
func (q *Query) doCommit() error {
	for _, cb := range q.beforeCommit {
		if err := cb(); err != nil {
			_ = q.tx.Rollback()
			return fmt.Errorf("beforeCommit hook error: %w", err)
		}
	}
	if err := q.tx.Commit(); err != nil {
		return err
	}
	for _, cb := range q.afterCommit {
		if err := cb(); err != nil {
			return fmt.Errorf("afterCommit hook error: %w", err)
		}
	}
	return nil
}

// doRollback runs beforeRollback hooks, rolls back, then runs afterRollback hooks.
func (q *Query) doRollback() error {
	for _, cb := range q.beforeRollback {
		_ = cb()
	}
	err := q.tx.Rollback()
	for _, cb := range q.afterRollback {
		_ = cb()
	}
	return err
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
	// Process args to handle func(Query)Query callbacks
	queryStr := fmt.Sprintf("%v", query)
	processedArgs := make([]any, 0, len(args))
	for _, arg := range args {
		// Check if arg is a func(Query)Query callback
		if fn, ok := arg.(func(contractsorm.Query) contractsorm.Query); ok {
			// Invoke the callback with a clone of the current query
			subQuery := fn(q.Clone())
			// Build the subquery SQL
			builder := NewBuilder(subQuery.(*Query))
			subSQL, subArgs := builder.BuildSelect()
			// Inline the subquery SQL into the query string, replacing the first ?
			queryStr = strings.Replace(queryStr, "?", fmt.Sprintf("(%s)", subSQL), 1)
			// Append subquery bound args
			processedArgs = append(processedArgs, subArgs...)
		} else {
			processedArgs = append(processedArgs, arg)
		}
	}
	q.havings = append(q.havings, havingClause{query: queryStr, args: processedArgs})
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
	subQ := q.Clone().(*Query)
	subQ = callback(subQ).(*Query)
	subSQL, subArgs := NewBuilder(subQ).BuildSelect()
	q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("EXISTS (%s)", subSQL), args: subArgs})
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
	// (col1 op ? OR col2 op ? OR ...)
	var parts []string
	var args []any
	for _, col := range columns {
		parts = append(parts, fmt.Sprintf("%s %s ?", col, operator))
		args = append(args, value)
	}
	q.wheres = append(q.wheres, whereClause{_type: "and", query: "(" + strings.Join(parts, " OR ") + ")", args: args})
	return q
}
func (q *Query) WhereAll(columns []string, operator string, value any) contractsorm.Query {
	// (col1 op ? AND col2 op ? AND ...)
	var parts []string
	var args []any
	for _, col := range columns {
		parts = append(parts, fmt.Sprintf("%s %s ?", col, operator))
		args = append(args, value)
	}
	q.wheres = append(q.wheres, whereClause{_type: "and", query: "(" + strings.Join(parts, " AND ") + ")", args: args})
	return q
}
func (q *Query) WhereNone(columns []string, operator string, value any) contractsorm.Query {
	// NOT (col1 op ? OR col2 op ? OR ...)
	var parts []string
	var args []any
	for _, col := range columns {
		parts = append(parts, fmt.Sprintf("%s %s ?", col, operator))
		args = append(args, value)
	}
	q.wheres = append(q.wheres, whereClause{_type: "and", query: "NOT (" + strings.Join(parts, " OR ") + ")", args: args})
	return q
}

func (q *Query) WhereJsonContains(column string, value any) contractsorm.Query {
	// SQLite uses different JSON functions than MySQL/Postgres
	if q.driver != nil && q.driver.Dialect() == "sqlite" {
		// SQLite: json_extract(json, path) = value
		// Convert -> to JSON path format
		path := strings.ReplaceAll(column, "->", ".")
		q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("json_extract(%s, '$%s') = ?", column, path), args: []any{value}})
	} else {
		q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("JSON_CONTAINS(%s, ?)", column), args: []any{value}})
	}
	return q
}
func (q *Query) OrWhereJsonContains(column string, value any) contractsorm.Query {
	if q.driver != nil && q.driver.Dialect() == "sqlite" {
		path := strings.ReplaceAll(column, "->", ".")
		q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("json_extract(%s, '$%s') = ?", column, path), args: []any{value}})
	} else {
		q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("JSON_CONTAINS(%s, ?)", column), args: []any{value}})
	}
	return q
}
func (q *Query) WhereJsonDoesntContain(column string, value any) contractsorm.Query {
	if q.driver != nil && q.driver.Dialect() == "sqlite" {
		path := strings.ReplaceAll(column, "->", ".")
		q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("json_extract(%s, '$%s') != ?", column, path), args: []any{value}})
	} else {
		q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("NOT JSON_CONTAINS(%s, ?)", column), args: []any{value}})
	}
	return q
}
func (q *Query) OrWhereJsonDoesntContain(column string, value any) contractsorm.Query {
	if q.driver != nil && q.driver.Dialect() == "sqlite" {
		path := strings.ReplaceAll(column, "->", ".")
		q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("json_extract(%s, '$%s') != ?", column, path), args: []any{value}})
	} else {
		q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("NOT JSON_CONTAINS(%s, ?)", column), args: []any{value}})
	}
	return q
}
func (q *Query) WhereJsonContainsKey(column string) contractsorm.Query {
	if q.driver != nil && q.driver.Dialect() == "sqlite" {
		// SQLite: json_type(json, path) is not null
		path := strings.ReplaceAll(column, "->", ".")
		q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("json_type(%s, '$%s') IS NOT NULL", column, path), args: nil})
	} else {
		q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("JSON_CONTAINS_PATH(%s, '$.%s')", column, column), args: nil})
	}
	return q
}
func (q *Query) OrWhereJsonContainsKey(column string) contractsorm.Query {
	if q.driver != nil && q.driver.Dialect() == "sqlite" {
		path := strings.ReplaceAll(column, "->", ".")
		q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("json_type(%s, '$%s') IS NOT NULL", column, path), args: nil})
	} else {
		q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("JSON_CONTAINS_PATH(%s, '$.%s')", column, column), args: nil})
	}
	return q
}
func (q *Query) WhereJsonDoesntContainKey(column string) contractsorm.Query {
	if q.driver != nil && q.driver.Dialect() == "sqlite" {
		path := strings.ReplaceAll(column, "->", ".")
		q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("json_type(%s, '$%s') IS NULL", column, path), args: nil})
	} else {
		q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("NOT JSON_CONTAINS_PATH(%s, '$.%s')", column, column), args: nil})
	}
	return q
}
func (q *Query) OrWhereJsonDoesntContainKey(column string) contractsorm.Query {
	if q.driver != nil && q.driver.Dialect() == "sqlite" {
		path := strings.ReplaceAll(column, "->", ".")
		q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("json_type(%s, '$%s') IS NULL", column, path), args: nil})
	} else {
		q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("NOT JSON_CONTAINS_PATH(%s, '$.%s')", column, column), args: nil})
	}
	return q
}
func (q *Query) WhereJsonLength(column string, operator string, value any) contractsorm.Query {
	if q.driver != nil && q.driver.Dialect() == "sqlite" {
		// SQLite: json_array_length(json, path)
		path := strings.ReplaceAll(column, "->", ".")
		q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("json_array_length(%s, '$%s') %s ?", column, path, operator), args: []any{value}})
	} else {
		q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("JSON_LENGTH(%s) %s ?", column, operator), args: []any{value}})
	}
	return q
}

func (q *Query) Count(count *int64) error {
	// Use a clone to avoid mutating the query state
	clone := q.Clone().(*Query)
	clone.aggregate = "COUNT"
	clone.aggregateCol = "*"
	clone.distinct = q.distinct
	clone.distinctCols = q.distinctCols

	// Build SELECT query
	builder := NewBuilder(clone)
	sql, args := builder.BuildSelect()

	// Execute query
	start := time.Now()
	var err error
	if q.tx != nil {
		err = q.tx.QueryRowContext(q.ctx, sql, args...).Scan(count)
	} else {
		databaseConn, err := q.ReadDB()
		if err != nil {
			return err
		}
		err = databaseConn.QueryRowContext(q.ctx, sql, args...).Scan(count)
	}

	if err != nil {
		return fmt.Errorf("failed to execute COUNT query: %w", err)
	}

	// Log query if enabled
	q.logQuery(sql, args, start)

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
	start := time.Now()
	var err error
	if q.tx != nil {
		err = q.tx.QueryRowContext(q.ctx, sql, args...).Scan(dest)
	} else {
		databaseConn, err := q.ReadDB()
		if err != nil {
			return err
		}
		err = databaseConn.QueryRowContext(q.ctx, sql, args...).Scan(dest)
	}

	if err != nil {
		return fmt.Errorf("failed to execute SUM query: %w", err)
	}

	// Log query if enabled
	q.logQuery(sql, args, start)

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
	start := time.Now()
	var err error
	if q.tx != nil {
		err = q.tx.QueryRowContext(q.ctx, sql, args...).Scan(dest)
	} else {
		databaseConn, err := q.ReadDB()
		if err != nil {
			return err
		}
		err = databaseConn.QueryRowContext(q.ctx, sql, args...).Scan(dest)
	}

	if err != nil {
		return fmt.Errorf("failed to execute AVG query: %w", err)
	}

	// Log query if enabled
	q.logQuery(sql, args, start)

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
	start := time.Now()
	var err error
	if q.tx != nil {
		err = q.tx.QueryRowContext(q.ctx, sql, args...).Scan(dest)
	} else {
		databaseConn, err := q.ReadDB()
		if err != nil {
			return err
		}
		err = databaseConn.QueryRowContext(q.ctx, sql, args...).Scan(dest)
	}

	if err != nil {
		return fmt.Errorf("failed to execute MIN query: %w", err)
	}

	// Log query if enabled
	q.logQuery(sql, args, start)

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
	start := time.Now()
	var err error
	if q.tx != nil {
		err = q.tx.QueryRowContext(q.ctx, sql, args...).Scan(dest)
	} else {
		databaseConn, err := q.ReadDB()
		if err != nil {
			return err
		}
		err = databaseConn.QueryRowContext(q.ctx, sql, args...).Scan(dest)
	}

	if err != nil {
		return fmt.Errorf("failed to execute MAX query: %w", err)
	}

	// Log query if enabled
	q.logQuery(sql, args, start)

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
	start := time.Now()
	var count int64
	var err error
	if q.tx != nil {
		err = q.tx.QueryRowContext(q.ctx, sql, args...).Scan(&count)
	} else {
		databaseConn, err := q.ReadDB()
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
	q.logQuery(sql, args, start)

	return nil
}

func (q *Query) Pluck(column string, dest any) error {
	// Set select to only the specified column
	q.selects = []selectClause{{expr: column}}

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

	databaseConn, err := q.ReadDB()
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
	q.selects = []selectClause{{expr: column}}
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
		databaseConn, err := q.ReadDB()
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

	databaseConn, err := q.ReadDB()
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
	// Build SELECT query without limit (we chunk in memory)
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

	databaseConn, err := q.ReadDB()
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

	databaseConn, err := q.ReadDB()
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
	// Build WHERE conditions from attributes
	clone := q.Clone().(*Query)

	// Handle map[string]any for attributes
	if attrs, ok := attributes.(map[string]any); ok {
		for col, val := range attrs {
			clone.Where(col, val)
		}
	} else {
		// For structs, extract fields and build WHERE conditions
		cols, vals, err := NewBuilder(q).extractSingleColumnsAndValues(attributes)
		if err == nil {
			for i, col := range cols {
				clone.Where(col, vals[i])
			}
		}
	}

	// Try to find the record first
	count := int64(0)
	if err := clone.Count(&count); err != nil {
		return err
	}

	if count > 0 {
		// Record exists, update it
		// Use the original query with WHERE conditions from attributes
		updateQ := q.Clone().(*Query)
		if attrs, ok := attributes.(map[string]any); ok {
			for col, val := range attrs {
				updateQ.Where(col, val)
			}
		} else {
			cols, vals, err := NewBuilder(q).extractSingleColumnsAndValues(attributes)
			if err == nil {
				for i, col := range cols {
					updateQ.Where(col, vals[i])
				}
			}
		}
		_, err := updateQ.Update(values)
		return err
	}

	// Record doesn't exist, create it
	// Merge attributes and values for the insert
	if attrsMap, ok := attributes.(map[string]any); ok {
		if valsMap, ok := values.(map[string]any); ok {
			// Merge both maps
			merged := make(map[string]any)
			for k, v := range attrsMap {
				merged[k] = v
			}
			for k, v := range valsMap {
				merged[k] = v
			}
			return q.Create(merged)
		}
	}

	// For struct values or mixed types, just create with values
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
	var order string
	if q.driver != nil && q.driver.Dialect() == "mysql" {
		order = "RAND()"
	} else {
		order = "RANDOM()"
	}
	newQ := *q
	newQ.orders = append(newQ.orders, orderClause{column: order, direction: ""})
	return &newQ
}
func (q *Query) LockForUpdate() contractsorm.Query {
	newQ := *q
	newQ.lockForUpdate = true
	return &newQ
}
func (q *Query) SharedLock() contractsorm.Query {
	newQ := *q
	newQ.sharedLock = true
	return &newQ
}
func (q *Query) Raw(sql string, values ...any) contractsorm.Query {
	// Store raw SQL for later use
	newQ := *q
	newQ.rawSQL = sql
	newQ.rawArgs = values
	return &newQ
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

func (q *Query) Omit(columns ...string) contractsorm.Query {
	newQ := *q
	newQ.omitColumns = append(newQ.omitColumns, columns...)
	return &newQ
}

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
	start := time.Now()
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
	q.logQuery(sql, args, start)

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
	start := time.Now()
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
	q.logQuery(sql, args, start)

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
	if err := q.doCommit(); err != nil {
		return err
	}
	q.inTransaction = false
	q.tx = nil
	return nil
}
func (q *Query) Rollback() error {
	if !q.inTransaction || q.tx == nil {
		return fmt.Errorf("not in a transaction")
	}
	err := q.doRollback()
	q.inTransaction = false
	q.tx = nil
	return err
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
func (q *Query) Scopes(funcs ...func(contractsorm.Query) contractsorm.Query) contractsorm.Query {
	newQ := *q
	newQ.scopes = append(newQ.scopes, funcs...)
	return &newQ
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
		databaseConn, err := q.ReadDB()
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

// BeforeCommit registers a callback to run before the transaction is committed.
func (q *Query) BeforeCommit(callback func() error) {
	q.beforeCommit = append(q.beforeCommit, callback)
}

// AfterCommit registers a callback to run after the transaction is committed.
func (q *Query) AfterCommit(callback func() error) {
	q.afterCommit = append(q.afterCommit, callback)
}

// BeforeRollback registers a callback to run before the transaction is rolled back.
func (q *Query) BeforeRollback(callback func() error) {
	q.beforeRollback = append(q.beforeRollback, callback)
}

// AfterRollback registers a callback to run after the transaction is rolled back.
func (q *Query) AfterRollback(callback func() error) {
	q.afterRollback = append(q.afterRollback, callback)
}

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

		columns, err := rows.Columns()
		if err != nil {
			return fmt.Errorf("failed to get columns: %w", err)
		}

		for rows.Next() {
			// Create new element
			elemPtr := reflect.New(elemType)
			elem := elemPtr.Elem()

			// Scan into element
			values := make([]any, len(columns))

			if elem.Kind() == reflect.Map {
				// Scan into a temporary []any then build the map
				ptrs := make([]any, len(columns))
				for i := range values {
					ptrs[i] = &values[i]
				}
				if err := rows.Scan(ptrs...); err != nil {
					return fmt.Errorf("failed to scan row: %w", err)
				}
				m := reflect.MakeMap(elemType)
				keyType := elemType.Key()
				for i, col := range columns {
					m.SetMapIndex(reflect.ValueOf(col).Convert(keyType), reflect.ValueOf(values[i]))
				}
				elem.Set(m)
			} else if elem.Kind() == reflect.Struct {
				values = structScanDests(elem, columns)
				if err := rows.Scan(values...); err != nil {
					return fmt.Errorf("failed to scan row: %w", err)
				}
				copyScanResults(elem, columns, values)
			} else {
				// Scalar slice element (e.g. []string, []int)
				if err := rows.Scan(elem.Addr().Interface()); err != nil {
					return fmt.Errorf("failed to scan row: %w", err)
				}
			}

			// Initialize relations if requested
			if len(q.withRelations) > 0 {
				q.initializeRelations(elem)
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

		values := structScanDests(destValue, columns)
		if err := rows.Scan(values...); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}
		copyScanResults(destValue, columns, values)

		// Initialize relations if requested
		if len(q.withRelations) > 0 {
			q.initializeRelations(destValue)
		}

		return rows.Err()
	}

	// Handle single map destination (*map[string]any)
	if destValue.Kind() == reflect.Map {
		if !rows.Next() {
			return nil
		}

		columns, err := rows.Columns()
		if err != nil {
			return fmt.Errorf("failed to get columns: %w", err)
		}

		values := make([]any, len(columns))
		ptrs := make([]any, len(columns))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		m := make(map[string]any, len(columns))
		for i, col := range columns {
			m[col] = values[i]
		}
		destValue.Set(reflect.ValueOf(m))

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

	// Get callback parameter type
	callbackType := callbackValue.Type()
	if callbackType.NumIn() != 1 {
		return fmt.Errorf("callback must accept exactly one parameter")
	}

	paramType := callbackType.In(0)

	// Determine if we need to convert to a specific type
	var isTypedSlice bool
	var elemType reflect.Type

	if paramType.Kind() == reflect.Slice {
		isTypedSlice = true
		elemType = paramType.Elem()
	}

	// Process rows in chunks
	chunk := make([]map[string]any, 0, size)

	for rows.Next() {
		// Scan into a map
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

		row := make(map[string]any)
		for i, col := range columns {
			row[col] = values[i]
		}

		chunk = append(chunk, row)

		// Call callback when chunk is full
		if len(chunk) >= size {
			var chunkValue reflect.Value
			if isTypedSlice {
				// Convert maps to typed slice
				chunkValue = reflect.MakeSlice(paramType, 0, len(chunk))
				for _, rowMap := range chunk {
					elem := reflect.New(elemType).Elem()
					// Map row data to struct fields
					for key, val := range rowMap {
						fieldName := toCamelCase(key)
						field := elem.FieldByName(fieldName)
						if field.IsValid() && field.CanSet() {
							if val != nil {
								// Convert value to field type
								valReflect := reflect.ValueOf(val)
								if valReflect.Type().ConvertibleTo(field.Type()) {
									field.Set(valReflect.Convert(field.Type()))
								}
							}
						}
					}
					chunkValue = reflect.Append(chunkValue, elem)
				}
			} else {
				chunkValue = reflect.ValueOf(chunk)
			}

			results := callbackValue.Call([]reflect.Value{chunkValue})
			if len(results) > 0 {
				if err, ok := results[0].Interface().(error); ok && err != nil {
					return err
				}
			}
			chunk = make([]map[string]any, 0, size)
		}
	}

	// Process remaining rows in the last chunk
	if len(chunk) > 0 {
		var chunkValue reflect.Value
		if isTypedSlice {
			// Convert maps to typed slice
			chunkValue = reflect.MakeSlice(paramType, 0, len(chunk))
			for _, rowMap := range chunk {
				elem := reflect.New(elemType).Elem()
				// Map row data to struct fields
				for key, val := range rowMap {
					fieldName := toCamelCase(key)
					field := elem.FieldByName(fieldName)
					if field.IsValid() && field.CanSet() {
						if val != nil {
							// Convert value to field type
							valReflect := reflect.ValueOf(val)
							if valReflect.Type().ConvertibleTo(field.Type()) {
								field.Set(valReflect.Convert(field.Type()))
							}
						}
					}
				}
				chunkValue = reflect.Append(chunkValue, elem)
			}
		} else {
			chunkValue = reflect.ValueOf(chunk)
		}

		results := callbackValue.Call([]reflect.Value{chunkValue})
		if len(results) > 0 {
			if err, ok := results[0].Interface().(error); ok && err != nil {
				return err
			}
		}
	}

	return rows.Err()
}

// toCamelCase converts snake_case to CamelCase
func toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	for i := range parts {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + strings.ToLower(parts[i][1:])
		}
	}
	return strings.Join(parts, "")
}

// getPrimaryKeyValue returns the primary key value (ID/Id) of a struct as int64, 0 if absent or zero.
func getPrimaryKeyValue(value any) int64 {
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return 0
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return 0
	}
	for _, name := range []string{"ID", "Id"} {
		field := v.FieldByName(name)
		if !field.IsValid() {
			continue
		}
		switch field.Kind() {
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return int64(field.Uint())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return field.Int()
		}
	}
	return 0
}

// structFieldColumnName returns the column name for a struct field by checking
// db, neat, gorm tags (in that order), then falling back to a snake_case of the field name.
func structFieldColumnName(f reflect.StructField) string {
	for _, tag := range []string{"db", "neat", "gorm"} {
		if v := f.Tag.Get(tag); v != "" && v != "-" {
			// take the first semicolon-delimited part; for gorm it may be "column:name"
			part := strings.SplitN(v, ";", 2)[0]
			if strings.HasPrefix(part, "column:") {
				return strings.TrimPrefix(part, "column:")
			}
			// db and neat tags use the value directly as the column name
			if tag == "db" || tag == "neat" {
				return part
			}
		}
	}
	// snake_case the Go field name
	return camelToSnake(f.Name)
}

// camelToSnake converts CamelCase to snake_case.
func camelToSnake(s string) string {
	var out []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			out = append(out, '_')
		}
		out = append(out, []rune(strings.ToLower(string(r)))...)
	}
	return string(out)
}

// structScanDests builds a scan-destination slice aligned to columns.
// Each element is either a pointer to the matching struct field or a *any placeholder.
// Non-addressable fields get a *T temporary that copyScanResults will copy back.
func structScanDests(v reflect.Value, columns []string) []any {
	// Build column → field index map
	t := v.Type()
	colToField := make(map[string]int, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		col := structFieldColumnName(t.Field(i))
		colToField[strings.ToLower(col)] = i
	}

	dests := make([]any, len(columns))
	for i, col := range columns {
		key := strings.ToLower(col)
		if fi, ok := colToField[key]; ok {
			field := v.Field(fi)
			if field.CanAddr() {
				dests[i] = field.Addr().Interface()
			} else {
				// allocate a temporary pointer of the field's type
				ptr := reflect.New(field.Type())
				dests[i] = ptr.Interface()
			}
		} else {
			var placeholder any
			dests[i] = &placeholder
		}
	}
	return dests
}

// copyScanResults copies values from non-addressable temporaries back into struct fields.
// For addressable fields the scan wrote directly into them; this is a no-op for those.
func copyScanResults(v reflect.Value, columns []string, dests []any) {
	t := v.Type()
	colToField := make(map[string]int, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		col := structFieldColumnName(t.Field(i))
		colToField[strings.ToLower(col)] = i
	}
	for i, col := range columns {
		key := strings.ToLower(col)
		fi, ok := colToField[key]
		if !ok {
			continue
		}
		field := v.Field(fi)
		if field.CanAddr() {
			continue // already written by Scan
		}
		// copy from the temporary pointer
		ptrVal := reflect.ValueOf(dests[i])
		if ptrVal.Kind() == reflect.Ptr && !ptrVal.IsNil() {
			val := ptrVal.Elem()
			if val.Type().AssignableTo(field.Type()) && field.CanSet() {
				field.Set(val)
			}
		}
	}
}

// setModelPrimaryKey sets the primary key field (ID or Id) on a struct model to the given value.
func setModelPrimaryKey(value any, id int64) {
	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return
	}
	for _, name := range []string{"ID", "Id"} {
		field := v.FieldByName(name)
		if !field.IsValid() || !field.CanSet() {
			continue
		}
		switch field.Kind() {
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			field.SetUint(uint64(id))
			return
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			field.SetInt(id)
			return
		}
	}
}
