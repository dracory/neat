package orm

import (
	"context"
	"database/sql"

	"github.com/dracory/neat/contracts/database"
)

type QueryLog struct {
	Query    string
	Bindings []any
	Time     float64 // duration in milliseconds
}

type Orm interface {
	// Connection gets an Orm instance from the connection pool.
	Connection(name string) Orm
	// DB gets the underlying database connection.
	DB() (*sql.DB, error)
	// DisableQueryLog disables the capturing of executed queries.
	DisableQueryLog()
	// EnableQueryLog enables the capturing of executed queries.
	EnableQueryLog()
	// FlushQueryLog clears the captured queries from the log.
	FlushQueryLog()
	// GetQueryLog retrieves the captured queries from the log.
	GetQueryLog() []QueryLog
	// Factory gets a new factory instance for the given model name.
	Factory() Factory
	// DatabaseName gets the current database name.
	DatabaseName() string
	// Name gets the current connection name.
	Name() string
	// Observe registers an observer with the Orm.
	Observe(model any, observer Observer)
	// Query gets a new query builder instance.
	Query() Query
	// Refresh resets the Orm instance.
	Refresh()
	// SetQuery sets the query builder instance.
	SetQuery(query Query)
	// Transaction runs a callback wrapped in a database transaction.
	Transaction(txFunc func(tx Query) error, opts ...*sql.TxOptions) error
	// WithContext sets the context to be used by the Orm.
	WithContext(ctx context.Context) Orm
}

