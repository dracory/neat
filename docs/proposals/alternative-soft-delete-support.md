# Proposal: Alternative Soft Delete Support

**Status**: Proposed
**Date**: June 1, 2026
**Author**: Neat ORM Team

## Summary

Add an alternative soft delete implementation that uses a maximum date time (9999-12-31 23:59:59) as the default value for non-deleted records, instead of NULL. Records are considered soft-deleted when their `deleted_at` timestamp is in the past (before current time), and not deleted when `deleted_at` is in the future.

## Motivation

The current soft delete implementation uses NULL for non-deleted records and a timestamp for deleted records. While this is a common pattern (used by Laravel Eloquent, Django, etc.), it has some limitations:

1. **NULL handling complexity**: Queries must explicitly handle NULL values with `IS NULL` or `IS NOT NULL` conditions
2. **Index inefficiency**: NULL values can complicate indexing strategies in some databases
3. **Alternative patterns**: Some systems prefer using a sentinel value (like max date) to avoid NULLs entirely
4. **Compatibility**: Some legacy systems or specific database schemas may not support NULL timestamps

The alternative approach using 9999-12-31 23:59:59 as a sentinel value offers:
- Simpler query logic (no NULL checks needed)
- Better index utilization in some database engines
- Compatibility with systems that avoid NULLs
- Clear semantic meaning: "not deleted until the end of time"
- Time-based logic: future = not deleted, past = deleted

## Proposed Changes

### 1. Add Alternative Soft Delete Struct

Add a new struct `SoftDeletesMaxDate` alongside the existing `SoftDeletes`:

```go
type SoftDeletesMaxDate struct {
    DeletedAt time.Time `db:"deleted_at"`
}
```

**Default value**: 9999-12-31 23:59:59

### 2. Implement Soft Delete Logic

When deleting a record with `SoftDeletesMaxDate`:

```go
func (s *SoftDeletesMaxDate) SoftDelete() error {
    s.DeletedAt = time.Now()
    return nil
}
```


### 3. Implement Query Scopes

Add query scope methods for the alternative implementation:

```go
// Exclude soft-deleted records (deleted_at > now)
func (q *Query) WithoutDeletedMaxDate() *Query {
    return q.Where("deleted_at > ?", time.Now())
}

// Include all records (no filter needed)
func (q *Query) WithTrashedMaxDate() *Query {
    return q // No filtering applied
}

// Only soft-deleted records (deleted_at <= now)
func (q *Query) OnlyTrashedMaxDate() *Query {
    return q.Where("deleted_at <= ?", time.Now())
}
```


### 4. Implement Helper Methods

```go
func (s *SoftDeletesMaxDate) IsDeleted() bool {
    return s.DeletedAt.Before(time.Now()) || s.DeletedAt.Equal(time.Now())
}

func (s *SoftDeletesMaxDate) Restore() error {
    s.DeletedAt = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
    return nil
}
```


### 5. Database Migration Support

Add migration support for creating columns with default max date:

```go
// In schema builder
table.Timestamp("deleted_at").Default("9999-12-31 23:59:59")
```


## Usage Examples

### Model Definition

```go
type User struct {
    neat.SoftDeletesMaxDate  // Alternative implementation
    ID   uint
    Name string
}
```


### Soft Deleting

```go
var user User
db.Query().Where("id", 1).First(&user)
db.Query().Delete(&user) // Sets deleted_at to current time
```

### Querying Non-Deleted Records

```go
var users []User
db.Query().WithoutDeletedMaxDate().Get(&users)
// Equivalent to: WHERE deleted_at > NOW()
```

### Querying All Records

```go
var users []User
db.Query().WithTrashedMaxDate().Get(&users)
// No WHERE clause added
```

### Querying Only Deleted Records

```go
var users []User
db.Query().OnlyTrashedMaxDate().Get(&users)
// Equivalent to: WHERE deleted_at <= NOW()
```

### Restoring

```go
db.Query().Restore(&user) // Sets deleted_at back to 9999-12-31 23:59:59 (future)
```

### Checking Status

```go
if user.IsDeleted() {
    fmt.Println("User is soft-deleted")
}
```

## Benefits

1. **No NULL handling**: Eliminates NULL-related query complexity
2. **Better indexing**: Some databases index non-NULL values more efficiently
3. **Simpler queries**: No need for `IS NULL` or `IS NOT NULL` conditions
4. **Clear semantics**: Max date clearly represents "never deleted"
5. **Compatibility**: Works with systems that avoid NULLs
6. **Coexistence**: Can exist alongside the NULL-based implementation

## Drawbacks

1. **API surface expansion**: Two soft delete implementations to maintain
2. **Potential confusion**: Developers must choose between two approaches
3. **Database-specific**: Some databases may have limits on max date values
4. **Migration complexity**: Existing NULL-based implementations would need migration
5. **Less common**: NULL-based approach is more widely used and understood

## Alternatives Considered

### Alternative 1: Replace Current Implementation

Replace the NULL-based soft delete with the max date approach entirely.

**Pros**: Single implementation, simpler codebase
**Cons**: Breaking change, existing users would need to migrate data

**Decision**: Rejected to maintain backward compatibility

### Alternative 2: Boolean Flag

Use a boolean `is_deleted` flag instead of timestamps.

**Pros**: Simple, clear semantics
**Cons**: Loses audit trail information (when was it deleted?)

**Decision**: Rejected - timestamp is valuable for audit purposes

### Alternative 3: No Changes

Keep only the current NULL-based implementation.

**Pros**: Simpler API, less maintenance
**Cons**: Doesn't address the use cases for non-NULL soft deletes

**Decision**: Rejected to provide flexibility for different use cases

## Implementation Plan

1. Add `SoftDeletesMaxDate` struct to the soft deletes package
2. Implement soft delete logic (set to current time)
3. Implement query scope methods (`WithoutDeletedMaxDate`, `WithTrashedMaxDate`, `OnlyTrashedMaxDate`)
4. Implement helper methods (`IsDeleted`, `Restore`)
5. Add schema builder support for default max date
6. Update documentation with examples for both implementations
7. Add tests for the alternative implementation
8. Add migration guide for converting between implementations

## Open Questions

1. Should the max date be configurable (e.g., via struct tag or global setting)?
2. Should we add a naming convention to distinguish between the two implementations (e.g., `DeletedAtMaxDate` vs `DeletedAt`)?
3. Should we provide a migration tool to convert existing NULL-based soft deletes to max date format?
4. Should the default behavior be configurable at the database or query level?

## References

- Current Soft Deletes Documentation: `docs/soft-deletes.md`
- Laravel Eloquent Soft Deletes: https://laravel.com/docs/eloquent#soft-deleting
- Django Soft Deletes (django-softdelete): https://github.com/s-n-osophist/django-softdelete
- SQL NULL vs Default Values: https://use-the-index-luke.com/sql/where-clause/functions/is-null
