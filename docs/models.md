# Models and Struct Tags

This document explains how Neat ORM maps Go structs to database tables and columns.

## Important: Two Separate Concepts

Neat has **two separate systems** for working with database tables:

1. **Struct Tags** - Map existing database columns to Go struct fields
2. **Schema Builder** - Define and create database table structures

**Struct tags do NOT define database types.** They only tell Neat which database column maps to which struct field.

To create tables with specific column types (VARCHAR, TEXT, INT, etc.), use the **Schema Builder** in migrations.

## Table Names

Neat determines table names in the following order:

1. **TableName() method** - If your model implements `TableName() string`, that method is used
2. **Automatic naming** - The struct name is converted to snake_case and pluralized

### Custom Table Name

```go
type User struct {
    ID   uint
    Name string
}

func (User) TableName() string {
    return "app_users"
}
```

### Automatic Table Name

```go
type User struct {
    ID   uint
    Name string
}
// Maps to table: "users"

type UserProfile struct {
    ID     uint
    UserID uint
    Bio    string
}
// Maps to table: "user_profiles"
```

## Column Names

Neat determines column names by checking struct tags in this priority order:

1. `db` tag - used directly as column name
2. `neat` tag - used directly as column name
3. `gorm` tag - parses `column:name` format for GORM compatibility
4. **Automatic naming** - field name converted to snake_case

### Using db Tags (Recommended)

```go
type User struct {
    ID        uint      `db:"id"`
    Name      string    `db:"name"`
    Email     string    `db:"email"`
    CreatedAt time.Time `db:"created_at"`
    UpdatedAt time.Time `db:"updated_at"`
}
```

### Using neat Tags

```go
type User struct {
    ID        uint      `neat:"id"`
    Name      string    `neat:"name"`
    Email     string    `neat:"email"`
}
```

### Using gorm Tags (For Compatibility)

```go
type User struct {
    ID        uint      `gorm:"primaryKey;column:id"`
    Name      string    `gorm:"size:255;column:name"`
    Email     string    `gorm:"uniqueIndex;column:email"`
}
```

### Automatic Column Names (No Tags)

```go
type User struct {
    ID        uint      // maps to "id"
    Name      string    // maps to "name"
    Email     string    // maps to "email"
    CreatedAt time.Time // maps to "created_at"
    UpdatedAt time.Time // maps to "updated_at"
    IsActive  bool      // maps to "is_active"
}
```

## How They Work Together

### Step 1: Create the Table (Schema Builder)

First, use a migration to define the table structure with explicit database types:

```go
// migration file
db.Schema().Create("users", func(table neat.Blueprint) {
    table.ID()                    // BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY
    table.String("name", 255)     // VARCHAR(255)
    table.String("email", 500)    // VARCHAR(500)
    table.Text("bio")             // TEXT
    table.Integer("age")          // INT
    table.Boolean("active")       // BOOLEAN
    table.DateTime("created_at")  // DATETIME
    table.DateTime("updated_at")  // DATETIME
})
```

### Step 2: Define the Struct (Struct Tags)

Then, create a Go struct that maps to those columns:

```go
type User struct {
    ID        uint      `db:"id"`         // maps to "id" column (BIGINT)
    Name      string    `db:"name"`       // maps to "name" column (VARCHAR)
    Email     string    `db:"email"`      // maps to "email" column (VARCHAR)
    Bio       string    `db:"bio"`        // maps to "bio" column (TEXT)
    Age       int       `db:"age"`        // maps to "age" column (INT)
    Active    bool      `db:"active"`     // maps to "active" column (BOOLEAN)
    CreatedAt time.Time `db:"created_at"` // maps to "created_at" column (DATETIME)
    UpdatedAt time.Time `db:"updated_at"` // maps to "updated_at" column (DATETIME)
}
```

### Key Points

