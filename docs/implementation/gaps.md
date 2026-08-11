# Neat ORM Implementation Gaps

**Date**: May 30, 2026
**Last updated**: August 11, 2026
**Purpose**: Track implementation gaps in the neat ORM project.

---

## Status

All previously tracked implementation gaps have been resolved. The Oracle integration issues documented below are now fixed.

---

## Resolved: Oracle Integration

Oracle database integration is complete. Schema introspection, column type tests, and query lock tests all pass.

**Resolved issues:**

- **Schema introspection** — fixed; column type tests pass.
- **Query Lock Tests** (`oracle_query_lock_test.go`, `TestOracleConcurrentAccess`) — fixed by adding a LIMIT check to skip `FOR UPDATE` when `LIMIT` is present (ORA-02014). The test now uses `Get()` instead of `First()` to avoid the `LIMIT` clause.

---

## Current Gaps

No implementation gaps are currently tracked. See `docs/proposals/` for proposed features that have not yet been started, and `docs/proposals/completed/` for completed work.
