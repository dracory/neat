# DeletedAt Soft Delete Example

This example demonstrates the NULL-based soft delete strategy using the `deleted_at` column name, which is compatible with Laravel and other frameworks that use this convention.

## Overview

The `deleted_at` soft delete strategy uses NULL to indicate "not deleted" records and a timestamp to indicate "soft deleted" records. This is the traditional Laravel-compatible approach. Records are considered:
- **Active** when `deleted_at IS NULL`
- **Soft Deleted** when `deleted_at IS NOT NULL` (contains the deletion timestamp)

## Features Demonstrated

- Using `DeletedAt` embed for Laravel-compatible soft deletes
- Creating tables with nullable `deleted_at` columns
- Soft deleting records (sets `deleted_at = NOW()`)
- Querying active records (default, excludes soft deleted)
- Including soft deleted records with `WithSoftDeleted()`
- Querying only soft deleted records with `OnlySoftDeleted()`
- Restoring soft deleted records (sets `deleted_at = NULL`)
- Force deleting records permanently with `ForceDelete()`

## Benefits of deleted_at Strategy

1. **Laravel Compatibility**: Uses the same column name as Laravel's soft delete feature
2. **Easy Migration**: Simple to migrate existing Laravel applications to Go
3. **Familiar Pattern**: Laravel developers will recognize this pattern immediately
4. **Widely Adopted**: Common convention used by many PHP frameworks

## Usage

### Model Definition

```go
import "github.com/dracory/neat/database/soft_delete"

type Post struct {
    soft_delete.DeletedAt  // Uses "deleted_at" column (Laravel-compatible)
    ID      uint   `json:"id" db:"id"`
    Title   string `json:"title" db:"title"`
    Content string `json:"content" db:"content"`
}
```

### Schema Definition

```go
import "github.com/dracory/neat/database/schema/constants"

db.Schema().Create("posts", func(blueprint schema.Blueprint) {
    blueprint.ID()
    blueprint.String("title")
    blueprint.Text("content")
    // For Laravel compatibility, deleted_at should be nullable
    blueprint.Timestamp(constants.DeletedAtColumnName).Nullable()
})
```

### Basic Operations

```go
// Create a post (deleted_at is NULL by default)
db.Query().Table("posts").Create(map[string]any{
    "title":   "My First Post",
    "content": "Hello, world!",
})

// Soft delete (sets deleted_at = NOW())
db.Query().Model(&Post{}).Where("title = ?", "My First Post").Delete()

// Query active posts (excludes soft deleted - default behavior)
var posts []Post
db.Query().Model(&Post{}).Get(&posts)

// Include soft deleted posts
db.Query().Model(&Post{}).WithSoftDeleted().Get(&posts)

// Only soft deleted posts
db.Query().Model(&Post{}).OnlySoftDeleted().Get(&posts)

// Restore a soft deleted post (sets deleted_at = NULL)
db.Query().Model(&Post{}).Where("id = ?", 1).RestoreSoftDeleted()

// Permanently delete (bypasses soft delete)
db.Query().Model(&Post{}).Where("id = ?", 1).ForceDelete()
```

### Checking Soft Delete Status

```go
var post Post
db.Query().Model(&Post{}).First(&post)

if post.IsSoftDeleted() {
    fmt.Println("This post is soft deleted")
    fmt.Printf("Deleted at: %v\n", post.DeletedAt.DeletedAt)
}
```

### Column Name Verification

```go
var post Post
columnName := post.SoftDeletedAtColumn()
// Returns "deleted_at" for Laravel compatibility
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

## Comparison with Other Soft Delete Strategies

| Aspect | deleted_at (DeletedAt) | soft_deleted_at (SoftDeletes) | Max-Date (SoftDeletesMaxDate) |
|--------|------------------------|-------------------------------|-------------------------------|
| Column name | `deleted_at` | `soft_deleted_at` | `soft_deleted_at` |
| "Not deleted" value | NULL | NULL | `9999-12-31 23:59:59` |
| Column constraint | Nullable | Nullable | NOT NULL (recommended) |
| Default value | NULL | NULL | `9999-12-31 23:59:59` |
| Active query condition | `IS NULL` | `IS NULL` | `> NOW()` |
| Deleted query condition | `IS NOT NULL` | `IS NOT NULL` | `<= NOW()` |
| Laravel compatible | Yes | No | No |
| Index efficiency | Varies by database | Varies by database | Better range scans |
| NOT NULL compatible | No | No | Yes |

## When to Use deleted_at Strategy

- When migrating from Laravel or other PHP frameworks
- When you need Laravel schema compatibility
- When working with existing databases using `deleted_at`
- When your team is familiar with Laravel conventions
- When you want to maintain consistency with Laravel-based systems

## When to Consider Other Strategies

- **soft_deleted_at (SoftDeletes)**: When you prefer more explicit column naming
- **max-date (SoftDeletesMaxDate)**: When you need NOT NULL constraints or better index performance

## Running the Example

```bash
cd examples/deleted-at-soft-delete
go run main.go
```

Or run the tests:

```bash
go test -v
```

## Laravel Migration Example

If you're migrating from Laravel, your existing schema might look like:

```php
// Laravel migration
Schema::create('posts', function (Blueprint $table) {
    $table->id();
    $table->string('title');
    $table->text('content');
    $table->string('author');
    $table->timestamps();
    $table->softDeletes(); // Creates deleted_at column
});
```

The equivalent in neat:

```go
db.Schema().Create("posts", func(blueprint schema.Blueprint) {
    blueprint.ID()
    blueprint.String("title")
    blueprint.Text("content")
    blueprint.String("author")
    blueprint.Timestamps() // created_at, updated_at
    blueprint.Timestamp("deleted_at").Nullable()
})
```

## Data Consistency

When using this strategy with existing Laravel applications:

1. **Column Name**: Ensure your Go models use `soft_delete.DeletedAt` to match Laravel's `deleted_at` column
2. **Timestamp Format**: Both Laravel and neat use similar timestamp formats (RFC3339/ISO8601)
3. **Query Behavior**: The soft delete query scopes work identically to Laravel's global scopes
4. **Restore Logic**: Both frameworks set the column to NULL when restoring records

## Additional Resources

- See [soft-deletes](../soft-deletes/) for the standard `soft_deleted_at` column approach
- See [max-date-soft-delete](../max-date-soft-delete/) for the max-date sentinel strategy
- Check the [soft_delete package documentation](https://github.com/dracory/neat) for more details
