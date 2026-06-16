# migrator Migrations (New Package)

This example demonstrates the new `database/migrator` package, which provides a cleaner API for managing database schema migrations with automatic schema injection and tracking.

## Features Demonstrated

### New migrator Package
- **Simplified API**: No need to manually register migrations with schema
- **Automatic Schema Injection**: Schema is automatically injected into each migration
- **Context Support**: All operations support context for cancellation and timeout handling
- **Clean Interface Naming**: Uses `MigratorInterface` and `Migrator` pattern
- **Flexible Migration Addition**: Add migrations individually or in batches
- **Migration Tracking**: Automatically tracks executed migrations in `migration_tracker` table
- **Advanced Rollback**: Support for rolling back by steps or by batch

### Migration Operations
- Creating tables with various column types
- Adding indexes to existing tables
- Adding columns to existing tables
- Rolling back migrations (single, by steps, or by batch)
- Migration status checking
- Fresh and reset operations

## Advantages Over Old System

1. **No Manual Registration**: No need to call `schema.Register()`
2. **Simpler API**: Pass neat db instance directly instead of separate schema and orm
3. **Auto-injection**: Schema is automatically injected into migrations
4. **Context Support**: All operations support context for proper cancellation/timeout handling
5. **Clear Rollback Methods**: `RollbackSteps()` and `RollbackToBatch()` instead of confusing parameters
6. **Auto-creation**: Automatically creates `migration_tracker` table on first run
7. **Better Organization**: Clear separation between schema building and migration execution

## Running the Example

```bash
cd examples/migrator-migrations
go run main.go
```

This will:
1. Create a SQLite database (`example_schema_migrations.db`)
2. Run all migrations using the new migrator package
3. Demonstrate rolling back the last migration
4. Show migration status

## Migration Structure

Each migration follows the same pattern as before, but now uses the migrator package for execution:

```go
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
        blueprint.Unique("email")
        blueprint.Timestamps()
    })
}

func (m *CreateUsersTable) Down() error {
    return m.GetSchema().DropIfExists("users")
}
```

## Usage Pattern

```go
// Create migrator instance with neat db
migrator := migrator.NewMigrator(db)

// Add migrations
migrator.AddMigration(&CreateUsersTable{})
migrator.AddMigration(&CreatePostsTable{})

// Or add multiple at once
migrator.AddMigrations([]migrator.MigrationInterface{
    &CreateUsersTable{},
    &CreatePostsTable{},
})

// Run migrations
ctx := context.Background()
if err := migrator.Up(ctx); err != nil {
    log.Fatal(err)
}

// Rollback last migration
if err := migrator.Down(ctx); err != nil {
    log.Fatal(err)
}

// Check status
status, err := migrator.Status()
```

## API Methods

### AddMigration
Adds a single migration to the migrator instance.

### AddMigrations
Adds multiple migrations at once.

### Up
Runs all pending migrations. Automatically creates the `migration_tracker` table if it doesn't exist.

### Down
Rolls back the last migration.

### RollbackSteps
Rolls back the specified number of migrations.

### RollbackToBatch
Rolls back all migrations to the specified batch.

### Status
Returns the current status of all migrations.

### Fresh
Drops all tables except `migration_tracker` and clears the tracker.

### Reset
Rolls back all migrations.

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

## Migration from Old System

**Before (Old System):**
```go
schema := db.Schema()
schema.Register(migrations)
manager := schema.NewMigrationManager(db)
manager.Run(migrations)
```

**After (New migrator Package):**
```go
migrator := migrator.NewMigrator(db)
migrator.AddMigrations(migrations)
migrator.Up(context.Background())
```

> **Note**: `schema.Register()`, `schema.Migrations()`, and `schema.NewMigrationManager()` have been removed. Use the `migrator` package as shown above.

## Testing

Run the tests with:

```bash
cd examples/migrator-migrations
go test -v
```

## Prerequisites

- SQLite database (or modify the DSN to use your preferred database)
- Neat ORM with migrator package support

## Related Documentation

- [migrator Package README](../../database/migrator/README.md)
- [migrator Package Proposal](../../docs/proposals/migrator-package.md)
