# Proposal: Alternative Soft Delete Column Naming

**Status**: Proposed
**Date**: June 1, 2026
**Author**: Neat ORM Team

## Summary

Add support for alternative column naming for soft delete fields, allowing users to use `soft_deleted_at` instead of `deleted_at` for clearer semantics. This provides better clarity by explicitly indicating the field is for soft deletion rather than hard deletion.

## Motivation

The current soft delete implementation uses `deleted_at` as the column name. While this is a common pattern (used by Laravel Eloquent, etc.), it can be ambiguous:

1. **Semantic clarity**: `deleted_at` doesn't distinguish between soft and hard deletion
2. **Team preferences**: Some teams prefer explicit naming conventions
3. **Legacy systems**: Some existing schemas may use `soft_deleted_at`
4. **Documentation clarity**: Makes it immediately obvious the field is for soft deletion

The alternative approach using `soft_deleted_at` offers:
- Clearer semantics - explicitly indicates soft deletion
- Better self-documenting code
- Compatibility with existing schemas that use this naming
- Easier onboarding for new team members
- Tag-free customization - uses method override instead of struct tags

## Proposed Changes

### 1. Add Column Name Customization Method

Enhance the soft delete structs to support a `DeletedAtColumn()` method that allows customizing the column name. Currently, the column name is determined by a constant. The method-based approach provides flexibility:

```go
type SoftDeletes struct {
    DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

func (s *SoftDeletes) DeletedAtColumn() string {
    return "deleted_at"
}
```

Users can override this in their models to use a different column name:

```go
type User struct {
    neat.SoftDeletes
    ID   uint `db:"id"`
    Name string `db:"name"`
}

func (u *User) DeletedAtColumn() string {
    return "soft_deleted_at"
}
```

**Note**: Neat ORM uses `db` tags for column mapping (e.g., `db:"id"`), but the current `SoftDeletes` implementation relies on a constant rather than a `db` tag. This proposal adds method-based customization as an alternative.

### 2. Add Built-in Alternative Struct

Provide a pre-configured struct that uses `soft_deleted_at` by default:

```go
type SoftDeletesAlt struct {
    SoftDeletedAt *time.Time `json:"soft_deleted_at,omitempty"`
}

func (s *SoftDeletesAlt) DeletedAtColumn() string {
    return "soft_deleted_at"
}
```

### 3. Update Query Scopes to Use Column Name Method

Update all query scope methods to use the `DeletedAtColumn()` method instead of hardcoded column names:

```go
func (q *Query) WithoutDeleted() *Query {
    column := q.getModel().DeletedAtColumn()
    return q.Where(column + " IS NULL")
}

func (q *Query) WithTrashed() *Query {
    return q // No filtering applied
}

func (q *Query) OnlyTrashed() *Query {
    column := q.getModel().DeletedAtColumn()
    return q.Where(column + " IS NOT NULL")
}
```

### 4. Update Helper Methods

Update helper methods to work with the customizable column name:

```go
func (s *SoftDeletes) IsDeleted() bool {
    return s.DeletedAt != nil
}

func (s *SoftDeletesAlt) IsDeleted() bool {
    return s.SoftDeletedAt != nil
}

func (s *SoftDeletes) Restore() error {
    s.DeletedAt = nil
    return nil
}

func (s *SoftDeletesAlt) Restore() error {
    s.SoftDeletedAt = nil
    return nil
}
```

### 5. Database Migration Support

Add migration support for creating columns with the alternative name:

```go
// In schema builder
table.Timestamp("soft_deleted_at").Nullable()
```

## Usage Examples

### Using Default `deleted_at`

```go
type User struct {
    neat.SoftDeletes
    ID   uint
    Name string
}
```

### Using Custom Column Name via Method Override

```go
type User struct {
    neat.SoftDeletes
    ID   uint
    Name string
}

func (u *User) DeletedAtColumn() string {
    return "soft_deleted_at"
}
```

### Using Built-in Alternative Struct

```go
type User struct {
    neat.SoftDeletesAlt  // Uses soft_deleted_at by default
    ID   uint
    Name string
}
```

### Soft Deleting

```go
var user User
db.Query().Where("id", 1).First(&user)
db.Query().Delete(&user) // Sets soft_deleted_at to current time
```

### Querying Non-Deleted Records

```go
var users []User
db.Query().WithoutDeleted().Get(&users)
// Equivalent to: WHERE soft_deleted_at IS NULL
```

### Querying All Records

```go
var users []User
db.Query().WithTrashed().Get(&users)
// No WHERE clause added
```

### Querying Only Deleted Records

```go
var users []User
db.Query().OnlyTrashed().Get(&users)
// Equivalent to: WHERE soft_deleted_at IS NOT NULL
```

### Restoring

```go
db.Query().Restore(&user) // Sets soft_deleted_at back to NULL
```

### Checking Status

```go
if user.IsDeleted() {
    fmt.Println("User is soft-deleted")
}
```

## Benefits

1. **Clearer semantics**: `soft_deleted_at` explicitly indicates soft deletion
2. **Self-documenting**: Code is more readable without needing documentation
3. **Flexibility**: Users can choose the naming convention that fits their team
4. **Backward compatible**: Existing `deleted_at` usage continues to work
5. **No breaking changes**: Both approaches can coexist
6. **Better onboarding**: New team members understand the schema faster

## Drawbacks

1. **API surface expansion**: Additional struct and methods to maintain
2. **Potential confusion**: Two ways to do the same thing
3. **Documentation complexity**: Need to document both approaches
4. **Inconsistency risk**: Different models might use different naming conventions

## Alternatives Considered

### Alternative 1: Rename Default Column

Change the default from `deleted_at` to `soft_deleted_at` for all users.

**Pros**: Consistent naming across all projects
**Cons**: Breaking change, existing users would need to migrate data

**Decision**: Rejected to maintain backward compatibility

### Alternative 2: Struct Tag Configuration

Use struct tags to configure the column name:

```go
type SoftDeletes struct {
    DeletedAt *time.Time `db:"deleted_at" soft_delete_column:"soft_deleted_at"`
}
```

**Pros**: Configuration in one place
**Cons**: More complex implementation, less Go-idiomatic

**Decision**: Rejected in favor of method override pattern

### Alternative 3: No Changes

Keep only `deleted_at` as the column name.

**Pros**: Simpler API, less maintenance
**Cons**: Doesn't address the clarity concerns

**Decision**: Rejected to provide flexibility for different team preferences

## Implementation Plan

1. Add `DeletedAtColumn()` method to existing `SoftDeletes` struct
2. Add `SoftDeletesAlt` struct with `soft_deleted_at` field
3. Update query scope methods to use `DeletedAtColumn()` method
4. Update helper methods to work with both field names
5. Add schema builder support for `soft_deleted_at` column
6. Update documentation with examples for both approaches
7. Add tests for the alternative implementation
8. Add migration guide for converting between column names

## Open Questions

1. Should we provide more alternative structs (e.g., `archived_at`, `removed_at`)?
2. Should the column name be configurable via a global setting?
3. Should we provide a migration tool to rename existing `deleted_at` columns to `soft_deleted_at`?
4. Should we add validation to ensure the column name exists in the database schema?

## References

- Current Soft Deletes Documentation: `docs/soft-deletes.md`
- Laravel Eloquent Soft Deletes: https://laravel.com/docs/eloquent#soft-deleting
- Django Soft Deletes (django-softdelete): https://github.com/s-n-osophist/django-softdelete
- Naming Conventions: https://github.com/golang/go/wiki/CodeReviewComments#variable-names
