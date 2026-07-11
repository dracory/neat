# Proposal: Explicit `SoftDelete()` and `HardDelete()` Methods on the Query Builder

**Status**: Implemented
**Date**: July 11, 2026
**Author**: Neat ORM Team

---

## Summary

Add an explicit `SoftDelete()` method and a `HardDelete()` alias to the query
builder (`Query`) and its contract interface, giving developers a clear,
intentional way to perform either type of deletion. `SoftDelete()` always
soft-deletes (errors if unsupported). `HardDelete()` is an intuitive alias for
the existing `ForceDelete()`, always performing a permanent `DELETE`.

---

## Motivation

### Current behavior

Today, `Query.Delete()` uses **implicit soft-delete detection**: it checks
whether the model implements `SoftDeleteColumnNamer` and, if so, performs an
`UPDATE` instead of a `DELETE`. If the model does not implement the interface,
it performs a hard `DELETE`.

This creates several issues:

1. **Ambiguity** — `Delete()` means different things depending on the model.
   A reader cannot tell at the call site whether a record is being soft-deleted
   or permanently destroyed.

2. **Asymmetry with `ForceDelete()`** — `ForceDelete()` is explicit about
   bypassing soft delete, but there is no explicit counterpart for *requesting*
   a soft delete. The API reads as:

   ```
   Delete()       → maybe soft, maybe hard (implicit)
   ForceDelete()  → always hard (explicit)
   ```

   The proposed API would read as:

   ```
   SoftDelete()   → always soft, error if unsupported (explicit)
   HardDelete()   → always hard (explicit, intuitive name)
   Delete()       → auto-detect (implicit, backward compatible)
   ForceDelete()  → always hard (explicit, Laravel-compatible alias)
   ```

   `SoftDelete()` / `HardDelete()` form the natural antonym pair.
   `ForceDelete()` is retained as a deprecated alias for backward
   compatibility (Laravel convention).

3. **No fail-fast for missing soft-delete support** — If a developer calls
   `Delete()` on a model without soft-delete capability, the record is
   permanently destroyed with no warning. An explicit `SoftDelete()` would
   return an error in this case, preventing accidental data loss.

4. **Inconsistency with the model-level API** — The `soft_delete` package
   already provides `SoftDelete()` on all five embed structs
   (`SoftDeletes`, `SoftDeletedAt`, `DeletedAt`, `SoftDeletesMaxDate`,
   `DeletedAtMaxDate`). The query builder should mirror this naming.

5. **Discoverability** — Developers looking for a `SoftDelete()` method on the
   query builder (by analogy with `ForceDelete()` and the model-level
   `SoftDelete()`) will find nothing. This leads to confusion and potential
   misuse of `Delete()`.

6. **`ForceDelete()` naming is unintuitive** — While `ForceDelete()` follows
   Laravel's convention, `SoftDelete()` / `HardDelete()` is the natural
   antonym pair. Developers who don't come from Laravel will look for
   `HardDelete()` and not find it. Adding `HardDelete()` as an alias costs
   nothing (one-line delegation) and significantly improves discoverability.

### Concrete example

```go
// Current — ambiguous
query.Where("status = ?", "inactive").Delete()
// Is this a soft delete or a hard delete? You have to check the model.

// Proposed — explicit
query.Where("status = ?", "inactive").SoftDelete()
// Always soft delete. Errors if the model doesn't support soft delete.

query.Where("status = ?", "inactive").HardDelete()
// Always hard delete. Permanent removal, bypasses soft delete.
```

---

## Design

### API surface

#### Query builder (`Query`)

```go
// SoftDelete soft-deletes records by setting the soft-delete timestamp column.
// Returns an error if the model does not implement SoftDeleteColumnNamer
// (i.e., does not support soft deletes).
//
// Example:
//   result, err := query.Where("status = ?", "inactive").SoftDelete()
SoftDelete(value ...any) (*Result, error)

// HardDelete permanently deletes records (bypasses soft delete).
// This is an intuitive alias for ForceDelete().
//
// Example:
//   result, err := query.Where("status = ?", "inactive").HardDelete()
HardDelete(value ...any) (*Result, error)
```

#### ToSql (`ToSql`)

