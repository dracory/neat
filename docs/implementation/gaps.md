# Neat ORM Implementation Gaps

**Date**: May 30, 2026
**Last updated**: June 3, 2026
**Purpose**: Track implementation gaps in the neat ORM project.

---

## Status

All implementation gaps have been resolved. The project has zero remaining gaps.

---

## Oracle Integration Plan

### Overview
Oracle database integration has core infrastructure implemented but most integration tests are stubs with known issues that need to be resolved.

### Stub Tests with Known Issues

**Schema Tests (case sensitivity issues):**
- `oracle_schema_column_types_test.go` - "case sensitivity issues with GetColumns"
- `oracle_schema_column_modifiers_test.go` - "case sensitivity issues"
- `oracle_schema_foreign_key_test.go` - "case sensitivity issues"
- `oracle_schema_timestamp_test.go` - "case sensitivity issues"
- `oracle_schema_table_test.go` - "case sensitivity issues with HasTable and table listing", "identity column with primary key syntax issue"
- `oracle_schema_rename_column_test.go` - "case sensitivity issues"
- `oracle_schema_column_methods_test.go` - "case sensitivity issues"
- `oracle_schema_column_change_test.go` - "case sensitivity issues"

**Query Tests (various Oracle-specific issues):**
- `oracle_query_json_test.go` - "to be implemented"
- `oracle_query_spatial_test.go` - "to be implemented"
- `oracle_where_advanced_test.go` - "to be implemented"
- `oracle_where_any_all_advanced_test.go` - "to be implemented"
- `oracle_update_test.go` - "to be implemented"
- `oracle_transaction_test.go` - "to be implemented"
- `oracle_soft_delete_test.go` - "to be implemented"
- `oracle_query_join_test.go` - "SQL syntax differs for join aliases - ORA-00933 error"
- `oracle_query_lock_test.go` - "ORA-02014 - cannot select FOR UPDATE from view with DISTINCT, GROUP BY, etc."
- `oracle_query_update_or_insert_test.go` - "UpdateOrInsert test failing - record not found after update"
- `oracle_query_to_sql_test.go` - "uses :1 placeholder syntax instead of interpolated values in ToRawSql"
- `oracle_raw_test.go` - "raw update with concatenation failing - ORA-00911 invalid character", "database functions test failing - ORA-00911"
- `oracle_query_paginate_test.go` - "data cleanup issues with Oracle"
- `oracle_query_belongs_to_test.go` - "With() method has known issues loading associations"
- `oracle_query_association_test.go` - "Delete method has known issues with WHERE clause"

### Required Tasks

#### 1. Fix Critical CRUD Operations (Highest Priority)
- Fix Delete method WHERE clause (`oracle_query_association_test.go`)

#### 2. Fix Other Query Builder Issues
- Fix join alias syntax (ORA-00933) (`oracle_query_join_test.go`)
- Fix FOR UPDATE with DISTINCT/GROUP BY (ORA-02014) (`oracle_query_lock_test.go`)
- Fix ToRawSql placeholder syntax (`oracle_query_to_sql_test.go`)
- Fix raw query concatenation (ORA-00911) (`oracle_raw_test.go`)
- Fix With() method association loading (`oracle_query_belongs_to_test.go`)
- Fix paginate data cleanup (`oracle_query_paginate_test.go`)

#### 3. Fix Schema Case Sensitivity Issues
- Fix `GetColumns` to handle Oracle's uppercase table/column names
- Fix `HasTable` and table listing methods
- Fix identity column with primary key syntax
- Files: All `oracle_schema_*.go` test files

#### 4. Implement Missing Features
- Implement update tests (`oracle_update_test.go`)
- Implement transaction tests (`oracle_transaction_test.go`)
- Implement soft delete tests (`oracle_soft_delete_test.go`)
- Implement where advanced tests (`oracle_where_advanced_test.go`)
- Implement where any/all advanced tests (`oracle_where_any_all_advanced_test.go`)
- Implement JSON query tests (`oracle_query_json_test.go`)
- Implement spatial query tests (`oracle_query_spatial_test.go`)

#### 5. CI/CD Configuration
- Add Oracle integration test execution to `.github/workflows/tests.yml` (after tests are working)
- Update `integration_tests/README.md` with Oracle documentation (after tests are working)
