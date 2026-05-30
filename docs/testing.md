# Testing Guide

This guide covers testing strategies and best practices for Neat ORM.

## Overview

Neat ORM provides comprehensive testing support including:
- Unit tests for individual components
- Integration tests with real databases
- Test database setup with Docker Compose
- Factory pattern for test data generation
- Seeder support for test data

## Running Tests

### Running Unit Tests

Unit tests test individual components without requiring a database connection:

```bash
# Run all unit tests
go test ./...

# Run unit tests with verbose output
go test -v ./...

# Run unit tests for a specific package
go test ./database/...

# Run unit tests with coverage
go test -cover ./...
```

### Running Integration Tests

Integration tests require real database connections. The project includes Docker Compose configurations for MySQL and PostgreSQL:

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

### Running All Tests

```bash
# Run all tests (unit and integration)
go test -v ./...

# Run all tests with coverage
go test -v -cover ./...
```

## Test Database Setup

### Using SQLite for Testing

SQLite is ideal for testing as it doesn't require a separate server:

```go
db, err := neat.NewFromDSN("sqlite://file::memory:?cache=shared")
if err != nil {
    t.Fatal(err)
}
defer db.Close()
```

### Using MySQL for Testing

```go
db, err := neat.NewFromDSN("mysql://root:root@localhost:3306/test")
if err != nil {
    t.Fatal(err)
}
defer db.Close()
```

### Using PostgreSQL for Testing

```go
db, err := neat.NewFromDSN("postgres://test:test@localhost:55432/test?sslmode=disable")
if err != nil {
    t.Fatal(err)
}
defer db.Close()
```

## Writing Unit Tests

### Testing Query Builder

```go
func TestQueryWhere(t *testing.T) {
    db := setupTestDatabase()
    defer db.Close()

    // Create test data
    user := &User{Name: "John", Email: "john@example.com"}
    err := db.Query().Create(user)
    if err != nil {
        t.Fatal(err)
    }

    // Test query
    var result User
    err = db.Query().Where("name", "John").First(&result)
    if err != nil {
        t.Fatal(err)
    }

    if result.Name != "John" {
        t.Errorf("Expected name 'John', got '%s'", result.Name)
    }
}
```

### Testing Schema Builder

```go
func TestSchemaCreateTable(t *testing.T) {
    db := setupTestDatabase()
    defer db.Close()

    err := db.Schema().Create("test_table", func(table neat.Blueprint) {
        table.ID()
        table.String("name")
    })
    if err != nil {
        t.Fatal(err)
    }

    // Verify table exists
    exists := db.Schema().HasTable("test_table")
    if !exists {
        t.Error("Table was not created")
    }
}
```

### Testing Migrations

```go
func TestMigration(t *testing.T) {
    db := setupTestDatabase()
    defer db.Close()

    // Register migration
    migration.RegisterMigration("test_migration", migration.Migration{
        Up: func(schema contractsschema.Schema) error {
            return schema.Create("test_table", func(table contractsschema.Blueprint) {
                table.ID()
                table.String("name")
            })
        },
        Down: func(schema contractsschema.Schema) error {
            return schema.DropTableIfExists("test_table")
        },
    })

    // Run migration
    err := db.Migrate("./migrations")
    if err != nil {
        t.Fatal(err)
    }

    // Verify table exists
    exists := db.Schema().HasTable("test_table")
    if !exists {
        t.Error("Migration did not create table")
    }

    // Rollback
    err = db.MigrateDown(1, "./migrations")
    if err != nil {
        t.Fatal(err)
    }
}
```

## Writing Integration Tests

### Integration Test Structure

Integration tests should be placed in the `integration_tests` directory:

```
integration_tests/
├── mysql/
│   ├── helper.go
│   ├── mysql_connection_test.go
│   └── ...
└── postgres/
    ├── helper.go
    ├── postgres_connection_test.go
    └── ...
```

### Example Integration Test

