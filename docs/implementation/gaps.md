# Neat ORM Implementation Gaps - Executable Plan

**Date**: May 25, 2026
**Last reviewed**: May 30, 2026
**Purpose**: Comprehensive executable plan to eliminate all implementation gaps in the neat ORM project.

---

## Executive Summary

This document provides a complete, step-by-step plan to bring the neat ORM to production-ready status with zero gaps. All items are prioritized and organized for sequential execution.

**Current Status:**
- ✅ Core ORM features: Complete
- ✅ SQLite integration tests: Complete (30+ tests)
- ✅ Turso integration tests: Complete (23 tests)
- ✅ Factory pattern: Implemented
- ✅ Migration system: Implemented
- ✅ Seeder system: Implemented
- ✅ MySQL integration tests: 46/46 files enabled
- ⚠️ PostgreSQL integration tests: 43/46 files disabled
- ❌ SQL Server integration tests: Not created

---

## Phase 1: Unit Test Coverage Gaps (HIGH PRIORITY)

These tests validate core functionality without requiring database connections.

### 1.1 Database/DB Tests

#### 1.1.1 database/db/dsn_test.go

**Status**: ❌ Missing
**Priority**: HIGH
**Purpose**: Test DSN parsing and generation

**Test cases needed**:
- Parse MySQL DSN
- Parse PostgreSQL DSN
- Parse SQLite DSN
- Parse SQL Server DSN
- Parse Turso DSN
- DSN with query parameters
- DSN with special characters in password
- DSN validation (max length 4096)
- DSN redaction for logging
- Invalid DSN error handling

**Files to reference**:
- `database/db/dsn.go` - Implementation to test
- `database/db/config_builder_test.go` - Similar test patterns

**Estimated effort**: 1 day

---

### 1.2 Database/Query Tests

#### 1.2.1 database/query/query_test.go

**Status**: ✅ Complete
**Priority**: HIGH
**Purpose**: Comprehensive query builder unit tests

**Test cases needed**:
- All Where variants (Where, OrWhere, WhereIn, WhereNotIn, WhereBetween, etc.)
- All Join variants (Join, LeftJoin, RightJoin, CrossJoin)
- Select with columns and aliases
- GroupBy and Having
- OrderBy with multiple columns and directions
- Limit and Offset
- Distinct
- Subqueries in Select, Where, From
- Raw expressions
- Query cloning
- Method chaining

**Note**: Can use mocks or ToSql() for testing without database

**Estimated effort**: 2 days

#### 1.2.2 database/query/clause_test.go

**Status**: ✅ Complete
**Priority**: HIGH
**Purpose**: Test SQL clause generation

**Test cases needed**:
- WHERE clause generation
- JOIN clause generation
- ORDER BY clause generation
- GROUP BY clause generation
- HAVING clause generation
- LIMIT/OFFSET clause generation
- Clause combination and ordering

**Estimated effort**: 1 day

#### 1.2.3 database/query/builder_test.go

**Status**: ✅ Complete
**Priority**: HIGH
**Purpose**: Test query builder internals

**Test cases needed**:
- Binding parameter handling
- Placeholder generation (?, $1, @p1)
- Table name wrapping
- Column name wrapping
- Identifier escaping
- Query compilation

**Estimated effort**: 1 day

#### 1.2.4 database/query/where_exists_test.go

**Status**: ✅ Complete
**Priority**: MEDIUM
**Purpose**: Test WhereExists subquery generation

**Test cases needed**:
- WhereExists with subquery
- WhereNotExists with subquery
- OrWhereExists
- OrWhereNotExists
- Nested exists clauses

**Estimated effort**: 0.5 days

---

### 1.3 Database/Schema Tests

#### 1.3.1 database/schema/index_test.go

**Status**: ❌ Missing
**Priority**: MEDIUM
**Purpose**: Test index creation and management

**Test cases needed**:
- Index creation
- Unique index creation
- Composite index creation
- Index naming
- Index dropping
- Index listing

**Estimated effort**: 0.5 days

---

### 1.4 Database/ORM Tests

#### 1.4.1 database/orm/buildquery_replica_test.go

**Status**: ❌ Missing
**Priority**: LOW
**Purpose**: Test that buildQuery wires replicas correctly

**Note**: Requires real database connections (read + write)

