package query

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/contracts/log"
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
	queryLog   *[]contractsorm.QueryLog
	enableLog  bool

	// Query state
	table        string
	tableArgs    []any
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
	inTransaction  bool
	tx             *sql.Tx
	savepointLevel int
	savepointName  string

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
	withRelations       []string
	relationConstraints map[string]func(contractsorm.Query) contractsorm.Query

	// Count and Exists subqueries
	withCountQueries  []countQuery
	withExistsQueries []existsQuery

	// Validation error
	buildError error
}

// countQuery represents a count subquery for eager loading.
type countQuery struct {
	relation   string
	column     string
	constraint func(contractsorm.Query) contractsorm.Query
}

// existsQuery represents an exists subquery for eager loading.
type existsQuery struct {
	relation   string
	constraint func(contractsorm.Query) contractsorm.Query
}

// whereClause represents a WHERE clause in a query.
type whereClause struct {
	_type string // "and", "or"
	query string
	args  []any
}

// joinClause represents a JOIN clause in a query.
type joinClause struct {
	_type string // "join", "left join", "right join", "cross join"
	query string
	args  []any
}

// havingClause represents a HAVING clause in a query.
type havingClause struct {
	query string
	args  []any
}

// selectClause represents a SELECT clause in a query.
type selectClause struct {
	expr string
	args []any
}

// RawExpression represents a raw SQL expression that can be used as a value in maps
type RawExpression struct {
	SQL  string
	Args []any
}

// RawExpr creates a new raw SQL expression for use in Create/Update values.
// WARNING: This function injects SQL directly without parameterization. NEVER pass user input to this function.
// The SQL will be injected directly into the query, creating a SQL injection vulnerability if used with untrusted data.
// Example: db.Table("users").Create(map[string]any{"created_at": RawExpr("NOW()")})
// Safe usage: Only use with hardcoded SQL expressions like "NOW()", "CURRENT_TIMESTAMP", etc.
// Dangerous usage: RawExpr(userInput) - DO NOT DO THIS
func RawExpr(sql string, args ...any) RawExpression {
	// Filter out nil arguments to prevent issues
	filteredArgs := make([]any, 0, len(args))
	for _, arg := range args {
		if arg != nil {
			filteredArgs = append(filteredArgs, arg)
		}
	}
	return RawExpression{SQL: sql, Args: filteredArgs}
}

// orderClause represents an ORDER BY clause in a query.
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
		queryLog:        &[]contractsorm.QueryLog{},
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

// isPostgres returns true if the driver dialect is PostgreSQL.
func (q *Query) isPostgres() bool {
	return q.driver != nil && q.driver.Dialect() == "postgres"
}

// isSQLServer returns true if the driver dialect is SQL Server.
func (q *Query) isSQLServer() bool {
	return q.driver != nil && q.driver.Dialect() == "sqlserver"
}

// isMySQL returns true if the driver dialect is MySQL.
func (q *Query) isMySQL() bool {
	return q.driver != nil && q.driver.Dialect() == "mysql"
}

// isSQLite returns true if the driver dialect is SQLite.
func (q *Query) isSQLite() bool {
	return q.driver != nil && q.driver.Dialect() == "sqlite"
}

// isOracle returns true if the driver dialect is Oracle.
func (q *Query) isOracle() bool {
	return q.driver != nil && q.driver.Dialect() == "oracle"
}