- **Go types** (`uint`, `string`, `int`, `bool`, `time.Time`) are for Go code only
- **Database types** (`BIGINT`, `VARCHAR`, `TEXT`, `INT`, `BOOLEAN`, `DATETIME`) are defined in Schema Builder
- **Struct tags** (`db:"name"`) just connect the two - they don't define types
- Neat handles type conversion between Go and database types automatically

## Complete Examples

### Simple Model with Tags

```go
type User struct {
    ID        uint      `db:"id"`
    Name      string    `db:"name"`
    Email     string    `db:"email"`
    Age       int       `db:"age"`
    CreatedAt time.Time `db:"created_at"`
    UpdatedAt time.Time `db:"updated_at"`
}

// Table: users
// Columns: id, name, email, age, created_at, updated_at
```

### Model with Custom Table Name

```go
type Customer struct {
    ID        uint      `db:"customer_id"`
    Name      string    `db:"customer_name"`
    Email     string    `db:"customer_email"`
}

func (Customer) TableName() string {
    return "customers"
}

// Table: customers
// Columns: customer_id, customer_name, customer_email
```

### Model with Mixed Naming

```go
type UserProfile struct {
    ID           uint      `db:"id"`
    UserID       uint      `db:"user_id"`
    DisplayName  string    `db:"display_name"`
    Bio          string    `db:"bio"`
    AvatarURL    string    `db:"avatar_url"`
    IsActive     bool      `db:"is_active"`
    LastLoginAt  time.Time `db:"last_login_at"`
}

// Table: user_profiles
// Columns: id, user_id, display_name, bio, avatar_url, is_active, last_login_at
```

### Model Without Tags (Automatic)

```go
type Product struct {
    ID          uint
    Name        string
    Description string
    Price       float64
    Stock       int
    CategoryID  uint
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// Table: products
// Columns: id, name, description, price, stock, category_id, created_at, updated_at
```

### Model with GORM Compatibility

```go
type Order struct {
    ID        uint      `gorm:"primaryKey;column:id"`
    UserID    uint      `gorm:"column:user_id;index"`
    Total     float64   `gorm:"column:total"`
    Status    string    `gorm:"column:status;size:50"`
    CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
    UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

// Table: orders
// Columns: id, user_id, total, status, created_at, updated_at
```

## Primary Keys

Neat automatically looks for fields named `ID` or `Id` as the primary key.

```go
type User struct {
    ID    uint // Primary key
    Name  string
    Email string
}
```

## Timestamps

Common convention for timestamp fields:

```go
type User struct {
    ID        uint
    Name      string
    CreatedAt time.Time `db:"created_at"`
    UpdatedAt time.Time `db:"updated_at"`
}
```

## Soft Deletes

For soft delete functionality, embed the `SoftDeletes` struct:

```go
import "github.com/dracory/neat/database/soft_delete"

type User struct {
    soft_delete.SoftDeletes
    ID    uint
    Name  string
    Email string
}

// Adds deleted_at column automatically
```

## Relationships

Relationships are defined using struct fields without special tags:

```go
type User struct {
    ID    uint
    Name  string
    Posts []Post // HasMany relationship
}

type Post struct {
    ID     uint
    Title  string
    UserID uint // Foreign key
    User   User // BelongsTo relationship
}
```

## Best Practices

1. **Use `db` tags** for explicit column naming - this is the recommended approach
2. **Keep field names in PascalCase** - Neat will convert to snake_case automatically
3. **Use `TableName()`** for non-standard table names
4. **Be consistent** - either use tags everywhere or rely on automatic naming
5. **Consider GORM compatibility** if migrating from GORM

## Tag Priority Summary

When Neat looks up a column name for a field:

1. Check `db` tag first
2. If not found, check `neat` tag
3. If not found, check `gorm` tag (parses `column:name`)
4. If not found, convert field name to snake_case

Example:
```go
type User struct {
    ID    uint `db:"user_id" neat:"uid" gorm:"column:id"`
    Name  string
}
// ID field maps to "user_id" (db tag takes priority)
// Name field maps to "name" (automatic snake_case)
```