**Test cases needed**:
- Read queries use read connection
- Write queries use write connection
- Connection switching
- Fallback when replica unavailable

**Estimated effort**: 1 day (requires test infrastructure)

---

## Phase 2: Integration Test Enablement (MEDIUM PRIORITY)

### 2.1 Enable PostgreSQL Integration Tests

**Status**: ⚠️ 43/46 files disabled
**Priority**: MEDIUM
**Files**: `integration_tests/postgres/*_test.go`

**Problem**:
- 43 PostgreSQL tests marked with `//go:build disabled`
- 3 tests already enabled with `//go:build integration` (helper.go, postgres_query_belongs_to_test.go, driver_registration_test.go)
- Tests exist but are not running in CI/CD

**Steps**:
1. Set up PostgreSQL test database in CI/CD
2. Update GitHub Actions workflow to include PostgreSQL service
3. Change `//go:build disabled` to `//go:build integration` in all files
4. Run tests and fix any failures
5. Document PostgreSQL test setup in integration_tests/README.md

**Files affected**: 43 test files

**Estimated effort**: 2-3 days

---

### 2.2 Enable MySQL Integration Tests

**Status**: ✅ 43/46 files enabled
**Priority**: LOW
**Files**: `integration_tests/mysql/*_test.go`

**Problem**:
- 43 MySQL tests marked with `//go:build integration` (enabled)
- 3 tests still disabled with `//go:build disabled`

**Steps**:
1. Identify why 3 tests are still disabled (check for failures)
2. Fix any implementation issues causing test failures
3. Change `//go:build disabled` to `//go:build integration`
4. Ensure MySQL service is configured in CI/CD
5. Run tests and verify all pass

**Files affected**: 3 test files

**Estimated effort**: 0.5 days

---

### 2.3 Create SQL Server Integration Tests

**Status**: ❌ Not created
**Priority**: MEDIUM

**Steps**:
1. Create `integration_tests/sqlserver/` directory
2. Create helper.go with SQL Server test setup
3. Port tests from MySQL/PostgreSQL patterns
4. Create ~40 test files covering:
   - Query operations (CRUD, aggregates, joins, etc.)
   - Schema operations (create, alter, drop tables)
   - Transactions
   - Soft deletes
5. Add SQL Server service to CI/CD
6. Document SQL Server test setup

**Estimated effort**: 3-4 days

---


## Phase 3: Advanced Integration Test Coverage (LOW PRIORITY)

### 3.1 Read/Write Replica Routing Test

**Status**: ❌ Missing
**Priority**: MEDIUM

**Purpose**: Test that reads go to replica and writes go to primary using different hosts

**Current state**: 
- Unit test exists: `database/query/query_routing_test.go`
- No integration test with real dual database setup

**Steps**:
1. Create test with two real database instances
2. Configure primary and replica in test config
3. Verify SELECT queries use replica connection
4. Verify INSERT/UPDATE/DELETE use primary connection
5. Test connection failover scenarios

**File to create**: `integration_tests/common/replica_routing_test.go`

**Estimated effort**: 1 day

---

### 3.2 InsertGetId PostgreSQL RETURNING Test

**Status**: ❌ Missing
**Priority**: MEDIUM

**Purpose**: Integration test asserting returned ID is non-zero for PostgreSQL

**Current state**:
- Unit test exists: `database/query/insert_get_id_test.go`
- No PostgreSQL-specific integration test

**Steps**:
1. Add test to `integration_tests/postgres/postgres_query_create_test.go`
2. Test InsertGetId returns correct ID
3. Verify PostgreSQL RETURNING clause is used
4. Test with serial and bigserial columns

**Estimated effort**: 0.5 days

---

### 3.3 SlowThreshold Warning Integration Test

**Status**: ❌ Missing
**Priority**: LOW

**Purpose**: Test that slow query logging triggers correctly

**Steps**:
1. Configure database with low SlowThreshold (e.g., 1ms)
2. Execute query that exceeds threshold
3. Capture log output
4. Verify slow query warning was logged
5. Test with different log levels

**File to create**: `integration_tests/common/slow_query_test.go`

**Estimated effort**: 0.5 days

---

## Phase 4: Documentation Gaps (MEDIUM PRIORITY)

### 4.1 Update Placeholder Documentation

**Status**: ⚠️ Marked as placeholders
**Priority**: MEDIUM

