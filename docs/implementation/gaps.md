# Neat ORM Implementation Gaps - Executable Plan

**Date**: May 25, 2026
**Last reviewed**: May 25, 2026
**Purpose**: Comprehensive executable plan to eliminate all implementation gaps in the neat ORM project.

---

## Executive Summary

This document provides a complete, step-by-step plan to bring the neat ORM to production-ready status with zero gaps. All items are prioritized and organized for sequential execution.

**Current Status:**
- ✅ Core ORM features: Complete
- ✅ SQLite integration tests: Complete (30+ tests)
- ⚠️ MySQL integration tests: 39/46 files disabled
- ⚠️ PostgreSQL integration tests: 46/46 files disabled
- ❌ SQL Server integration tests: Not created
- ❌ Turso integration tests: Not created
- ❌ Factory pattern: Stub implementation only
- ❌ Migration system: Contracts only, no implementation
- ❌ Seeder system: Contracts only, no implementation

---

## Phase 1: Critical Implementation Gaps (MUST FIX)

These are incomplete features that will cause runtime errors if used.

### 1.1 Factory Pattern Implementation

**Status**: ❌ Returns "not implemented" errors
**Priority**: CRITICAL
**Files**: `database/orm/factory.go`

**Problem**: 
- `Factory.Create()`, `Factory.CreateQuietly()`, `Factory.Make()` all return errors
- `Orm.Factory()` returns nil (line 253-255 in database/orm/orm.go)
- Contract exists but no working implementation

**Decision Required**: 
- [ ] **Option A**: Implement full factory pattern (similar to Laravel factories)
- [ ] **Option B**: Remove factory from contracts and mark as future feature
- [ ] **Option C**: Keep stub but document as "not yet implemented" in README

**If implementing (Option A), steps:**
1. Design factory definition system (how users define factories)
2. Implement attribute merging and overrides
3. Implement sequence generation for unique values
4. Implement relationship factory support
5. Add factory tests
6. Add factory documentation and examples

**Estimated effort**: 3-5 days for full implementation

---

### 1.2 Turso Driver Implementation

**Status**: ❌ Returns "not yet implemented" error
**Priority**: HIGH
**Files**: `database/driver/turso.go`

**Problem**:
- `Turso.Open()` returns error (line 20-22)
- Listed as supported in README but doesn't work
- Commented out libsql import

**Decision Required**:
- [ ] **Option A**: Implement Turso driver using libsql-client-go
- [ ] **Option B**: Remove Turso from supported databases list
- [ ] **Option C**: Mark as "experimental" or "coming soon"

**If implementing (Option A), steps:**
1. Uncomment and configure libsql-client-go dependency
2. Implement Turso.Open() with proper DSN parsing
3. Test connection pooling with Turso
4. Create Turso integration test suite
5. Update documentation

**Estimated effort**: 2-3 days

---

### 1.3 Migration System Implementation

**Status**: ❌ Contracts exist, no implementation
**Priority**: HIGH
**Files**: `contracts/migration/`, `contracts/database/migration/`

**Problem**:
- Migration contracts defined but no implementation exists
- Documentation in `docs/migrations.md` describes non-existent features
- Examples in `examples/migrations/` may reference unimplemented code

**Decision Required**:
- [ ] **Option A**: Implement full migration system
- [ ] **Option B**: Remove migration contracts and mark as future feature
- [ ] **Option C**: Use schema builder as migration alternative (document this)

**If implementing (Option A), steps:**
1. Create migration repository (tracks applied migrations)
2. Create migration runner (applies/rolls back migrations)
3. Implement migration file generation
4. Add migration status tracking
5. Create migration tests
6. Update documentation to match implementation

**Estimated effort**: 5-7 days for full implementation

---

### 1.4 Seeder System Implementation

**Status**: ❌ Contracts exist, no implementation
**Priority**: MEDIUM
**Files**: `contracts/seeder/`, `contracts/database/seeder/`

**Problem**:
- Seeder contracts defined but no implementation exists
- No seeder facade or runner

**Decision Required**:
- [ ] **Option A**: Implement seeder system
- [ ] **Option B**: Remove seeder contracts and mark as future feature
- [ ] **Option C**: Document manual seeding patterns as alternative

**If implementing (Option A), steps:**
1. Create seeder registry
2. Implement seeder runner with dependency ordering
3. Add CallOnce tracking (prevent duplicate runs)
4. Create seeder tests
5. Add seeder documentation and examples

**Estimated effort**: 2-3 days

---

## Phase 2: Input Validation & Security Gaps (HIGH PRIORITY)

These gaps could lead to SQL injection or runtime panics.

### 2.1 Column Name Validation

**Status**: ❌ Not implemented
**Priority**: HIGH
**Evidence**: Tests skipped in `sqlite_query_aggregate_test.go` lines 196-220

