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

## Additional Column Types

### Binary

```go
table.Binary("data")
```

### Enum

```go
table.Enum("status", []string{"active", "inactive", "pending"})
```

### UUID

```go
table.UUID("id")
```

### IP Address

```go
table.IPAddress("ip")
```

### MAC Address

```go
table.MacAddress("mac")
```

## Additional Column Modifiers

### Primary Key

```go
table.Integer("id").Primary()
```

### Auto Increment

```go
table.Integer("id").AutoIncrement()
```

### Comment

```go
table.String("name").Comment("User's full name")
```

### After (MySQL)

```go
table.String("phone").After("email")
```

### Default Value (Raw)

```go
table.Timestamp("created_at").Default("CURRENT_TIMESTAMP")
```

### Unsigned

```go
table.Integer("age").Unsigned()
```

## Additional Index Operations

### Drop Index

```go
table.DropIndex("email")
```

### Rename Index

```go
table.RenameIndex("old_name", "new_name")
```

## Foreign Key Constraints

### On Delete

```go
table.ForeignKey("user_id").References("id").On("users").OnDelete("CASCADE")
```

### On Update

```go
table.ForeignKey("user_id").References("id").On("users").OnUpdate("CASCADE")
```

## Engine and Charset (MySQL)

```go
err := db.Schema().Create("users", func(table neat.Blueprint) {
    table.ID()
    table.String("name")
    table.Timestamps()
}).Engine("InnoDB").Charset("utf8mb4")
```

## Best Practices

1. **Use descriptive column names**: Make column names clear and consistent
2. **Add indexes strategically**: Index columns used in WHERE, JOIN, and ORDER BY clauses
3. **Use appropriate data types**: Choose the smallest data type that fits your needs
4. **Set default values**: Use defaults for columns that should have a value
5. **Use NOT NULL**: Mark columns as NOT NULL when they must have a value
6. **Add foreign keys**: Use foreign keys to enforce referential integrity
7. **Use timestamps**: Add created_at and updated_at for audit trails
8. **Consider soft deletes**: Use deleted_at instead of hard deletes for data recovery

## Troubleshooting

### Table creation fails
- Check if the table already exists (use DropIfExists first)
- Verify column names don't conflict with reserved keywords
- Ensure foreign key references exist

### Foreign key errors
- Verify the referenced table and column exist
- Check data types match between columns
- Ensure the referenced column is indexed

### Index issues
- Some databases have limits on index length
- Composite indexes have size limitations
- Too many indexes can slow down writes

## API Reference

### Schema Methods

- `Create(table string, callback func(Blueprint)) error` - Create a new table
- `Table(table string, callback func(Blueprint)) error` - Modify an existing table
- `Drop(table string) error` - Drop a table
- `DropIfExists(table string) error` - Drop a table if it exists
- `Rename(from, to string) error` - Rename a table
- `HasTable(table string) bool` - Check if table exists
- `HasColumn(table, column string) bool` - Check if column exists
- `GetTableListing() []string` - Get all table names
- `GetColumnListing(table string) []string` - Get column names for a table
- `GetIndexListing(table string) []string` - Get index names for a table

### Blueprint Column Methods

- `ID()` - Add auto-incrementing ID column
- `String(name string, length ...int) ColumnDefinition` - Add VARCHAR column
- `Integer(name string) ColumnDefinition` - Add INTEGER column
- `BigInteger(name string) ColumnDefinition` - Add BIGINT column
- `Boolean(name string) ColumnDefinition` - Add BOOLEAN column
- `DateTime(name string) ColumnDefinition` - Add DATETIME column
- `Timestamp(name string) ColumnDefinition` - Add TIMESTAMP column
- `Date(name string) ColumnDefinition` - Add DATE column
- `Time(name string) ColumnDefinition` - Add TIME column
- `Text(name string) ColumnDefinition` - Add TEXT column
- `Decimal(name string, precision, scale int) ColumnDefinition` - Add DECIMAL column
- `JSON(name string) ColumnDefinition` - Add JSON column
- `Binary(name string) ColumnDefinition` - Add BINARY column
- `Enum(name string, values []string) ColumnDefinition` - Add ENUM column
- `UUID(name string) ColumnDefinition` - Add UUID column
- `IPAddress(name string) ColumnDefinition` - Add IP address column
- `MacAddress(name string) ColumnDefinition` - Add MAC address column

### Blueprint Modifiers

- `Nullable() ColumnDefinition` - Allow NULL values
- `Unique() ColumnDefinition` - Add unique constraint
- `Default(value any) ColumnDefinition` - Set default value
- `Index() ColumnDefinition` - Add index
- `Primary() ColumnDefinition` - Set as primary key
- `AutoIncrement() ColumnDefinition` - Set as auto-increment
- `Comment(comment string) ColumnDefinition` - Add comment
- `After(column string) ColumnDefinition` - Place after another column (MySQL)
- `Unsigned() ColumnDefinition` - Set as unsigned (numeric types)

### Blueprint Index Methods

- `Index(columns ...string)` - Add index
- `UniqueIndex(columns ...string)` - Add unique index
- `DropIndex(name string)` - Drop index
- `RenameIndex(from, to string)` - Rename index

### Blueprint Foreign Key Methods

- `ForeignKey(column string) ForeignKeyDefinition` - Add foreign key
- `On(table string) ForeignKeyDefinition` - Set referenced table
- `References(column string) ForeignKeyDefinition` - Set referenced column
- `OnDelete(action string) ForeignKeyDefinition` - Set ON DELETE action
- `OnUpdate(action string) ForeignKeyDefinition` - Set ON UPDATE action

### Blueprint Helper Methods

- `Timestamps()` - Add created_at and updated_at timestamps
- `SoftDeletes()` - Add deleted_at timestamp for soft deletes
- `DropColumn(name string)` - Drop a column
- `RenameColumn(from, to string)` - Rename a column
