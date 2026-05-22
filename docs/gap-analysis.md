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

## 3. Read/Write Replica Support ✅ COMPLETED

`neat` exports `ReplicaConfig` and `ConnectionConfig.Read`/`Write` fields. `buildQuery` in `database/orm/orm.go` opens separate `*sql.DB` connections for read replicas and write primaries. `Query` has `readDB`/`writeDB` fields with `readConn()`/`writeConn()` helpers; all SELECT paths (`Find`, `First`, `Get`, aggregates, `Pluck`, `Scan`, `Chunk`, `Paginate`, `Cursor`) route to the read replica while write paths use the write primary. Integration tests cover both single-connection and read/write-separated configs.

**Files**: `neat/config.go`, `neat/database/db/config_builder.go`, `neat/database/orm/orm.go`, `neat/database/query/query.go`.

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

## 7. JSON Where Clauses ✅ COMPLETED

All 9 JSON where methods were already fully implemented in `neat` — they generate correct `JSON_CONTAINS`, `JSON_CONTAINS_PATH`, and `JSON_LENGTH` SQL. The original gap analysis was incorrect.

---

## 8. Subquery / Nested WHERE Support ✅ COMPLETED (WhereExists)

`WhereExists` is now implemented: the callback receives a cloned `Query`, builds its SELECT SQL, and the result is embedded as `EXISTS (SELECT ...)`. Full arbitrary subquery support in `Where`/`Table` is a larger scope; `WhereExists` covers the primary use case.

---

## 9. Soft Delete via Model vs. Manual ✅ COMPLETED

`neat` soft-delete is now fully automatic at the builder level. `buildWheresWithSoftDelete()` in `builder.go` inspects `hasSoftDeleteCapability(model)` and prepends:
- `deleted_at IS NULL` by default (excludes soft-deleted rows)
- `deleted_at IS NOT NULL` when `OnlyTrashed()` is active
- No filter when `WithTrashed()` is active

This applies to both `BuildSelect` and `BuildDelete`. `Delete` already did a soft-delete UPDATE when the model has `DeletedAt`.

---

## 10. Struct Scan Field Mapping ✅ COMPLETED

`neat`'s `scanRows` now uses name/tag-based mapping via `structScanDests` and `copyScanResults` helpers. Column names are matched to struct fields by checking `db`, `neat`, and `gorm` tags (in that order), falling back to a snake_case conversion of the Go field name. Unmatched columns scan into a `*any` placeholder.

---

## 11. `Turso` Driver Support ✅ COMPLETED (parsing)

`parseDSN` in `database/db.go` already handles `turso://` DSNs. The `Turso` driver struct exists in `database/driver/turso.go` but `Open` returns an error (libSQL dependency not yet added). DSN parsing and driver wiring are done; the actual libSQL connection requires adding the `tursodatabase/go-libsql` dependency separately.

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
