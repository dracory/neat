# Proposal: Alternative Soft Delete Column Naming

**Status**: Implemented
**Date**: June 1, 2026
**Implemented**: June 14, 2026
**Author**: Neat ORM Team

## Summary

Add support for customizing the soft delete column name, allowing users to use
`soft_deleted_at` (or any other name) instead of the hardcoded `deleted_at`.
This provides better semantic clarity for teams that prefer explicit naming
conventions, and fixes the gap between the existing documentation and the
actual implementation.

## Motivation

The soft delete implementation hardcoded `deleted_at` throughout the query
builder (`builder_where.go`). While this is a common convention (Laravel
Eloquent, etc.), it created friction in three scenarios:

1. **Semantic clarity**: `deleted_at` doesn't explicitly indicate soft deletion
   vs. hard deletion
2. **Legacy schemas**: Existing databases may already use `soft_deleted_at` or
   another column name
3. **Team conventions**: Some teams enforce explicit naming for all "soft"
   operations

The existing `docs/soft-deletes.md` documented a `DeletedAtColumn() string`
method override as the customization mechanism, but it was not implemented —
`builder_where.go` hardcoded `"deleted_at"` regardless of any such method on
the model.

## What Was Changed

### 1. `SoftDeleteColumnNamer` interface — `contracts/database/orm/soft_delete.go`

A single interface that any model can implement to override the column name:

```go
// SoftDeleteColumnNamer is implemented by models that customize the soft delete column name.
type SoftDeleteColumnNamer interface {
    DeletedAtColumn() string
}
```

### 2. `DeletedAtColumn()` added to `SoftDeletes`

Both `SoftDeletes` implementations now implement the interface:

`database/soft_delete/soft_delete.go` (`*time.Time` variant — canonical):
```go
func (sd *SoftDeletes) DeletedAtColumn() string { return "deleted_at" }
```

`database/orm/model.go` (`sql.NullTime` variant):
```go
func (sd *SoftDeletes) DeletedAtColumn() string { return "deleted_at" }
```

### 3. New `SoftDeletedAt` struct — `database/soft_delete/soft_delete.go`

A built-in alternative embed using `soft_deleted_at` as the column name.
Named following the same convention as `CreatedAt` and `UpdatedAt` — the
struct name matches the field it provides:

```go
// SoftDeletedAt provides soft delete functionality using the "soft_deleted_at" column name.
// Follows the same naming convention as CreatedAt and UpdatedAt.
type SoftDeletedAt struct {
    SoftDeletedAt *time.Time `json:"soft_deleted_at,omitempty" db:"soft_deleted_at"`
}

func (sd *SoftDeletedAt) DeletedAtColumn() string { return "soft_deleted_at" }
func (sd *SoftDeletedAt) IsDeleted() bool         { return sd.SoftDeletedAt != nil }
func (sd *SoftDeletedAt) Delete()                 { now := time.Now(); sd.SoftDeletedAt = &now }
func (sd *SoftDeletedAt) Restore()                { sd.SoftDeletedAt = nil }
func (sd *SoftDeletedAt) GetDeletedAt() *time.Time { return sd.SoftDeletedAt }
```

### 4. `hasSoftDeleteCapability` — `database/query/query_delete.go`

Replaced the fragile reflection-based field-name check with a clean interface
assertion:

```go
// Before: reflected on field named "DeletedAt" of type *time.Time
// After:
func hasSoftDeleteCapability(model any) bool {
    if model == nil {
        return false
    }
    _, ok := model.(contractsorm.SoftDeleteColumnNamer)
    return ok
}
```

### 5. `getSoftDeleteColumn` helper — `database/query/query_delete.go`

```go
// getSoftDeleteColumn returns the soft delete column name for the given model,
// falling back to "deleted_at" if the model does not implement SoftDeleteColumnNamer.
func getSoftDeleteColumn(model any) string {
    if namer, ok := model.(contractsorm.SoftDeleteColumnNamer); ok {
        return namer.DeletedAtColumn()
    }
    return "deleted_at"
}
```

### 6. Query builder updated to use `getSoftDeleteColumn`

All hardcoded `"deleted_at"` references replaced in:

- `builder_where.go` — `buildWheresWithSoftDelete()` and `buildWheresWithSoftDeleteIndex()`
- `query_delete.go` — `Delete()`
- `query_advanced.go` — `Restore()`
- `builder_update.go` — `isSoftDeleteOperation` detection

## Usage

### Default `deleted_at` (unchanged)

```go
type User struct {
    soft_delete.SoftDeletes  // uses "deleted_at"
    ID   uint
    Name string
}
```

### Built-in `SoftDeletedAt` embed

