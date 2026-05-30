# Performance Guide

This guide covers performance optimization strategies and best practices for Neat ORM.

## Overview

Neat ORM is designed for performance, but like any ORM, it requires careful usage to achieve optimal results. This guide covers connection pooling, query optimization, eager loading, batch operations, and benchmarking.

## Connection Pooling

### Understanding Connection Pooling

Connection pooling reuses database connections instead of creating a new one for each query, significantly improving performance.

### Configuring Connection Pool

```go
config := neat.DBConfig{
    Default: "default",
    Connections: map[string]neat.ConnectionConfig{
        "default": {
            Driver:   "mysql",
            Host:     "localhost",
            Port:     3306,
            Database: "mydb",
            Username: "user",
            Password: "password",
            Pool: neat.PoolConfig{
                MaxOpenConns: 25,    // Maximum open connections
                MaxIdleConns: 10,    // Maximum idle connections
                MaxLifetime:  300,  // Connection lifetime in seconds
            },
        },
    },
}

db, err := neat.New(config)
```

### Connection Pool Best Practices

1. **Set appropriate MaxOpenConns**: Start with `CPU cores * 2` and adjust based on load
2. **Set MaxIdleConns**: Usually 50-75% of MaxOpenConns
3. **Set MaxLifetime**: Prevents stale connections (typically 5-10 minutes)
4. **Monitor pool metrics**: Check for connection exhaustion
5. **Use context with timeout**: Prevent long-running queries from holding connections

### Monitoring Connection Pool

```go
stats := db.Stats()
fmt.Printf("Open connections: %d\n", stats.OpenConnections)
fmt.Printf("In use: %d\n", stats.InUse)
fmt.Printf("Idle: %d\n", stats.Idle)
```

## Query Optimization

### Select Only Needed Columns

Avoid selecting all columns when you only need a few:

```go
// Bad - selects all columns
var users []User
db.Query().Get(&users)

// Good - selects only needed columns
var users []User
db.Query().Select("id", "name").Get(&users)
```

### Use Indexes

Ensure columns used in WHERE, JOIN, and ORDER BY clauses are indexed:

```go
// Create index on frequently queried column
db.Schema().Table("users", func(table neat.Blueprint) {
    table.String("email").Index()
})
```

### Limit Results

Always use LIMIT when you don't need all records:

```go
// Bad - could return millions of records
var users []User
db.Query().Get(&users)

// Good - limits results
var users []User
db.Query().Limit(100).Get(&users)
```

### Use WHERE Instead of Filtering in Code

```go
// Bad - fetches all then filters
var users []User
db.Query().Get(&users)
var activeUsers []User
for _, user := range users {
    if user.Active {
        activeUsers = append(activeUsers, user)
    }
}

// Good - filters at database level
var users []User
db.Query().Where("active", true).Get(&users)
```

### Avoid N+1 Queries

N+1 queries occur when you execute one query to fetch records, then additional queries for each record:

```go
// Bad - N+1 query problem
var posts []Post
db.Query().Get(&posts)
for _, post := range posts {
    var user User
    db.Query().Where("id", post.UserID).First(&user) // N queries
    post.User = user
}

// Good - eager loading
var posts []Post
db.Query().With("user").Get(&posts) // 2 queries total
```

### Use Exists Instead of Count

When checking if records exist, use EXISTS instead of COUNT:

```go
// Bad - counts all records
count, _ := db.Query().Where("active", true).Count()
if count > 0 {
    // ...
}

// Good - checks existence
exists, _ := db.Query().Where("active", true).Exists()
if exists {
    // ...
}
```

## Eager Loading vs Lazy Loading

### Eager Loading

Eager loading loads related records in a single query using JOINs:

```go
// Eager load posts with user
var posts []Post
db.Query().With("user").Get(&posts)
```

**Use when:**
- You know you'll need the related data
- Loading a small number of records
- Performance is critical

**Pros:**
- Single query (or few queries)
- Predictable performance
- No N+1 problem

**Cons:**
- Can load unnecessary data
- Complex JOINs can be slow

### Lazy Loading

Lazy loading loads related records only when accessed:

```go
var post Post
db.Query().First(&post)
// Load user only when needed
db.Query().Load(&post, "user")
```

**Use when:**
- Related data might not be needed
- Loading many records
- Memory is a concern

**Pros:**
- Only loads data when needed
- Lower memory usage
- Simpler initial queries

**Cons:**
- N+1 query problem
- Unpredictable performance
- More round trips to database

### Hybrid Approach

Combine both approaches for optimal performance:

```go
// Eager load primary relationships
var posts []Post
db.Query().With("user").Get(&posts)

// Lazy load secondary relationships as needed
for _, post := range posts {
    if post.NeedsComments {
        db.Query().Load(&post, "comments")
    }
}
```

## Batch Operations

### Bulk Insert

Use bulk insert instead of individual inserts:

```go
// Bad - individual inserts
for _, user := range users {
    db.Query().Create(user)
}

// Good - bulk insert
db.Query().Create(users)
```

### Bulk Update

Use bulk update instead of individual updates:

```go
// Bad - individual updates
for _, id := range ids {
    db.Query().Where("id", id).Update("status", "active")
}

// Good - bulk update
db.Query().WhereIn("id", ids).Update("status", "active")
```

