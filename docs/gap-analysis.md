# Gap Analysis: neat vs eloquent

**Date**: May 2026  
**Summary**: `neat` uses native `database/sql` + hand-rolled query builder; `eloquent` uses GORM under the hood. The contracts (interfaces) are **identical** — gaps are purely in the implementation layer.

---

## 1. Query Implementation Gaps

### 1.1 Stub / Not-Implemented Methods ✅ COMPLETED

These methods exist in `neat`'s `Query` struct but return errors or are no-ops:

| Method | Status in neat | eloquent status |
|---|---|---|
| `FindOrFail` | `return fmt.Errorf("not implemented")` | Fully implemented |
| `FirstOrFail` | `return fmt.Errorf("not implemented")` | Fully implemented |
| `Save` | `return fmt.Errorf("not implemented")` | Fully implemented |
| `SaveQuietly` | `return fmt.Errorf("not implemented")` | Fully implemented |
| `Load` | Returns error "lazy loading not fully implemented yet" | Fully implemented |
| `LoadMissing` | Returns error "lazy loading not fully implemented yet" | Fully implemented |
| `Without` | No-op (`return q`) | Fully implemented |
| `WithCount` | No-op (`return q`) | Fully implemented |
| `WithExists` | No-op (`return q`) | Fully implemented |
| `WhereExists` | No-op (TODO comment) | Fully implemented |
| `WhereAny` | No-op (TODO comment) | Fully implemented |
| `WhereAll` | No-op (TODO comment) | Fully implemented |
| `WhereNone` | No-op (TODO comment) | Fully implemented |
| `LockForUpdate` | No-op (`return q`) | Implemented (SELECT ... FOR UPDATE) |
| `SharedLock` | No-op (`return q`) | Implemented (SELECT ... LOCK IN SHARE MODE) |
| `InRandomOrder` | No-op (`return q`) | Implemented (ORDER BY RAND()/RANDOM()) |
| `Scopes` | No-op (`return q`) | Fully implemented |
| `Omit` | No-op (`return q`) | Fully implemented |

### 1.2 Query Logging Duration ✅ COMPLETED

All query log entries in `neat` now record actual execution duration in milliseconds via the `logQuery(sql, bindings, start)` helper. Slow-query warnings are emitted to the logger when `DBConfig.SlowThreshold` (ms) is set and exceeded.

---

## 2. Transaction Lifecycle Hooks ✅ COMPLETED

`eloquent` supports four transaction callback hooks; `neat` has the interface stubs but the implementations are empty (`{}`):

| Hook | neat | eloquent |
|---|---|---|
| `BeforeCommit(func() error)` | Stub (no-op) | Executes before commit; rolls back on error |
| `AfterCommit(func() error)` | Stub (no-op) | Executes after commit; returns `MultiCallbackError` |
| `BeforeRollback(func() error)` | Stub (no-op) | Executes before rollback |
| `AfterRollback(func() error)` | Stub (no-op) | Executes after rollback |

**File**: `d:\PROJECTs\_modules_dracory\neat\database\query\query.go` lines ~1717–1718.

---

## 3. Read/Write Replica Support ✅ COMPLETED (fields added; query routing not yet wired)

`eloquent`'s `ConnectionConfig` has `Read []database.Config` and `Write []database.Config` fields and uses GORM's `dbresolver` plugin to route reads to replicas and writes to primaries.

`neat`'s `ConnectionConfig` has **no** `Read`/`Write` fields — all queries go to a single connection. No read/write splitting is possible.

**File**: `d:\PROJECTs\_modules_dracory\neat\config.go:29-46` vs `d:\PROJECTs\_modules_dracory\eloquent\config.go:28-47`.

---

## 4. EventBus / Model Lifecycle Events (Public API) ✅ COMPLETED

`eloquent` exposes a top-level `EventBus` with `Listen`, `Dispatch`, and `Forget` methods and named constants (`EventCreating`, `EventCreated`, etc.) wired to the `WithEventBus(eventBus)` option.

`neat` has no equivalent public `EventBus`. Model lifecycle events exist internally via `observer.Dispatcher` but there is no way for application code to subscribe to events via a bus — only via the `Observe(model, observer)` pattern.

**File**: `d:\PROJECTs\_modules_dracory\eloquent\event.go` (missing entirely from neat).

---

## 5. Connection Switching (`Database.Connection`) ✅ COMPLETED

`Database.Connection(name)` was already fully implemented in `neat`. `Query.Connection(name)` was a no-op and is now implemented: it looks up the named connection in `dbConfig`, constructs the appropriate driver via `newDriverForDialect`, builds the DSN, opens a new `*sql.DB`, and returns a fresh `Query` scoped to that connection.

