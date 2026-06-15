# Schemer Package Proposal

**Date**: June 15, 2026
**Status**: Implemented
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
2. **Simpler API**: Pass neat db instance instead of separate schema and orm dependencies
3. **Better Organization**: Each package has a single, clear responsibility
4. **User-Friendly**: Less boilerplate and clearer usage patterns
5. **Extensibility**: Easier to add migration-specific features without affecting schema package

## Proposed Architecture

### Package Structure

```
database/
├── schema/           # Schema building (blueprints, grammars, etc.)
├── schemer/          # Migration management (NEW)
│   ├── schemer.go    # SchemerInterface and SchemerImplementation
│   └── tracker.go    # MigrationTracker and MigrationStatus types
└── migration/        # Deprecated file-based system
```

### New API

```go
package schemer

import (
    "context"
    "github.com/dracory/neat"
    contractsschema "github.com/dracory/neat/contracts/database/schema"
)

// SchemerInterface defines the contract for migration management
type SchemerInterface interface {
    AddMigration(migration contractsschema.MigrationInterface) error
    AddMigrations(migrations []contractsschema.MigrationInterface) error
    Up(ctx context.Context) error
    Down(ctx context.Context) error
    RollbackSteps(ctx context.Context, steps int) error
    RollbackToBatch(ctx context.Context, batch int) error
    Status() ([]MigrationStatus, error)
    Fresh(ctx context.Context) error
    Reset(ctx context.Context) error
}

// SchemerImplementation handles execution and tracking of interface-based migrations
type SchemerImplementation struct {
    db         *neat.Neat
    migrations []contractsschema.MigrationInterface
}

// NewSchemer creates a new SchemerImplementation instance
// Takes neat db instance as dependency, extracts schema and orm internally
func NewSchemer(db *neat.Neat) SchemerInterface {
    return &SchemerImplementation{
        db:         db,
        migrations: []contractsschema.MigrationInterface{},
    }
}

// AddMigration adds a new migration to the list
func (s *SchemerImplementation) AddMigration(migration contractsschema.MigrationInterface) error {
    s.migrations = append(s.migrations, migration)
    return nil
}

// AddMigrations adds multiple migrations to the runner
func (s *SchemerImplementation) AddMigrations(migrations []contractsschema.MigrationInterface) error {
    s.migrations = append(s.migrations, migrations...)
    return nil
}

// Up runs all pending migrations
// Automatically injects schema into each migration before execution
func (s *SchemerImplementation) Up(ctx context.Context) error {
    // Inject schema into each migration
    for _, migration := range s.migrations {
        migration.SetSchema(s.db.Schema())
    }

    // Run migrations with tracking
    // ... existing implementation
}

// Down rolls back the last migration
func (s *SchemerImplementation) Down(ctx context.Context) error {
    // ... existing implementation
}

// RollbackSteps rolls back the specified number of migrations
func (s *SchemerImplementation) RollbackSteps(ctx context.Context, steps int) error {
    // ... existing implementation
}

// RollbackToBatch rolls back all migrations to the specified batch
func (s *SchemerImplementation) RollbackToBatch(ctx context.Context, batch int) error {
    // ... existing implementation
}

// Status returns migration status
func (s *SchemerImplementation) Status() ([]MigrationStatus, error) {
    // ... existing implementation
}

// Fresh drops all tables and re-runs migrations
func (s *SchemerImplementation) Fresh(ctx context.Context) error {
    // ... existing implementation
}

// Reset rolls back and re-runs all migrations
func (s *SchemerImplementation) Reset(ctx context.Context) error {
    // ... existing implementation
}
```

### MigrationTracker and Status Types

```go
package schemer

import "time"

// MigrationTracker represents a migration record stored in the migration_tracker table
// This is the database model/entity used for persistence
type MigrationTracker struct {
    ID          string    // The migration signature
    Batch       int       // Timestamp ID (YYYYMMDDHHMMSS)
    Description string    // The migration description
    StartedAt   time.Time // When the migration started
    CompletedAt time.Time // When the migration finished
}

// MigrationStatus represents the status of a migration returned to users
// This is a DTO/response type derived from MigrationTracker data
type MigrationStatus struct {
    ID          string    `json:"id"`
    Description string    `json:"description"`
    Batch       int       `json:"batch"`
    StartedAt   time.Time `json:"started_at"`
    CompletedAt time.Time `json:"completed_at"`
    State       string    `json:"state"` // "pending", "completed", "failed"
}
```

