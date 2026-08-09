# Proposal: TiDB Integration Tests

**Date**: August 8, 2026
**Status**: Proposal
**Priority**: Medium

## Problem

Neat ORM claims TiDB support via the MySQL driver (documented in `docs/driver-registration.html`), but there are **no integration tests** verifying this. The comparison tables list TiDB as `✅` for Neat, yet we have zero test coverage proving that Neat's query builder, schema builder, migrations, associations, and soft deletes actually work against a real TiDB instance.

If a user picks Neat because of the TiDB claim and something breaks, we won't know until they report it.

## Goal

Add a full integration test suite for TiDB that mirrors the existing MySQL integration tests, running against a real TiDB container in Docker Compose and GitHub Actions CI.

## Background

TiDB is a distributed SQL database with **very high MySQL compatibility**. It speaks the MySQL wire protocol and accepts the `go-sql-driver/mysql` driver. Neat's MySQL driver should work with TiDB out of the box, but there are edge cases:

- TiDB does not support `SAVEPOINT` (savepoints are emulated by Neat's transaction layer, but this needs verification)
- TiDB's `AUTO_RANDOM` vs `AUTO_INCREMENT` behavior differs
- TiDB may have different defaults for `sql_mode`
- TiDB's handling of `FOREIGN KEY` constraints has historically been limited (though TiDB v6.6+ supports them)
- TiDB's `JSON` functions are a subset of MySQL's

These differences are exactly why we need integration tests.

## Design

### 1. Docker Compose Service

Add a TiDB service to `docker-compose.yml`:

```yaml
  tidb:
    image: pingcap/tidb:v8.5.0
    container_name: neat-tidb-test
    ports:
      - "4000:4000"
    # TiDB doesn't have a built-in healthcheck, so we use mysqladmin
    # against the TiDB port (4000) since TiDB speaks the MySQL protocol
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "127.0.0.1", "-P", "4000", "-u", "root"]
      interval: 10s
      timeout: 5s
      retries: 15
```

TiDB listens on port `4000` by default and accepts MySQL connections with no password by default (`root` user, no password).

### 2. Integration Test Directory

Create `integration_tests/tidb/` with the same structure as `integration_tests/mysql/`:

```
integration_tests/tidb/
├── helper.go                          # Connection config, setup, table creation
├── tidb_connection_test.go            # Basic connection test
├── tidb_find_test.go                  # Find/First/Get queries
├── tidb_query_create_test.go          # INSERT operations
├── tidb_query_update_test.go          # UPDATE operations
├── tidb_query_delete_test.go          # DELETE operations
├── tidb_query_count_test.go           # COUNT and aggregates
├── tidb_query_where_test.go           # WHERE clauses
├── tidb_query_join_test.go            # JOIN operations
├── tidb_query_order_limit_offset_test.go  # ORDER BY, LIMIT, OFFSET
├── tidb_query_group_having_test.go    # GROUP BY, HAVING
├── tidb_query_association_test.go     # Associations (BelongsTo, HasMany, HasOne)
├── tidb_query_belongs_to_test.go      # BelongsTo-specific tests
├── tidb_query_chunk_test.go           # Chunk processing
├── tidb_query_paginate_test.go        # Pagination
├── tidb_query_pluck_test.go           # Pluck
├── tidb_query_value_test.go           # Value extraction
├── tidb_query_distinct_test.go        # DISTINCT
├── tidb_query_select_test.go          # SELECT specific columns
├── tidb_query_omit_test.go            # Omit columns
├── tidb_query_load_test.go            # Eager loading
├── tidb_query_lock_test.go            # Pessimistic locking
├── tidb_query_scopes_test.go          # Query scopes
├── tidb_query_to_sql_test.go          # ToSql interface
├── tidb_query_increment_decrement_test.go  # Increment/Decrement
├── tidb_query_update_or_insert_test.go     # UpdateOrInsert
├── tidb_query_json_test.go            # JSON column operations
├── tidb_query_aggregate_test.go       # SUM, AVG, MIN, MAX
├── tidb_raw_test.go                   # Raw SQL queries
├── tidb_transaction_test.go           # Transactions and savepoints
├── tidb_soft_delete_test.go           # Soft deletes
├── tidb_schema_table_test.go          # Schema builder: table operations
├── tidb_schema_column_types_test.go   # Schema builder: column types
├── tidb_schema_column_methods_test.go # Schema builder: column methods
├── tidb_schema_column_modifiers_test.go    # Schema builder: column modifiers
├── tidb_schema_column_change_test.go  # Schema builder: column changes
├── tidb_schema_foreign_key_test.go    # Schema builder: foreign keys
├── tidb_schema_index_test.go          # Schema builder: indexes
├── tidb_schema_rename_column_test.go  # Schema builder: rename column
├── tidb_schema_timestamp_test.go      # Schema builder: timestamps
├── tidb_schema_view_test.go           # View management
└── tidb_dotted_column_test.go         # Dotted column references
```

### 3. Helper File

`integration_tests/tidb/helper.go`:

```go
package tidb_test

import (
    "fmt"
    "testing"
    "time"

    "github.com/dracory/neat"
    "github.com/dracory/neat/contracts/log"
    "github.com/dracory/neat/database"
    "github.com/dracory/neat/integration_tests/common"
    _ "github.com/go-sql-driver/mysql"
)

// GetTiDBConfig returns a TiDB connection config from environment variables.
// TiDB speaks the MySQL wire protocol, so we use the MySQL driver.
func GetTiDBConfig() neat.DBConfig {
    host := common.GetEnv("TIDB_HOST", "127.0.0.1")
    port := common.GetEnvInt("TIDB_PORT", 4000)
    database := common.GetEnv("TIDB_DATABASE", "test")
    username := common.GetEnv("TIDB_USER", "root")
    password := common.GetEnv("TIDB_PASS", "")

    return neat.DBConfig{
        Default: "tidb",
        Connections: map[string]neat.ConnectionConfig{
            "tidb": {
                Driver:   "mysql", // TiDB uses the MySQL driver
                Host:     host,
                Port:     port,
                Database: database,
                Username: username,
                Password: password,
                Charset:  "utf8mb4",
                Loc:      "Local",
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

// SetupTiDBTest sets up a TiDB connection and creates test tables.
func SetupTiDBTest(t *testing.T) *database.DB {
    t.Helper()
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    config := GetTiDBConfig()
    db, err := neat.New(config)
    if err != nil {
        t.Fatalf("Failed to connect to TiDB: %v", err)
    }

    // Create test database if it doesn't exist
    // TiDB defaults to no password, root user
    err = createTiDBDatabase(db)
    if err != nil {
        t.Fatalf("Failed to create TiDB test database: %v", err)
    }

    // Create tables
    createTiDBTables(t, db)

    return db
}

// SetupTiDBConnection sets up a TiDB connection without creating tables.
func SetupTiDBConnection(t *testing.T) *database.DB {
    t.Helper()
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    config := GetTiDBConfig()
    db, err := neat.New(config)
    if err != nil {
        t.Fatalf("Failed to connect to TiDB: %v", err)
    }

    return db
}
```

The table creation SQL should be identical to the MySQL helper's SQL, since TiDB is MySQL-compatible. Any divergences (e.g., foreign key syntax, JSON function differences) should be documented in the helper file.

### 4. Environment Variables

Add TiDB environment variables to the CI workflow:

```
TIDB_HOST: 127.0.0.1
TIDB_PORT: 4000
TIDB_DATABASE: test
TIDB_USER: root
TIDB_PASS: ""
```

### 5. GitHub Actions CI

Add a TiDB service container to `.github/workflows/tests.yml`:

```yaml
      tidb:
        image: pingcap/tidb:v8.5.0
        ports:
          - "4000:4000"
        options: >-
          --health-cmd="mysqladmin ping -h 127.0.0.1 -P 4000 -u root"
          --health-interval=10s
          --health-timeout=5s
          --health-retries=15
```

Add a wait-for-TiDB step (similar to the MySQL wait step) and add `./integration_tests/tidb/...` to the integration test run command.

### 6. Known TiDB-Specific Test Considerations

1. **Savepoints**: TiDB does not support `SAVEPOINT` syntax. Neat's transaction layer should gracefully handle this. Tests should verify that savepoint calls don't cause errors (they may be no-ops or emulated).

2. **Foreign Keys**: TiDB v6.6+ supports foreign keys, but behavior may differ from MySQL. The foreign key tests should be run but may need TiDB-specific assertions.

3. **Auto-Increment**: TiDB's `AUTO_INCREMENT` is not guaranteed to be sequential (distributed allocation). Tests should not assert on sequential IDs, only on uniqueness and non-zero values.

4. **JSON Functions**: TiDB supports a subset of MySQL's JSON functions. JSON tests should focus on basic operations (store, retrieve, query) rather than advanced functions like `JSON_TABLE`.

5. **Charset/Collation**: TiDB defaults to `utf8mb4_bin` collation, which is case-sensitive. MySQL defaults to `utf8mb4_0900_ai_ci` (case-insensitive). Tests that rely on case-insensitive string comparisons may need adjustment.

## Implementation Plan

### Phase 1: Infrastructure (Low Effort)
- Add TiDB service to `docker-compose.yml`
- Create `integration_tests/tidb/helper.go`
- Add TiDB environment variables to CI workflow
- Add TiDB service container to CI workflow
- Add wait-for-TiDB step to CI

### Phase 2: Core Tests (Medium Effort)
- Port the MySQL connection, find, create, update, delete, count, where, and order/limit/offset tests
- These are the most critical tests and should be straightforward since TiDB is MySQL-compatible

### Phase 3: Advanced Tests (Medium Effort)
- Port association, transaction, soft delete, schema builder, and view tests
- Pay attention to savepoint and foreign key behavior

### Phase 4: Edge Case Tests (Low Effort)
- Port JSON, raw SQL, and dotted column tests
- Document any TiDB-specific behavior differences

## Estimated Effort

- **Phase 1**: 2-3 hours (Docker Compose + helper + CI)
- **Phase 2**: 4-6 hours (core test porting — mostly mechanical copy with `tidb` prefix)
- **Phase 3**: 3-4 hours (advanced tests + TiDB-specific adjustments)
- **Phase 4**: 1-2 hours (edge cases)

**Total**: ~10-15 hours

## Risks

1. **TiDB container size**: The TiDB Docker image is large (~1GB). CI runs may slow down. Mitigation: use a specific version tag, not `latest`.

2. **TiDB startup time**: TiDB may take longer to start than MySQL. The healthcheck retries should account for this (15 retries × 10s = 150s max wait).

3. **Savepoint incompatibility**: If Neat's savepoint implementation relies on native `SAVEPOINT` SQL and TiDB doesn't support it, tests will fail. This would be a real bug to fix, not just a test issue.

4. **Foreign key behavior**: TiDB's foreign key enforcement may differ. Tests may need conditional assertions.

## Success Criteria

1. `docker-compose up -d tidb` starts a healthy TiDB instance
2. `go test -v ./integration_tests/tidb/...` passes all tests locally
3. GitHub Actions CI runs TiDB integration tests on every push/PR
4. All TiDB-specific behavior differences are documented in the helper file
5. The comparison tables' `✅` for TiDB is backed by actual test evidence

## References

- TiDB Documentation: https://docs.pingcap.com/tidb/
- TiDB MySQL Compatibility: https://docs.pingcap.com/tidb/stable/mysql-compatibility
- TiDB Docker Image: https://hub.docker.com/r/pingcap/tidb
- Neat Driver Registration: `docs/driver-registration.html`
- Existing MySQL Integration Tests: `integration_tests/mysql/`