type Query interface {
	// Connection sets the connection name for the query.
	Connection(name string) Query
	// Association gets an association instance by name.
	Association(association string) Association
	// AfterCommit registers a callback to be executed after the transaction is committed.
	AfterCommit(callback func() error)
	// AfterRollback registers a callback to be executed after the transaction is rolled back.
	AfterRollback(callback func() error)
	// Begin begins a new transaction
	Begin(opts ...*sql.TxOptions) (Query, error)
	// BeforeCommit registers a callback to be executed before the transaction is committed.
	BeforeCommit(callback func() error)
	// BeforeRollback registers a callback to be executed before the transaction is rolled back.
	BeforeRollback(callback func() error)
	// Commit commits the changes in a transaction.
	Commit() error
	// Count retrieve the "count" result of the query.
	Count(count *int64) error
	// Chunk the results of the query.
	Chunk(size int, callback any) error
	// Create inserts new record into the database.
	Create(value any) error
	// InsertGetId inserts a new record and returns the last inserted ID.
	InsertGetId(values any) (uint, error)
	// Cursor returns a cursor, use scan to iterate over the returned rows.
	Cursor() (chan Cursor, error)
	// DB gets the underlying database connection.
	DB() (*sql.DB, error)
	// Delete deletes records matching given conditions, if the conditions are empty will delete all records.
	Delete(value ...any) (*Result, error)
	// Distinct specifies distinct fields to query.
	Distinct(args ...any) Query
	// Driver gets the driver for the query.
	Driver() database.Driver
	// Exec executes raw sql
	Exec(sql string, values ...any) (*Result, error)
	// Exists returns true if matching records exist; otherwise, it returns false.
	Exists(exists *bool) error
	// Find finds records that match given conditions.
	Find(dest any, conds ...any) error
	// FindOrFail finds records that match given conditions or throws an error.
	FindOrFail(dest any, conds ...any) error
	// First finds record that match given conditions.
	First(dest any) error
	// FirstOrCreate finds the first record that matches the given attributes
	// or create a new one with those attributes if none was found.
	FirstOrCreate(dest any, conds ...any) error
	// FirstOr finds the first record that matches the given conditions or
	// execute the callback and return its result if no record is found.
	FirstOr(dest any, callback func() error) error
	// FirstOrFail finds the first record that matches the given conditions or throws an error.
	FirstOrFail(dest any) error
	// FirstOrNew finds the first record that matches the given conditions or
	// return a new instance of the model initialized with those attributes.
	FirstOrNew(dest any, attributes any, values ...any) error
	// ForceDelete forces delete records matching given conditions.
	ForceDelete(value ...any) (*Result, error)
	// Get retrieves all rows from the database.
	Get(dest any) error
	// Group specifies the group method on the query.
	Group(name string) Query
	// Having specifying HAVING conditions for the query.
	Having(query any, args ...any) Query
	// InRandomOrder specifies the order randomly.
	InRandomOrder() Query
	// InTransaction checks if the query is in a transaction.
	InTransaction() bool
	// Join specifying JOIN conditions for the query.
	Join(query string, args ...any) Query
	// LeftJoin specifying LEFT JOIN conditions for the query.
	LeftJoin(query string, args ...any) Query
	// RightJoin specifying RIGHT JOIN conditions for the query.
	RightJoin(query string, args ...any) Query
	// CrossJoin specifying CROSS JOIN conditions for the query.
	CrossJoin(query string, args ...any) Query
	// Limit the number of records returned.
	Limit(limit int) Query
	// Load loads a relationship for the model.
	Load(dest any, relation string, args ...any) error
	// LoadMissing loads a relationship for the model that is not already loaded.
	LoadMissing(dest any, relation string, args ...any) error
	// LockForUpdate locks the selected rows in the table for updating.
	LockForUpdate() Query
	// Model sets the model instance to be queried.
	Model(value any) Query
	// Offset specifies the number of records to skip before starting to return the records.
	Offset(offset int) Query
	// Omit specifies columns that should be omitted from the query.
	Omit(columns ...string) Query
	// Order specifies the order in which the results should be returned.
	Order(value any) Query
	// OrderBy specifies the order should be ascending.
	OrderBy(column string, direction ...string) Query
	// OrderByDesc specifies the order should be descending.
	OrderByDesc(column string) Query
	// OrWhere add an "or where" clause to the query.
	OrWhere(query any, args ...any) Query
	// OrWhereIn adds an "or where column in" clause to the query.
	OrWhereIn(column string, values []any) Query
	// OrWhereNotIn adds an "or where column not in" clause to the query.
	OrWhereNotIn(column string, values []any) Query
	// OrWhereBetween adds an "or where column between x and y" clause to the query.
	OrWhereBetween(column string, x, y any) Query
	// OrWhereNotBetween adds an "or where column not between x and y" clause to the query.
	OrWhereNotBetween(column string, x, y any) Query
	// OrWhereNull adds a "or where column is null" clause to the query.
	OrWhereNull(column string) Query
	// Paginate the given query into a simple paginator.
	Paginate(page, limit int, dest any, total *int64) error
	// Pluck retrieves a single column from the database.
	Pluck(column string, dest any) error
	// Value retrieves a single column's value from the first record.
	Value(column string, dest any) error
	// Raw creates a raw query.
	Raw(sql string, values ...any) Query
	// Restore restores a soft deleted model.
	Restore(model ...any) (*Result, error)
	// Rollback rolls back the changes in a transaction.
	Rollback() error
	// RollbackTo rolls back the changes in a transaction to a specific savepoint.
	RollbackTo(name string) error
	// Save updates value in a database
	Save(value any) error
	// SaveQuietly updates value in a database without firing events
	SaveQuietly(value any) error
	// SavePoint creates a new savepoint in the transaction.
	SavePoint(name string) error
	// Scan scans the query result and populates the destination object.
	Scan(dest any) error
	// Scopes applies one or more query scopes.
	Scopes(funcs ...func(Query) Query) Query
	// Select specifies fields that should be retrieved from the database.
	Select(query any, args ...any) Query
	// SharedLock locks the selected rows in the table.
	SharedLock() Query
	// Avg calculates the average of a column's values and populates the destination object.
	// If the result set is empty, the destination object will be set to its zero value.
	// To distinguish between a zero average and an empty result set, use a pointer as the destination.
	Avg(column string, dest any) error
	// Max calculates the maximum of a column's values and populates the destination object.
	// If the result set is empty, the destination object will be set to its zero value.
	// To distinguish between a zero maximum and an empty result set, use a pointer as the destination.
	Max(column string, dest any) error
	// Min calculates the minimum of a column's values and populates the destination object.
	// If the result set is empty, the destination object will be set to its zero value.
	// To distinguish between a zero minimum and an empty result set, use a pointer as the destination.
	Min(column string, dest any) error
	// Sum calculates the sum of a column's values and populates the destination object.
	// If the result set is empty, the destination object will be set to its zero value.
	// To distinguish between a zero sum and an empty result set, use a pointer as the destination.
	Sum(column string, dest any) error
	// Table specifies the table for the query.
	Table(name string, args ...any) Query
	// ToSql returns the query as a SQL string.
	ToSql() ToSql
	// ToRawSql returns the query as a raw SQL string.
	ToRawSql() ToSql
	// Transaction executes a function within a database transaction.
	Transaction(txFunc func(tx Query) error, opts ...*sql.TxOptions) error
	// Update updates records with the given column and values
	Update(column any, value ...any) (*Result, error)
	// Increment increments a column's value by a given amount.
	Increment(column string, amount ...any) (*Result, error)
	// Decrement decrements a column's value by a given amount.
	Decrement(column string, amount ...any) (*Result, error)
	// UpdateOrCreate finds the first record that matches the given attributes
	// or create a new one with those attributes if none was found.
	UpdateOrCreate(dest any, attributes any, values any) error
	// UpdateOrInsert updates a record in the database or inserts it if it doesn't exist.
	UpdateOrInsert(attributes any, values any) error
	// Where add a "where" clause to the query.
	Where(query any, args ...any) Query
	// WhereIn adds a "where column in" clause to the query.
	WhereIn(column string, values []any) Query
	// WhereNotIn adds a "where column not in" clause to the query.
	WhereNotIn(column string, values []any) Query
	// WhereBetween adds a "where column between x and y" clause to the query.
	WhereBetween(column string, x, y any) Query
	// WhereNotBetween adds a "where column not between x and y" clause to the query.
	WhereNotBetween(column string, x, y any) Query
	// WhereNull adds a "where column is null" clause to the query.
	WhereNull(column string) Query
	// WhereNotNull adds a "where column is not null" clause to the query.
	WhereNotNull(column string) Query
	// WhereColumn adds a "where column1 operator column2" clause to the query.
	WhereColumn(first, operator, second string) Query
	// OrWhereColumn adds an "or where column1 operator column2" clause to the query.
	OrWhereColumn(first, operator, second string) Query
	// WhereExists adds a "where exists" clause to the query.
	WhereExists(callback func(Query) Query) Query
	// WhereNot adds a "where not" clause to the query.
	WhereNot(query any, args ...any) Query
	// OrWhereNot adds an "or where not" clause to the query.
	OrWhereNot(query any, args ...any) Query
	// WhereAny adds a "where any" clause to the query.
	WhereAny(columns []string, operator string, value any) Query
	// WhereAll adds a "where all" clause to the query.
	WhereAll(columns []string, operator string, value any) Query
	// WhereNone adds a "where none" clause to the query.
	WhereNone(columns []string, operator string, value any) Query
	// WhereJsonContains adds a "where json contains" clause to the query.
	WhereJsonContains(column string, value any) Query
	// OrWhereJsonContains adds an "or where json contains" clause to the query.
	OrWhereJsonContains(column string, value any) Query
	// WhereJsonDoesntContain adds a "where json doesn't contain" clause to the query.
	WhereJsonDoesntContain(column string, value any) Query
	// OrWhereJsonDoesntContain adds an "or where json doesn't contain" clause to the query.
	OrWhereJsonDoesntContain(column string, value any) Query
	// WhereJsonContainsKey adds a "where json contains key" clause to the query.
	WhereJsonContainsKey(column string) Query
	// OrWhereJsonContainsKey adds an "or where json contains key" clause to the query.
	OrWhereJsonContainsKey(column string) Query
	// WhereJsonDoesntContainKey adds a "where json doesn't contain key" clause to the query.
	WhereJsonDoesntContainKey(column string) Query
	// OrWhereJsonDoesntContainKey adds an "or where json doesn't contain key" clause to the query.
	OrWhereJsonDoesntContainKey(column string) Query
	// WhereJsonLength adds a "where json length" clause to the query.
	WhereJsonLength(column string, operator string, value any) Query
	// WithoutEvents disables event firing for the query.
	WithoutEvents() Query
	// OnlyTrashed allows only soft deleted models to be included in the results.
	OnlyTrashed() Query
	// WithTrashed allows soft deleted models to be included in the results.
	WithTrashed() Query
	// WithoutTrashed allows soft deleted models to be excluded from the results.
	WithoutTrashed() Query
	// With returns a new query instance with the given relationships eager loaded.
	With(query string, args ...any) Query
	// Without returns a new query instance with the given relationships excluded from eager loading.
	Without(relations ...string) Query
	// WithCount returns a new query instance with the count of the given relationship.
	WithCount(query string, args ...any) Query
	// WithExists returns a new query instance with the existence of the given relationship.
	WithExists(query string, args ...any) Query
	// DisableQueryLog disables the capturing of executed queries.
	DisableQueryLog()
	// EnableQueryLog enables the capturing of executed queries.
	EnableQueryLog()
	// FlushQueryLog clears the captured queries from the log.
	FlushQueryLog()
	// GetQueryLog retrieves the captured queries from the log.
	GetQueryLog() []QueryLog
}