**Persistence:**
- MigrationTracker records are stored in a `migration_tracker` table in the database
- The SchemerImplementation automatically creates this table if it doesn't exist
- MigrationStatus is computed by comparing registered migrations against MigrationTracker records

## Usage Examples

### Before (Current Approach)

```go
package main

import (
    "context"
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
    manager := schema.NewMigrationManager(db)

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
    "context"
    "github.com/dracory/neat"
    contractsschema "github.com/dracory/neat/contracts/database/schema"
    "github.com/dracory/neat/database/schemer"
)

func main() {
    db, _ := neat.NewFromDSN("sqlite://./app.db")
    defer db.Close()

    // Create schemer with neat db instance
    schemer := schemer.NewSchemer(db)

    // Add migrations
    schemer.AddMigration(&CreateUsersTable{})
    schemer.AddMigration(&CreatePostsTable{})

    // Or add multiple at once
    // schemer.AddMigrations([]contractsschema.MigrationInterface{
    //     &CreateUsersTable{},
    //     &CreatePostsTable{},
    // })

    // Run migrations
    if err := schemer.Up(context.Background()); err != nil {
        log.Fatal(err)
    }

    // Rollback last 3 migrations
    if err := schemer.RollbackSteps(context.Background(), 3); err != nil {
        log.Fatal(err)
    }

    // Or rollback to specific batch
    // if err := schemer.RollbackToBatch(context.Background(), 20240115120000); err != nil {
    //     log.Fatal(err)
    // }

    // Fresh: drop all tables and re-run migrations
    // if err := schemer.Fresh(context.Background()); err != nil {
    //     log.Fatal(err)
    // }

    // Reset: rollback and re-run all migrations
    // if err := schemer.Reset(context.Background()); err != nil {
    //     log.Fatal(err)
    // }
}
```

## Migration Path

### Phase 1: Create Schemer Package ✅

1. Create `database/schemer` package ✅
2. Create SchemerInterface defining the migration management contract ✅
3. Implement SchemerImplementation (extract logic from MigrationManager) ✅
4. Move MigrationTracker and MigrationStatus types to `database/schemer` ✅
5. Update SchemerImplementation to accept neat db instance and auto-inject schema ✅
6. Add auto-creation of migration_tracker table ✅

### Phase 2: Update Examples ✅

1. Update `examples/schemer-migrations/main.go` to use new package ✅
2. Update `examples/schemer-migrations/main_test.go` to use new package ✅
3. Remove `schema.Register()` calls from examples ✅
4. Update migration calls from `Run()` to `Up()` ✅
5. Add new schemer-specific tests ✅

### Phase 3: Deprecate Old Approach

1. Keep `schema.NewMigrationManager` for backward compatibility
2. Add deprecation notice to `schema.NewMigrationManager`
3. Update documentation to recommend `schemer` package

### Phase 4: Documentation ✅

1. Create comprehensive README.md for schemer package ✅
2. Add API reference and usage examples ✅
3. Document migration from old system ✅
4. Add best practices section ✅

### Phase 5: Remove Old Code (Future)

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
- ✅ Create `database/schemer` package
- ✅ Create SchemerInterface defining migration management contract
- ✅ Implement SchemerImplementation with migration management logic
- ✅ Move MigrationTracker and MigrationStatus types to schemer package
- ✅ Update SchemerImplementation to accept neat db instance and auto-inject schema
- ✅ Add comprehensive tests for schemer package
- ✅ Update examples to use schemer package
- ✅ Update documentation

## Related Documents

- See [migrations-part-1.md](./migrations-part-1.md) for the interface-based migration system
- See [migrations-part-2.md](./migrations-part-2.md) for the migration tracking system
- See [migrations.md](./migrations.md) for the complete proposal overview
