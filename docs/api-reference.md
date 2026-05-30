# API Reference

This document provides a comprehensive reference for the Neat ORM API. For detailed usage examples and best practices, see the specific documentation for each component.

## Table of Contents

- [Database](#database)
- [Query Builder](#query-builder)
- [Schema Builder](#schema-builder)
- [ORM](#orm)
- [Factory](#factory)
- [Seeder](#seeder)
- [Migration](#migration)
- [Observer](#observer)
- [Association](#association)

## Database

### Configuration

```go
type DBConfig struct {
    Default     string
    Connections map[string]ConnectionConfig
    Migrations  MigrationConfig
}

type ConnectionConfig struct {
    Driver   string
    Host     string
    Port     int
    Database string
    Username string
    Password string
    Options  map[string]string
    Pool     PoolConfig
}

type PoolConfig struct {
    MaxOpenConns int
    MaxIdleConns int
    MaxLifetime  int
}
```

### Methods

- `New(config DBConfig) (*Database, error)` - Create database from config
- `NewFromDSN(dsn string) (*Database, error)` - Create database from DSN string
- `Close() error` - Close database connection
- `Query() Query` - Get query builder instance
- `Schema() Schema` - Get schema builder instance
- `Orm() Orm` - Get ORM instance
- `Factory() Factory` - Get factory instance
- `Seeder() Facade` - Get seeder facade
- `Transaction(fn func(Query) error) error` - Execute transaction
- `Migrate(paths ...string) error` - Run migrations
- `MigrateDown(steps int, paths ...string) error` - Rollback migrations
- `MigrateFresh(paths ...string) error` - Drop all tables and re-run migrations
- `MigrateReset(paths ...string) error` - Rollback all and re-run migrations
- `MigrationStatus(paths ...string) ([]MigrationStatus, error)` - Get migration status
- `Seed(seeders []Seeder) error` - Run seeders
- `SeedOnce(seeders []Seeder) error` - Run seeders only once

**See also:** [Configuration Guide](./configuration.md)

## Query Builder

### Query Methods

#### Selection

- `Get(dest any) error` - Execute query and populate dest
- `First(dest any) error` - Get first matching record
- `Find(dest any, id any) error` - Find record by ID
- `Select(columns ...string) Query` - Select specific columns
- `SelectRaw(sql string, bindings ...any) Query` - Select with raw SQL

#### Where Clauses

- `Where(column string, value ...any) Query` - Add WHERE clause
- `OrWhere(column string, value ...any) Query` - Add OR WHERE clause
- `WhereIn(column string, values []any) Query` - Add WHERE IN clause
- `WhereNotIn(column string, values []any) Query` - Add WHERE NOT IN clause
- `WhereNull(column string) Query` - Add WHERE NULL clause
- `WhereNotNull(column string) Query` - Add WHERE NOT NULL clause
- `WhereBetween(column string, min, max any) Query` - Add WHERE BETWEEN clause
- `WhereRaw(sql string, bindings ...any) Query` - Add raw WHERE clause

#### Ordering and Limiting

- `OrderBy(column string, direction string) Query` - Add ORDER BY clause
- `OrderByRaw(sql string, bindings ...any) Query` - Add raw ORDER BY
- `Limit(limit int) Query` - Set result limit
- `Offset(offset int) Query` - Set result offset
- `Skip(offset int) Query` - Alias for Offset
- `Take(limit int) Query` - Alias for Limit

#### Aggregations

- `Count() (int64, error)` - Count matching records
- `Sum(column string) (float64, error)` - Sum column values
- `Avg(column string) (float64, error)` - Average column values
- `Min(column string) (float64, error)` - Minimum column value
- `Max(column string) (float64, error)` - Maximum column value
- `Exists() (bool, error)` - Check if records exist

#### Joins

- `Join(table, first, operator, second string) Query` - Add INNER JOIN
- `LeftJoin(table, first, operator, second string) Query` - Add LEFT JOIN
- `RightJoin(table, first, operator, second string) Query` - Add RIGHT JOIN
- `JoinRaw(sql string, bindings ...any) Query` - Add raw JOIN

#### Grouping

- `GroupBy(columns ...string) Query` - Add GROUP BY clause
- `Having(column string, operator string, value any) Query` - Add HAVING clause
- `HavingRaw(sql string, bindings ...any) Query` - Add raw HAVING

#### CRUD Operations

- `Create(value any) error` - Insert new record
- `Update(column string, value any) error` - Update records
- `UpdateMany(updates map[string]any) error` - Update multiple columns
- `Delete() (sql.Result, error)` - Delete records
- `Increment(column string, amount ...int64) error` - Increment column value
- `Decrement(column string, amount ...int64) error` - Decrement column value

#### Advanced

- `With(relations ...string) Query` - Eager load relationships
- `WithTrashed() Query` - Include soft-deleted records
- `OnlyTrashed() Query` - Only soft-deleted records
- `WithoutEvents() Query` - Disable model events
- `Table(table string) Query` - Set table name
- `Distinct() Query` - Select distinct records
- `Paginate(page, perPage int) (Paginator, error)` - Paginate results
- `Chunk(size int, fn func([]any) error) error` - Process in chunks
- `Cursor(fn func(Cursor) error) error` - Process with cursor

#### Utilities

- `ToSql() (string, error)` - Generate SQL without executing
- `Clone() Query` - Clone query builder
- `NewQuery() Query` - Create new query instance

**See also:** [Query Builder Documentation](./query-builder.md)

## Schema Builder

### Schema Methods

- `Create(table string, callback func(Blueprint)) error` - Create a new table
- `Table(table string, callback func(Blueprint)) error` - Modify an existing table
- `Drop(table string) error` - Drop a table
- `DropIfExists(table string) error` - Drop a table if it exists
- `Rename(from, to string) error` - Rename a table
- `HasTable(table string) bool` - Check if table exists
- `HasColumn(table, column string) bool` - Check if column exists
- `GetTableListing() []string` - Get all table names
- `GetColumnListing(table string) []string` - Get column names for a table
- `GetIndexListing(table string) []string` - Get index names for a table

### Blueprint Column Methods

#### Numeric

- `ID()` - Add auto-incrementing ID column
- `Integer(name string) ColumnDefinition` - Add INTEGER column
- `BigInteger(name string) ColumnDefinition` - Add BIGINT column
- `SmallInteger(name string) ColumnDefinition` - Add SMALLINT column
- `Decimal(name string, precision, scale int) ColumnDefinition` - Add DECIMAL column
- `Float(name string, precision, scale int) ColumnDefinition` - Add FLOAT column
- `Double(name string, precision, scale int) ColumnDefinition` - Add DOUBLE column

#### String

- `String(name string, length ...int) ColumnDefinition` - Add VARCHAR column
- `Char(name string, length int) ColumnDefinition` - Add CHAR column
- `Text(name string) ColumnDefinition` - Add TEXT column
- `MediumText(name string) ColumnDefinition` - Add MEDIUMTEXT column
- `LongText(name string) ColumnDefinition` - Add LONGTEXT column

#### Date/Time

- `DateTime(name string) ColumnDefinition` - Add DATETIME column
- `Timestamp(name string) ColumnDefinition` - Add TIMESTAMP column
- `Date(name string) ColumnDefinition` - Add DATE column
- `Time(name string) ColumnDefinition` - Add TIME column
- `Year(name string) ColumnDefinition` - Add YEAR column

#### Boolean

- `Boolean(name string) ColumnDefinition` - Add BOOLEAN column

#### Special

- `JSON(name string) ColumnDefinition` - Add JSON column
- `Binary(name string) ColumnDefinition` - Add BINARY column
- `Enum(name string, values []string) ColumnDefinition` - Add ENUM column
- `UUID(name string) ColumnDefinition` - Add UUID column
- `IPAddress(name string) ColumnDefinition` - Add IP address column
- `MacAddress(name string) ColumnDefinition` - Add MAC address column

### Blueprint Modifiers

- `Nullable() ColumnDefinition` - Allow NULL values
- `Unique() ColumnDefinition` - Add unique constraint
- `Default(value any) ColumnDefinition` - Set default value
- `Index() ColumnDefinition` - Add index
- `Primary() ColumnDefinition` - Set as primary key
- `AutoIncrement() ColumnDefinition` - Set as auto-increment
- `Comment(comment string) ColumnDefinition` - Add comment
- `After(column string) ColumnDefinition` - Place after another column (MySQL)
- `Unsigned() ColumnDefinition` - Set as unsigned (numeric types)

### Blueprint Index Methods

- `Index(columns ...string)` - Add index
- `UniqueIndex(columns ...string)` - Add unique index
- `DropIndex(name string)` - Drop index
- `RenameIndex(from, to string)` - Rename index

### Blueprint Foreign Key Methods

- `ForeignKey(column string) ForeignKeyDefinition` - Add foreign key
- `On(table string) ForeignKeyDefinition` - Set referenced table
- `References(column string) ForeignKeyDefinition` - Set referenced column
- `OnDelete(action string) ForeignKeyDefinition` - Set ON DELETE action
- `OnUpdate(action string) ForeignKeyDefinition` - Set ON UPDATE action

### Blueprint Helper Methods

- `Timestamps()` - Add created_at and updated_at timestamps
- `SoftDeletes()` - Add deleted_at timestamp for soft deletes
- `DropColumn(name string)` - Drop a column
- `RenameColumn(from, to string)` - Rename a column

**See also:** [Schema Builder Documentation](./schema-builder.md)

## ORM

### Orm Methods

- `Query() Query` - Get query builder
- `Observe(observers []ModelToObserver) error` - Register model observers
- `Model(model any) Query` - Get query for specific model

### Model Interface

```go
type Model interface {
    Factory() Factory
}
```

### Observer Interface

```go
type Observer interface {
    Creating(event Event) error
    Created(event Event) error
    Updating(event Event) error
    Updated(event Event) error
    Saving(event Event) error
    Saved(event Event) error
    Deleting(event Event) error
    Deleted(event Event) error
    Restoring(event Event) error
    Restored(event Event) error
}
```

### Soft Deletes

```go
type SoftDeletes struct {
    DeletedAt *time.Time
}
```

**See also:** [Associations Documentation](./associations.md), [Observers Documentation](./observers.md)

## Factory

### Factory Methods

- `Count(count int) Factory` - Set number of models to generate
- `Table(table string) Factory` - Set table name for database operations
- `Create(value any, attributes ...map[string]any) (any, error)` - Create and persist
- `CreateQuietly(value any, attributes ...map[string]any) (any, error)` - Create without events
- `Make(value any, attributes ...map[string]any) (any, error)` - Create in memory

### Factory Interface

```go
type Factory interface {
    Definition() map[string]any
}
```

**See also:** [Factory Documentation](./factory.md)

## Seeder

### Database Methods

- `Seed(seeders []Seeder) error` - Run specified seeders
- `SeedOnce(seeders []Seeder) error` - Run seeders only once
- `Seeder() Facade` - Get seeder facade

### Facade Methods

- `Register(seeders []Seeder)` - Register seeders
- `GetSeeder(name string) Seeder` - Get seeder by signature
- `GetSeeders() []Seeder` - Get all registered seeders
- `Call(seeders []Seeder) error` - Execute specified seeders
- `CallOnce(seeders []Seeder) error` - Execute seeders only once
- `ResetCallOnce()` - Reset call-once tracking

### Seeder Interface

```go
type Seeder interface {
    Signature() string
    Run() error
}
```

### Registry Functions

- `RegisterSeeder(name string, s Seeder)` - Register seeder globally
- `GetSeeder(name string) Seeder` - Retrieve from global registry
- `GetSeeders() []Seeder` - Retrieve all from global registry
- `ClearRegistry()` - Clear global registry

**See also:** [Seeder Documentation](./seeder.md)

## Migration

### Database Methods

- `Migrate(paths ...string) error` - Run all pending migrations
- `MigrateDown(steps int, paths ...string) error` - Rollback migrations
- `MigrateFresh(paths ...string) error` - Drop all and re-run
- `MigrateReset(paths ...string) error` - Rollback all and re-run
- `MigrationStatus(paths ...string) ([]MigrationStatus, error)` - Get status

### Migration Interface

```go
type Migration struct {
    Up   func(Schema) error
    Down func(Schema) error
}
```

### Registry Functions

- `RegisterMigration(name string, migration Migration)` - Register migration

**See also:** [Migrations Documentation](./migrations.md)

## Observer

### Event Types

- `Creating` - Before model is created
- `Created` - After model is created
- `Updating` - Before model is updated
- `Updated` - After model is updated
- `Saving` - Before model is saved (create or update)
- `Saved` - After model is saved (create or update)
- `Deleting` - Before model is deleted
- `Deleted` - After model is deleted
- `Restoring` - Before soft-deleted model is restored
- `Restored` - After soft-deleted model is restored

### Event Structure

```go
type Event struct {
    Model  any
    Query  Query
    Action string
}
```

**See also:** [Observers Documentation](./observers.md)

## Association

### Association Methods

- `Association(relation string) Association` - Get association instance
- `Load(model any, relations ...string) error` - Load relationships
- `With(relations ...string) Query` - Eager load relationships

### Association Operations

- `Append(model any, related ...any) error` - Add related records
- `Detach(model any, related ...any) error` - Remove related records
- `Sync(model any, related []any) error` - Sync related records
- `Count(model any) (int64, error)` - Count related records
- `Get(model any) (any, error)` - Get related records

### Relationship Types

- `BelongsTo` - Many-to-one relationship
- `HasMany` - One-to-many relationship
- `HasOne` - One-to-one relationship

**See also:** [Associations Documentation](./associations.md)

## Configuration Types

### DBConfig

```go
type DBConfig struct {
    Default     string
    Connections map[string]ConnectionConfig
    Migrations  MigrationConfig
}
```

### ConnectionConfig

```go
type ConnectionConfig struct {
    Driver   string
    Host     string
    Port     int
    Database string
    Username string
    Password string
    Options  map[string]string
    Pool     PoolConfig
}
```

### PoolConfig

```go
type PoolConfig struct {
    MaxOpenConns int
    MaxIdleConns int
    MaxLifetime  int
}
```

### MigrationConfig

```go
type MigrationConfig struct {
    Driver string
    Table  string
}
```

## Error Types

### Common Errors

- `ErrConnectionFailed` - Database connection failed
- `ErrQueryFailed` - Query execution failed
- `ErrMigrationFailed` - Migration failed
- `ErrTransactionFailed` - Transaction failed
- `ErrModelNotFound` - Model not found
- `ErrInvalidConfig` - Invalid configuration

**See also:** [errors/errors.go](../errors/errors.go)

## Supported Databases

- MySQL 5.7+
- PostgreSQL 12+
- SQLite 3+
- SQL Server 2017+
- Turso (SQLite edge)

## DSN Formats

- **MySQL**: `mysql://user:password@localhost:3306/database`
- **PostgreSQL**: `postgres://user:password@localhost:5432/database?sslmode=disable`
- **SQLite**: `sqlite:///path/to/database.db`
- **SQL Server**: `sqlserver://user:password@localhost:1433?database=database`
- **Turso**: `turso://database.turso.io?auth-token=token`

## Additional Documentation

- [Query Builder](./query-builder.md) - Detailed query builder guide
- [Schema Builder](./schema-builder.md) - Detailed schema builder guide
- [Migrations](./migrations.md) - Migration system guide
- [Associations](./associations.md) - Relationship management
- [Observers](./observers.md) - Model lifecycle events
- [Factory](./factory.md) - Test data generation
- [Seeder](./seeder.md) - Database seeding
- [Testing](./testing.md) - Testing guide
- [Performance](./performance.md) - Performance optimization