**Problem**:
- Aggregate functions (Sum, Avg, Min, Max) don't validate column names
- SQL injection possible via column names
- Invalid column names cause database errors instead of early validation

**Implementation steps:**
1. Add column name validation function (alphanumeric + underscore + dot for table.column)
2. Apply validation in all aggregate methods
3. Apply validation in Select, OrderBy, GroupBy
4. Add validation tests
5. Enable skipped validation tests

**Files to modify**:
- `database/query/query.go` - Add validation to Sum, Avg, Min, Max, Count
- Add new file: `database/query/validation.go`
- Enable tests in `integration_tests/sqlite/sqlite_query_aggregate_test.go`

**Estimated effort**: 1 day

---

### 2.2 Nil Destination Validation

**Status**: ❌ Not implemented
**Priority**: HIGH
**Evidence**: Test skipped in `sqlite_query_aggregate_test.go` line 218

**Problem**:
- Methods don't check for nil destination pointers
- Causes panic instead of returning error

**Implementation steps:**
1. Add nil checks to all methods that accept destination pointers
2. Return descriptive error for nil destinations
3. Enable skipped nil validation tests

**Files to modify**:
- `database/query/query.go` - Add nil checks to Find, First, Get, Sum, Avg, etc.

**Estimated effort**: 0.5 days

---

### 2.3 COUNT(DISTINCT column) Support

**Status**: ❌ Not implemented
**Priority**: MEDIUM
**Evidence**: Test skipped in `sqlite_query_distinct_test.go` line 81

**Problem**:
- `Distinct().Count()` doesn't generate `COUNT(DISTINCT column)`
- Currently generates incorrect SQL

**Implementation steps:**
1. Detect when Distinct is used with Count
2. Generate `COUNT(DISTINCT column)` instead of `SELECT DISTINCT ... COUNT(*)`
3. Enable skipped test

**Files to modify**:
- `database/query/query.go` - Modify Count() method
- `database/query/builder.go` - Add distinct count logic

**Estimated effort**: 0.5 days

---

## Phase 3: Unit Test Coverage Gaps (HIGH PRIORITY)

These tests validate core functionality without requiring database connections.

### 3.1 Root Level Configuration Tests

**Status**: ❌ Missing
**Priority**: HIGH

#### 3.1.1 db_context_test.go
**Purpose**: Test database context handling
**Test cases needed**:
- WithContext() sets context correctly
- Context propagates to queries
- Context cancellation stops operations
- Nil context handling

**Estimated effort**: 0.5 days

#### 3.1.2 db_pool_test.go
**Purpose**: Test connection pool configuration
**Test cases needed**:
- MaxOpenConns configuration
- MaxIdleConns configuration
- ConnMaxLifetime configuration
- ConnMaxIdleTime configuration
- Pool exhaustion behavior

**Estimated effort**: 0.5 days

#### 3.1.3 db_ssl_test.go
**Purpose**: Test SSL/TLS connection configuration
**Test cases needed**:
- SSL mode configuration (disable, require, verify-ca, verify-full)
- Certificate path configuration
- SSL parameter parsing from DSN
- SSL configuration for each driver

**Estimated effort**: 0.5 days

---

### 3.2 Database/DB Tests

#### 3.2.1 database/db/dsn_test.go

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

### 3.3 Database/Query Tests

#### 3.3.1 database/query/query_test.go

**Status**: ❌ Missing
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

#### 3.3.2 database/query/clause_test.go

**Status**: ❌ Missing
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

#### 3.3.3 database/query/builder_test.go

**Status**: ❌ Missing
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

#### 3.3.4 database/query/where_exists_test.go

**Status**: ❌ Missing
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

### 3.4 Database/Schema Tests

#### 3.4.1 database/schema/index_test.go

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

### 3.5 Database/ORM Tests

#### 3.5.1 database/orm/buildquery_replica_test.go

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

## Phase 4: Integration Test Enablement (MEDIUM PRIORITY)

### 4.1 Enable PostgreSQL Integration Tests

**Status**: ⚠️ 46/46 files disabled
**Priority**: MEDIUM
**Files**: `integration_tests/postgres/*_test.go`

**Problem**:
- All PostgreSQL tests marked with `//go:build disabled`
- Tests exist but are not running in CI/CD

**Steps**:
1. Set up PostgreSQL test database in CI/CD
2. Update GitHub Actions workflow to include PostgreSQL service
3. Change `//go:build disabled` to `//go:build integration` in all files
4. Run tests and fix any failures
5. Document PostgreSQL test setup in integration_tests/README.md

**Files affected**: 46 test files

**Estimated effort**: 2-3 days

---

### 4.2 Enable MySQL Integration Tests

**Status**: ⚠️ 39/46 files disabled
**Priority**: MEDIUM
**Files**: `integration_tests/mysql/*_test.go`

