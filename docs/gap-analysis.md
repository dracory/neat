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

## 13. Unit Test Gaps — Tests That Don't Catch Regressions

**Root problem**: All existing unit tests (`database/query/to_sql_test.go`, `database/orm/orm_test.go`, `database/db/config_builder_test.go`) only assert non-empty / non-nil / non-panic. They would pass even if every implemented feature were reverted. The following unit tests need to be **created** to provide real regression coverage.

### 13.1 `database/query/query_routing_test.go` ✅ CREATED

Tests that `readConn()` / `writeConn()` / `ReadDB()` / `DB()` route to the correct `*sql.DB` instance.

| Test | Asserts |
|---|---|
| `TestReadConnFallsBackToPrimary` | `readConn()` returns `db` when `readDB == nil` |
| `TestReadConnUsesReplicaWhenSet` | `readConn()` returns `readDB` when set |
| `TestWriteConnFallsBackToPrimary` | `writeConn()` returns `db` when `writeDB == nil` |
| `TestWriteConnUsesWriteWhenSet` | `writeConn()` returns `writeDB` when set |
| `TestNewQueryWithReplicasSetsFields` | `NewQueryWithReplicas` sets `readDB` and `writeDB` |
| `TestClonePropagatesReplicas` | `Clone()` copies `readDB` and `writeDB` to the clone |
| `TestDBErrorsDuringTransaction` | `DB()` returns error when `tx != nil` |
| `TestReadDBErrorsDuringTransaction` | `ReadDB()` returns error when `tx != nil` |

### 13.2 `database/query/query_log_test.go` ✅ CREATED

Tests for query logging and slow threshold behaviour (§1.2 / §12).

| Test | Asserts |
|---|---|
| `TestLogQueryRecordsDuration` | `QueryLog` entry has `Time > 0` after a query |
| `TestLogQuerySlowThreshold` | Logger receives `Warningf` when elapsed ≥ `SlowThreshold` |
| `TestLogQueryNoWarnBelowThreshold` | No warning logged when elapsed < `SlowThreshold` |
| `TestEnableDisableQueryLog` | `EnableQueryLog`/`DisableQueryLog` toggle capture |
| `TestFlushQueryLog` | `FlushQueryLog` empties the log |

### 13.3 `database/query/soft_delete_builder_test.go` ✅ CREATED

Tests that the query builder injects the correct `deleted_at` filter (§9).

| Test | Asserts |
|---|---|
| `TestBuildSelectInjectsSoftDeleteFilter` | Generated SQL contains `deleted_at IS NULL` for soft-delete model |
| `TestBuildSelectWithTrashedSkipsFilter` | `WithTrashed()` produces no `deleted_at` clause |
| `TestBuildSelectOnlyTrashedFilter` | `OnlyTrashed()` produces `deleted_at IS NOT NULL` |
| `TestBuildDeleteSoftDeleteUpdateNotHardDelete` | `BuildDelete` emits `UPDATE … SET deleted_at=` for soft-delete model |
| `TestBuildSelectNoFilterForNonSoftDeleteModel` | Plain model gets no `deleted_at` in SQL |

### 13.4 `database/query/where_exists_test.go` ❌ MISSING

Tests for `WhereExists` subquery generation (§8).

| Test | Asserts |
|---|---|
| `TestWhereExistsGeneratesExistsSql` | SQL contains `EXISTS (SELECT` |
| `TestWhereExistsCallbackReceivesClone` | Callback's query is independent of outer query |

### 13.5 `database/query/insert_get_id_test.go` ✅ CREATED

Tests for `InsertGetId` driver branching (§6).

| Test | Asserts |
|---|---|
| `TestInsertGetIdPostgresUsesReturning` | SQL generated contains `RETURNING id` for postgres driver |
| `TestInsertGetIdMysqlUsesLastInsertId` | SQL generated does NOT contain `RETURNING` for mysql driver |

### 13.6 `database/query/scan_mapping_test.go` ✅ CREATED

Tests for name/tag-based struct scanning (§10). **Writing these tests also revealed and fixed a real bug**: `structFieldColumnName` was not treating `neat:"col"` as a plain column name (only `db` tag had that treatment), so `neat`-tagged fields silently fell through to gorm-tag resolution. Fixed in `query.go`.

| Test | Asserts |
|---|---|
| `TestScanRowsByDbTag` | `db:"col"` tag maps column to field correctly |
| `TestScanRowsByNeatTag` | `neat:"col"` tag maps column to field correctly |
| `TestScanRowsByGormTag` | `gorm:"column:col"` maps correctly |
| `TestScanRowsBySnakeCase` | Field `UserName` maps from column `user_name` with no tag |
| `TestScanRowsUnmatchedColumnIgnored` | Extra columns in result don't panic or error |
| `TestScanRowsIntoSlice` | Scanning multiple rows into `[]Struct` populates all elements |

### 13.7 `database/query/connection_switch_test.go` ✅ CREATED

Tests for `Query.Connection(name)` switching (§5). **Note**: `Connection()` returns the original query on unknown names rather than an error — tests document actual behaviour.

