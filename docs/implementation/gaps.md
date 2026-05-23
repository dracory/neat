# Neat Implementation Gaps

**Date**: May 2026  
**Purpose**: Catalogue remaining implementation gaps in the neat ORM project.

---

## ORM Feature Gaps

### GAP-09: `Having()` does not support `func(Query)Query` callback subqueries

**Status**: ❌ Open  
**Root cause**: `Having` only accepts string + args. There is no code path for a callback that builds a subquery.  
**Failing tests**:
- `integration_tests/sqlite/sqlite_query_group_having_test.go` — `Having with subquery callback`, `Having with subquery in args`

**Implementation hint**: Add a callback overload check in the `Having` method, similar to how `Where` handles closures.

---

## Test Coverage Gaps

### Root Level Tests

| File | Priority | Notes |
|------|----------|-------|
| db_context_test.go | High | Tests database context handling |
| db_pool_test.go | High | Tests connection pool configuration |
| db_ssl_test.go | High | Tests SSL/TLS connection configuration |

### Database/DB Tests

| File | Priority | Notes |
|------|----------|-------|
| database/db/dsn_test.go | High | Tests DSN parsing and generation |

### Database/Schema Tests

| File | Priority | Notes |
|------|----------|-------|
| database/schema/index_test.go | Medium | Tests index creation and management |

### Database/Query Tests

| File | Priority | Notes |
|------|----------|-------|
| database/query/query_test.go | High | Tests full query builder functionality |
| database/query/clause_test.go | High | Tests SQL clause generation |
| database/query/builder_test.go | High | Tests query builder internals |

### Database/Query Unit Tests (Specific)

| File | Priority | Notes |
|------|----------|-------|
| database/query/where_exists_test.go | Medium | Tests for `WhereExists` subquery generation |
| database/orm/buildquery_replica_test.go | Low | Tests that `buildQuery` wires replicas correctly (requires real DB) |

### Integration Tests

| Directory | Priority | Notes |
|-----------|----------|-------|
| integration_tests/sqlserver/*_test.go | Medium | ~8 SQL Server integration test files |
| integration_tests/turso/*_test.go | Medium | ~3 Turso integration test files |

### Integration Test Coverage Gaps

| Area | Priority | Notes |
|---|---|---|
| Read/write routing with real dual DB | Medium | No test that reads go to replica and writes go to primary using *different* hosts |
| `InsertGetId` Postgres `RETURNING id` | Medium | No integration test asserting the returned ID is non-zero |
| `Query.Connection(name)` switching | Medium | Integration test needed (connection test exercises `Database.Connection`, not `Query.Connection`) |
| `SlowThreshold` warning in integration | Low | No test that triggers and captures a slow-query log entry against a real DB |
| Transaction hooks (`BeforeCommit` etc.) | Medium | Integration test needed to confirm callbacks fire with a real DB |

---

## Implementation Priority

### High Priority

1. **Add root level tests** (db_context_test.go, db_pool_test.go, db_ssl_test.go)
   - These tests validate core configuration and connection handling
   - Can be implemented without database connections

2. **Add database/db/dsn_test.go**
   - Tests DSN parsing logic
   - Can be implemented without database connections

3. **Add database/query/query_test.go**
   - Tests full query builder functionality
   - Can be implemented with mocks

4. **Fix GAP-09: Having callback subqueries**
   - Add callback support to Having method
   - Required for full GROUP BY/HAVING functionality

### Medium Priority

1. **Add database/schema/index_test.go**
   - Tests index management in schema builder
   - Can be implemented without database connections

2. **Add database/query/clause_test.go and builder_test.go**
   - Tests internal query building components
   - Can be implemented with mocks

3. **Add database/query/where_exists_test.go**
   - Tests for WhereExists subquery generation
   - Can be implemented with mocks

4. **Integration tests for SQL Server and Turso**
   - Require database setup and configuration
   - Follow the pattern from existing integration tests

5. **Integration test coverage gaps**
   - Read/write routing with real dual DB
   - Query.Connection switching
   - Transaction hooks

### Low Priority

1. **database/orm/buildquery_replica_test.go**
   - Requires real database available in test environment
   - Tests replica connection wiring

2. **SlowThreshold warning integration test**
   - Tests slow-query logging against real database

---

## Dependencies

Some tests have dependencies on other components:

- Integration tests require database services (MySQL, PostgreSQL, SQLite, SQL Server, Turso)
- Query builder tests require mock database connections
- Schema tests require schema builder to be fully functional
- buildquery_replica_test.go requires real database connections

---

## Notes

- **Completed items removed**: All ORM feature gaps (GAP-01 through GAP-08, GAP-10 through GAP-24) have been completed and removed from this document.
- **Test files marked as EXISTS**: Test files that already exist (config_test.go, database/db/config_builder_test.go, database/schema/blueprint_test.go, database/schema/column_test.go, database/query/to_sql_test.go, database/query/query_bench_test.go, etc.) are not listed as gaps.
- **Integration tests**: MySQL, PostgreSQL, and SQLite integration tests are complete. SQL Server and Turso integration tests are still missing.