---

## 6. `InsertGetId` Implementation ✅ COMPLETED

`InsertGetId` is now fully implemented with driver-aware ID retrieval: Postgres uses `INSERT ... RETURNING id` with `QueryRowContext`, all other drivers use `ExecContext` + `LastInsertId()`. The previous variable-shadowing bug (`err` in else-branch) is also fixed.

---

## 7. JSON Where Clauses

`neat` has the JSON where methods in the contract but not implemented in the query builder:

| Method | neat | eloquent |
|---|---|---|
| `WhereJsonContains` | Not in builder (no SQL generated) | Implemented |
| `OrWhereJsonContains` | Not in builder | Implemented |
| `WhereJsonDoesntContain` | Not in builder | Implemented |
| `OrWhereJsonDoesntContain` | Not in builder | Implemented |
| `WhereJsonContainsKey` | Not in builder | Implemented |
| `OrWhereJsonContainsKey` | Not in builder | Implemented |
| `WhereJsonDoesntContainKey` | Not in builder | Implemented |
| `OrWhereJsonDoesntContainKey` | Not in builder | Implemented |
| `WhereJsonLength` | Not in builder | Implemented |

---

## 8. Subquery / Nested WHERE Support

`eloquent` supports subqueries in `Where`, `WhereExists`, `Count` (with auto-subquery for GROUP BY / DISTINCT), and `Table` (derived table).

`neat`'s builder has no subquery support — `WhereExists` is a no-op and `Table` / `Where` only accept strings.

---

## 9. Soft Delete via Model vs. Manual

`eloquent` soft-delete is model-driven: GORM auto-detects `DeletedAt gorm.DeletedAt` and handles filter injection globally per query.

`neat` soft-delete is implemented manually in `query.go` with explicit `withTrashed` / `onlyTrashed` flags, but the `Delete` method does not automatically set `deleted_at` — it calls a hard `DELETE` unless the query builder checks the model's soft-delete capability at runtime. This needs audit.

---

## 10. Struct Scan Field Mapping ✅ COMPLETED

`neat`'s `scanRows` now uses name/tag-based mapping via `structScanDests` and `copyScanResults` helpers. Column names are matched to struct fields by checking `db`, `neat`, and `gorm` tags (in that order), falling back to a snake_case conversion of the Go field name. Unmatched columns scan into a `*any` placeholder.

---

## 11. `Turso` Driver Support

`eloquent` parses `turso://` DSNs and supports the Turso (libSQL) driver via `database/gorm/turso.go`.

`neat` has a `turso` driver stub in `database/driver/` but `parseDSN` in `db.go` does not handle `turso://` scheme.

---

## 12. `SlowThreshold` / Debug Logging ✅ COMPLETED

`SlowThreshold` is now propagated from `neat.DBConfig` → `db.DBConfig` → `Query`. The `logQuery` helper emits a `Warningf` log entry whenever a query's elapsed time meets or exceeds the threshold (in milliseconds).

---

## 13. Integration Test Coverage Gaps (mysql)

Eloquent has these test files that neat does not:

| Test File | Description |
|---|---|
| `mysql_connection_test.go` | Connection lifecycle, ping, reconnect |
| `mysql_query_omit_test.go` | Omit column behaviour |
| `mysql_query_scopes_test.go` | Reusable query scopes |
| `mysql_query_log_test.go` | Query log capture |
| `mysql_query_lock_test.go` | SELECT FOR UPDATE / SHARE MODE |
| `mysql_where_any_all_advanced_test.go` | WhereAny/WhereAll/WhereNone |
| `mysql_soft_delete_test.go` | Full soft delete lifecycle |

---

## Priority Recommendations

1. **High** — Fix struct scan field mapping (positional → name/tag based) — breaks real-world models.
2. **High** — Implement `Save` / `SaveQuietly` — core ORM pattern.
3. **High** — Implement `LockForUpdate` / `SharedLock` — required for concurrent writes.
4. **Medium** — Implement transaction hooks (`BeforeCommit`, `AfterCommit`, `BeforeRollback`, `AfterRollback`).
5. **Medium** — Add `Connection(name)` switching at `Database` level.
6. **Medium** — Implement `WhereExists` and subquery support.
7. **Medium** — Implement `Scopes`, `Omit`, `InRandomOrder`.
8. **Low** — Add `EventBus` public API.
9. **Low** — Add read/write replica support (`Read`/`Write` in `ConnectionConfig`).
10. **Low** — Wire up `SlowThreshold` logging and query duration tracking.
