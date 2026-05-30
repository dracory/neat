# Neat ORM Implementation Gaps - Executable Plan

**Date**: May 25, 2026
**Last reviewed**: May 30, 2026
**Last updated**: May 30, 2026
**Purpose**: Comprehensive executable plan to eliminate all implementation gaps in the neat ORM project.

---

## Executive Summary

This document provides a complete, step-by-step plan to bring the neat ORM to production-ready status with zero gaps. All items are prioritized and organized for sequential execution.

**Current Status:**
- ⚠️ PostgreSQL integration tests: 35 files enabled, 11 files skipped due to unimplemented features
- ❌ SQL Server integration tests: Not created

---

## Phase 1: Integration Test Enablement (MEDIUM PRIORITY)

### 1.1 Enable PostgreSQL Integration Tests

**Status**: ⚠️ Partially Complete (35/46 files passing, 11 skipped due to gaps)
**Priority**: MEDIUM
**Files**: `integration_tests/postgres/*_test.go`

**Skipped tests (gaps to address)**:
- ~~postgres_query_join_test.go - PostgreSQL custom type conflicts~~ (FIXED)
- ~~postgres_query_json_test.go - MySQL/SQLite JSON syntax incompatibility~~ (FIXED - implemented PostgreSQL JSONB operators)
- ~~postgres_query_omit_test.go - Soft-delete filter incompatibility~~ (FIXED)
- postgres_query_paginate_test.go - Soft-delete filter incompatibility with count query (SKIPPED)
- ~~postgres_query_select_test.go (specific columns, subqueries) - Soft-delete filter and subquery parameter numbering~~ (FIXED - subquery test skipped due to parameterized subqueries not supported)
- ~~postgres_query_to_sql_test.go (Count, Update, RawSql, Value) - SQL format variations~~ (FIXED)
- ~~postgres_query_update_or_insert_test.go (struct tests) - Soft-delete filter~~ (FIXED)
- ~~postgres_query_value_test.go (ToSql) - SQL format variations~~ (FIXED)
- ~~postgres_query_lock_test.go (SharedLock, ConcurrentAccess) - PostgreSQL FOR SHARE syntax~~ (FIXED)
- postgres_query_order_limit_offset_test.go (negative limit) - PostgreSQL doesn't allow negative LIMIT
- postgres_query_group_having_test.go (subquery tests) - Subquery parameter numbering not implemented
- postgres_query_increment_decrement_test.go (decrement ID) - Invalid operation on auto-increment
- postgres_schema_* tests (9 files) - Schema builder not implemented for PostgreSQL

**Remaining steps**:
1. Set up PostgreSQL test database in CI/CD
2. Update GitHub Actions workflow to include PostgreSQL service
3. Document PostgreSQL test setup in integration_tests/README.md

---

### 1.2 Create SQL Server Integration Tests

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


## Phase 2: Advanced Integration Test Coverage (LOW PRIORITY)

### 2.1 Read/Write Replica Routing Test

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

### 2.2 InsertGetId PostgreSQL RETURNING Test

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

### 2.3 SlowThreshold Warning Integration Test

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

## Phase 3: Code Quality Improvements (LOW PRIORITY)

### 3.1 Resolve TODO Comments

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

### 3.2 Remove Unused Dependencies

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

### 3.3 Add Code Coverage Reporting

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

## Phase 4: CI/CD Improvements (LOW PRIORITY)

### 4.1 Enhance GitHub Actions Workflows

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

## Phase 5: Advanced Features (LOW PRIORITY)

### 5.1 Spatial Data Type Support

**Status**: ❌ Not implemented
**Priority**: LOW

**Purpose**: Add support for MySQL/PostgreSQL spatial data types (GEOMETRY, POINT, LINESTRING, POLYGON, etc.) to match Laravel's spatial capabilities.

**Current state**:
- Test exists: `integration_tests/mysql/mysql_query_spatial_test.go` (skipped)
- No spatial type definitions in ORM
- No WKT/WKB format support
- No spatial function support

**Steps**:
1. Add spatial type definitions in models (Geometry, Point, LineString, Polygon, etc.)
2. Implement WKT (Well-Known Text) parsing and serialization
3. Implement WKB (Well-Known Binary) parsing and serialization
4. Add spatial column types to schema builder
5. Add spatial query functions (ST_GeomFromText, ST_AsText, ST_Distance, etc.)
6. Update Create() method to handle spatial data types
7. Add spatial query scopes (WhereDistance, etc.)
8. Write integration tests for spatial operations

**Estimated effort**: 3-4 days

---

## Execution Plan Summary

### Recommended Execution Order

**Week 1: Integration Test Enablement**
1. Enable PostgreSQL integration tests (Phase 1.1)
2. Create SQL Server integration tests (Phase 1.2)

**Week 2: Advanced Integration & Code Quality**
3. Read/Write Replica Routing Test (Phase 2.1)
4. InsertGetId PostgreSQL Test (Phase 2.2)
5. SlowThreshold Warning Test (Phase 2.3)
6. Resolve TODO comments (Phase 3.1)
7. Remove unused dependencies (Phase 3.2)

**Week 3: CI/CD & Polish**
8. Add code coverage reporting (Phase 3.3)
9. Enhance GitHub Actions workflows (Phase 4.1)

**Optional (if time permits)**:
- Additional CI/CD improvements

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

### Phase 1: Integration Tests
- [ ] 1.1 Enable PostgreSQL tests (35/46 files passing, 11 skipped due to gaps)
- [ ] 1.2 Create SQL Server tests (~40 files)

### Phase 2: Advanced Integration
- [ ] 2.1 Read/Write Replica Routing Test
- [ ] 2.2 InsertGetId PostgreSQL Test
- [ ] 2.3 SlowThreshold Warning Test

### Phase 3: Code Quality
- [ ] 3.1 Resolve TODO comments (13 items)
- [ ] 3.2 Remove unused dependencies
- [ ] 3.3 Add code coverage reporting

### Phase 4: CI/CD
- [ ] 4.1 Enhance GitHub Actions workflows

### Phase 5: Advanced Features
- [ ] 5.1 Spatial Data Type Support

---

## Estimated Total Effort

- **Phase 1**: 5-7 days
- **Phase 2**: 2 days
- **Phase 3**: 2 days
- **Phase 4**: 1.5 days
- **Phase 5**: 3-4 days

**Total**: 13.5-16.5 days (2.5-3.5 weeks for one developer)

With 2-3 developers working in parallel: **1-1.5 weeks to zero gaps**

---

## Notes

- Integration test enablement may reveal additional bugs requiring fixes
- Documentation effort can be parallelized with implementation work
- CI/CD improvements can be done incrementally
- Code coverage and quality improvements are ongoing

**Last Updated**: May 30, 2026
