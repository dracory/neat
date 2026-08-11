# Proposal: CockroachDB Integration Tests

**Date**: August 8, 2026
**Status**: Completed
**Priority**: Medium

## Problem

Neat ORM claims CockroachDB support via the PostgreSQL driver (documented in `docs/driver-registration.html`), but there are **no integration tests** verifying this. The comparison tables list CockroachDB as `✅` for Neat, yet we have zero test coverage proving that Neat's query builder, schema builder, migrations, associations, and soft deletes actually work against a real CockroachDB instance.

If a user picks Neat because of the CockroachDB claim and something breaks, we won't know until they report it.

## Goal

Add a full integration test suite for CockroachDB that mirrors the existing PostgreSQL integration tests, running against a real CockroachDB container in Docker Compose and GitHub Actions CI.

## Background

CockroachDB is a distributed SQL database with **high PostgreSQL compatibility**. It speaks the PostgreSQL wire protocol and accepts the `lib/pq` or `pgx` driver. Neat's PostgreSQL driver should work with CockroachDB out of the box, but there are edge cases:

- CockroachDB does not support `SAVEPOINT` in the same way PostgreSQL does (it uses `SAVEPOINT` but with different semantics around nested transactions)
- CockroachDB's `SERIAL` type behaves differently from PostgreSQL's `SERIAL` (CockroachDB uses `unique_rowid()` by default)
- CockroachDB may have different behavior for `RETURNING` clauses in some edge cases
- CockroachDB's JSON support (`JSONB`) is compatible but may have minor function differences
- CockroachDB's handling of `SELECT ... FOR UPDATE` differs in distributed mode
- CockroachDB does not support all PostgreSQL index types (e.g., partial indexes have different syntax)
- CockroachDB's `ALTER TABLE ... RENAME COLUMN` may have different behavior

These differences are exactly why we need integration tests.

## Design

### 1. Docker Compose Service

Add a CockroachDB service to `docker-compose.yml`:

```yaml
  cockroachdb:
    image: cockroachdb/cockroach:v24.3.0
    container_name: neat-cockroachdb-test
    command: start-single-node --insecure
    ports:
      - "26257:26257"
      - "8080:8080"
    healthcheck:
      test: ["CMD", "cockroach", "sql", "--insecure", "--host", "127.0.0.1:26257", "-e", "SELECT 1"]
      interval: 10s
      timeout: 5s
      retries: 15
```

CockroachDB listens on port `26257` for SQL connections (PostgreSQL wire protocol) and port `8080` for the admin UI. The `--insecure` flag is used for testing only (no TLS, no authentication).

### 2. Integration Test Directory

Create `integration_tests/cockroachdb/` with the same structure as `integration_tests/postgres/`:

```
integration_tests/cockroachdb/
├── helper.go                              # Connection config, setup, table creation
├── cockroachdb_connection_test.go         # Basic connection test
├── cockroachdb_find_test.go               # Find/First/Get queries
├── cockroachdb_query_create_test.go       # INSERT operations
├── cockroachdb_query_update_test.go       # UPDATE operations
├── cockroachdb_query_delete_test.go       # DELETE operations
├── cockroachdb_query_count_test.go        # COUNT and aggregates
├── cockroachdb_query_where_test.go        # WHERE clauses
├── cockroachdb_query_join_test.go         # JOIN operations
├── cockroachdb_query_order_limit_offset_test.go  # ORDER BY, LIMIT, OFFSET
├── cockroachdb_query_group_having_test.go # GROUP BY, HAVING
├── cockroachdb_query_association_test.go  # Associations (BelongsTo, HasMany, HasOne)
├── cockroachdb_query_belongs_to_test.go   # BelongsTo-specific tests
├── cockroachdb_query_chunk_test.go        # Chunk processing
├── cockroachdb_query_paginate_test.go     # Pagination
├── cockroachdb_query_pluck_test.go        # Pluck
├── cockroachdb_query_value_test.go        # Value extraction
├── cockroachdb_query_distinct_test.go     # DISTINCT
├── cockroachdb_query_select_test.go       # SELECT specific columns
├── cockroachdb_query_omit_test.go         # Omit columns
├── cockroachdb_query_load_test.go         # Eager loading
├── cockroachdb_query_lock_test.go         # Pessimistic locking
├── cockroachdb_query_scopes_test.go       # Query scopes
├── cockroachdb_query_to_sql_test.go       # ToSql interface
├── cockroachdb_query_increment_decrement_test.go  # Increment/Decrement
├── cockroachdb_query_update_or_insert_test.go     # UpdateOrInsert
├── cockroachdb_query_json_test.go         # JSON column operations
├── cockroachdb_query_aggregate_test.go    # SUM, AVG, MIN, MAX
├── cockroachdb_raw_test.go                # Raw SQL queries
├── cockroachdb_transaction_test.go        # Transactions and savepoints
├── cockroachdb_soft_delete_test.go        # Soft deletes
├── cockroachdb_schema_table_test.go       # Schema builder: table operations
├── cockroachdb_schema_column_types_test.go    # Schema builder: column types
├── cockroachdb_schema_column_methods_test.go  # Schema builder: column methods
├── cockroachdb_schema_column_modifiers_test.go # Schema builder: column modifiers
├── cockroachdb_schema_column_change_test.go    # Schema builder: column changes
├── cockroachdb_schema_foreign_key_test.go      # Schema builder: foreign keys
├── cockroachdb_schema_index_test.go            # Schema builder: indexes
├── cockroachdb_schema_rename_column_test.go    # Schema builder: rename column
├── cockroachdb_schema_timestamp_test.go        # Schema builder: timestamps
├── cockroachdb_schema_view_test.go             # View management
└── cockroachdb_dotted_column_test.go           # Dotted column references
```

