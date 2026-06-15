# Enhanced Schema Migration Interface

**Date**: June 15, 2026
**Status**: Implemented
**Priority**: High
**Author**: Neat ORM Team

## Overview

This proposal enhances the existing schema Migration interface to make it practical and competitive with the current function-based migration system. The design leverages the existing `schema.MigrationInterface` with a `BaseMigration` pattern for clean, type-safe migrations inspired by Goravel's implementation.

## Motivation

### Current System Limitations

The current migration system has several architectural limitations:

1. **Function-Based Design**: Uses functions instead of interfaces, limiting extensibility and testability
2. **Global Registration Pattern**: Migrations must be registered in a global registry, creating coupling and initialization order dependencies
3. **No Context Support**: Cannot handle cancellation, timeouts, or request-scoped values
4. **Schema Builder Only**: Migrations are limited to schema builder operations, cannot execute raw SQL
5. **No Self-Contained IDs**: Migration IDs are derived from registration, not intrinsic to the migration object

### Existing Schema Migration Interface

The codebase has a `schema.MigrationInterface` that has been enhanced with schema management:

```go
type MigrationInterface interface {
    // Signature Get the migration signature.
    Signature() string
    // Up Run the migrations.
    Up() error
    // Down Reverse the migrations.
    Down() error
    // SetSchema sets the schema for this migration
    SetSchema(schema Schema)
    // GetSchema returns the schema for this migration
    GetSchema() Schema
}
```

This interface was inspired by Goravel's migration system and has been enhanced with direct schema management methods, eliminating the need for optional interfaces.

### Design Philosophy

The enhanced system embraces these principles:

1. **Leverage Existing Interface**: Use the existing `schema.MigrationInterface`
2. **BaseMigration Pattern**: Provide a base struct with schema management
3. **Automatic Schema Injection**: Schema is set during registration
4. **Interface-Based Design**: Better testability and extensibility
5. **Minimal Boilerplate**: Clean migration implementation
6. **Backward Compatible**: Can coexist with current system during transition

## Proposed Architecture

### Enhanced Migration Interface

```go
package schema

// MigrationInterface defines the contract for a single migration
type MigrationInterface interface {
    // Signature Get the migration signature.
    Signature() string
    // Up Run the migrations.
    Up() error
    // Down Reverse the migrations.
    Down() error
    // SetSchema sets the schema for this migration
    SetSchema(schema Schema)
    // GetSchema returns the schema for this migration
    GetSchema() Schema
}
```

### BaseMigration Implementation

```go
package schema

// BaseMigration provides common functionality for all migrations
type BaseMigration struct {
    schema Schema
}

// SetSchema sets the schema for this migration
func (b *BaseMigration) SetSchema(schema Schema) {
    b.schema = schema
}

// GetSchema returns the schema for this migration
func (b *BaseMigration) GetSchema() Schema {
    return b.schema
}
```

### Enhanced Register Method

```go
package schema

// Register migrations.
func (r *Schema) Register(migrations []MigrationInterface) {
    for _, migration := range migrations {
        migration.SetSchema(r)
    }
    r.migrations = migrations
}
```

## Migration Implementation

### Basic Migration

```go
package migrations

import (
    contractsschema "github.com/dracory/neat/contracts/database/schema"
)

// CreateUsersTable creates the users table
type CreateUsersTable struct {
    contractsschema.BaseMigration
}

func (m *CreateUsersTable) Signature() string {
    return "create_users_table"
}

func (m *CreateUsersTable) Up() error {
    return m.GetSchema().Create("users", func(blueprint contractsschema.Blueprint) {
        blueprint.ID()
        blueprint.String("name")
        blueprint.String("email")
        blueprint.Unique("email")
        blueprint.String("password")
        blueprint.Timestamps()
        blueprint.SoftDeletes()
    })
}

func (m *CreateUsersTable) Down() error {
    return m.GetSchema().DropIfExists("users")
}
```

### Migration with Indexes

