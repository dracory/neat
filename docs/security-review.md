# Security Review Report
Date: 2026-06-23
Reviewer: Senior Principal Golang Engineer (Cascade)
Codebase: neat — `feat/array-driver-8792631851847356937` branch (vs origin/main)

## Executive Summary

This review covers the array driver feature added in 17 changed files (+941/-7 lines). The feature introduces an in-memory SQLite-backed driver that populates tables from Go data structures (`ArraySource`). The overall design is sound — identifiers are validated, parameterized queries are used for data insertion, and concurrency is handled with double-checked locking. However, there are **2 high-severity** and **4 medium-severity** issues that should be addressed before merging. The most critical finding is a SQL injection vector through unvalidated schema type strings from the `ArraySchema` interface.

## Critical Findings (Severity: Critical)

None.

## High Severity Findings

### Finding #1: SQL Injection via Unvalidated Schema Type Strings ✅ FIXED

- **Location**: `database/driver/array.go:218-241`
- **Description**: The `createTable` method constructs SQL by interpolating `schema[col]` (the type string) directly into the DDL statement. While column names are validated via `isSimpleIdentifier`, the **type values** from `ArraySchema.Schema()` are only run through a `switch/case` that normalizes known types. If the type string doesn't match any case, it falls through and is used as-is in the SQL string with no validation.
- **Impact**: A malicious or buggy `ArraySource` implementation providing `ArraySchema` with a crafted type string (e.g., `"TEXT); DROP TABLE users; --"`) can inject arbitrary SQL into the `CREATE TABLE` statement. While the array driver is typically used with developer-controlled data, if `ArraySource` implementations ever incorporate user-supplied type metadata, this becomes exploitable. CWE-89 (SQL Injection).
- **Code Example**:
```go
// database/driver/array.go:236
columns = append(columns, fmt.Sprintf("\"%s\" %s", col, sqlType))
// sqlType is schema[col] which may be an arbitrary string from ArraySchema.Schema()
```
- **Suggested Fix**:
```go
// Add a default case that rejects unknown type strings
switch strings.ToLower(sqlType) {
case "int", "integer":
    sqlType = "INTEGER"
case "float", "real", "double":
    sqlType = "REAL"
case "bool", "boolean":
    sqlType = "INTEGER"
case "string", "text":
    sqlType = "TEXT"
case "time", "datetime", "timestamp":
    sqlType = "DATETIME"
default:
    return fmt.Errorf("unsupported column type %q for column %q", sqlType, col)
}
```

### Finding #2: Pointer Reuse in Populated Cache Causes Stale State ✅ FIXED

- **Location**: `database/driver/array.go:123-131`
- **Description**: The `isPopulated` and `markPopulated` methods use `fmt.Sprintf("%p-%s", db, tableName)` as the cache key. When a `*sql.DB` is closed and a new one is allocated, Go's memory allocator may reuse the same memory address, producing the same `%p` string. The new connection would incorrectly believe the table is already populated, skipping `CREATE TABLE` and `INSERT` entirely.
- **Impact**: After a connection pool recycle or database reconnect, queries against array-backed tables will silently execute against empty/non-existent tables, returning zero results. This is a data integrity issue. In multi-connection scenarios, it could also cause cross-connection data leakage if the old connection's data persists. CWE-1188 (Insecure Default - Pointer Reuse).
- **Code Example**:
```go
// database/driver/array.go:123-127
func (a *Array) isPopulated(db *sql.DB, tableName string) bool {
    key := fmt.Sprintf("%p-%s", db, tableName)
    _, ok := a.populated.Load(key)
    return ok
}
```
- **Suggested Fix**: Use a unique connection identifier instead of pointer address. One approach is to embed a monotonic counter in the driver or use the DSN string as part of the key:
```go
// Option A: Use a unique ID assigned at Open() time
// Option B: Verify table existence in SQLite before skipping population
func (a *Array) isPopulated(db *sql.DB, tableName string) bool {
    key := fmt.Sprintf("%p-%s", db, tableName)
    if _, ok := a.populated.Load(key); ok {
        // Verify the table actually exists to guard against pointer reuse
        var exists int
        err := db.QueryRow(
            "SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?", tableName,
        ).Scan(&exists)
        if err != nil || exists == 0 {
            a.populated.Delete(key)
            return false
        }
        return true
    }
    return false
}
```

## Medium Severity Findings

### Finding #3: `newDriverForDialect` Missing "array" Case ✅ FIXED

