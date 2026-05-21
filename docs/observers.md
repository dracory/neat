# Observers

This document describes the observer system in Neat ORM for reacting to model lifecycle events.

## What are Observers?

Observers allow you to hook into model lifecycle events to perform actions before or after database operations.

## Observer Events

### Creating

Called before a model is created:

```go
func (o *UserObserver) Creating(event neat.Event) error {
    // Validate or modify data before creation
    return nil
}
```

### Created

Called after a model is created:

```go
func (o *UserObserver) Created(event neat.Event) error {
    // Perform post-creation actions (e.g., send email)
    return nil
}
```

### Updating

Called before a model is updated:

```go
func (o *UserObserver) Updating(event neat.Event) error {
    // Validate or modify data before update
    return nil
}
```

### Updated

Called after a model is updated:

```go
func (o *UserObserver) Updated(event neat.Event) error {
    // Perform post-update actions
    return nil
}
```

### Deleting

Called before a model is deleted:

```go
func (o *UserObserver) Deleting(event neat.Event) error {
    // Perform pre-deletion actions
    return nil
}
```

### Deleted

Called after a model is deleted:

```go
func (o *UserObserver) Deleted(event neat.Event) error {
    // Perform post-deletion actions
    return nil
}
```

### Force Deleted

Called after a model is force deleted (permanently deleted):

```go
func (o *UserObserver) ForceDeleted(event neat.Event) error {
    // Perform post-force-deletion actions
    return nil
}
```

## Creating an Observer

```go
type UserObserver struct{}

func (o *UserObserver) Creating(event neat.Event) error {
    log.Println("Creating user")
    return nil
}

func (o *UserObserver) Created(event neat.Event) error {
    log.Println("User created")
    return nil
}
```

## Registering Observers

```go
db.Orm().Observe([]neat.ModelToObserver{
    {Model: User{}, Observer: UserObserver{}},
})
```

## Event Interface

The Event interface provides access to:

- `Context()`: Get the context
- `GetAttribute()`: Get a model attribute
- `GetOriginal()`: Get the original attribute value
- `IsClean()`: Check if an attribute is unchanged
- `IsDirty()`: Check if an attribute has changed
- `Query()`: Get the query instance
- `SetAttribute()`: Set a model attribute

## Disabling Events

To disable event dispatching for a specific query:

```go
db.Query().WithoutEvents().Create(&user)
```

## Note

This documentation is a placeholder and will be expanded as the observer system is fully implemented.