**Problem**:
- 39 MySQL tests marked with `//go:build disabled`
- 7 tests already enabled with `//go:build integration`

**Steps**:
1. Identify why 39 tests are disabled (check for failures)
2. Fix any implementation issues causing test failures
3. Change `//go:build disabled` to `//go:build integration`
4. Ensure MySQL service is configured in CI/CD
5. Run tests and verify all pass

**Files affected**: 39 test files

**Estimated effort**: 2-3 days

---

### 4.3 Enable SQLite Disabled Tests

**Status**: ⚠️ 5 specific test cases disabled (not short mode skips)
**Priority**: LOW
**Files**: 
- `sqlite_schema_index_test.go` (1 test)
- `sqlite_query_paginate_test.go` (2 tests)
- `sqlite_query_json_test.go` (1 test)
- `sqlite_query_join_test.go` (3 tests)

**Note**: Most SQLite tests are only skipped in "short mode" (via `testing.Short()`). These are not broken - they just require a database connection. The 5 tests below are disabled with specific reasons indicating actual issues.

**Disabled tests with specific reasons**:

1. ~~**sqlite_schema_index_test.go:236**~~ ✅ FIXED
   - ~~Reason: `RenameIndex is currently problematic in SQLite with savepoints`~~
   - Action: Fixed by skipping transaction wrapper for RenameIndex operations

2. ~~**sqlite_query_json_test.go:149**~~ ✅ FIXED
   - ~~Reason: `Update with JSON path requires JSON_SET which is more complex - skipping for now`~~
   - Action: Implemented JSON_SET support in BuildUpdate for SQLite using json_set() function

3. **sqlite_query_join_test.go:222, 256, 283**
   - Reason: `RIGHT JOIN requires SQLite 3.39.0 or higher`
   - Action: Either require SQLite 3.39.0+ or implement workaround

**Steps**:
1. Fix RenameIndex savepoint issue
2. Implement JSON_SET support for SQLite
3. Either upgrade SQLite requirement or implement RIGHT JOIN workaround
4. Enable tests
5. Verify all pass

**Estimated effort**: 1-2 days

---

### 4.4 Create SQL Server Integration Tests

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

### 4.5 Create Turso Integration Tests

**Status**: ❌ Not created
**Priority**: LOW (depends on Turso driver implementation)

**Prerequisites**: Turso driver must be implemented first (Phase 1.2)

**Steps**:
1. Create `integration_tests/turso/` directory
2. Create helper.go with Turso test setup
3. Port relevant SQLite tests (Turso is SQLite-based)
4. Create ~20 test files covering core functionality
5. Document Turso test setup (may require Turso account/token)

**Estimated effort**: 2-3 days

---

## Phase 5: Advanced Integration Test Coverage (LOW PRIORITY)

### 5.1 Read/Write Replica Routing Test

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

### 5.2 InsertGetId PostgreSQL RETURNING Test

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

### 5.3 SlowThreshold Warning Integration Test

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

## Phase 6: Documentation Gaps (MEDIUM PRIORITY)

### 6.1 Update Placeholder Documentation

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

### 6.2 Create Missing Documentation

**Status**: ❌ Missing
**Priority**: MEDIUM

**Documents to create**:

#### 6.2.1 docs/factory.md
- Factory pattern usage (if implemented)
- Defining factories
- Using factories in tests
- Factory relationships

#### 6.2.2 docs/seeder.md
- Seeder usage (if implemented)
- Creating seeders
- Running seeders
- Seeder dependencies

#### 6.2.3 docs/testing.md
- Testing guide for contributors
- Running unit tests
- Running integration tests
- Writing new tests
- Test database setup

#### 6.2.4 docs/performance.md
- Connection pooling best practices
- Query optimization tips
- Eager loading vs lazy loading
- Batch operations
- Benchmarking results

#### 6.2.5 docs/api-reference.md
- Complete API reference
- All methods with signatures
- Parameter descriptions
- Return values
- Examples for each method

**Estimated effort**: 3-4 days total

---

### 6.3 Update README.md

**Status**: ⚠️ Contains inaccuracies
**Priority**: HIGH

**Issues to fix**:
1. Remove Turso from supported databases (or mark as experimental)
2. Remove Factory from features (or mark as not implemented)
3. Remove Migration from features (or mark as not implemented)
4. Remove Seeder from features (or mark as not implemented)
5. Add "Roadmap" section with planned features
6. Add "Contributing" guide link
7. Add badges (build status, coverage, version, license)

**Estimated effort**: 0.5 days

---

## Phase 7: Code Quality Improvements (LOW PRIORITY)

### 7.1 Resolve TODO Comments

**Status**: ⚠️ 8 TODO comments found
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

