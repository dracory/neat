# Proposal: Alternative Soft Delete Support (Max-Date Sentinel)

**Status**: Proposed  
**Date**: June 1, 2026  
**Author**: Neat ORM Team

---

## Summary

Add a `SoftDeletesMaxDate` embed that uses a maximum datetime sentinel
(`9999-12-31 23:59:59 UTC`) as the "not deleted" value instead of NULL.
Records are soft-deleted when their `soft_deleted_at` timestamp is in the
past (≤ current time), and active when it is in the future.

---

## Motivation

The NULL-based soft delete strategy (current default) has real limitations
in certain schemas:

1. **NULL handling complexity** — queries must use `IS NULL` / `IS NOT NULL`
2. **Index inefficiency** — many databases store NULLs outside B-tree indexes
   or require special `IS NULL`-aware indexes
3. **NOT NULL constraint conflicts** — some schemas enforce `NOT NULL` on all
   timestamp columns
4. **Legacy compatibility** — some existing systems use a sentinel value to
   avoid NULLs entirely

The max-date approach offers:
- Simpler range queries (`soft_deleted_at > NOW()` instead of `IS NULL`)
- Better partial index support (`WHERE soft_deleted_at > NOW()` is a range scan)
- Compatible with `NOT NULL` column constraints
- Clear semantic: "deleted until the end of time means not deleted"

---

## Current State (as implemented)

### How soft delete detection works today

`hasSoftDeleteCapability` in `database/query/query_delete.go` uses a clean
**interface assertion** — no reflection:

```go
func hasSoftDeleteCapability(model any) bool {
    if model == nil {
        return false
    }
    _, ok := model.(contractsorm.SoftDeleteColumnNamer)
    return ok
}
```

Any model that implements `SoftDeleteColumnNamer` is considered soft-deletable.
The three built-in embeds (`SoftDeletes`, `SoftDeletedAt`, `DeletedAt`) all
satisfy this interface via their promoted `SoftDeletedAtColumn() string` method.

### Column name resolution

`getSoftDeleteColumn` in `database/query/query_delete.go`:

```go
func getSoftDeleteColumn(model any) string {
    if namer, ok := model.(contractsorm.SoftDeleteColumnNamer); ok {
        return namer.SoftDeletedAtColumn()
    }
    return "soft_deleted_at"  // default fallback
}
```

### `SoftDeleteColumnNamer` interface — `contracts/database/orm/soft_delete.go`

```go
type SoftDeleteColumnNamer interface {
    SoftDeletedAtColumn() string
}
```

### WHERE clause construction — `database/query/builder_where.go`

Currently NULL-only — uses `IS NULL` / `IS NOT NULL`:

```go
col := b.quoteIdentifier(getSoftDeleteColumn(b.query.model))
switch {
case b.query.onlySoftDeleted:
    prefix = col + " IS NOT NULL"
case b.query.includeSoftDeleted:
    // no filter
default:
    prefix = col + " IS NULL"
}
```

### Query builder state fields (private)

| Field | Meaning |
|---|---|
| `includeSoftDeleted` | include soft-deleted rows (like `WITH TRASHED`) |
| `onlySoftDeleted` | only soft-deleted rows |
| `excludeSoftDeleted` | force-exclude soft-deleted rows |

### Delete — `database/query/query_delete.go`

```go
if useSoftDelete && !q.includeSoftDeleted && !q.onlySoftDeleted {
    clone := q.Clone().(*Query)
    clone.includeSoftDeleted = true
    builder := NewBuilder(clone)
    col := getSoftDeleteColumn(q.model)
    deleteSQL, args = builder.BuildUpdate(map[string]any{col: time.Now()})
}
```

### Restore — `database/query/query_advanced.go`

Sets the column to `nil` (NULL):

```go
sql, args := builder.BuildUpdate(map[string]any{col: nil})
```

For max-date strategy, restore must set the column back to `MaxSoftDeletedAt`,
not `nil`. **This is the primary gap to close.**

### Five embeds (three existing, two proposed)

