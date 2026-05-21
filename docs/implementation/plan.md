# Neat Implementation Plan

**Date:** May 21, 2026
**Status**: Detailed Implementation Plan for Greenfield Development

## Overview

This document provides the detailed implementation plan for building **neat**, a Go ORM and database abstraction layer that provides the same functionality as eloquent but without GORM dependency. Neat will be implemented from scratch using native `database/sql` with a custom query builder and SQL generation.

## Project Goals

1. **Feature Parity**: Implement all features currently available in eloquent
2. **No GORM Dependency**: Use native `database/sql` and custom SQL builder
3. **Clean Architecture**: Maintain the same contract-based interface design
4. **Performance**: Optimize for better performance than GORM-based implementation
5. **Maintainability**: Simplify codebase by removing GORM abstraction layer

## Eloquent Feature Analysis

Based on the current eloquent implementation, neat must implement:

### Core ORM Features (contracts/database/orm/orm.go)
- **80+ Query Methods**: Where, Select, Order, Limit, Offset, Join, Group, Having, etc.
- **CRUD Operations**: Create, Read, Update, Delete with various methods
- **Advanced Query**: WhereIn, WhereBetween, WhereColumn, WhereExists, WhereJson, etc.
- **Aggregations**: Count, Sum, Avg, Min, Max, Exists
- **Result Retrieval**: Get, First, Find, Pluck, Value, Scan, Chunk, Paginate, Cursor
- **Transactions**: Begin, Commit, Rollback, Transaction with callbacks and savepoints
- **Model Lifecycle**: Observer pattern with Creating, Created, Updating, Updated, etc.
- **Soft Deletes**: WithTrashed, OnlyTrashed, WithoutTrashed, Restore, ForceDelete
- **Associations**: Belongs-to, Has-many, Has-one with eager and lazy loading
- **Query Logging**: EnableQueryLog, DisableQueryLog, GetQueryLog, FlushQueryLog
- **ToSql Interface**: Generate SQL without execution for debugging

### Schema Builder (contracts/database/schema/)
- **Blueprint Interface**: 50+ methods for table/column definition
- **Column Types**: Integer, BigInteger, String, Text, Boolean, Json, Decimal, etc.
- **Indexes**: Primary, Unique, Index, FullText
- **Foreign Keys**: Foreign constraints with onUpdate/onDelete
- **Table Operations**: Create, Drop, Rename, Modify
- **Column Operations**: Add, Drop, Rename, Modify
- **Database-specific Grammars**: MySQL, PostgreSQL, SQLite, SQL Server
- **Database-specific Processors**: Type conversion and processing

### Configuration System (config.go)
- **DBConfig**: Default connection, connection map, pool config
- **ConnectionConfig**: Driver, host, port, database, username, password, SSL, etc.
- **PoolConfig**: MaxIdleConns, MaxOpenConns, ConnMaxLifetime, ConnMaxIdleTime
- **DSN Parsing**: Support postgres://, mysql://, sqlite://, turso:// formats

### Support Utilities (support/)
- **collect**: Collection utilities for data manipulation
- **convert**: Type conversion utilities
- **database**: Database-specific utilities
- **env**: Environment variable handling
- **str**: String manipulation utilities

### Event System (event.go)
- **EventBus**: Event dispatching and handling
- **Event Types**: Model lifecycle events
- **Observer Registration**: Register observers for models

### Factory Pattern
- **Factory Interface**: Create models with default attributes
- **Count**: Set number of models to generate
- **Create**: Create and persist models
- **Make**: Create models without persistence
- **CreateQuietly**: Create without firing events

## Architecture Design

### Package Structure

