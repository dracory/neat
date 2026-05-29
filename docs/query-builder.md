# Query Builder

This document describes the query builder API in Neat ORM.

## Basic Queries

### Selecting Records

```go
// Get all records
var users []User
err := db.Query().Get(&users)

// Get first record
var user User
err := db.Query().First(&user)

// Find by ID
var user User
err := db.Query().Find(&user, 1)
```

### Where Clauses

```go
// Laravel-style where (implicit = operator)
db.Query().Where("name", "John")

// Multiple conditions
db.Query().Where("name", "John").Where("age", 30)

// Explicit operator with spaces
db.Query().Where("age > ?", 18)

// Explicit operator without spaces
db.Query().Where("age>?", 18)

// OrWhere (also supports Laravel-style)
db.Query().Where("name", "John").OrWhere("name", "Jane")

// Complex conditions
db.Query().Where("name LIKE ?", "John%")
db.Query().Where("age BETWEEN ? AND ?", 18, 65)
db.Query().Where("id IN (?)", []any{1, 2, 3})
```

**Where Syntax Options:**

The query builder supports multiple Where syntax styles:

1. **Laravel-style** (implicit `=` operator):
   - `Where("column", "value")` → `column = ?`
   - `OrWhere("column", "value")` → `column = ?`

2. **Explicit operator with spaces**:
   - `Where("column = ?", "value")`
   - `Where("age > ?", 18)`
   - `Where("name LIKE ?", "pattern")`

3. **Explicit operator without spaces**:
   - `Where("column=?", "value")`
   - `Where("age>?", 18)`
   - `Where("name LIKE ?", "pattern")`

The Laravel-style syntax is automatically applied when:
- Exactly one argument is provided
- The query string does not contain an SQL operator
- The query string does not contain operator-like keywords (LIKE, IN, BETWEEN, etc.)

### Ordering

```go
// Order by
db.Query().OrderBy("created_at", "desc")

// Multiple order by
db.Query().OrderBy("name", "asc").OrderBy("created_at", "desc")
```

### Limit and Offset

```go
// Limit
db.Query().Limit(10)

// Offset
db.Query().Offset(20)

// Pagination
db.Query().Limit(10).Offset(20)
```

### Aggregations

```go
// Count
count, err := db.Query().Count()

// Sum
sum, err := db.Query().Sum("amount")

// Average
avg, err := db.Query().Avg("amount")

// Min
min, err := db.Query().Min("amount")

// Max
max, err := db.Query().Max("amount")
```

## Advanced Queries

### Joins

```go
// Inner join
db.Query().Join("posts", "users.id", "=", "posts.user_id")

// Left join
db.Query().LeftJoin("posts", "users.id", "=", "posts.user_id")
```

### Group By and Having

```go
db.Query().GroupBy("category").Having("count", ">", 10)
```

### Subqueries

```go
// Subquery in where clause
subquery := db.Query().Select("id").From("orders").Where("status", "pending")
db.Query().WhereIn("user_id", subquery)
```

## Creating, Updating, Deleting

### Create

```go
user := User{Name: "John", Email: "john@example.com"}
err := db.Query().Create(&user)
```

### Update

```go
err := db.Query().Where("id", 1).Update("name", "Jane")
```

### Delete

```go
result, err := db.Query().Where("id", 1).Delete()
```

## ToSql Interface

Generate SQL without executing:

```go
sql, err := db.Query().Where("name", "John").ToSql()
```

## Note

This documentation is a placeholder and will be expanded as the query builder API is finalized.
