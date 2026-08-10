# Query Analyzer

**Date**: June 1, 2026
**Last Reviewed**: August 10, 2026
**Status**: Not Started
**Priority**: Medium
**Impact**: High
**Effort**: Medium
**ROI**: High

## Description

Add query analysis and optimization recommendations.

## Rationale

- Performance tuning
- Slow query detection
- Best practices enforcement

## Proposed Features

```go
// EXPLAIN query analysis
analyzer := db.Query().Model(&User{}).Where("name = ?", "John").Explain()
analyzer.Suggestions() // "Add index on name column"

// Slow query detection
db.Query().SetSlowQueryThreshold(100 * time.Millisecond)
db.Query().On("slowQuery", func(q Query, duration time.Duration) {
    log.Printf("Slow query: %s took %v", q.ToSQL(), duration)
})
```

## Implementation Details

- EXPLAIN parsing
- Index recommendations
- Query pattern detection
- Performance metrics

## References

- Extracted from `docs/proposals/feature-requests.md` (item #20)