| Embed | Column | Strategy |
|---|---|---|
| `soft_delete.SoftDeletes` | `soft_deleted_at` | NULL-based (default) |
| `soft_delete.SoftDeletedAt` | `soft_deleted_at` | NULL-based (explicit naming) |
| `soft_delete.DeletedAt` | `deleted_at` | NULL-based (Laravel compat) |
| `soft_delete.SoftDeletesMaxDate` | `soft_deleted_at` | **Max-date (proposed)** |
| `soft_delete.DeletedAtMaxDate` | `deleted_at` | **Max-date, Laravel compat (proposed)** |

All live in `database/soft_delete/soft_delete.go`.

---

## Proposed Changes

### 1. Add two new max-date embed structs

Add to `database/soft_delete/soft_delete.go` alongside the existing embeds.
Following the same pattern as the NULL-based embeds, **two** max-date structs
are provided — one per supported column name:

| Struct | Column | Use case |
|---|---|---|
| `SoftDeletesMaxDate` | `soft_deleted_at` | Default max-date — new projects |
| `DeletedAtMaxDate` | `deleted_at` | Laravel-compatible max-date |

```go
// MaxSoftDeletedAt is the sentinel "not deleted" value for the max-date strategy.
var MaxSoftDeletedAt = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)

// SoftDeletesMaxDate provides soft delete functionality using a max-date sentinel
// instead of NULL. The column used is "soft_deleted_at".
// Embed this in models where the schema enforces NOT NULL on timestamp columns.
//
// Example:
//
//  type User struct {
//      soft_delete.SoftDeletesMaxDate  // uses "soft_deleted_at" with sentinel
//      ID   uint
//      Name string
//  }
type SoftDeletesMaxDate struct {
    SoftDeletedAt time.Time `json:"soft_deleted_at,omitempty" db:"soft_deleted_at"`
}

func (s *SoftDeletesMaxDate) SoftDeletedAtColumn() string { return "soft_deleted_at" }
func (s *SoftDeletesMaxDate) IsSoftDeleted() bool         { return !s.SoftDeletedAt.After(time.Now()) }
func (s *SoftDeletesMaxDate) SoftDelete()                 { s.SoftDeletedAt = time.Now() }
func (s *SoftDeletesMaxDate) RestoreSoftDeleted()         { s.SoftDeletedAt = MaxSoftDeletedAt }
func (s *SoftDeletesMaxDate) GetSoftDeletedAt() time.Time { return s.SoftDeletedAt }

// Deprecated aliases
func (s *SoftDeletesMaxDate) IsDeleted() bool         { return s.IsSoftDeleted() }
func (s *SoftDeletesMaxDate) Delete()                 { s.SoftDelete() }
func (s *SoftDeletesMaxDate) Restore()                { s.RestoreSoftDeleted() }
func (s *SoftDeletesMaxDate) DeletedAtColumn() string { return s.SoftDeletedAtColumn() }

// DeletedAtMaxDate provides soft delete functionality using a max-date sentinel
// with the "deleted_at" column name (Laravel-compatible).
// Use this when your schema uses "deleted_at" and enforces NOT NULL.
//
// Example:
//
//  type Post struct {
//      soft_delete.DeletedAtMaxDate  // uses "deleted_at" with sentinel
//      ID    uint
//      Title string
//  }
type DeletedAtMaxDate struct {
    DeletedAt time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

func (s *DeletedAtMaxDate) SoftDeletedAtColumn() string { return "deleted_at" }
func (s *DeletedAtMaxDate) IsSoftDeleted() bool         { return !s.DeletedAt.After(time.Now()) }
func (s *DeletedAtMaxDate) SoftDelete()                 { s.DeletedAt = time.Now() }
func (s *DeletedAtMaxDate) RestoreSoftDeleted()         { s.DeletedAt = MaxSoftDeletedAt }
func (s *DeletedAtMaxDate) GetSoftDeletedAt() time.Time { return s.DeletedAt }

// Deprecated aliases
func (s *DeletedAtMaxDate) IsDeleted() bool         { return s.IsSoftDeleted() }
func (s *DeletedAtMaxDate) Delete()                 { s.SoftDelete() }
func (s *DeletedAtMaxDate) Restore()                { s.RestoreSoftDeleted() }
func (s *DeletedAtMaxDate) DeletedAtColumn() string { return s.SoftDeletedAtColumn() }
```

