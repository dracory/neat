# Transaction Failure and Automatic Rollback Example

This example demonstrates how transaction failure behavior works in the schemer package, showing automatic rollback when migrations fail.

## Features Demonstrated

### Transaction Failure Scenarios

1. **Migration Failure WITHOUT Transactions**
   - Shows partial migration state when transactions are disabled
   - Tables created before failure remain in database
   - Manual cleanup required

2. **Migration Failure WITH Transactions**
   - Demonstrates automatic rollback on failure
   - All changes within transaction are rolled back
   - Database remains in consistent state
   - Migration tracker reflects actual state

3. **Successful Migration WITH Transactions**
   - Shows normal successful migration execution
   - All changes committed atomically
   - Database state is consistent

## Running the Example

```bash
cd examples/schemer-transaction-failure
go run main.go
```

This will:
1. Run a failing migration without transactions (shows partial state)
2. Run a failing migration with transactions (shows automatic rollback)
3. Run successful migrations with transactions (shows normal execution)

## Expected Output

**Note**: Transaction wrapping is currently disabled in the schemer package pending verification of schema transaction detection. The example below shows the current behavior (no automatic rollback) and demonstrates what would happen once transaction wrapping is enabled.

```
=== Transaction Failure and Automatic Rollback Example ===

=== Example 1: Migration Failure WITHOUT Transactions ===
Transactions enabled: false
Migration failed (expected): migration 2024_06_15_120200_failing_migration failed: intentional migration failure for demonstration

=== Checking Database State After Failure ===
⚠️  'users' table still exists (partial migration state)

=== Example 2: Migration Failure WITH Transactions ===
Transactions enabled: true
Migration failed (expected): migration 2024_06_15_120200_failing_migration failed: intentional migration failure for demonstration

=== Checking Database State After Failure ===
⚠️  'users' table still exists (transaction wrapping not yet enabled)
✓ 'migration_tracker' table exists (created before failure)

=== Migration Status ===
⚠️  2 migrations recorded (transaction wrapping not yet enabled)
  - 2024_06_15_000000_create_migration_tracker_table: completed
  - 2024_06_15_120000_create_users_table: completed

=== Example 3: Successful Migration WITH Transactions ===
Transactions enabled: true
✓ All migrations completed successfully

=== Verifying Database State ===
✓ 'users' table exists
✓ 'posts' table exists

=== Summary ===
⚠️ Transaction wrapping currently disabled (pending schema transaction verification)
✓ Once enabled, transactions will prevent partial migration states
✓ Once enabled, failed migrations will automatically roll back
✓ Once enabled, database will remain in consistent state
✓ Once enabled, migration tracker will reflect actual state
```

## Key Concepts

### Without Transactions
- **Partial State**: Tables created before failure remain
- **Manual Cleanup**: Must manually clean up failed migrations
- **Inconsistent State**: Database may be in invalid state
- **Risk**: Production deployments can leave database in bad state

### With Transactions
- **Atomic Execution**: Either all changes succeed, or none do
- **Automatic Rollback**: Failed migrations automatically roll back
- **Consistent State**: Database always in valid state
- **Safe**: Production deployments are much safer

## Benefits of Transaction Wrapping

1. **Data Safety**: Prevents partial migration states
2. **Automatic Cleanup**: No manual rollback needed on failure
3. **Consistency**: Database and tracker always in sync
4. **Production Safety**: Reduces deployment risk significantly
5. **Peace of Mind**: Know that failures won't corrupt database

## When Transactions Are Critical

- **Production Deployments**: Always use transactions
- **Multi-Table Migrations**: Ensure all tables created together
- **Data Migrations**: Protect data integrity
- **Complex Migrations**: Multiple steps should be atomic
- **Critical Systems**: Where data consistency is paramount

## Note on Current Implementation

Transaction wrapping is currently disabled by default in the schemer package pending verification of schema transaction detection. This example demonstrates the behavior once transaction wrapping is enabled.

## Related Documentation

- [Schemer Transaction Control Example](../schemer-transactions/)
- [Schemer Package README](../../database/schemer/README.md)
- [Transaction Support Proposal](../../docs/proposals/schemer-transaction-support.md)
