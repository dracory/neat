# Advanced Scopes

**Date**: June 1, 2026
**Last Reviewed**: August 25, 2026
**Status**: Completed
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

- Global scopes defined on models via the `GlobalScope` interface (`GlobalScopes() []func(Query) Query`) and automatically registered during `Model()` resolution.
- Per-query scope composition via `Scopes()` in `database/query/query_scopes.go`.
- Scope removal methods:
  - `WithoutScope(funcs ...func(Query) Query)` for disabling specific scope functions.
  - `WithoutScopes()` for disabling all per-query scope functions.
  - `WithoutGlobalScopes()` for disabling model-level global scopes for a query.
- Scope functions are applied via `applyScopes()` and chained on cloned queries.
- Comprehensive unit and integration tests for scopes in `database/query/query_scopes_test.go` and database driver integration tests.

## Remaining Work

None.

## Notes

- The `Scopes()` API accepts `func(contractsorm.Query) contractsorm.Query` functions and applies them on a cloned query, ensuring immutability of the original query.
- Models implement `GlobalScope` by defining a `GlobalScopes() []func(contractsorm.Query) contractsorm.Query` method.

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