> **Note**: `GetSoftDeletedAt()` returns `time.Time` (not `*time.Time`) because
> the field is always set. The zero value `0001-01-01` would incorrectly report
> `IsSoftDeleted() == true`, so the ORM automatically sets the field to
> `MaxSoftDeletedAt` before executing INSERT on any model that implements
> `SoftDeleteStrategy`. A database-level `DEFAULT` is still recommended as a
> safety net for rows inserted outside the ORM.

### 2. Extend `SoftDeleteColumnNamer` with a restore-value method — OR — add `SoftDeleteStrategy`

`SoftDeleteColumnNamer` currently only provides the column name. The query builder
also needs to know **what value to set on restore** — `nil` for NULL-based, or
`MaxSoftDeletedAt` for max-date.

**Option A — extend `SoftDeleteColumnNamer`** (simpler, one interface):

```go
type SoftDeleteColumnNamer interface {
    SoftDeletedAtColumn() string
    // SoftDeleteRestoreValue returns the value to set when restoring a soft-deleted record.
    // NULL-based embeds return nil. Max-date embed returns MaxSoftDeletedAt.
    SoftDeleteRestoreValue() any
}
```

All three existing embeds would return `nil`. `SoftDeletesMaxDate` returns `MaxSoftDeletedAt`.

**Option B — separate `SoftDeleteStrategy` interface** (more flexible, handles
custom WHERE conditions for the max-date query filter):

```go
// SoftDeleteStrategy is implemented by models that use a non-NULL soft delete strategy.
// Models that do NOT implement this interface are assumed to use the NULL strategy.
type SoftDeleteStrategy interface {
    SoftDeleteColumnNamer
    // SoftDeleteValue returns the value to write on soft delete (e.g. time.Now()).
    SoftDeleteValue() any
    // SoftDeleteRestoreValue returns the value to write on restore.
    RestoreValue() any
    // IsDeletedCondition returns the SQL fragment + args for the "only deleted" filter.
    IsDeletedCondition(quoteIdentifier func(string) string) (string, []any)
    // IsActiveCondition returns the SQL fragment + args for the "not deleted" filter.
    IsActiveCondition(quoteIdentifier func(string) string) (string, []any)
}
```

`SoftDeletesMaxDate` implements it:

```go
func (s *SoftDeletesMaxDate) SoftDeleteValue() any { return time.Now() }
func (s *SoftDeletesMaxDate) RestoreValue() any     { return MaxSoftDeletedAt }

func (s *SoftDeletesMaxDate) IsDeletedCondition(q func(string) string) (string, []any) {
    return q("soft_deleted_at") + " <= ?", []any{time.Now()}
}
func (s *SoftDeletesMaxDate) IsActiveCondition(q func(string) string) (string, []any) {
    return q("soft_deleted_at") + " > ?", []any{time.Now()}
}
```

**Recommendation**: Option B — it is the more complete design. The NULL-based
embeds do not implement `SoftDeleteStrategy`, so the query builder falls back
to `IS NULL` / `IS NOT NULL` for backward compatibility.

### 3. `hasSoftDeleteCapability` — no change needed

The current implementation already works correctly via `SoftDeleteColumnNamer`.
`SoftDeletesMaxDate` implements `SoftDeletedAtColumn()`, so it is automatically
detected. **No changes required to `hasSoftDeleteCapability`.**

### 4. Update `Delete()` in `query_delete.go`

Add a strategy check alongside the existing `getSoftDeleteColumn` call:

```go
col := getSoftDeleteColumn(q.model)
var deleteValue any = time.Now()
if strat, ok := q.model.(contractsorm.SoftDeleteStrategy); ok {
    deleteValue = strat.SoftDeleteValue()
}
deleteSQL, args = builder.BuildUpdate(map[string]any{col: deleteValue})
```

For NULL-based models (no `SoftDeleteStrategy`), this is identical to today.

