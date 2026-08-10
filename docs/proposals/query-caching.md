# Query Caching

**Date**: June 1, 2026
**Last Reviewed**: August 10, 2026
**Status**: Not Started
**Priority**: Medium
**Impact**: Medium
**Effort**: Medium
**ROI**: Medium

## Description

Add query result caching with TTL support.

## Rationale

- GORM has built-in caching support
- Improves performance for frequently accessed data
- Reduces database load

## Proposed API

```go
// Cache query for 5 minutes
db.Query().Model(&User{}).Cache(5*time.Minute).Where("id = ?", 1).First(&user)

// Cache with custom key
db.Query().Model(&User{}).Cache("user:1", 10*time.Minute).Where("id = ?", 1).First(&user)

// Invalidate cache
db.Query().Model(&User{}).CacheInvalidate("user:1")
```

## Implementation Details

- Pluggable cache backend (Redis, in-memory)
- Automatic cache invalidation on model updates
- Cache tags for bulk invalidation
- Configurable default TTL

## References

- Extracted from `docs/proposals/feature-requests.md` (item #2)