4. **database/orm/orm.go:253**
   - `TODO: Implement factory when needed`
   - Resolve based on Phase 1.1 decision

5. **database/orm/factory.go:31, 37, 43**
   - Three TODOs for factory methods
   - Resolve based on Phase 1.1 decision

**Estimated effort**: 1 day

---

### 7.2 Remove Unused Dependencies

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

### 7.3 Add Code Coverage Reporting

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

## Phase 8: CI/CD Improvements (LOW PRIORITY)

### 8.1 Enhance GitHub Actions Workflows

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
1. ✅ Make decisions on Factory, Migration, Seeder (implement or remove)
2. ✅ Implement input validation (Phase 2.1, 2.2, 2.3)
3. ✅ Fix Turso driver or remove from supported list
4. ✅ Update README to reflect actual features

**Week 3-4: Unit Test Coverage**
5. ✅ Add root level tests (Phase 3.1)
6. ✅ Add database/db/dsn_test.go (Phase 3.2)
7. ✅ Add database/query tests (Phase 3.3)
8. ✅ Add database/schema tests (Phase 3.4)

**Week 5-6: Integration Test Enablement**
9. ✅ Enable PostgreSQL integration tests (Phase 4.1)
10. ✅ Enable MySQL integration tests (Phase 4.2)
11. ✅ Enable SQLite disabled tests (Phase 4.3)
12. ✅ Create SQL Server integration tests (Phase 4.4)

**Week 7-8: Documentation & Polish**
13. ✅ Update placeholder documentation (Phase 6.1)
14. ✅ Create missing documentation (Phase 6.2)
15. ✅ Resolve TODO comments (Phase 7.1)
16. ✅ CI/CD improvements (Phase 8.1)

**Optional (if time permits)**:
- Advanced integration tests (Phase 5)
- Code coverage reporting (Phase 7.3)
- Turso integration tests (Phase 4.5)

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

### Phase 1: Critical Implementation
- [ ] 1.1 Factory Pattern (decision + implementation)
- [ ] 1.2 Turso Driver (decision + implementation)
- [ ] 1.3 Migration System (decision + implementation)
- [ ] 1.4 Seeder System (decision + implementation)

### Phase 2: Validation & Security
- [ ] 2.1 Column Name Validation
- [ ] 2.2 Nil Destination Validation
- [ ] 2.3 COUNT(DISTINCT) Support

### Phase 3: Unit Tests
- [ ] 3.1.1 db_context_test.go
- [ ] 3.1.2 db_pool_test.go
- [ ] 3.1.3 db_ssl_test.go
- [ ] 3.2.1 database/db/dsn_test.go
- [ ] 3.3.1 database/query/query_test.go
- [ ] 3.3.2 database/query/clause_test.go
- [ ] 3.3.3 database/query/builder_test.go
- [ ] 3.3.4 database/query/where_exists_test.go
- [ ] 3.4.1 database/schema/index_test.go
- [ ] 3.5.1 database/orm/buildquery_replica_test.go

### Phase 4: Integration Tests
- [ ] 4.1 Enable PostgreSQL tests (46 files)
- [ ] 4.2 Enable MySQL tests (39 files)
- [ ] 4.3 Enable SQLite tests (4 files)
- [ ] 4.4 Create SQL Server tests (~40 files)
- [ ] 4.5 Create Turso tests (~20 files)

### Phase 5: Advanced Integration
- [ ] 5.1 Read/Write Replica Routing Test
- [ ] 5.2 InsertGetId PostgreSQL Test
- [ ] 5.3 SlowThreshold Warning Test

### Phase 6: Documentation
- [ ] 6.1 Update placeholder docs (3 files)
- [ ] 6.2 Create missing docs (5 files)
- [ ] 6.3 Update README.md

### Phase 7: Code Quality
- [ ] 7.1 Resolve TODO comments (8 items)
- [ ] 7.2 Remove unused dependencies
- [ ] 7.3 Add code coverage reporting

### Phase 8: CI/CD
- [ ] 8.1 Enhance GitHub Actions workflows

---

## Estimated Total Effort

- **Phase 1**: 12-18 days (depends on decisions)
- **Phase 2**: 2 days
- **Phase 3**: 7.5 days
- **Phase 4**: 8-12 days
- **Phase 5**: 2 days
- **Phase 6**: 5.5 days
- **Phase 7**: 2 days
- **Phase 8**: 1.5 days

**Total**: 40-50 days (8-10 weeks for one developer)

With 2-3 developers working in parallel: **4-6 weeks to zero gaps**

---

## Notes

- This plan assumes decisions on Factory, Migration, and Seeder are made quickly
- Integration test enablement may reveal additional bugs requiring fixes
- Documentation effort can be parallelized with implementation work
- CI/CD improvements can be done incrementally
- Code coverage and quality improvements are ongoing

**Last Updated**: May 25, 2026
