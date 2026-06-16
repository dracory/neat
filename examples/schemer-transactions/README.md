# Schemer Transaction Control Example

This example demonstrates transaction control in the schemer package for safe migration execution with automatic rollback on failure.

## Features Demonstrated

### Transaction Control
- **SetTransactionsEnabled**: Enable or disable transaction wrapping
- **SetTransactionIsolationLevel**: Set transaction isolation level
- **Automatic Rollback**: Failed migrations automatically roll back when transactions are enabled
- **Flexible Configuration**: Can disable transactions for large migrations

### Transaction Isolation Levels
- `READ UNCOMMITTED`
- `READ COMMITTED`
- `REPEATABLE READ`
- `SERIALIZABLE`
- `SNAPSHOT`

## Running the Example

```bash
cd examples/schemer-transactions
go run main.go
```

This will:
1. Create a SQLite database (`example_transactions.db`)
2. Configure transaction settings (enabled, SERIALIZABLE isolation)
3. Run migrations with transaction wrapping
4. Disable transactions for large migrations
5. Run migrations without transaction wrapping

## Usage Pattern

```go
// Create schemer instance
schemer := schemer.NewSchemer(db)

// Configure transaction settings
schemer.SetTransactionsEnabled(true)
schemer.SetTransactionIsolationLevel("SERIALIZABLE")

// Add migrations
schemer.AddMigration(&CreateUsersTable{})
schemer.AddMigration(&CreatePostsTable{})

// Run migrations with transaction wrapping
ctx := context.Background()
err := schemer.Up(ctx)
if err != nil {
    // All changes automatically rolled back
    log.Fatal("Migration failed, database rolled back:", err)
}

// Disable for large migrations
schemer.SetTransactionsEnabled(false)
schemer.AddMigration(&AddPostsIndexes{})
schemer.Up(ctx)
```

## When to Use Transactions

### Enable Transactions (Recommended)
- **Production deployments**: Always use transactions for safety
- **Multi-table migrations**: Ensure all tables are created together
- **Data migrations**: Use when migrating data
- **Complex migrations**: Use for migrations with multiple steps

### Disable Transactions
- **Very large migrations**: May exceed database transaction limits
- **DDL-heavy operations**: Some databases don't support DDL in transactions
- **Long-running migrations**: May cause transaction timeouts

## Benefits

1. **Automatic Safety**: No need to think about transactions - just configure and run
2. **Atomic Execution**: Either all changes succeed, or none do
3. **Automatic Rollback**: Failed migrations automatically roll back
4. **Data Consistency**: Database is always in a valid state
5. **Flexible**: Can disable when needed for specific scenarios

## Note

Transaction wrapping is currently disabled by default in the schemer package pending verification of schema transaction detection. The infrastructure is in place and demonstrated in this example. Once schema transaction behavior is properly tested, transaction wrapping can be enabled by default for safe migration execution.

## Related Documentation

- [Schemer Package README](../../database/schemer/README.md)
- [Transaction Support Proposal](../../docs/proposals/schemer-transaction-support.md)
