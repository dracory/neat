# Proposal: ClickHouse Support Investigation and Integration Tests

**Date**: August 8, 2026
**Status**: Proposal
**Priority**: Low

## Problem

GORM supports ClickHouse via a dedicated driver (`gorm.io/driver/clickhouse`), and the comparison tables currently show `❌` for Neat's ClickHouse support. The question is whether Neat **should** support ClickHouse, and if so, what that looks like.

ClickHouse is fundamentally different from MySQL, PostgreSQL, TiDB, CockroachDB, and GaussDB — it is a **columnar OLAP database** designed for analytical workloads, not an OLTP database. This has significant implications for ORM support.

## Background

### What ClickHouse Is

ClickHouse is a column-oriented database management system for online analytical processing (OLAP). It is optimized for:
- High-throughput INSERT operations (millions of rows per second)
- Fast aggregation queries over large datasets
- Real-time analytics on event/streaming data

It is **not** designed for:
- Point queries (single-row lookups by ID)
- Frequent UPDATE/DELETE operations
- Transactional workloads with ACID guarantees
- Foreign key enforcement
- Traditional ORM use cases

### ClickHouse SQL Dialect

ClickHouse uses a SQL dialect that differs significantly from MySQL/PostgreSQL:

1. **No FOREIGN KEYS**: ClickHouse does not support foreign key constraints. This means association loading (BelongsTo, HasMany, HasOne) cannot rely on database-level referential integrity.

2. **No SAVEPOINT**: ClickHouse does not support savepoints. Transactions are limited.

3. **No traditional AUTO_INCREMENT**: ClickHouse does not have `AUTO_INCREMENT`. It uses `UUID` or sequence-based approaches for ID generation.

4. **UPDATE/DELETE are expensive**: ClickHouse does not support traditional `UPDATE` or `DELETE` statements efficiently. They use `ALTER TABLE ... UPDATE` and `ALTER TABLE ... DELETE` which are heavyweight mutations. Lightweight DELETE (`DELETE FROM`) was added in recent versions but is still not as efficient as OLTP databases.

5. **Different data types**: ClickHouse uses its own type system (`UInt64`, `String`, `DateTime64`, `Array(T)`, `Map(K,V)`, `Tuple(T1,T2)`) rather than MySQL/PostgreSQL types.

6. **No traditional indexes**: ClickHouse uses sparse indexes (primary key index, skip indexes) rather than B-tree indexes. The concept of "creating an index" is different.

7. **MergeTree engine**: Tables are backed by MergeTree family engines, which have specific requirements (ORDER BY clause is mandatory for most engines, data is merged asynchronously in the background).

### Connection Options

ClickHouse offers multiple connection protocols:

1. **Native protocol** (port 9000 TCP): ClickHouse's own binary protocol. Best performance. Requires the `clickhouse-go` driver.

2. **HTTP interface** (port 8123): REST-like interface. Works with `clickhouse-go` driver in HTTP mode.

3. **MySQL wire protocol** (port 9004): Emulates MySQL protocol for compatibility with MySQL clients. **Known to have issues with `go-sql-driver/mysql`** — there are open GitHub issues about connection failures and incomplete compatibility.

4. **PostgreSQL wire protocol** (port 9005): Emulates PostgreSQL protocol. Less commonly used.

### Go Driver

The official ClickHouse Go driver is `github.com/ClickHouse/clickhouse-go`. It supports:
- Native protocol (fast, recommended)
- HTTP protocol
- `database/sql` interface (slower than native but compatible with ORMs)

The driver registers under the name `"clickhouse"` for `database/sql`.

## Feasibility Assessment

### What Would Work

If we integrate the `clickhouse-go` driver and register it in Neat:

| Feature | Feasibility | Notes |
|---------|-------------|-------|
| Query Builder (SELECT) | ✅ Likely works | Basic SELECT, WHERE, ORDER BY, LIMIT, GROUP BY should work |
| INSERT | ✅ Likely works | Batch inserts are ClickHouse's strength |
| Aggregates (COUNT, SUM, AVG) | ✅ Likely works | ClickHouse is optimized for this |
| Raw SQL | ✅ Works | Any SQL ClickHouse accepts |
| ToSql Interface | ✅ Works | SQL generation without execution |
| Connection Pooling | ✅ Works | Via `database/sql` |
| Context Support | ✅ Works | Via `database/sql` |

