# Neat ORM Implementation Gaps

**Date**: May 30, 2026
**Last updated**: June 3, 2026
**Purpose**: Track implementation gaps in the neat ORM project.

---

## Status

Most implementation gaps have been resolved. All critical CRUD operation gaps have been addressed.

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
- `oracle_query_update_or_insert_test.go` - "UpdateOrInsert test failing - record not found after update"

### Required Tasks

#### 1. Implement Missing Features
- Implement update tests (`oracle_update_test.go`)
- Implement transaction tests (`oracle_transaction_test.go`)
- Implement soft delete tests (`oracle_soft_delete_test.go`)
- Implement where advanced tests (`oracle_where_advanced_test.go`)
- Implement where any/all advanced tests (`oracle_where_any_all_advanced_test.go`)
- Implement JSON query tests (`oracle_query_json_test.go`)
- Implement spatial query tests (`oracle_query_spatial_test.go`)

#### 3. CI/CD Configuration
- Add Oracle integration test execution to `.github/workflows/tests.yml` (after tests are working)
- Update `integration_tests/README.md` with Oracle documentation (after tests are working)