```
neat/
├── config.go                 # Configuration structures and parsing
├── db.go                     # Main Database entry point
├── event.go                  # Event bus implementation
├── contracts/                # Interface definitions (copied from eloquent)
│   ├── config/
│   ├── database/
│   │   ├── orm/
│   │   ├── schema/
│   │   ├── factory/
│   │   └── migration/
│   ├── errors/
│   ├── factory/
│   ├── log/
│   ├── migration/
│   ├── schema/
│   └── seeder/
├── database/
│   ├── db/                   # Database connection management
│   │   ├── config_builder.go
│   │   ├── dsn.go
│   │   └── expression.go
│   ├── orm/                  # ORM abstraction layer
│   │   ├── orm.go           # Orm implementation
│   │   ├── model.go         # Model types (Model, SoftDeletes, Timestamps)
│   │   └── factory.go       # Factory implementation
│   ├── query/                # Query builder implementation
│   │   ├── query.go         # Main Query struct
│   │   ├── builder.go       # SQL builder
│   │   ├── clause.go        # Query clauses
│   │   ├── where.go         # Where clause handling
│   │   ├── join.go          # Join handling
│   │   ├── aggregate.go     # Aggregation methods
│   │   └── to_sql.go        # ToSql interface implementation
│   ├── association/          # Association handling
│   │   ├── association.go   # Association interface
│   │   ├── belongs_to.go    # Belongs-to relationship
│   │   ├── has_many.go      # Has-many relationship
│   │   └── has_one.go       # Has-one relationship
│   ├── observer/             # Observer pattern
│   │   ├── observer.go      # Observer implementation
│   │   ├── event.go         # Event implementation
│   │   └── dispatcher.go    # Event dispatcher
│   ├── soft_delete/          # Soft delete implementation
│   │   └── soft_delete.go
│   ├── transaction/          # Transaction handling
│   │   ├── transaction.go   # Transaction implementation
│   │   ├── savepoint.go     # Savepoint handling
│   │   └── callback.go      # Transaction callbacks
│   ├── cursor/               # Cursor for streaming
│   │   └── cursor.go
│   ├── schema/               # Schema builder (reuse from eloquent)
│   │   ├── blueprint.go
│   │   ├── schema.go
│   │   ├── column.go
│   │   ├── index.go
│   │   ├── constants/
│   │   ├── grammars/
│   │   └── processors/
│   └── driver/               # Native database drivers
│       ├── driver.go        # Driver interface
│       ├── mysql.go         # MySQL driver
│       ├── postgres.go      # PostgreSQL driver
│       ├── sqlite.go        # SQLite driver
│       ├── sqlserver.go     # SQL Server driver
│       └── turso.go         # Turso driver
├── support/                  # Utility functions (copied from eloquent)
│   ├── collect/
│   ├── convert/
│   ├── database/
│   ├── env/
│   └── str/
├── errors/                   # Error definitions
│   └── errors.go
└── log/                      # Logger interface
    ├── log.go
    ├── std_logger.go
    └── noop_logger.go
```

### Key Design Decisions

#### 1. Native database/sql Usage
- Use `database/sql` directly for all database operations
- Implement custom connection pooling logic
- Use prepared statements for all queries
- Implement dialect-specific placeholder handling

#### 2. Custom SQL Builder
- Build SQL strings programmatically
- Support dialect-specific syntax
- Use parameterized queries to prevent SQL injection
- Implement clause-based query building (Where, Join, Group, etc.)

#### 3. Tag-based Model Mapping
- Use custom struct tags (e.g., `neat:"column:name;primaryKey"`)
- Support GORM tags for migration compatibility
- Implement tag parser for both neat and GORM tags
- Auto-detect table name from struct name

#### 4. Observer Pattern
- Implement observer pattern for model lifecycle events
- Support all event types: Creating, Created, Updating, Updated, etc.
- Allow multiple observers per model
- Execute events in correct order

#### 5. Association Loading
- Implement eager loading with `With()`
- Implement lazy loading with `Load()`
- Support belongs-to, has-many, has-one relationships
- Use separate queries for association loading (N+1 problem avoidance)

#### 6. Transaction Management
- Use `sql.Tx` for transactions
- Implement transaction callbacks
- Support nested transactions with savepoints
- Handle transaction rollback on errors

#### 7. Soft Deletes
- Use `deleted_at` timestamp column
- Exclude soft-deleted records by default
- Implement WithTrashed, OnlyTrashed, WithoutTrashed
- Support restore and force delete

## Implementation Phases

### Phase 1: Project Setup and Foundation (Week 1-2) ✅ COMPLETED

#### Objective
Set up the project structure, copy contracts, implement configuration system, and create database driver foundation.

#### Tasks

**1.1 Project Initialization**
- Initialize Go module: `go mod init github.com/dracory/neat`
- Set up project structure as defined in Architecture Design
- Copy contracts directory from eloquent (no changes needed)
- Copy support directory from eloquent (no changes needed)
- Copy errors directory from eloquent (no changes needed)
- Copy log directory from eloquent (no changes needed)
- Set up go.mod with initial dependencies:
  - github.com/dromara/carbon/v2 (for timestamps)
  - Database drivers: mysql, postgres, sqlite, sqlserver

**1.2 Configuration System**
- Implement `config.go` with DBConfig, ConnectionConfig, PoolConfig structures
- Implement DSN parsing in `config.go` (parseDSN function)
- Support postgres://, mysql://, sqlite://, turso:// formats
- Implement connection pool configuration

