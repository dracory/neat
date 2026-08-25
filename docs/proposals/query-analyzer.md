# Query Analyzer

**Date**: June 1, 2026
**Last Reviewed**: August 25, 2026
**Status**: Partially Done
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

## Implemented

- **Slow query detection** — `SlowThreshold` field on `DBConfig` (in milliseconds) configures the threshold. When a query exceeds it, a warning is emitted via the query helpers (`database/query/query_helpers.go`). See also the `Debug` toggle on `DBConfig`.
- **DSN redaction** — `redactDSN()` in `database/db.go` strips credentials from DSN strings for safe logging.
- **Query validation** — `validate()` in `database/query/query_helpers.go` checks for nil DB, missing table, and build errors before executing terminal methods.

## Not Yet Implemented

- EXPLAIN parsing and index recommendations
- Query pattern detection
- Performance metrics aggregation
- The `.Explain()` and `.On("slowQuery", ...)` API shown below is still proposed, not implemented

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