```go
// integration_tests/mysql/query_test.go
package mysql

import (
    "testing"
    
    "github.com/dracory/neat"
)

func TestMySQLQuery(t *testing.T) {
    db := setupMySQLTestDatabase()
    defer db.Close()

    // Test query functionality
    var users []User
    err := db.Query().Where("name", "John").Get(&users)
    if err != nil {
        t.Fatal(err)
    }

    if len(users) == 0 {
        t.Error("Expected to find users")
    }
}
```

## Using Factories in Tests

### In-Memory Test Data

```go
func TestUserValidation(t *testing.T) {
    db := setupTestDatabase()
    defer db.Close()

    // Create user in memory
    user, err := db.Factory().Make(&User{Name: "Test"})
    if err != nil {
        t.Fatal(err)
    }

    // Test validation logic
    if user.Name == "" {
        t.Error("Name is required")
    }
}
```

### Database Test Data

```go
func TestUserCreation(t *testing.T) {
    db := setupTestDatabase()
    defer db.Close()

    // Create table
    err := db.Schema().Create("users", func(table neat.Blueprint) {
        table.ID()
        table.String("name")
        table.String("email").Unique()
    })
    if err != nil {
        t.Fatal(err)
    }

    // Create test data
    user, err := db.Factory().Table("users").Create(&User{Name: "John"})
    if err != nil {
        t.Fatal(err)
    }

    // Verify creation
    if user.ID == 0 {
        t.Error("Expected ID to be set")
    }
}
```

### Bulk Test Data

```go
func TestBulkOperations(t *testing.T) {
    db := setupTestDatabase()
    defer db.Close()

    // Create bulk test data
    users, err := db.Factory().Table("users").Count(100).Create(&User{Name: "Template"})
    if err != nil {
        t.Fatal(err)
    }

    // Verify count
    count, _ := db.Query().Table("users").Count()
    if count != 100 {
        t.Errorf("Expected 100 users, got %d", count)
    }
}
```

## Using Seeders in Tests

```go
func TestWithSeeders(t *testing.T) {
    db := setupTestDatabase()
    defer db.Close()

    // Create seeder
    seeder := &UserSeeder{db: db}

    // Run seeder
    err := db.Seed([]contractsseeder.Seeder{seeder})
    if err != nil {
        t.Fatal(err)
    }

    // Run test
    var users []User
    err = db.Query().Table("users").Get(&users)
    if err != nil {
        t.Fatal(err)
    }

    if len(users) == 0 {
        t.Error("Expected seeded data")
    }
}
```

### Testing with SeedOnce

```go
func TestSeedOnce(t *testing.T) {
    db := setupTestDatabase()
    defer db.Close()

    seeder := &UserSeeder{db: db}
    facade := db.Seeder()
    facade.Register([]contractsseeder.Seeder{seeder})

    // First call - runs the seeder
    err := facade.CallOnce([]contractsseeder.Seeder{seeder})
    if err != nil {
        t.Fatal(err)
    }

    // Second call - skips the seeder
    err = facade.CallOnce([]contractsseeder.Seeder{seeder})
    if err != nil {
        t.Fatal(err)
    }

    // Reset for next test
    facade.ResetCallOnce()
}
```

## Test Helpers

### Setup Test Database

```go
func setupTestDatabase() *neat.Database {
    db, err := neat.NewFromDSN("sqlite://file::memory:?cache=shared")
    if err != nil {
        panic(err)
    }
    return db
}
```

### Setup Test Table

```go
func setupTestTable(db *neat.Database, tableName string) error {
    return db.Schema().Create(tableName, func(table neat.Blueprint) {
        table.ID()
        table.String("name")
        table.String("email").Unique()
        table.Timestamps()
    })
}
```

### Cleanup Test Database

```go
func cleanupTestDatabase(db *neat.Database) {
    tables := db.Schema().GetTableListing()
    for _, table := range tables {
        db.Schema().Drop(table)
    }
}
```

## Test Best Practices

