# Neat Implementation Gaps

**Date**: May 23, 2026
**Last reviewed**: May 23, 2026
**Purpose**: Catalogue remaining implementation gaps in the neat ORM project.

---

## ORM Feature Gaps

✅ **All ORM feature gaps have been completed.** The `Having()` method now supports `func(Query)Query` callback subqueries (implemented in database/query/query.go lines 1006-1027).

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
| `SlowThreshold` warning in integration | Low | No test that triggers and captures a slow-query log entry against a real DB |

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
- **Recently completed test coverage** (commit 36e2056, May 23, 2026):
  - Added `database/db_test.go` with comprehensive Database method tests
  - Added `database/query/connection_switch_test.go` for Query.Connection switching
  - Added `database/query/insert_get_id_test.go` for InsertGetId functionality
  - Added `database/query/transaction_hooks_test.go` for transaction hooks
  - Added `database/query/query_routing_test.go` for read/write routing
