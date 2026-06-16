# Schema Transaction Context Support

**Date**: June 15, 2026
**Status**: Implemented
**Priority**: High
**Author**: Neat ORM Team

## Overview

This proposal suggests adding transaction context support to the schema package, enabling schema operations to participate in outer transactions. This is a prerequisite for the schemer package transaction support.

## Motivation

### Current Limitation

The schema package has built-in transaction support for its own operations, but **does not detect or use outer transaction context**:

```go
// Current behavior - schema doesn't detect outer transaction
db.Transaction(func(tx orm.Query) error {
    schema := db.Schema()
    schema.Create("users", func(blueprint Blueprint) {
        blueprint.ID()
        blueprint.String("name")
    })
    // Schema creates its own transaction, ignoring outer transaction
    return nil
})
```

Verification tests confirm:
```go
db.Transaction(func(tx orm.Query) error {
    query := db.Schema().Orm().Query()
    if query.InTransaction() {
        // This is FALSE - schema's ORM doesn't know about outer transaction
    }
    return nil
})
```

### Why This Matters

1. **Schemer Transaction Support**: The schemer package cannot provide transaction safety without this
2. **Migration Atomicity**: Migrations need to be atomic - all changes succeed or all roll back
3. **Data Consistency**: Prevents partial migration states in production
4. **User Expectations**: Users expect schema operations to respect transaction boundaries

## Proposed Solutions

### Solution 1: Transaction-Aware Schema Methods

Add transaction-aware methods to the Schema interface that accept a transaction context:

```go
type Schema interface {
    // Existing methods (no changes)
    Create(table string, callback func(Blueprint)) error
    Drop(table string) error
    DropIfExists(table string) error
    Table(table string, callback func(Blueprint)) error
    HasTable(table string) bool

    // New transaction-aware methods
    CreateWithTx(tx orm.Query, table string, callback func(Blueprint)) error
    DropWithTx(tx orm.Query, table string) error
    DropIfExistsWithTx(tx orm.Query, table string) error
    TableWithTx(tx orm.Query, table string, callback func(Blueprint)) error
    HasTableWithTx(tx orm.Query, table string) bool
}
```

#### Usage Example

```go
db.Transaction(func(tx orm.Query) error {
    schema := db.Schema()
    
    // Use transaction-aware method
    err := schema.CreateWithTx(tx, "users", func(blueprint Blueprint) {
        blueprint.ID()
        blueprint.String("name")
    })
    if err != nil {
        return err // Transaction automatically rolls back
    }
    
    return nil // Transaction commits
})
```

#### Benefits
- ✅ Backward compatible (existing methods unchanged)
- ✅ Explicit transaction usage (clear intent)
- ✅ No breaking changes
- ✅ Easy to implement

#### Drawbacks
- ⚠️ Method duplication (WithTx variants)
- ⚠️ Users must remember to use WithTx methods

### Solution 2: Transaction-Aware Schema Instance  (Recommended)

Create a method to get a transaction-aware schema instance:

```go
type Schema interface {
    // Existing methods...
    Create(table string, callback func(Blueprint)) error
    Drop(table string) error
    
    // New method to create transaction-aware instance
    WithTransaction(tx orm.Query) Schema
}
```

#### Usage Example

```go
db.Transaction(func(tx orm.Query) error {
    // Create transaction-aware schema
    schema := db.Schema().WithTransaction(tx)
    
    // All operations use the transaction
    err := schema.Create("users", func(blueprint Blueprint) {
        blueprint.ID()
        blueprint.String("name")
    })
    if err != nil {
        return err // Transaction rolls back
    }
    
    return nil // Transaction commits
})
```

#### Benefits
- ✅ Clean API (no method duplication)
- ✅ All operations automatically use transaction
- ✅ Backward compatible
- ✅ Fluent interface

#### Drawbacks
- ⚠️ More complex implementation
- ⚠️ Need to track transaction state in schema instance

### Solution 3: Automatic Transaction Detection (Complex)

Enhance schema to automatically detect and use outer transaction context:

```go
// Schema automatically detects outer transaction
db.Transaction(func(tx orm.Query) error {
    schema := db.Schema()
    
    // Schema automatically uses the outer transaction
    schema.Create("users", func(blueprint Blueprint) {
        blueprint.ID()
        blueprint.String("name")
    })
    
    return nil
})
```

#### Benefits
- ✅ Transparent to users
- ✅ No API changes needed
- ✅ Works automatically

#### Drawbacks
- ❌ Very complex implementation
- ❌ Requires thread-local storage or context propagation
- ❌ May have performance implications
- ❌ Difficult to debug

## Recommended Approach

**Solution 2: Transaction-Aware Schema Instance** is recommended because:

1. **Clean API**: No method duplication
2. **Clear Intent**: Explicit transaction usage
3. **Backward Compatible**: Existing code works unchanged
4. **Flexible**: Can be extended for other use cases
5. **Reasonable Complexity**: Not too simple, not too complex

## Implementation Plan

### Phase 1: Schema Package Enhancement

1. Add `WithTransaction(tx orm.Query) Schema` method to Schema interface
2. Implement transaction-aware schema instance
3. Update internal schema operations to use transaction-aware ORM
4. Add tests for transaction-aware schema operations

### Phase 2: Schemer Integration

1. Update schemer to use transaction-aware schema
2. Enable transaction wrapping in schemer methods
3. Update migration execution to use `WithTransaction`
4. Add comprehensive transaction tests

