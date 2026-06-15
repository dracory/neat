# Schemer Package

The `schemer` package provides a clean API for managing database migrations with automatic schema injection and tracking.

## Overview

The schemer package offers a simplified migration management system that:

- **Automatic Schema Injection**: No need to manually register migrations with schema
- **Clean Interface-based API**: Uses `SchemerInterface` and `SchemerImplementation` pattern
- **Context Support**: All operations support context for cancellation and timeout handling
- **Migration Tracking**: Automatically tracks executed migrations in a `migration_tracker` table
- **Flexible Rollback**: Support for rolling back by steps or by batch

## Installation

```go
import "github.com/dracory/neat/database/schemer"
```

## Quick Start

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

    // Create schemer instance
    schemer := schemer.NewSchemer(db)

    // Add migrations
    schemer.AddMigration(&CreateUsersTable{})
    schemer.AddMigration(&CreatePostsTable{})

    // Run migrations
    ctx := context.Background()
    if err := schemer.Up(ctx); err != nil {
        log.Fatal(err)
    }
}
```

## API Reference

### SchemerInterface

```go
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
    SetTransactionsEnabled(enabled bool)
    SetTransactionIsolationLevel(level string)
}
```

### Methods

#### AddMigration
Adds a single migration to the schemer instance.

```go
schemer.AddMigration(&CreateUsersTable{})
```

#### AddMigrations
Adds multiple migrations at once.

```go
schemer.AddMigrations([]contractsschema.MigrationInterface{
    &CreateUsersTable{},
    &CreatePostsTable{},
})
```

#### Up
Runs all pending migrations. Automatically creates the `migration_tracker` table if it doesn't exist.

```go
ctx := context.Background()
err := schemer.Up(ctx)
```

#### Down
Rolls back the last migration.

```go
ctx := context.Background()
err := schemer.Down(ctx)
```

#### RollbackSteps
Rolls back the specified number of migrations.

```go
ctx := context.Background()
err := schemer.RollbackSteps(ctx, 3) // Rollback last 3 migrations
```

#### RollbackToBatch
Rolls back all migrations to the specified batch.

```go
ctx := context.Background()
err := schemer.RollbackToBatch(ctx, 20240615120000)
```

#### Status
Returns the current status of all migrations.

```go
status, err := schemer.Status()
for _, s := range status {
    fmt.Printf("Migration: %s - State: %s\n", s.ID, s.State)
}
```

#### Fresh
Drops all tables except `migration_tracker` and clears the tracker.

```go
ctx := context.Background()
err := schemer.Fresh(ctx)
```

#### Reset
Rolls back all migrations.

```go
ctx := context.Background()
err := schemer.Reset(ctx)
```

#### SetTransactionsEnabled
Enables or disables transaction wrapping for migration operations. Transactions are enabled by default for safety.

```go
schemer.SetTransactionsEnabled(true)  // Enable transactions (default)
schemer.SetTransactionsEnabled(false) // Disable transactions for large migrations
```

#### SetTransactionIsolationLevel
Sets the transaction isolation level for migration operations.

```go
schemer.SetTransactionIsolationLevel("SERIALIZABLE")
schemer.SetTransactionIsolationLevel("READ COMMITTED")
```

Supported isolation levels:
- `READ UNCOMMITTED`
- `READ COMMITTED`
- `REPEATABLE READ`
- `SERIALIZABLE`
- `SNAPSHOT`

## Migration Implementation

Migrations must implement the `MigrationInterface` from the contracts package:

```go
import (
    contractsschema "github.com/dracory/neat/contracts/database/schema"
    "github.com/dracory/neat/database/schema"
)

type CreateUsersTable struct {
    schema.BaseMigration
}

func (m *CreateUsersTable) Signature() string {
    return "2024_06_15_120000_create_users_table"
}

func (m *CreateUsersTable) Description() string {
    return "Creates users table"
}

func (m *CreateUsersTable) Up() error {
    return m.GetSchema().Create("users", func(blueprint contractsschema.Blueprint) {
        blueprint.ID()
        blueprint.String("name")
        blueprint.String("email")
        blueprint.Timestamps()
    })
}

func (m *CreateUsersTable) Down() error {
    return m.GetSchema().DropIfExists("users")
}
```

## Migration Tracking

The schemer automatically tracks migrations in a `migration_tracker` table with the following structure:

```go
type MigrationTracker struct {
    ID          string    // Migration signature
    Batch       int       // Batch number (timestamp)
    Description string    // Migration description
    StartedAt   time.Time // When migration started
    CompletedAt time.Time // When migration finished
}
```

## Transaction Support

The schemer package supports transaction wrapping for safe migration execution. Transactions are enabled by default to ensure atomic execution.

### Enabling/Disabling Transactions

```go
schemer := schemer.NewSchemer(db)

// Transactions are enabled by default
schemer.SetTransactionsEnabled(true)

// Disable for large migrations or specific needs
schemer.SetTransactionsEnabled(false)
```

### Transaction Isolation Levels

```go
schemer.SetTransactionIsolationLevel("SERIALIZABLE")
```

Supported isolation levels:
- `READ UNCOMMITTED`
- `READ COMMITTED`
- `REPEATABLE READ`
- `SERIALIZABLE`
- `SNAPSHOT`

### Note on Current Implementation

Transaction wrapping is currently disabled by default pending verification of schema transaction detection. The infrastructure is in place and can be enabled once schema transaction behavior is properly tested.

See [examples/schemer-transactions](../../examples/schemer-transactions/) for a complete example of transaction control usage.

## Migration Status

The `Status()` method returns `MigrationStatus` objects:

```go
type MigrationStatus struct {
    ID          string    `json:"id"`
    Description string    `json:"description"`
    Batch       int       `json:"batch"`
    StartedAt   time.Time `json:"started_at"`
    CompletedAt time.Time `json:"completed_at"`
    State       string    `json:"state"` // "pending", "completed", "failed"
}
```

## Best Practices

1. **Migration Naming**: Use timestamp-based signatures for ordering
   ```go
   "2024_06_15_120000_create_users_table"
   ```

2. **Idempotent Up Methods**: Check if resources exist before creating
   ```go
   func (m *CreateUsersTable) Up() error {
       if m.GetSchema().HasTable("users") {
           return nil
       }
       // Create table
   }
   ```

3. **Context Usage**: Always use context for production applications
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
   defer cancel()
   err := schemer.Up(ctx)
   ```

4. **Error Handling**: Handle migration errors appropriately
   ```go
   if err := schemer.Up(ctx); err != nil {
       log.Fatalf("Migration failed: %v", err)
   }
   ```

## Migration from Old System

If you're migrating from the old `schema.NewMigrationManager`:

**Before:**
```go
schema := db.Schema()
schema.Register(migrations)
manager := schema.NewMigrationManager(db)
manager.Run(migrations)
```

**After:**
```go
schemer := schemer.NewSchemer(db)
schemer.AddMigrations(migrations)
schemer.Up(context.Background())
```

## Examples

See the `examples/schemer-migrations` directory for complete examples of using the schemer package.

## Testing

The schemer package includes comprehensive tests. Run them with:

```bash
go test ./database/schemer/...
```

## Notes

- The schemer automatically creates the `migration_tracker` table on first run
- Migrations are executed in the order they are added
- Already-run migrations are automatically skipped
- Schema is automatically injected into each migration before execution
