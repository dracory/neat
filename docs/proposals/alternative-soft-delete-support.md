# Proposal: Alternative Soft Delete Support (Max-Date Sentinel)

**Status**: Proposed
**Date**: June 1, 2026
**Author**: Neat ORM Team

## Summary

Add an alternative soft delete implementation that uses a maximum datetime sentinel (`9999-12-31 23:59:59 UTC`) as the "not deleted" value instead of NULL. Records are considered soft-deleted when their `deleted_at` timestamp is in the past (≤ current time), and active when `deleted_at` is in the future.

## Motivation

The current soft delete implementation uses `NULL` for active records and a past timestamp for deleted records. This is the most common pattern (Laravel Eloquent, Django, etc.), but it has real limitations:

1. **NULL handling complexity**: Queries must explicitly use `IS NULL` / `IS NOT NULL`
2. **Index inefficiency**: Many databases store NULLs outside B-tree indexes or require special `IS NULL`-aware indexes
3. **NOT NULL constraint conflicts**: Some schemas enforce `NOT NULL` on all timestamp columns
4. **Legacy compatibility**: Some existing systems use a sentinel value to avoid NULLs entirely

The max-date approach offers:
- Simpler range queries (`deleted_at > NOW()` instead of `IS NULL`)
- Better partial index support (`WHERE deleted_at > NOW()` is directly indexable)
- Compatible with `NOT NULL` column constraints
- Clear semantic: "deleted until the end of time means not deleted"

## Current State

### How soft deletes work today

`hasSoftDeleteCapability` in `database/query/query_delete.go` detects a soft-delete-capable model by reflection:

```go
deletedAtField := val.FieldByName("DeletedAt")
if deletedAtField.IsValid() && deletedAtField.Type() == reflect.TypeOf(&time.Time{}) {
    return true
}
```

**This checks for a field named `DeletedAt` of type `*time.Time` specifically.** A `time.Time` (non-pointer) field does not satisfy this check and will cause all deletes to fall through to hard deletes silently.

### Hardcoded column name in four places

The column name `"deleted_at"` is hardcoded in:

| File | Location | Usage |
|------|----------|-------|
| `database/query/builder_where.go` | `buildWheresWithSoftDelete()` | `IS NULL` / `IS NOT NULL` filter |
| `database/query/builder_where.go` | `buildWheresWithSoftDeleteIndex()` | Same, with placeholder index |
| `database/query/query_delete.go` | `Delete()` | Soft delete UPDATE sets `deleted_at` to `time.Now()` |
| `database/query/query_advanced.go` | `Restore()` | Restore UPDATE sets `deleted_at` to `nil` |

Additionally, `builder_update.go` detects whether an UPDATE is a soft-delete operation by checking `m["deleted_at"]` — a hardcoded key name.

### Restore() hardcodes NULL

`query_advanced.go` line 205:
```go
sql, args := builder.BuildUpdate(map[string]any{"deleted_at": nil})
```

For the max-date approach, restoring must set the column back to the sentinel value, not `nil`. This path needs to be aware of which soft-delete strategy the model uses.

## Proposed Changes

### 1. Add `SoftDeletesMaxDate` struct

Place it alongside `SoftDeletes` in `database/soft_delete/soft_delete.go`:

```go
// MaxDeletedAt is the sentinel "not deleted" value for the max-date soft delete strategy.
var MaxDeletedAt = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)

// SoftDeletesMaxDate provides soft delete functionality using a max-date sentinel
// instead of NULL. Embed this struct instead of SoftDeletes in models where
// NULL timestamps are undesirable or the schema enforces NOT NULL.
type SoftDeletesMaxDate struct {
    DeletedAt time.Time `json:"deleted_at,omitempty"`
}

func (s *SoftDeletesMaxDate) IsDeleted() bool {
    return !s.DeletedAt.After(time.Now())
}

func (s *SoftDeletesMaxDate) Delete() {
    s.DeletedAt = time.Now()
}

func (s *SoftDeletesMaxDate) Restore() {
    s.DeletedAt = MaxDeletedAt
}

func (s *SoftDeletesMaxDate) GetDeletedAt() time.Time {
    return s.DeletedAt
}
```

