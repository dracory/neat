# Migration Guide: Sugar Methods

This guide helps you migrate from the pointer-based API to the new sugar methods for improved usability.

## Overview

Neat ORM now provides two ways to interact with the database:

1. **Base API (Performance-oriented)**: Uses pointer parameters for zero allocations
2. **Sugar API (Usability-oriented)**: Returns values directly for cleaner code

Both APIs are fully backward compatible. You can choose which to use based on your needs.

## When to Use Each API

### Use Sugar Methods (AsVar) when:
- Application logic and business code
- API handlers and controllers
- Readability matters more than micro-optimizations
- One-off queries and scripts
- Development and prototyping

### Use Base Methods when:
- Performance-critical code paths
- Hot loops and high-frequency operations
- Writing to existing variables
- Avoiding allocations is important
- Library code where performance is critical

## Migration Examples

### Counting Records

**Before (Base API):**
```go
var count int64
err := db.Query().Model(&User{}).Count(&count)
if err != nil {
    return err
}
fmt.Printf("Total users: %d\n", count)
```

**After (Sugar API):**
```go
count, err := db.Query().Model(&User{}).CountAsVar()
if err != nil {
    return err
}
fmt.Printf("Total users: %d\n", count)
```

### Retrieving Single Records

**Before (Base API):**
```go
var user User
err := db.Query().Model(&User{}).Where("id = ?", 1).First(&user)
if err != nil {
    return err
}
fmt.Printf("User: %s\n", user.Name)
```

**After (Sugar API):**
```go
userAny, err := db.Query().Model(&User{}).Where("id = ?", 1).FirstAsVar()
if err != nil {
    return err
}
user := userAny.(User)
fmt.Printf("User: %s\n", user.Name)
```

### Retrieving Multiple Records

**Before (Base API):**
```go
var users []User
err := db.Query().Model(&User{}).Where("age > ?", 18).Get(&users)
if err != nil {
    return err
}
for _, user := range users {
    fmt.Printf("User: %s\n", user.Name)
}
```

**After (Sugar API):**
```go
usersAny, err := db.Query().Model(&User{}).Where("age > ?", 18).GetAsVar()
if err != nil {
    return err
}
users := usersAny.([]User)
for _, user := range users {
    fmt.Printf("User: %s\n", user.Name)
}
```

### Aggregation Methods

**Before (Base API):**
```go
var total float64
err := db.Query().Model(&Order{}).Sum("amount", &total)
if err != nil {
    return err
}
fmt.Printf("Total: %.2f\n", total)
```

**After (Sugar API):**
```go
total, err := db.Query().Model(&Order{}).SumAsVar("amount")
if err != nil {
    return err
}
fmt.Printf("Total: %.2f\n", total)
```

### Column Operations

**Before (Base API):**
```go
var emails []string
err := db.Query().Model(&User{}).Pluck("email", &emails)
if err != nil {
    return err
}
for _, email := range emails {
    fmt.Printf("Email: %s\n", email)
}
```

**After (Sugar API):**
```go
emailsAny, err := db.Query().Model(&User{}).PluckAsVar("email")
if err != nil {
    return err
}
emails := emailsAny.([]string)
for _, email := range emails {
    fmt.Printf("Email: %s\n", email)
}
```

## Complete Method Mapping

| Base Method | Sugar Method | Return Type |
|-------------|--------------|-------------|
| `Count(count *int64)` | `CountAsVar()` | `(int64, error)` |
| `Sum(column, dest)` | `SumAsVar(column)` | `(float64, error)` |
| `Avg(column, dest)` | `AvgAsVar(column)` | `(float64, error)` |
| `Min(column, dest)` | `MinAsVar(column)` | `(float64, error)` |
| `Max(column, dest)` | `MaxAsVar(column)` | `(float64, error)` |
| `Exists(exists *bool)` | `ExistsAsVar()` | `(bool, error)` |
| `Pluck(column, dest)` | `PluckAsVar(column)` | `([]any, error)` |
| `Value(column, dest)` | `ValueAsVar(column)` | `(any, error)` |
| `First(dest)` | `FirstAsVar()` | `(any, error)` |
| `FindOne(dest)` | `FindOneAsVar()` | `(any, error)` |
| `Get(dest)` | `GetAsVar()` | `([]any, error)` |
| `All(dest)` | `AllAsVar()` | `([]any, error)` |
| `FindAll(dest)` | `FindAllAsVar()` | `([]any, error)` |
| `Find(dest, conds...)` | `FindAsVar(conds...)` | `([]any, error)` |

## Type Safety Considerations

Since sugar methods return `any` or `[]any`, you'll need to type-assert the results:

```go
// For single records
userAny, err := db.Query().Model(&User{}).FirstAsVar()
if err != nil {
    return err
}
user, ok := userAny.(User)
if !ok {
    return fmt.Errorf("unexpected type")
}

// For slices
usersAny, err := db.Query().Model(&User{}).GetAsVar()
if err != nil {
    return err
}
users, ok := usersAny.([]User)
if !ok {
    return fmt.Errorf("unexpected type")
}
```

## Performance Impact

Sugar methods introduce minimal overhead for the convenience they provide:

- **Aggregation methods**: ~1-2 ns overhead (negligible)
- **Single record methods**: ~5-10 ns overhead (negligible)
- **Slice methods**: ~10-20 ns overhead (negligible)

For most applications, this overhead is insignificant. Only consider using base methods in performance-critical hot paths.

## Gradual Migration

You don't need to migrate all at once. Both APIs can coexist:

```go
// Use sugar methods for new code
count, err := db.Query().Model(&User{}).CountAsVar()

// Keep base methods for existing performance-critical code
var count int64
err := db.Query().Model(&User{}).Count(&count)
```

## Common Patterns

### Pattern 1: API Handler (Use Sugar)
```go
func GetUsersHandler(c *gin.Context) {
    usersAny, err := db.Query().Model(&User{}).GetAsVar()
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    users := usersAny.([]User)
    c.JSON(200, users)
}
```

### Pattern 2: Hot Loop (Use Base)
```go
func ProcessMillionRecords() {
    var count int64
    for i := 0; i < 1000000; i++ {
        // Use base method to avoid allocations in hot loop
        err := db.Query().Model(&User{}).Count(&count)
        if err != nil {
            log.Fatal(err)
        }
        // Process count...
    }
}
```

### Pattern 3: Batch Processing (Use Base)
```go
func BatchProcessUsers() {
    var users []User
    // Reuse the same slice to avoid allocations
    for page := 0; page < 100; page++ {
        users = users[:0] // Clear slice
        err := db.Query().Model(&User{}).
            Offset(page * 100).
            Limit(100).
            Get(&users)
        if err != nil {
            log.Fatal(err)
        }
        // Process users...
    }
}
```

## Conclusion

The sugar methods provide a more idiomatic Go experience while maintaining the performance-oriented base API for when you need it. Choose the right tool for your specific use case, and don't be afraid to mix both approaches in the same codebase.