**1.3 Database Driver Interface**
- Create `database/driver/driver.go` with Driver interface
- Define driver methods: Open, Close, Ping, BeginTx
- Implement dialect-specific placeholder handling
- Create placeholder.go for dialect-specific placeholders:
  - MySQL: `?`
  - PostgreSQL: `$1, $2, $3`
  - SQLite: `?`
  - SQL Server: `@p1, @p2, @p3`

**1.4 Native Database Drivers**
- Implement `database/driver/mysql.go`:
  - Use github.com/go-sql-driver/mysql
  - Implement MySQL-specific connection logic
  - Handle MySQL-specific connection options
- Implement `database/driver/postgres.go`:
  - Use github.com/lib/pq
  - Implement PostgreSQL-specific connection logic
  - Handle SSL modes and schema options
- Implement `database/driver/sqlite.go`:
  - Use github.com/mattn/go-sqlite3
  - Implement SQLite-specific connection logic
  - Handle in-memory and file-based databases
- Implement `database/driver/sqlserver.go`:
  - Use github.com/microsoft/go-odbc/sqlserver
  - Implement SQL Server-specific connection logic
- Implement `database/driver/turso.go`:
  - Use github.com/tursodatabase/libsql-client-go
  - Implement Turso-specific connection logic

**1.5 Database Connection Management**
- Implement `database/db/config_builder.go`:
  - Build DSN from ConnectionConfig
  - Handle driver-specific DSN formats
  - Apply connection pool settings
- Implement `database/db/dsn.go`:
  - DSN parsing utilities
  - DSN building utilities
- Implement `database/db/expression.go`:
  - Database expression handling
  - Query expression building

**Files to create:**
- `config.go`
- `database/driver/driver.go`
- `database/driver/placeholder.go`
- `database/driver/mysql.go`
- `database/driver/postgres.go`
- `database/driver/sqlite.go`
- `database/driver/sqlserver.go`
- `database/driver/turso.go`
- `database/db/config_builder.go`
- `database/db/dsn.go`
- `database/db/expression.go`

**Files to copy from eloquent:**
- `contracts/` (entire directory)
- `support/` (entire directory)
- `errors/` (entire directory)
- `log/` (entire directory)

**Success criteria:**
- Project structure is set up correctly
- All database drivers can connect to their respective databases
- DSN parsing works for all supported formats
- Connection pooling is configured correctly

---

### Phase 2: ORM Abstraction Layer (Week 3-4) ✅ COMPLETED

#### Objective
Implement the ORM abstraction layer with connection management and query initialization.

#### Tasks

**2.1 ORM Implementation**
- Implement `database/orm/orm.go`:
  - Orm struct with connection management
  - Connection(name string) Orm method
  - DB() (*sql.DB, error) method
  - Query() Query method
  - DatabaseName() string method
  - Name() string method
  - Transaction(txFunc func(tx Query) error) error method
  - WithContext(ctx context.Context) Orm method
  - Query log methods: EnableQueryLog, DisableQueryLog, FlushQueryLog, GetQueryLog
  - Factory() Factory method
  - Observe(model any, observer Observer) method
  - Refresh() method

**2.2 Model Types**
- Implement `database/orm/model.go`:
  - Model struct with ID and Timestamps
  - SoftDeletes struct with DeletedAt *time.Time
  - Timestamps struct with CreatedAt and UpdatedAt
  - Use custom tags: `neat:"column:name;primaryKey"`
  - Support GORM tags for compatibility
  - Implement tag parser for both neat and GORM tags

**2.3 Factory Implementation**
- Implement `database/orm/factory.go`:
  - Factory struct
  - Count(count int) Factory method
  - Create(value any, attributes ...map[string]any) error method
  - CreateQuietly(value any, attributes ...map[string]any) error method
  - Make(value any, attributes ...map[string]any) error method

**2.4 Event System**
- Implement `event.go`:
  - EventBus struct
  - Subscribe(event string, handler func(any)) method
  - Publish(event string, data any) method
  - Unsubscribe(event string) method

**2.5 Database Entry Point**
- Implement `db.go`:
  - Database struct
  - New(cfg DBConfig, opts ...Option) function
  - NewFromDSN(dsn string, opts ...Option) function
  - Query() Query method
  - Schema() Schema method
  - Close() error method
  - Connection(name string) method
  - Transaction(txFunc func(tx Query) error) error method
  - Query log methods
  - DatabaseName() string method

**Files to create:**
- `database/orm/orm.go`
- `database/orm/model.go`
- `database/orm/factory.go`
- `event.go`
- `db.go`

**Success criteria:**
- ORM can manage multiple connections
- Query() returns a valid Query instance
- Factory can create models
- Event bus can dispatch events
- Database entry point is functional