### What Would NOT Work (or would require significant changes)

| Feature | Feasibility | Notes |
|---------|-------------|-------|
| UPDATE | ⚠️ Problematic | ClickHouse uses `ALTER TABLE ... UPDATE`, not standard `UPDATE` |
| DELETE | ⚠️ Problematic | ClickHouse uses `ALTER TABLE ... DELETE` or lightweight `DELETE FROM` |
| Soft Deletes | ❌ Impractical | UPDATE-based soft deletes would be extremely expensive on large tables |
| Transactions | ❌ Not supported | ClickHouse does not support ACID transactions |
| Savepoints | ❌ Not supported | ClickHouse does not support savepoints |
| Foreign Keys | ❌ Not supported | ClickHouse does not enforce foreign keys |
| Associations (BelongsTo, HasMany, HasOne) | ⚠️ Limited | Can work at query level (JOINs) but no DB-level enforcement |
| Schema Builder (indexes) | ❌ Different model | ClickHouse uses sparse indexes, not B-tree indexes |
| Schema Builder (foreign keys) | ❌ Not supported | ClickHouse has no foreign keys |
| Migrations | ⚠️ Limited | DDL works but semantics differ (MergeTree engines, mandatory ORDER BY) |
| Auto-increment IDs | ❌ Not available | ClickHouse uses UUID or other ID strategies |
| Observers (before/after update) | ⚠️ Limited | UPDATE is a mutation, not a row-level operation |
| Polymorphic Associations | ❌ Impractical | Depends on foreign key semantics |

### The Core Question

**Is ClickHouse a good fit for an OLTP ORM like Neat?**

**Short answer**: No, not as a full-featured ORM target. ClickHouse is an OLAP database, and most ORM features (transactions, foreign keys, soft deletes, auto-increment, efficient UPDATE/DELETE) either don't exist or work fundamentally differently.

**However**, there is a valid use case: **read-only analytics**. A user might want to:
1. Use Neat's query builder to run analytical queries against ClickHouse
2. Use aggregates, GROUP BY, and JOINs for reporting
3. Insert batch data for analytics
4. Use ToSql to generate SQL without executing

This "analytics-only mode" is a subset of Neat's features, but it could be valuable.

## Design

### Option A: Full ClickHouse Driver (Not Recommended)

Integrate `clickhouse-go` as a new driver and attempt to make all Neat features work. This would require:
- New driver registration
- Dialect-specific SQL generation for UPDATE/DELETE (ALTER TABLE mutations)
- Graceful degradation for unsupported features (transactions, savepoints, foreign keys)
- Custom schema builder for ClickHouse engines and sparse indexes
- Custom migration system for ClickHouse DDL

**Effort**: Very High (40+ hours)
**Risk**: High — many features would be stubs or no-ops, creating a confusing user experience
**Verdict**: Not recommended. The effort doesn't match the value.

### Option B: Read-Only Analytics Mode (Recommended)

Integrate `clickhouse-go` as a new driver with a **clearly documented reduced feature set**. Only support:

1. **Connection**: `clickhouse-go` driver via `database/sql`
2. **Query Builder**: SELECT, WHERE, ORDER BY, LIMIT, OFFSET, GROUP BY, HAVING, JOIN
3. **Aggregates**: COUNT, SUM, AVG, MIN, MAX
4. **INSERT**: Batch inserts (ClickHouse's strength)
5. **Raw SQL**: For anything the query builder doesn't cover
6. **ToSql**: SQL generation without execution

Explicitly **not supported** (with clear error messages):
- UPDATE (suggest raw SQL with `ALTER TABLE ... UPDATE`)
- DELETE (suggest raw SQL with `DELETE FROM` or `ALTER TABLE ... DELETE`)
- Transactions
- Savepoints
- Soft deletes
- Foreign keys
- Associations (at DB level — JOINs in queries still work)
- Auto-increment IDs
- Schema builder
- Migrations

**Effort**: Medium (15-20 hours)
**Risk**: Low — we're honest about what works and what doesn't
**Verdict**: Recommended. Provides real value for analytics use cases without overpromising.

### Option C: No Support (Also Valid)

Don't support ClickHouse. Keep `❌` in the comparison tables. Document that Neat is an OLTP ORM and ClickHouse is an OLAP database — they serve different use cases.

**Effort**: Zero
**Risk**: None
**Verdict**: Valid if we don't see demand from users.

## Recommended Approach: Option B

### Phase 1: Driver Integration (5-6 hours)

1. Add `github.com/ClickHouse/clickhouse-go` as a dependency
2. Register `"clickhouse"` as a new driver in Neat's driver registry
3. Create a ClickHouse dialect that:
   - Uses ClickHouse-specific type mappings
   - Generates ClickHouse-compatible SQL for SELECT, INSERT, aggregates
   - Returns clear errors for unsupported operations (UPDATE, DELETE, transactions, etc.)
4. Add ClickHouse to `docs/driver-registration.html`

### Phase 2: Docker Compose and CI (2-3 hours)

Add a ClickHouse service to `docker-compose.yml`:

```yaml
  clickhouse:
    image: clickhouse/clickhouse-server:24.8
    container_name: neat-clickhouse-test
    ports:
      - "9000:9000"    # Native protocol
      - "8123:8123"    # HTTP interface
    ulimits:
      nofile:
        soft: 262144
        hard: 262144
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8123/ping"]
      interval: 10s
      timeout: 5s
      retries: 15
```

Add to GitHub Actions CI:

```yaml
      clickhouse:
        image: clickhouse/clickhouse-server:24.8
        ports:
          - "9000:9000"
          - "8123:8123"
        options: >-
          --health-cmd="wget --no-verbose --tries=1 --spider http://localhost:8123/ping"
          --health-interval=10s
          --health-timeout=5s
          --health-retries=15
```

### Phase 3: Integration Tests (8-10 hours)

Create `integration_tests/clickhouse/` with a **reduced test suite** (only supported features):

```
integration_tests/clickhouse/
├── helper.go                              # Connection config, setup, table creation
├── clickhouse_connection_test.go          # Basic connection test
├── clickhouse_find_test.go                # SELECT/Find/First/Get
├── clickhouse_query_create_test.go        # INSERT (batch insert focus)
├── clickhouse_query_count_test.go         # COUNT
├── clickhouse_query_where_test.go         # WHERE clauses
├── clickhouse_query_join_test.go          # JOIN operations
├── clickhouse_query_order_limit_offset_test.go  # ORDER BY, LIMIT, OFFSET
├── clickhouse_query_group_having_test.go  # GROUP BY, HAVING (ClickHouse strength)
├── clickhouse_query_aggregate_test.go     # SUM, AVG, MIN, MAX
├── clickhouse_query_distinct_test.go      # DISTINCT
├── clickhouse_query_select_test.go        # SELECT specific columns
├── clickhouse_query_pluck_test.go         # Pluck
├── clickhouse_query_value_test.go         # Value extraction
├── clickhouse_query_to_sql_test.go        # ToSql interface
├── clickhouse_raw_test.go                 # Raw SQL queries
├── clickhouse_dotted_column_test.go       # Dotted column references
├── clickhouse_unsupported_test.go         # Tests that verify clear errors for unsupported ops
└── clickhouse_query_chunk_test.go         # Chunk processing (for large analytics)
```

Note: No tests for UPDATE, DELETE, transactions, soft deletes, associations, schema builder, migrations, or foreign keys — these are explicitly unsupported.

### Phase 4: Helper File

`integration_tests/clickhouse/helper.go`:

```go
package clickhouse_test

import (
    "fmt"
    "testing"
    "time"

    "github.com/dracory/neat"
    "github.com/dracory/neat/database"
    "github.com/dracory/neat/integration_tests/common"
    _ "github.com/ClickHouse/clickhouse-go/v2"
)

// GetClickHouseConfig returns a ClickHouse connection config from environment variables.
func GetClickHouseConfig() neat.DBConfig {
    host := common.GetEnv("CLICKHOUSE_HOST", "127.0.0.1")
    port := common.GetEnvInt("CLICKHOUSE_PORT", 9000)
    database := common.GetEnv("CLICKHOUSE_DATABASE", "test")

    return neat.DBConfig{
        Default: "clickhouse",
        Connections: map[string]neat.ConnectionConfig{
            "clickhouse": {
                Driver:   "clickhouse",
                Host:     host,
                Port:     port,
                Database: database,
                // ClickHouse default user has no password
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

// SetupClickHouseTest sets up a ClickHouse connection and creates test tables.
func SetupClickHouseTest(t *testing.T) *database.DB {
    t.Helper()
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    config := GetClickHouseConfig()
    db, err := neat.New(config)
    if err != nil {
        t.Fatalf("Failed to connect to ClickHouse: %v", err)
    }

    // Create test tables using ClickHouse DDL
    // Note: ClickHouse requires MergeTree engine and ORDER BY clause
    createClickHouseTables(t, db)

    return db
}
```

Table creation for ClickHouse uses different DDL:

```sql
CREATE TABLE IF NOT EXISTS test_models (
    id UInt64,
    name String,
    email String,
    age UInt8,
    active UInt8,
    created_at DateTime,
    updated_at DateTime
) ENGINE = MergeTree()
ORDER BY id;
```

### Phase 5: Documentation (2-3 hours)

1. Update `docs/driver-registration.html` to list ClickHouse with clear caveats
2. Create `docs/clickhouse.html` documenting:
   - Supported features (read, insert, aggregates)
   - Unsupported features (update, delete, transactions, soft deletes, associations, schema builder)
   - Why ClickHouse is different (OLAP vs OLTP)
   - Recommended use cases (analytics, reporting)
   - How to use Neat with ClickHouse (connection config, query examples)
3. Update comparison tables from `❌` to `⚠️` (partial support) for ClickHouse
4. Update conclusion sections in comparison files

### Phase 6: Dialect Implementation (3-4 hours)

Create a ClickHouse dialect that:
- Maps Neat's column types to ClickHouse types (`uint` → `UInt64`, `string` → `String`, `time.Time` → `DateTime`)
- Generates ClickHouse-compatible SELECT SQL
- Generates ClickHouse-compatible INSERT SQL (with batch support)
- Returns `fmt.Errorf("ClickHouse does not support UPDATE operations; use raw SQL with ALTER TABLE ... UPDATE")` for UPDATE
- Returns `fmt.Errorf("ClickHouse does not support DELETE operations; use raw SQL with DELETE FROM or ALTER TABLE ... DELETE")` for DELETE
- Returns `fmt.Errorf("ClickHouse does not support transactions")` for transaction operations

## Known ClickHouse-Specific Test Considerations

1. **ID Generation**: ClickHouse doesn't have auto-increment. Tests must generate IDs explicitly (e.g., using `UUID()` or incrementing counters).

2. **No UPDATE/DELETE in tests**: Tests that rely on updating or deleting records (most association and soft delete tests) cannot be ported.

3. **Batch Insert**: ClickHouse excels at batch inserts. Tests should include batch insert scenarios.

4. **Eventual Consistency**: ClickHouse's MergeTree engine merges data parts asynchronously. Tests that count rows immediately after insert may see slightly different results. Use `OPTIMIZE TABLE ... FINAL` or wait for merges in tests.

5. **Data Types**: ClickHouse uses unsigned integers (`UInt8`, `UInt64`) by default. Tests must use appropriate types.

6. **No Boolean Type**: ClickHouse uses `UInt8` (0 or 1) for boolean values.

7. **DateTime Precision**: ClickHouse's `DateTime` has second precision; `DateTime64` supports sub-second precision.

8. **String vs FixedString**: ClickHouse has both `String` (variable length) and `FixedString(N)` types.

9. **Array and Map Types**: ClickHouse has native `Array(T)` and `Map(K,V)` types that don't have direct MySQL/PostgreSQL equivalents.

10. **ORDER BY is Mandatory**: MergeTree tables require an `ORDER BY` clause in the table definition. This affects schema builder tests.

## Implementation Plan

### Phase 1: Driver Integration (5-6 hours)
- Add `clickhouse-go` dependency
- Register ClickHouse driver
- Create ClickHouse dialect with type mappings
- Implement clear error messages for unsupported operations

### Phase 2: Docker Compose and CI (2-3 hours)
- Add ClickHouse service to `docker-compose.yml`
- Add ClickHouse service to GitHub Actions CI
- Add environment variables

### Phase 3: Integration Tests (8-10 hours)
- Create `integration_tests/clickhouse/helper.go`
- Create connection test (go/no-go)
- Create read/insert/aggregate tests
- Create unsupported-operation tests (verify clear errors)

### Phase 4: Dialect Refinement (3-4 hours)
- Handle ClickHouse-specific SQL generation
- Test edge cases (batch insert, large result sets)
- Handle type mapping edge cases

### Phase 5: Documentation (2-3 hours)
- Update driver-registration docs
- Create ClickHouse-specific documentation
- Update comparison tables to `⚠️` (partial support)
- Update conclusion sections

## Estimated Effort

- **Phase 1**: 5-6 hours (driver + dialect)
- **Phase 2**: 2-3 hours (Docker + CI)
- **Phase 3**: 8-10 hours (tests)
- **Phase 4**: 3-4 hours (dialect refinement)
- **Phase 5**: 2-3 hours (documentation)

**Total**: ~20-26 hours

## Risks

1. **Feature mismatch confusion**: Users may expect all Neat features to work with ClickHouse. Clear documentation and error messages are critical.

2. **`clickhouse-go` driver stability**: The driver is actively maintained by ClickHouse, but the `database/sql` interface is slower than the native interface. Performance-sensitive users should use the native interface directly.

3. **MergeTree async merges**: Tests may be flaky if they rely on immediate consistency. Use `OPTIMIZE TABLE ... FINAL` in test setup.

4. **Type mapping complexity**: ClickHouse's type system is rich (Array, Map, Tuple, Nullable, LowCardinality). Mapping these to Go types requires careful handling.

5. **Limited ORM value**: An ORM that can't UPDATE, DELETE, or use transactions is of limited value for typical web applications. The analytics-only use case must be clearly communicated.

## Success Criteria

1. `docker-compose up -d clickhouse` starts a healthy ClickHouse instance
2. `clickhouse_connection_test.go` passes — Neat can connect to ClickHouse
3. `go test -v ./integration_tests/clickhouse/...` passes all supported-feature tests
4. Unsupported operations return clear, helpful error messages
5. `docs/driver-registration.html` lists ClickHouse with clear caveats
6. `docs/clickhouse.html` documents supported/unsupported features and use cases
7. Comparison tables show `⚠️` (partial support) for ClickHouse, not `✅`

## Decision Required

**Should Neat support ClickHouse as a "read-only analytics" target (Option B), or should we explicitly not support it (Option C)?**

Arguments for Option B:
- GORM supports it, so users comparing ORMs may expect it
- Analytics queries are a valid use case
- The query builder and aggregates work well for reporting
- It differentiates Neat from ORMs that only support OLTP databases

Arguments for Option C:
- ClickHouse is fundamentally different from OLTP databases
- Most ORM features don't apply
- The effort (20-26 hours) is significant for a partial feature set
- Users who need ClickHouse analytics may be better served by direct SQL or the native `clickhouse-go` API
- Supporting it may create false expectations about feature parity

**Recommendation**: Start with Option C (no support) unless there is user demand. If demand exists, implement Option B (analytics mode) with very clear documentation about the reduced feature set. Do not attempt Option A (full support).

## References

- ClickHouse Documentation: https://clickhouse.com/docs
- ClickHouse Go Driver: https://github.com/ClickHouse/clickhouse-go
- ClickHouse MySQL Interface: https://clickhouse.com/docs/interfaces/mysql
- ClickHouse MySQL Driver Issue: https://github.com/ClickHouse/ClickHouse/issues/64071
- GORM ClickHouse Driver: https://github.com/go-gorm/clickhouse
- Neat Driver Registration: `docs/driver-registration.html`
