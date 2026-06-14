# Soft Deletes

Soft deletes mark records as deleted without removing them from the database.
The row remains in the table and can be restored. Normal queries automatically
exclude soft-deleted records.

## What are Soft Deletes?

Soft deletes are useful for:
- Data recovery and audit trails
- Compliance requirements (data retention)
- Undo / restore workflows

---

## Choosing an Embed

Neat provides three built-in soft delete embeds. Pick the one that matches your schema:

| Embed | Column | When to use |
|---|---|---|
| `soft_delete.SoftDeletes` | `soft_deleted_at` | Default — new projects |
| `soft_delete.SoftDeletedAt` | `soft_deleted_at` | Explicit naming (mirrors `CreatedAt`/`UpdatedAt`) |
| `soft_delete.DeletedAt` | `deleted_at` | Laravel-compatible or existing `deleted_at` schemas |

All three embeds are in the `github.com/dracory/neat/database/soft_delete` package.
Equivalent `sql.NullTime` variants exist in `github.com/dracory/neat/database/orm`.

---

## Setting Up

### 1. Add the embed to your model

```go
import "github.com/dracory/neat/database/soft_delete"

// New project — default column soft_deleted_at
type User struct {
    soft_delete.SoftDeletes
    ID   uint
    Name string
}

// Laravel-compatible — column deleted_at
type Post struct {
    soft_delete.DeletedAt
    ID    uint
    Title string
}
```

### 2. Add the column to your schema

```sql
-- SoftDeletes / SoftDeletedAt
ALTER TABLE users ADD COLUMN soft_deleted_at DATETIME NULL;

-- DeletedAt
ALTER TABLE posts ADD COLUMN deleted_at DATETIME NULL;
```

Or using the schema builder:

```go
table.Timestamp("soft_deleted_at").Nullable()
table.Timestamp("deleted_at").Nullable()
```

---

## Soft Deleting Records

```go
// Via query builder — sets soft_deleted_at = NOW() in the database
db.Query().Model(&User{}).Where("id = ?", 1).SoftDelete()

// Via struct method — sets the field in memory (does not hit the database)
user.SoftDelete()
```

---

## Querying

```go
var users []User

// Default — excludes soft-deleted records (soft_deleted_at IS NULL)
db.Query().Model(&User{}).Get(&users)

// Include all records (including soft-deleted)
db.Query().Model(&User{}).WithSoftDeleted().Get(&users)

// Only soft-deleted records
db.Query().Model(&User{}).OnlySoftDeleted().Get(&users)

// Explicitly exclude (reset a previous WithSoftDeleted)
db.Query().Model(&User{}).WithoutSoftDeleted().Get(&users)
```

---

## Restoring Soft-Deleted Records

```go
// Restore by condition — sets soft_deleted_at = NULL
res, err := db.Query().Model(&User{}).
    WithSoftDeleted().
    Where("id = ?", 1).
    RestoreSoftDeleted()

// Restore by model instance
res, err := db.Query().Model(&User{}).
    WithSoftDeleted().
    RestoreSoftDeleted(&user)
```

---

## Force Deleting (Permanent)

```go
// Permanently removes the row — bypasses soft delete
res, err := db.Query().Model(&User{}).Where("id = ?", 1).ForceDelete()
```

---

## Checking Status on a Struct

```go
if user.IsSoftDeleted() {
    fmt.Println("User is soft-deleted")
}

ts := user.GetSoftDeletedAt()  // returns *time.Time (nil if not deleted)
```

---

## Custom Column Name

If none of the three built-in embeds match your column name, override
`SoftDeletedAtColumn()` on the model:

```go
type Order struct {
    soft_delete.SoftDeletes
    ID uint
}

// Override — query builder will use "removed_at" instead of "soft_deleted_at"
func (o *Order) SoftDeletedAtColumn() string { return "removed_at" }
```

> **Note:** When overriding the column name, ensure the `db` struct tag on the
> embedded field also matches, or scan/hydration will break. Using `SoftDeletedAt`
> (which has `db:"soft_deleted_at"`) or `DeletedAt` (which has `db:"deleted_at"`)
> avoids this mismatch entirely.

---

## Deprecated API

The following identifiers still work but produce deprecation warnings from
`golangci-lint` (SA1019). Migrate to the new names at your convenience.

| Deprecated | Use instead |
|---|---|
| `user.Delete()` | `user.SoftDelete()` |
| `user.IsDeleted()` | `user.IsSoftDeleted()` |
| `user.GetDeletedAt()` | `user.GetSoftDeletedAt()` |
| `user.DeletedAtColumn()` | `user.SoftDeletedAtColumn()` |
| `user.Restore()` | `user.RestoreSoftDeleted()` |
| `query.WithTrashed()` | `query.WithSoftDeleted()` |
| `query.OnlyTrashed()` | `query.OnlySoftDeleted()` |
| `query.WithoutTrashed()` | `query.WithoutSoftDeleted()` |
| `query.Restore(...)` | `query.RestoreSoftDeleted(...)` |
| `EventRestoring` | `EventSoftDeleteRestoring` |
| `EventRestored` | `EventSoftDeleteRestored` |

---

## Migrating from `deleted_at` to `soft_deleted_at`

If you were using `soft_delete.SoftDeletes` with a `deleted_at` column
(the previous default), you have three options:

**Option A — rename the column (recommended for new schemas):**

```sql
-- PostgreSQL / MySQL 8+
ALTER TABLE users RENAME COLUMN deleted_at TO soft_deleted_at;
```

**Option B — swap the embed (zero schema change):**

```go
// Before
type User struct { soft_delete.SoftDeletes ... }

// After
type User struct { soft_delete.DeletedAt ... }  // keeps deleted_at column
```

**Option C — override the column name:**

```go
func (u *User) SoftDeletedAtColumn() string { return "deleted_at" }
```
