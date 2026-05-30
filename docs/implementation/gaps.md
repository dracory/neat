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

## Phase 2: ORM Feature Completion (HIGH PRIORITY)

### 2.1 Implement WithCount and WithExists

**Status**: ✅ Completed
**Priority**: MEDIUM

**Purpose**: Allow eager loading of relationship counts and existence.

**Current state**:
- `WithCount` and `WithExists` methods are fully implemented in `database/query/query_relations.go`.
- Subquery generation is handled in `database/query/builder_select.go`.
- Constraint callbacks are supported for both methods.
- Unit and integration tests have been added.

**Steps**:
1. ✅ Implement `WithCount` to add subqueries for relationship counts.
2. ✅ Implement `WithExists` to add subqueries for relationship existence.
3. ✅ Add support for constraints in `WithCount` and `WithExists`.
4. ✅ Add unit and integration tests.

**Estimated effort**: 2 days

---

## Phase 3: Advanced Query Support (MEDIUM PRIORITY)

### 3.1 Raw Expressions in Create and Update

**Status**: ✅ Implemented
**Priority**: MEDIUM

**Purpose**: Support using `Raw()` expressions as values in `Create()` and `Update()` calls.

**Current state**:
- Implemented `RawExpr()` function to create raw SQL expressions for use in Create/Update values.
- Modified `BuildInsert` and `BuildUpdate` to detect `RawExpression` in values and inject SQL directly without parameterization.
- Raw expressions can include their own arguments (e.g., `RawExpr("score + ?", 10)`).
- This enables features like spatial data inserts (e.g., `ST_GeomFromText`) and database functions (e.g., `NOW()`).

**Usage example**:
```go
db.Table("users").Create(map[string]any{
    "name":       "John",
    "created_at": RawExpr("NOW()"),
})

db.Table("users").Update(map[string]any{
    "score": RawExpr("score + ?", 10),
})
```

**Implementation details**:
- Added `RawExpr()` public function in `database/query/query.go`
- Modified `builder_insert.go` to handle raw expressions in both single and bulk inserts
- Modified `builder_update.go` to handle raw expressions in map and struct updates
- Added comprehensive tests in `builder_raw_expression_test.go`

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
- [x] PostgreSQL subquery parameter numbering is fixed
- [x] Schema `Change()` works for PostgreSQL
- [x] `Association` API is fully functional (including polymorphic)
- [ ] `WithCount` and `WithExists` are implemented
- [ ] Raw expressions can be used in `Create`/`Update`
- [ ] CI/CD runs all tests (MySQL, Postgres, SQLite, SQL Server) successfully
- [ ] Code coverage is measured and reported

---

## Tracking Progress

### Phase 1: Integration Tests
- [ ] 1.1 Create SQL Server tests

### Phase 2: ORM Features
- [x] 2.1 Complete Association API
- [x] 2.2 Implement Polymorphic Association Support
- [ ] 2.3 Implement WithCount and WithExists

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

**Total**: 14-18 days (~3-4 weeks for one developer)

**Last Updated**: May 30, 2026 (Polymorphic Association API completed)
