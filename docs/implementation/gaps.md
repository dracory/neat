# Neat ORM Implementation Gaps - Executable Plan

**Date**: May 30, 2026
**Last reviewed**: May 30, 2026
**Last updated**: May 30, 2026
**Purpose**: Comprehensive executable plan to eliminate all implementation gaps in the neat ORM project.

---

## Executive Summary

This document provides a complete, step-by-step plan to bring the neat ORM to production-ready status with zero gaps. All items are prioritized and organized for sequential execution.

**Current Status:**
- ✅ PostgreSQL integration tests: Core query and schema tests enabled (Association tests completed)
- ✅ MySQL integration tests: Association tests completed (Spatial tests still skipped)
- ❌ SQL Server integration tests: Not created

---

## Phase 1: Integration Test Enablement (HIGH PRIORITY)

### 1.1 Create SQL Server Integration Tests

**Status**: ❌ Not created
**Priority**: HIGH

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

## Phase 2: CI/CD & Quality Improvements (LOW PRIORITY)

### 2.1 Enhance GitHub Actions Workflows

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

## Phase 3: Advanced Features (LOW PRIORITY)

### 3.1 Spatial Data Type Support

**Status**: ❌ Not implemented
**Priority**: LOW

**Purpose**: Add support for MySQL/PostgreSQL spatial data types.

**Current state**:
- Test exists: `integration_tests/mysql/mysql_query_spatial_test.go` (skipped)
- Raw expressions are now supported (enables spatial data inserts)

**Estimated effort**: 3-4 days

---

## Success Criteria

The project will have **ZERO GAPS** when:

- [ ] All stub implementations are either completed or removed
- [ ] All disabled integration tests are enabled and passing
- [x] PostgreSQL subquery parameter numbering is fixed
- [x] Schema `Change()` works for PostgreSQL
- [x] `Association` API is fully functional (including polymorphic)
- [x] `WithCount` and `WithExists` are implemented
- [x] Raw expressions can be used in `Create`/`Update`
- [ ] CI/CD runs all tests (MySQL, Postgres, SQLite, SQL Server) successfully
- [ ] Code coverage is measured and reported

---

## Tracking Progress

### Phase 1: Integration Tests
- [ ] 1.1 Create SQL Server tests

### Phase 2: CI/CD & Quality
- [ ] 2.1 Enhance GitHub Actions Workflows

### Phase 3: Advanced Features
- [ ] 3.1 Spatial Data Type Support

---

## Estimated Total Effort

- **Phase 1**: 3-4 days
- **Phase 2**: 1-2 days
- **Phase 3**: 3-4 days

**Total**: 7-10 days (~2 weeks for one developer)

**Last Updated**: May 30, 2026 (Raw Expressions in Create/Update completed)
