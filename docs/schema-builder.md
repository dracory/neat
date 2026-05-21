# Schema Builder

This document describes the schema builder API in Neat ORM.

## Creating Tables

```go
err := db.Schema().Create("users", func(table neat.Blueprint) {
    table.ID()
    table.String("name")
    table.String("email").Unique()
    table.Timestamps()
})
```

## Modifying Tables

```go
err := db.Schema().Table("users", func(table neat.Blueprint) {
    table.String("phone").Nullable()
    table.DropColumn("old_column")
})
```

## Dropping Tables

```go
// Drop table
err := db.Schema().Drop("users")

// Drop if exists
err := db.Schema().DropIfExists("users")
```

## Renaming Tables

```go
err := db.Schema().Rename("users", "accounts")
```

## Column Types

### String

```go
table.String("name")
table.String("name", 100) // with length
```

### Integer

```go
table.Integer("age")
table.BigInteger("id")
```

### Boolean

```go
table.Boolean("active")
```

### DateTime

```go
table.DateTime("created_at")
table.Timestamp("updated_at")
table.Date("birth_date")
table.Time("start_time")
```

### Text

```go
table.Text("description")
```

### Decimal

```go
table.Decimal("price", 10, 2)
```

### JSON

```go
table.JSON("metadata")
```

## Column Modifiers

### Unique

```go
table.String("email").Unique()
```

### Nullable

```go
table.String("phone").Nullable()
```

### Default

```go
table.Boolean("active").Default(true)
```

### Index

```go
table.String("email").Index()
```

## Indexes

```go
// Single column index
table.Index("email")

// Composite index
table.Index(["first_name", "last_name"])

// Unique index
table.UniqueIndex("email")
```

## Foreign Keys

```go
table.ForeignKey("user_id").References("id").On("users")
```

## Timestamps

```go
table.Timestamps() // created_at and updated_at
table.SoftDeletes() // deleted_at
```

## Checking Table/Column Existence

```go
// Check if table exists
exists := db.Schema().HasTable("users")

// Check if column exists
hasColumn := db.Schema().HasColumn("users", "email")
```

## Getting Table/Column Listing

```go
// Get all tables
tables := db.Schema().GetTableListing()

// Get columns in table
columns := db.Schema().GetColumnListing("users")

// Get indexes in table
indexes := db.Schema().GetIndexListing("users")
```

## Note

This documentation is a placeholder and will be expanded once the schema builder API is fully implemented with the config adapter.