1. **Use in-memory databases**: Prefer SQLite in-memory for unit tests
2. **Clean up after tests**: Always clean up test data in teardown
3. **Use factories for test data**: Use factories instead of manual data creation
4. **Test edge cases**: Test error conditions and edge cases
5. **Use table-driven tests**: Use table-driven tests for multiple scenarios
6. **Mock external dependencies**: Mock external services when possible
7. **Keep tests isolated**: Each test should be independent
8. **Use descriptive test names**: Make test names clear and descriptive

## Table-Driven Tests

```go
func TestQueryWhere(t *testing.T) {
    tests := []struct {
        name     string
        where    string
        value    any
        expected int
    }{
        {"simple equals", "name", "John", 1},
        {"greater than", "age > ?", 18, 2},
        {"in clause", "id IN (?)", []any{1, 2}, 2},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            db := setupTestDatabase()
            defer db.Close()

            // Setup test data
            // ...

            count, err := db.Query().Where(tt.where, tt.value).Count()
            if err != nil {
                t.Fatal(err)
            }

            if count != tt.expected {
                t.Errorf("Expected %d results, got %d", tt.expected, count)
            }
        })
    }
}
```

## Testing Observers

```go
func TestObserver(t *testing.T) {
    db := setupTestDatabase()
    defer db.Close()

    // Register observer
    observer := &TestObserver{}
    db.Orm().Observe([]neat.ModelToObserver{
        {Model: User{}, Observer: observer},
    })

    // Create user
    user := &User{Name: "John"}
    err := db.Query().Create(user)
    if err != nil {
        t.Fatal(err)
    }

    // Verify observer was called
    if !observer.CreatingCalled {
        t.Error("Creating observer was not called")
    }
    if !observer.CreatedCalled {
        t.Error("Created observer was not called")
    }
}

type TestObserver struct {
    CreatingCalled bool
    CreatedCalled  bool
}

func (o *TestObserver) Creating(event neat.Event) error {
    o.CreatingCalled = true
    return nil
}

func (o *TestObserver) Created(event neat.Event) error {
    o.CreatedCalled = true
    return nil
}
```

## Testing Transactions

```go
func TestTransaction(t *testing.T) {
    db := setupTestDatabase()
    defer db.Close()

    // Test successful transaction
    err := db.Transaction(func(tx neat.Query) error {
        user := &User{Name: "John"}
        return tx.Create(user)
    })
    if err != nil {
        t.Fatal(err)
    }

    // Test failed transaction (rollback)
    err = db.Transaction(func(tx neat.Query) error {
        user := &User{Name: "Jane"}
        err := tx.Create(user)
        if err != nil {
            return err
        }
        return fmt.Errorf("force rollback")
    })
    if err == nil {
        t.Error("Expected transaction to fail")
    }

    // Verify rollback
    count, _ := db.Query().Table("users").Count()
    if count != 1 {
        t.Errorf("Expected 1 user after rollback, got %d", count)
    }
}
```

## Continuous Integration

### GitHub Actions Example

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      mysql:
        image: mysql:8.0
        env:
          MYSQL_ROOT_PASSWORD: root
          MYSQL_DATABASE: test
        ports:
          - 3306:3306
      
      postgres:
        image: postgres:15
        env:
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
          POSTGRES_DB: test
        ports:
          - 5432:5432
    
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run unit tests
        run: go test -v ./...
      
      - name: Run integration tests
        run: go test -v -tags=integration ./integration_tests/...
```

## Troubleshooting

### Tests fail with database connection error
- Ensure Docker containers are running for integration tests
- Check database credentials in test configuration
- Verify database is accessible on the expected port

### Tests are slow
- Use SQLite in-memory databases for unit tests
- Limit the number of test records created
- Use parallel testing with `t.Parallel()`

### Test data conflicts
- Use unique test data for each test
- Clean up test data in teardown
- Use transactions that rollback after each test

### Integration tests fail locally but pass in CI
- Check CI database configuration matches local setup
- Verify database versions are compatible
- Check for environment-specific settings