### 5. Update `RestoreSoftDeleted()` in `query_advanced.go`

```go
col := getSoftDeleteColumn(q.model)
var restoreValue any = nil  // NULL-based default
if strat, ok := q.model.(contractsorm.SoftDeleteStrategy); ok {
    restoreValue = strat.RestoreValue()
}
sql, args := builder.BuildUpdate(map[string]any{col: restoreValue})
```

### 6. Update `buildWheresWithSoftDelete` in `builder_where.go`

```go
func (b *Builder) buildWheresWithSoftDelete() (string, []any) {
    var prefix string
    var prefixArgs []any

    if hasSoftDeleteCapability(b.query.model) {
        if strat, ok := b.query.model.(contractsorm.SoftDeleteStrategy); ok {
            // Max-date (or custom) strategy: use range conditions
            switch {
            case b.query.onlySoftDeleted:
                prefix, prefixArgs = strat.IsDeletedCondition(b.quoteIdentifier)
            case b.query.includeSoftDeleted:
                // no filter
            default:
                prefix, prefixArgs = strat.IsActiveCondition(b.quoteIdentifier)
            }
        } else {
            // NULL-based strategy (default)
            col := b.quoteIdentifier(getSoftDeleteColumn(b.query.model))
            switch {
            case b.query.onlySoftDeleted:
                prefix = col + " IS NOT NULL"
            case b.query.includeSoftDeleted:
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
    return prefix + " AND " + base, append(prefixArgs, args...)
}
```

Apply the same pattern to `buildWheresWithSoftDeleteIndex` — note that the
max-date conditions carry a bind parameter, so the placeholder index counter
must be incremented by 1 before the remaining WHERE args.

### 7. Update `isSoftDeleteOperation` detection in `builder_update.go`

Currently checks `m["deleted_at"]` hardcoded. Already partially fixed by the
column-naming work (uses `getSoftDeleteColumn`). Confirm it reads:

```go
softDeleteCol := getSoftDeleteColumn(b.query.model)
if _, has := m[softDeleteCol]; has {
    isSoftDeleteOperation = true
}
```

No additional change needed if `getSoftDeleteColumn` already handles this.

### 8. Schema

The `soft_deleted_at` column must be `NOT NULL` with a sentinel default:

```go
table.Timestamp("soft_deleted_at").Default("9999-12-31 23:59:59").NotNullable()
```

New records must arrive with `soft_deleted_at = MaxSoftDeletedAt`. Options:
- Database-level `DEFAULT '9999-12-31 23:59:59'` (preferred — no ORM hook needed)
- `BeforeCreate` observer that sets `model.SoftDeletedAt = MaxSoftDeletedAt`

---

## Portability of `9999-12-31 23:59:59`

| Database | Max type | Notes |
|---|---|---|
| MySQL / MariaDB | `DATETIME` max = `9999-12-31 23:59:59` | Do NOT use `TIMESTAMP` (max `2038-01-19`) |
| PostgreSQL | `TIMESTAMP` / `TIMESTAMPTZ` | `9999-12-31 23:59:59` is valid |
| SQLite | No hard limit | Stored as text; ISO 8601 comparison works correctly |
| SQL Server | `datetime2` max = `9999-12-31 23:59:59.9999999` | Use `datetime2`, not legacy `datetime` |
| Oracle | `DATE` / `TIMESTAMP` max = `9999-12-31 23:59:59` | Valid |

**Always use UTC. Never use the short-range `TIMESTAMP` type in MySQL.**

---

## Usage Examples

### Model definition

```go
import "github.com/dracory/neat/database/soft_delete"

// New project — soft_deleted_at, NOT NULL, max-date sentinel
type User struct {
    soft_delete.SoftDeletesMaxDate
    ID   uint
    Name string
}

// Laravel-compatible schema — deleted_at, NOT NULL, max-date sentinel
type Post struct {
    soft_delete.DeletedAtMaxDate
    ID    uint
    Title string
}
```

### Schema migration

