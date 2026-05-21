# Transactions

This document describes the transaction functionality in Neat ORM.

## What are Transactions?

Transactions allow you to execute multiple database operations as a single atomic unit. If any operation fails, all changes are rolled back.

## Basic Transaction

```go
err := db.Transaction(func(tx neat.Query) error {
    err := tx.Create(&user1)
    if err != nil {
        return err // Rollback
    }
    
    err = tx.Create(&user2)
    if err != nil {
        return err // Rollback
    }
    
    return nil // Commit
})
```

## Transaction Isolation Levels

You can specify transaction isolation levels:

```go
opts := &sql.TxOptions{
    Isolation: sql.LevelSerializable,
}

err := db.Transaction(func(tx neat.Query) error {
    // Transaction operations
    return nil
}, opts)
```

Available isolation levels:
- `sql.LevelReadUncommitted`
- `sql.LevelReadCommitted`
- `sql.LevelRepeatableRead`
- `sql.LevelSerializable`

## Manual Transaction Control

### Begin Transaction

```go
tx, err := db.Begin()
if err != nil {
    return err
}
```

### Commit Transaction

```go
err := tx.Commit()
if err != nil {
    return err
}
```

### Rollback Transaction

```go
err := tx.Rollback()
if err != nil {
    return err
}
```

## Transaction with Context

```go
ctx := context.Background()
err := db.Transaction(func(tx neat.Query) error {
    // Transaction operations with context
    return nil
}, nil, ctx)
```

## Nested Transactions

Neat ORM supports nested transactions using savepoints:

```go
err := db.Transaction(func(tx1 neat.Query) error {
    err := tx1.Transaction(func(tx2 neat.Query) error {
        // Nested transaction operations
        return nil
    })
    return err
})
```

## Transaction Best Practices

1. **Keep transactions short**: Long-running transactions can cause locking issues
2. **Handle errors properly**: Always return errors to ensure rollback
3. **Use appropriate isolation levels**: Choose the right level for your use case
4. **Avoid user interaction**: Don't wait for user input within a transaction

## Note

This documentation is a placeholder and will be expanded as the transaction system is fully implemented.