### 3. Helper File

`integration_tests/cockroachdb/helper.go`:

```go
package cockroachdb_test

import (
    "fmt"
    "testing"
    "time"

    "github.com/dracory/neat"
    "github.com/dracory/neat/contracts/log"
    "github.com/dracory/neat/database"
    "github.com/dracory/neat/integration_tests/common"
    _ "github.com/lib/pq"
)

// GetCockroachDBConfig returns a CockroachDB connection config from environment variables.
// CockroachDB speaks the PostgreSQL wire protocol, so we use the PostgreSQL driver.
func GetCockroachDBConfig() neat.DBConfig {
    host := common.GetEnv("COCKROACHDB_HOST", "127.0.0.1")
    port := common.GetEnvInt("COCKROACHDB_PORT", 26257)
    database := common.GetEnv("COCKROACHDB_DATABASE", "test")
    username := common.GetEnv("COCKROACHDB_USER", "root")
    password := common.GetEnv("COCKROACHDB_PASS", "")

    return neat.DBConfig{
        Default: "cockroachdb",
        Connections: map[string]neat.ConnectionConfig{
            "cockroachdb": {
                Driver:   "postgres", // CockroachDB uses the PostgreSQL driver
                Host:     host,
                Port:     port,
                Database: database,
                Username: username,
                Password: password,
                SSLMode:  "disable",
            },
        },
        Pool: neat.PoolConfig{
            MaxIdleConns:    5,
            MaxOpenConns:    10,
            ConnMaxLifetime: time.Hour,
            ConnMaxIdleTime: time.Hour,
        },
    }
}

// SetupCockroachDBTest sets up a CockroachDB connection and creates test tables.
func SetupCockroachDBTest(t *testing.T) *database.DB {
    t.Helper()
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    config := GetCockroachDBConfig()
    db, err := neat.New(config)
    if err != nil {
        t.Fatalf("Failed to connect to CockroachDB: %v", err)
    }

    // Create test database if it doesn't exist
    // CockroachDB defaults to no password, root user
    err = createCockroachDBDatabase(db)
    if err != nil {
        t.Fatalf("Failed to create CockroachDB test database: %v", err)
    }

    // Create tables
    createCockroachDBTables(t, db)

    return db
}

// SetupCockroachDBConnection sets up a CockroachDB connection without creating tables.
func SetupCockroachDBConnection(t *testing.T) *database.DB {
    t.Helper()
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    config := GetCockroachDBConfig()
    db, err := neat.New(config)
    if err != nil {
        t.Fatalf("Failed to connect to CockroachDB: %v", err)
    }

    return db
}
```

The table creation SQL should be nearly identical to the PostgreSQL helper's SQL, since CockroachDB is PostgreSQL-compatible. Any divergences (e.g., `SERIAL` vs `unique_rowid()`, index syntax) should be documented in the helper file.

### 4. Environment Variables

Add CockroachDB environment variables to the CI workflow:

```
COCKROACHDB_HOST: 127.0.0.1
COCKROACHDB_PORT: 26257
COCKROACHDB_DATABASE: test
COCKROACHDB_USER: root
COCKROACHDB_PASS: ""
```

### 5. GitHub Actions CI

Add a CockroachDB service container to `.github/workflows/tests.yml`:

```yaml
      cockroachdb:
        image: cockroachdb/cockroach:v24.3.0
        ports:
          - "26257:26257"
        options: >-
          --health-cmd="cockroach sql --insecure --host 127.0.0.1:26257 -e 'SELECT 1'"
          --health-interval=10s
          --health-timeout=5s
          --health-retries=15
```

Add a wait-for-CockroachDB step (similar to the PostgreSQL wait step) and add `./integration_tests/cockroachdb/...` to the integration test run command.

### 6. Known CockroachDB-Specific Test Considerations

