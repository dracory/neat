# Event System

**Date**: June 1, 2026
**Last Reviewed**: August 10, 2026
**Status**: Completed
**Priority**: Medium
**Impact**: Medium
**Effort**: Medium
**ROI**: Medium

## Description

Add comprehensive event system for model and query lifecycle.

## Rationale

- Hooks for business logic
- Audit logging
- Cache invalidation
- Notifications

## Implemented

- Full observer/event system in `database/observer/` with `Event` and `Dispatcher`
- Event types: `creating`, `created`, `updating`, `updated`, `saving`, `saved`, `deleting`, `deleted`, `force_deleting`, `force_deleted`, `restoring`, `restored`, `retrieved`
- `DispatchesEvents` contract on models for per-event handlers
- `Observer` interface with lifecycle hooks (`Creating`, `Created`, `Saving`, `Saved`, etc.)
- `WithoutEvents()` to disable event firing for a query
- Event carries context, model, original/dirty attributes, and the query instance

## References

- Extracted from `docs/proposals/feature-requests.md` (item #16)