---

### Phase 3: Core Query Builder (Week 5-8) ✅ COMPLETED

#### Objective
Implement the core query builder with CRUD operations and basic query building methods.

#### Tasks

**3.1 Query Struct and Basic Methods**
- Implement `database/query/query.go`:
  - Query struct with all query state (where, select, join, etc.)
  - Connection(name string) Query method
  - Model(value any) Query method
  - Table(name string, args ...any) Query method
  - DB() (*sql.DB, error) method
  - Driver() database.Driver method
  - InTransaction() bool method

**3.2 CRUD Operations**
- Implement basic CRUD methods in query.go:
  - Find(dest any, conds ...any) error
  - First(dest any) error
  - FirstOrFail(dest any) error
  - Get(dest any) error
  - Create(value any) error
  - Save(value any) error
  - SaveQuietly(value any) error
  - Update(column any, value ...any) (*Result, error)
  - Delete(value ...any) (*Result, error)
  - InsertGetId(values any) (uint, error)

**3.3 Query Building Methods**
- Implement query building methods:
  - Where(query any, args ...any) Query
  - OrWhere(query any, args ...any) Query
  - Select(query any, args ...any) Query
  - Order(value any) Query
  - OrderBy(column string, direction ...string) Query
  - OrderByDesc(column string) Query
  - Limit(limit int) Query
  - Offset(offset int) Query
  - Distinct(args ...any) Query

**3.4 Advanced Where Clauses**
- Implement `database/query/where.go`:
  - WhereIn(column string, values []any) Query
  - WhereNotIn(column string, values []any) Query
  - OrWhereIn(column string, values []any) Query
  - OrWhereNotIn(column string, values []any) Query
  - WhereBetween(column string, x, y any) Query
  - WhereNotBetween(column string, x, y any) Query
  - OrWhereBetween(column string, x, y any) Query
  - OrWhereNotBetween(column string, x, y any) Query
  - WhereNull(column string) Query
  - WhereNotNull(column string) Query
  - OrWhereNull(column string) Query
  - WhereColumn(first, operator, second string) Query
  - OrWhereColumn(first, operator, second string) Query
  - WhereExists(callback func(Query) Query) Query
  - WhereNot(query any, args ...any) Query
  - OrWhereNot(query any, args ...any) Query
  - WhereAny(columns []string, operator string, value any) Query
  - WhereAll(columns []string, operator string, value any) Query
  - WhereNone(columns []string, operator string, value any) Query

**3.5 JSON Query Methods**
- Implement JSON methods in where.go:
  - WhereJsonContains(column string, value any) Query
  - OrWhereJsonContains(column string, value any) Query
  - WhereJsonDoesntContain(column string, value any) Query
  - OrWhereJsonDoesntContain(column string, value any) Query
  - WhereJsonContainsKey(column string) Query
  - OrWhereJsonContainsKey(column string) Query
  - WhereJsonDoesntContainKey(column string) Query
  - OrWhereJsonDoesntContainKey(column string) Query
  - WhereJsonLength(column string, operator string, value any) Query

**3.6 SQL Builder**
- Implement `database/query/builder.go`:
  - Build SELECT query
  - Build INSERT query
  - Build UPDATE query
  - Build DELETE query
  - Handle dialect-specific syntax
  - Use parameterized queries
  - Implement placeholder replacement

**3.7 Clause Handling**
- Implement `database/query/clause.go`:
  - Clause types: Where, Select, Join, Group, Having, Order, Limit, Offset
  - Clause builder interface
  - Clause composition
  - Clause to SQL conversion

**Files to create:**
- `database/query/query.go`
- `database/query/where.go`
- `database/query/builder.go`
- `database/query/clause.go`

**Success criteria:**
- Basic CRUD operations work
- Query building generates correct SQL
- Where clauses work correctly
- JSON queries work correctly
- Parameterized queries prevent SQL injection

---

### Phase 4: Advanced Query Features (Week 9-11)

#### Objective
Implement advanced query features including joins, aggregations, and special query methods.

#### Tasks

**4.1 Join Operations**
- Implement `database/query/join.go`:
  - Join(query string, args ...any) Query
  - LeftJoin(query string, args ...any) Query
  - RightJoin(query string, args ...any) Query
  - CrossJoin(query string, args ...any) Query
  - Handle different join syntax for dialects
  - Support subqueries in joins

**4.2 Group By and Having**
- Implement in query.go:
  - Group(name string) Query
  - Having(query any, args ...any) Query
  - Support multiple group by clauses
  - Support complex having conditions

