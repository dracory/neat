# Lessons from Database Package

This document outlines key lessons and best practices that Neat ORM can learn from the database package, based on a direct comparison of both codebases.

> Status markers: ✅ Already done in Neat | ⚠️ Partial | ❌ Not yet implemented

## Overview

The database package demonstrates excellent practices in building a focused, production-ready database connection library with automatic optimizations, sensible defaults, and clean API design. These lessons can help improve Neat ORM's database connection handling, performance, and developer experience.

## 1. Automatic Database Optimizations

### SQLite-Specific Optimizations ❌

**Database Package Approach:**
```go
if databaseType == DATABASE_TYPE_SQLITE {
    // Enable WAL mode for better concurrency
    _, _ = db.Exec("PRAGMA journal_mode=WAL;")
    // Use NORMAL synchronous for WAL (durable enough, faster writes)
    _, _ = db.Exec("PRAGMA synchronous=NORMAL;")
    // Ensure foreign keys are enforced
    _, _ = db.Exec("PRAGMA foreign_keys=ON;")
    // Back off up to 5s when the database is busy
    _, _ = db.Exec("PRAGMA busy_timeout=5000;")
}
```

**Current Neat state (`database/db.go`):** No SQLite-specific PRAGMA configuration is applied after opening a connection. The `New()` and `NewFromDSN()` constructors delegate immediately to `databaseorm.BuildOrm()` without any post-open optimization hooks.

**Lesson for Neat:**
- Add a post-open hook in `BuildOrm` (or in `New`/`NewFromDSN`) to apply SQLite PRAGMAs when `driver == "sqlite"`
- Enable WAL mode for better concurrency under concurrent readers
- Set `busy_timeout=5000` to avoid immediate SQLITE_BUSY errors
- Enable `foreign_keys=ON` since SQLite disables them by default
- Ignore PRAGMA errors (they are optimizations, not requirements): `_, _ = db.Exec(...)`

### Connection Pool Defaults ❌

**Database Package Approach:**
```go
// SQLite: Constrain pool to avoid multiple concurrent writers
maxOpen := 1
maxIdle := 1

// MySQL/Postgres: Sensible production defaults
maxOpen := 25
maxIdle := 5
```

**Current Neat state (`database/db.go`, `NewFromDSN`):**
```go
poolConfig := db.PoolConfig{
    MaxIdleConns:    10,
    MaxOpenConns:    100,  // too high — no per-driver tuning
    ConnMaxLifetime: int(time.Hour.Seconds()),
    ConnMaxIdleTime: int(time.Hour.Seconds()),
}
```

MaxOpen=100 is applied uniformly to all drivers, including SQLite, which does not support concurrent writers. This will cause write contention and `database is locked` errors under load.

**Lesson for Neat:**
- Apply driver-aware defaults when no explicit pool config is given
- SQLite: MaxOpen=1, MaxIdle=1 — single writer, WAL allows concurrent readers
- MySQL/Postgres: MaxOpen=25, MaxIdle=5 for production workloads
- The `PoolConfig` struct in `database/db/config_builder.go` already supports this — it just needs driver-aware initialization logic
- Document rationale for each default in `PoolConfig` struct comments

### Query Timeout Default ⚠️

**Current Neat state (`database/db/config_builder.go`):**
```go
case "database.pool.query_timeout":
    if c.Pool.QueryTimeout > 0 {
        return c.Pool.QueryTimeout
    }
    return 30 // Default 30 seconds
```

Neat already has a `QueryTimeout` field in `PoolConfig` — the database package does not. However, Neat's `NewFromDSN` does not populate this field in the default `poolConfig`, so it defaults to zero and falls back to the 30s default only via the config accessor.

**Lesson for Neat:**
- Explicitly set `QueryTimeout: 30` in the `NewFromDSN` default pool config for clarity
- Document that this timeout applies to individual query execution, not to the connection itself
- This is an advantage Neat has over the database package — it should be surfaced more clearly in docs