```go
type User struct {
    soft_delete.SoftDeletedAt  // uses "soft_deleted_at"
    ID   uint
    Name string
}
```

### Fully custom column name

```go
type Order struct {
    soft_delete.SoftDeletes
    ID uint
}

func (o *Order) DeletedAtColumn() string { return "removed_at" }
```

> **Note**: When overriding `DeletedAtColumn()` on a model that embeds
> `SoftDeletes`, the struct field tag (`db:"deleted_at"`) no longer matches
> the column. The scan/hydration will break unless the `db` tag is also
> updated. Using `SoftDeletedAt` avoids this inconsistency entirely.

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
// Soft delete — sets soft_deleted_at to now
db.Query().Model(&User{}).Where("id = ?", 1).Delete()

if user.IsDeleted() {
    fmt.Println("User is soft-deleted")
}

// Restore — sets soft_deleted_at to nil
db.Query().Model(&User{}).WithTrashed().Where("id = ?", 1).Restore()
```

### Schema

No changes required to the schema builder — it already accepts any column name:

```go
table.Timestamp("soft_deleted_at").Nullable()
```

## Benefits

1. **Naming consistency**: `SoftDeletedAt` follows the same convention as
   `CreatedAt` and `UpdatedAt` — struct name matches the field it provides
2. **Correctness**: Fixes the gap between existing docs and actual implementation
3. **Interface-driven**: Clean Go idiom; no fragile reflection on field names
4. **Backward compatible**: Existing `deleted_at` usage works unchanged;
   `hasSoftDeleteCapability` falls back gracefully for models that don't
   implement the interface
5. **Extensible**: Any column name works via `DeletedAtColumn()` override

## Drawbacks

1. **Two structs to maintain**: `SoftDeletes` and `SoftDeletedAt` diverge in
   field name, requiring parallel maintenance
2. **Method-override inconsistency**: Overriding `DeletedAtColumn()` on a model
   embedding `SoftDeletes` mismatches the `db:"deleted_at"` field tag —
   developers must use `SoftDeletedAt` or update the tag manually
3. **API surface grows**: `SoftDeleteColumnNamer` is a permanent addition to the
   contracts package

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
**Cons**: Does not signal to `hasSoftDeleteCapability` or the query builder
what column to filter on at query-build time  
**Decision**: Workable for scan/hydration but insufficient alone; can be
combined with `DeletedAtColumn()` for a complete solution

### Alternative 3: Struct tag for the query builder column name

```go
type SoftDeletes struct {
    DeletedAt *time.Time `db:"deleted_at" neat_soft_delete:"deleted_at"`
}
```

**Pros**: Declarative; no interface required  
**Cons**: Requires reflection at query-build time; less idiomatic in Go;
hard to override in embedding models  
**Decision**: Rejected in favor of the interface-based method override pattern

### Alternative 4: No changes

Keep the hardcoded `deleted_at` column name.

**Pros**: Zero maintenance cost  
**Cons**: Contradicts the existing documentation; blocks legitimate use cases  
**Decision**: Rejected

## Open Questions (resolved)

1. **Canonical soft delete type**: `*time.Time` (`database/soft_delete`) is the
   canonical type. `sql.NullTime` (`database/orm`) is a secondary variant that
   also received `DeletedAtColumn()` for consistency. Both implement
   `SoftDeleteColumnNamer`.

2. **Scan/hydration consistency**: When a user overrides `DeletedAtColumn()` on
   a model embedding `SoftDeletes`, the `db:"deleted_at"` tag mismatch is the
   user's responsibility. The recommended solution is to use `SoftDeletedAt`
   instead, which has a matching field tag from the start.

3. **More alternative structs**: The method-override pattern (`DeletedAtColumn()
   string`) is sufficient for any arbitrary column name. No additional pre-built
   structs are planned beyond `SoftDeletes` and `SoftDeletedAt`.

4. **Global default override**: Not implemented. Per-model override via
   `DeletedAtColumn()` is the supported mechanism.

5. **Migration tooling**: Out of scope for this proposal.

## References

- Soft Deletes Documentation: `docs/soft-deletes.md`
- `SoftDeletes` and `SoftDeletedAt` implementation: `database/soft_delete/soft_delete.go`
- `SoftDeleteColumnNamer` interface: `contracts/database/orm/soft_delete.go`
- ORM model structs: `database/orm/model.go`
- Query builder soft delete logic: `database/query/builder_where.go`,
  `database/query/query_delete.go`, `database/query/query_advanced.go`,
  `database/query/builder_update.go`
- Laravel Eloquent Soft Deletes: https://laravel.com/docs/eloquent#soft-deleting
