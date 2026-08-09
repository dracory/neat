# Neat ORM

[![Tests Status](https://github.com/dracory/neat/actions/workflows/tests.yml/badge.svg?branch=main)](https://github.com/dracory/neat/actions/workflows/tests.yml)
[![golangci-lint](https://golangci-lint.run/badge.svg)](https://golangci-lint.run)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/dracory/neat)](https://pkg.go.dev/github.com/dracory/neat)
[![codecov](https://codecov.io/gh/dracory/neat/branch/main/graph/badge.svg)](https://codecov.io/gh/dracory/neat)

A powerful and elegant ORM (Object-Relational Mapping) library for Go, designed to provide a clean and intuitive interface for database operations. Neat aims to feature parity with Laravel's Eloquent ORM while being built from scratch without GORM dependencies.

## Features

- **Query Builder**: Fluent and intuitive query building interface
- **ORM**: Full ORM support with models and relationships
- **Schema Builder**: Database schema creation and modification
- **Migrations**: Complete database migration system with the `Migrator` package, schema builder, rollback support, and automatic tracking (major advantage over most Go ORMs)
- **Seeders**: Database seeding for test and initial data
- **Factories**: Test data generation with factory pattern
- **Multiple Database Support**: MySQL, PostgreSQL, SQLite, SQL Server, Turso, Oracle, CSVDB
- **Transactions**: Robust transaction support
- **Observers**: Model lifecycle event system
- **Soft Deletes**: Soft delete functionality with multiple strategies (NULL-based and max-date sentinel)
- **Associations**: BelongsTo, HasMany, HasOne, PolymorphicBelongsTo, PolymorphicHasMany relationships with eager and lazy loading
- **Views**: Create, drop, and introspect database views via `CreateView`, `CreateViewRaw`, `DropView`, `DropViewIfExists`, `HasView` across all supported drivers
- **Array-Backed Sources**: Query in-memory slices of structs or `[]map[string]any` as if they were database tables using `NewArraySourceFrom` — zero boilerplate, no custom `ArraySource` struct required
- **CSVDB Driver**: Query a directory of CSV files as if they were database tables — each `.csv` file becomes a table, with automatic type inference, BOM stripping, and transaction-wrapped bulk loading
- **Connection Pooling**: Efficient connection management
- **Context Support**: Full context.Context support throughout
- **Query Method Aliases**: Sequelize-style (FindAll, FindOne, Destroy) and Django-style (Filter, Exclude, All)
- **Sugar Methods**: Convenience methods (`CountAsVar`, `FirstAsVar`, etc.) that return values directly for improved usability
- **ToSql Interface**: SQL generation without execution
- **Dotted Column References**: `table.column` syntax supported in `OrderBy`, `OrderByDesc`, `Group`, `Distinct`, and `WhereColumn`
- **Security Hardening**: SQL injection prevention with identifier validation

## Key Advantage: Complete Migration System

> **🚀 Most Go ORMs lack comprehensive schema migration support.** Neat ORM includes a complete migration system with the `Migrator` package, schema builder, rollback support, and automatic tracking - something most competitors either lack entirely or require third-party tools for.

## Installation

```bash
go get github.com/dracory/neat
```

## Documentation

- **[HTML Documentation](https://html-preview.github.io/?url=https://github.com/dracory/neat/blob/main/docs/index.html)** - Browse documentation in your browser
- **[Examples](./examples)** - Practical examples for various features

## Quick Start

```go
package main

import (
    "context"
    "log"
    
    "github.com/dracory/neat"
)

type User struct {
    ID    uint
    Name  string
    Email string
}

func main() {
    // Create database connection
    db, err := neat.NewFromDSN("mysql://user:password@localhost:3306/mydb")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Query users
    var users []User
    err = db.Query().Where("name", "John").Get(&users)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Found %d users", len(users))
}
```

## Configuration

### Using DSN String

```go
db, err := neat.NewFromDSN("mysql://user:password@localhost:3306/mydb")
```

### Using DBConfig

```go
config := neat.DBConfig{
    Default: "default",
    Connections: map[string]neat.ConnectionConfig{
        "default": {
            Driver:   "mysql",
            Host:     "localhost",
            Port:     3306,
            Database: "mydb",
            Username: "user",
            Password: "password",
        },
    },
}

db, err := neat.New(config)
```

### Supported DSN Formats

- **MySQL**: `mysql://user:password@localhost:3306/database`
- **PostgreSQL**: `postgres://user:password@localhost:5432/database?sslmode=disable`
- **SQLite**: `sqlite:///path/to/database.db`
- **SQL Server**: `sqlserver://user:password@localhost:1433?database=database`

## ORM Usage

### Models

Neat maps Go structs to database tables using struct tags. For detailed information on table names, column names, and tag priority, see [Models Documentation](./docs/models.html).

### Creating Records

```go
user := User{
    Name:  "John Doe",
    Email: "john@example.com",
}
err := db.Query().Create(&user)
```

### Querying Records

```go
var user User
err := db.Query().Where("id", 1).First(&user)

var users []User
err := db.Query().Where("name", "John").Get(&users)
```

### Updating Records

```go
err := db.Query().Where("id", 1).Update("name", "Jane")
```

### Deleting Records

```go
result, err := db.Query().Where("id", 1).Delete()
```

### Transactions

```go
err := db.Transaction(func(tx neat.Query) error {
    err := tx.Create(&user1)
    if err != nil {
        return err
    }
    
    err = tx.Create(&user2)
    if err != nil {
        return err
    }
    
    return nil
})
```

## Schema Builder

```go
err := db.Schema().Create("users", func(table neat.Blueprint) {
    table.ID()
    table.String("name")
    table.String("email").Unique()
    table.Timestamps()
})
```

### Views

```go
// Create a view from a query builder
err := db.Schema().CreateView("active_users", db.Query().Table("users").Where("active", true))

// Create a view from raw SQL
err := db.Schema().CreateViewRaw("user_summary", "SELECT user_id, COUNT(*) FROM orders GROUP BY user_id")

// Check existence
exists := db.Schema().HasView("active_users")

// Drop (with if-exists guard)
err := db.Schema().DropViewIfExists("active_users")
```

See [Views Documentation](./docs/views.html) for per-driver notes.

## Observers

```go
type UserObserver struct{}

func (o *UserObserver) Creating(event neat.Event) error {
    log.Println("Creating user")
    return nil
}

func (o *UserObserver) Created(event neat.Event) error {
    log.Println("User created")
    return nil
}

// Register observer
db.Orm().Observe([]neat.ModelToObserver{
    {Model: User{}, Observer: UserObserver{}},
})
```

## Soft Deletes

```go
type User struct {
    neat.SoftDeletes
    ID   uint
    Name string
}

// Soft delete
db.Query().Where("id", 1).Delete()

// Include soft-deleted records
db.Query().WithTrashed().Where("id", 1).First(&user)

// Only soft-deleted records
db.Query().OnlyTrashed().Where("id", 1).First(&user)

// Restore soft-deleted record
db.Query().Restore(&user)

// Force delete (permanent)
db.Query().ForceDelete(&user)
```

## Associations

```go
type Post struct {
    ID     uint
    Title  string
    UserID uint
}

type User struct {
    ID    uint
    Name  string
    Posts []Post
}

// Eager loading
db.Query().With("posts").Where("id", 1).First(&user)

// Lazy loading
db.Query().Load(&user, "posts")

// Association operations
db.Query().Association("posts").Append(&user, &post)
```

## Array-Backed Sources

Query in-memory slices as if they were database tables — useful for static data (statuses, countries), mocking in tests, or querying computed datasets.

```go
type Status struct {
    ID    int    `db:"id"`
    Name  string `db:"name"`
    Color string `db:"color"`
}

statuses := []Status{
    {ID: 1, Name: "Pending", Color: "yellow"},
    {ID: 2, Name: "Active",  Color: "green"},
    {ID: 3, Name: "Inactive", Color: "red"},
}

var results []Status
err := db.Query().
    Model(neat.NewArraySourceFrom(statuses)).
    Where("name = ?", "Active").
    OrderBy("id", "asc").
    Get(&results)
```

`NewArraySourceFrom` accepts a slice of structs or `[]map[string]any`. The table name is auto-generated and the schema is inferred from the data. See [Array Source Documentation](./docs/array-source.html) and the [array driver examples](./examples/array-driver).

## CSVDB Driver

Query a directory of CSV files as if they were database tables — useful for data exports, reports, test fixtures, and datasets. The directory is the database; each `.csv` file is a table; the filename (without `.csv`) is the table name.

```go
config := neat.DBConfig{
    Default: "csv_db",
    Connections: map[string]neat.ConnectionConfig{
        "csv_db": {
            Driver:   "csvdb",
            Database: "data/",   // directory path
        },
    },
}

db, _ := neat.New(config)
defer db.Close()

// data/users.csv → "users" table
var users []User
err := db.Query().Model(&User{}).Where("active = ?", true).Get(&users)
```

Column types are inferred from the CSV data (INTEGER, REAL, DATETIME, TEXT). The CSV header row defines column names. All tables are loaded into an in-memory SQLite database at connection open time, so the full query builder works: WHERE, JOIN, ORDER BY, aggregates, etc. See the [csvdb-driver example](./examples/csvdb-driver) and the [proposal](./docs/proposals/csv-directory-driver.md).

## Supported Databases

- MySQL 5.7+
- PostgreSQL 12+
- SQLite 3+
- SQL Server 2017+
- Turso (SQLite edge)
- Oracle
- CSVDB (CSV directory as database)

### Driver Compatibility Matrix

| Feature | SQLite | MySQL | PostgreSQL | Oracle | Turso | SQL Server | CSVDB |
|---------|--------|-------|------------|--------|-------|------------|-------|
| **Basic Operations** |
| Open Connection | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Close Connection | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Ping/Health Check | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Transactions** |
| BeginTx with Options | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Savepoints | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Isolation Levels | Limited | Full | Full | Full | Limited | Full | Limited |
| **Placeholder Style** |
| Placeholder Format | `?` | `?` | `$1, $2` | `:1, :2` | `?` | `@p1, @p2` | `?` |
| **DSN Support** |
| URL-based DSN | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Query Parameters | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | N/A |
| **Connection Pool** |
| MaxOpenConns | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ (pinned to 1) |
| MaxIdleConns | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ (pinned to 1) |
| QueryTimeout | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Optimizations** |
| SQLite PRAGMAs | ✅ | ❌ | ❌ | ❌ | ✅ | ❌ | ✅ |
| MySQL Charset | ❌ | ✅ (utf8mb4) | ❌ | ❌ | ❌ | ❌ | ❌ |
| PostgreSQL SSL | ❌ | ❌ | ✅ (require) | ❌ | ❌ | ❌ | ❌ |

**Notes:**
- **Turso** is a SQLite edge database, so it shares SQLite's placeholder style and PRAGMA support
- **CSVDB** uses in-memory SQLite under the hood, so it shares SQLite's placeholder style, PRAGMA support, and single-connection constraint. The `Database` field holds a directory path (not a file path or DSN)
- **Transaction Isolation Levels**: SQLite has limited isolation level support (SERIALIZABLE only), MySQL/PostgreSQL/Oracle/SQL Server support all standard levels
- **Savepoints**: All drivers support savepoints through the standard `database/sql` interface
- **Connection Pool**: All drivers support standard `database/sql` connection pooling parameters

## Connection Pool Configuration

Neat ORM provides sensible defaults for connection pooling, but you can customize these settings based on your application's needs.

### Pool Configuration Options

Pool settings are configured on the `DBConfig.Pool` field using `neat.PoolConfig` (durations use `time.Duration`):

```go
config := neat.DBConfig{
    Default: "default",
    Connections: map[string]neat.ConnectionConfig{
        "default": {
            Driver:   "mysql",
            Host:     "localhost",
            Port:     3306,
            Database: "mydb",
            Username: "user",
            Password: "password",
        },
    },
    Pool: neat.PoolConfig{
        MaxIdleConns:    5,                    // Maximum number of idle connections
        MaxOpenConns:    25,                   // Maximum number of open connections
        ConnMaxLifetime: 3600 * time.Second,   // Connection lifetime (1 hour)
        ConnMaxIdleTime: 300 * time.Second,    // Maximum idle time (5 minutes)
        QueryTimeout:    30 * time.Second,     // Query timeout (default: 30 seconds)
    },
}

db, err := neat.New(config)
```

### SQLite-Specific Configuration

**Why SQLite uses MaxOpen=1:**

SQLite has a fundamental limitation: it allows only one writer at a time. Multiple concurrent write operations will cause "database is locked" errors. To prevent this, Neat automatically enforces `MaxOpenConns=1` and `MaxIdleConns=1` for SQLite connections, regardless of your pool configuration.

**SQLite Pool Defaults:**
- `MaxOpenConns`: 1 (enforced to prevent writer contention)
- `MaxIdleConns`: 1 (enforced to prevent writer contention)
- `QueryTimeout`: 30 seconds
- **PRAGMA Optimizations**: Automatically applied (WAL mode, foreign keys, busy timeout)

**Turso (SQLite Edge):**

Turso is a SQLite edge database that inherits SQLite's single-writer limitation. The same pool constraints apply to Turso connections:
- `MaxOpenConns`: 1 (enforced to prevent writer contention)
- `MaxIdleConns`: 1 (enforced to prevent writer contention)
- `QueryTimeout`: 30 seconds
- **PRAGMA Optimizations**: Automatically applied (WAL mode, foreign keys, busy timeout)

**When to use SQLite/Turso:**
- Development and testing
- Low-traffic applications
- Single-process services
- Embedded applications
- Edge computing scenarios (Turso)

**When to avoid SQLite/Turso:**
- High-concurrency write workloads
- Multi-process services requiring concurrent writes
- Production applications with significant write traffic

### MySQL/PostgreSQL/SQL Server/Oracle Configuration

These databases support true concurrent connections and can handle larger connection pools.

**Production Defaults:**
- `MaxOpenConns`: 25 (adjust based on your database server capacity)
- `MaxIdleConns`: 5 (keeps a small pool of ready connections)
- `ConnMaxLifetime`: 3600 seconds (1 hour)
- `ConnMaxIdleTime`: 300 seconds (5 minutes)
- `QueryTimeout`: 30 seconds

**Development Defaults:**
- `MaxOpenConns`: 10 (lower for local development)
- `MaxIdleConns`: 2 (minimal idle connections)
- `ConnMaxLifetime`: 1800 seconds (30 minutes)
- `ConnMaxIdleTime`: 300 seconds (5 minutes)
- `QueryTimeout`: 30 seconds

### Workload-Specific Recommendations

The examples below show only the `Pool` field of `neat.DBConfig`. Durations use `time.Duration`.

**Read-Heavy Workloads:**
```go
Pool: neat.PoolConfig{
    MaxIdleConns:    10,                  // More idle connections for quick reads
    MaxOpenConns:    50,                  // Higher open connection limit
    ConnMaxLifetime: 7200 * time.Second,  // Longer lifetime (2 hours)
    QueryTimeout:    10 * time.Second,    // Shorter timeout for reads
}
```

**Write-Heavy Workloads:**
```go
Pool: neat.PoolConfig{
    MaxIdleConns:    5,                   // Fewer idle connections
    MaxOpenConns:    20,                  // Moderate open connection limit
    ConnMaxLifetime: 3600 * time.Second,  // Standard lifetime (1 hour)
    QueryTimeout:    60 * time.Second,    // Longer timeout for writes
}
```

**High-Concurrency Applications:**
```go
Pool: neat.PoolConfig{
    MaxIdleConns:    20,                  // Larger idle pool
    MaxOpenConns:    100,                 // High open connection limit
    ConnMaxLifetime: 1800 * time.Second,  // Shorter lifetime (30 minutes)
    ConnMaxIdleTime: 120 * time.Second,   // Shorter idle time (2 minutes)
    QueryTimeout:    30 * time.Second,
}
```

**Low-Traffic Services:**
```go
Pool: neat.PoolConfig{
    MaxIdleConns:    2,                   // Minimal idle connections
    MaxOpenConns:    5,                   // Low open connection limit
    ConnMaxLifetime: 3600 * time.Second,  // Standard lifetime
    QueryTimeout:    30 * time.Second,
}
```

### Monitoring and Tuning

Monitor your connection pool metrics to optimize performance:

- **Pool Hit Rate**: High hit rate indicates good pool utilization
- **Wait Time**: Long wait times suggest increasing `MaxOpenConns`
- **Connection Age**: Frequent reconnections suggest increasing `ConnMaxLifetime`
- **Idle Connections**: Too many idle connections waste resources, reduce `MaxIdleConns`

### Important Notes

- **SQLite Constraints**: SQLite pool settings are automatically overridden to prevent "database is locked" errors
- **Query Timeout**: Default is 30 seconds, adjust based on your query complexity
- **Connection Lifetime**: Set shorter lifetimes for cloud databases with connection limits
- **Pool Size**: Never set `MaxOpenConns` higher than your database server's max connection limit

## API Documentation

For detailed API documentation, see the [docs](./docs) directory.

## Examples

For more examples, see the [examples](./examples) directory.

## License

This project is licensed under the GNU Affero General Public License v3.0 - see the LICENSE file for details.

## Testing

### Running Integration Tests with Docker Compose

The project includes a Docker Compose configuration for running integration tests locally with MySQL and PostgreSQL:

```bash
# Start the database containers
docker-compose up -d

# Run MySQL integration tests
go test -v -tags=integration ./integration_tests/mysql/...

# Run PostgreSQL integration tests
go test -v -tags=integration ./integration_tests/postgres/...

# Stop the containers when done
docker-compose down
```

The Docker Compose setup includes:
- **MySQL 8.0** on port `3306` (user: `root`, password: `root`, database: `test`)
- **PostgreSQL 15** on port `55432` (user: `test`, password: `test`, database: `test`)

### Running Unit Tests

```bash
go test ./...
```

### Running All Tests

```bash
go test -v ./...
```

### Generating Coverage Reports

To generate a coverage report locally:

```bash
# Generate coverage profile
go test -coverprofile=coverage.out -covermode=atomic ./...

# View coverage percentage in terminal
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html
```

The HTML report can be opened in a browser to see detailed coverage information for each file and function.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## Roadmap

### Current Status

Neat ORM is actively developed with the following features implemented:
- ✅ Query Builder with fluent interface and Sugar Methods
- ✅ ORM with model support
- ✅ Schema Builder for database operations
- ✅ Advanced Migration system (`Migrator` package)
- ✅ Seeder system for data seeding
- ✅ Factory pattern for test data
- ✅ Multiple database support (MySQL, PostgreSQL, SQLite, SQL Server, Turso, Oracle, CSVDB)
- ✅ Transaction support with savepoints and callbacks
- ✅ Observer system for model events
- ✅ Soft deletes with multiple strategies (NULL and Max-Date)
- ✅ Associations (BelongsTo, HasMany, HasOne, Polymorphic)
- ✅ View management (CreateView, CreateViewRaw, DropView, DropViewIfExists, HasView)
- ✅ Array-backed sources (NewArraySourceFrom for struct and map slices)
- ✅ CSVDB driver (query CSV directory as database with type inference)
- ✅ Connection pooling
- ✅ Context support

### Planned Features

- [ ] Additional migration drivers (SQL, custom drivers)
- [ ] More relationship types (HasManyThrough, BelongsToMany)
- [ ] Query caching
- [ ] Full-text search support
- [ ] Scopes and global scopes
- [ ] Mutators and accessors
- [ ] Model casting
- [ ] Validation integration
- [ ] Query builder debugging tools
- [ ] Additional database drivers

For detailed implementation plans, see [docs/implementation/gaps.md](./docs/implementation/gaps.md).