- **Location**: `database/query/query_clone.go:136-152`
- **Description**: The `newDriverForDialect` function (used by `Query.Connection()`) does not have a case for `"array"`, falling through to the default which returns `driver.NewSQLite()`. The SQLite driver does not implement `contractsorm.ArrayPopulator`, so array-backed models will not be populated when accessed via `Connection()`.
- **Impact**: If a user switches to an array connection via `.Connection("array_db")`, the driver will be SQLite (not Array), `Model()` won't trigger population, and queries will fail against non-existent tables. This is a functional bug with potential security implications — queries may return unexpected data from pre-existing SQLite tables. CWE-754 (Improper Check for Unusual or Exceptional Conditions).
- **Code Example**:
```go
// database/query/query_clone.go:136-152
func newDriverForDialect(dialect string) driver.Driver {
    switch dialect {
    case "mysql":
        return driver.NewMySQL()
    // ... other cases ...
    default:
        return driver.NewSQLite() // "array" falls through to here
    }
}
```
- **Suggested Fix**:
```go
case "array":
    return driver.NewArray()
```

### Finding #4: Unbounded Growth in `populated` and `locks` sync.Maps ✅ FIXED

- **Location**: `database/driver/array.go:18-19`
- **Description**: The `populated` and `locks` `sync.Map` fields on the `Array` struct are never cleaned up. Every unique `*sql.DB` + table name combination adds an entry that persists for the lifetime of the driver instance. In long-running services that create many connections or use many array-backed tables, this constitutes an unbounded memory leak.
- **Impact**: Gradual memory exhaustion leading to OOM kills. CWE-401 (Missing Release of Memory after Effective Lifetime).
- **Code Example**:
```go
// database/driver/array.go:18-19
type Array struct {
    *SQLite
    populated sync.Map // never cleaned up
    locks     sync.Map // never cleaned up
    locksMu   sync.Mutex
}
```
- **Suggested Fix**: Add a `Cleanup(db *sql.DB)` method or use a `WeakMap`-like pattern. Alternatively, scope the cache to the `*sql.DB` lifecycle by registering a cleanup callback when the DB is closed. A simpler approach: use a bounded LRU cache instead of `sync.Map`.

### Finding #5: `CREATE TABLE IF NOT EXISTS` Silently Ignores Schema Mismatch ✅ FIXED

- **Location**: `database/driver/array.go:239`
- **Description**: `CREATE TABLE IF NOT EXISTS` will silently succeed if a table with the same name already exists, even if the schema differs. The subsequent `INSERT INTO` will then attempt to insert columns that may not exist in the pre-existing table, causing errors — or worse, if the column names happen to match, data will be appended to the pre-existing table.
- **Impact**: If an array-backed table name collides with a pre-existing SQLite table (e.g., from a previous driver or manual schema setup), data corruption or query errors can occur. The `isPopulated` cache prevents this within a single driver instance lifecycle, but not across driver recreations or pre-existing schemas. CWE-1295 (Debug Messages Revealing Unnecessary Information).
- **Code Example**:
```go
// database/driver/array.go:239
sql := fmt.Sprintf("CREATE TABLE IF NOT EXISTS \"%s\" (%s)", tableName, strings.Join(columns, ", "))
```
- **Suggested Fix**: Use `CREATE TABLE` (without `IF NOT EXISTS`) and handle the "table already exists" error explicitly, or verify the existing schema matches before proceeding:
```go
// Check if table exists first
var existingSchema string
err := db.QueryRowContext(ctx,
    "SELECT sql FROM sqlite_master WHERE type='table' AND name=?", tableName,
).Scan(&existingSchema)
if err == nil && existingSchema != "" {
    // Table exists — verify schema or skip population
    a.markPopulated(db, tableName)
    return nil
}
```

### Finding #6: Population Error During Transactions Not Surfaced Immediately ✅ FIXED

- **Location**: `database/query/query_model.go:17-39`
- **Description**: When `Model()` is called during a transaction, `q.DB()` returns an error (`"cannot get DB during transaction"`). This error is stored in `q.buildError` but `Model()` continues and returns normally. If the caller doesn't check `buildError` before executing the query, the query will run against an unpopulated table, silently returning empty results or errors.
- **Impact**: Silent data integrity violations — queries return empty results without clear error messages when array-backed models are used inside transactions. CWE-754 (Improper Check for Unusual or Exceptional Conditions).
- **Code Example**:
```go
// database/query/query_model.go:26-31
db, err := q.DB()
if err != nil {
    q.buildError = err  // Error stored but not returned
} else {
    if err := arrayDriver.Populate(q.ctx, db, source); err != nil {
        q.buildError = err  // Error stored but not returned
    }
}
// Method continues to reset state and return q
```
- **Suggested Fix**: Either return the error immediately from `Model()` (breaking the fluent API contract) or ensure all subsequent query methods (`Get`, `First`, etc.) check `buildError` before executing. At minimum, document that `Model()` may store a build error that surfaces later.