## 2. Nil Database Connection Handling ❌

**Database Package Approach:**
```go
func (o *openOptions) Verify() error {
    if !o.HasDatabaseType() {
        return errors.New(`database type is required`)
    }
    // ... validates all required fields before opening
}
```

**Current Neat state (`database/query/query_errors_test.go`):**
```go
// Recover from panic since nil DB causes panic in database/sql package
defer func() {
    if r := recover(); r != nil {
        // Expected panic for nil database connection
    }
}()
err := q.First(&result)
```

Six tests use `defer recover()` to catch panics from `database/sql` when `db` is nil. This is not graceful degradation — it is a known panic that is being masked in tests. The `NewQuery` function in `database/query/` does not guard against a nil `*sql.DB`.

**Lesson for Neat:**
- `NewQuery` should validate that `db` is non-nil and return an error-state query or a sentinel error
- The query's terminal methods (`First`, `Find`, `Create`, `Update`, `Delete`) should check for nil DB and return `ErrNilDatabase` instead of panicking
- Remove `defer recover()` blocks from tests once this is fixed — they hide real bugs

## 3. Simplicity and Focus

### Single Responsibility ⚠️

**Database Package Approach:**
- Focused solely on database connection and basic operations
- No ORM, no query builder, no schema builder
- 16 Go files with clear responsibilities
- Minimal dependencies

**Current Neat state:** Neat intentionally combines ORM, query builder, schema builder, migration, seeding, and associations in one module. This is by design and not inherently wrong. However, the internal packages (`database/query/`, `database/schema/`, `database/migration/`) are already well-separated — the issue is surface-area creep in `database/db.go` which exposes all of them through one `Database` facade.

**Lesson for Neat:**
- The `Database` facade (`database/db.go`) is 496 lines. Consider whether all methods belong on it or if some should be accessed via sub-objects only (e.g. `d.Schema()`, `d.Migrate()` already do this correctly)
- Keep the `Database` struct as an entry point, not a god object
- Avoid adding new top-level methods to `Database` that could live on a sub-object

### Clean API Design ✅

**Database Package Approach:**
```go
db, err := database.Open(database.Options().
    SetDatabaseType(DbDriver).
    SetDatabaseHost(DbHost).
    SetDatabasePort(DbPort).
    SetDatabaseName(DbName).
    SetCharset(`utf8mb4`).
    SetUserName(DbUser).
    SetPassword(DbPass))
```

**Current Neat state:** Neat provides two clean constructors: `New(cfg db.DBConfig, opts ...Option)` and `NewFromDSN(dsn string, opts ...Option)`. The `Option` functional pattern matches this philosophy and is well-implemented. The `NewFromDSN` approach is even simpler for common cases.

**No action required.** This is already well-designed.

## 4. Driver Management

### Driver Explicitness ✅

**Database Package Approach:**
- Drivers not included to prevent size bloat
- Users import only required drivers
- Clear documentation of required imports
- Supports SQLite, MySQL, PostgreSQL, PGX

**Current Neat state:** Neat supports SQLite, MySQL, PostgreSQL, Oracle, Turso, and SQL Server via the `database/driver/` package. Drivers are registered through the `driver` sub-package rather than being bundled unconditionally. This already avoids driver bloat.

**Lesson for Neat:**
- Document the required driver import for each supported database in the README and `database/driver/` package godoc
- Consider adding build tags to exclude drivers at compile time if binary size becomes a concern

## 5. Error Handling

### Validation Before Connection ⚠️

**Database Package Approach:**
```go
func (o *openOptions) Verify() error {
    if !o.HasDatabaseType() {
        return errors.New(`database type is required`)
    }
    // ... comprehensive validation
    return nil
}
```