```go
// SoftDelete generates the SQL for a soft-delete UPDATE query.
// Returns the SQL string with placeholders.
//
// Example:
//   sql := query.ToSql().SoftDelete()
SoftDelete(value ...any) string

// HardDelete generates the SQL for a permanent DELETE query (bypasses soft delete).
// This is an intuitive alias for ForceDelete().
HardDelete(value ...any) string
```

### Behavior

| Condition | `SoftDelete()` | `HardDelete()` | `Delete()` (unchanged) | `ForceDelete()` (unchanged) |
|---|---|---|---|---|
| Model has soft-delete support | UPDATE sets timestamp | DELETE row | UPDATE sets timestamp | DELETE row |
| Model lacks soft-delete support | **Returns error** | DELETE row | DELETE row | DELETE row |

`HardDelete()` and `ForceDelete()` are functionally identical. `HardDelete()`
is the recommended name going forward; `ForceDelete()` is retained as a
deprecated alias for backward compatibility.

### Error handling

When the model does not implement `SoftDeleteColumnNamer`, `SoftDelete()`
returns:

```go
nil, fmt.Errorf("SoftDelete() requires a model that implements SoftDeleteColumnNamer; use Delete() or HardDelete() instead")
```

This is a fail-fast guard against accidental permanent deletion.

### Events

`SoftDelete()` fires the same events as the current soft-delete path in
`Delete()`:

- **Before**: `EventDeleting` (same as today)
- **After**: `EventDeleted` (same as today)

No new event types are introduced. The existing `deleting` / `deleted` events
are semantically correct — they represent the act of deleting, whether soft or
hard. The distinction between soft and hard delete is already captured by the
method name and the resulting database state.

### Strategy support

`SoftDelete()` respects the `SoftDeleteStrategy` interface, just like the
current soft-delete path in `Delete()`:

- If the model implements `SoftDeleteStrategy`, the custom `SoftDeleteValue()`
  is used (e.g., `time.Now()` for max-date strategies).
- Otherwise, `time.Now()` is used as the default value.

### `includeSoftDeleted` flag

`SoftDelete()` sets `includeSoftDeleted = true` on the clone, matching the
current behavior in `Delete()`'s soft-delete branch. This ensures that
already-soft-deleted records are also updated (idempotent soft delete).

### Interaction with `OnlySoftDeleted()` / `WithSoftDeleted()`

`SoftDelete()` always performs a soft delete regardless of whether
`OnlySoftDeleted()` or `WithSoftDeleted()` was called before it. Unlike
`Delete()`, which checks `!query.includeSoftDeleted && !query.onlySoftDeleted`
to decide between soft and hard, `SoftDelete()` unconditionally builds an
UPDATE. The `includeSoftDeleted = true` setting ensures the UPDATE reaches
all matching rows, including already-soft-deleted ones.

### `ToRawSql()` interaction

`ToRawSql()` returns a `ToSql` with `useValues: true`. Both `SoftDelete()` and
`HardDelete()` work correctly with `ToRawSql()` — `SoftDelete()` checks
`t.useValues` for placeholder replacement, and `HardDelete()` delegates to
`ForceDelete()` which already handles `t.useValues`.

---

## Implementation plan

### 1. Contract changes

**File**: `contracts/database/orm/orm.go`

Add `SoftDelete` and `HardDelete` to the `Query` interface:

```go
// SoftDelete soft-deletes records by setting the soft-delete timestamp column.
// Returns an error if the model does not implement SoftDeleteColumnNamer.
//
// Example:
//   result, err := query.Where("status = ?", "inactive").SoftDelete()
SoftDelete(value ...any) (*Result, error)

// HardDelete permanently deletes records (bypasses soft delete).
// This is an intuitive alias for ForceDelete().
//
// Example:
//   result, err := query.Where("status = ?", "inactive").HardDelete()
HardDelete(value ...any) (*Result, error)
```

Add `SoftDelete` and `HardDelete` to the `ToSql` interface:

```go
SoftDelete(value ...any) string
HardDelete(value ...any) string
```

### 2. Query builder implementation

**File**: `database/query/query_soft_delete.go` (existing file, alongside
`WithSoftDeleted` / `OnlySoftDeleted` / `WithoutSoftDeleted`)