// Clone returns a new Query with shared connection state but empty query-builder state.
func (q *Query) Clone() contractsorm.Query {
	clone := q.newQuery()
	clone.table = q.table
	clone.tableArgs = append([]any{}, q.tableArgs...)
	clone.model = q.model
	clone.omitColumns = append([]string{}, q.omitColumns...)
	clone.distinct = q.distinct
	clone.distinctCols = append([]string{}, q.distinctCols...)

	clone.wheres = make([]whereClause, len(q.wheres))
	for i, w := range q.wheres {
		clone.wheres[i] = w
		clone.wheres[i].args = append([]any{}, w.args...)
	}

	clone.selects = make([]selectClause, len(q.selects))
	for i, s := range q.selects {
		clone.selects[i] = s
		clone.selects[i].args = append([]any{}, s.args...)
	}

	clone.joins = make([]joinClause, len(q.joins))
	for i, j := range q.joins {
		clone.joins[i] = j
		clone.joins[i].args = append([]any{}, j.args...)
	}

	clone.havings = make([]havingClause, len(q.havings))
	for i, h := range q.havings {
		clone.havings[i] = h
		clone.havings[i].args = append([]any{}, h.args...)
	}

	clone.groups = append([]string{}, q.groups...)
	clone.orders = append([]orderClause{}, q.orders...)

	if q.limit != nil {
		limit := *q.limit
		clone.limit = &limit
	}
	if q.offset != nil {
		offset := *q.offset
		clone.offset = &offset
	}

	clone.withTrashed = q.withTrashed
	clone.onlyTrashed = q.onlyTrashed
	clone.withoutTrashed = q.withoutTrashed
	clone.queryLog = q.queryLog
	clone.rawSQL = q.rawSQL
	clone.rawArgs = append([]any{}, q.rawArgs...)
	clone.lockForUpdate = q.lockForUpdate
	clone.sharedLock = q.sharedLock
	clone.scopes = append([]func(contractsorm.Query) contractsorm.Query{}, q.scopes...)
	clone.withRelations = append([]string{}, q.withRelations...)
	if q.relationConstraints != nil {
		clone.relationConstraints = make(map[string]func(contractsorm.Query) contractsorm.Query)
		for k, v := range q.relationConstraints {
			clone.relationConstraints[k] = v
		}
	}
	clone.withCountQueries = append([]countQuery{}, q.withCountQueries...)
	clone.withExistsQueries = append([]existsQuery{}, q.withExistsQueries...)
	clone.buildError = q.buildError

	// Transaction state
	clone.inTransaction = q.inTransaction
	clone.tx = q.tx
	clone.savepointLevel = q.savepointLevel
	clone.savepointName = q.savepointName

	// Observer state
	clone.modelToObserver = append([]contractsorm.ModelToObserver{}, q.modelToObserver...)
	clone.withoutEvents = q.withoutEvents
	clone.dispatcher = q.dispatcher

	// Transaction lifecycle hooks
	clone.beforeCommit = append([]func() error{}, q.beforeCommit...)
	clone.afterCommit = append([]func() error{}, q.afterCommit...)
	clone.beforeRollback = append([]func() error{}, q.beforeRollback...)
	clone.afterRollback = append([]func() error{}, q.afterRollback...)

	return clone
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
	case "oracle":
		return driver.NewOracle()
	default:
		return driver.NewSQLite()
	}
}

// Model sets the model for the query.
func (q *Query) Model(value any) contractsorm.Query {
	q.model = value
	q.table = q.resolveTableName(value)
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
	// q.withTrashed = false
	// q.onlyTrashed = false
	// q.withoutTrashed = false
	q.withRelations = nil
	q.relationConstraints = nil
	q.withCountQueries = nil
	q.withExistsQueries = nil
	return q
}

// initializeRelations initializes association fields for relations requested via With.
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

