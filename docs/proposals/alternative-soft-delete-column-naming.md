# Proposal: Soft Delete Column Naming & Semantic API

**Status**: Implemented  
**Date**: June 1, 2026  
**Implemented**: June 14–15, 2026  
**Author**: Neat ORM Team

---

## Summary

Redesign the soft delete surface of Neat ORM for semantic clarity:

1. Change the **default soft delete column** from `deleted_at` to `soft_deleted_at`
2. Provide three **built-in embeds** covering all common use-cases
3. Rename all **methods and query builder verbs** to make the "soft" nature explicit
4. Keep **deprecated aliases** for every renamed identifier so nothing breaks silently

---

## Motivation

The original implementation had three problems:

1. **Semantic ambiguity** — `deleted_at`, `IsDeleted()`, `WithTrashed()` etc. did not indicate
   *soft* deletion. A developer reading code could not tell whether a record was permanently
   or soft-deleted.

2. **Laravel coupling** — `WithTrashed()` / `OnlyTrashed()` / `WithoutTrashed()` are
   Laravel Eloquent terms with no meaning outside that ecosystem.

3. **Hardcoded column** — `builder_where.go` hardcoded `"deleted_at"` regardless of any
   override, making the documented customization mechanism non-functional.

---

## What Was Changed

### 1. Default column: `deleted_at` → `soft_deleted_at`

The fallback in `getSoftDeleteColumn()` (`database/query/query_delete.go`) now returns
`"soft_deleted_at"`. The `SoftDeletes` embed also now uses `soft_deleted_at`.

This is a **breaking change** for users migrating from a schema with `deleted_at`.
See the [Migration Guide](#migration-guide) below.

### 2. Three built-in embed structs

| Struct | Column | Use case |
|---|---|---|
| `soft_delete.SoftDeletes` | `soft_deleted_at` | Default — new projects |
| `soft_delete.SoftDeletedAt` | `soft_deleted_at` | Explicit naming (mirrors `CreatedAt`/`UpdatedAt`) |
| `soft_delete.DeletedAt` | `deleted_at` | Laravel-compatible / existing `deleted_at` schemas |

The same three structs exist in `database/orm` for the `sql.NullTime` variant.

### 3. `SoftDeleteColumnNamer` interface — `contracts/database/orm/soft_delete.go`

Any model can implement this interface to use any column name:

```go
type SoftDeleteColumnNamer interface {
    SoftDeletedAtColumn() string
}
```

### 4. Renamed methods (deprecated aliases kept)

**Struct methods** (on all three embeds):

| New (primary) | Old (deprecated alias) |
|---|---|
| `SoftDelete()` | `Delete()` |
| `IsSoftDeleted() bool` | `IsDeleted() bool` |
| `GetSoftDeletedAt() *time.Time` | `GetDeletedAt() *time.Time` |
| `SoftDeletedAtColumn() string` | `DeletedAtColumn() string` |
| `RestoreSoftDeleted()` | `Restore()` |

**Query builder methods**:

| New (primary) | Old (deprecated alias) |
|---|---|
| `WithSoftDeleted()` | `WithTrashed()` |
| `OnlySoftDeleted()` | `OnlyTrashed()` |
| `WithoutSoftDeleted()` | `WithoutTrashed()` |
| `RestoreSoftDeleted(...)` | `Restore(...)` |

**Event constants** (`contracts/database/orm/events.go`):

| New (primary) | Old (deprecated alias) |
|---|---|
| `EventSoftDeleteRestoring` | `EventRestoring` |
| `EventSoftDeleteRestored` | `EventRestored` |

**Package-level constants** (`database/soft_delete`):

| New (primary) | Old (deprecated alias) |
|---|---|
| `SoftDeleteAtColumn = "soft_deleted_at"` | `DeletedAtColumn` (points to `SoftDeleteAtColumn`) |
| `DeletedAtColumnName = "deleted_at"` | _(new, no alias needed)_ |

### 5. Internal query struct field renames

Private fields in the query struct — no API impact, purely internal clarity:

| New | Old |
|---|---|
| `includeSoftDeleted` | `withTrashed` |
| `onlySoftDeleted` | `onlyTrashed` |
| `excludeSoftDeleted` | `withoutTrashed` |

---

## Usage

### Default — new projects (`soft_deleted_at`)

```go
type User struct {
    soft_delete.SoftDeletes  // uses "soft_deleted_at"
    ID   uint
    Name string
}
```

Schema:
```sql
ALTER TABLE users ADD COLUMN soft_deleted_at DATETIME NULL;
```

### Explicit naming (identical behavior)

```go
type User struct {
    soft_delete.SoftDeletedAt  // also uses "soft_deleted_at"
    ID   uint
    Name string
}
```

### Laravel-compatible (`deleted_at`)

```go
type User struct {
    soft_delete.DeletedAt  // uses "deleted_at"
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

func (o *Order) SoftDeletedAtColumn() string { return "removed_at" }
```

Schema:
```sql
ALTER TABLE orders ADD COLUMN removed_at DATETIME NULL;
```

### Querying

```go
// Excludes soft-deleted records (default)
db.Query().Model(&User{}).Get(&users)
// WHERE soft_deleted_at IS NULL

// Include all records
db.Query().Model(&User{}).WithSoftDeleted().Get(&users)

// Only soft-deleted records
db.Query().Model(&User{}).OnlySoftDeleted().Get(&users)
// WHERE soft_deleted_at IS NOT NULL

// Soft delete — sets soft_deleted_at to now
db.Query().Model(&User{}).Where("id = ?", 1).SoftDelete()

// Restore — sets soft_deleted_at to nil
db.Query().Model(&User{}).WithSoftDeleted().Where("id = ?", 1).RestoreSoftDeleted()

// Check status on struct
if user.IsSoftDeleted() {
    fmt.Println("User is soft-deleted")
}
```

---

## Migration Guide

### From `deleted_at` (previous default) to `soft_deleted_at` (new default)

**Option A — rename the column (recommended for new schemas):**

```sql
-- PostgreSQL / MySQL
ALTER TABLE users RENAME COLUMN deleted_at TO soft_deleted_at;

-- SQLite (no RENAME COLUMN before 3.25.0)
ALTER TABLE users ADD COLUMN soft_deleted_at DATETIME;
UPDATE users SET soft_deleted_at = deleted_at;
ALTER TABLE users DROP COLUMN deleted_at;
```

**Option B — use the `DeletedAt` embed (zero code change, keep schema as-is):**

```go
// Before
type User struct {
    soft_delete.SoftDeletes  // was deleted_at, now soft_deleted_at — BREAKS
    ID uint
}

// After (Option B)
type User struct {
    soft_delete.DeletedAt  // deleted_at, unchanged
    ID uint
}
```

**Option C — override `SoftDeletedAtColumn()` on the model:**

```go
type User struct {
    soft_delete.SoftDeletes
    ID uint
}

func (u *User) SoftDeletedAtColumn() string { return "deleted_at" }
```

---

## File Reference

| Concern | File |
|---|---|
| Embed structs | `database/soft_delete/soft_delete.go` |
| `sql.NullTime` variants | `database/orm/model.go` |
| `SoftDeleteColumnNamer` interface | `contracts/database/orm/soft_delete.go` |
| Query builder soft delete | `database/query/builder_where.go`, `query_delete.go`, `query_advanced.go`, `builder_update.go` |
| Soft delete state methods | `database/query/query_soft_delete.go` |
| Event constants | `contracts/database/orm/events.go` |
| Query interface | `contracts/database/orm/orm.go` |