1. **Savepoints**: CockroachDB supports `SAVEPOINT` but with different semantics. Nested transactions in CockroachDB use `SAVEPOINT` internally. Neat's savepoint implementation should work, but tests should verify correct rollback behavior.

2. **Auto-Increment / Serial**: CockroachDB's `SERIAL` type uses `unique_rowid()` which generates non-sequential IDs (combines timestamp + random). Tests should not assert on sequential IDs, only on uniqueness and non-zero values. Consider using `BITINT` with explicit sequence for tests that need sequential IDs.

3. **RETURNING Clause**: CockroachDB supports `INSERT ... RETURNING` and `UPDATE ... RETURNING` like PostgreSQL. This should work with Neat's PostgreSQL driver.

4. **JSON/JSONB**: CockroachDB supports `JSONB` type and most PostgreSQL JSON functions. Basic JSON operations should work. Advanced functions like `jsonb_each` may have differences.

5. **SELECT FOR UPDATE**: CockroachDB supports `SELECT ... FOR UPDATE` but in distributed mode, locking semantics differ from PostgreSQL. Lock tests should verify basic functionality without asserting on specific locking behavior.

6. **Index Types**: CockroachDB supports B-tree, GIN, and GiST indexes but with some limitations compared to PostgreSQL. Index tests should focus on basic index creation and usage.

7. **Charset/Collation**: CockroachDB defaults to UTF-8 encoding. Collation support is more limited than PostgreSQL. Tests relying on specific collations may need adjustment.

8. **ALTER TABLE**: CockroachDB's `ALTER TABLE` support is comprehensive but some operations may be online-only (non-blocking) compared to PostgreSQL. `RENAME COLUMN` should work but may have slightly different behavior.

9. **Foreign Keys**: CockroachDB fully supports foreign keys. This should work the same as PostgreSQL.

## Implementation Plan

### Phase 1: Infrastructure (Low Effort)
- Add CockroachDB service to `docker-compose.yml`
- Create `integration_tests/cockroachdb/helper.go`
- Add CockroachDB environment variables to CI workflow
- Add CockroachDB service container to CI workflow
- Add wait-for-CockroachDB step to CI

### Phase 2: Core Tests (Medium Effort)
- Port the PostgreSQL connection, find, create, update, delete, count, where, and order/limit/offset tests
- These are the most critical tests and should be straightforward since CockroachDB is PostgreSQL-compatible

### Phase 3: Advanced Tests (Medium Effort)
- Port association, transaction, soft delete, schema builder, and view tests
- Pay attention to savepoint behavior and `SERIAL`/`RETURNING` differences

### Phase 4: Edge Case Tests (Low Effort)
- Port JSON, raw SQL, and dotted column tests
- Document any CockroachDB-specific behavior differences

## Estimated Effort

- **Phase 1**: 2-3 hours (Docker Compose + helper + CI)
- **Phase 2**: 4-6 hours (core test porting — mostly mechanical copy with `cockroachdb` prefix)
- **Phase 3**: 3-4 hours (advanced tests + CockroachDB-specific adjustments)
- **Phase 4**: 1-2 hours (edge cases)

**Total**: ~10-15 hours

## Risks

1. **CockroachDB container size**: The CockroachDB Docker image is large (~300MB compressed). CI runs may slow down. Mitigation: use a specific version tag, not `latest`.

2. **CockroachDB startup time**: CockroachDB in single-node mode starts relatively quickly, but the healthcheck should account for initialization time (15 retries × 10s = 150s max wait).

3. **Serial ID non-sequentiality**: Tests that assert on sequential auto-increment IDs will fail. All such tests need to be updated to assert on non-zero uniqueness only.

4. **Placeholder syntax**: CockroachDB uses `$1, $2, ...` like PostgreSQL. Neat's PostgreSQL driver should handle this correctly, but tests should verify.

5. **SSL Mode**: CockroachDB in `--insecure` mode requires `sslmode=disable` in the connection string. The helper config must set this correctly.

## Success Criteria

1. `docker-compose up -d cockroachdb` starts a healthy CockroachDB instance
2. `go test -v ./integration_tests/cockroachdb/...` passes all tests locally
3. GitHub Actions CI runs CockroachDB integration tests on every push/PR
4. All CockroachDB-specific behavior differences are documented in the helper file
5. The comparison tables' `✅` for CockroachDB is backed by actual test evidence

## References

- CockroachDB Documentation: https://www.cockroachlabs.com/docs/
- CockroachDB PostgreSQL Compatibility: https://www.cockroachlabs.com/docs/stable/postgresql-compatibility.html
- CockroachDB Docker Image: https://hub.docker.com/r/cockroachdb/cockroach
- Neat Driver Registration: `docs/driver-registration.html`
- Existing PostgreSQL Integration Tests: `integration_tests/postgres/`