```go
// SoftDeletesMaxDate
table.Timestamp("soft_deleted_at").Default("9999-12-31 23:59:59").NotNullable()

// DeletedAtMaxDate
table.Timestamp("deleted_at").Default("9999-12-31 23:59:59").NotNullable()
```

### Soft deleting

```go
db.Query().Model(&User{}).Where("id = ?", 1).SoftDelete()
// UPDATE users SET soft_deleted_at = '2026-06-15 10:00:00' WHERE id = 1
```

### Querying active records (default)

```go
var users []User
db.Query().Model(&User{}).Get(&users)
// WHERE soft_deleted_at > '2026-06-15 10:00:00'
```

### Including all records

```go
db.Query().Model(&User{}).WithSoftDeleted().Get(&users)
// (no WHERE clause added)
```

### Only soft-deleted records

```go
db.Query().Model(&User{}).OnlySoftDeleted().Get(&users)
// WHERE soft_deleted_at <= '2026-06-15 10:00:00'
```

### Restoring

```go
db.Query().Model(&User{}).WithSoftDeleted().Where("id = ?", 1).RestoreSoftDeleted()
// UPDATE users SET soft_deleted_at = '9999-12-31 23:59:59' WHERE id = 1
```

### Checking status on struct

```go
if user.IsSoftDeleted() {
    fmt.Println("User is soft-deleted")
}
```

---

## Benefits

1. **No NULL handling** — all queries use simple range comparisons
2. **Better indexing** — a B-tree index on `soft_deleted_at` supports range scans
3. **NOT NULL compatible** — works with strict schemas
4. **Backward compatible** — existing NULL-based models are completely unchanged;
   `SoftDeleteStrategy` is opt-in
5. **Consistent column naming** — uses `soft_deleted_at` by default, matching
   the new default established for all other embeds

---

## Drawbacks

1. **Two implementations to maintain** — `SoftDeletes` and `SoftDeletesMaxDate`
   diverge and must evolve in parallel
2. **New record initialization** — `MaxSoftDeletedAt` must be set explicitly at
   insert time; the zero value `0001-01-01` would immediately report deleted
3. **`IsSoftDeleted()` is time-sensitive** — result depends on `time.Now()`;
   requires mocking in time-sensitive tests
4. **Bind parameter in WHERE** — `soft_deleted_at > ?` adds a bind arg where
   `IS NULL` adds none; `buildWheresWithSoftDeleteIndex` must handle this
5. **Less common** — NULL-based soft deletes are the dominant convention;
   max-date may be surprising to developers unfamiliar with the pattern

---

## Alternatives Considered

### Alternative 1: Replace the NULL implementation globally

**Pros**: Single implementation, no ambiguity  
**Cons**: Breaking change; every existing user must migrate schema and data  
**Decision**: Rejected

### Alternative 2: Boolean `is_deleted` flag

**Pros**: Simple; no NULL or sentinel issues; easy indexing  
**Cons**: Loses the deletion timestamp — essential for audit trails  
**Decision**: Rejected

### Alternative 3: Separate `soft_deleted_at` (NULL) + `is_deleted` (boolean)

**Pros**: Fast boolean index; keeps the full audit timestamp  
**Cons**: Data redundancy; two columns must stay in sync  
**Decision**: Out of scope for this proposal

### Alternative 4: No changes

**Pros**: Zero maintenance cost, zero risk  
**Cons**: Legitimate NOT NULL schemas remain unservable  
**Decision**: Rejected

---

## Implementation Plan

1. Add `MaxSoftDeletedAt` sentinel and both `SoftDeletesMaxDate` and
   `DeletedAtMaxDate` structs to `database/soft_delete/soft_delete.go`
2. Define `SoftDeleteStrategy` interface in `contracts/database/orm/soft_delete.go`
   (extending `SoftDeleteColumnNamer`)
3. Implement `SoftDeleteStrategy` on `SoftDeletesMaxDate`
4. Update `Delete()` in `query_delete.go` — use `SoftDeleteStrategy.SoftDeleteValue()`
   when available
5. Update `RestoreSoftDeleted()` in `query_advanced.go` — use
   `SoftDeleteStrategy.RestoreValue()` when available (currently hardcodes `nil`)
