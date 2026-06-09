# Runtime Debug Toggle Proposal

**Date**: June 8, 2026
**Status**: Completed
**Priority**: Medium

## Overview

This proposal adds runtime debug toggle methods to Neat ORM, allowing developers to enable/disable detailed SQL error messages without restarting the application. This addresses the current limitation where debug mode can only be set at initialization via `DBConfig.Debug`.

## Problem Statement

### Current Issues
- **No runtime debug control**: Debug mode can only be set via `DBConfig.Debug` at initialization
- **Production debugging difficulty**: When encountering "database operation failed" errors in production, developers must restart the application with debug mode enabled to see the actual SQL error
- **Poor developer experience**: AI assistants and developers cannot deduce the real error from the generic "database operation failed" message
- **Security vs usability tradeoff**: The current error sanitization is too aggressive, hiding all SQL errors in production even from logs

### Impact
- Difficult to debug database errors in production environments
- AI assistants cannot provide accurate troubleshooting advice
- Developers must modify code and redeploy to enable debug mode
- Increased time to resolve database-related issues

## Proposed Solution

### 1. Runtime Debug Toggle Methods

Add methods to enable/disable debug mode at runtime:

```go
// Database methods
func (d *Database) EnableDebug()
func (d *Database) DisableDebug()
func (d *Database) IsDebug() bool

// Orm methods
func (o *Orm) EnableDebug()
func (o *Orm) DisableDebug()
func (o *Orm) IsDebug() bool

// Query methods
func (q *Query) EnableDebug()
func (q *Query) DisableDebug()
func (q *Query) IsDebug() bool
```

### 2. Thread-Safe Implementation

Use atomic operations or mutex to ensure thread-safe debug state changes:

```go
type Query struct {
    // ... existing fields
    debugMu    sync.RWMutex
    debugState bool
}

func (q *Query) EnableDebug() {
    q.debugMu.Lock()
    defer q.debugMu.Unlock()
    q.debugState = true
    q.dbConfig.Debug = true
}

func (q *Query) DisableDebug() {
    q.debugMu.Lock()
    defer q.debugMu.Unlock()
    q.debugState = false
    q.dbConfig.Debug = false
}

func (q *Query) isDebugEnabled() bool {
    q.debugMu.RLock()
    defer q.debugMu.RUnlock()
    if q.debugState {
        return true
    }
    if q.dbConfig != nil {
        return q.dbConfig.Debug
    }
    return false
}
```

### 3. Enhanced Error Sanitization

Improve error sanitization as a method on Query that logs full errors when debug is enabled:

```go
func (q *Query) sanitizeError(err error) error {
    if err == nil || q.isDebugEnabled() {
        return err
    }

    // Never suppress context errors
    if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
        return err
    }

    // Log full error when debug is enabled
    if q.log != nil && q.isDebugEnabled() {
        q.log.Errorf("Database error: %v", err)
    }

    // In production (debug disabled), return generic error to users
    errMsg := err.Error()
    if strings.Contains(strings.ToLower(errMsg), "sql") ||
       strings.Contains(strings.ToLower(errMsg), "query") ||
       strings.Contains(strings.ToLower(errMsg), "syntax") {
        return fmt.Errorf("database operation failed")
    }

    return err
}
```

### 4. Environment Variable Support

Allow debug mode to be controlled via environment variable:

```go
// In DBConfig initialization
debugEnv := strings.ToLower(os.Getenv("NEAT_DEBUG"))
if debugEnv == "true" || debugEnv == "1" {
    cfg.Debug = true
}
```

## Implementation Plan

### Phase 1: Core Methods
1. Add `EnableDebug()`, `DisableDebug()`, `IsDebug()` methods to `Query` struct
2. Implement thread-safe state management with `sync.RWMutex`
3. Add `isDebugEnabled()` method to check debug state
4. Add unit tests for thread-safety

### Phase 2: Orm and Database Methods
1. Add debug toggle methods to `Orm` struct
2. Add debug toggle methods to `Database` struct
3. Propagate debug state changes to underlying queries
4. Add integration tests

### Phase 3: Enhanced Error Sanitization
1. Update `sanitizeError()` to be a method on Query
2. Add logging of full errors only when debug is enabled
3. Update all call sites to use `q.sanitizeError(err)`
4. Add tests for error logging behavior

