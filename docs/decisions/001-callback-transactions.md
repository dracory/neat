# ADR-001: Callback Transaction Pattern

## Status

Accepted

## Context

Neat ORM needed a transaction management system that would handle database transactions safely and prevent common issues like:
- Transactions not being committed or rolled back
- Resource leaks from unclosed transactions
- Complex error handling in transactional code

## Decision

Neat ORM uses a callback-based transaction pattern where transactions are wrapped in a function that receives a transaction query object:

```go
err := db.Transaction(func(tx contractsorm.Query) error {
    // Transactional operations
    return tx.Create(data)
})
```

The callback automatically handles:
- Beginning the transaction
- Committing if the callback returns nil
- Rolling back if the callback returns an error
- Ensuring the transaction is always closed

## Rationale

The callback pattern provides several advantages over explicit Begin/Commit:

1. **Safety**: Cannot forget to commit or rollback - the library handles it
2. **Resource cleanup**: Guarantees transactions are closed even on panic
3. **Simpler error handling**: Just return an error to rollback, nil to commit
4. **No resource leaks**: Automatic cleanup prevents connection pool exhaustion
5. **Clear scope**: Transaction boundaries are visually obvious in code

## Consequences

**Positive:**
- Eliminates entire class of transaction-related bugs
- Reduces boilerplate code
- Makes transactional code easier to read and maintain

**Negative:**
- Less flexible for complex multi-step transaction workflows (can be worked around with nested transactions)
- May feel unfamiliar to developers used to explicit Begin/Commit
- Cannot easily implement savepoints within the callback (though Neat supports savepoints via separate methods)

**Mitigations:**
- Neat provides savepoint methods for complex transaction scenarios
- The pattern is well-established in Go (similar to sql.Tx with defer)
- Documentation clearly explains the pattern and its benefits