**Important**: `GetDeletedAt()` returns `time.Time` (not `*time.Time`) because the field is always set.

### 2. Define the `SoftDeleteStrategy` interface

The query builder needs to distinguish between the two strategies at runtime. Introduce an interface in `contracts/database/orm`:

```go
// SoftDeleteStrategy is implemented by models that use a non-NULL soft delete strategy.
// Models that do NOT implement this interface are assumed to use the NULL strategy (default).
type SoftDeleteStrategy interface {
    // SoftDeleteColumn returns the column name used for soft deletes.
    SoftDeleteColumn() string
    // SoftDeleteValue returns the value to set when soft-deleting (e.g. time.Now()).
    SoftDeleteValue() any
    // RestoreValue returns the value to set when restoring (e.g. MaxDeletedAt for max-date, nil for NULL).
    RestoreValue() any
    // IsDeletedCondition returns the SQL fragment and args for the "only deleted" filter.
    // e.g. ("deleted_at <= ?", []any{time.Now()}) for max-date strategy.
    IsDeletedCondition(quoteIdentifier func(string) string) (string, []any)
    // IsActiveCondition returns the SQL fragment and args for the "not deleted" filter.
    // e.g. ("deleted_at > ?", []any{time.Now()}) for max-date strategy.
    IsActiveCondition(quoteIdentifier func(string) string) (string, []any)
}
```

Implement this on `SoftDeletesMaxDate`:

```go
func (s *SoftDeletesMaxDate) SoftDeleteColumn() string { return "deleted_at" }
func (s *SoftDeletesMaxDate) SoftDeleteValue() any     { return time.Now() }
func (s *SoftDeletesMaxDate) RestoreValue() any        { return MaxDeletedAt }

func (s *SoftDeletesMaxDate) IsDeletedCondition(q func(string) string) (string, []any) {
    return q("deleted_at") + " <= ?", []any{time.Now()}
}
func (s *SoftDeletesMaxDate) IsActiveCondition(q func(string) string) (string, []any) {
    return q("deleted_at") + " > ?", []any{time.Now()}
}
```

The existing NULL-based `SoftDeletes` does **not** implement `SoftDeleteStrategy`; the query builder falls back to the current NULL logic for backward compatibility.

### 3. Update `hasSoftDeleteCapability`

The current reflection-based check must handle both `*time.Time` (NULL strategy) and `time.Time` (max-date strategy). The cleanest fix is to check the interface instead of the field:

```go
func hasSoftDeleteCapability(model any) bool {
    if model == nil {
        return false
    }
    // NULL-based strategy: field named DeletedAt of type *time.Time
    val := reflect.ValueOf(model)
    if val.Kind() == reflect.Ptr {
        val = val.Elem()
    }
    if val.Kind() == reflect.Struct {
        f := val.FieldByName("DeletedAt")
        if f.IsValid() && f.Type() == reflect.TypeOf(&time.Time{}) {
            return true
        }
        // Also check embedded structs
        t := val.Type()
        for i := 0; i < t.NumField(); i++ {
            field := t.Field(i)
            if field.Anonymous && field.Type.Kind() == reflect.Struct {
                embedded := val.Field(i)
                ef := embedded.FieldByName("DeletedAt")
                if ef.IsValid() && ef.Type() == reflect.TypeOf(&time.Time{}) {
                    return true
                }
            }
        }
    }
    // Max-date strategy (or any custom strategy): implements SoftDeleteStrategy
    if _, ok := model.(SoftDeleteStrategy); ok {
        return true
    }
    return false
}
```

### 4. Update `Delete()` in `query_delete.go`