**Files to update**:
1. `docs/query-builder.md` - Expand with comprehensive examples
2. `docs/schema-builder.md` - Add all column types and modifiers
3. `docs/migrations.md` - Update based on implementation decision (Phase 1.3)

**Steps for each**:
1. Add comprehensive API reference
2. Add code examples for all major features
3. Add best practices section
4. Add troubleshooting section
5. Remove "placeholder" markers

**Estimated effort**: 2 days total

---

### 4.2 Create Missing Documentation

**Status**: ❌ Missing
**Priority**: MEDIUM

**Documents to create**:

#### 4.2.1 docs/factory.md
- Factory pattern usage (if implemented)
- Defining factories
- Using factories in tests
- Factory relationships

#### 4.2.2 docs/seeder.md
- Seeder usage (if implemented)
- Creating seeders
- Running seeders
- Seeder dependencies

#### 4.2.3 docs/testing.md
- Testing guide for contributors
- Running unit tests
- Running integration tests
- Writing new tests
- Test database setup

#### 4.2.4 docs/performance.md
- Connection pooling best practices
- Query optimization tips
- Eager loading vs lazy loading
- Batch operations
- Benchmarking results

#### 4.2.5 docs/api-reference.md
- Complete API reference
- All methods with signatures
- Parameter descriptions
- Return values
- Examples for each method

**Estimated effort**: 3-4 days total

---

### 4.3 Update README.md

**Status**: ⚠️ Contains inaccuracies
**Priority**: HIGH

**Issues to fix**:
1. Remove Migration from features (or mark as not implemented)
2. Remove Seeder from features (or mark as not implemented)
5. Add "Roadmap" section with planned features
6. Add "Contributing" guide link
7. Add badges (build status, coverage, version, license)

**Estimated effort**: 0.5 days

---

## Phase 5: Code Quality Improvements (LOW PRIORITY)

### 5.1 Resolve TODO Comments

**Status**: ⚠️ 13 TODO comments found
**Priority**: LOW

**TODOs to resolve**:

1. **database/schema/grammars/sqlserver.go:161**
   - `TODO Add change logic` for DropDefaultConstraint
   - Implement or document limitation

2. **database/schema/grammars/sqlite.go:126**
   - `TODO check Sqlite 3.35` for DropColumn
   - Verify SQLite version support and update

3. **integration_tests/postgres/postgres_query_lock_test.go:10**
   - `TODO: package doesn't exist in neat`
   - Remove comment or fix reference

4. **support/str/str_test.go:2** - Additional TODOs in test files
5. **database/query/query_advanced_test.go:20** - Additional TODOs in test files
6. **database/query/query_bench_test.go:5** - Additional TODOs in test files
7. **database/query/to_sql_test.go:15** - Additional TODOs in test files
8. **database/query/query_where_test.go:79** - Additional TODOs in test files
9. **database/query/query_raw_test.go:3** - Additional TODOs in test files
10. **database/query/query_aggregate_test.go:7** - Additional TODOs in test files
11. **database/query/query_builder_test.go:29** - Additional TODOs in test files
12. **CHANGELOG.md:1** - Additional TODOs in documentation

**Estimated effort**: 1.5 days

---

### 5.2 Remove Unused Dependencies

**Status**: ⚠️ Unused dependencies in go.mod
**Priority**: LOW

**From CHANGELOG.md**:
> Some dependencies (github.com/tursodatabase/libsql-client-go, github.com/antlr4-go/antlr/v4, github.com/coder/websocket) are not currently used

**Steps**:
1. Verify dependencies are truly unused
2. Remove from go.mod if not needed for Turso implementation
3. Run `go mod tidy`
4. Update CHANGELOG.md

**Estimated effort**: 0.5 days

---

### 5.3 Add Code Coverage Reporting

**Status**: ❌ Not configured
**Priority**: LOW

**Steps**:
1. Add coverage reporting to GitHub Actions
2. Integrate with Codecov or Coveralls
3. Add coverage badge to README
4. Set coverage thresholds
5. Generate coverage reports locally

**Estimated effort**: 0.5 days

---

## Phase 6: CI/CD Improvements (LOW PRIORITY)

### 6.1 Enhance GitHub Actions Workflows

**Status**: ⚠️ Basic workflows exist
**Priority**: LOW