**Current Neat state (`database/db.go`, `parseDSN`):** DSN validation checks for empty string and max length but does not validate driver-specific requirements (e.g. missing host for MySQL, missing database name for PostgreSQL). The `DBConfig` struct has no `Verify()` method.

**Lesson for Neat:**
- Add a `Verify()` or `Validate()` method to `DBConfig` that checks required fields per driver
- `parseDSN` already handles some edge cases — consolidate validation logic there
- Provide clear error messages that identify which field is missing and why

### Graceful Degradation for Optimizations ❌

**Database Package Approach:**
```go
// PRAGMA errors ignored to prevent failures
_, _ = db.Exec("PRAGMA journal_mode=WAL;")
```

**Lesson for Neat:**
- When SQLite PRAGMA calls are added (see section 1), always use `_, _ = db.Exec(...)` — never fail the connection over optimization PRAGMAs
- Optionally log warnings through the injected `log.Log` interface when a PRAGMA fails
- This distinction (critical error vs. optimization failure) should be documented in the SQLite driver

## 6. Database-Specific Handling

### Dialect-Aware Configuration ✅

**Database Package Approach:**
```go
func dsn(driver string, ...) string {
    if strings.EqualFold(driver, DATABASE_TYPE_SQLITE) { return databaseName }
    if strings.EqualFold(driver, DATABASE_TYPE_MYSQL) { /* MySQL DSN */ }
    if strings.EqualFold(driver, DATABASE_TYPE_POSTGRES) { /* PostgreSQL DSN */ }
    return ""
}
```

**Current Neat state:** `database/db/config_builder.go` `ConfigBuilder.BuildDSN()` already handles this with a clean `switch b.config.Driver` statement covering mysql, postgres, sqlite, sqlserver, turso, and oracle. This is more complete than the database package.

**No action required.** Neat's implementation is ahead here.

### Feature Detection ⚠️

**Database Package Approach:**
```go
// SSL mode only for PostgreSQL, defaults to "disable"
// Charset only for MySQL, defaults to "utf8mb4"
```

**Current Neat state (`database/db.go`, `parseDSN`):**
```go
case "postgres":
    config.SSLMode = query.Get("sslmode")
    if config.SSLMode == "" {
        config.SSLMode = "require"  // defaults to require, not disable
    }
case "mysql":
    config.Charset = query.Get("charset")  // no default charset set here
```

**Differences to note:**
- Neat defaults PostgreSQL SSLMode to `"require"` (secure by default), the database package defaults to `"disable"` — Neat's approach is better
- MySQL charset has no default in `parseDSN`; `buildMySQLDSN()` only adds it if non-empty. Consider defaulting to `utf8mb4`
- Document these defaults clearly so users are not surprised

## 7. Documentation Quality

### Practical Examples ⚠️

**Database Package Approach:** Comprehensive inline godoc examples on every public method.

**Current Neat state:** Most methods in `database/db.go` have godoc comments, but they are minimal descriptions without usage examples. The `NewFromDSN` method has excellent inline format documentation (DSN examples as comments), but `New()` has none.

**Lesson for Neat:**
- Add `Example*` functions to test files for the primary constructors (`New`, `NewFromDSN`)
- Expand `NewFromDSN` godoc to show supported DSN formats (already partly done)
- Add godoc examples for `Transaction()`, `Query()`, and `Schema()` methods

### Transaction Documentation ⚠️

**Current Neat state:** `Transaction(txFunc func(tx orm.Query) error, opts ...*sql.TxOptions)` has a one-line comment. The pattern for nested transactions and savepoints is undocumented.

**Lesson for Neat:**
- Document the `Transaction` callback pattern with a full example showing error handling and rollback behavior
- Document savepoint support (`SavePoint`, `RollbackTo`) which exists in `database/query/` but is not surfaced on the `Database` facade
- Document transaction isolation level usage via `*sql.TxOptions`
- Note that Neat uses a callback pattern (`func(tx orm.Query) error`) rather than explicit Begin/Commit — document why this is safer