type QueryWithContext interface {
	WithContext(ctx context.Context) Query
}

type QueryWithObserver interface {
	Observe(model any, observer Observer)
}

type Association interface {
	// Find finds records that match given conditions.
	Find(out any, conds ...any) error
	// Append appending a model to the association.
	Append(values ...any) error
	// Replace replaces the association with the given value.
	Replace(values ...any) error
	// Delete deletes the given value from the association.
	Delete(values ...any) error
	// Clear clears the association.
	Clear() error
	// Count returns the number of records in the association.
	Count() int64
}

type ConnectionModel interface {
	// Connection gets the connection name for the model.
	Connection() string
}

type Cursor interface {
	// Scan scans the current row into the given destination.
	Scan(value any) error
}

type Result struct {
	RowsAffected int64
}

type Attribute struct {
	Get func(value any, attributes map[string]any) any
	Set func(value any, attributes map[string]any) any
}

type ToSql interface {
	Count() string
	Create(value any) string
	InsertGetId(values any) string
	Delete(value ...any) string
	Find(dest any, conds ...any) string
	First(dest any) string
	ForceDelete(value ...any) string
	Get(dest any) string
	Pluck(column string, dest any) string
	Value(column string, dest any) string
	Save(value any) string
	Avg(column string, dest any) string
	Max(column string, dest any) string
	Min(column string, dest any) string
	Sum(column string, dest any) string
	Update(column any, value ...any) string
	Increment(column string, amount ...any) string
	Decrement(column string, amount ...any) string
}
