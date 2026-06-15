# Enhanced Schema Migration Interface

**Date**: June 15, 2026
**Status**: Open for Discussion
**Priority**: High
**Author**: Neat ORM Team

## Overview

This proposal proposes enhancing the existing schema Migration interface to make it practical and competitive with the current function-based migration system. The design leverages the existing `schema.Migration` interface with a `BaseMigration` pattern for clean, type-safe migrations inspired by Goravel's implementation.

## Motivation

### Current System Limitations

The current migration system has several architectural limitations:

1. **Function-Based Design**: Uses functions instead of interfaces, limiting extensibility and testability
2. **Global Registration Pattern**: Migrations must be registered in a global registry, creating coupling and initialization order dependencies
3. **No Context Support**: Cannot handle cancellation, timeouts, or request-scoped values
4. **Schema Builder Only**: Migrations are limited to schema builder operations, cannot execute raw SQL
5. **No Self-Contained IDs**: Migration IDs are derived from registration, not intrinsic to the migration object

### Existing Schema Migration Interface

The codebase already has a `schema.Migration` interface that is currently unused:

```go
type Migration interface {
    Signature() string
    Up() error
    Down() error
}
```

This interface was likely inspired by Goravel's migration system but was abandoned because it lacked a practical way to access the schema.

### Design Philosophy

The enhanced system embraces these principles:

1. **Leverage Existing Interface**: Use the existing `schema.Migration` interface
2. **BaseMigration Pattern**: Provide a base struct with schema management
3. **Automatic Schema Injection**: Schema is set during registration
4. **Interface-Based Design**: Better testability and extensibility
5. **Minimal Boilerplate**: Clean migration implementation
6. **Backward Compatible**: Can coexist with current system during transition

## Proposed Architecture

### Enhanced Migration Interface

```go
package schema

// Migration defines the contract for a single migration
type Migration interface {
    // Signature returns the unique identifier for this migration
    Signature() string
    
    // Up applies the migration
    Up() error
    
    // Down rolls back the migration
    Down() error
}

// SchemaSetter is an optional interface for migrations that need schema access
type SchemaSetter interface {
    SetSchema(schema Schema)
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

func (r *Schema) Register(migrations []Migration) {
    for _, migration := range migrations {
        if setter, ok := migration.(SchemaSetter); ok {
            setter.SetSchema(r)
        }
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
    migrations := []contractsschema.Migration{
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

The current system has comprehensive migration tracking (database table, batches, metadata). This would need to be integrated with the new interface-based approach. Options:

1. **Add tracking to Schema**: Extend Schema interface with migration tracking methods
2. **Separate Migrator**: Keep current migrator for tracking, use schema interface for migrations
3. **Hybrid Approach**: Use schema interface for migration definitions, current migrator for execution

## Implementation Plan

### Phase 1: Core Infrastructure
1. Add `SchemaSetter` interface to schema contracts
2. Implement `BaseMigration` struct
3. Update `Schema.Register()` to inject schema
4. Add migration tracking methods to Schema interface

### Phase 2: Migration System Integration
1. Create migrator that works with schema Migration interface
2. Add batch management
3. Add metadata tracking (timestamps, duration)
4. Add rollback support

### Phase 3: Advanced Features
1. Add context support
2. Add transaction support
3. Add raw SQL support
4. Add migration status commands

### Phase 4: Migration and Documentation
1. Update examples to use new approach
2. Create migration guide from current system
3. Update documentation
4. Deprecate old function-based approach

## Open Questions

1. **Migration Tracking**: Should tracking be integrated into Schema or kept separate?
2. **Context Support**: Should we add context to Up/Down methods?
3. **Transaction Support**: How should transactions be handled?
4. **Backward Compatibility**: How long should we support the old system?
5. **Migration Path**: How to help users migrate from old to new system?

## Conclusion

This enhanced schema Migration interface approach provides a clean, practical alternative to the current function-based system. By leveraging the existing interface with a BaseMigration pattern, we can achieve better design while maintaining simplicity. The approach is inspired by Goravel's proven pattern and addresses the limitations of the current system.