**4.3 Aggregation Methods**
- Implement `database/query/aggregate.go`:
  - Count(count *int64) error
  - Sum(column string, dest any) error
  - Avg(column string, dest any) error
  - Min(column string, dest any) error
  - Max(column string, dest any) error
  - Exists(exists *bool) error

**4.4 Result Retrieval Methods**
- Implement in query.go:
  - Pluck(column string, dest any) error
  - Value(column string, dest any) error
  - Scan(dest any) error
  - Chunk(size int, callback any) error
  - Paginate(page, limit int, dest any, total *int64) error

**4.5 Special Query Methods**
- Implement in query.go:
  - FirstOr(dest any, callback func() error) error
  - FirstOrCreate(dest any, conds ...any) error
  - FirstOrNew(dest any, attributes any, values ...any) error
  - UpdateOrCreate(dest any, attributes any, values any) error
  - UpdateOrInsert(attributes any, values any) error
  - Increment(column string, amount ...any) (*Result, error)
  - Decrement(column string, amount ...any) (*Result, error)
  - InRandomOrder() Query
  - LockForUpdate() Query
  - SharedLock() Query
  - Raw(sql string, values ...any) Query
  - Exec(sql string, values ...any) (*Result, error)

**4.6 Cursor Implementation**
- Implement `database/cursor/cursor.go`:
  - Cursor interface with Scan method
  - Cursor() (chan Cursor, error) method in query.go
  - Handle streaming results
  - Handle context cancellation

**Files to create:**
- `database/query/join.go`
- `database/query/aggregate.go`
- `database/cursor/cursor.go`

**Success criteria:**
- All join operations work correctly
- Group by and having work correctly
- Aggregation methods return correct results
- Special query methods work
- Cursor streaming works correctly

---

### Phase 5: Transaction Handling (Week 12)

#### Objective
Implement transaction management with callbacks and savepoints.

#### Tasks

**5.1 Basic Transaction Methods**
- Implement `database/transaction/transaction.go`:
  - Begin(opts ...*sql.TxOptions) (Query, error) method
  - Commit() error method
  - Rollback() error method
  - Transaction(txFunc func(tx Query) error, opts ...*sql.TxOptions) error method
  - InTransaction() bool method
  - Use sql.Tx for transaction management

**5.2 Transaction Callbacks**
- Implement `database/transaction/callback.go`:
  - BeforeCommit(callback func() error) method
  - AfterCommit(callback func() error) method
  - BeforeRollback(callback func() error) method
  - AfterRollback(callback func() error) method
  - Callback execution in correct order
  - Error aggregation with MultiCallbackError

**5.3 Savepoint Handling**
- Implement `database/transaction/savepoint.go`:
  - SavePoint(name string) error method
  - RollbackTo(name string) error method
  - Handle savepoint syntax for different dialects
  - Support nested transactions

**Files to create:**
- `database/transaction/transaction.go`
- `database/transaction/callback.go`
- `database/transaction/savepoint.go`

**Success criteria:**
- Basic transactions work correctly
- Callbacks execute in correct order
- Nested transactions with savepoints work
- Error handling in transactions is correct

---

### Phase 6: Observer Pattern (Week 13)

#### Objective
Implement observer pattern for model lifecycle events.

#### Tasks

**6.1 Observer Implementation**
- Implement `database/observer/observer.go`:
  - Observer struct
  - Register observers for models
  - Execute observers in correct order
  - Handle observer errors

**6.2 Event Implementation**
- Implement `database/observer/event.go`:
  - Event struct implementing contracts/database/orm/events.Event
  - Context() context.Context method
  - GetAttribute(key string) any method
  - GetOriginal(key string, def ...any) any method
  - IsClean(columns ...string) bool method
  - IsDirty(columns ...string) bool method
  - Query() Query method
  - SetAttribute(key string, value any) method

**6.3 Event Dispatcher**
- Implement `database/observer/dispatcher.go`:
  - Dispatcher struct
  - Dispatch events for all event types:
    - Creating, Created
    - Updating, Updated
    - Saving, Saved
    - Deleting, Deleted
    - ForceDeleting, ForceDeleted
    - Restoring, Restored
    - Retrieved
  - Handle observer errors
  - Support multiple observers per model

**6.4 Integration with Query**
- Add observer execution to Create, Save, Update, Delete methods
- Add WithoutEvents() method to query.go
- Ensure events fire in correct order

**Files to create:**
- `database/observer/observer.go`
- `database/observer/event.go`
- `database/observer/dispatcher.go`

