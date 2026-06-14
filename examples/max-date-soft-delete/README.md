# Max-Date Sentinel Soft Delete Example

This example demonstrates the max-date sentinel soft delete strategy, an alternative to NULL-based soft deletes that offers better compatibility with NOT NULL column constraints and improved index performance.

## Overview

The max-date sentinel strategy uses a far-future timestamp (`9999-12-31 23:59:59 UTC`) as the "not deleted" value instead of NULL. Records are considered:
- **Active** when `soft_deleted_at > NOW()` (in the future, i.e., the sentinel value)
- **Soft Deleted** when `soft_deleted_at <= NOW()` (in the past or present)

## Features Demonstrated

- Using `SoftDeletesMaxDate` embed for max-date sentinel soft deletes
- Creating tables with NOT NULL `soft_deleted_at` columns
- Soft deleting records (sets `soft_deleted_at = NOW()`)
- Querying active records (default, excludes soft deleted)
- Including soft deleted records with `WithSoftDeleted()`
- Querying only soft deleted records with `OnlySoftDeleted()`
- Restoring soft deleted records (sets `soft_deleted_at = MaxSoftDeletedAt`)
- Force deleting records permanently with `ForceDelete()`

## Benefits of Max-Date Strategy

1. **NOT NULL Compatibility**: Works with schemas that enforce NOT NULL on timestamp columns
2. **Better Index Performance**: Range scans (`> NOW()`) vs `IS NULL` lookups
3. **Simpler Queries**: No NULL handling complexity
4. **Database Portability**: Some databases handle NULLs differently in indexes

## Usage

### Model Definition

```go
import "github.com/dracory/neat/database/soft_delete"

type Product struct {
    soft_delete.SoftDeletesMaxDate  // Uses max-date sentinel strategy
    ID          uint    `json:"id" db:"id"`
    Name        string  `json:"name" db:"name"`
    Price       float64 `json:"price" db:"price"`
}
```

### Schema Definition

```go
db.Schema().Create("products", func(blueprint schema.Blueprint) {
    blueprint.ID()
    blueprint.String("name")
    blueprint.Float("price", 10, 2)
    // For max-date strategy, set default to the sentinel value
    blueprint.Timestamp("soft_deleted_at").Default("9999-12-31 23:59:59")
})
```

### Basic Operations

```go
// Create a product (soft_deleted_at automatically set to MaxSoftDeletedAt)
db.Query().Table("products").Create(map[string]any{
    "name": "Laptop",
    "price": 999.99,
})

// Soft delete (sets soft_deleted_at = NOW())
db.Query().Model(&Product{}).Where("name = ?", "Laptop").Delete()

// Query active products (excludes soft deleted - default behavior)
var products []Product
db.Query().Model(&Product{}).Get(&products)

// Include soft deleted products
db.Query().Model(&Product{}).WithSoftDeleted().Get(&products)

// Only soft deleted products
db.Query().Model(&Product{}).OnlySoftDeleted().Get(&products)

// Restore a soft deleted product (sets soft_deleted_at = MaxSoftDeletedAt)
db.Query().Model(&Product{}).Where("id = ?", 1).RestoreSoftDeleted()

// Permanently delete (bypasses soft delete)
db.Query().Model(&Product{}).Where("id = ?", 1).ForceDelete()
```

## Sentinel Value

The max-date sentinel value is exposed as a constant:

```go
import "github.com/dracory/neat/database/soft_delete"

// MaxSoftDeletedAt = 9999-12-31 23:59:59 UTC
fmt.Println(soft_delete.MaxSoftDeletedAt)
```

## Alternative: DeletedAtMaxDate

For Laravel-compatible schemas that use `deleted_at` column name:

```go
type Post struct {
    soft_delete.DeletedAtMaxDate  // Uses "deleted_at" column
    ID    uint
    Title string
}
```

## Running the Example

```bash
cd examples/max-date-soft-delete
go run main.go
```

Or run the tests:

```bash
go test -v
```

## Comparison with NULL-Based Strategy

| Aspect | NULL-Based (SoftDeletes) | Max-Date (SoftDeletesMaxDate) |
|--------|--------------------------|-------------------------------|
| "Not deleted" value | NULL | `9999-12-31 23:59:59` |
| Column constraint | Nullable | NOT NULL (recommended) |
| Default value | NULL | `9999-12-31 23:59:59` |
| Active query condition | `IS NULL` | `> NOW()` |
| Deleted query condition | `IS NOT NULL` | `<= NOW()` |
| Index efficiency | Varies by database | Better range scans |
| NOT NULL compatible | No | Yes |