// Select adds columns to the select clause.
func (q *Query) newQuery() *Query {
	newQ := NewQuery(q.ctx, q.db, q.driver, q.connection, q.dbConfig, q.log)
	newQ.readDB = q.readDB
	newQ.writeDB = q.writeDB
	newQ.enableLog = q.enableLog
	newQ.queryLog = q.queryLog
	newQ.withRelations = nil
	newQ.relationConstraints = nil
	// Note: buildError is intentionally NOT copied to newQuery()
	// newQuery() creates a fresh query without inheriting build errors
	// Use Clone() to preserve buildError across query copies
	return newQ
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
	// SQL Server: uses OUTPUT clause (already added in BuildInsert)
	// Oracle: use RETURNING id for identity columns
	if q.isPostgres() || q.isOracle() {
		insertSQL = insertSQL + " RETURNING id"
	}

	start := time.Now()
	var lastID int64

	if q.isPostgres() || q.isSQLServer() || q.isOracle() {
		// For PostgreSQL with RETURNING or SQL Server with OUTPUT, use QueryRow instead of Exec
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
		if q.isOracle() {
			// Oracle may return different types, try scanning into interface{} first
			var idInterface any
			if err := row.Scan(&idInterface); err != nil {
				return 0, fmt.Errorf("failed to get inserted ID: %w", err)
			}
			// Convert to int64
			switch v := idInterface.(type) {
			case int64:
				lastID = v
			case int:
				lastID = int64(v)
			case float64:
				lastID = int64(v)
			case []byte:
				// Oracle may return as bytes
				var num int64
				if _, err := fmt.Sscanf(string(v), "%d", &num); err == nil {
					lastID = num
				} else {
					return 0, fmt.Errorf("failed to convert Oracle ID bytes '%s' to int64: %w", string(v), err)
				}
			default:
				return 0, fmt.Errorf("unsupported Oracle ID type: %T", idInterface)
			}
		} else {
			if err := row.Scan(&lastID); err != nil {
				return 0, fmt.Errorf("failed to get inserted ID: %w", err)
			}
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
	if q.enableLog && q.queryLog != nil {
		*q.queryLog = append(*q.queryLog, contractsorm.QueryLog{
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

func (q *Query) validateAggregate(column string, dest any) error {
	if dest == nil {
		return fmt.Errorf("destination cannot be nil")
	}

	// Validate column name: alphanumeric, underscores, dots or *
	for _, r := range column {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '.' || r == '*') {
			return fmt.Errorf("invalid column name: %s", column)
		}
	}

	return nil
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
		} else {
			// Attributes is map, values is struct - merge them
			merged := make(map[string]any)
			for k, v := range attrsMap {
				merged[k] = v
			}
			// Extract struct fields and add to merged map
			cols, vals, err := NewBuilder(q).extractSingleColumnsAndValues(values)
			if err == nil {
				for i, col := range cols {
					merged[col] = vals[i]
				}
			}
			return q.Create(merged)
		}
	}

	// For struct values or mixed types, just create with values
	return q.Create(values)
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

			// Note: Relations are loaded after rows are closed to avoid SQLite deadlock
			// This is handled by the calling methods (First, Get, etc.)

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

		// Note: Relations are loaded after rows are closed to avoid SQLite deadlock
		// This is handled by the calling methods (First, Get, etc.)

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
	if paramType.Kind() != reflect.Slice {
		return fmt.Errorf("callback parameter must be a slice")
	}

	elemType := paramType.Elem()
	realElemType := elemType
	isPtr := elemType.Kind() == reflect.Ptr
	if isPtr {
		realElemType = elemType.Elem()
	}

	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to get columns: %w", err)
	}

	// Process rows in chunks
	chunk := reflect.MakeSlice(paramType, 0, size)

	for rows.Next() {
		// Create a new element
		var elem reflect.Value
		if isPtr {
			elem = reflect.New(realElemType)
		} else {
			elem = reflect.New(elemType).Elem()
		}

		// Scan into element
		if realElemType.Kind() == reflect.Struct {
			var scanElem reflect.Value
			if isPtr {
				scanElem = elem.Elem()
			} else {
				scanElem = elem
			}
			values := structScanDests(scanElem, columns)
			if err := rows.Scan(values...); err != nil {
				return fmt.Errorf("failed to scan row into struct: %w", err)
			}
			copyScanResults(scanElem, columns, values)
		} else if realElemType.Kind() == reflect.Map {
			// Scan into a map
			values := make([]any, len(columns))
			valuePtrs := make([]any, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}
			if err := rows.Scan(valuePtrs...); err != nil {
				return fmt.Errorf("failed to scan row into map: %w", err)
			}
			m := reflect.MakeMap(realElemType)
			keyType := realElemType.Key()
			for i, col := range columns {
				m.SetMapIndex(reflect.ValueOf(col).Convert(keyType), reflect.ValueOf(values[i]))
			}
			if isPtr {
				elem.Elem().Set(m)
			} else {
				elem.Set(m)
			}
		} else {
			// Scalar element
			var scanDest reflect.Value
			if isPtr {
				scanDest = elem
			} else {
				scanDest = elem.Addr()
			}
			if err := rows.Scan(scanDest.Interface()); err != nil {
				return fmt.Errorf("failed to scan row into scalar: %w", err)
			}
		}

		chunk = reflect.Append(chunk, elem)

		// Call callback when chunk is full
		if chunk.Len() >= size {
			results := callbackValue.Call([]reflect.Value{chunk})
			if len(results) > 0 {
				if err, ok := results[0].Interface().(error); ok && err != nil {
					return err
				}
			}
			chunk = reflect.MakeSlice(paramType, 0, size)
		}
	}

	// Process remaining rows in the last chunk
	if chunk.Len() > 0 {
		results := callbackValue.Call([]reflect.Value{chunk})
		if len(results) > 0 {
			if err, ok := results[0].Interface().(error); ok && err != nil {
				return err
			}
		}
	}

	return rows.Err()
}