## 8. Testing Strategy

### Focused Test Coverage ⚠️

**Database Package Approach:**
- Tests for database-specific optimizations (PRAGMA verification)
- Validation tests for configuration before connection
- Error handling tests for all failure modes

**Current Neat state:** `database/query/query_errors_test.go` has 35+ error scenario tests — this is excellent coverage. However, six tests use `defer recover()` to absorb panics from nil `*sql.DB`, which masks real defects rather than testing correct behavior. There are no tests for pool configuration logic or driver-specific defaults.

**Lesson for Neat:**
- Fix the six nil-DB panic tests (see Section 2) so they test proper error returns instead
- Add tests for `ConfigBuilder.BuildDSN()` covering all six drivers
- Add tests verifying that `NewFromDSN("sqlite://...")` sets appropriate pool limits once driver-aware defaults are implemented
- Add a test that `parseDSN` with empty DSN returns a descriptive error

## Implementation Roadmap

### Priority 1: Correctness Fixes (High Impact, Low Risk)
1. Fix nil `*sql.DB` in `NewQuery` — return `ErrNilDatabase` instead of panicking (eliminates 6 `defer recover()` blocks in tests)
2. Fix `NewFromDSN` SQLite pool default — set MaxOpen=1, MaxIdle=1 when `driver == "sqlite"`
3. Explicitly set `QueryTimeout: 30` in `NewFromDSN` default `poolConfig`

### Priority 2: SQLite Optimizations (Medium Impact, Low Risk)
1. Add post-open hook in `BuildOrm` or `New` to apply SQLite PRAGMAs
2. Apply: `journal_mode=WAL`, `synchronous=NORMAL`, `foreign_keys=ON`, `busy_timeout=5000`
3. Ignore PRAGMA errors with `_, _ = db.Exec(...)`
4. Add tests verifying PRAGMAs are set on SQLite connections

### Priority 3: Validation & Documentation (Medium Impact, Medium Effort)
1. Add `Validate()` method to `DBConfig` with driver-aware field checking
2. Default MySQL charset to `utf8mb4` in `buildMySQLDSN` when empty
3. Add godoc `Example*` functions for `New`, `NewFromDSN`, `Transaction`
4. Document savepoint support and transaction isolation options

### Priority 4: Documentation Completeness
1. Document all supported DSN formats with examples
2. Create a driver compatibility matrix (supported features per driver)
3. Document pool configuration rationale and recommendations

## What Neat Already Does Better

| Feature | Database Package | Neat |
|---------|-----------------|------|
| DSN parsing | Limited | Full URL parsing with 6 drivers |
| PostgreSQL SSL default | `disable` | `require` (more secure) |
| Query timeout | Not supported | `PoolConfig.QueryTimeout` |
| Driver support | 4 drivers | 6 drivers (+ Oracle, Turso) |
| Functional options | Not used | `WithContext`, `WithPool`, etc. |
| Transaction API | Manual Begin/Commit | Callback pattern (safer) |
| Schema builder | Not included | Full schema builder included |

## Success Metrics

- **Nil safety**: Zero `defer recover()` blocks in query tests
- **SQLite safety**: No `database is locked` errors under concurrent load
- **Pool defaults**: Driver-aware defaults applied without explicit configuration
- **Validation**: `DBConfig.Validate()` catches all configuration errors before connection
- **Documentation**: `Example*` tests for all primary entry points

## Conclusion

The database package demonstrates how to build a focused, production-ready database connection library with automatic optimizations and sensible defaults. The most actionable lessons for Neat are the nil-DB panic fix, SQLite-specific connection pool limits, and automatic PRAGMA configuration — all of which can be implemented without any breaking API changes.

Neat already exceeds the database package in several areas (DSN flexibility, driver support, query timeout, transaction safety). The gap is primarily in default safety for SQLite connections and runtime nil-safety in the query layer.
