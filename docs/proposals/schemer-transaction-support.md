# Transaction Support for Schemer Package

**Date**: June 15, 2026
**Status**: Proposed
**Priority**: Medium
**Author**: Neat ORM Team

## Overview

This proposal suggests adding transparent transaction support to the schemer package. Users should not need to think about transactions - the schemer should automatically handle transaction wrapping for safe migration execution.

## Motivation

### Current Issues

1. **No Atomicity**: Current migration execution doesn't use transactions, so if a migration fails partway through, the database can be left in an inconsistent state
2. **Partial Migrations**: If a migration creates multiple tables and fails halfway, some tables exist while others don't
3. **Manual Cleanup**: Users must manually clean up failed migrations, which is error-prone
4. **Data Safety**: No automatic rollback on migration failures

### Benefits of Transparent Transaction Support

1. **Automatic Safety**: Users don't need to think about transactions - it just works
2. **Atomic Execution**: Either all changes in a migration succeed, or none do
3. **Automatic Rollback**: Failed migrations automatically roll back all changes
4. **Data Consistency**: Database is always in a valid state
5. **Safer Deployments**: Reduces risk of production deployment failures

## Proposed Approach: Transparent Transactions

### Key Principle

**Users should not need to choose between transaction and non-transaction methods.** The schemer should automatically handle transaction wrapping based on configuration and database capabilities.

### No New Methods Required

The existing API remains unchanged:

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
}
```

### Usage Example

```go
// User code - no changes needed
schemer := schemer.NewSchemer(db)
schemer.AddMigration(&CreateUsersTable{})
schemer.AddMigration(&CreatePostsTable{})

// Just call Up() - transactions handled automatically
ctx := context.Background()
err := schemer.Up(ctx)
if err != nil {
    // All changes automatically rolled back if transaction enabled
    log.Fatal("Migration failed:", err)
}
```

## Implementation Approach

### Schema Package Already Has Transaction Support

The schema package already has built-in transaction support and automatically detects transaction context:

```go
// Schema automatically detects if in transaction
if query.InTransaction() {
    return blueprint.Build(query, r.grammar)
}

// If not in transaction, wraps in transaction
return r.orm.Transaction(func(tx contractsorm.Query) error {
    return blueprint.Build(tx, r.grammar)
})
```

### Simple Implementation

Since the schema package is already transaction-aware, the implementation is much simpler:

```go
type SchemerImplementation struct {
    db              *database.Database
    migrations      []contractsschema.MigrationInterface
    useTransactions bool   // Default: true for safety
    isolationLevel  string // Optional isolation level
}

// SetTransactionsEnabled enables or disables transaction wrapping
func (s *SchemerImplementation) SetTransactionsEnabled(enabled bool) {
    s.useTransactions = enabled
}

// SetTransactionIsolationLevel sets the transaction isolation level
func (s *SchemerImplementation) SetTransactionIsolationLevel(level string) {
    s.isolationLevel = level
}
```

### Automatic Transaction Wrapping

The existing methods will internally handle transaction wrapping:

```go
func (s *SchemerImplementation) Up(ctx context.Context) error {
    if s.useTransactions {
        // Wrap in transaction - schema will automatically detect and use it
        return s.db.Transaction(func(tx orm.Query) error {
            return s.executeMigrations(tx, ctx)
        })
    }
    return s.executeMigrationsWithoutTx(ctx)
}

func (s *SchemerImplementation) executeMigrations(tx orm.Query, ctx context.Context) error {
    for _, migration := range s.migrations {
        migration.SetSchema(s.db.Schema())
        if err := migration.Up(); err != nil {
            return err  // Transaction automatically rolls back
        }
    }
    return nil  // Transaction automatically commits
}
```

### Schema Package Already Has Transaction Support

The schema package already has built-in transaction support:

```go
// Schema automatically detects if in transaction
if query.InTransaction() {
    return blueprint.Build(query, r.grammar)
}

