# Proposal: Alternative Soft Delete Column Naming

**Status**: Proposed
**Date**: June 1, 2026
**Author**: Neat ORM Team

## Summary

Add support for customizing the soft delete column name, allowing users to use `soft_deleted_at` (or any other name) instead of the hardcoded `deleted_at`. This provides better semantic clarity for teams that prefer explicit naming conventions.

## Motivation

The current soft delete implementation hardcodes `deleted_at` throughout the query builder (`builder_where.go`). While this is a common convention (Laravel Eloquent, etc.), it creates friction in two scenarios:

1. **Semantic clarity**: `deleted_at` doesn't explicitly indicate soft deletion vs. hard deletion
2. **Legacy schemas**: Existing databases may already use `soft_deleted_at` or another column name
3. **Team conventions**: Some teams enforce explicit naming for all "soft" operations

The existing `docs/soft-deletes.md` already documents a `DeletedAtColumn() string` method override as the customization mechanism, but **this is not yet implemented** — `builder_where.go` hardcodes `"deleted_at"` regardless of any such method on the model.

## Current State

### Two separate `SoftDeletes` structs exist in the codebase

`database/soft_delete/soft_delete.go`:
```go
type SoftDeletes struct {
    DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
```

`database/orm/model.go`:
```go
type SoftDeletes struct {
    DeletedAt sql.NullTime `json:"deleted_at,omitempty"`
}
```

These differ in field type (`*time.Time` vs `sql.NullTime`). The proposal must clarify which package is the canonical one, and whether both need updating.

### Column name is hardcoded in four places

`database/query/builder_where.go` hardcodes `"deleted_at"` in:
- `buildWheresWithSoftDelete()` — `IS NULL` and `IS NOT NULL` conditions
- `buildWheresWithSoftDeleteIndex()` — same, with placeholder index variant

`database/query/query_delete.go` hardcodes `"deleted_at"` in:
- `Delete()` — soft delete UPDATE sets `"deleted_at"` directly

`database/query/builder_update.go` — hardcodes `"deleted_at"` in restore/update logic.

### `hasSoftDeleteCapability` detects by field name, not interface

`query_delete.go`:
```go
func hasSoftDeleteCapability(model any) bool {
    // checks for a field named "DeletedAt" of type *time.Time
    deletedAtField := val.FieldByName("DeletedAt")
    if deletedAtField.IsValid() && deletedAtField.Type() == reflect.TypeOf(&time.Time{}) {
        return true
    }
    // also checks embedded structs...
}
```

A model with a `SoftDeletedAt` field (named differently) would **not** be detected and soft deletes would silently fall back to hard deletes.

## Proposed Changes

### 1. Define a `SoftDeleteColumnNamer` interface

Introduce a single interface that any model can implement to override the column name:

```go
// SoftDeleteColumnNamer is implemented by models that customize the soft delete column.
type SoftDeleteColumnNamer interface {
    DeletedAtColumn() string
}
```

This interface should live in the `contracts/database/orm` package alongside other ORM contracts.

### 2. Add `DeletedAtColumn()` method to `SoftDeletes`

Update the canonical `SoftDeletes` struct (resolve the `*time.Time` vs `sql.NullTime` ambiguity first) to implement the interface with the default value:

```go
func (s *SoftDeletes) DeletedAtColumn() string {
    return "deleted_at"
}
```

This makes `SoftDeletes` itself implement `SoftDeleteColumnNamer`, so the query builder can call it uniformly.

### 3. Add a built-in `SoftDeletesAlt` struct

```go
// SoftDeletesAlt provides soft delete functionality using the soft_deleted_at column name.
type SoftDeletesAlt struct {
    SoftDeletedAt *time.Time `json:"soft_deleted_at,omitempty"`
}

func (s *SoftDeletesAlt) DeletedAtColumn() string {
    return "soft_deleted_at"
}

func (s *SoftDeletesAlt) IsDeleted() bool {
    return s.SoftDeletedAt != nil
}

func (s *SoftDeletesAlt) Delete() {
    now := time.Now()
    s.SoftDeletedAt = &now
}

func (s *SoftDeletesAlt) Restore() {
    s.SoftDeletedAt = nil
}

func (s *SoftDeletesAlt) GetDeletedAt() *time.Time {
    return s.SoftDeletedAt
}
```

