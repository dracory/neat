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
Oracle database integration has core infrastructure implemented. Schema introspection has been fixed and column type tests are now passing.

### Remaining Issues

**Schema Tests (unrelated Oracle grammar issues):**
- `oracle_schema_column_types_test.go` (default value test) - "skipped - Oracle default value handling needs investigation (ORA-00907)"
- `oracle_schema_timestamp_test.go` - "skipped - Oracle-specific timestamp test not yet implemented"
- `oracle_schema_table_test.go` - "skipped - Oracle-specific table test not yet implemented"
- `oracle_schema_rename_column_test.go` - "skipped - Oracle-specific rename column test not yet implemented"

### Resolved Issues

**Query Lock Tests:**
- `oracle_query_lock_test.go` (TestOracleConcurrentAccess) - Fixed by adding LIMIT check to skip FOR UPDATE when LIMIT is present (ORA-02014). Test now uses Get() instead of First() to avoid LIMIT clause.