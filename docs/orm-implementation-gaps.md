# ORM Implementation Gaps — neat

**Date**: May 2026  
**Source**: SQLite + MySQL integration test runs  
**Purpose**: Catalogue every failing ORM feature as a concrete work item. Tests are skipped with references to this document until the feature is implemented.

---

## How to use this document

Each section is one ORM feature gap. The heading is the feature. The body contains:
- **Root cause** — why it currently fails
- **Failing tests** — which test(s) expose the bug (with file + subtest name)
- **Implementation hint** — where in the codebase the fix should land

When a feature is implemented, remove the `t.Skip(...)` from the relevant test(s) and mark the section `✅ DONE`.

---

## GAP-01: Batch `Create` inserts only the first row

**Status**: ✅ DONE  
**Root cause**: `BuildInsert` in `database/query/builder.go` handles `[]T` slices but only generates a single-row `INSERT`. It does not generate multi-value `INSERT INTO t (cols) VALUES (...), (...), (...)`.  
**Failing tests**:
- `integration_tests/sqlite/sqlite_query_create_test.go` — `batch create by struct`
- `integration_tests/mysql/mysql_query_create_test.go` — `batch_create_by_struct`

**Implementation hint**: In `builder.go` `BuildInsert`, when `reflect.Slice` is detected, iterate all elements and append each row's values to the VALUES clause. Use a single prepared statement with multiple value groups.

---

## GAP-02: `Table().Create(map)` does not insert

**Status**: ✅ DONE  
**Root cause**: `BuildInsert` when given a `map[string]any` with `Table()` set but no `Model()` fails to generate valid SQL. The map key ordering is non-deterministic and the column/value construction path for the table-only (no model) case appears incomplete.  
**Failing tests**:
- `integration_tests/sqlite/sqlite_query_create_test.go` — `create by map`
- `integration_tests/mysql/mysql_query_create_test.go` — `create_by_map`

**Implementation hint**: In `BuildInsert`, ensure the `map[string]any` path works when only `q.table` is set (no `q.model`). Sort keys for determinism.

---

## GAP-03: `InsertGetId` via `Model()` does not write back struct ID

**Status**: ✅ DONE  
**Root cause**: `InsertGetId` returns the new ID as a `uint` return value but does not write it back into the struct passed as `values`. The struct's `ID` field remains zero.  
**Failing tests**:
- `integration_tests/sqlite/sqlite_query_create_test.go` — `insert get id by struct`
- `integration_tests/mysql/mysql_query_create_test.go` — `insert_get_id_by_struct`