```go
// SoftDelete soft-deletes records by setting the soft-delete timestamp column.
// Returns an error if the model does not implement SoftDeleteColumnNamer.
func (q *Query) SoftDelete(value ...any) (*contractsorm.Result, error) {
    query := q.Clone().(*Query)
    if len(value) > 0 {
        applyConditions(query, value)
    }

    if err := query.validate(); err != nil {
        return nil, err
    }

    // Fail-fast: model must support soft delete
    if !hasSoftDeleteCapability(query.model) {
        return nil, fmt.Errorf(
            "SoftDelete() requires a model that implements SoftDeleteColumnNamer; " +
            "use Delete() or HardDelete() instead",
        )
    }

    // Fire Deleting event
    if !query.withoutEvents && query.model != nil {
        attributes := observer.ExtractModelAttributes(query.model)
        if err := query.dispatcher.DispatchDeleting(
            query.ctx, query.model, query.modelToObserver,
            nil, attributes, nil, query,
        ); err != nil {
            return nil, fmt.Errorf("deleting event error: %w", err)
        }
    }

    // Build UPDATE to set the soft delete column
    query.includeSoftDeleted = true
    builder := NewBuilder(query)
    col := getSoftDeleteColumn(query.model)

    var deleteValue any = time.Now()
    if strat, ok := query.model.(contractsorm.SoftDeleteStrategy); ok {
        deleteValue = strat.SoftDeleteValue()
    }

    sql, args := builder.BuildUpdate(map[string]any{col: deleteValue})
    if sql == "" {
        return nil, fmt.Errorf("failed to build SOFT DELETE query")
    }

    // Execute
    ctx, cancel := query.timeoutContext()
    defer cancel()

    start := time.Now()
    var result interface{ RowsAffected() (int64, error) }
    var err error

    if query.tx != nil {
        result, err = query.tx.ExecContext(ctx, sql, args...)
    } else {
        var dbConn *sql.DB
        dbConn, err = query.DB()
        if err != nil {
            return nil, err
        }
        result, err = dbConn.ExecContext(ctx, sql, args...)
    }

    if err != nil {
        return nil, query.sanitizeError(
            fmt.Errorf("failed to execute SOFT DELETE query: %w", err),
        )
    }
    query.logQuery(sql, args, start)

    // Fire Deleted event
    if !query.withoutEvents && query.model != nil {
        attributes := observer.ExtractModelAttributes(query.model)
        if err := query.dispatcher.DispatchDeleted(
            query.ctx, query.model, query.modelToObserver,
            nil, attributes, nil, query,
        ); err != nil {
            return nil, fmt.Errorf("deleted event error: %w", err)
        }
    }

    rowsAffected, _ := result.RowsAffected()
    return &contractsorm.Result{RowsAffected: rowsAffected}, nil
}
```

### 3. HardDelete implementation

**File**: `database/query/query_advanced.go`

```go
// HardDelete permanently deletes a record (bypasses soft delete).
// This is an intuitive alias for ForceDelete().
func (q *Query) HardDelete(value ...any) (*contractsorm.Result, error) {
    return q.ForceDelete(value...)
}
```

**File**: `database/query/to_sql.go`

```go
// HardDelete generates the SQL for a permanent DELETE query (bypasses soft delete).
// This is an intuitive alias for ForceDelete().
func (t *ToSql) HardDelete(value ...any) string {
    return t.ForceDelete(value...)
}
```

### 4. Deprecate ForceDelete

Add a deprecation comment to the existing `ForceDelete()` methods in both the
contract and implementation. The method is **not removed** — it remains as a
backward-compatible alias:

```go
// ForceDelete permanently deletes a record (bypasses soft delete).
//
// Deprecated: Use HardDelete() instead.
ForceDelete(value ...any) (*Result, error)
```

### 5. ToSql implementation

**File**: `database/query/to_sql.go`

```go
// SoftDelete generates the SQL for a soft-delete UPDATE query.
func (t *ToSql) SoftDelete(value ...any) string {
    query := t.query.Clone().(*Query)
    if len(value) > 0 {
        applyConditions(query, value)
    }

    // Build UPDATE to set the soft delete column
    query.includeSoftDeleted = true
    builder := NewBuilder(query)
    col := getSoftDeleteColumn(query.model)

    var deleteValue any = time.Now()
    if strat, ok := query.model.(contractsorm.SoftDeleteStrategy); ok {
        deleteValue = strat.SoftDeleteValue()
    }

    sql, args := builder.BuildUpdate(map[string]any{col: deleteValue})
    if t.useValues {
        return t.replacePlaceholdersWithValues(sql, args)
    }
    return t.replacePlaceholders(sql, args)
}
```

