# Factory Pattern Usage

This example demonstrates the Factory pattern for creating test data in your database.

## Features Demonstrated

- Creating single model instances with `Factory.Table().Create()`
- Bulk creation with `Factory.Table().Count().Create()`
- Creating models without firing events with `Factory.Table().CreateQuietly()`
- Creating in-memory instances without persistence with `Factory.Make()`
- Bulk make operations with `Factory.Count().Make()`

## Running the Example

```bash
cd examples/factory
go run main.go
```

## Running the Tests

```bash
cd examples/factory
go test
```

## Prerequisites

- SQLite database (or modify the DSN to use your preferred database)
- The example creates a `users` table automatically

## Factory Methods

### Table(table string) Factory
Sets the table name for database operations. Required for Create() and CreateQuietly().

### Count(count int) Factory
Sets the number of models that should be generated. Can be chained with Create, CreateQuietly, or Make for bulk operations.

### Create(value any, attributes ...map[string]any) (any, error)
Creates a model and persists it to the database, returning the created instance(s). Requires Table() to be called first.

### CreateQuietly(value any, attributes ...map[string]any) (any, error)
Creates a model and persists it to the database without firing any model events, returning the created instance(s). Requires Table() to be called first.

### Make(value any, attributes ...map[string]any) (any, error)
Creates a model instance in memory but does not persist it to the database. Useful for testing without side effects. Returns the created instance(s).

## Use Cases

- **In-memory testing**: Use Factory.Make() to create instances without database persistence
- **Bulk in-memory operations**: Use Factory.Count().Make() to create multiple instances in memory
- **Database seeding**: Use Factory.Table().Create() to persist models to the database
- **Event-free creation**: Use Factory.Table().CreateQuietly() when you don't want to trigger model observers
- **Bulk database operations**: Use Factory.Table().Count().Create() to create multiple records efficiently

## Example Usage

```go
// Single database creation
user, err := db.Factory().Table("users").Create(&User{Name: "John"})

// Bulk database creation
users, err := db.Factory().Table("users").Count(3).Create(&User{Name: "Template"})

// In-memory creation
user, err := db.Factory().Make(&User{Name: "Test"})

// Bulk in-memory creation
users, err := db.Factory().Count(3).Make(&User{Name: "Template"})
```
