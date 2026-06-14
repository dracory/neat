# Neat ORM

[![Tests Status](https://github.com/dracory/neat/actions/workflows/tests.yml/badge.svg?branch=main)](https://github.com/dracory/neat/actions/workflows/tests.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/dracory/neat)](https://goreportcard.com/report/github.com/dracory/neat)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/dracory/neat)](https://pkg.go.dev/github.com/dracory/neat)
[![codecov](https://codecov.io/gh/dracory/neat/branch/main/graph/badge.svg)](https://codecov.io/gh/dracory/neat)

A powerful and elegant ORM (Object-Relational Mapping) library for Go, designed to provide a clean and intuitive interface for database operations. Neat aims to feature parity with Laravel's Eloquent ORM while being built from scratch without GORM dependencies.

## Features

- **Query Builder**: Fluent and intuitive query building interface
- **ORM**: Full ORM support with models and relationships
- **Schema Builder**: Database schema creation and modification
- **Migrations**: Complete database migration system with schema builder, rollback support, and ORM driver integration (major advantage over most Go ORMs)
- **Seeders**: Database seeding for test and initial data
- **Factories**: Test data generation with factory pattern
- **Multiple Database Support**: MySQL, PostgreSQL, SQLite, SQL Server, Turso, Oracle
- **Transactions**: Robust transaction support
- **Observers**: Model lifecycle event system
- **Soft Deletes**: Soft delete functionality with multiple strategies (NULL-based and max-date sentinel)
- **Associations**: BelongsTo, HasMany, HasOne, PolymorphicBelongsTo, PolymorphicHasMany relationships
- **Connection Pooling**: Efficient connection management
- **Context Support**: Full context.Context support throughout
- **Query Method Aliases**: Sequelize-style (FindAll, FindOne, Destroy) and Django-style (Filter, Exclude, All)
- **ToSql Interface**: SQL generation without execution
- **Security Hardening**: SQL injection prevention with identifier validation

## Key Advantage: Complete Migration System

> **🚀 Most Go ORMs lack comprehensive schema migration support.** Neat ORM includes a complete migration system with schema builder, rollback support, and ORM driver integration - something most competitors either lack entirely or require third-party tools for.

## Installation

```bash
go get github.com/dracory/neat
```

## Documentation

- **[HTML Documentation](https://html-preview.github.io/?url=https://github.com/dracory/neat/blob/main/docs/index.html)** - Browse documentation in your browser
- **[Examples](./examples)** - Practical examples for various features
- **[API Reference](./docs/api-reference.md)** - Complete API documentation
```

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

Neat maps Go structs to database tables using struct tags. For detailed information on table names, column names, and tag priority, see [Models Documentation](./docs/models.md).

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

## Supported Databases

- MySQL 5.7+
- PostgreSQL 12+
- SQLite 3+
- SQL Server 2017+
- Turso (SQLite edge)
- Oracle

### Driver Compatibility Matrix

| Feature | SQLite | MySQL | PostgreSQL | Oracle | Turso | SQL Server |
|---------|--------|-------|------------|--------|-------|------------|
| **Basic Operations** |
| Open Connection | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Close Connection | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Ping/Health Check | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Transactions** |
| BeginTx with Options | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Savepoints | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Isolation Levels | Limited | Full | Full | Full | Limited | Full |
| **Placeholder Style** |
| Placeholder Format | `?` | `?` | `$1, $2` | `:1, :2` | `?` | `@p1, @p2` |
| **DSN Support** |
| URL-based DSN | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Query Parameters | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Connection Pool** |
| MaxOpenConns | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| MaxIdleConns | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| QueryTimeout | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Optimizations** |
| SQLite PRAGMAs | ✅ | ❌ | ❌ | ❌ | ✅ | ❌ |
| MySQL Charset | ❌ | ✅ (utf8mb4) | ❌ | ❌ | ❌ | ❌ |
| PostgreSQL SSL | ❌ | ❌ | ✅ (require) | ❌ | ❌ | ❌ |

**Notes:**
- **Turso** is a SQLite edge database, so it shares SQLite's placeholder style and PRAGMA support
- **Transaction Isolation Levels**: SQLite has limited isolation level support (SERIALIZABLE only), MySQL/PostgreSQL/Oracle/SQL Server support all standard levels
- **Savepoints**: All drivers support savepoints through the standard `database/sql` interface
- **Connection Pool**: All drivers support standard `database/sql` connection pooling parameters

## Connection Pool Configuration

Neat ORM provides sensible defaults for connection pooling, but you can customize these settings based on your application's needs.

### Pool Configuration Options

```go
poolConfig := db.PoolConfig{
    MaxIdleConns:    5,   // Maximum number of idle connections
    MaxOpenConns:    25,  // Maximum number of open connections
    ConnMaxLifetime: 3600, // Connection lifetime in seconds (1 hour)
    ConnMaxIdleTime: 300, // Maximum idle time in seconds (5 minutes)
    QueryTimeout:    30,  // Query timeout in seconds (default: 30)
}

db, err := neat.New(config, neat.WithPool(poolConfig))
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

**Read-Heavy Workloads:**
```go
poolConfig := db.PoolConfig{
    MaxIdleConns:    10,  // More idle connections for quick reads
    MaxOpenConns:    50,  // Higher open connection limit
    ConnMaxLifetime: 7200, // Longer lifetime (2 hours)
    QueryTimeout:    10,  // Shorter timeout for reads
}
```

**Write-Heavy Workloads:**
```go
poolConfig := db.PoolConfig{
    MaxIdleConns:    5,   // Fewer idle connections
    MaxOpenConns:    20,  // Moderate open connection limit
    ConnMaxLifetime: 3600, // Standard lifetime (1 hour)
    QueryTimeout:    60,  // Longer timeout for writes
}
```

**High-Concurrency Applications:**
```go
poolConfig := db.PoolConfig{
    MaxIdleConns:    20,  // Larger idle pool
    MaxOpenConns:    100, // High open connection limit
    ConnMaxLifetime: 1800, // Shorter lifetime (30 minutes)
    ConnMaxIdleTime: 120, // Shorter idle time (2 minutes)
    QueryTimeout:    30,
}
```

**Low-Traffic Services:**
```go
poolConfig := db.PoolConfig{
    MaxIdleConns:    2,   // Minimal idle connections
    MaxOpenConns:    5,   // Low open connection limit
    ConnMaxLifetime: 3600, // Standard lifetime
    QueryTimeout:    30,
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

Contributions are welcome! Please open an issue or submit a pull request. For detailed contribution guidelines, see [CONTRIBUTING.md](./CONTRIBUTING.md).

## Roadmap

### Current Status

Neat ORM is actively developed with the following features implemented:
- ✅ Query Builder with fluent interface
- ✅ ORM with model support
- ✅ Schema Builder for database operations
- ✅ Migration system (ORM driver)
- ✅ Seeder system for data seeding
- ✅ Factory pattern for test data
- ✅ Multiple database support (MySQL, PostgreSQL, SQLite, SQL Server, Turso)
- ✅ Transaction support
- ✅ Observer system for model events
- ✅ Soft deletes
- ✅ Associations (BelongsTo, HasMany, HasOne)
- ✅ Connection pooling
- ✅ Context support

### Planned Features

- [ ] Additional migration drivers (SQL, custom drivers)
- [ ] More relationship types (HasManyThrough, BelongsToMany)
- [ ] Query caching
- [ ] Full-text search support
- [ ] Polymorphic relationships
- [ ] Scopes and global scopes
- [ ] Mutators and accessors
- [ ] Model casting
- [ ] Validation integration
- [ ] Query builder debugging tools
- [ ] Additional database drivers

For detailed implementation plans, see [docs/implementation/plan.md](./docs/implementation/plan.md).