**Current workflows**:
- `.github/workflows/tests.yml` - Unit tests
- `.github/workflows/integration-tests.yml` - Integration tests

**Improvements needed**:
1. Add PostgreSQL service to integration tests
2. Add MySQL service to integration tests
3. Add SQL Server service to integration tests
4. Add code coverage reporting
5. Add linting (golangci-lint)
6. Add security scanning (gosec)
7. Add dependency vulnerability scanning
8. Add build matrix (multiple Go versions)
9. Add caching for faster builds

**Estimated effort**: 1-2 days

---

## Execution Plan Summary

### Recommended Execution Order

**Week 1-2: Critical Decisions & Fixes**
1. Make decisions on Migration, Seeder (implement or remove)
2. Update README to reflect actual features

**Week 3-4: Unit Test Coverage**
3. Add root level tests (Phase 2.1)
4. Add database/db/dsn_test.go (Phase 2.2)
5. Add database/query tests (Phase 2.3)
6. Add database/schema tests (Phase 2.4)

**Week 5-6: Integration Test Enablement**
7. Enable PostgreSQL integration tests (Phase 3.1)
8. Enable MySQL integration tests (Phase 3.2)
9. Enable SQLite disabled tests (Phase 3.3)
10. Create SQL Server integration tests (Phase 3.4)

**Week 7-8: Documentation & Polish**
11. Update placeholder documentation (Phase 5.1)
12. Create missing documentation (Phase 5.2)
13. Resolve TODO comments (Phase 6.1)
14. CI/CD improvements (Phase 7.1)

**Optional (if time permits)**:
- Advanced integration tests (Phase 4)
- Code coverage reporting (Phase 6.3)

---

## Success Criteria

The project will have **ZERO GAPS** when:

- [ ] All stub implementations are either completed or removed
- [ ] All disabled integration tests are enabled and passing
- [ ] All missing unit tests are created and passing
- [ ] All input validation is implemented
- [ ] All documentation is accurate and complete
- [ ] All TODO comments are resolved
- [ ] README accurately reflects implemented features
- [ ] CI/CD runs all tests successfully
- [ ] Code coverage is measured and reported
- [ ] No "not implemented" errors exist in codebase

---

## Tracking Progress

Use this checklist to track completion:

### Phase 1: Unit Tests
- [ ] 1.1.1 database/db/dsn_test.go
- [ ] 1.2.1 database/query/query_test.go
- [ ] 1.2.2 database/query/clause_test.go
- [ ] 1.2.3 database/query/builder_test.go
- [ ] 1.2.4 database/query/where_exists_test.go
- [ ] 1.3.1 database/schema/index_test.go
- [ ] 1.4.1 database/orm/buildquery_replica_test.go

### Phase 2: Integration Tests
- [ ] 2.1 Enable PostgreSQL tests (43 files)
- [ ] 2.2 Enable MySQL tests (3 files remaining)
- [ ] 2.3 Create SQL Server tests (~40 files)

### Phase 3: Advanced Integration
- [ ] 3.1 Read/Write Replica Routing Test
- [ ] 3.2 InsertGetId PostgreSQL Test
- [ ] 3.3 SlowThreshold Warning Test

### Phase 4: Documentation
- [ ] 4.1 Update placeholder docs (3 files)
- [ ] 4.2 Create missing docs (5 files)
- [ ] 4.3 Update README.md

### Phase 5: Code Quality
- [ ] 5.1 Resolve TODO comments (13 items)
- [ ] 5.2 Remove unused dependencies
- [ ] 5.3 Add code coverage reporting

### Phase 6: CI/CD
- [ ] 6.1 Enhance GitHub Actions workflows

---

## Estimated Total Effort

- **Phase 1**: 7.5 days
- **Phase 2**: 5.5-9.5 days
- **Phase 3**: 2 days
- **Phase 4**: 5.5 days
- **Phase 5**: 2 days
- **Phase 6**: 1.5 days

**Total**: 24-28 days (5-6 weeks for one developer)

With 2-3 developers working in parallel: **3-4 weeks to zero gaps**

---

## Notes

- Integration test enablement may reveal additional bugs requiring fixes
- Documentation effort can be parallelized with implementation work
- CI/CD improvements can be done incrementally
- Code coverage and quality improvements are ongoing

**Last Updated**: May 30, 2026
