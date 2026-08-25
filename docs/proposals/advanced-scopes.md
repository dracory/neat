# Advanced Scopes

**Date**: June 1, 2026
**Last Reviewed**: August 25, 2026
**Status**: Partially Done
**Priority**: Medium
**Impact**: Medium
**Effort**: Low
**ROI**: High

## Description

Add global scopes and scope composition features.

## Rationale

- GORM has advanced scope support
- Enables reusable query logic
- Supports multi-tenant applications

## Implemented

- Per-query scope composition via `Scopes()` in `database/query/query_scopes.go`
- Scope functions are applied via `applyScopes()` and chained on cloned queries
- Integration tests for scopes in `integration_tests/cockroachdb/cockroachdb_query_scopes_test.go`

## Remaining Work

- Global scopes defined on the model and applied automatically to all queries
- Scope removal for specific queries (e.g. `WithoutScope`)

## Notes

- The `Scopes()` API accepts `func(contractsorm.Query) contractsorm.Query` functions and applies them on a cloned query, ensuring immutability of the original query.
- No `GlobalScopes()` method on models exists yet — this would require model interface changes or struct tag parsing.

## Proposed API

```go
// Global scope (applied to all queries)
func (u *User) GlobalScopes() []Scope {
    return []Scope{
        func(q Query) Query {
            return q.Where("deleted_at IS NULL")
        },
    }
}

// Scope composition
activeAdmins := db.Query().Model(&User{}).Scopes(Active, Admin).Find(&users)

// Dynamic scope parameters
func CreatedAfter(date time.Time) Scope {
    return func(q Query) Query {
        return q.Where("created_at > ?", date)
    }
}
```

## Implementation Details

- Global scopes defined on model
- Scope chaining and composition
- Dynamic scope parameters
- Scope removal for specific queries

## References

- Extracted from `docs/proposals/feature-requests.md` (item #3)