**Success criteria:**
- Observer pattern works correctly
- All event types fire in correct order
- Multiple observers can be registered per model
- Observer errors are handled correctly
- WithoutEvents() disables event firing

---

### Phase 7: Soft Deletes (Week 14)

#### Objective
Implement soft delete functionality.

#### Tasks

**7.1 Soft Delete Implementation**
- Implement `database/soft_delete/soft_delete.go`:
  - Soft delete logic with deleted_at column
  - Set deleted_at timestamp instead of hard delete
  - Exclude soft-deleted records from queries by default

**7.2 Soft Delete Query Methods**
- Implement in query.go:
  - WithTrashed() Query - include soft-deleted records
  - OnlyTrashed() Query - only soft-deleted records
  - WithoutTrashed() Query - exclude soft-deleted records (default)
  - Restore(model ...any) (*Result, error) - restore soft-deleted records
  - ForceDelete(value ...any) (*Result, error) - permanent delete

**7.3 Integration with Query Builder**
- Add deleted_at clause to queries by default
- Handle WithTrashed, OnlyTrashed, WithoutTrashed
- Update Delete method to use soft delete
- Update Restore and ForceDelete methods

**Files to create:**
- `database/soft_delete/soft_delete.go`

**Success criteria:**
- Soft deletes work correctly
- WithTrashed, OnlyTrashed, WithoutTrashed work
- Restore and ForceDelete work correctly
- Soft-deleted records are excluded by default

---

### Phase 8: Associations (Week 15-17)

#### Objective
Implement association handling for belongs-to, has-many, and has-one relationships.

#### Tasks

**8.1 Association Interface**
- Implement `database/association/association.go`:
  - Association struct
  - Implement contracts/database/orm/orm.Association interface:
    - Find(out any, conds ...any) error
    - Append(values ...any) error
    - Replace(values ...any) error
    - Delete(values ...any) error
    - Clear() error
    - Count() int64

**8.2 Belongs-to Relationship**
- Implement `database/association/belongs_to.go`:
  - Belongs-to relationship logic
  - Load related model
  - Set foreign key
  - Handle eager loading

**8.3 Has-many Relationship**
- Implement `database/association/has_many.go`:
  - Has-many relationship logic
  - Load related collection
  - Append, Replace, Delete, Clear operations
  - Handle eager loading

**8.4 Has-one Relationship**
- Implement `database/association/has_one.go`:
  - Has-one relationship logic
  - Load related model
  - Handle eager loading

**8.5 Eager Loading**
- Implement With(query string, args ...any) Query in query.go
- Implement Load(dest any, relation string, args ...any) error in query.go
- Implement LoadMissing(dest any, relation string, args ...any) error in query.go
- Implement Without(relations ...string) Query in query.go
- Implement WithCount(query string, args ...any) Query in query.go
- Implement WithExists(query string, args ...any) Query in query.go
- Use separate queries to avoid N+1 problem

**8.6 Association Method**
- Implement Association(association string) Association in query.go
- Return appropriate association instance

**Files to create:**
- `database/association/association.go`
- `database/association/belongs_to.go`
- `database/association/has_many.go`
- `database/association/has_one.go`

**Success criteria:**
- All association types work correctly
- Eager loading with With() works
- Lazy loading with Load() works
- LoadMissing works correctly
- Association operations (Append, Replace, Delete, Clear) work

---

### Phase 9: ToSql Interface (Week 18)

#### Objective
Implement ToSql interface for generating SQL without execution.

#### Tasks

**9.1 ToSql Implementation**
- Implement `database/query/to_sql.go`:
  - ToSql struct implementing contracts/database/orm/orm.ToSql
  - Count() string method
  - Create(value any) string method
  - InsertGetId(values any) string method
  - Delete(value ...any) string method
  - Find(dest any, conds ...any) string method
  - First(dest any) string method
  - ForceDelete(value ...any) string method
  - Get(dest any) string method
  - Pluck(column string, dest any) string method
  - Value(column string, dest any) string method
  - Save(value any) string method
  - Avg(column string, dest any) string method
  - Max(column string, dest any) string method
  - Min(column string, dest any) string method
  - Sum(column string, dest any) string method
  - Update(column any, value ...any) string method
  - Increment(column string, amount ...any) string method
  - Decrement(column string, amount ...any) string method

**9.2 ToRawSql Method**
- Implement ToRawSql() ToSql method in query.go
- Return SQL with placeholders (not parameterized)

**9.3 Integration with Query**
- Add ToSql() ToSql method to query.go
- Ensure SQL generation is consistent with actual execution

**Files to create:**
- `database/query/to_sql.go`

