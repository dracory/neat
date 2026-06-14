# Neat Examples

This directory contains examples demonstrating various use cases of the neat Go ORM library.

## Available Examples

- **[basic-orm](./basic-orm/)** - Basic ORM operations with the query builder (CRUD operations, simple queries)
- **[schema-builder](./schema-builder/)** - Creating and modifying database tables using the schema builder
- **[models](./models/)** - Using struct-based models for type-safe database operations
- **[configuration](./configuration/)** - Various configuration options for database connections
- **[migrations](./migrations/)** - Database migration examples with relationships and constraints
- **[advanced-queries](./advanced-queries/)** - Advanced query builder features (joins, aggregations, subqueries)
- **[soft-deletes](./soft-deletes/)** - NULL-based soft delete with `soft_deleted_at` column
- **[deleted-at-soft-delete](./deleted-at-soft-delete/)** - NULL-based soft delete with `deleted_at` column (Laravel-compatible)
- **[max-date-soft-delete](./max-date-soft-delete/)** - Max-date sentinel soft delete strategy for NOT NULL compatibility
- **[associations](./associations/)** - Model relationships (BelongsTo, HasMany, HasOne)
- **[observers](./observers/)** - Model lifecycle event observers
- **[factory](./factory/)** - Model factory patterns for testing
- **[json-queries](./json-queries/)** - JSON column querying capabilities
- **[seeders](./seeders/)** - Database seeding examples
- **[transactions](./transactions/)** - Transaction handling patterns
- **[sugar-methods](./sugar-methods/)** - Convenience methods for improved API usability (AsVar methods)

## Soft Delete Strategies

Neat provides three different soft delete strategies to accommodate different use cases and database requirements:

### 1. NULL-Based with `soft_deleted_at` (Standard)
- **Example**: [soft-deletes](./soft-deletes/)
- **Embed**: `soft_delete.SoftDeletes`
- **Column**: `soft_deleted_at` (nullable)
- **Active when**: `soft_deleted_at IS NULL`
- **Deleted when**: `soft_deleted_at IS NOT NULL`
- **Use when**: You want explicit column naming and don't need NOT NULL constraints

### 2. NULL-Based with `deleted_at` (Laravel-Compatible)
- **Example**: [deleted-at-soft-delete](./deleted-at-soft-delete/)
- **Embed**: `soft_delete.DeletedAt`
- **Column**: `deleted_at` (nullable)
- **Active when**: `deleted_at IS NULL`
- **Deleted when**: `deleted_at IS NOT NULL`
- **Use when**: Migrating from Laravel or need Laravel schema compatibility

### 3. Max-Date Sentinel Strategy
- **Example**: [max-date-soft-delete](./max-date-soft-delete/)
- **Embed**: `soft_delete.SoftDeletesMaxDate`
- **Column**: `soft_deleted_at` (NOT NULL)
- **Active when**: `soft_deleted_at > NOW()` (future sentinel value)
- **Deleted when**: `soft_deleted_at <= NOW()` (past or present)
- **Use when**: You need NOT NULL constraints or better index performance

### Strategy Comparison

| Strategy | Column Name | Nullable | Default Value | Index Performance | Laravel Compatible |
|----------|-------------|----------|---------------|------------------|-------------------|
| `SoftDeletes` | `soft_deleted_at` | Yes | NULL | Varies by DB | No |
| `DeletedAt` | `deleted_at` | Yes | NULL | Varies by DB | Yes |
| `SoftDeletesMaxDate` | `soft_deleted_at` | No | `9999-12-31 23:59:59` | Better (range scans) | No |

### Choosing a Strategy

- **Use `SoftDeletes`** for new projects where you want explicit naming
- **Use `DeletedAt`** when migrating from Laravel or working with Laravel-based systems
- **Use `SoftDeletesMaxDate`** when you need NOT NULL constraints or have performance-critical queries

All strategies support the same query methods:
- `Delete()` - Soft delete a record
- `WithSoftDeleted()` - Include soft deleted records in queries
- `OnlySoftDeleted()` - Query only soft deleted records
- `RestoreSoftDeleted()` - Restore a soft deleted record
- `ForceDelete()` - Permanently delete a record

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
