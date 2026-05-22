# Migrations

This example demonstrates how to create and run database migrations using the schema builder.

## Features Demonstrated

- Creating tables with various column types
- Adding foreign key constraints with cascade delete
- Creating relationships between tables (users, posts, comments)
- Adding indexes to existing tables
- Adding columns to existing tables
- Using default values and nullable columns
- Timestamps and soft deletes

## Running the Example

```bash
cd examples/migrations
go run main.go
```

## Prerequisites

- SQLite database (or modify the DSN to use your preferred database)
