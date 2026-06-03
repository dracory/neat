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

### Fixed Issues

**Schema Introspection (ORA-00923):**
- Fixed `CompileColumns` in Oracle grammar to use `user_tab_columns` instead of `all_tab_columns`, avoiding schema parameter issues
- Added join with `user_tab_identity_cols` to properly detect auto-increment columns
- Updated column mapping to handle Oracle's uppercase column names and nullable values ('Y'/'N')
- Fixed `TypeFloat` to not use precision parameter (Oracle BINARY_FLOAT doesn't support it)

### Remaining Issues

**Schema Tests (unrelated Oracle grammar issues):**
- `oracle_schema_column_modifiers_test.go` - "skipped - Oracle column modifiers need investigation (ORA-00907)"
- `oracle_schema_foreign_key_test.go` - "skipped - Oracle foreign key handling needs investigation (ORA-01735)"
- `oracle_schema_column_types_test.go` (default value test) - "skipped - Oracle default value handling needs investigation (ORA-00907)"
- `oracle_schema_timestamp_test.go` - "skipped - Oracle-specific timestamp test not yet implemented"
- `oracle_schema_table_test.go` - "skipped - Oracle-specific table test not yet implemented"
- `oracle_schema_rename_column_test.go` - "skipped - Oracle-specific rename column test not yet implemented"
- `oracle_schema_column_methods_test.go` - "skipped - Oracle-specific column methods test not yet implemented"
- `oracle_schema_column_change_test.go` - "skipped - Oracle-specific column change test not yet implemented"