// If not in transaction, wraps in transaction
return r.orm.Transaction(func(tx contractsorm.Query) error {
    return blueprint.Build(tx, r.grammar)
})
```

This means the schema package is already transaction-aware and can detect when it's running in a transaction context.

### Migration Tracking Integration

Migration tracker updates will be part of the transaction:
- If migration succeeds, tracker record is committed
- If migration fails, tracker record is rolled back
- Database state and tracker state remain consistent

### Database Limitations Handling

For operations that can't be in transactions:
- Automatically detect and skip transaction wrapping
- Log warnings when transactions are skipped
- Provide clear error messages for unsupported operations
- Allow configuration to force non-transaction mode for specific migrations

## Configuration Options

### Default Behavior

```go
// Default: transactions enabled for safety
schemer := schemer.NewSchemer(db)
// Automatically uses transactions (default enabled)
```

### Enable Transactions Explicitly

```go
schemer := schemer.NewSchemer(db)
schemer.SetTransactionsEnabled(true)
schemer.SetTransactionIsolationLevel("READ COMMITTED")
```

### Disable Transactions

```go
schemer := schemer.NewSchemer(db)
schemer.SetTransactionsEnabled(false)  // For large migrations or specific needs
```

## Benefits and Trade-offs

### Benefits

1. **Transparent to Users**: No API changes, no method proliferation
2. **Automatic Safety**: Transactions enabled by default
3. **Flexibility**: Configuration allows disabling when needed
4. **Backward Compatible**: Existing code works without changes
5. **Production Ready**: Safe defaults for production use

### Trade-offs

1. **Schema Package Changes**: Requires schema package to support transaction-aware operations
2. **Performance**: Default transaction wrapping may impact performance for large migrations
3. **Complexity**: Internal implementation more complex
4. **Database Limitations**: Some operations may not support transactions

## Migration Path

### Phase 1: Schemer Integration (Simplified)

Since the schema package already has transaction support, implementation is straightforward:

1. Add `useTransactions` field to SchemerImplementation (default: true)
2. Add `SetTransactionsEnabled()` setter method
3. Add `SetTransactionIsolationLevel()` setter method
4. Implement automatic transaction wrapping in existing methods
5. Update migration tracking to use transactions

### Phase 2: Testing and Documentation

1. Add comprehensive tests for transaction behavior
2. Add tests for setter methods
3. Update documentation with setter method usage
4. Add examples showing automatic transaction behavior

### Phase 3: Rollout

1. Set default to use transactions (safe default)
2. Monitor performance impact
3. Gather user feedback
4. Adjust default behavior based on real-world usage

## Implementation Status

- ✅ Schema package already has transaction support
- ⏳ Add `useTransactions` field to SchemerImplementation
- ⏳ Add `SetTransactionsEnabled()` setter method
- ⏳ Add `SetTransactionIsolationLevel()` setter method
- ⏳ Implement automatic transaction wrapping in existing methods
- ⏳ Add comprehensive transaction tests
- ⏳ Update documentation with setter method usage
- ⏳ Add performance monitoring and optimization

## Usage Recommendations

### Default Behavior (Recommended)

For most users, the default behavior with transactions enabled is recommended:

```go
// Just use the schemer normally - transactions handled automatically
schemer := schemer.NewSchemer(db)
schemer.AddMigration(&CreateUsersTable{})
schemer.Up(context.Background())
```

### When to Disable Transactions

Users can disable transactions for specific scenarios:

```go
config := schemer.SchemerConfig{
    UseTransactions: false,  // For very large migrations
}
schemer := schemer.NewSchemerWithConfig(db, config)
```

### When Transactions Are Automatically Disabled

The implementation will automatically disable transactions for:
- Operations that don't support transactions (CREATE DATABASE, etc.)
- Very large migrations that exceed transaction limits
- Cross-database operations
- User-configured exclusions

## Example Scenarios

### Default Usage (Automatic Transactions)

```go
// User doesn't need to think about transactions
schemer := schemer.NewSchemer(db)
schemer.AddMigration(&CreateUsersTable{})
schemer.AddMigration(&CreatePostsTable{})

// Automatic transaction wrapping (default enabled)
ctx := context.Background()
err := schemer.Up(ctx)
if err != nil {
    // All changes automatically rolled back
    log.Fatal("Migration failed, database rolled back:", err)
}
```

### Custom Transaction Settings

```go
// For specific needs
schemer := schemer.NewSchemer(db)
schemer.SetTransactionsEnabled(true)
schemer.SetTransactionIsolationLevel("SERIALIZABLE")  // For high consistency
schemer.Up(context.Background())
```

### Large Migration (Disable Transactions)

```go
// For very large data migrations
schemer := schemer.NewSchemer(db)
schemer.SetTransactionsEnabled(false)  // Avoid transaction limits
schemer.Up(context.Background())
```

## Related Documents

- See [schemer-package.md](./schemer-package.md) for the base schemer package implementation
- See [migrations-part-1.md](./migrations-part-1.md) for interface-based migration system
- See [migrations-part-2.md](./migrations-part-2.md) for migration tracking system