### 4. Update `hasSoftDeleteCapability` to check by interface, not field name

Replace the current reflection-based field name check with an interface assertion:

```go
func hasSoftDeleteCapability(model any) bool {
    if model == nil {
        return false
    }
    _, ok := model.(SoftDeleteColumnNamer)
    return ok
}
```

This is simpler, more robust, and correctly handles any custom column name — including `SoftDeletesAlt`. Models that embed `SoftDeletes` or `SoftDeletesAlt` will satisfy the interface automatically via promoted methods.

### 5. Add a `getSoftDeleteColumn` helper

```go
// getSoftDeleteColumn returns the soft delete column name for the model,
// falling back to "deleted_at" if the model does not implement SoftDeleteColumnNamer.
func getSoftDeleteColumn(model any) string {
    if namer, ok := model.(SoftDeleteColumnNamer); ok {
        return namer.DeletedAtColumn()
    }
    return "deleted_at"
}
```

### 6. Update `builder_where.go` to use the helper

Replace all hardcoded `"deleted_at"` references:

```go
func (b *Builder) buildWheresWithSoftDelete() (string, []any) {
    var prefix string
    if hasSoftDeleteCapability(b.query.model) {
        col := b.quoteIdentifier(getSoftDeleteColumn(b.query.model))
        switch {
        case b.query.onlyTrashed:
            prefix = fmt.Sprintf("%s IS NOT NULL", col)
        case b.query.withTrashed:
            // include all rows — no filter
        default:
            prefix = fmt.Sprintf("%s IS NULL", col)
        }
    }
    // ... rest unchanged
}
```

Apply the same change to `buildWheresWithSoftDeleteIndex`.

### 7. Update `query_delete.go` to use the helper

```go
if useSoftDelete && !q.withTrashed && !q.onlyTrashed {
    col := getSoftDeleteColumn(q.model)
    deleteSQL, args = builder.BuildUpdate(map[string]any{col: now})
}
```

### 8. Schema migration support

No changes required. The schema builder already accepts any column name:

```go
table.Timestamp("soft_deleted_at").Nullable()
```

## Usage Examples

### Default `deleted_at` (unchanged)

```go
type User struct {
    neat.SoftDeletes
    ID   uint
    Name string
}
```

### Custom column via method override

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

Note: with this approach the struct field is still named `DeletedAt` and tagged `db:"deleted_at"`. The `db` tag on the struct field must also be updated, or a custom scan mapping applied, to correctly read the column back from the database. Using `SoftDeletesAlt` avoids this inconsistency.

### Using the built-in `SoftDeletesAlt`

```go
type User struct {
    neat.SoftDeletesAlt  // field: SoftDeletedAt, column: soft_deleted_at
    ID   uint
    Name string
}
```

### Querying

```go
// Excludes soft-deleted records (default)
db.Query().Model(&User{}).Get(&users)
// WHERE soft_deleted_at IS NULL

// Include all records
db.Query().Model(&User{}).WithTrashed().Get(&users)

// Only deleted records
db.Query().Model(&User{}).OnlyTrashed().Get(&users)
// WHERE soft_deleted_at IS NOT NULL
```

### Soft deleting, restoring, checking status

```go
db.Query().Model(&User{}).Where("id", 1).Delete()  // Sets soft_deleted_at to now

if user.IsDeleted() {
    fmt.Println("User is soft-deleted")
}

user.Restore()  // Sets soft_deleted_at to nil; then save the model
```

## Benefits

1. **Semantic clarity**: `soft_deleted_at` explicitly indicates soft deletion
2. **Correctness**: Fixes the gap between existing docs and actual implementation
3. **Interface-driven**: Clean Go idiom; no fragile reflection on field names
4. **Backward compatible**: Existing `deleted_at` usage continues to work unchanged
5. **Extensible**: Any column name works, not just two hardcoded options

## Drawbacks

