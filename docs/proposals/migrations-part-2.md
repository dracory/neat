# Enhanced Schema Migration Interface - Part 2: Migration Tracking

**Date**: June 15, 2026
**Status**: Proposed
**Priority**: High
**Author**: Neat ORM Team

## Overview

This proposal (Part 2) focuses on implementing a new migration tracking system specifically designed for the interface-based migrations introduced in Part 1. This tracking system will record which migrations have been executed, their batch information, and execution status.

## Motivation

### Why Migration Tracking?

Migration tracking is critical for:

1. **Prevent Duplicate Execution**: Ensure each migration runs only once
2. **Support Rollback**: Know which migrations to rollback and in what order
3. **Batch Management**: Group migrations for logical rollback points
4. **Audit Trail**: Track when migrations were executed for debugging and compliance
5. **Status Monitoring**: Determine which migrations are pending vs completed

### Why a New Implementation?

The existing `database/migration` package is designed for file-based migrations and uses global registration patterns. The new tracking system should:

- Work directly with `MigrationInterface` instead of file discovery
- Support the explicit registration pattern from Part 1
- Eliminate file system scanning dependencies
- Provide a cleaner, more type-safe interface
- Be designed specifically for interface-based migrations

## Migration Table Schema

### Table Structure

The migration tracking table will be called `migration_tracker` and use the following structure. This table will be created by the schema builder as one of the initial migrations.

```go
// Migration tracker structure
type MigrationTracker struct {
    ID          string    // The migration signature (e.g., "2024_06_15_120000_create_users_table")
    Batch       int       // Timestamp ID (YYYYMMDDHHMMSS). Groups the run
    Description string    // The migration description from Description() method
    StartedAt   time.Time // When the migration started
    CompletedAt time.Time // When the migration finished
}
```

### Initial Migration

The `migration_tracker` table itself will be created via an initial migration:

```go
// CreateMigrationTrackerTable creates the migration tracking table
type CreateMigrationTrackerTable struct {
    schema.BaseMigration
}

func (m *CreateMigrationTrackerTable) Signature() string {
    return "2024_06_15_000000_create_migration_tracker_table"
}

func (m *CreateMigrationTrackerTable) Description() string {
    return "Creates the migration_tracker table for tracking migration execution"
}

func (m *CreateMigrationTrackerTable) Up() error {
    return m.GetSchema().Create("migration_tracker", func(blueprint contractsschema.Blueprint) {
        blueprint.String("id").Primary()
        blueprint.Integer("batch")
        blueprint.Text("description")
        blueprint.Timestamp("started_at")
        blueprint.Timestamp("completed_at")
    })
}

func (m *CreateMigrationTrackerTable) Down() error {
    return m.GetSchema().DropIfExists("migration_tracker")
}
```

## Batch Management

### Batch Concept

Migrations should be executed in batches to support logical rollback points:

```go
// Example batch execution
batch1 := []contractsschema.MigrationInterface{
    &CreateUsersTable{},
    &CreatePostsTable{},
}

batch2 := []contractsschema.MigrationInterface{
    &AddPostsIndexes{},
    &AddCommentsTable{},
}
```

### Batch Operations

```go
// Run migrations (creates new batch)
manager.Run(migrations) // All pending migrations run in single batch

// Rollback last batch
manager.Rollback(0, 0) // Rolls back most recent batch

// Rollback specific batch
manager.Rollback(0, 5) // Rolls back batch #5

// Rollback specific number of migrations
manager.Rollback(3, 0) // Rolls back last 3 migrations
```

## Integration with Interface-Based Migrations

The new tracking system will integrate with the existing interface-based registration. The `MigrationInterface` already includes the `Signature()` and `Description()` methods needed for tracking.

```go
// Register interface-based migrations
migrations := []contractsschema.MigrationInterface{
    &CreateUsersTable{},
    &CreatePostsTable{},
}

db.Schema().Register(migrations)

// Run migrations (tracking to be implemented)
for _, migration := range migrations {
    if err := migration.Up(); err != nil {
        log.Fatal(err)
    }
}
```

### Signature-Based Tracking

The migration tracking will use the `Signature()` method from the interface as the `id` and the `Description()` method for the `description` field in the migration_tracker table.

## Status Monitoring

### Migration Status

```go
type Status struct {
    ID          string
    Description string
    Ran         bool
    Batch       int
    StartedAt   *time.Time
    CompletedAt *time.Time
}

// Get migration status
status, err := manager.Status()
for _, s := range status {
    if s.Ran {
        duration := s.CompletedAt.Sub(*s.StartedAt)
        fmt.Printf("✓ %s: %s (batch %d, %v)\n", s.ID, s.Description, s.Batch, duration)
    } else {
        fmt.Printf("○ %s: %s (pending)\n", s.ID, s.Description)
    }
}
```

### Status Output Example

