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
- ✅ SQL Server integration tests: Completed (43 test files, CI/CD configured)

---

## Phase 1: Integration Test Enablement (HIGH PRIORITY)

### 1.1 Create SQL Server Integration Tests

**Status**: ✅ Completed
**Priority**: HIGH

**Completed**:
1. ✅ Created `integration_tests/sqlserver/` directory
2. ✅ Created helper.go with SQL Server test setup
3. ✅ Ported tests from MySQL/PostgreSQL patterns
4. ✅ Created 43 test files covering:
   - Query operations (CRUD, aggregates, joins, etc.)
   - Schema operations (create, alter, drop tables)
   - Transactions
   - Soft deletes
5. ✅ Added SQL Server service to CI/CD
6. ✅ Documented SQL Server test setup

**Completed on**: May 31, 2026

---

## Phase 2: CI/CD & Quality Improvements (LOW PRIORITY)

### 2.1 Enhance GitHub Actions Workflows

**Status**: ✅ Completed
**Priority**: LOW

**Completed**:
1. ✅ Add PostgreSQL service to integration tests
2. ✅ Add MySQL service to integration tests
3. ✅ Add SQL Server service to integration tests
4. ✅ Add code coverage reporting (Codecov)
5. ✅ Add linting (golangci-lint)
6. ✅ Add security scanning (gosec)

**Completed on**: May 31, 2026

---

## Phase 3: Advanced Features (LOW PRIORITY)

### 3.1 Spatial Data Type Support

**Status**: ✅ Implemented
**Priority**: LOW

**Purpose**: Add support for MySQL/PostgreSQL spatial data types.

**Completed**:
- ✅ Blueprint methods for all spatial types (Point, Geometry, LineString, Polygon, GeometryCollection, MultiPoint, MultiLineString, MultiPolygon)
- ✅ MySQL type handlers for all spatial types
- ✅ PostgreSQL type handlers for all spatial types
- ✅ Integration tests for MySQL and PostgreSQL
- ✅ Raw expressions enable spatial data operations (ST_GeomFromText, ST_AsText)

**Note**: PostgreSQL tests require PostGIS extension to be installed; tests skip gracefully if not available.

**Completed on**: May 31, 2026

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
- [x] CI/CD runs all tests (MySQL, Postgres, SQLite, SQL Server) successfully
- [ ] Code coverage is measured and reported

---

## Tracking Progress

### Phase 1: Integration Tests
- [ ] 1.1 Create SQL Server tests

### Phase 2: CI/CD & Quality
- [x] 2.1 Enhance GitHub Actions Workflows

### Phase 3: Advanced Features
- [x] 3.1 Spatial Data Type Support

---

## Estimated Total Effort

- **Phase 1**: 3-4 days
- **Phase 2**: 1-2 days
- **Phase 3**: 3-4 days

**Total**: 7-10 days (~2 weeks for one developer)

**Last Updated**: May 31, 2026 (All phases completed - spatial data type support verified)