**Success criteria:**
- ToSql interface is fully implemented
- Generated SQL matches actual execution SQL
- ToRawSql returns raw SQL with placeholders

---

### Phase 10: Schema Builder Integration (Week 19)

#### Objective
Integrate the schema builder from eloquent with the new database implementation.

#### Tasks

**10.1 Copy Schema Builder**
- Copy entire `database/schema/` directory from eloquent
- No changes needed to schema builder code
- Schema builder is independent of GORM

**10.2 Schema Integration**
- Implement `database/schema/schema.go`:
  - Schema struct
  - Connection(name string) Schema method
  - Create(table string, callback func(Blueprint)) error method
  - Drop(table string) error method
  - DropIfExists(table string) error method
  - Table(table string, callback func(Blueprint)) error method
  - Rename(from, to string) error method
  - GetColumnListing(table string) []string method
  - GetIndexListing(table string) []string method
  - GetTableListing() []string method
  - HasColumn(table, column string) bool method
  - HasColumns(table string, columns []string) bool method
  - HasIndex(table, index string) bool method
  - HasTable(name string) bool method
  - Sql(sql string) error method
  - Register migrations
  - Orm() orm.Orm method

**10.3 Database Entry Point Integration**
- Add Schema() Schema method to db.go
- Initialize Schema in New() function
- Pass ORM instance to Schema

**Files to create:**
- `database/schema/schema.go` (updated integration)

**Files to copy:**
- `database/schema/` (entire directory from eloquent)

**Success criteria:**
- Schema builder works with native implementation
- All schema operations work correctly
- Schema is accessible from Database entry point

---

### Phase 11: Testing Setup (Week 20)

#### Objective
Set up comprehensive testing infrastructure.

#### Tasks