### Bulk Delete

Use bulk delete instead of individual deletes:

```go
// Bad - individual deletes
for _, id := range ids {
    db.Query().Where("id", id).Delete()
}

// Good - bulk delete
db.Query().WhereIn("id", ids).Delete()
```

### Use Transactions for Batch Operations

Wrap batch operations in transactions for better performance:

```go
err := db.Transaction(func(tx neat.Query) error {
    for _, user := range users {
        if err := tx.Create(user); err != nil {
            return err
        }
    }
    return nil
})
```

## Caching Strategies

### Query Result Caching

Cache frequently accessed query results:

```go
type Cache struct {
    data map[string]interface{}
    mu   sync.RWMutex
}

func (c *Cache) Get(key string) (interface{}, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    val, ok := c.data[key]
    return val, ok
}

func (c *Cache) Set(key string, value interface{}) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.data[key] = value
}

// Usage
cache := &Cache{data: make(map[string]interface{})}

// Check cache first
if users, ok := cache.Get("active_users"); ok {
    return users.([]User)
}

// Query and cache
var users []User
db.Query().Where("active", true).Get(&users)
cache.Set("active_users", users)
```

### Schema Caching

Neat ORM caches schema information automatically. Ensure your schema doesn't change frequently for optimal performance.

## Benchmarking

### Writing Benchmarks

```go
func BenchmarkQueryWhere(b *testing.B) {
    db := setupBenchmarkDatabase()
    defer db.Close()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        var users []User
        db.Query().Where("active", true).Get(&users)
    }
}
```

### Running Benchmarks

```bash
# Run all benchmarks
go test -bench=. ./...

# Run specific benchmark
go test -bench=BenchmarkQueryWhere ./...

# Run with memory allocation stats
go test -bench=. -benchmem ./...

# Run with CPU profiling
go test -bench=. -cpuprofile=cpu.prof ./...
```

### Benchmark Results Analysis

```bash
# Analyze CPU profile
go tool pprof cpu.prof

# View top functions
(pprof) top10

# View graph
(pprof) web
```

## Performance Monitoring

### Enable Query Logging

```go
// Enable query logging for debugging
db.Query().Log()
```

### Measure Query Execution Time

```go
start := time.Now()
var users []User
db.Query().Where("active", true).Get(&users)
duration := time.Since(start)
fmt.Printf("Query took: %v\n", duration)
```

### Use EXPLAIN to Analyze Queries

```go
// Get the SQL and run EXPLAIN
sql, _ := db.Query().Where("active", true).ToSql()
// Run EXPLAIN on your database to analyze the query plan
```

## Database-Specific Optimizations

### MySQL

```go
config := neat.ConnectionConfig{
    Driver: "mysql",
    // MySQL-specific optimizations
    Options: map[string]string{
        "parseTime": "true",
        "loc":       "Local",
    },
}
```

### PostgreSQL

```go
config := neat.ConnectionConfig{
    Driver: "postgres",
    // PostgreSQL-specific optimizations
    Options: map[string]string{
        "sslmode": "disable",
    },
}
```

### SQLite

```go
// Use WAL mode for better concurrency
config := neat.ConnectionConfig{
    Driver: "sqlite",
    DSN:    "sqlite:///path/to/db.db?cache=shared&_journal_mode=WAL",
}
```

## Common Performance Pitfalls

### 1. Not Using Indexes

**Problem**: Queries on unindexed columns are slow.

**Solution**: Add indexes to frequently queried columns.

### 2. Selecting Too Many Columns

**Problem**: Selecting all columns wastes bandwidth and memory.

**Solution**: Use `Select()` to specify only needed columns.

### 3. N+1 Queries

**Problem**: Loading related records one at a time.

**Solution**: Use eager loading with `With()`.

### 4. Not Limiting Results

**Problem**: Queries can return millions of records.

**Solution**: Always use `Limit()` when appropriate.

### 5. Inefficient Connection Pooling

**Problem**: Too few or too many connections.

**Solution**: Tune pool settings based on your workload.

### 6. Not Using Transactions

**Problem**: Multiple round trips to database.

**Solution**: Wrap related operations in transactions.

### 7. Ignoring Query Plans

**Problem**: Queries use suboptimal execution plans.

**Solution**: Use EXPLAIN to analyze and optimize queries.

## Performance Checklist

- [ ] Connection pool configured appropriately
- [ ] Indexes on frequently queried columns
- [ ] SELECT only needed columns
- [ ] LIMIT results when appropriate
- [ ] Use eager loading to avoid N+1 queries
- [ ] Use batch operations for bulk changes
- [ ] Wrap related operations in transactions
- [ ] Cache frequently accessed data
- [ ] Monitor query performance
- [ ] Profile slow queries
- [ ] Use appropriate data types
- [ ] Regularly analyze and optimize database schema

## Additional Resources

- [MySQL Performance Tuning](https://dev.mysql.com/doc/refman/8.0/en/optimization.html)
- [PostgreSQL Performance Tips](https://www.postgresql.org/docs/current/performance-tips.html)
- [SQLite Optimization](https://www.sqlite.org/optoverview.html)