```go
if useSoftDelete && !q.withTrashed && !q.onlyTrashed {
    clone := q.Clone().(*Query)
    clone.withTrashed = true
    builder := NewBuilder(clone)

    var updateMap map[string]any
    if strat, ok := q.model.(SoftDeleteStrategy); ok {
        updateMap = map[string]any{strat.SoftDeleteColumn(): strat.SoftDeleteValue()}
    } else {
        updateMap = map[string]any{"deleted_at": time.Now()}
    }
    deleteSQL, args = builder.BuildUpdate(updateMap)
}
```

### 5. Update `Restore()` in `query_advanced.go`

```go
var updateMap map[string]any
if strat, ok := q.model.(SoftDeleteStrategy); ok {
    updateMap = map[string]any{strat.SoftDeleteColumn(): strat.RestoreValue()}
} else {
    updateMap = map[string]any{"deleted_at": nil}
}
sql, args := builder.BuildUpdate(updateMap)
```

### 6. Update `buildWheresWithSoftDelete` in `builder_where.go`

```go
func (b *Builder) buildWheresWithSoftDelete() (string, []any) {
    var prefix string
    var prefixArgs []any

    if hasSoftDeleteCapability(b.query.model) {
        if strat, ok := b.query.model.(SoftDeleteStrategy); ok {
            // Max-date (or custom) strategy: use range conditions
            switch {
            case b.query.onlyTrashed:
                prefix, prefixArgs = strat.IsDeletedCondition(b.quoteIdentifier)
            case b.query.withTrashed:
                // no filter
            default:
                prefix, prefixArgs = strat.IsActiveCondition(b.quoteIdentifier)
            }
        } else {
            // NULL-based strategy (default)
            col := b.quoteIdentifier("deleted_at")
            switch {
            case b.query.onlyTrashed:
                prefix = col + " IS NOT NULL"
            case b.query.withTrashed:
                // no filter
            default:
                prefix = col + " IS NULL"
            }
        }
    }

    if len(b.query.wheres) == 0 {
        return prefix, prefixArgs
    }
    base, args := b.buildWheres()
    if prefix == "" {
        return base, args
    }
    combined := append(prefixArgs, args...)
    return prefix + " AND " + base, combined
}
```

Apply the same pattern to `buildWheresWithSoftDeleteIndex`.

### 7. Update `isSoftDeleteOperation` detection in `builder_update.go`

Currently:
```go
if _, hasDeletedAt := m["deleted_at"]; hasDeletedAt {
    isSoftDeleteOperation = true
}
```

Replace with a check against the actual soft delete column for the model:

```go
softDeleteCol := "deleted_at"
if strat, ok := b.query.model.(SoftDeleteStrategy); ok {
    softDeleteCol = strat.SoftDeleteColumn()
}
if _, has := m[softDeleteCol]; has {
    isSoftDeleteOperation = true
}
```

### 8. Schema migration support

Use the schema builder to add a `NOT NULL` `deleted_at` column with the sentinel default:

```go
table.Timestamp("deleted_at").Default("9999-12-31 23:59:59").NotNullable()
```

Database-specific notes are in the Portability section below.

## Portability of 9999-12-31 23:59:59

| Database | Max DATETIME | Notes |
|----------|-------------|-------|
| MySQL / MariaDB | `9999-12-31 23:59:59` | Valid for `DATETIME`. Beware `TIMESTAMP` max is `2038-01-19`. |
| PostgreSQL | `9999-12-31 23:59:59` for `TIMESTAMP` | Store as `TIMESTAMPTZ` for TZ-aware comparisons. |
| SQLite | No hard limit | Stored as text; ISO 8601 comparison works correctly. |
| SQL Server | `9999-12-31 23:59:59.997` for `datetime`; `9999-12-31 23:59:59.9999999` for `datetime2` | Use `datetime2` to store the exact sentinel. |
| Oracle | `9999-12-31 23:59:59` for `DATE` / `TIMESTAMP` | Valid. |

**Always use `DATETIME` / `TIMESTAMP` (not the short-range `TIMESTAMP` in MySQL) and store all values in UTC.**

## Usage Examples

### Model definition