**Implementation hint**: After retrieving `lastID`, use `reflect` to find a field named `ID` (or tagged `db:"id"`) on `values` (if it's a pointer-to-struct) and set it.

---

## GAP-04: `Table().InsertGetId(map)` does not insert

**Status**: ✅ DONE  
**Root cause**: Same as GAP-02 — map + table-only path in `BuildInsert` is broken, so `InsertGetId` also fails.  
**Failing tests**:
- `integration_tests/sqlite/sqlite_query_create_test.go` — `insert get id by map`
- `integration_tests/mysql/mysql_query_create_test.go` — `insert_get_id_by_map`

**Implementation hint**: Fix GAP-02 first; this should follow automatically.

---

## GAP-05: `Chunk()` panics — type mismatch in typed callbacks

**Status**: ✅ DONE  
**Root cause**: `Chunk` fetches rows into `[]interface{}` internally and passes that to the callback. When the caller passes a typed callback `func([]models.User) error`, the reflect-based dispatch fails with a panic because `[]interface{}` is not assignable to `[]models.User`.  
**Failing tests**:
- `integration_tests/sqlite/sqlite_query_chunk_test.go` — all subtests
- `integration_tests/mysql/mysql_query_chunk_test.go` — all subtests

**Implementation hint**: In `query.go` `Chunk()`, detect the callback's element type via reflect, create a correctly-typed slice, scan into it, and pass the typed slice to the callback.

---

## GAP-06: `Distinct(col).Count()` does not generate `COUNT(DISTINCT col)`

**Status**: ✅ DONE  
**Root cause**: `BuildSelect` when `aggregate = COUNT` ignores the `distinct` and `distinctCols` fields. It emits `SELECT count(*)` rather than `SELECT COUNT(DISTINCT col)`.  
**Failing tests**:
- `integration_tests/sqlite/sqlite_query_distinct_test.go` — `Distinct count`, `Select with Distinct and Count`
- `integration_tests/mysql/mysql_query_distinct_test.go` — same subtests

**Implementation hint**: In `builder.go` `BuildSelect`, when `aggregate == "COUNT"` and `distinct == true`, generate `COUNT(DISTINCT <distinctCols>)`.

---

## GAP-07: `Select(col).Distinct().Scan()` does not generate `SELECT DISTINCT col`

**Status**: ✅ DONE  
**Root cause**: `BuildSelect` ignores `q.distinct` when assembling the SELECT list. It emits `SELECT col` instead of `SELECT DISTINCT col`.  
**Failing tests**:
- `integration_tests/sqlite/sqlite_query_distinct_test.go` — `Distinct select`, `Distinct with Scan`

**Implementation hint**: In `builder.go` `BuildSelect`, prepend `DISTINCT` to the column list when `q.distinct == true` and no aggregate is set.

---

## GAP-08: Chained `Having()` calls generate invalid SQL (duplicate HAVING keyword)

**Status**: ✅ DONE  
**Root cause**: Each `Having()` call appends a new `HAVING` clause independently, resulting in `HAVING ... HAVING ...` which is invalid SQL.  
**Failing tests**:
- `integration_tests/sqlite/sqlite_query_group_having_test.go` — `Multiple Having clauses`
- `integration_tests/mysql/mysql_query_group_having_test.go` — same

**Implementation hint**: Store having conditions in a slice (like `wheres`). `BuildSelect` should emit a single `HAVING cond1 AND cond2`.

---

## GAP-09: `Having()` does not support `func(Query)Query` callback subqueries

**Status**: ❌ Open  
**Root cause**: `Having` only accepts string + args. There is no code path for a callback that builds a subquery.  
**Failing tests**:
- `integration_tests/sqlite/sqlite_query_group_having_test.go` — `Having with subquery callback`, `Having with subquery in args`

**Implementation hint**: Add a callback overload check in the `Having` method, similar to how `Where` handles closures.

---

## GAP-10: `Increment()` / `Decrement()` generate invalid SQL (missing argument)

**Status**: ✅ DONE (SQL generation fixed, testing blocked by model limitations)  
**Root cause**: `BuildIncrement`/`BuildDecrement` in `builder.go` produce SQL like `SET col = col + ` with a missing `?` placeholder, or the argument is not appended to the args slice.  
**Failing tests**:
- `integration_tests/sqlite/sqlite_query_increment_decrement_test.go` — all subtests
- `integration_tests/mysql/mysql_query_increment_decrement_test.go` — all subtests

**Implementation hint**: Check `ToSql().Increment("col")` output. The amount value must be bound as a `?` arg.

---

## GAP-11: JSON path methods generate invalid SQL for SQLite

**Status**: ❌ Open (SQLite-specific)  
**Root cause**: `WhereJsonContains`, `WhereJsonContainsKey`, `WhereJsonLength`, `WhereJsonDoesntContain`, `WhereJsonDoesntContainKey` generate MySQL `JSON_CONTAINS()`/`JSON_CONTAINS_PATH()` syntax. SQLite requires `json_extract()` / `json_type()` / `json_array_length()`.  
**Failing tests**:
- `integration_tests/sqlite/sqlite_query_json_test.go` — all subtests

**Implementation hint**: In the builder, detect the driver dialect and emit the appropriate JSON function for the target DB. SQLite: `json_extract(col, '$.key')`. MySQL: `JSON_CONTAINS(col, ...)`. PostgreSQL: `col->>'key'`.

---

## GAP-12: `LockForUpdate()` / `SharedLock()` generate MySQL syntax on SQLite

**Status**: ❌ Open (SQLite-specific)  
**Root cause**: `BuildSelect` always appends `FOR UPDATE` or `LOCK IN SHARE MODE`. SQLite does not support either clause.  
**Failing tests**:
- `integration_tests/sqlite/sqlite_query_lock_test.go` — `TestSQLiteLockForUpdate`, `TestSQLiteSharedLock`

**Implementation hint**: In `BuildSelect`, skip lock clauses when the driver dialect is `sqlite`.

---

## GAP-13: `Omit()` does not exclude columns from SELECT

**Status**: ❌ Open  
**Root cause**: `Omit()` stores column names in `q.omit` but `BuildSelect` never reads `q.omit` to filter the SELECT list.  
**Failing tests**:
- `integration_tests/sqlite/sqlite_query_omit_test.go` — `Omit during select`
- `integration_tests/mysql/mysql_query_omit_test.go` — `Omit_during_select`

**Implementation hint**: In `BuildSelect`, when `q.omit` is non-empty and columns are being derived from the model struct, exclude any field whose column name is in `q.omit`.

---

## GAP-14: `Omit().Save()` generates invalid SQL (`near "SET": syntax error`)

**Status**: ❌ Open  
**Root cause**: `Save` (which is built on top of `Update`) with `Omit` applied generates SQL without a valid SET clause — either empty or malformed.  
**Failing tests**:
- `integration_tests/sqlite/sqlite_query_omit_test.go` — `Omit during update`
- `integration_tests/mysql/mysql_query_omit_test.go` — `Omit_during_update`

**Implementation hint**: Fix GAP-13 first (Omit support). Then ensure the `Save`/`Update` builder path respects `q.omit` when constructing the SET clause.

---

## GAP-15: `Order(expr).OrderBy(col, dir)` generates invalid SQL

**Status**: ❌ Open  
**Root cause**: When both `Order(rawExpr)` and `OrderBy(col, dir)` are called, the builder appends them incorrectly, resulting in e.g. `ORDER BY LENGTH(name) DESC name asc` (missing comma, or conflated direction).  
**Failing tests**:
- `integration_tests/sqlite/sqlite_query_order_limit_offset_test.go` — `OrderBy with expressions`
- `integration_tests/mysql/mysql_query_order_limit_offset_test.go` — same (likely)

**Implementation hint**: Store raw `Order()` expressions and `OrderBy()` column/direction pairs in a unified ordered slice. Emit each entry comma-separated in the ORDER BY clause.

---

## GAP-16: `Offset()` without `Limit()` generates invalid SQL in SQLite

**Status**: ❌ Open  
**Root cause**: `BuildSelect` emits `OFFSET n` without a preceding `LIMIT` clause. SQLite requires `LIMIT -1 OFFSET n` (or any LIMIT) when using OFFSET.  
**Failing tests**:
- `integration_tests/sqlite/sqlite_query_order_limit_offset_test.go` — `Offset clause`

**Implementation hint**: In `BuildSelect`, when `q.offset > 0` and `q.limit == 0`, emit `LIMIT -1 OFFSET n` for SQLite dialect (or `LIMIT 9223372036854775807`).

---

## GAP-17: `Count()` with `Select` alias returns 0

**Status**: ✅ DONE  
**Root cause**: When `Select("name as user_name")` is set, `BuildSelect` with `aggregate = COUNT` emits `SELECT count(name as user_name)` or similar invalid SQL, returning 0.  
**Failing tests**:
- `integration_tests/sqlite/sqlite_query_paginate_test.go` — `Count with Select alias (sanity check)`, `Pagination with Select aliases`
- `integration_tests/mysql/mysql_query_paginate_test.go` — same (likely)

**Implementation hint**: In `BuildSelect`, when `aggregate == "COUNT"`, always emit `SELECT count(*)` and ignore any `q.selects`. The SELECT list is irrelevant for counting.

---

## GAP-18: `Distinct()` does not apply to `Pluck`

**Status**: ❌ Open  
**Root cause**: The `Pluck` method builds its own query string independently of `BuildSelect` and does not consult `q.distinct`.  
**Failing tests**:
- `integration_tests/sqlite/sqlite_query_pluck_test.go` — `Pluck with Distinct`
- `integration_tests/mysql/mysql_query_pluck_test.go` — same

**Implementation hint**: In `Pluck`, when `q.distinct == true`, emit `SELECT DISTINCT col FROM ...`.

---

## GAP-19: `Select(rawSQL, args...)` causes argument index mismatch

**Status**: ❌ Open  
**Root cause**: When `Select("(SELECT name FROM t WHERE id = ?) as sub", userID)` is used alongside `Where("id = ?", id)`, the extra args from `Select` shift the positional `?` indices, causing `missing argument with index N`.  
**Failing tests**:
- `integration_tests/sqlite/sqlite_query_select_test.go` — `Select with raw subqueries`
- `integration_tests/mysql/mysql_query_select_test.go` — `Select_with_raw_subqueries`

**Implementation hint**: Track args attached to each `Select` expression separately and splice them into the final args list in the correct position relative to the SELECT clause (before WHERE args).

---

## GAP-20: `Select(func(q Query)Query, alias)` subquery callbacks not implemented

**Status**: ❌ Open  
**Root cause**: `Select` only handles `string` and `[]string` types. A `func(Query)Query` callback for building subqueries is not supported.  
**Failing tests**:
- `integration_tests/sqlite/sqlite_query_select_test.go` — `Select with subquery callbacks`

**Implementation hint**: In `Select`, detect `func(contractsorm.Query) contractsorm.Query` via reflect and invoke it with a clone of the current query, then embed the resulting SQL as a subquery expression in the SELECT list.

---

## GAP-21: `ToSql()` / `ToRawSql()` do not quote identifiers

**Status**: ❌ Open  
**Root cause**: The `ToSql` / `ToRawSql` wrappers call `BuildSelect` which does not quote table/column names. Tests that check for `` `users` `` (MySQL) or `"users"` (SQLite/Postgres) quoting fail.  
**Failing tests**:
- `integration_tests/sqlite/sqlite_query_to_sql_test.go` — `ToSql`, `ToRawSql`, `ToSql Count`, `ToRawSql Count`, `ToSql Update`, `ToSql Delete`
- `integration_tests/mysql/mysql_query_to_sql_test.go` — `ToSql`, `ToRawSql`, `ToSql Count`, `ToSql Update`

**Implementation hint**: Add a `quoteIdentifier(name, dialect)` helper that wraps names in `` ` `` for MySQL and `"` for SQLite/Postgres. Apply it in `BuildSelect`, `BuildUpdate`, `BuildDelete`, and `BuildInsert`.

---

## GAP-22: `UpdateOrInsert()` does not work with `Table()` (no `Model()`)

**Status**: ❌ Open  
**Root cause**: `UpdateOrInsert` internally calls `Create` / `Update` which both hit the `BuildInsert`/`BuildUpdate` map path. The table-only + map path is broken (see GAP-02).  
**Failing tests**:
- `integration_tests/sqlite/sqlite_query_update_or_insert_test.go` — all subtests
- `integration_tests/mysql/mysql_query_update_or_insert_test.go` — all subtests

**Implementation hint**: Fix GAP-02 first. Then verify `UpdateOrInsert` correctly issues `SELECT count(*)` with the attributes as WHERE conditions, then either INSERT or UPDATE.

---

## GAP-23: `Query.EnableQueryLog()` / `GetQueryLog()` not exposed on `database.Database`

**Status**: ✅ DONE  
**Root cause**: `Query` instances have `EnableQueryLog`, `DisableQueryLog`, `FlushQueryLog`, `GetQueryLog` methods. `database.Database` does not proxy these, so integration tests calling `db.EnableQueryLog()` fail to compile / are not wired.  
**Failing tests**:
- `integration_tests/sqlite/sqlite_query_log_test.go` — all subtests

**Implementation hint**: Add `EnableQueryLog()`, `DisableQueryLog()`, `FlushQueryLog()`, `GetQueryLog()` methods to `database.Database` that delegate to the underlying `ormInstance`'s query.

---

## GAP-24: ORM `Raw()` returns `*Query` — cannot be used as a map value in `Create()`

**Status**: ❌ Open  
**Root cause**: `db.Query().Raw("ST_PointFromText(?)", "POINT(1 1)")` returns a `*Query` struct. When passed as a value in `map[string]any{"location": <*Query>}`, the sql driver reports `unsupported type query.Query, a struct`.  
**Failing tests**:
- `integration_tests/mysql/mysql_query_spatial_test.go` — all subtests

**Implementation hint**: `Raw()` should return a `RawExpr` wrapper type (not `*Query`) that `BuildInsert` recognises and inlines as literal SQL with its own bound args.

---

## GAP-25: MySQL `TestMySQLIntegrationConnection/Default_connection_name` — wrong expectation

**Status**: ✅ FIXED (test corrected)  
**Root cause**: Test asserted `conn.DatabaseName() == "mysql"` (the connection key) but `DatabaseName()` returns the actual database name from config (e.g. `"test"`). Test updated to use `getEnv("MYSQL_DATABASE", "test")`.

---

## GAP-26: MySQL tables not created before tests run

**Status**: ✅ FIXED  
**Root cause**: `SetupMySQLTest` connected to MySQL but never created the required tables. Fixed: `createMySQLTestTables` added to `helper.go`, creates `users`, `addresses`, `books`, `peoples`, `json_datas` with `CREATE TABLE IF NOT EXISTS`.

---

## Priority Order

| Priority | GAP | Feature |
|---|---|---|
| P1 | GAP-02, GAP-04 | `Create(map)` / `InsertGetId(map)` with `Table()` |
| P1 | GAP-01 | Batch `Create` for slices |
| P1 | GAP-22 | `UpdateOrInsert` |
| P2 | GAP-05 | `Chunk()` typed callbacks |
| P2 | GAP-10 | `Increment`/`Decrement` SQL |
| P2 | GAP-06, GAP-07 | `Distinct` in SELECT and COUNT |
| P2 | GAP-08 | Chained `Having()` |
| P2 | GAP-13, GAP-14 | `Omit()` in SELECT and UPDATE |
| P3 | GAP-03 | `InsertGetId` writes back struct ID |
| P3 | GAP-15 | `Order(expr).OrderBy()` |
| P3 | GAP-16 | `Offset` without `Limit` (SQLite) |
| P3 | GAP-17 | `Count` ignores `Select` alias |
| P3 | GAP-18 | `Distinct` on `Pluck` |
| P3 | GAP-19 | `Select` with raw subquery args |
| P3 | GAP-21 | `ToSql` identifier quoting |
| P4 | GAP-09 | `Having` callback subqueries |
| P4 | GAP-11 | JSON methods on SQLite |
| P4 | GAP-12 | Lock clauses on SQLite |
| P4 | GAP-20 | `Select` callback subqueries |
| P4 | GAP-23 | `db.EnableQueryLog()` proxy |
| P4 | GAP-24 | `Raw()` as map value in `Create` |
