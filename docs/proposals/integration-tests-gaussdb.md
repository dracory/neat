# Proposal: GaussDB Support and Integration Tests

**Date**: August 8, 2026
**Status**: Proposal
**Priority**: Medium

## Problem

GaussDB is Huawei's distributed database, widely used in enterprise environments (particularly in China). GORM supports it via a dedicated driver (`gorm.io/driver/gaussdb`), and the comparison tables currently show `❌` for Neat's GaussDB support.

GaussDB is PostgreSQL-compatible at the wire protocol level, which means Neat's PostgreSQL driver **may** work with it. However, this is unverified — there are enough differences (type mappings, SQL dialect variations, authentication) that GORM forked their PostgreSQL driver specifically for GaussDB.

We need to:
1. Investigate whether Neat's PostgreSQL driver works with GaussDB as-is
2. If not, add GaussDB-specific support (driver or dialect adjustments)
3. Add integration tests to verify and maintain compatibility
4. Update the comparison tables from `❌` to `✅` (only after tests pass)

## Background

### GaussDB vs openGauss

- **GaussDB** is Huawei's commercial distributed database, available on Huawei Cloud
- **openGauss** is the open-source core of GaussDB, freely available and Docker-izable
- For integration testing, we use the **openGauss** Docker image since commercial GaussDB is cloud-only

### PostgreSQL Compatibility

GaussDB/openGauss is PostgreSQL-compatible but with important differences:

1. **Wire Protocol**: GaussDB speaks PostgreSQL wire protocol, but authentication may differ (SHA256-based password authentication by default vs PostgreSQL's MD5/scram-sha-256)

2. **Go Driver**: Huawei provides a dedicated Go driver (`github.com/HuaweiCloudDeveloper/gaussdb-go`) separate from `lib/pq`. The Huawei compatibility docs state: "The Go driver of the database cannot coexist with that of PostgreSQL" — meaning the `gaussdb-go` driver may conflict with `lib/pq` if both are registered for the same `database/sql` driver name.

3. **SQL Dialect**: GORM's GaussDB driver (a fork of `gorm/postgres`) handles "GaussDB specific SQL differences" and "proper type mappings for GaussDB specific data types". The exact differences are not fully documented but include:
   - Type name differences (e.g., some PostgreSQL types may have different names)
   - Index syntax variations
   - Potential differences in `RETURNING` clause behavior
   - GaussDB-specific data types not present in PostgreSQL

4. **Default Port**: openGauss listens on port `5432` (same as PostgreSQL), so Docker Compose must use a different port mapping.

5. **Default User**: openGauss uses `gaussdb` as the default user (not `postgres`).

## Design

### Phase 1: Investigation (Critical — determines the approach)

Before writing tests, we need to determine whether Neat's existing PostgreSQL driver (`lib/pq`) can connect to openGauss, or whether we need to integrate the `gaussdb-go` driver.

#### Option A: `lib/pq` works directly

If openGauss accepts `lib/pq` connections (it may, since it speaks PostgreSQL protocol), then:
- No code changes needed in Neat
- GaussDB support is "free" via the PostgreSQL driver
- We only need integration tests to verify

**Risk**: openGauss's default authentication (`sha256`) may not be supported by `lib/pq`. This can be worked around by configuring openGauss to accept `md5` or `trust` authentication in the test container.

#### Option B: Need `gaussdb-go` driver

If `lib/pq` doesn't work, we need to:
- Add `github.com/HuaweiCloudDeveloper/gaussdb-go` as a dependency
- Register it as a new driver in Neat's driver registry (e.g., `gaussdb`)
- Handle the potential conflict with `lib/pq` (both may register for `database/sql` under similar names)
- Add GaussDB-specific dialect adjustments if needed

**Risk**: The `gaussdb-go` driver and `lib/pq` may conflict in the same binary. The Huawei docs explicitly warn about this. We may need to use build tags or conditional compilation to handle this.

### Phase 2: Docker Compose Service (regardless of Phase 1 outcome)

Add an openGauss service to `docker-compose.yml`:

```yaml
  opengauss:
    image: enmotech/opengauss:5.0.0
    container_name: neat-opengauss-test
    environment:
      GS_PASSWORD: GaussDB@123
      GS_USERNAME: gaussdb
      GS_DB: test
    ports:
      - "6432:5432"
    healthcheck:
      test: ["CMD", "gsql", "-d", "test", "-U", "gaussdb", "-W", "GaussDB@123", "-c", "SELECT 1"]
      interval: 10s
      timeout: 5s
      retries: 15
```

**Notes**:
- We use port `6432` externally to avoid conflicts with the PostgreSQL service on `55432`
- The `enmotech/opengauss` image is a community-maintained image with better Docker defaults than the official `opengauss/opengauss` image
- `GS_PASSWORD` sets the password for the default `gaussdb` user
- Authentication is configured to accept password-based connections

### Phase 3: Integration Test Directory

Create `integration_tests/gaussdb/` with the same structure as `integration_tests/postgres/`:

```
integration_tests/gaussdb/
├── helper.go                                  # Connection config, setup, table creation
├── gaussdb_connection_test.go                 # Basic connection test (CRITICAL — validates Phase 1)
├── gaussdb_find_test.go                       # Find/First/Get queries
├── gaussdb_query_create_test.go               # INSERT operations
├── gaussdb_query_update_test.go               # UPDATE operations
├── gaussdb_query_delete_test.go               # DELETE operations
├── gaussdb_query_count_test.go                # COUNT and aggregates
├── gaussdb_query_where_test.go                # WHERE clauses
├── gaussdb_query_join_test.go                 # JOIN operations
├── gaussdb_query_order_limit_offset_test.go   # ORDER BY, LIMIT, OFFSET
├── gaussdb_query_group_having_test.go         # GROUP BY, HAVING
├── gaussdb_query_association_test.go          # Associations
├── gaussdb_query_belongs_to_test.go           # BelongsTo-specific tests
├── gaussdb_query_chunk_test.go                # Chunk processing
├── gaussdb_query_paginate_test.go             # Pagination
├── gaussdb_query_pluck_test.go                # Pluck
├── gaussdb_query_value_test.go                # Value extraction
├── gaussdb_query_distinct_test.go             # DISTINCT
├── gaussdb_query_select_test.go               # SELECT specific columns
├── gaussdb_query_omit_test.go                 # Omit columns
├── gaussdb_query_load_test.go                 # Eager loading
├── gaussdb_query_lock_test.go                 # Pessimistic locking
├── gaussdb_query_scopes_test.go               # Query scopes
├── gaussdb_query_to_sql_test.go               # ToSql interface
├── gaussdb_query_increment_decrement_test.go  # Increment/Decrement
├── gaussdb_query_update_or_insert_test.go     # UpdateOrInsert
├── gaussdb_query_json_test.go                 # JSON column operations
├── gaussdb_query_aggregate_test.go            # SUM, AVG, MIN, MAX
├── gaussdb_raw_test.go                        # Raw SQL queries
├── gaussdb_transaction_test.go                # Transactions and savepoints
├── gaussdb_soft_delete_test.go                # Soft deletes
├── gaussdb_schema_table_test.go               # Schema builder: table operations
├── gaussdb_schema_column_types_test.go        # Schema builder: column types
├── gaussdb_schema_column_methods_test.go      # Schema builder: column methods
├── gaussdb_schema_column_modifiers_test.go    # Schema builder: column modifiers
├── gaussdb_schema_column_change_test.go       # Schema builder: column changes
├── gaussdb_schema_foreign_key_test.go         # Schema builder: foreign keys
├── gaussdb_schema_index_test.go               # Schema builder: indexes
├── gaussdb_schema_rename_column_test.go       # Schema builder: rename column
├── gaussdb_schema_timestamp_test.go           # Schema builder: timestamps
├── gaussdb_schema_view_test.go                # View management
└── gaussdb_dotted_column_test.go              # Dotted column references
```

### Phase 4: Helper File

`integration_tests/gaussdb/helper.go` (Option A — using PostgreSQL driver):

```go
package gaussdb_test

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

// GetGaussDBConfig returns a GaussDB connection config from environment variables.
// GaussDB/openGauss speaks the PostgreSQL wire protocol, so we use the PostgreSQL driver.
func GetGaussDBConfig() neat.DBConfig {
    host := common.GetEnv("GAUSSDB_HOST", "127.0.0.1")
    port := common.GetEnvInt("GAUSSDB_PORT", 6432)
    database := common.GetEnv("GAUSSDB_DATABASE", "test")
    username := common.GetEnv("GAUSSDB_USER", "gaussdb")
    password := common.GetEnv("GAUSSDB_PASS", "GaussDB@123")

    return neat.DBConfig{
        Default: "gaussdb",
        Connections: map[string]neat.ConnectionConfig{
            "gaussdb": {
                Driver:   "postgres", // GaussDB uses the PostgreSQL driver
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
```

If Option A fails and we need the `gaussdb-go` driver (Option B), the helper would change to:

```go
    _ "github.com/HuaweiCloudDeveloper/gaussdb-go"
    // ... and the Driver field would be "gaussdb" instead of "postgres"
```

### Phase 5: Environment Variables

Add GaussDB environment variables to the CI workflow:

```
GAUSSDB_HOST: 127.0.0.1
GAUSSDB_PORT: 6432
GAUSSDB_DATABASE: test
GAUSSDB_USER: gaussdb
GAUSSDB_PASS: GaussDB@123
```

### Phase 6: GitHub Actions CI

Add an openGauss service container to `.github/workflows/tests.yml`:

```yaml
      opengauss:
        image: enmotech/opengauss:5.0.0
        env:
          GS_PASSWORD: GaussDB@123
          GS_USERNAME: gaussdb
          GS_DB: test
        ports:
          - "6432:5432"
        options: >-
          --health-cmd="gsql -d test -U gaussdb -W GaussDB@123 -c 'SELECT 1'"
          --health-interval=10s
          --health-timeout=5s
          --health-retries=15
```

### Known GaussDB-Specific Test Considerations

1. **Authentication**: openGauss defaults to SHA256 password authentication. `lib/pq` may not support this. The test container should be configured to accept `md5` or `trust` authentication. This can be done by modifying `pg_hba.conf` in the container or using an image that defaults to `md5`.

2. **Type Mappings**: GaussDB has some type name differences from PostgreSQL. The schema builder tests should verify that column types are created correctly. GORM's GaussDB driver includes "proper type mappings for GaussDB specific data types" — we need to identify what these are and test for them.

3. **Savepoints**: openGauss supports `SAVEPOINT` syntax like PostgreSQL. Transaction tests should work, but verify behavior.

4. **Serial/Auto-Increment**: openGauss supports `SERIAL` and `BIGSERIAL` like PostgreSQL. Auto-increment behavior should be identical.

5. **RETURNING Clause**: openGauss supports `INSERT ... RETURNING` like PostgreSQL. This should work with Neat's PostgreSQL driver.

6. **JSON/JSONB**: openGauss supports `JSONB` type. Basic JSON operations should work.

7. **Index Types**: openGauss supports B-tree indexes. Some PostgreSQL-specific index types (GIN, GiST) may have different behavior or availability.

8. **Connection String**: openGauss accepts PostgreSQL-style connection strings. The `sslmode=disable` parameter is required for non-TLS connections.

9. **Driver Conflict**: If using `gaussdb-go` driver (Option B), it may conflict with `lib/pq` in the same binary. This needs to be resolved before mixing GaussDB and PostgreSQL tests in the same test run.

## Implementation Plan

### Phase 1: Investigation (2-3 hours)
- Start openGauss container locally
- Attempt connection with `lib/pq` driver
- If connection fails, attempt with `gaussdb-go` driver
- Document findings and determine Option A or Option B
- If Option B, design the driver registration approach

### Phase 2: Infrastructure (2-3 hours)
- Add openGauss service to `docker-compose.yml`
- Create `integration_tests/gaussdb/helper.go` (based on Phase 1 findings)
- Add GaussDB environment variables to CI workflow
- Add openGauss service container to CI workflow
- Add wait-for-GaussDB step to CI

### Phase 3: Core Tests (4-6 hours)
- Start with `gaussdb_connection_test.go` — this is the go/no-go test
- Port the PostgreSQL connection, find, create, update, delete, count, where, and order/limit/offset tests
- If any test fails due to GaussDB-specific behavior, document the difference

### Phase 4: Advanced Tests (3-4 hours)
- Port association, transaction, soft delete, schema builder, and view tests
- Pay attention to type mappings and index creation

### Phase 5: Edge Case Tests (1-2 hours)
- Port JSON, raw SQL, and dotted column tests
- Document any GaussDB-specific behavior differences

### Phase 6: Documentation Update (1 hour)
- If tests pass with `lib/pq` (Option A): Add GaussDB to `docs/driver-registration.html` as compatible
- If tests require `gaussdb-go` (Option B): Add GaussDB as a new supported driver in docs
- Update comparison tables from `❌` to `✅` for GaussDB support
- Update conclusion sections in comparison files

## Estimated Effort

- **Phase 1**: 2-3 hours (investigation — critical path)
- **Phase 2**: 2-3 hours (Docker Compose + helper + CI)
- **Phase 3**: 4-6 hours (core test porting)
- **Phase 4**: 3-4 hours (advanced tests)
- **Phase 5**: 1-2 hours (edge cases)
- **Phase 6**: 1 hour (documentation)

**Total**: ~13-19 hours (more than TiDB/CockroachDB due to the investigation phase and potential driver integration work)

## Risks

1. **`lib/pq` incompatibility**: If openGauss's authentication or protocol differences prevent `lib/pq` from connecting, we need to integrate a new driver. This adds complexity and potential dependency conflicts.

2. **Driver conflict**: The `gaussdb-go` driver may conflict with `lib/pq` in the same binary. This could require build tags, conditional compilation, or a separate test binary.

3. **openGauss vs GaussDB differences**: openGauss is the open-source core, but commercial GaussDB may have additional features or differences. Tests against openGauss may not cover all commercial GaussDB scenarios.

4. **Container image availability**: The `enmotech/opengauss` image is community-maintained. If it becomes unmaintained, we may need to switch to `opengauss/opengauss` or build a custom image.

5. **Authentication configuration**: openGauss's default SHA256 authentication may require container customization to work with `lib/pq`. This adds complexity to the Docker Compose setup.

6. **Limited documentation**: GaussDB's PostgreSQL compatibility documentation is not comprehensive. We may discover unexpected differences during testing.

## Success Criteria

1. Phase 1 investigation determines whether `lib/pq` or `gaussdb-go` is needed
2. `docker-compose up -d opengauss` starts a healthy openGauss instance
3. `gaussdb_connection_test.go` passes — Neat can connect to openGauss
4. `go test -v ./integration_tests/gaussdb/...` passes all tests locally
5. GitHub Actions CI runs GaussDB integration tests on every push/PR
6. All GaussDB-specific behavior differences are documented in the helper file
7. `docs/driver-registration.html` is updated to list GaussDB as compatible
8. Comparison tables are updated from `❌` to `✅` for GaussDB support

## Decision Required

Before starting implementation, the following question must be answered:

**Should we use `lib/pq` (Option A, simpler but may not work) or `gaussdb-go` (Option B, more work but guaranteed compatibility)?**

Recommendation: Start with Option A (`lib/pq`). If the connection test fails, fall back to Option B. This minimizes risk and effort if `lib/pq` works.

## References

- GaussDB Documentation: https://support.huaweicloud.com/intl/en-us/distributed-devg-v8-gaussdb/gaussdb-12-0245.html
- openGauss Docker Image: https://hub.docker.com/r/opengauss/opengauss
- gaussdb-go Driver: https://github.com/HuaweiCloudDeveloper/gaussdb-go
- GORM GaussDB Driver: https://github.com/go-gorm/gaussdb
- GORM GaussDB PR: https://github.com/go-gorm/gorm/pull/7508
- Neat Driver Registration: `docs/driver-registration.html`
- Existing PostgreSQL Integration Tests: `integration_tests/postgres/`
