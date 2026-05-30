# Factory Pattern

The Factory pattern in Neat ORM provides a convenient way to create test data for your database. It allows you to define model templates and generate multiple instances with optional attribute overrides.

## Overview

The Factory system consists of:

- **Factory Interface**: Defines the contract for creating model instances
- **Factory Implementation**: Provides methods for creating and persisting models
- **Model Factory**: Optional interface for models to define their own factory

## Basic Usage

### Creating a Single Model

```go
// Create a single user in the database
user := &User{Name: "John", Email: "john@example.com"}
createdUser, err := db.Factory().Table("users").Create(user)
```

### Creating Multiple Models

```go
// Create 5 users in the database
template := &User{Name: "Template User"}
users, err := db.Factory().Table("users").Count(5).Create(template)
```

### Creating Models with Custom Attributes

```go
// Create a user with custom attributes
template := &User{Name: "Template"}
user, err := db.Factory().Table("users").Create(template, map[string]any{
    "Name":  "Custom Name",
    "Email": "custom@example.com",
})
```

### Creating In-Memory Models

```go
// Create a user without persisting to database
user, err := db.Factory().Make(&User{Name: "Test User"})
```

### Creating Multiple In-Memory Models

```go
// Create 3 users in memory
users, err := db.Factory().Count(3).Make(&User{Name: "Template"})
```

## Factory Methods

### Table(table string) Factory

Sets the table name for database operations. This is required for `Create()` and `CreateQuietly()` operations.

```go
factory := db.Factory().Table("users")
```

### Count(count int) Factory

Sets the number of models that should be generated. Can be chained with `Create`, `CreateQuietly`, or `Make` for bulk operations.

```go
factory := db.Factory().Count(10)
```

### Create(value any, attributes ...map[string]any) (any, error)

Creates a model and persists it to the database, returning the created instance(s). Requires `Table()` to be called first.

**Parameters:**
- `value`: The model template (single struct or slice)
- `attributes`: Optional map of attributes to override

**Returns:**
- The created instance(s)
- Error if creation fails

```go
user, err := db.Factory().Table("users").Create(&User{Name: "John"})
users, err := db.Factory().Table("users").Count(5).Create(&User{Name: "Template"})
```

### CreateQuietly(value any, attributes ...map[string]any) (any, error)

Creates a model and persists it to the database without firing any model events, returning the created instance(s). Requires `Table()` to be called first.

**Parameters:**
- `value`: The model template (single struct or slice)
- `attributes`: Optional map of attributes to override

**Returns:**
- The created instance(s)
- Error if creation fails

```go
user, err := db.Factory().Table("users").CreateQuietly(&User{Name: "John"})
```

This is useful when you want to avoid triggering model observers during data seeding.

### Make(value any, attributes ...map[string]any) (any, error)

Creates a model instance in memory but does not persist it to the database. Useful for testing without side effects.

**Parameters:**
- `value`: The model template (single struct or slice)
- `attributes`: Optional map of attributes to override

**Returns:**
- The created instance(s)
- Error if creation fails

```go
user, err := db.Factory().Make(&User{Name: "Test User"})
users, err := db.Factory().Count(3).Make(&User{Name: "Template"})
```

## Defining Model Factories

You can define a factory for your models by implementing the `Factory` interface:

```go
type UserFactory struct{}

func (f *UserFactory) Definition() map[string]any {
    return map[string]any{
        "Name":  "Default User",
        "Email": "default@example.com",
    }
}

type User struct {
    ID    uint
    Name  string
    Email string
}

func (u *User) Factory() contractsfactory.Factory {
    return &UserFactory{}
}
```

## Use Cases

### In-Memory Testing

Use `Factory.Make()` to create instances without database persistence:

```go
func TestUserValidation(t *testing.T) {
    user, err := db.Factory().Make(&User{Name: "Test"})
    if err != nil {
        t.Fatal(err)
    }
    
    // Test validation logic without database
    if user.Name == "" {
        t.Error("Name is required")
    }
}
```

### Bulk In-Memory Operations

Use `Factory.Count().Make()` to create multiple instances in memory:

```go
users, err := db.Factory().Count(100).Make(&User{Name: "Template"})
```

### Database Seeding

Use `Factory.Table().Create()` to persist models to the database:

```go
user, err := db.Factory().Table("users").Create(&User{Name: "John"})
```

### Event-Free Creation

Use `Factory.Table().CreateQuietly()` when you don't want to trigger model observers:

```go
user, err := db.Factory().Table("users").CreateQuietly(&User{Name: "John"})
```

This is useful for seeding data where you don't want observers to run.

### Bulk Database Operations

Use `Factory.Table().Count().Create()` to create multiple records efficiently:

```go
users, err := db.Factory().Table("users").Count(50).Create(&User{Name: "Template"})
```

## Attribute Overrides

You can override template attributes when creating models:

```go
template := &User{Name: "Template", Email: "template@example.com"}
user, err := db.Factory().Table("users").Create(template, map[string]any{
    "Name":  "Custom Name",
    "Email": "custom@example.com",
})
```

The factory will merge the attributes with the template, with the attributes taking precedence.

## Working with Slices

You can also pass slices to the factory for bulk operations:

```go
templates := []User{
    {Name: "User 1"},
    {Name: "User 2"},
}
users, err := db.Factory().Table("users").Count(5).Create(templates)
```

This will create 5 users, using the templates as a base.

## Best Practices

1. **Use descriptive templates**: Make your template models clear and representative
2. **Use Make for unit tests**: Prefer in-memory creation for unit tests to avoid database overhead
3. **Use Create for integration tests**: Use database creation when you need to test database interactions
4. **Use CreateQuietly for seeding**: Avoid triggering observers when seeding initial data
5. **Clean up after tests**: Always clean up created data in test teardown
6. **Use Count for bulk operations**: Generate multiple instances efficiently with Count()
7. **Override attributes wisely**: Use attribute overrides to create variations without defining multiple templates

## Troubleshooting

### Create fails with "query not initialized"
- Ensure you have called `Table()` before `Create()` or `CreateQuietly()`
- Verify the database connection is properly configured

### Attributes not applied
- Check that attribute keys match struct field names (case-sensitive)
- For JSON-tagged fields, use the JSON tag name instead of the struct field name
- Verify attribute values are convertible to the target field types

### Make returns slice instead of single instance
- When using `Count() > 1`, `Make()` always returns a slice
- Use `Count(1)` or omit `Count()` for single instances

### Events still firing with CreateQuietly
- Verify you're using `CreateQuietly()` and not `Create()`
- Check that observers are properly registered

## API Reference

### Factory Interface

```go
type Factory interface {
    // Count sets the number of models that should be generated
    Count(count int) Factory
    // Table sets the table name for database operations
    Table(table string) Factory
    // Create creates a model and persists it to the database
    Create(value any, attributes ...map[string]any) (any, error)
    // CreateQuietly creates a model without firing events
    CreateQuietly(value any, attributes ...map[string]any) (any, error)
    // Make creates a model in memory without persisting
    Make(value any, attributes ...map[string]any) (any, error)
}
```

### Model Interface (Optional)

```go
type Model interface {
    // Factory creates a new factory instance for the model
    Factory() Factory
}
```

## Comparison with Seeders

**Factories** are best for:
- Test data generation
- Creating variations of the same model
- In-memory testing
- Bulk data creation

**Seeders** are best for:
- Initial database population
- Reference data
- Production data setup
- One-time data insertion

You can use both together - use seeders for initial data and factories for test data.
