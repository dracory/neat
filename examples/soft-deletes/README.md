# Soft Deletes Example

This example demonstrates the NULL-based soft delete strategy, the traditional approach for implementing soft deletes using NULL values to indicate active records.

## Overview

The NULL-based soft delete strategy uses NULL to indicate "not deleted" records and a timestamp to indicate "soft deleted" records. Records are considered:
- **Active** when `soft_deleted_at IS NULL`
- **Soft Deleted** when `soft_deleted_at IS NOT NULL` (contains the deletion timestamp)

## Features Demonstrated

- Using `SoftDeletes` embed for NULL-based soft deletes
- Creating tables with nullable `soft_deleted_at` columns
- Soft deleting records (sets `soft_deleted_at = NOW()`)
- Querying active records (default, excludes soft deleted)
- Including soft deleted records with `WithSoftDeleted()`
- Querying only soft deleted records with `OnlySoftDeleted()`
- Restoring soft deleted records (sets `deleted_at = NULL`)
- Force deleting records permanently with `ForceDelete()`

## Benefits of NULL-Based Strategy

1. **Simple and Intuitive**: NULL = active, non-NULL = deleted is easy to understand
2. **Widely Used**: Common pattern in many ORMs (Laravel, Django, etc.)
3. **Easy to Debug**: NULL values are immediately recognizable in database tools
4. **No Special Values**: No need for sentinel values or magic dates

## Usage

### Model Definition

```go
import "github.com/dracory/neat/database/soft_delete"

type Product struct {
    soft_delete.SoftDeletes  // Uses NULL-based strategy
    ID          uint    `json:"id" db:"id"`
    Name        string  `json:"name" db:"name"`
    Price       float64 `json:"price" db:"price"`
}
```

### Schema Definition

```go
import "github.com/dracory/neat/database/schema/constants"

db.Schema().Create("products", func(blueprint schema.Blueprint) {
    blueprint.ID()
    blueprint.String("name")
    blueprint.Float("price", 10, 2)
    // For NULL-based strategy, soft_deleted_at should be nullable
    blueprint.Timestamp(constants.SoftDeleteAtColumn).Nullable()
})
```

### Basic Operations

```go
// Create a product (deleted_at is NULL by default)
db.Query().Table("products").Create(map[string]any{
    "name": "Laptop",
    "price": 999.99,
})

// Soft delete (sets deleted_at = NOW())
db.Query().Model(&Product{}).Where("name = ?", "Laptop").Delete()

// Query active products (excludes soft deleted - default behavior)
var products []Product
db.Query().Model(&Product{}).Get(&products)

// Include soft deleted products
db.Query().Model(&Product{}).WithSoftDeleted().Get(&products)

// Only soft deleted products
db.Query().Model(&Product{}).OnlySoftDeleted().Get(&products)

// Restore a soft deleted product (sets deleted_at = NULL)
db.Query().Model(&Product{}).Where("id = ?", 1).RestoreSoftDeleted()

// Permanently delete (bypasses soft delete)
db.Query().Model(&Product{}).Where("id = ?", 1).ForceDelete()
```

### Checking Soft Delete Status

```go
var product Product
db.Query().Model(&Product{}).First(&product)

if product.IsSoftDeleted() {
    fmt.Println("This product is soft deleted")
    fmt.Printf("Soft deleted at: %v\n", product.SoftDeletedAt)
}
```


## Using Constants

The neat package provides constants for all default column names to avoid hardcoding:

- `constants.SoftDeleteAtColumn` - Default soft delete column name ("soft_deleted_at")
- `constants.DeletedAtColumnName` - Laravel-compatible column name ("deleted_at")  
- `constants.MaxSoftDeletedAtDefault` - Max-date sentinel value ("9999-12-31 23:59:59")
- `constants.DefaultIDColumn` - Default ID column name ("id")
- `constants.DefaultCreatedAtColumn` - Default created_at column name ("created_at")
- `constants.DefaultUpdatedAtColumn` - Default updated_at column name ("updated_at")

Import the constants package to use them:

```go
import "github.com/dracory/neat/database/schema/constants"
```

## Alternative: DeletedAt

For Laravel-compatible schemas that use `deleted_at` column name:

```go
type Post struct {
    soft_delete.DeletedAt  // Uses "deleted_at" column (Laravel-compatible)
    ID    uint
    Title string
}
```

## Running the Example

```bash
cd examples/soft-deletes
go run main.go
```

Or run the tests:

```bash
go test -v
```

## Comparison with Max-Date Strategy

| Aspect | NULL-Based (SoftDeletes) | Max-Date (SoftDeletesMaxDate) |
|--------|--------------------------|-------------------------------|
| "Not deleted" value | NULL | `9999-12-31 23:59:59` |
| Column name | `soft_deleted_at` | `soft_deleted_at` |
| Column constraint | Nullable | NOT NULL (recommended) |
| Default value | NULL | `9999-12-31 23:59:59` |
| Active query condition | `IS NULL` | `> NOW()` |
| Deleted query condition | `IS NOT NULL` | `<= NOW()` |
| Index efficiency | Varies by database | Better range scans |
| NOT NULL compatible | No | Yes |
| Simplicity | Very simple | Slightly more complex |

## When to Use NULL-Based Strategy

- When you need simplicity and clarity
- When NOT NULL constraints are not required
- When you want the most common and widely-understood pattern
- When database index performance is not a critical concern

## When to Consider Max-Date Strategy

- When you need NOT NULL constraints on the soft delete column
- When you need better index performance on large datasets
- When working with databases that have poor NULL index handling
- When you want to avoid NULL handling complexity in queries
