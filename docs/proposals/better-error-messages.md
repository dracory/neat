# Better Error Messages

**Date**: June 1, 2026
**Last Reviewed**: August 25, 2026
**Status**: Partially Done
**Priority**: High
**Impact**: High
**Effort**: Low
**ROI**: Very High

## Description

Improve error messages with SQL context and debugging information.

## Rationale

- Current errors can be cryptic
- Hard to debug complex queries
- Improves developer productivity

## Implemented

- Runtime debug toggle via `EnableDebug()` / `DisableDebug()` / `IsDebug()` in `database/query/query_debug.go`
- Error sanitization that strips SQL details in production and logs full errors in debug mode (`database/query/error_sanitizer.go`)
- Thread-safe debug state
- `validate()` method in `database/query/query_helpers.go` provides structured error messages for common issues (nil DB, missing table, build errors)
- `SlowThreshold` field on `DBConfig` for slow query detection with warning emission
- `redactDSN()` in `database/db.go` strips credentials from DSN strings in error messages

## Remaining Work

- Embed SQL query context and bound parameter values directly in returned error messages (currently only logged)
- Execution time tracking in error messages
- Stack traces for query errors

## Proposed Improvements

```go
// Before: "SQL error: near 'WHERE'"
// After: "SQL error: near 'WHERE' in query: SELECT * FROM users WHERE name = ? AND"

// Debug mode
db.Debug().Query().Model(&User{}).Where("name = ?", "John").First(&user)
// Output: [SQL] SELECT * FROM users WHERE name = 'John' LIMIT 1 [0.5ms]
```

## Implementation Details

- SQL query context in errors
- Parameter values in error messages
- Debug mode with query logging
- Execution time tracking
- Stack traces for query errors

## References

- Extracted from `docs/proposals/feature-requests.md` (item #4)