### 6. Tests

**File**: `database/query/query_soft_delete_test.go` (new)

Test cases:

- `TestSoftDelete_Success` — model with `SoftDeletes` embed, verify UPDATE
  sets `soft_deleted_at` to non-NULL
- `TestSoftDelete_DeletedAtColumn` — model with `DeletedAt` embed, verify
  UPDATE sets `deleted_at`
- `TestSoftDelete_MaxDateStrategy` — model with `SoftDeletesMaxDate`, verify
  `SoftDeleteValue()` is used
- `TestSoftDelete_NoSoftDeleteCapability` — model without any soft-delete
  embed, verify error is returned and no DELETE is executed
- `TestSoftDelete_WithWhereConditions` — verify WHERE clauses are applied
- `TestSoftDelete_EventsFired` — verify `deleting` and `deleted` events fire
- `TestSoftDelete_WithoutEvents` — verify events are suppressed
- `TestSoftDelete_Idempotent` — soft-deleting an already-soft-deleted record
  updates the timestamp (does not error)
- `TestSoftDelete_InTransaction` — verify soft delete works within a
  transaction
- `TestToSql_SoftDelete` — verify SQL generation without execution
- `TestHardDelete_BypassesSoftDelete` — model with `SoftDeletes` embed, verify
  row is permanently deleted (not soft-deleted)
- `TestHardDelete_NoSoftDeleteCapability` — model without soft-delete embed,
  verify row is permanently deleted (same as `ForceDelete`)
- `TestHardDelete_MatchesForceDelete` — verify `HardDelete()` and
  `ForceDelete()` produce identical SQL and results
- `TestToSql_HardDelete` — verify SQL generation without execution

### 7. Integration tests

Add `SoftDelete()` and `HardDelete()` test cases to each database integration
test suite:

- `integration_tests/mysql/mysql_soft_delete_test.go`
- `integration_tests/postgres/postgres_soft_delete_test.go`
- `integration_tests/sqlite/sqlite_soft_delete_test.go`
- `integration_tests/oracle/oracle_soft_delete_test.go`
- `integration_tests/sqlserver/sqlserver_soft_delete_test.go`
- `integration_tests/turso/turso_soft_delete_test.go`

### 8. Documentation

- Update `docs/soft-deletes.html` with `SoftDelete()` and `HardDelete()` method
  documentation
- Update `docs/api-reference.html` with the new methods
- Update `docs/query-builder.html` if it references delete methods

### 9. Examples

Update the existing soft-delete examples to demonstrate the new explicit
methods:

- `examples/soft-deletes/main.go` — add `SoftDelete()` and `HardDelete()` usage
- `examples/soft-deletes/main_test.go` — add test cases
- `examples/soft-delete-alt-deleted-at/main.go` — add `SoftDelete()` usage with
  `deleted_at` column
- `examples/soft-delete-max-date/main.go` — add `SoftDelete()` usage with
  max-date strategy

---

## Backward compatibility

- `Delete()` behavior is **unchanged** — it continues to auto-detect soft
  delete capability. No existing code breaks. **Not deprecated**, but a
  documentation note is added recommending `SoftDelete()` or `HardDelete()`
  when the model's soft-delete support is known (see below).
- `SoftDelete()` is a **new method** — purely additive.
- `HardDelete()` is a **new method** — purely additive, delegates to
  `ForceDelete()`.
- `ForceDelete()` is **deprecated** but **not removed** — existing code
  continues to work. A deprecation comment directs users to `HardDelete()`.
- `RestoreSoftDeleted()` is **unchanged**.

### Deprecation policy

| Method | Status | Reason |
|---|---|---|
| `ForceDelete()` | **Deprecated** | Pure rename to `HardDelete()`. No behavioral difference. Keeping both as primary names would be confusing. |
| `Delete()` | **Not deprecated** | Has a legitimate use case: generic code that works with both soft-deletable and non-soft-deletable models (e.g., a generic repository). Deprecating it would force unnecessary churn with no benefit. |
| `Destroy()` | **Not deprecated** | Sequelize-compatible alias for `Delete()`. Same auto-detect behavior. Gets a documentation note recommending explicit methods. |

### Documentation note for `Delete()`

While `Delete()` is not deprecated, its doc comment should be updated to
recommend the explicit methods:

