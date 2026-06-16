# migrator Package

The `migrator` package provides a clean API for managing database migrations with automatic schema injection and tracking.

## Overview

The migrator package offers a simplified migration management system that:

- **Automatic Schema Injection**: No need to manually register migrations with schema
- **Clean Interface-based API**: Uses `MigratorInterface` and `Migrator` pattern
- **Context Support**: All operations support context for cancellation and timeout handling
- **Migration Tracking**: Automatically tracks executed migrations in a `migration_tracker` table
- **Flexible Rollback**: Support for rolling back by steps or by batch

## Installation

```go
import "github.com/dracory/neat/database/migrator"
```

## Quick Start

```go
package main

import (
    "context"
    "github.com/dracory/neat"
    contractsschema "github.com/dracory/neat/contracts/database/schema"
    "github.com/dracory/neat/database/migrator"
)

func main() {
    db, _ := neat.NewFromDSN("sqlite://./app.db")
    defer db.Close()

    // Create migrator instance
    migrator := migrator.NewMigrator(db)

    // Add migrations
    migrator.AddMigration(&CreateUsersTable{})
    migrator.AddMigration(&CreatePostsTable{})

    // Run migrations
    ctx := context.Background()
    if err := migrator.Up(ctx); err != nil {
        log.Fatal(err)
    }
}
```

## API Reference

### MigratorInterface

```go
type MigratorInterface interface {
    AddMigration(migration migrator.MigrationInterface) error
    AddMigrations(migrations []migrator.MigrationInterface) error
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
Adds a single migration to the migrator instance.

```go
migrator.AddMigration(&CreateUsersTable{})
```

#### AddMigrations
Adds multiple migrations at once.

```go
migrator.AddMigrations([]migrator.MigrationInterface{
    &CreateUsersTable{},
    &CreatePostsTable{},
})
```

#### Up
Runs all pending migrations. Automatically creates the `migration_tracker` table if it doesn't exist.

```go
ctx := context.Background()
err := migrator.Up(ctx)
```

#### Down
Rolls back the last migration.

```go
ctx := context.Background()
err := migrator.Down(ctx)
```

#### RollbackSteps
Rolls back the specified number of migrations.

```go
ctx := context.Background()
err := migrator.RollbackSteps(ctx, 3) // Rollback last 3 migrations
```

#### RollbackToBatch
Rolls back all migrations to the specified batch.

```go
ctx := context.Background()
err := migrator.RollbackToBatch(ctx, 20240615120000)
```

#### Status
Returns the current status of all migrations.

```go
status, err := migrator.Status()
for _, s := range status {
    fmt.Printf("Migration: %s - State: %s\n", s.ID, s.State)
}
```

#### Fresh
Drops all tables except `migration_tracker` and clears the tracker.

```go
ctx := context.Background()
err := migrator.Fresh(ctx)
```

#### Reset
Rolls back all migrations.

```go
ctx := context.Background()
err := migrator.Reset(ctx)
```

#### SetTransactionsEnabled
Enables or disables transaction wrapping for migration operations. Transactions are enabled by default for safety.

```go
migrator.SetTransactionsEnabled(true)  // Enable transactions (default)
migrator.SetTransactionsEnabled(false) // Disable transactions for large migrations
```

#### SetTransactionIsolationLevel
Sets the transaction isolation level for migration operations.

```go
migrator.SetTransactionIsolationLevel("SERIALIZABLE")
migrator.SetTransactionIsolationLevel("READ COMMITTED")
```

Supported isolation levels:
- `READ UNCOMMITTED`
- `READ COMMITTED`
- `REPEATABLE READ`
- `SERIALIZABLE`
- `SNAPSHOT`

## Migration Implementation

Migrations must implement the `MigrationInterface` from the migrator package:

```go
import (
    contractsschema "github.com/dracory/neat/contracts/database/schema"
    "github.com/dracory/neat/database/schema"
)

type CreateUsersTable struct {
    migrator.BaseMigration
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

The migrator automatically tracks migrations in a `migration_tracker` table with the following structure:

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

The migrator package supports transaction wrapping for safe migration execution. Transactions are enabled by default to ensure atomic execution.

### Enabling/Disabling Transactions

```go
migrator := migrator.NewMigrator(db)

// Transactions are enabled by default
migrator.SetTransactionsEnabled(true)

// Disable for large migrations or specific needs
migrator.SetTransactionsEnabled(false)
```

### Transaction Isolation Levels

```go
migrator.SetTransactionIsolationLevel("SERIALIZABLE")
```

Supported isolation levels:
- `READ UNCOMMITTED`
- `READ COMMITTED`
- `REPEATABLE READ`
- `SERIALIZABLE`
- `SNAPSHOT`

### Note on Current Implementation

Transaction wrapping is currently disabled by default pending verification of schema transaction detection. The infrastructure is in place and can be enabled once schema transaction behavior is properly tested.

See [examples/migrator-transactions](../../examples/migrator-transactions/) for a complete example of transaction control usage.

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
   err := migrator.Up(ctx)
   ```

4. **Error Handling**: Handle migration errors appropriately
   ```go
   if err := migrator.Up(ctx); err != nil {
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
migrator := migrator.NewMigrator(db)
migrator.AddMigrations(migrations)
migrator.Up(context.Background())
```

> **Note**: `schema.Register()`, `schema.Migrations()`, and `schema.NewMigrationManager()` have been removed. Use the `migrator` package as shown above.

## Examples

See the `examples/migrator-migrations` directory for complete examples of using the migrator package.

## Testing

The migrator package includes comprehensive tests. Run them with:

```bash
go test ./database/migrator/...
```

## Notes

- The migrator automatically creates the `migration_tracker` table on first run
- Migrations are executed in the order they are added
- Already-run migrations are automatically skipped
- Schema is automatically injected into each migration before execution