```go
type User struct {
    neat.SoftDeletesMaxDate  // uses deleted_at with max-date sentinel
    ID   uint
    Name string
}
```

### Creating the schema

```go
schema.Create("users", func(table blueprint.Blueprint) {
    table.ID()
    table.String("name")
    table.Timestamp("deleted_at").Default("9999-12-31 23:59:59").NotNullable()
    table.Timestamps()
})
```

When inserting a new record the ORM must populate `deleted_at` with `MaxDeletedAt`. This requires either:
- A `BeforeCreate` hook that sets the sentinel, or
- A database-level `DEFAULT` constraint (preferred).

### Soft deleting

```go
db.Query().Model(&User{}).Where("id", 1).Delete()
// UPDATE users SET deleted_at = '2026-06-13 10:00:00' WHERE id = 1
```

### Querying active records (default)

```go
var users []User
db.Query().Model(&User{}).Get(&users)
// WHERE deleted_at > '2026-06-13 10:00:00'
```

### Querying all records

```go
db.Query().Model(&User{}).WithTrashed().Get(&users)
// no WHERE clause added
```

### Querying only deleted records

```go
db.Query().Model(&User{}).OnlyTrashed().Get(&users)
// WHERE deleted_at <= '2026-06-13 10:00:00'
```

### Restoring

```go
db.Query().Model(&User{}).Where("id", 1).Restore()
// UPDATE users SET deleted_at = '9999-12-31 23:59:59' WHERE id = 1
```

### Checking status in application code

```go
if user.IsDeleted() {
    fmt.Println("User is soft-deleted")
}
```

## Benefits

1. **No NULL handling**: All queries use simple range comparisons
2. **Better indexing**: A B-tree index on `deleted_at` supports `deleted_at > NOW()` efficiently as a range scan
3. **NOT NULL compatible**: Works with strict schemas that disallow NULLs in timestamp columns
4. **Backward compatible**: Existing NULL-based models are unchanged
5. **Clean interface boundary**: `SoftDeleteStrategy` provides a well-defined extension point for future strategies

## Drawbacks

1. **Two implementations to maintain**: `SoftDeletes` and `SoftDeletesMaxDate` diverge and must evolve in parallel
2. **New record initialization**: Unlike NULL (the zero value for pointers), `MaxDeletedAt` must be explicitly set at insert time, requiring either a hook or a database default constraint
3. **`IsDeleted()` is time-sensitive**: The result depends on `time.Now()`. In tests or batch operations with a fixed clock, this must be mocked or accounted for
4. **Query conditions contain bind parameters**: `deleted_at > ?` with `time.Now()` adds a bind parameter where `IS NULL` adds none — this affects placeholder index counting in `buildWheresWithSoftDeleteIndex`; the implementation must pass the extra arg through correctly
5. **Less common**: NULL-based soft deletes are the dominant convention; developers unfamiliar with max-date may find the queries surprising
6. **Database-specific max-date limits**: See portability table above

## Alternatives Considered

### Alternative 1: Replace the NULL implementation

Replace the NULL-based implementation globally with max-date.

**Pros**: Single implementation, no ambiguity
**Cons**: Breaking change; every existing user must migrate schema and data

**Decision**: Rejected to maintain backward compatibility

### Alternative 2: Boolean `is_deleted` flag

Use a boolean column rather than a timestamp.

**Pros**: Simple; no NULL or sentinel issues; easy indexing
**Cons**: Loses the deletion timestamp, which is essential for audit trails and time-based queries

**Decision**: Rejected — the timestamp is valuable and both soft delete strategies preserve it

### Alternative 3: Separate `deleted_at` (NULL-based) + `is_deleted` (boolean)

Use both columns together: `deleted_at` for when, `is_deleted` for filtering.

**Pros**: Fast boolean index; keeps full audit timestamp
**Cons**: Data redundancy; two columns must stay in sync; more complex schema

**Decision**: Out of scope for this proposal

### Alternative 4: No changes

