# Query Optimization

**Date**: June 1, 2026
**Last Reviewed**: August 10, 2026
**Status**: Partially Done
**Priority**: High
**Impact**: High
**Effort**: Medium
**ROI**: High

## Description

Add automatic query optimization features.

## Rationale

- Prevent N+1 query problems
- Improve eager loading efficiency
- Connection pool recommendations

## Implemented

- Eager loading via `With()` in `database/query/query_relations.go`
- Nested eager loading (e.g. `With("Posts.Comments")`)
- `Without()` to remove relations from eager loading
- Connection pool configuration on the database layer

## Remaining Work

- Automatic query batching to prevent N+1 across collections
- Automatic join optimization (collapse `With` into a single query where possible)
- Connection pool health checks and recommendations
- Performance monitoring hooks

## Proposed Features

```go
// Automatic query batching
db.Query().Model(&User{}).With("Posts.Comments").Find(&users)
// Automatically batches queries instead of N+1

// Eager loading optimization
db.Query().Model(&User{}).With("Posts", "Profile").Find(&users)
// Optimizes to single query with joins where possible

// Connection pool recommendations
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(10)
db.SetConnMaxLifetime(5 * time.Minute)
```

## Implementation Details

- Query batching for eager loading
- Automatic join optimization
- Connection pool health checks
- Performance monitoring

## References

- Extracted from `docs/proposals/feature-requests.md` (item #7)
