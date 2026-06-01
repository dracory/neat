# Integration Test Deduplication Plan

## Overview
Integration tests across MySQL, PostgreSQL, SQLite, SQL Server, and Turso have significant duplication. This plan outlines the refactoring to move shared test logic to `integration_tests/common/`.

## Completed
- ✅ **Query Count Tests** - Created `query_count_helpers.go` with `QueryCountBasic` and `QueryCountWithTable` functions
- ✅ **Paginate Tests** - Created `paginate_helpers.go` with shared functions and refactored all 5 database-specific test files
- ✅ **Create Tests** - Created `create_helpers.go` with shared functions and refactored all 5 database-specific test files (PostgreSQL BigSerial tests kept in postgres package)
- ✅ **Delete Tests** - Created `delete_helpers.go` with shared functions and refactored all 5 database-specific test files
- ✅ **Join Tests** - Created `join_helpers.go` with shared functions and refactored MySQL and PostgreSQL test files (PostgreSQL-specific tests kept in postgres package)
- ✅ **Group/Having Tests** - Created `group_having_helpers.go` with shared functions and refactored all 5 database-specific test files (SQLite and Turso-specific tests kept in their packages)
- ✅ **Aggregate Tests** - Created `aggregate_helpers.go` with shared functions and refactored all 5 database-specific test files (PostgreSQL, SQLite, and Turso-specific tests kept in their packages)

## Priority 1: High Duplication Tests

### 1. Paginate Tests ✅
**Files:** `*_query_paginate_test.go` (5 databases)
**Shared Helper:** `seedPaginateTestData(t, db)`
**Test Functions:**
- `Test*IntegrationPaginateFirstPage`
- `Test*IntegrationPaginateSecondPage`
- `Test*IntegrationPaginateWithConditions`
- `Test*IntegrationPaginateWithSelectAliases`
- `Test*IntegrationPaginateLastPage` (SQLite only)
- `Test*IntegrationPaginatePageBeyondBounds` (SQLite only)
- `Test*IntegrationPaginateEmptyResults` (SQLite only)
- `Test*IntegrationCountWithSelectAlias` (SQLite only)

**Action:** ✅ Created `paginate_helpers.go` with shared functions

### 2. Create Tests ✅
**Files:** `*_query_create_test.go` (5 databases)
**Shared Tests:**
- `Test*IntegrationQueryCreateByStruct` - identical across all databases
- `Test*IntegrationQueryBatchCreateByStruct` - identical across all databases
- `Test*IntegrationQueryCreateByMap` - identical across MySQL, PostgreSQL, SQLite
- `Test*IntegrationQueryInsertGetIdByStruct` - identical across all databases
- `Test*IntegrationQueryInsertGetIdByMap` - identical across all databases
**Database-Specific:**
- PostgreSQL: `TestPostgreSQLIntegrationQueryInsertGetIdBigSerial` (BigSerial type)
- PostgreSQL: `TestPostgreSQLIntegrationQueryInsertGetIdBigSerialByMap` (BigSerial type)
**Action:** ✅ Created `create_helpers.go` with shared functions, kept BigSerial tests in postgres package

### 3. Delete Tests ✅
**Files:** `*_query_delete_test.go` (5 databases)
**Shared Tests:**
- `Test*IntegrationQueryDeleteByModel` - identical across all databases (SQLite has extra nil check)
- `Test*IntegrationQueryDeleteByTable` - identical across all databases (SQLite has extra nil check)
- `Test*IntegrationQueryDeleteByModelWithWhere` - identical across all databases (SQLite has extra nil check)
**Note:** SQLite adds `if res == nil` checks that MySQL/PostgreSQL don't have. These can be harmonized or kept as-is.
**Action:** ✅ Created `delete_helpers.go` with shared functions (harmonized nil checks)

### 4. Find Tests
**Files:** `*_find_test.go` (4 databases: mysql, postgres, sqlite, sqlserver)
**Action:** Create `find_helpers.go` with shared test functions

## Priority 2: Medium Duplication Tests

### 5. Aggregate Tests ✅
**Files:** `*_query_aggregate_test.go` (5 databases)
**Note:** MySQL uses `seedAggregateTestData` helper with time fields, PostgreSQL creates users inline with explicit IDs. Different seeding strategies.
**Action:** ✅ Created `aggregate_helpers.go` with shared functions, kept database-specific tests

### 6. Chunk Tests
**Files:** `*_query_chunk_test.go` (5 databases)
**Note:** MySQL uses `seedChunkTestData` helper and deletes existing data first, PostgreSQL creates users inline. Different approaches.
**Action:** Review if seeding can be harmonized before creating helpers

### 7. Pluck Tests
**Files:** `*_query_pluck_test.go` (5 databases)
**Note:** MySQL uses `seedPluckTestData` helper, PostgreSQL creates users inline. Same data structure, different seeding approach.
**Action:** Can be deduplicated after harmonizing seeding strategy

