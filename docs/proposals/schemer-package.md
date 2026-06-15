# Schemer Package Proposal

**Date**: June 15, 2026
**Status**: Proposed
**Priority**: Medium
**Author**: Neat ORM Team

## Overview

This proposal suggests moving the MigrationManager from the `database/schema` package to a new dedicated `database/schemer` package. This change will improve separation of concerns, reduce package coupling, and provide a cleaner API for migration management.

## Motivation

### Current Issues

1. **Package Coupling**: MigrationManager in `database/schema` creates circular dependencies between schema operations and migration management
2. **Registration Complexity**: Users must call `schema.Register(migrations)` before using MigrationManager, which is an extra step
3. **Package Naming**: The `schema` package is focused on schema building, not migration execution
4. **User Confusion**: Having migration management in the schema package is conceptually confusing

### Benefits of Schemer Package

1. **Clear Separation**: Schema building vs migration execution are separate concerns
2. **Simpler API**: Pass schema as dependency instead of requiring registration
3. **Better Organization**: Each package has a single, clear responsibility
4. **User-Friendly**: Less boilerplate and clearer usage patterns
5. **Extensibility**: Easier to add migration-specific features without affecting schema package

## Proposed Architecture

### Package Structure

```
database/
├── schema/           # Schema building (blueprints, grammars, etc.)
├── schemer/          # Migration management (NEW)
│   ├── schemer.go    # MigrationManager and related types
│   └── tracker.go    # MigrationTracker and Status types
└── migration/        # Deprecated file-based system
```

### New API

```go
package schemer

import (
    contractsschema "github.com/dracory/neat/contracts/database/schema"
    contractsorm "github.com/dracory/neat/contracts/database/orm"
)

// MigrationManager handles execution and tracking of interface-based migrations
type MigrationManager struct {
    schema contractsschema.Schema
    orm    contractsorm.Orm
}

// NewMigrationManager creates a new MigrationManager instance
// Takes schema and orm as dependencies, no registration needed
func NewMigrationManager(schema contractsschema.Schema, orm contractsorm.Orm) *MigrationManager {
    return &MigrationManager{
        schema: schema,
        orm:    orm,
    }
}

// Run executes pending migrations
// Automatically injects schema into each migration before execution
func (m *MigrationManager) Run(migrations []contractsschema.MigrationInterface) error {
    // Inject schema into each migration
    for _, migration := range migrations {
        migration.SetSchema(m.schema)
    }

    // Run migrations with tracking
    // ... existing implementation
}

// Rollback reverts migrations
func (m *MigrationManager) Rollback(step, batch int) error {
    // ... existing implementation
}

// Status returns migration status
func (m *MigrationManager) Status() ([]Status, error) {
    // ... existing implementation
}

// Fresh drops all tables and re-runs migrations
func (m *MigrationManager) Fresh() error {
    // ... existing implementation
}

// Reset rolls back and re-runs all migrations
func (m *MigrationManager) Reset() error {
    // ... existing implementation
}
```

### MigrationTracker and Status Types

```go
package schemer

import "time"

// MigrationTracker represents a migration record from the migration_tracker table
type MigrationTracker struct {
    ID          string    // The migration signature
    Batch       int       // Timestamp ID (YYYYMMDDHHMMSS)
    Description string    // The migration description
    StartedAt   time.Time // When the migration started
    CompletedAt time.Time // When the migration finished
}

// Status represents the status of a migration
type Status struct {
    ID          string    `json:"id"`
    Description string    `json:"description"`
    Batch       int       `json:"batch"`
    StartedAt   time.Time `json:"started_at"`
    CompletedAt time.Time `json:"completed_at"`
    State       string    `json:"state"` // "pending", "completed", "failed"
}
```

## Usage Examples

### Before (Current Approach)

```go
package main

import (
    "github.com/dracory/neat"
    contractsschema "github.com/dracory/neat/contracts/database/schema"
    "github.com/dracory/neat/database/schema"
)

func main() {
    db, _ := neat.NewFromDSN("sqlite://./app.db")
    defer db.Close()

    migrations := []contractsschema.MigrationInterface{
        &CreateUsersTable{},
        &CreatePostsTable{},
    }

    // Step 1: Register with schema
    db.Schema().Register(migrations)

    // Step 2: Create manager
    manager := schema.NewMigrationManager(db.Schema(), db.Orm())

    // Step 3: Run migrations
    if err := manager.Run(migrations); err != nil {
        log.Fatal(err)
    }
}
```

### After (Proposed Approach)

```go
package main

import (
    "github.com/dracory/neat"
    contractsschema "github.com/dracory/neat/contracts/database/schema"
    "github.com/dracory/neat/database/schemer"
)

func main() {
    db, _ := neat.NewFromDSN("sqlite://./app.db")
    defer db.Close()

    migrations := []contractsschema.MigrationInterface{
        &CreateUsersTable{},
        &CreatePostsTable{},
    }

    // Single step: Create manager and run
    manager := schemer.NewMigrationManager(db.Schema(), db.Orm())
    if err := manager.Run(migrations); err != nil {
        log.Fatal(err)
    }
}
```

## Migration Path

### Phase 1: Create Schemer Package

1. Create `database/schemer` package
2. Move MigrationManager from `database/schema` to `database/schemer`
3. Move MigrationTracker and Status types to `database/schemer`
4. Update MigrationManager to auto-inject schema

### Phase 2: Update Examples

1. Update `examples/schema-migrations/main.go` to use new package
2. Update `examples/schema-migrations/main_test.go` to use new package
3. Remove `schema.Register()` calls from examples

### Phase 3: Deprecate Old Approach

1. Keep `schema.NewMigrationManager` for backward compatibility
2. Add deprecation notice to `schema.NewMigrationManager`
3. Update documentation to recommend `schemer` package

### Phase 4: Remove Old Code

1. Remove `schema.NewMigrationManager` after deprecation period
2. Remove schema registration requirement from MigrationInterface
3. Update all internal code to use schemer package

## Impact Analysis

### Breaking Changes

- **Import path change**: Users importing `database/schema` for MigrationManager need to update to `database/schemer`
- **API change**: `schema.Register()` is no longer required (can be kept for backward compatibility initially)

### Non-Breaking Changes

- MigrationInterface remains unchanged
- BaseMigration remains unchanged
- Migration signature format remains unchanged
- MigrationTracker table structure remains unchanged

### Benefits

1. **Simpler API**: One less step for users (no registration)
2. **Clearer Organization**: Better separation of concerns
3. **Easier Testing**: Can test schemer independently of schema package
4. **Better Documentation**: Each package has clear, focused documentation

## Implementation Status

- ✅ Mark database/migrator package as deprecated
- ✅ Mark schema.NewMigrationManager as deprecated
- ✅ Add deprecation notices to migrator files (migrator.go, format.go, repository.go)
- ⏳ Create `database/schemer` package
- ⏳ Move MigrationManager to schemer package
- ⏳ Move MigrationTracker and Status types to schemer package
- ⏳ Update MigrationManager to auto-inject schema
- ⏳ Update examples to use schemer package
- ⏳ Update documentation

## Related Documents

- See [migrations-part-1.md](./migrations-part-1.md) for the interface-based migration system
- See [migrations-part-2.md](./migrations-part-2.md) for the migration tracking system
- See [migrations.md](./migrations.md) for the complete proposal overview