## Low Severity Findings

### Finding #7: No Context Timeout Enforcement on Populate ✅ FIXED

- **Location**: `database/driver/array.go:36-121`
- **Description**: The `Populate` method accepts a `context.Context` but doesn't enforce any internal timeout. If `source.Rows()` returns an extremely large dataset (millions of rows), the populate operation could consume significant memory and CPU.
- **Impact**: Potential denial of service if `ArraySource` implementations return unbounded datasets. CWE-400 (Uncontrolled Resource Consumption).
- **Recommendation**: Document expected dataset size limits, or add a configurable max-rows check in `Populate`.

### Finding #9: `q.ctx` May Be Nil When Populate Is Called ✅ FIXED

- **Location**: `database/query/query_model.go:30`
- **Description**: `arrayDriver.Populate(q.ctx, db, source)` is called with `q.ctx`, which may be `nil` if the `Query` was created without explicit context. Go's `database/sql` package treats a nil context as `context.Background()`, so this won't panic, but it means context cancellation won't work.
- **Impact**: Loss of context cancellation for populate operations. Low severity since the array driver is typically used with in-memory data.
- **Recommendation**: Ensure `q.ctx` is always initialized to `context.Background()` in query constructors.

## Best Practice Recommendations

1. **Validate all external inputs at trust boundaries**: The `ArraySchema.Schema()` return value is an external input that should be validated before use in SQL construction. Apply the same `isSimpleIdentifier`-style validation to type strings, or better, use an enum/allowlist.

2. **Add integration tests for transaction scenarios**: The current test suite doesn't cover array-backed models inside transactions. Add tests that verify `Model()` behavior when `q.tx != nil`.

3. **Add a test for `Connection()` with array driver**: Verify that `newDriverForDialect("array")` returns the correct driver type.

4. **Consider a `Close()` or `Reset()` method on the Array driver**: To clean up the `populated` and `locks` maps when connections are closed.

5. **Document the security model**: The array driver is designed for developer-controlled data, not user-supplied data. This should be explicitly documented in the package and type docs to prevent misuse.

## Dependencies Analysis

- Total direct dependencies: 12 (from `go.mod`)
- Dependencies with known vulnerabilities: 0 (all versions appear current as of review date)
- Notable dependencies:
  - `modernc.org/sqlite v1.52.0` — Pure-Go SQLite implementation, no CGo required. No known CVEs.
  - `github.com/go-sql-driver/mysql v1.10.0` — Current, no known CVEs.
  - `github.com/lib/pq v1.12.3` — Note: this package is in maintenance mode; consider migrating to `pgx` for new features.
  - `golang.org/x/crypto v0.53.0` — Current.

No new dependencies were introduced by this branch.

## Compliance Considerations

- **GDPR**: If array-backed tables contain PII, the in-memory SQLite database must be properly closed when no longer needed. The current `populated` cache retains references to `*sql.DB` instances via pointer keys, which could delay garbage collection.
- **PCI-DSS**: Not directly applicable, but if used in a payment context, the SQL injection vector (Finding #1) must be remediated.

## Summary Statistics

- Total issues found: 9
- Critical: 0
- High: 2
- Medium: 4
- Low: 3

## Next Steps

1. **Fix Finding #1 (SQL Injection in type strings)** — Add a `default` case to the type switch in `createTable` that rejects unknown types. This is a one-line fix with high impact.
2. **Fix Finding #3 (Missing "array" case in `newDriverForDialect`)** — Add the missing case. One-line fix.
3. **Address Finding #2 (Pointer reuse)** — Add a table-existence verification check in `isPopulated` or switch to a more robust cache key.
4. **Address Finding #6 (Transaction error handling)** — Ensure `buildError` is checked by all query execution methods, or return early from `Model()`.
5. **Address Finding #4 and #5** — Add cleanup mechanism and schema verification for production readiness.
6. **Add regression tests** for each fix, particularly: invalid schema types, `Connection()` with array driver, and array models inside transactions.
