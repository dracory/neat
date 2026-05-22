# Models

This example demonstrates using struct-based models with the ORM for type-safe database operations.

## Features Demonstrated

- Defining models with struct tags
- Custom table names with TableName() method
- Creating records with models
- Finding records by ID
- Updating records with models
- Querying multiple records into slices
- Working with foreign keys
- Soft deletes

## Running the Example

```bash
cd examples/models
go run main.go
```

## Prerequisites

- SQLite database (or modify the DSN to use your preferred database)
- `users` and `posts` tables should exist in the database with appropriate schema