### Phase 3: Documentation and Examples

1. Update schema package documentation
2. Update schemer package documentation
3. Add examples showing transaction usage
4. Update migration guides

## Implementation Details

### Transaction-Aware Schema Instance

```go
// Schema struct enhancement
type Schema struct {
    orm     orm.Orm
    grammar grammar.Grammar
    tx      orm.Query  // Transaction context (nil if not in transaction)
}

// WithTransaction creates a transaction-aware schema instance
func (s *Schema) WithTransaction(tx orm.Query) Schema {
    return &Schema{
        orm:     s.orm,
        grammar: s.grammar,
        tx:      tx,  // Store transaction context
    }
}

// Create uses transaction if available
func (s *Schema) Create(table string, callback func(Blueprint)) error {
    blueprint := NewBlueprint(table)
    callback(blueprint)
    
    // Use transaction if available, otherwise use regular ORM
    var query orm.Query
    if s.tx != nil {
        query = s.tx  // Use transaction
    } else {
        query = s.orm.Query()  // Use regular ORM
    }
    
    // Check if already in transaction
    if query.InTransaction() {
        return blueprint.Build(query, s.grammar)
    }
    
    // Wrap in transaction if not already in one
    return s.orm.Transaction(func(tx orm.Query) error {
        return blueprint.Build(tx, s.grammar)
    })
}
```

### Schemer Integration

```go
func (s *SchemerImplementation) Up(ctx context.Context) error {
    if s.useTransactions {
        return s.db.Transaction(func(tx orm.Query) error {
            return s.upWithTx(ctx, tx)
        })
    }
    return s.up(ctx)
}

func (s *SchemerImplementation) upWithTx(ctx context.Context, tx orm.Query) error {
    // Create transaction-aware schema
    schema := s.db.Schema().WithTransaction(tx)
    
    // Get next batch number
    batch, err := s.getNextBatchNumber()
    if err != nil {
        return fmt.Errorf("failed to get next batch number: %w", err)
    }
    
    // Run pending migrations
    for _, migration := range s.migrations {
        // Inject transaction-aware schema
        migration.SetSchema(schema)
        
        // Run migration
        if err := migration.Up(); err != nil {
            return fmt.Errorf("migration %s failed: %w", migration.Signature(), err)
        }
        
        // Log migration (also uses transaction)
        if err := s.logMigrationWithTx(tx, migration.Signature(), migration.Description(), batch); err != nil {
            return fmt.Errorf("failed to log migration: %w", err)
        }
    }
    
    return nil
}
```

## Benefits

1. **Migration Safety**: Migrations become atomic - all changes succeed or all roll back
2. **Data Consistency**: Prevents partial migration states
3. **Production Ready**: Safe for production deployments
4. **User Expectations**: Schema operations respect transaction boundaries
5. **Backward Compatible**: Existing code continues to work

## Trade-offs

### Pros
- ✅ Enables schemer transaction support
- ✅ Clean API with `WithTransaction`
- ✅ Backward compatible
- ✅ Explicit transaction usage
- ✅ Flexible for future enhancements

### Cons
- ⚠️ Requires schema package changes
- ⚠️ Adds complexity to schema implementation
- ⚠️ Need to track transaction state
- ⚠️ Testing complexity increases

## Implementation Status

- ⏳ Add `WithTransaction` method to Schema interface
- ⏳ Implement transaction-aware schema instance
- ⏳ Update schema operations to use transaction context
- ⏳ Add comprehensive tests for transaction-aware schema
- ⏳ Update schemer to use transaction-aware schema
- ⏳ Enable transaction wrapping in schemer
- ⏳ Update documentation and examples

## Testing Strategy

### Unit Tests
- Schema operations with transaction context
- Schema operations without transaction context
- Transaction rollback behavior
- Nested transaction handling

### Integration Tests
- Schemer with transaction-aware schema
- Failed migrations with automatic rollback
- Successful migrations with commit
- Migration tracker consistency

### Performance Tests
- Transaction overhead measurement
- Large migration performance
- Concurrent migration handling

## Migration Path

### For Schema Package Users

No changes required - existing code continues to work:

```go
// Existing code - still works
schema := db.Schema()
schema.Create("users", func(blueprint Blueprint) {
    blueprint.ID()
})
```

New transaction-aware usage is opt-in:

```go
// New transaction-aware usage
db.Transaction(func(tx orm.Query) error {
    schema := db.Schema().WithTransaction(tx)
    schema.Create("users", func(blueprint Blueprint) {
        blueprint.ID()
    })
    return nil
})
```

### For Schemer Package Users

Once implemented, transaction support works automatically:

```go
// Transactions enabled by default
schemer := schemer.NewSchemer(db)
schemer.SetTransactionsEnabled(true)  // Default

// Migrations automatically use transactions
schemer.Up(context.Background())
```

## Related Documents

- [Schemer Transaction Support Proposal](./schemer-transaction-support.md)
- [Schemer Package Proposal](./schemer-package.md)
- [Schema Package Documentation](../../database/schema/README.md)

## Success Criteria

1. ✅ Schema operations can use outer transaction context
2. ✅ Transaction rollback works correctly
3. ✅ Backward compatibility maintained
4. ✅ Schemer transaction support enabled
5. ✅ Comprehensive tests passing
6. ✅ Documentation complete
7. ✅ Examples demonstrate usage