**Pros**: Zero maintenance cost, zero risk
**Cons**: Legitimate use cases for non-NULL soft deletes remain unaddressed; the NULL-based approach cannot serve schemas with `NOT NULL` constraints

**Decision**: Rejected

## Implementation Plan

1. Add `MaxDeletedAt` sentinel constant and `SoftDeletesMaxDate` struct to `database/soft_delete/soft_delete.go`
2. Define the `SoftDeleteStrategy` interface in `contracts/database/orm`
3. Implement `SoftDeleteStrategy` on `SoftDeletesMaxDate`
4. Update `hasSoftDeleteCapability` to detect `time.Time` (non-pointer) fields and `SoftDeleteStrategy` implementors
5. Update `Delete()` in `query_delete.go` to use `SoftDeleteStrategy.SoftDeleteValue()` when available
6. Update `Restore()` in `query_advanced.go` to use `SoftDeleteStrategy.RestoreValue()` when available
7. Update `buildWheresWithSoftDelete` and `buildWheresWithSoftDeleteIndex` in `builder_where.go` to use `IsActiveCondition`/`IsDeletedCondition`
8. Update `isSoftDeleteOperation` detection in `builder_update.go` to use `SoftDeleteStrategy.SoftDeleteColumn()`
9. Add tests covering:
   - `SoftDeletesMaxDate.IsDeleted()` with past, present, and future timestamps
   - `Delete()` and `Restore()` SQL generation for max-date models
   - `WithTrashed`, `OnlyTrashed`, `WithoutTrashed` (default) query filters for max-date models
   - Placeholder index correctness when max-date conditions carry bind args
10. Decide how new records get their initial `MaxDeletedAt` value (hook vs. database default)
11. Update `docs/soft-deletes.md` to document both strategies with schema examples

## Open Questions

1. **New record initialization**: Should the ORM automatically set `DeletedAt = MaxDeletedAt` during `Insert` when the model embeds `SoftDeletesMaxDate`, or is it the caller's responsibility to set a database-level `DEFAULT`? A missing default would cause the sentinel value to be the zero `time.Time` (year 0001), which `IsDeleted()` would immediately treat as deleted.
2. **Placeholder index in `buildWheresWithSoftDeleteIndex`**: The max-date conditions (`deleted_at > ?`) add a bind parameter. The index-tracking variant must correctly advance the placeholder counter. This is a non-trivial coordination point that deserves its own test matrix.
3. **`time.Now()` in `IsActiveCondition`/`IsDeletedCondition`**: Should these methods accept a `time.Time` argument (allowing callers to pass a fixed clock) rather than calling `time.Now()` internally? This would improve testability.
4. **`SoftDeleteStrategy` interface placement**: Should it live in `contracts/database/orm` (alongside the ORM interface) or in `database/soft_delete` (closer to the implementation)?
5. **Coexistence with column name customization**: This proposal hardcodes `"deleted_at"` inside `SoftDeletesMaxDate`. The column-naming proposal (see `docs/proposals/alternative-soft-delete-column-naming.md`) introduces `DeletedAtColumn()`. These two proposals must be reconciled — `SoftDeleteStrategy.SoftDeleteColumn()` can subsume `DeletedAtColumn()` and should be considered the unified solution.

## References

- Current Soft Deletes Documentation: `docs/soft-deletes.md`
- Column Naming Proposal: `docs/proposals/alternative-soft-delete-column-naming.md`
- Existing `SoftDeletes` implementation: `database/soft_delete/soft_delete.go`
- ORM model structs: `database/orm/model.go`
- Query soft-delete detection: `database/query/query_delete.go` (`hasSoftDeleteCapability`)
- Query builder WHERE construction: `database/query/builder_where.go`
- Restore implementation: `database/query/query_advanced.go` (`Restore`)
- Use The Index, Luke — IS NULL: https://use-the-index-luke.com/sql/where-clause/functions/is-null
- Laravel Eloquent Soft Deletes: https://laravel.com/docs/eloquent#soft-deleting