### Phase 4: Environment Variable Support
1. Add environment variable reading in config initialization
2. Update documentation with environment variable usage
3. Add examples showing environment-based debug control

## Alternatives Considered

### Alternative 1: Logging Only (No Runtime Toggle)
- **Pros**: Simpler implementation, maintains security
- **Cons**: Still requires code changes to enable, doesn't solve immediate debugging need
- **Rejected**: Doesn't address the core need for runtime control

### Alternative 2: Error Codes Instead of Sanitization
- **Pros**: Provides structured error information without exposing SQL
- **Cons**: Complex to implement, requires mapping all database errors to codes
- **Rejected**: Too complex for the benefit gained

### Alternative 3: Separate Debug Database Instance
- **Pros**: Complete isolation of debug vs production
- **Cons**: Requires infrastructure changes, resource overhead
- **Rejected**: Overkill for this use case

## Benefits

- **Improved debugging**: Developers can enable debug mode without restarting
- **Better AI assistance**: AI can see actual errors when debug is enabled
- **Production safety**: Debug mode can be disabled quickly if needed
- **Selective logging**: Full errors only logged when debug is enabled (security)
- **Backward compatible**: Existing code continues to work without changes

## Risks and Mitigations

### Risk: Accidentally leaving debug enabled in production
- **Mitigation**: Add health check endpoint to report debug status
- **Mitigation**: Add monitoring/alerting for debug mode in production
- **Mitigation**: Document best practices for debug mode usage

### Risk: Thread-safety issues with runtime state changes
- **Mitigation**: Use proper locking mechanisms (sync.RWMutex)
- **Mitigation**: Comprehensive unit tests for concurrent access
- **Mitigation**: Code review focused on concurrency patterns

### Risk: Performance overhead from locking
- **Mitigation**: Use RWMutex for read-heavy operations
- **Mitigation**: Benchmark to ensure minimal performance impact
- **Mitigation**: Consider atomic operations if appropriate

### Risk: Logging sensitive data in production
- **Mitigation**: Only log when debug is explicitly enabled
- **Mitigation**: Document that debug mode should not be enabled in production
- **Mitigation**: Consider adding separate "always log" config if needed

## Migration Guide

### For Existing Applications
No changes required - existing applications continue to work as before.

### To Use Runtime Debug Toggle
```go
// Enable debug mode when needed
db.Orm().EnableDebug()

// Execute query that was failing
err := db.Table("users").Create(&user)

// Disable debug mode
db.Orm().DisableDebug()
```

### To Use Environment Variable
```bash
# Set environment variable
export NEAT_DEBUG=true

# Run application
./your-app
```

## Testing Strategy

### Unit Tests
- Test `EnableDebug()`/`DisableDebug()` methods
- Test `IsDebug()` returns correct state
- Test thread-safety with concurrent calls
- Test `isDebugEnabled()` respects runtime state

### Integration Tests
- Test debug toggle affects error sanitization
- Test logging behavior with debug enabled/disabled
- Test environment variable configuration
- Test debug state propagation through Orm/Database layers

### Manual Testing
- Enable debug mode in running application
- Verify SQL errors are shown
- Disable debug mode
- Verify errors are sanitized
- Check logs contain full error details only when debug is enabled

## Documentation Updates

- Add runtime debug toggle methods to API reference
- Update error handling documentation
- Add examples of debug toggle usage
- Document environment variable support
- Add security considerations for debug mode

## Open Questions

1. Should debug toggle be scoped per-connection or global?
   - **Recommendation**: Global for simplicity, can add per-connection later if needed

2. Should we add a middleware to automatically disable debug after timeout?
   - **Recommendation**: No, let developers control explicitly

3. Should we add audit logging for debug mode changes?
   - **Recommendation**: Yes, log when debug mode is enabled/disabled in production

## Success Criteria

- [x] Runtime debug toggle methods implemented and tested
- [x] Thread-safety verified with concurrent access tests
- [x] Error sanitization enhanced with conditional logging
- [x] Environment variable support added
- [x] Documentation updated
- [x] Backward compatibility maintained
- [ ] Performance impact < 1% overhead
- [x] All tests passing