1. **Two structs to maintain**: `SoftDeletes` and `SoftDeletesAlt` diverge in field name and type, requiring parallel maintenance
2. **Method-override inconsistency**: If a user overrides `DeletedAtColumn()` on a model embedding `SoftDeletes`, the struct field tag (`db:"deleted_at"`) no longer matches the column, which will break scan/hydration unless the field tag is also updated
3. **API surface grows**: Adding `SoftDeleteColumnNamer` to the contracts package is a minor but permanent API addition

## Alternatives Considered

### Alternative 1: Rename the default column globally

Change the default from `deleted_at` to `soft_deleted_at` for all users.

**Pros**: Consistent and self-documenting across all projects
**Cons**: Breaking change; all existing users must migrate schema and data

**Decision**: Rejected to maintain backward compatibility

### Alternative 2: Struct tag configuration

```go
type SoftDeletes struct {
    DeletedAt *time.Time `db:"soft_deleted_at"`
}
```

**Pros**: Standard Go convention; column name is co-located with the field
**Cons**: The `db` tag already controls scan mapping; no additional ORM-level hook is needed — but it does not signal to `hasSoftDeleteCapability` or the query builder what column to filter on

**Decision**: Workable for scan/hydration but insufficient on its own; the query builder still needs to know the column name at query-build time. Can be combined with `DeletedAtColumn()` for a complete solution.

### Alternative 3: Struct tag for the query builder column name

```go
type SoftDeletes struct {
    DeletedAt *time.Time `db:"deleted_at" neat_soft_delete:"deleted_at"`
}
```

**Pros**: Declarative; no interface required
**Cons**: Requires reflection at query-build time; less idiomatic in Go; hard to override in embedding models

**Decision**: Rejected in favor of the interface-based method override pattern

### Alternative 4: No changes

Keep the hardcoded `deleted_at` column name.

**Pros**: Zero maintenance cost
**Cons**: Contradicts the existing documentation; blocks legitimate use cases

**Decision**: Rejected

## Implementation Plan

1. Resolve the `*time.Time` vs `sql.NullTime` discrepancy between `database/soft_delete` and `database/orm` packages — decide the canonical type
2. Define the `SoftDeleteColumnNamer` interface in `contracts/database/orm`
3. Add `DeletedAtColumn() string` to both `SoftDeletes` implementations
4. Add `SoftDeletesAlt` struct in `database/soft_delete`
5. Add `getSoftDeleteColumn` helper in `database/query`
6. Refactor `hasSoftDeleteCapability` to use interface assertion
7. Update `buildWheresWithSoftDelete` and `buildWheresWithSoftDeleteIndex` in `builder_where.go`
8. Update `Delete()` in `query_delete.go`
9. Audit `builder_update.go` for any additional hardcoded `deleted_at` references
10. Update tests in `query_soft_delete_test.go` and `builder_where_test.go` to cover both column names
11. Update `docs/soft-deletes.md` to remove the placeholder notice and document the full behavior

## Open Questions

1. **Canonical soft delete type**: Should `DeletedAt` be `*time.Time` (nullable pointer, `database/soft_delete`) or `sql.NullTime` (`database/orm`)? These currently diverge and both must be handled by `hasSoftDeleteCapability`.
2. **Scan/hydration consistency**: When a user overrides `DeletedAtColumn()` on a model that embeds `SoftDeletes`, the `db:"deleted_at"` tag on the embedded field still points to the old column. Should the ORM automatically remap the scan target, or is it the user's responsibility to also update the tag (or use `SoftDeletesAlt`)?
3. **More alternative structs**: Should we provide additional pre-built structs for other conventions (e.g., `archived_at`, `removed_at`)? Or is the method-override pattern sufficient?
4. **Global default override**: Should the default column name be configurable at the ORM/database level, not just per model?
5. **Migration tooling**: Should we provide a helper to generate the SQL to rename an existing `deleted_at` column to `soft_deleted_at`?

## References

- Current Soft Deletes Documentation: `docs/soft-deletes.md`
- Existing `SoftDeletes` implementation: `database/soft_delete/soft_delete.go`
- ORM model structs: `database/orm/model.go`
- Query builder soft delete logic: `database/query/builder_where.go`, `database/query/query_delete.go`
- Laravel Eloquent Soft Deletes: https://laravel.com/docs/eloquent#soft-deleting
