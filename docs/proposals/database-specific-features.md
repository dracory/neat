# Database-Specific Features

**Date**: June 1, 2026
**Last Reviewed**: August 25, 2026
**Status**: Partially Done
**Priority**: Medium
**Impact**: High
**Effort**: High
**ROI**: Medium

## Description

Add support for database-specific features.

## Rationale

- Leverage database capabilities
- Better performance for specific databases
- Feature parity with native drivers

## Implemented

- JSON query methods (`WhereJsonContains`, `WhereJsonContainsKey`, `WhereJsonLength`) using `->` operator syntax for MySQL, PostgreSQL, and SQLite
- Schema grammars for PostgreSQL, MySQL, SQLite, SQL Server, and Oracle with column type support (JSONB, fulltext indexes)
- Open proposal for driver-specific JSON grammar at `docs/proposals/driver-specific-json-query-support.md`
- Integration tests for JSON queries across MySQL, PostgreSQL, SQLite, CockroachDB, and TiDB
- CockroachDB and TiDB integration test suites (37 and 48 test files respectively)
- Turso integration test suite (37 test files)
- Embedded filesystem (embed.FS) support for CSVDB, JSONDB, XMLDB drivers
- Driver constants (`contracts.Driver*`) replacing string literals throughout the codebase for type-safe driver references

## Remaining Work

- PostgreSQL: JSONB array operators, array operations, full-text search
- MySQL: Spatial functions, window functions
- SQLite: FTS5, R-Tree, generated columns
- SQL Server: Spatial data, hierarchyid
- Driver-specific JSON grammar for Oracle and SQL Server (see `driver-specific-json-query-support.md`)
- Feature detection and fallback for unsupported databases

## Proposed Features

- PostgreSQL: JSONB operators, array operations, full-text search
- MySQL: Full-text search, spatial functions, window functions
- SQLite: FTS5, R-Tree, generated columns
- SQL Server: Spatial data, hierarchyid

## Implementation Details

- Database-specific query methods
- Type-safe operators
- Feature detection
- Fallback for unsupported databases

## References

- Extracted from `docs/proposals/feature-requests.md` (item #15)
- Related proposal: `docs/proposals/driver-specific-json-query-support.md`
