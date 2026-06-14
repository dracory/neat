# Sugar Methods Examples

This directory contains comprehensive examples demonstrating the new sugar methods in Neat ORM.

## Overview

Sugar methods provide a more convenient, idiomatic Go API by returning values directly instead of requiring pointer parameters. This reduces boilerplate and improves code readability while maintaining the existing performance-oriented API for when you need it.

## Running the Examples

```bash
cd examples/sugar-methods
go run main.go
```

## What's Demonstrated

This example demonstrates all 14 sugar methods:

### Aggregation Methods
1. **CountAsVar()** - Count records matching the query
2. **SumAsVar(column)** - Sum of column values
3. **AvgAsVar(column)** - Average of column values
4. **MinAsVar(column)** - Minimum value in column
5. **MaxAsVar(column)** - Maximum value in column
6. **ExistsAsVar()** - Check if any records match the query

### Column Operations
7. **PluckAsVar(column)** - Retrieve single column values as a slice
8. **ValueAsVar(column)** - Retrieve single column value from first record

### Retrieval Methods
9. **FirstAsVar()** - Retrieve first record matching the query
10. **FindOneAsVar()** - Alias for FirstAsVar (Sequelize-style)
11. **GetAsVar()** - Retrieve all records matching the query
12. **AllAsVar()** - Alias for GetAsVar (Django-style)
13. **FindAllAsVar()** - Alias for GetAsVar (Sequelize-style)
14. **FindAsVar(conds...)** - Retrieve records with conditions

## Comparison: Base API vs Sugar API

### Base API (Performance-oriented)
```go
var count int64
err := db.Query().Model(&User{}).Count(&count)
```

### Sugar API (Usability-oriented)
```go
count, err := db.Query().Model(&User{}).CountAsVar()
```

## Type Safety

Since sugar methods return `any` or `[]any`, you'll need to type-assert the results:

```go
// For single records
userAny, err := db.Query().Model(&User{}).FirstAsVar()
if err != nil {
    return err
}
user := userAny.(User)

// For slices
usersAny, err := db.Query().Model(&User{}).GetAsVar()
if err != nil {
    return err
}
users := usersAny.([]User)
```

## When to Use Sugar Methods

**Use sugar methods when:**
- Application logic and business code
- API handlers and controllers
- Readability matters more than micro-optimizations
- One-off queries and scripts
- Development and prototyping

**Use base methods when:**
- Performance-critical code paths
- Hot loops and high-frequency operations
- Writing to existing variables
- Avoiding allocations is important

## Performance Impact

Sugar methods introduce minimal overhead:
- Aggregation methods: ~1-2 ns overhead
- Single record methods: ~5-10 ns overhead
- Slice methods: ~10-20 ns overhead

For most applications, this overhead is insignificant.

## Learn More

- [Migration Guide](../../docs/migration-guide-sugar-methods.md) - Detailed guide for migrating existing code
- [API Reference](../../docs/api-reference.html) - Complete API documentation
- [Query Builder Documentation](../../docs/query-builder.html) - Query builder usage guide
