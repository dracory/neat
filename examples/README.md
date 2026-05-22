# Neat Examples

This directory contains examples demonstrating various use cases of the neat Go ORM library.

## Available Examples

- **[basic-orm](./basic-orm/)** - Basic ORM operations with the query builder (CRUD operations, simple queries)
- **[schema-builder](./schema-builder/)** - Creating and modifying database tables using the schema builder
- **[models](./models/)** - Using struct-based models for type-safe database operations
- **[configuration](./configuration/)** - Various configuration options for database connections
- **[migrations](./migrations/)** - Database migration examples with relationships and constraints
- **[advanced-queries](./advanced-queries/)** - Advanced query builder features (joins, aggregations, subqueries)

## Running Examples

Each example is self-contained and can be run independently:

```bash
cd examples/<example-name>
go run main.go
```

## Prerequisites

Most examples use SQLite by default for simplicity. You can modify the DSN in each example to use PostgreSQL, MySQL, or SQL Server as needed.

Make sure you have the required database server running and valid credentials before running examples that connect to external databases.

## Database Setup

For a quick start with SQLite, no additional setup is required. The examples will create a local SQLite file.

For other databases, ensure you have:

1. The database server installed and running
2. A database created
3. Valid credentials with appropriate permissions

Modify the connection string in each example's `main.go` file to match your database configuration.