### 8. Value Tests
**Files:** `*_query_value_test.go` (5 databases)
**Note:** MySQL uses `seedValueTestData` helper, PostgreSQL creates users inline. Same data structure, different seeding approach.
**Action:** Can be deduplicated after harmonizing seeding strategy

### 9. Distinct Tests
**Files:** `*_query_distinct_test.go` (4 databases: mysql, postgres, sqlite, sqlserver)
**Note:** MySQL uses `seedDistinctTestData` helper with time fields, PostgreSQL creates users inline. Same data structure.
**Action:** Can be deduplicated after harmonizing seeding strategy

### 10. Join Tests ✅
**Files:** `*_query_join_test.go` (5 databases)
**Note:** Both MySQL and PostgreSQL use identical `seedJoinTestData` helper with time fields. Good candidate for deduplication.
**Action:** ✅ Created `join_helpers.go` with shared seed function and test functions

### 11. Group/Having Tests ✅
**Files:** `*_query_group_having_test.go` (5 databases)
**Note:** Both MySQL and PostgreSQL use identical `seedGroupHavingTestData` helper. Good candidate for deduplication.
**Action:** ✅ Created `group_having_helpers.go` with shared seed function and test functions

### 12. Order/Limit/Offset Tests
**Files:** `*_query_order_limit_offset_test.go` (5 databases)
**Note:** MySQL uses `seedOrderLimitOffsetTestData` helper, PostgreSQL creates users inline. Same data structure.
**Action:** Can be deduplicated after harmonizing seeding strategy

### 13. JSON Tests
**Files:** `*_query_json_test.go` (5 databases)
**Action:** Create `json_helpers.go`

### 14. Scopes Tests
**Files:** `*_query_scopes_test.go` (5 databases)
**Action:** Create `scopes_helpers.go`

### 15. Soft Delete Tests
**Files:** `*_soft_delete_test.go` (5 databases)
**Action:** Create `soft_delete_helpers.go`

### 16. Where Tests
**Files:** `*_where_test.go` (3 databases: mysql, postgres, sqlserver)
**Action:** Create `where_helpers.go`

## Implementation Guidelines

1. **File Naming:** Use `*_helpers.go` (not `*_test.go`) to allow cross-package imports
2. **Function Naming:** Use descriptive names without `Test` prefix (e.g., `QueryCountBasic`)
3. **Function Signature:** `func FunctionName(t *testing.T, db *database.Database)`
4. **Test Wrapper:** Keep database-specific test files with minimal setup:
   ```go
   func TestDatabaseIntegrationFeature(t *testing.T) {
       if testing.Short() {
           t.Skip("Skipping integration test in short mode")
       }
       db := SetupDatabaseTest(t)
       common.TestFeature(t, db)
   }
   ```
5. **Data Seeding:** Move seed functions to helpers where applicable
6. **Database-Specific Logic:** Keep database-specific tests in their own files (e.g., spatial tests for MySQL/PostgreSQL)

## Revised Priority Based on Analysis

### Immediate (High ROI - Identical Code):
1. **Paginate Tests** ✅ - Identical `seedPaginateTestData` helper, test logic nearly identical
2. **Create Tests** ✅ - Most tests identical, only PostgreSQL has BigSerial-specific tests
3. **Delete Tests** ✅ - Identical logic, SQLite has extra nil checks (harmonized)
4. **Join Tests** ✅ - Identical `seedJoinTestData` helper across MySQL/PostgreSQL
5. **Group/Having Tests** ✅ - Identical `seedGroupHavingTestData` helper across MySQL/PostgreSQL

### Secondary (Medium ROI - Needs Harmonization):
6. **Pluck Tests** - Same data, different seeding approach (helper vs inline)
7. **Value Tests** - Same data, different seeding approach (helper vs inline)
8. **Distinct Tests** - Same data, different seeding approach (helper vs inline with time fields)
9. **Order/Limit/Offset Tests** - Same data, different seeding approach (helper vs inline)
10. **Aggregate Tests** ✅ - Different seeding strategies (time fields vs explicit IDs) - completed with shared helpers
11. **Chunk Tests** - Different seeding strategies (helper with cleanup vs inline)

### Review Needed:
12. **JSON Tests** - Need to review for database-specific JSON syntax differences
13. **Scopes Tests** - Need to review for duplication
14. **Soft Delete Tests** - Need to review for duplication
15. **Where Tests** - Need to review for duplication
16. **Find Tests** - Need to review for duplication

## Estimated Impact
- **Lines of Code Reduction:** ~1500-2500 lines (revised down due to seeding strategy differences)
- **Files Modified:** ~50-60 test files
- **New Helper Files:** ~10-12 files (fewer due to database-specific variations)

## Notes
- Some tests may have database-specific variations (e.g., SQLite has additional paginate tests)
- PostgreSQL has array-specific tests that should remain in postgres package
- MySQL has spatial tests that should remain in mysql package
- SQL Server may have specific T-SQL variations that need special handling
