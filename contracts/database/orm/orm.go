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
	// The name parameter specifies which connection to retrieve.
	// If name is empty, the default connection is returned.
	//
	// Example:
	//   secondary := db.Connection("secondary")
	Connection(name string) Orm
	// DB gets the underlying database connection.
	// Returns the *sql.DB instance for direct database operations.
	// Returns an error if the connection is not available.
	DB() (*sql.DB, error)
	// DisableQueryLog disables the capturing of executed queries.
	// When disabled, queries are not logged to memory.
	DisableQueryLog()
	// EnableQueryLog enables the capturing of executed queries.
	// When enabled, all executed queries are logged to memory for debugging.
	EnableQueryLog()
	// FlushQueryLog clears the captured queries from the log.
	// This removes all previously logged queries from memory.
	FlushQueryLog()
	// GetQueryLog retrieves the captured queries from the log.
	// Returns a slice of QueryLog containing all executed queries since logging was enabled.
	GetQueryLog() []QueryLog
	// EnableDebug enables debug mode at runtime for all queries.
	// When enabled, detailed SQL error messages are shown instead of generic errors.
	EnableDebug()
	// DisableDebug disables debug mode at runtime for all queries.
	// When disabled, SQL error messages are sanitized for security.
	DisableDebug()
	// IsDebug returns true if debug mode is enabled.
	IsDebug() bool
	// Factory gets a new factory instance for the given model name.
	// Factories are used for generating test data and seeding databases.
	Factory() Factory
	// DatabaseName gets the current database name.
	// Returns the name of the database for the current connection.
	DatabaseName() string
	// Name gets the current connection name.
	// Returns the name of the connection (e.g., "default", "secondary").
	Name() string
	// Observe registers an observer with the Orm.
	// Observers can listen to model events like creating, updating, deleting.
	// The model parameter specifies which model to observe.
	// The observer parameter is the observer instance to register.
	Observe(model any, observer Observer)
	// Query gets a new query builder instance.
	// Returns a fresh Query instance for building database queries.
	Query() Query
	// Refresh resets the Orm instance.
	// Clears any cached state and resets the instance to its initial state.
	Refresh()
	// SetQuery sets the query builder instance.
	// Allows replacing the current query builder with a custom instance.
	SetQuery(query Query)
	// Transaction runs a callback wrapped in a database transaction.
	// The callback receives a transaction query object.
	// If the callback returns nil, the transaction is committed.
	// If the callback returns an error, the transaction is rolled back.
	// Options can be provided to configure transaction isolation level, etc.
	//
	// Example:
	//   err := db.Transaction(func(tx orm.Query) error {
	//       if err := tx.Create(&user); err != nil {
	//           return err
	//       }
	//       return tx.Create(&profile)
	//   })
	Transaction(txFunc func(tx Query) error, opts ...*sql.TxOptions) error
	// WithContext sets the context to be used by the Orm.
	// The context is used for query timeouts, cancellation, and request-scoped values.
	// Returns a new Orm instance with the context set.
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
	// Count retrieves the number of records matching the query conditions.
	// The count parameter must be a pointer to an int64.
	//
	// Example:
	//   var count int64
	//   err := query.Count(&count)
	Count(count *int64) error
	// Chunk the results of the query into smaller batches.
	// Processes records in chunks of the specified size to reduce memory usage.
	// The callback receives each chunk of records.
	// Returns an error if the callback fails or chunking encounters an error.
	//
	// Example:
	//   err := query.Chunk(100, func(users []User) error {
	//       for _, user := range users {
	//           fmt.Println(user.Name)
	//       }
	//       return nil
	//   })
	Chunk(size int, callback any) error
	// Create inserts a new record into the database.
	// The value parameter must be a pointer to a struct or a struct value.
	// The struct's TableName() method or the struct name (lowercased) determines the table.
	//
	// Example:
	//   user := User{Name: "John", Email: "john@example.com"}
	//   err := query.Create(&user)
	Create(value any) error
	// InsertGetId inserts a new record and returns the last inserted ID.
	// The values parameter must be a pointer to a struct or a struct value.
	// Returns the ID of the inserted record.
	// Returns an error if the insert fails.
	//
	// Example:
	//   user := User{Name: "John", Email: "john@example.com"}
	//   id, err := query.InsertGetId(&user)
	InsertGetId(values any) (uint, error)
	// Cursor returns a cursor for efficient iteration over large result sets.
	// Use Scan to iterate over the returned rows one at a time.
	// This is memory-efficient for processing large datasets.
	// Returns a channel of Cursor instances and an error if the query fails.
	//
	// Example:
	//   cursor, err := query.Cursor()
	//   if err != nil {
	//       return err
	//   }
	//   for row := range cursor {
	//       var user User
	//       if err := row.Scan(&user); err != nil {
	//           return err
	//       }
	//       fmt.Println(user.Name)
	//   }
	Cursor() (chan Cursor, error)
	// DB gets the underlying database connection.
	DB() (*sql.DB, error)
	// Delete deletes records matching the query conditions.
	// If soft deletes are enabled on the model, this sets a deleted_at timestamp instead of permanently deleting.
	// Use ForceDelete for permanent deletion when soft deletes are enabled.
	// Returns the number of rows affected.
	//
	// Example:
	//   result, err := query.Where("status = ?", "inactive").Delete()
	Delete(value ...any) (*Result, error)
	// Distinct specifies distinct fields to query.
	// Removes duplicate rows from the result set based on the specified columns.
	// Args can be column names or expressions.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.Distinct("email")
	//   query.Distinct("first_name", "last_name")
	Distinct(args ...any) Query
	// Driver gets the driver for the query.
	// Returns the database driver instance for driver-specific operations.
	Driver() database.Driver
	// Exec executes raw SQL with optional parameter values.
	// The sql parameter is the raw SQL statement to execute.
	// The values parameter provides parameter values for prepared statements.
	// Returns the result containing rows affected and any error.
	//
	// Example:
	//   result, err := query.Exec("UPDATE users SET status = ? WHERE id = ?", "active", 1)
	Exec(sql string, values ...any) (*Result, error)
	// Exists checks if matching records exist in the database.
	// The exists parameter must be a pointer to a bool.
	// Sets the bool to true if records match the query conditions, false otherwise.
	// Returns an error if the query fails.
	//
	// Example:
	//   var exists bool
	//   err := query.Where("email = ?", "test@example.com").Exists(&exists)
	Exists(exists *bool) error
	// Find retrieves all records matching the given conditions.
	// The dest parameter must be a pointer to a slice (e.g., *[]User).
	// Conditions can be provided as variadic arguments for WHERE clauses.
	//
	// Example:
	//   var users []User
	//   err := query.Find(&users)
	//
	//   var users []User
	//   err := query.Find(&users, "age > ?", 18)
	Find(dest any, conds ...any) error
	// FindOrFail finds records that match given conditions or returns an error.
	// Similar to Find, but returns an error if no records are found.
	// The dest parameter must be a pointer to a slice (e.g., *[]User).
	// Conditions can be provided as variadic arguments for WHERE clauses.
	// Returns ErrRecordNotFound if no records match the conditions.
	//
	// Example:
	//   var users []User
	//   err := query.FindOrFail(&users, "age > ?", 18)
	FindOrFail(dest any, conds ...any) error
	// First retrieves the first record matching the query conditions.
	// The dest parameter must be a pointer to a struct (e.g., *User).
	// Returns an error if no record is found.
	//
	// Example:
	//   var user User
	//   err := query.First(&user)
	//
	//   var user User
	//   err := query.Where("email = ?", "test@example.com").First(&user)
	First(dest any) error
	// FirstOrCreate finds the first record matching the given attributes,
	// or creates a new one with those attributes if none was found.
	// The dest parameter must be a pointer to a struct (e.g., *User).
	// If a record is found, it's loaded into dest.
	// If no record is found, a new one is created with the given conditions.
	// Returns an error if the find or create operation fails.
	//
	// Example:
	//   var user User
	//   err := query.FirstOrCreate(&user, "email = ?", "john@example.com")
	FirstOrCreate(dest any, conds ...any) error
	// FirstOr finds the first record matching the given conditions,
	// or executes the callback and returns its result if no record is found.
	// The dest parameter must be a pointer to a struct (e.g., *User).
	// The callback is executed if no record is found, allowing custom logic.
	// Returns an error if the find or callback operation fails.
	//
	// Example:
	//   var user User
	//   err := query.FirstOr(&user, func() error {
	//       user = User{Name: "Default User"}
	//       return nil
	//   })
	FirstOr(dest any, callback func() error) error
	// FirstOrFail finds the first record matching the given conditions or returns an error.
	// Similar to First, but returns an error if no record is found.
	// The dest parameter must be a pointer to a struct (e.g., *User).
	// Returns ErrRecordNotFound if no record matches the conditions.
	//
	// Example:
	//   var user User
	//   err := query.Where("email = ?", "test@example.com").FirstOrFail(&user)
	FirstOrFail(dest any) error
	// FirstOrNew finds the first record matching the given conditions,
	// or returns a new instance of the model initialized with those attributes.
	// The dest parameter must be a pointer to a struct (e.g., *User).
	// The attributes parameter specifies the conditions to match.
	// The values parameter provides additional values for the new instance.
	// If a record is found, it's loaded into dest.
	// If no record is found, dest is initialized with the attributes and values.
	// Returns an error if the find operation fails.
	//
	// Example:
	//   var user User
	//   err := query.FirstOrNew(&user, map[string]any{"email": "john@example.com"})
	FirstOrNew(dest any, attributes any, values ...any) error
	// ForceDelete permanently deletes records matching the query conditions.
	// Unlike Delete, this bypasses soft delete mechanisms and permanently removes records.
	// Returns the number of rows affected and any error.
	//
	// Example:
	//   result, err := query.Where("status = ?", "inactive").ForceDelete()
	ForceDelete(value ...any) (*Result, error)
	// Get retrieves all rows from the database matching the query conditions.
	// The dest parameter must be a pointer to a slice of structs or maps.
	// Returns an error if the query fails.
	//
	// Example:
	//   var users []User
	//   err := query.Where("age > ?", 18).Get(&users)
	//
	//   var results []map[string]any
	//   err := query.Table("users").Get(&results)
	Get(dest any) error
	// Group specifies a GROUP BY clause for the query.
	// Groups results by the specified column or expression.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.Group("status")
	//   query.Group("department", "role")
	Group(name string) Query
	// Having specifies HAVING conditions for the query.
	// Used with GROUP BY to filter grouped results.
	// Query can be a string (SQL fragment) or a map.
	// Args provides parameter values for prepared statements.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.Group("status").Having("COUNT(*) > ?", 10)
	Having(query any, args ...any) Query
	// InRandomOrder specifies that results should be returned in random order.
	// Useful for sampling or random selection.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.InRandomOrder().Limit(10)
	InRandomOrder() Query
	// InTransaction checks if the query is currently within a transaction.
	// Returns true if the query is part of an active transaction, false otherwise.
	InTransaction() bool
	// Join specifies INNER JOIN conditions for the query.
	// The query parameter specifies the join condition (e.g., "users ON orders.user_id = users.id").
	// Args provides parameter values for prepared statements.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.Join("users ON orders.user_id = users.id")
	Join(query string, args ...any) Query
	// LeftJoin specifies LEFT JOIN conditions for the query.
	// The query parameter specifies the join condition.
	// Args provides parameter values for prepared statements.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.LeftJoin("profiles ON users.id = profiles.user_id")
	LeftJoin(query string, args ...any) Query
	// RightJoin specifies RIGHT JOIN conditions for the query.
	// The query parameter specifies the join condition.
	// Args provides parameter values for prepared statements.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.RightJoin("orders ON users.id = orders.user_id")
	RightJoin(query string, args ...any) Query
	// CrossJoin specifies CROSS JOIN conditions for the query.
	// The query parameter specifies the table to join.
	// Args provides parameter values for prepared statements.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.CrossJoin("categories")
	CrossJoin(query string, args ...any) Query
	// Limit restricts the number of records returned by the query.
	// The limit parameter specifies the maximum number of records to return.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.Limit(10)
	Limit(limit int) Query
	// Load eager loads a relationship for the model.
	// The dest parameter is the model instance to load the relationship for.
	// The relation parameter specifies the relationship name (e.g., "posts", "profile").
	// Args provides additional conditions for the relationship query.
	// Returns an error if the load operation fails.
	//
	// Example:
	//   var user User
	//   err := query.First(&user)
	//   err = query.Load(&user, "posts")
	Load(dest any, relation string, args ...any) error
	// LoadMissing eager loads a relationship only if it's not already loaded.
	// The dest parameter is the model instance to load the relationship for.
	// The relation parameter specifies the relationship name.
	// Args provides additional conditions for the relationship query.
	// Returns an error if the load operation fails.
	// This is more efficient than Load for relationships that may already be loaded.
	LoadMissing(dest any, relation string, args ...any) error
	// LockForUpdate locks the selected rows in the table for updating.
	// This prevents other transactions from modifying the selected rows.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.Where("id = ?", 1).LockForUpdate().First(&user)
	LockForUpdate() Query
	// Model sets the model instance to be queried.
	// The value parameter must be a struct or pointer to a struct.
	// The model's TableName() method or struct name (lowercased) determines the table.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.Model(&User{})
	//   query.Model(user)
	Model(value any) Query
	// Offset specifies the number of records to skip before returning results.
	// Often used with Limit for pagination.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.Offset(10).Limit(10) // Skip first 10, return next 10
	Offset(offset int) Query
	// Omit specifies columns that should be omitted from the query.
	// Useful for excluding sensitive fields or large columns from results.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.Omit("password", "secret_key")
	Omit(columns ...string) Query
	// Order specifies the order in which results should be returned.
	// The value parameter can be a string (column name) or expression.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.Order("created_at DESC")
	//   query.Order("name ASC")
	Order(value any) Query
	// OrderBy specifies that results should be ordered by a column in ascending order.
	// The column parameter specifies the column to order by.
	// The optional direction parameter can override the default ASC direction.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.OrderBy("name")
	//   query.OrderBy("created_at", "DESC")
	OrderBy(column string, direction ...string) Query
	// OrderByDesc specifies that results should be ordered by a column in descending order.
	// The column parameter specifies the column to order by.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.OrderByDesc("created_at")
	OrderByDesc(column string) Query
	// OrWhere adds an OR WHERE clause to the query.
	// Query can be a string (SQL fragment) or a map[string]any for multiple conditions.
	// Args provides parameter values for prepared statements.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.Where("status = ?", "active").OrWhere("status = ?", "pending")
	OrWhere(query any, args ...any) Query
	// OrWhereIn adds an OR WHERE column IN clause to the query.
	// The column parameter specifies the column to check.
	// The values parameter provides the list of values to match against.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.Where("status = ?", "active").OrWhereIn("id", []any{1, 2, 3})
	OrWhereIn(column string, values []any) Query
	// OrWhereNotIn adds an OR WHERE column NOT IN clause to the query.
	// The column parameter specifies the column to check.
	// The values parameter provides the list of values to exclude.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.Where("status = ?", "active").OrWhereNotIn("id", []any{1, 2, 3})
	OrWhereNotIn(column string, values []any) Query
	// OrWhereBetween adds an OR WHERE column BETWEEN clause to the query.
	// The column parameter specifies the column to check.
	// The x and y parameters specify the range boundaries.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.Where("status = ?", "active").OrWhereBetween("age", 18, 65)
	OrWhereBetween(column string, x, y any) Query
	// OrWhereNotBetween adds an OR WHERE column NOT BETWEEN clause to the query.
	// The column parameter specifies the column to check.
	// The x and y parameters specify the range boundaries.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.Where("status = ?", "active").OrWhereNotBetween("age", 18, 65)
	OrWhereNotBetween(column string, x, y any) Query
	// OrWhereNull adds an OR WHERE column IS NULL clause to the query.
	// The column parameter specifies the column to check.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.Where("status = ?", "active").OrWhereNull("deleted_at")
	OrWhereNull(column string) Query
	// Paginate paginates the query results into pages.
	// The page parameter specifies the current page number (1-based).
	// The limit parameter specifies the number of records per page.
	// The dest parameter must be a pointer to a slice for the results.
	// The total parameter must be a pointer to int64 for the total record count.
	// Returns an error if the query fails.
	//
	// Example:
	//   var users []User
	//   var total int64
	//   err := query.Paginate(1, 10, &users, &total)
	Paginate(page, limit int, dest any, total *int64) error
	// Pluck retrieves a single column's values from the database.
	// The column parameter specifies the column to retrieve.
	// The dest parameter must be a pointer to a slice (e.g., *[]string, *[]int).
	// Returns an error if the query fails.
	//
	// Example:
	//   var names []string
	//   err := query.Pluck("name", &names)
	Pluck(column string, dest any) error
	// Value retrieves a single column's value from the first matching record.
	// The column parameter specifies the column to retrieve.
	// The dest parameter must be a pointer to the appropriate type.
	// Returns an error if no record is found or the query fails.
	//
	// Example:
	//   var name string
	//   err := query.Where("id = ?", 1).Value("name", &name)
	Value(column string, dest any) error
	// Raw creates a raw SQL query.
	// The sql parameter is the raw SQL statement to execute.
	// The values parameter provides parameter values for prepared statements.
	// Returns a query instance for method chaining.
	//
	// Example:
	//   query.Raw("SELECT * FROM users WHERE status = ?", "active")
	Raw(sql string, values ...any) Query
	// RestoreSoftDeleted restores a soft-deleted model.
	// Sets the soft-delete timestamp column to NULL for the specified model(s).
	// Returns the number of rows affected and any error.
	//
	// Example:
	//   result, err := query.Where("id = ?", 1).RestoreSoftDeleted()
	RestoreSoftDeleted(model ...any) (*Result, error)
	// Restore restores a soft-deleted model.
	//
	// Deprecated: Use RestoreSoftDeleted() instead.
	Restore(model ...any) (*Result, error)
	// Rollback rolls back the current transaction.
	// All changes made in the transaction are discarded.
	// Returns an error if the rollback fails.
	Rollback() error
	// RollbackTo rolls back the transaction to a specific savepoint.
	// The name parameter specifies the savepoint name.
	// Returns an error if the rollback fails.
	//
	// Example:
	//   query.SavePoint("before_update")
	//   // ... perform operations ...
	//   err := query.RollbackTo("before_update")
	RollbackTo(name string) error
	// Save updates a model in the database.
	// If the model has a primary key value, it performs an UPDATE.
	// If the model has no primary key value, it performs an INSERT.
	// The value parameter must be a pointer to a struct.
	// Returns an error if the save operation fails.
	//
	// Example:
	//   user := User{Name: "John", Email: "john@example.com"}
	//   err := query.Save(&user) // INSERT
	//   user.Name = "Jane"
	//   err = query.Save(&user) // UPDATE
	Save(value any) error
	// SaveQuietly updates a model in the database without firing events.
	// Similar to Save, but does not trigger model events (creating, updating, etc.).
	// The value parameter must be a pointer to a struct.
	// Returns an error if the save operation fails.
	//
	// Example:
	//   err := query.SaveQuietly(&user)
	SaveQuietly(value any) error
	// SavePoint creates a new savepoint in the current transaction.
	// Savepoints allow partial rollback within a transaction.
	// The name parameter specifies the savepoint name.
	// Returns an error if the savepoint creation fails.
	//
	// Example:
	//   err := query.SavePoint("before_update")
	SavePoint(name string) error
	// Scan scans the query result and populates the destination object.
	// The dest parameter can be a struct, slice of structs, or map.
	// Useful for raw queries or when you need custom result mapping.
	// Returns an error if the scan fails.
	//
	// Example:
	//   var user User
	//   err := query.Raw("SELECT * FROM users WHERE id = ?", 1).Scan(&user)
	Scan(dest any) error
	// Scopes applies one or more query scopes to the query.
	// Scopes are reusable query building functions.
	// Each scope function receives and returns a Query instance.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   func Active(query Query) Query {
	//       return query.Where("status = ?", "active")
	//   }
	//   query.Scopes(Active)
	Scopes(funcs ...func(Query) Query) Query
	// Select specifies which columns should be retrieved from the database.
	// Query can be a string (column name or comma-separated list) or a slice of strings.
	// By default, all columns are selected.
	//
	// Example:
	//   query.Select("name, email")
	//
	//   query.Select([]string{"name", "email"})
	Select(query any, args ...any) Query
	// SharedLock locks the selected rows in the table for reading.
	// Other transactions can read but not modify the locked rows.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.Where("id = ?", 1).SharedLock().First(&user)
	SharedLock() Query
	// Avg calculates the average of a column's values.
	// The column parameter specifies the column to average.
	// The dest parameter must be a pointer to a numeric type (float64, int, etc.).
	// If the result set is empty, dest is set to its zero value.
	// Use a pointer to distinguish between zero average and empty result set.
	// Returns an error if the query fails.
	//
	// Example:
	//   var avg float64
	//   err := query.Avg("price", &avg)
	Avg(column string, dest any) error
	// Max calculates the maximum value of a column.
	// The column parameter specifies the column to find the maximum of.
	// The dest parameter must be a pointer to the appropriate type.
	// If the result set is empty, dest is set to its zero value.
	// Use a pointer to distinguish between zero maximum and empty result set.
	// Returns an error if the query fails.
	//
	// Example:
	//   var maxPrice float64
	//   err := query.Max("price", &maxPrice)
	Max(column string, dest any) error
	// Min calculates the minimum value of a column.
	// The column parameter specifies the column to find the minimum of.
	// The dest parameter must be a pointer to the appropriate type.
	// If the result set is empty, dest is set to its zero value.
	// Use a pointer to distinguish between zero minimum and empty result set.
	// Returns an error if the query fails.
	//
	// Example:
	//   var minPrice float64
	//   err := query.Min("price", &minPrice)
	Min(column string, dest any) error
	// Sum calculates the sum of a column's values.
	// The column parameter specifies the column to sum.
	// The dest parameter must be a pointer to a numeric type.
	// If the result set is empty, dest is set to its zero value.
	// Use a pointer to distinguish between zero sum and empty result set.
	// Returns an error if the query fails.
	//
	// Example:
	//   var total float64
	//   err := query.Sum("price", &total)
	Sum(column string, dest any) error
	// Table specifies the table name for the query.
	// If the model is set via Model(), this is typically not needed.
	// Args can be used for table aliases or dynamic table names.
	//
	// Example:
	//   query.Table("users")
	//
	//   query.Table("users as u")
	Table(name string, args ...any) Query
	// ToSql returns the query as a SQL string.
	ToSql() ToSql
	// ToRawSql returns the query as a raw SQL string.
	ToRawSql() ToSql
	// Transaction executes a function within a database transaction.
	// The function receives a transaction query object.
	// If the function returns nil, the transaction is committed.
	// If the function returns an error, the transaction is rolled back.
	// Options can be provided to configure transaction isolation level, etc.
	//
	// Example:
	//   err := query.Transaction(func(tx Query) error {
	//       if err := tx.Create(&user); err != nil {
	//           return err
	//       }
	//       return tx.Create(&profile)
	//   })
	Transaction(txFunc func(tx Query) error, opts ...*sql.TxOptions) error
	// Update updates records matching the query conditions.
	// Column can be a string (column name) or a map[string]any for multiple columns.
	// When column is a string, value is the new value for that column.
	// When column is a map, value is ignored.
	// Returns the number of rows affected.
	//
	// Example:
	//   result, err := query.Update("status", "active")
	//
	//   result, err := query.Update(map[string]any{"status": "active", "updated_at": time.Now()})
	Update(column any, value ...any) (*Result, error)
	// Increment increments a column's value by a given amount.
	// The column parameter specifies the column to increment.
	// The optional amount parameter specifies the increment amount (default is 1).
	// Returns the number of rows affected and any error.
	//
	// Example:
	//   result, err := query.Where("id = ?", 1).Increment("views")
	//   result, err := query.Increment("stock", 5)
	Increment(column string, amount ...any) (*Result, error)
	// Decrement decrements a column's value by a given amount.
	// The column parameter specifies the column to decrement.
	// The optional amount parameter specifies the decrement amount (default is 1).
	// Returns the number of rows affected and any error.
	//
	// Example:
	//   result, err := query.Where("id = ?", 1).Decrement("stock")
	//   result, err := query.Decrement("quantity", 5)
	Decrement(column string, amount ...any) (*Result, error)
	// UpdateOrCreate finds the first record matching the given attributes,
	// or creates a new one with those attributes if none was found.
	// The dest parameter must be a pointer to a struct for the result.
	// The attributes parameter specifies the conditions to match.
	// The values parameter provides the values to update/create with.
	// If a record is found, it's updated with values.
	// If no record is found, a new one is created with attributes and values.
	// Returns an error if the operation fails.
	//
	// Example:
	//   var user User
	//   err := query.UpdateOrCreate(&user, map[string]any{"email": "john@example.com"}, map[string]any{"name": "John"})
	UpdateOrCreate(dest any, attributes any, values any) error
	// UpdateOrInsert updates a record in the database or inserts it if it doesn't exist.
	// The attributes parameter specifies the conditions to match.
	// The values parameter provides the values to update/insert with.
	// If a matching record is found, it's updated with values.
	// If no matching record is found, a new one is inserted with attributes and values.
	// Returns an error if the operation fails.
	//
	// Example:
	//   err := query.UpdateOrInsert(map[string]any{"email": "john@example.com"}, map[string]any{"name": "John"})
	UpdateOrInsert(attributes any, values any) error
	// Where adds a WHERE clause to the query.
	// Query can be a string (SQL fragment) or a map[string]any for multiple conditions.
	// When query is a string, args provides the parameter values.
	// When query is a map, args is ignored.
	//
	// Example:
	//   query.Where("age > ?", 18)
	//
	//   query.Where(map[string]any{"status": "active", "age > ?": 18})
	Where(query any, args ...any) Query
	// WhereIn adds a WHERE column IN clause to the query.
	// The column parameter specifies the column to check.
	// The values parameter provides the list of values to match against.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.WhereIn("id", []any{1, 2, 3})
	WhereIn(column string, values []any) Query
	// WhereNotIn adds a WHERE column NOT IN clause to the query.
	// The column parameter specifies the column to check.
	// The values parameter provides the list of values to exclude.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.WhereNotIn("id", []any{1, 2, 3})
	WhereNotIn(column string, values []any) Query
	// WhereBetween adds a WHERE column BETWEEN clause to the query.
	// The column parameter specifies the column to check.
	// The x and y parameters specify the range boundaries.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.WhereBetween("age", 18, 65)
	WhereBetween(column string, x, y any) Query
	// WhereNotBetween adds a WHERE column NOT BETWEEN clause to the query.
	// The column parameter specifies the column to check.
	// The x and y parameters specify the range boundaries.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.WhereNotBetween("age", 18, 65)
	WhereNotBetween(column string, x, y any) Query
	// WhereNull adds a WHERE column IS NULL clause to the query.
	// The column parameter specifies the column to check.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.WhereNull("deleted_at")
	WhereNull(column string) Query
	// WhereNotNull adds a WHERE column IS NOT NULL clause to the query.
	// The column parameter specifies the column to check.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.WhereNotNull("deleted_at")
	WhereNotNull(column string) Query
	// WhereColumn adds a WHERE column1 operator column2 clause to the query.
	// Compares two columns directly without using values.
	// The first and second parameters specify the columns to compare.
	// The operator parameter specifies the comparison operator (=, >, <, etc.).
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.WhereColumn("updated_at", ">", "created_at")
	WhereColumn(first, operator, second string) Query
	// OrWhereColumn adds an OR WHERE column1 operator column2 clause to the query.
	// Compares two columns directly without using values.
	// The first and second parameters specify the columns to compare.
	// The operator parameter specifies the comparison operator (=, >, <, etc.).
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.Where("status = ?", "active").OrWhereColumn("updated_at", ">", "created_at")
	OrWhereColumn(first, operator, second string) Query
	// WhereExists adds a WHERE EXISTS clause to the query.
	// The callback receives a query builder for the subquery.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.WhereExists(func(q Query) Query {
	//       return q.Table("orders").Where("user_id = users.id")
	//   })
	WhereExists(callback func(Query) Query) Query
	// WhereNot adds a WHERE NOT clause to the query.
	// Query can be a string (SQL fragment) or a map[string]any for multiple conditions.
	// Args provides parameter values for prepared statements.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.WhereNot("status = ?", "inactive")
	WhereNot(query any, args ...any) Query
	// OrWhereNot adds an OR WHERE NOT clause to the query.
	// Query can be a string (SQL fragment) or a map[string]any for multiple conditions.
	// Args provides parameter values for prepared statements.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.Where("status = ?", "active").OrWhereNot("deleted = ?", true)
	OrWhereNot(query any, args ...any) Query
	// WhereAny adds a WHERE ANY clause for JSON arrays.
	// Checks if any element in the JSON array matches the condition.
	// The columns parameter specifies the JSON columns to check.
	// The operator parameter specifies the comparison operator.
	// The value parameter specifies the value to compare against.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.WhereAny([]string{"tags"}, "=", "important")
	WhereAny(columns []string, operator string, value any) Query
	// WhereAll adds a WHERE ALL clause for JSON arrays.
	// Checks if all elements in the JSON array match the condition.
	// The columns parameter specifies the JSON columns to check.
	// The operator parameter specifies the comparison operator.
	// The value parameter specifies the value to compare against.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.WhereAll([]string{"tags"}, "=", "important")
	WhereAll(columns []string, operator string, value any) Query
	// WhereNone adds a WHERE NONE clause for JSON arrays.
	// Checks if no elements in the JSON array match the condition.
	// The columns parameter specifies the JSON columns to check.
	// The operator parameter specifies the comparison operator.
	// The value parameter specifies the value to compare against.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.WhereNone([]string{"tags"}, "=", "important")
	WhereNone(columns []string, operator string, value any) Query
	// WhereJsonContains adds a WHERE JSON contains clause to the query.
	// Checks if a JSON column contains the specified value.
	// The column parameter specifies the JSON column to check.
	// The value parameter specifies the value to search for.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.WhereJsonContains("tags", "\"important\"")
	WhereJsonContains(column string, value any) Query
	// OrWhereJsonContains adds an OR WHERE JSON contains clause to the query.
	// Checks if a JSON column contains the specified value.
	// The column parameter specifies the JSON column to check.
	// The value parameter specifies the value to search for.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.Where("status = ?", "active").OrWhereJsonContains("tags", "\"important\"")
	OrWhereJsonContains(column string, value any) Query
	// WhereJsonDoesntContain adds a WHERE JSON doesn't contain clause to the query.
	// Checks if a JSON column does not contain the specified value.
	// The column parameter specifies the JSON column to check.
	// The value parameter specifies the value to search for.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.WhereJsonDoesntContain("tags", "\"important\"")
	WhereJsonDoesntContain(column string, value any) Query
	// OrWhereJsonDoesntContain adds an OR WHERE JSON doesn't contain clause to the query.
	// Checks if a JSON column does not contain the specified value.
	// The column parameter specifies the JSON column to check.
	// The value parameter specifies the value to search for.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.Where("status = ?", "active").OrWhereJsonDoesntContain("tags", "\"important\"")
	OrWhereJsonDoesntContain(column string, value any) Query
	// WhereJsonContainsKey adds a WHERE JSON contains key clause to the query.
	// Checks if a JSON column contains the specified key.
	// The column parameter specifies the JSON column to check.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.WhereJsonContainsKey("metadata")
	WhereJsonContainsKey(column string) Query
	// OrWhereJsonContainsKey adds an OR WHERE JSON contains key clause to the query.
	// Checks if a JSON column contains the specified key.
	// The column parameter specifies the JSON column to check.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.Where("status = ?", "active").OrWhereJsonContainsKey("metadata")
	OrWhereJsonContainsKey(column string) Query
	// WhereJsonDoesntContainKey adds a WHERE JSON doesn't contain key clause to the query.
	// Checks if a JSON column does not contain the specified key.
	// The column parameter specifies the JSON column to check.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.WhereJsonDoesntContainKey("metadata")
	WhereJsonDoesntContainKey(column string) Query
	// OrWhereJsonDoesntContainKey adds an OR WHERE JSON doesn't contain key clause to the query.
	// Checks if a JSON column does not contain the specified key.
	// The column parameter specifies the JSON column to check.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.Where("status = ?", "active").OrWhereJsonDoesntContainKey("metadata")
	OrWhereJsonDoesntContainKey(column string) Query
	// WhereJsonLength adds a WHERE JSON length clause to the query.
	// Checks the length of a JSON array or object.
	// The column parameter specifies the JSON column to check.
	// The operator parameter specifies the comparison operator (=, >, <, etc.).
	// The value parameter specifies the length value to compare against.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.WhereJsonLength("tags", ">", 5)
	WhereJsonLength(column string, operator string, value any) Query
	// WithoutEvents disables event firing for the query.
	// Model events (creating, updating, deleting, etc.) will not be triggered.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.WithoutEvents().Save(&user)
	WithoutEvents() Query
	// OnlyTrashed includes only soft deleted models in the results.
	// Filters results to only include records where deleted_at is not NULL.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.OnlyTrashed().Find(&users)
	OnlyTrashed() Query
	// WithTrashed includes soft deleted models in the results.
	// By default, soft deleted records are excluded.
	// This method includes them in the results.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.WithTrashed().Find(&users)
	WithTrashed() Query
	// WithoutTrashed excludes soft deleted models from the results.
	// This is the default behavior, but can be used to reset WithTrashed.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.WithTrashed().WithoutTrashed().Find(&users)
	WithoutTrashed() Query
	// With eager loads the specified relationships.
	// The query parameter specifies the relationship name(s) to load.
	// Args provides additional conditions for the relationship query.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.With("posts")
	//   query.With("posts", "profile")
	//   query.With("posts.status = ?", "published")
	With(query string, args ...any) Query
	// Without excludes the specified relationships from eager loading.
	// The relations parameter specifies the relationship names to exclude.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.Without("posts", "profile")
	Without(relations ...string) Query
	// WithCount eager loads the count of the specified relationship.
	// Adds a {relationship}_count attribute to the model.
	// The query parameter specifies the relationship name.
	// Args provides additional conditions for the relationship query.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.WithCount("posts")
	WithCount(query string, args ...any) Query
	// WithExists adds a relationship existence check to the query.
	// Filters results to only include records that have the specified relationship.
	// The query parameter specifies the relationship name.
	// Args provides additional conditions for the relationship query.
	// Returns the query instance for method chaining.
	//
	// Example:
	//   query.WithExists("posts")
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