**11.1 Unit Tests**
- Create unit tests for each component:
  - database/driver/* tests
  - database/db/* tests
  - database/orm/* tests
  - database/query/* tests
  - database/transaction/* tests
  - database/observer/* tests
  - database/soft_delete/* tests
  - database/association/* tests

**11.2 Integration Tests**
- Set up integration test infrastructure:
  - Test database setup (MySQL, PostgreSQL, SQLite, SQL Server)
  - Test fixtures
  - Test models
- Create integration tests for:
  - CRUD operations
  - Query building
  - Transactions
  - Observers
  - Soft deletes
  - Associations
  - Schema builder

**11.3 CI/CD Setup**
- Set up GitHub Actions workflow
- Test on all supported databases
- Run tests on PR and main branch
- Code coverage reporting

**Files to create:**
- `database/driver/*_test.go`
- `database/db/*_test.go`
- `database/orm/*_test.go`
- `database/query/*_test.go`
- `database/transaction/*_test.go`
- `database/observer/*_test.go`
- `database/soft_delete/*_test.go`
- `database/association/*_test.go`
- `integration_tests/` directory structure
- `.github/workflows/integration-tests.yml`

**Success criteria:**
- Unit tests cover all components
- Integration tests work on all databases
- CI/CD runs automatically
- Code coverage is high (>80%)

---

### Phase 12: Documentation and Examples (Week 21)

#### Objective
Create comprehensive documentation and examples.

#### Tasks

**12.1 README**
- Create comprehensive README.md:
  - Project overview
  - Installation instructions
  - Quick start guide
  - Configuration examples
  - ORM usage examples
  - Schema builder examples
  - Supported databases
  - API documentation

**12.2 Examples**
- Create examples directory with:
  - basic-orm example
  - advanced-queries example
  - models example
  - schema-builder example
  - migrations example
  - configuration example
  - transactions example
  - observers example
  - associations example
  - soft-deletes example

**12.3 Documentation**
- Create docs directory with:
  - driver-registration.md
  - query-builder.md
  - schema-builder.md
  - observers.md
  - associations.md
  - soft-deletes.md
  - transactions.md
  - migrations.md

**Files to create:**
- `README.md`
- `examples/*/main.go`
- `docs/*.md`

**Success criteria:**
- README is comprehensive and clear
- Examples are runnable and cover all features
- Documentation is complete and accurate

---

### Phase 13: Performance Optimization (Week 22)

#### Objective
Optimize performance and benchmark against eloquent.

#### Tasks

**13.1 Benchmarking**
- Create benchmarks for:
  - CRUD operations
  - Query building
  - Aggregation
  - Transactions
  - Associations
- Benchmark against eloquent implementation

**13.2 Optimization**
- Optimize SQL generation
- Optimize query building
- Optimize connection pooling
- Optimize memory usage
- Optimize hot paths

**13.3 Profiling**
- Profile with pprof
- Identify bottlenecks
- Optimize identified bottlenecks

**Files to create:**
- `database/query/*_bench_test.go`
- Benchmark results documentation

**Success criteria:**
- Performance is better or equal to eloquent
- No memory leaks
- Hot paths are optimized

---

### Phase 14: Final Polish and Release (Week 23)

#### Objective
Final polish, cleanup, and release preparation.

#### Tasks

**14.1 Code Cleanup**
- Remove unused code
- Remove commented-out code
- Format code with gofmt
- Run linters and fix issues
- Ensure consistent code style

**14.2 Dependency Management**
- Run go mod tidy
- Update dependencies to latest versions
- Remove unused dependencies
- Ensure go.sum is up to date

**14.3 Final Testing**
- Run full test suite
- Test on all supported databases
- Manual testing of all features
- Edge case testing

**14.4 Release Preparation**
- Update version in go.mod
- Create CHANGELOG.md
- Create release notes
- Tag release
- Publish to GitHub

**Files to create:**
- `CHANGELOG.md`

**Success criteria:**
- Code is clean and consistent
- All tests pass
- Documentation is complete
- Ready for release

## Timeline Summary

- **Phase 1**: Weeks 1-2 (Project Setup and Foundation)
- **Phase 2**: Weeks 3-4 (ORM Abstraction Layer)
- **Phase 3**: Weeks 5-8 (Core Query Builder)
- **Phase 4**: Weeks 9-11 (Advanced Query Features)
- **Phase 5**: Week 12 (Transaction Handling)
- **Phase 6**: Week 13 (Observer Pattern)
- **Phase 7**: Week 14 (Soft Deletes)
- **Phase 8**: Weeks 15-17 (Associations)
- **Phase 9**: Week 18 (ToSql Interface)
- **Phase 10**: Week 19 (Schema Builder Integration)
- **Phase 11**: Week 20 (Testing Setup)
- **Phase 12**: Week 21 (Documentation and Examples)
- **Phase 13**: Week 22 (Performance Optimization)
- **Phase 14**: Week 23 (Final Polish and Release)

**Total estimated timeline**: 23 weeks (5.75 months)

## Resource Requirements

### Team
- **Team size**: 1-2 developers
- **Skills required**:
  - Go development expertise
  - database/sql package knowledge
  - SQL dialect knowledge (MySQL, PostgreSQL, SQLite, SQL Server)
  - ORM design patterns
  - Testing and benchmarking

### Infrastructure
- **Testing**: Access to MySQL, PostgreSQL, SQLite, SQL Server for integration testing
- **CI/CD**: GitHub Actions for automated testing
- **Development**: Local development environments for all database types

### Tools
- Go toolchain (go test, go build, go mod)
- Database clients for all supported databases
- Benchmarking tools (go test -bench)
- Code coverage tools (go test -cover)
- Profiling tools (pprof)

## Success Criteria

1. **Feature Parity**: All features from eloquent are implemented
2. **No GORM Dependency**: Zero GORM dependencies in go.mod
3. **API Compatibility**: Same contract interface as eloquent
4. **Performance**: Performance improvement or parity with eloquent
5. **Test Coverage**: >80% code coverage
6. **Documentation**: Comprehensive documentation and examples
7. **Database Support**: All supported databases work correctly (SQLite, MySQL, PostgreSQL, SQL Server, Turso)

## Open Questions

1. **Tag Format**: Should we use `neat:` tags or a different format?
   - Option A: `neat:"column:name;primaryKey"`
   - Option B: `db:"column:name;primaryKey"`
   - Option C: `orm:"column:name;primaryKey"`
   - Recommendation: Option A for clear branding

2. **GORM Tag Support**: Should we support GORM tags for migration?
   - Option A: Yes, support GORM tags indefinitely
   - Option B: Yes, support with deprecation warning
   - Option C: No, only support neat tags
   - Recommendation: Option B for easier migration

3. **External Dependencies**: Should we use any external SQL builders?
   - Option A: Use sb as dependency
   - Option B: Use squirrel as dependency
   - Option C: Implement our own SQL builder
   - Recommendation: Option C for full control

4. **Version Strategy**: Should we version separately from eloquent?
   - Option A: Same version as eloquent
   - Option B: Independent versioning
   - Recommendation: Option B for independent development

## Next Steps

1. Review and approve this implementation plan
2. Set up development environment
3. Begin Phase 1: Project Setup and Foundation
4. Establish weekly progress reviews
5. Set up continuous integration testing

---

**Created by:** Cascade AI Assistant
**Date:** May 21, 2026
**Status**: Awaiting review and approval
