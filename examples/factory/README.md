# Factory Pattern Usage

This example demonstrates the Factory pattern for creating test data in your database.

## Features Demonstrated

- Creating single model instances with `Query().Table().Create()`
- Creating models without firing events with `Query().Table().WithoutEvents().Create()`
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

### Make(value any, attributes ...map[string]any) error
Creates a model instance in memory but does not persist it to the database. Useful for testing without side effects.

### Count(count int) Factory
Sets the number of models that should be generated. Can be chained with Make for bulk operations.

## Query Methods for Database Operations

### Query().Table(name).Create(value) error
Creates a model and persists it to the database. You must specify the table name before calling Create.

### Query().Table(name).WithoutEvents().Create(value) error
Creates a model and persists it to the database without firing any model events. Useful for seeding data without triggering observers.

## Use Cases

- **In-memory testing**: Use Factory.Make() to create instances without database persistence
- **Bulk in-memory operations**: Use Factory.Count().Make() to create multiple instances in memory
- **Database seeding**: Use Query().Table().Create() to persist models to the database
- **Event-free creation**: Use Query().Table().WithoutEvents().Create() when you don't want to trigger model observers

## Note

The Factory pattern in this codebase is primarily used for creating in-memory instances for testing. For database operations, use the Query builder with Table() to specify the target table.