| Test | Asserts |
|---|---|
| `TestConnectionSwitchUnknownNameReturnsSelf` | `Connection("nonexistent")` returns the original query |
| `TestConnectionSwitchReturnsNewQuery` | Returned query is a different instance for a valid name |
| `TestConnectionSwitchUsesCorrectDriver` | Returned query's `Driver()` matches the named connection's driver |
| `TestConnectionSwitchEmptyNameReturnsSelf` | `Connection("")` returns the original query |

### 13.8 `database/db/config_builder_replica_test.go` ✅ CREATED

Tests for `ReplicaConfig` fields on `ConnectionConfig` (§3).

| Test | Asserts |
|---|---|
| `TestConnectionConfigReadFieldSet` | `Read` slice is stored and accessible |
| `TestConnectionConfigWriteFieldSet` | `Write` slice is stored and accessible |
| `TestReplicaConfigFields` | `ReplicaConfig` struct has all five fields (Host, Port, Database, Username, Password) |

### 13.9 `database/orm/buildquery_replica_test.go` ❌ STILL MISSING

Tests that `buildQuery` in `orm.go` wires replicas correctly (§3). Deferred — requires a real database available in the test environment to open replica connections.

| Test | Asserts |
|---|---|
| `TestBuildQueryNoReplicasUsesNewQuery` | With empty `Read`/`Write`, returns plain `*query.Query` (no `readDB`) |
| `TestBuildQueryWithReadReplicaOpensReadDB` | With `Read` set, returned query's `ReadDB()` differs from `DB()` |
| `TestBuildQueryWritePrimaryOverridesPrimary` | With `Write` set, returned query's `DB()` is the write connection |

### 13.10 `config_test.go` ✅ CREATED (top-level `neat` package)

Tests for `neat.ReplicaConfig` propagation to `db.ConnectionConfig` (§3).

| Test | Asserts |
|---|---|
| `TestReplicaConfigPropagatedToDbConfig` | `neat.New` with `Read`/`Write` replicas produces `db.ConnectionConfig` with matching replica entries |
| `TestDatabaseTypeAlias` | `neat.Database` is assignable to `*database.Database` |

### 13.11 `database/query/transaction_hooks_test.go` ✅ CREATED

Tests for transaction lifecycle callbacks (§2).

| Test | Asserts |
|---|---|
| `TestBeforeCommitCalledOnCommit` | Registered `BeforeCommit` callback executes before commit |
| `TestAfterCommitCalledOnCommit` | Registered `AfterCommit` callback executes after commit |
| `TestBeforeRollbackCalledOnRollback` | Registered `BeforeRollback` callback executes on rollback |
| `TestAfterRollbackCalledOnRollback` | Registered `AfterRollback` callback executes after rollback |
| `TestBeforeCommitErrorAbortsCommit` | If `BeforeCommit` callback returns error, transaction is rolled back |

---

## 14. Integration Test Coverage Gaps

### Previously missing, now added
- `mysql_connection_test.go`, `postgres_connection_test.go` — connection switching, read/write separation, pool settings
- `postgres_raw_test.go`, `postgres_where_test.go`, `postgres_where_advanced_test.go`, `postgres_where_any_all_advanced_test.go`
- `postgres_update_test.go`, `postgres_query_load_test.go`, `postgres_query_association_test.go`, `postgres_query_belongs_to_test.go`

### Still missing integration coverage
| Area | Gap |
|---|---|
| Read/write routing with real dual DB | No test that reads go to replica and writes go to primary using *different* hosts |
| `InsertGetId` Postgres `RETURNING id` | No integration test asserting the returned ID is non-zero |
| `Query.Connection(name)` switching | Integration test needed (connection test exercises `Database.Connection`, not `Query.Connection`) |
| `SlowThreshold` warning in integration | No test that triggers and captures a slow-query log entry against a real DB |
| Transaction hooks (`BeforeCommit` etc.) | Integration test needed to confirm callbacks fire with a real DB |

---

## Priority Recommendations (updated)

1. ~~**Critical** — `query_routing_test.go` (§13.1)~~ ✅ Done.
2. ~~**Critical** — `soft_delete_builder_test.go` (§13.3)~~ ✅ Done.
3. ~~**Critical** — `scan_mapping_test.go` (§13.6)~~ ✅ Done — also revealed and fixed a `neat` tag bug.
4. ~~**High** — `insert_get_id_test.go` (§13.5)~~ ✅ Done.
5. ~~**High** — `transaction_hooks_test.go` (§13.11)~~ ✅ Done.
6. ~~**High** — `query_log_test.go` (§13.2)~~ ✅ Done.
7. ~~**Medium** — `connection_switch_test.go` (§13.7)~~ ✅ Done. `where_exists_test.go` (§13.4) ❌ Still pending.
8. ~~**Medium** — `config_builder_replica_test.go` (§13.8), `config_test.go` (§13.10)~~ ✅ Done. `buildquery_replica_test.go` (§13.9) deferred — needs real DB.
9. **Low** — Fix existing `to_sql_test.go` assertions to check SQL correctness, not just non-empty.