6. Update `buildWheresWithSoftDelete` and `buildWheresWithSoftDeleteIndex` in
   `builder_where.go` — use `IsActiveCondition`/`IsDeletedCondition` for
   `SoftDeleteStrategy` models
7. Confirm `isSoftDeleteOperation` in `builder_update.go` uses
   `getSoftDeleteColumn` (already done)
8. Fix `buildWheresWithSoftDeleteIndex` in `builder_where.go` — the max-date
   conditions (`soft_deleted_at > ?`) add one bind parameter ahead of any user
   WHERE args; the placeholder counter must be incremented by 1 before the
   remaining clauses so numbered placeholders (e.g. `$2`, `$3` in PostgreSQL)
   are generated correctly
9. Add tests:
   - `SoftDeletesMaxDate` and `DeletedAtMaxDate`: `IsSoftDeleted()` with past,
     now, and sentinel timestamps
   - `SoftDelete()` and `RestoreSoftDeleted()` SQL generation for both embeds
   - `WithSoftDeleted`, `OnlySoftDeleted`, `WithoutSoftDeleted` WHERE filters
     for both embeds
   - `buildWheresWithSoftDeleteIndex` placeholder index correctness: verify
     subsequent WHERE args receive the correct `$N` / `?` index when a
     max-date prefix arg is present
10. Auto-initialize new records: during `Insert`, if the model implements
    `SoftDeleteStrategy`, set the soft delete field to `RestoreValue()` before
    executing the INSERT (so in-memory state is correct without a re-fetch)
11. Update `docs/soft-deletes.md` to document the max-date strategy

---

## Open Questions

1. **New record initialization** — **Resolved: the ORM should automatically set
   `SoftDeletedAt = MaxSoftDeletedAt` during `Insert`** when the model implements
   `SoftDeleteStrategy`. A database-level `DEFAULT` alone is insufficient — it
   does not populate the in-memory struct after insert, leaving `IsSoftDeleted()`
   in an incorrect state until the record is re-fetched. The ORM will detect the
   interface and set the sentinel before executing the INSERT.

2. **`time.Now()` inside `IsActiveCondition`/`IsDeletedCondition`** — **Resolved:
   call `time.Now()` internally**, consistent with every other `time.Now()` call
   in the codebase (query builder, soft delete structs, observers). No clock
   abstraction exists anywhere in Neat ORM and introducing one here would be
   inconsistent. Tests that need a fixed timestamp should construct the condition
   SQL directly and assert the bound arg falls within an acceptable window
   (e.g. `time.Since(arg) < time.Second`), the same pattern already used in
   `query_helpers_test.go` and `query_accessors_test.go`.

3. **Interface placement** — **Resolved: `SoftDeleteStrategy` goes in
   `contracts/database/orm/soft_delete.go`** alongside `SoftDeleteColumnNamer`.
   Both interfaces govern soft delete behavior at the contract layer; keeping
   them in the same file makes the relationship explicit.

4. **`DeletedAt`-column max-date variant** — ~~should a `DeletedAtMaxDate` embed
   also be provided?~~ **Resolved: yes.** `DeletedAtMaxDate` is included in the
   proposal. The pattern is consistent: every NULL-based embed has a max-date
   counterpart with the same column name.

---

## References

- Soft Deletes Documentation: `docs/soft-deletes.md`
- Column naming proposal (implemented): `docs/proposals/alternative-soft-delete-column-naming.md`
- Embed structs: `database/soft_delete/soft_delete.go`
- `sql.NullTime` variants: `database/orm/model.go`
- `SoftDeleteColumnNamer` interface: `contracts/database/orm/soft_delete.go`
- Soft delete detection + column resolution: `database/query/query_delete.go`
- WHERE construction: `database/query/builder_where.go`
- Restore implementation: `database/query/query_advanced.go`
- Use The Index, Luke — IS NULL: https://use-the-index-luke.com/sql/where-clause/functions/is-null
- Laravel Eloquent Soft Deletes: https://laravel.com/docs/eloquent#soft-deleting