```go
package migrations

// AddPostsIndexes adds indexes to the posts table
type AddPostsIndexes struct {
    contractsschema.BaseMigration
}

func (m *AddPostsIndexes) Signature() string {
    return "add_posts_indexes"
}

func (m *AddPostsIndexes) Up() error {
    return m.GetSchema().Table("posts", func(blueprint contractsschema.Blueprint) {
        blueprint.Index("user_id")
        blueprint.Index("status")
    })
}

func (m *AddPostsIndexes) Down() error {
    // Note: Dropping indexes is database-specific
    return nil
}
```

## Usage Examples

### Basic Usage

```go
package main

import (
    "log"
    
    "github.com/dracory/neat"
    "github.com/dracory/neat/migrations"
)

func main() {
    db, err := neat.NewFromDSN("sqlite://./app.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Create migrations
    migrations := []contractsschema.MigrationInterface{
        &migrations.CreateUsersTable{},
        &migrations.CreatePostsTable{},
        &migrations.AddPostsIndexes{},
    }
    
    // Register migrations (schema is automatically injected)
    db.Schema().Register(migrations)
    
    // Run migrations
    for _, migration := range db.Schema().Migrations() {
        log.Printf("Running migration: %s", migration.Signature())
        if err := migration.Up(); err != nil {
            log.Fatalf("Migration failed: %v", err)
        }
    }
}
```

### With Rollback

```go
// Rollback in reverse order
migrationsList := db.Schema().Migrations()
for i := len(migrationsList) - 1; i >= 0; i-- {
    migration := migrationsList[i]
    log.Printf("Rolling back migration: %s", migration.Signature())
    if err := migration.Down(); err != nil {
        log.Fatalf("Rollback failed: %v", err)
    }
}
```

## Comparison with Current System

### Current Function-Based Approach

```go
migrator.RegisterMigration("001_create_users_table", migrator.Migration{
    Up: func(schema contractsschema.Schema) error {
        return schema.Create("users", func(blueprint contractsschema.Blueprint) {
            blueprint.ID()
            blueprint.String("name")
            blueprint.String("email")
            blueprint.Unique("email")
            blueprint.Timestamps()
        })
    },
    Down: func(schema contractsschema.Schema) error {
        return schema.DropIfExists("users")
    },
})
```

### Enhanced Interface-Based Approach

```go
type CreateUsersTable struct {
    contractsschema.BaseMigration
}

func (m *CreateUsersTable) Signature() string {
    return "create_users_table"
}

func (m *CreateUsersTable) Up() error {
    return m.GetSchema().Create("users", func(blueprint contractsschema.Blueprint) {
        blueprint.ID()
        blueprint.String("name")
        blueprint.String("email")
        blueprint.Unique("email")
        blueprint.Timestamps()
    })
}

func (m *CreateUsersTable) Down() error {
    return m.GetSchema().DropIfExists("users")
}
```

## Benefits

1. **Interface-Based Design**: Better testability and extensibility
2. **Minimal Boilerplate**: Just embed `BaseMigration`
3. **Type-Safe**: Schema access through methods
4. **Self-Contained**: Migration objects with intrinsic IDs
5. **No Global State**: Explicit registration and management
6. **Consistent Pattern**: All migrations follow the same structure
7. **Leverages Existing Code**: Uses existing schema interface
8. **Goravel-Inspired**: Proven pattern from successful framework

## Migration Tracking

The current system has comprehensive migration tracking (database table, batches, metadata). This has been integrated with the new interface-based approach through the existing `database/migration` package, which supports both function-based and interface-based migrations.

## Implementation Status

- ✅ MigrationInterface enhanced with SetSchema/GetSchema methods
- ✅ BaseMigration struct implemented
- ✅ Register method updated to inject schema
- ✅ Schema interface includes Register method
- ✅ Migration tracking integration (database/migration package)
- ✅ Documentation updates (examples/schema-migrations/README.md)
- ✅ Example migrations (examples/schema-migrations/)
- ✅ Migration runner service (database/migration/migrator.go)

## Next Steps

1. Add comprehensive tests for interface-based migrations
2. Plan migration path from function-based to interface-based
3. Consider deprecation timeline for function-based system
