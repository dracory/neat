# Neat ORM Implementation Gaps - Executable Plan

**Date**: May 25, 2026
**Last reviewed**: May 30, 2026
**Last updated**: May 30, 2026
**Purpose**: Comprehensive executable plan to eliminate all implementation gaps in the neat ORM project.

---

## Executive Summary

This document provides a complete, step-by-step plan to bring the neat ORM to production-ready status with zero gaps. All items are prioritized and organized for sequential execution.

**Current Status:**
- ⚠️ PostgreSQL integration tests: 46 files total, many skipped due to unimplemented features
- ⚠️ MySQL integration tests: Several files skipped (Association, Spatial, etc.)
- ❌ SQL Server integration tests: Not created

---

## Phase 1: Integration Test Enablement (HIGH PRIORITY)

### 1.1 Enable PostgreSQL Integration Tests

**Status**: ✅ COMPLETED (May 30, 2026)
**Priority**: HIGH

**Completed gaps**:
1. ✅ **Subquery Parameter Numbering**: Fixed parameter numbering for SELECT and HAVING clauses with subqueries. PostgreSQL's numbered parameters ($1, $2, etc.) are now correctly handled across all query clauses.
2. ✅ **Schema Change Syntax**: Enhanced PostgreSQL's `ALTER COLUMN` syntax in the schema builder's `Change()` method to handle type, nullable, and default changes with proper SQL statements.
3. ⏳ **Association Tests**: Still pending Phase 2.1 completion (Association API implementation).

**Changes made**:
- Updated `database/query/builder_select.go` to add placeholder numbering for SELECT clauses and FROM subqueries
- Enhanced `database/schema/grammars/postgres.go` CompileChange() to generate separate ALTER COLUMN statements for type, nullable, and default changes
- Implemented full test suite in `integration_tests/postgres/postgres_schema_column_change_test.go`
- Enabled HAVING subquery tests in `integration_tests/postgres/postgres_query_group_having_test.go`
- Enabled SELECT subquery test in `integration_tests/postgres/postgres_query_select_test.go`
- Removed skip for Change modifier test in `integration_tests/postgres/postgres_schema_column_modifiers_test.go`

**Estimated effort**: 3-4 days (completed)

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

## Phase 2: ORM Feature Completion (HIGH PRIORITY)

### 2.1 Complete Association API

**Status**: ⚠️ Stubs exist
**Priority**: HIGH

**Purpose**: Implement full support for managing relationships via the `Association` method.

**Current state**:
- `Association` method exists but returns errors for `Append`, `Replace`, `Delete`, and `Clear` in most relationship types.
- Integration tests are skipped across all drivers.

**Steps**:
1. Implement `Append`, `Replace`, `Delete`, and `Clear` for `HasOne` relationships.
2. Implement `Append`, `Replace`, `Delete`, and `Clear` for `HasMany` relationships.
3. Implement `Append`, `Replace`, `Delete`, and `Clear` for `BelongsTo` relationships.
4. Implement polymorphic association support.
5. Enable and verify integration tests in `mysql`, `postgres`, and `sqlite`.

**Estimated effort**: 3-4 days

### 2.2 Implement WithCount and WithExists

**Status**: ❌ Stubs only
**Priority**: MEDIUM

**Purpose**: Allow eager loading of relationship counts and existence.

**Current state**:
- `WithCount` and `WithExists` methods are stubs in `database/query/query_relations.go`.

**Steps**:
1. Implement `WithCount` to add subqueries for relationship counts.
2. Implement `WithExists` to add subqueries for relationship existence.
3. Add support for constraints in `WithCount` and `WithExists`.
4. Add unit and integration tests.

**Estimated effort**: 2 days

---

## Phase 3: Advanced Query Support (MEDIUM PRIORITY)

### 3.1 Raw Expressions in Create and Update

**Status**: ❌ Not supported
**Priority**: MEDIUM

**Purpose**: Support using `Raw()` expressions as values in `Create()` and `Update()` calls.

**Current state**:
- `Raw()` returns a `*Query` or similar structure that is not correctly handled when passed as a value in maps for `Create` or `Update`.
- This blocks features like spatial data inserts (e.g., `ST_GeomFromText`).

**Steps**:
1. Modify `Create` and `Update` logic to detect raw expressions in values.
2. Ensure raw expressions are not parameterized but injected directly into the SQL.
3. Update `structScanDests` and related logic to handle these cases.

**Estimated effort**: 2 days

---

## Phase 4: CI/CD & Quality Improvements (LOW PRIORITY)

### 4.1 Enhance GitHub Actions Workflows

**Status**: ⚠️ Basic workflows exist
**Priority**: LOW

**Improvements needed**:
1. Add PostgreSQL service to integration tests
2. Add MySQL service to integration tests
3. Add SQL Server service to integration tests
4. Add code coverage reporting
5. Add linting (golangci-lint)
6. Add security scanning (gosec)

**Estimated effort**: 1-2 days

---

## Phase 5: Advanced Features (LOW PRIORITY)

### 5.1 Spatial Data Type Support

**Status**: ❌ Not implemented
**Priority**: LOW

**Purpose**: Add support for MySQL/PostgreSQL spatial data types.

**Current state**:
- Test exists: `integration_tests/mysql/mysql_query_spatial_test.go` (skipped)
- Dependent on Phase 3.1.

**Estimated effort**: 3-4 days

---

## Success Criteria

The project will have **ZERO GAPS** when:

- [ ] All stub implementations are either completed or removed
- [ ] All disabled integration tests are enabled and passing
- [ ] PostgreSQL subquery parameter numbering is fixed
- [ ] Schema `Change()` works across all supported databases
- [ ] `Association` API is fully functional
- [ ] `WithCount` and `WithExists` are implemented
- [ ] Raw expressions can be used in `Create`/`Update`
- [ ] CI/CD runs all tests (MySQL, Postgres, SQLite, SQL Server) successfully
- [ ] Code coverage is measured and reported

---

## Tracking Progress

### Phase 1: Integration Tests
- [x] 1.1 Enable PostgreSQL tests
- [ ] 1.2 Create SQL Server tests

### Phase 2: ORM Features
- [ ] 2.1 Complete Association API
- [ ] 2.2 Implement WithCount and WithExists

### Phase 3: Advanced Query Support
- [ ] 3.1 Raw Expressions in Create/Update

### Phase 5: Advanced Features
- [ ] 5.1 Spatial Data Type Support

---

## Estimated Total Effort

- **Phase 1**: 6-8 days
- **Phase 2**: 5-6 days
- **Phase 3**: 2 days
- **Phase 4**: 1-2 days
- **Phase 5**: 3-4 days

**Total**: 17-22 days (~4 weeks for one developer)

**Last Updated**: May 30, 2026
