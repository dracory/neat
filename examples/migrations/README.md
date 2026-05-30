# Migrations

This example demonstrates both approaches for managing database schema changes in Neat ORM:

1. **Schema Builder Approach** - Direct schema manipulation without version control
2. **Migration System Approach** - Version-controlled migrations with rollback capabilities

## Features Demonstrated

### Schema Builder Approach
- Direct table creation using the schema builder
- Adding columns to existing tables
- Creating indexes
- Simple and straightforward for one-time setup
- No version control or rollback capabilities

### Migration System Approach
- Version-controlled migrations with Up/Down methods
- Migration registration in global registry
- Running pending migrations
- Rolling back migrations
- Checking migration status
- Batch tracking for migration groups

## Running the Example

```bash
cd examples/migrations
go run main.go
```

This will run both examples sequentially:
1. First, it demonstrates the schema builder approach (creates `example_schema.db`)
2. Then, it demonstrates the migration system approach (creates `example_migration.db`)

## Prerequisites

- SQLite database (or modify the DSN to use your preferred database)

## When to Use Each Approach

### Use Schema Builder When:
- You need a simple one-time database setup
- You don't need version control for schema changes
- You're working on a prototype or development environment
- Schema changes are infrequent and simple

### Use Migration System When:
- You need to track schema changes over time
- You work in a team and need consistent database states
- You need the ability to rollback schema changes
- You're deploying to production environments
- You have complex schema evolution requirements