```
✓ create_users_table: Creates users table with authentication fields (batch 20240615120000, 125ms)
✓ create_posts_table: Creates posts table with user relationships (batch 20240615120000, 89ms)
○ add_posts_indexes: Adds indexes to posts table for performance (pending)
○ add_comments_table: Creates comments table with moderation (pending)
```

## Migration Manager Service

### MigrationManager Interface

```go
package schema

import contractsschema "github.com/dracory/neat/contracts/database/schema"

// MigrationManager handles execution and tracking of interface-based migrations
type MigrationManager interface {
    // Run executes pending migrations
    Run(migrations []contractsschema.MigrationInterface) error
    
    // Rollback reverts migrations
    Rollback(step, batch int) error
    
    // Status returns migration status
    Status() ([]Status, error)
    
    // Fresh drops all tables and re-runs migrations
    Fresh() error
    
    // Reset rolls back and re-runs all migrations
    Reset() error
}
```

### MigrationManager Implementation

The manager should:

1. **Accept Interface-Based Migrations**: Work directly with `MigrationInterface` slices
2. **Automatic Tracking**: Log migrations using their signatures and descriptions
3. **Batch Management**: Group migrations into logical batches
4. **Error Handling**: Provide clear error messages and rollback on failure
5. **Schema Integration**: Use the schema from registered migrations
6. **Table Management**: Handle migration_tracker table operations

## Security Considerations

### SQL Injection Prevention

The new repository should implement security measures:

1. **Table Name Validation**: Validate migration table names to prevent SQL injection
2. **Parameterized Queries**: Use parameterized queries for all database operations
3. **Input Validation**: Validate migration signatures before storage

### Validation Example

```go
// Table name validation
func isValidTableName(name string) bool {
    // Check for SQL injection patterns
    // Ensure only alphanumeric characters and underscores
    // Prevent SQL keywords as table names
}
```

## Database Compatibility

### Supported Databases

The new tracking system should support:

- **SQLite**: For development and testing
- **MySQL**: For production applications
- **PostgreSQL**: For production applications
- **SQL Server**: For enterprise applications
- **Oracle**: For enterprise applications

### Database-Specific Handling

The schema builder automatically handles database-specific SQL generation, so the migration_tracker table creation will work across all supported databases without manual SQL handling.

## Implementation Status

- ✅ MigrationInterface enhanced with Description() method (completed in Part 1)
- ✅ BaseMigration struct with Description() implementation (completed in Part 1)
- ✅ Migration signature validation logic copied to schema package
- ✅ MigrationTracker table structure defined
- ✅ CreateMigrationTrackerTable migration implemented in examples
- ✅ Migration tracking table schema specification
- ⏳ MigrationManager service implementation
- ⏳ Batch management for interface-based migrations
- ⏳ Status monitoring for interface-based migrations
- ⏳ Fresh and reset operations for interface-based migrations
- ⏳ Integration with migration_tracker table operations

## Usage Examples

### Basic Migration Tracking

```go
// Register interface-based migrations
migrations := []contractsschema.MigrationInterface{
    &CreateUsersTable{},
    &CreatePostsTable{},
}

db.Schema().Register(migrations)

// Run migrations with tracking
manager := NewMigrationManager(db.Schema())
err := manager.Run(migrations)
if err != nil {
    log.Fatal(err)
}

// Check status
status, err := manager.Status()
for _, s := range status {
    if s.Ran {
        fmt.Printf("✓ %s (batch %d)\n", s.Name, s.Batch)
    } else {
        fmt.Printf("○ %s (pending)\n", s.Name)
    }
}
```

### Rollback Operations

```go
// Rollback last migration
err := manager.Rollback(1, 0)

// Rollback entire last batch
err := manager.Rollback(0, 0)

// Rollback specific batch
err := manager.Rollback(0, 3)
```

### Fresh Start

```go
// Complete database reset
err := manager.Fresh()
if err != nil {
    log.Fatal(err)
}
```

## Benefits

1. **Interface-Based Design**: Works seamlessly with `MigrationInterface` from Part 1
2. **No File Dependencies**: Eliminates file system scanning
3. **Type-Safe**: Compile-time checking of migration signatures
4. **Explicit Registration**: Clear migration management
5. **Reliable Tracking**: Prevents duplicate migration execution
6. **Batch Management**: Logical grouping for rollback operations
7. **Audit Trail**: Complete history of migration executions
8. **Status Monitoring**: Easy to check migration state
9. **Database Agnostic**: Works across multiple database systems
10. **Secure**: Built-in SQL injection prevention

## Next Steps

1. **Implement MigrationManager service** for interface-based migrations
2. **Add migration locking** to prevent concurrent execution
3. **Implement migration dependencies** between interface-based migrations
4. **Add dry-run mode** for testing interface-based migrations
5. **Enhance error reporting** and recovery for interface-based migrations
6. **Add migration validation** before execution for interface-based migrations
7. **Plan deprecation timeline** for the old `database/migration` package

## Related Documents

- See [migrations-part-1.md](./migrations-part-1.md) for the interface-based migration system
- See [migrations.md](./migrations.md) for the complete proposal overview