```go
// Delete deletes records from the database.
//
// If the model implements SoftDeleteColumnNamer, this performs a soft delete
// (UPDATE setting the soft-delete timestamp). Otherwise, it performs a hard
// DELETE.
//
// For clarity, prefer SoftDelete() when the model supports soft deletes, or
// HardDelete() when you want a permanent deletion. Use Delete() only when the
// model's soft-delete capability is unknown or intentionally auto-detected.
//
// Example:
//   result, err := query.Where("status = ?", "inactive").Delete()
Delete(value ...any) (*Result, error)
```

### Documentation note for `Destroy()`

`Destroy()` is a Sequelize-compatible alias for `Delete()` and shares the same
ambiguity. Its doc comment should also be updated:

```go
// Destroy is an alias for Delete, providing Sequelize-style syntax.
//
// If the model implements SoftDeleteColumnNamer, this performs a soft delete.
// Otherwise, it performs a hard DELETE.
//
// For clarity, prefer SoftDelete() or HardDelete() when the model's
// soft-delete support is known.
//
// Deprecated: Prefer SoftDelete() or HardDelete() for explicit intent.
Destroy(value ...any) (*Result, error)
```

Note: `Destroy()` is marked as deprecated in its doc comment since it adds no
functionality over `Delete()` and the Sequelize-style naming is not needed when
explicit methods are available.

## Alternatives considered

### A. Rename `Delete()` to `SoftDelete()` and make `Delete()` hard-only

**Rejected** — Breaking change. All existing `Delete()` callers would need to
update their code. The auto-detect behavior of `Delete()` is convenient and
well-established.

### B. Add a `SoftDeleteOnly()` boolean flag on the query

**Rejected** — Fluent method chaining is more idiomatic in Go ORMs and matches
the existing `ForceDelete()` pattern. A boolean flag would require:
`query.Where(...).SoftDeleteOnly(true).Delete()` which is verbose and
error-prone.

### C. Make `Delete()` always soft-delete and remove auto-detect

**Rejected** — Breaking change. Models without soft-delete support would
silently fail or require a different method. The auto-detect behavior is a
feature, not a bug — it lets developers swap embeds without changing query
code.

### D. Add `SoftDelete()` as a no-op alias for `Delete()`

**Rejected** — This defeats the purpose. The value of `SoftDelete()` is the
**fail-fast error** when the model doesn't support soft delete. Without that,
it's just another name for the same ambiguous behavior.

---

## Naming consistency

The proposed method aligns with the existing naming conventions:

| Level | Method | Purpose |
|---|---|---|
| Model embed (`soft_delete.SoftDeletes`) | `SoftDelete()` | Marks struct field in memory |
| Query builder (`Query`) | `SoftDelete()` | Executes soft-delete UPDATE in DB |
| Query builder (`Query`) | `HardDelete()` | Executes hard DELETE in DB (recommended) |
| Query builder (`Query`) | `ForceDelete()` | Alias for `HardDelete()` (deprecated, Laravel-compatible) |
| Query builder (`Query`) | `Delete()` | Auto-detects soft vs hard (not deprecated, see note) |
| Query builder (`Query`) | `RestoreSoftDeleted()` | Reverses a soft delete |
| ToSql (`ToSql`) | `SoftDelete()` | Generates soft-delete SQL |
| ToSql (`ToSql`) | `HardDelete()` | Generates hard-delete SQL (recommended) |
| ToSql (`ToSql`) | `ForceDelete()` | Alias for `HardDelete()` (deprecated) |

---

## Open questions

1. **Should `Delete()` be deprecated?** — No. `Delete()` remains useful for
   generic code that works with both soft-deletable and non-soft-deletable
   models (e.g., a generic repository). Deprecating it would force unnecessary
   churn with no benefit. Instead, a documentation note recommends `SoftDelete()`
   or `HardDelete()` when the model's soft-delete support is known. See the
   **Deprecation policy** section for details.

2. **Should a `SoftDeleting` / `SoftDeleted` event pair be added?** — Not in
   this proposal. The existing `deleting` / `deleted` events already cover
   soft deletes. Adding separate events would require all event listeners to
   register twice. This can be revisited if there's demand.

3. **Should `SoftDelete()` accept a custom timestamp?** — Not in v1. The
   method uses `time.Now()` (or `SoftDeleteValue()` for strategy models).
   A custom timestamp can be added later as an optional parameter if needed.
