# Feature Requests Proposal

**Date**: June 1, 2026
**Last Reviewed**: August 10, 2026
**Status**: Open for Discussion

## Overview

This document is an index of proposed feature improvements for Neat ORM. Each feature has been extracted into its own proposal file for detailed discussion. The proposals aim to enhance developer experience, close feature gaps with competing ORMs, and expand Neat's capabilities.

## Status Legend

- **Not Started** — no implementation exists; the proposal is fully relevant
- **Partially Done** — some sub-features are implemented; remaining work is listed in the proposal file
- **Completed** — implemented; the proposal has been moved to `completed/`
- **Rejected** — dropped from the roadmap with rationale

## Priority Matrix

Updated August 10, 2026 to reflect current implementation status.

| # | Feature | Status | Priority | Impact | Effort | ROI |
|---|---------|--------|----------|--------|--------|-----|
| 1 | [ManyToMany Relationships](many-to-many-relationships.md) | Not Started | High | High | High | High |
| 2 | [Query Caching](query-caching.md) | Not Started | Medium | Medium | Medium | Medium |
| 3 | [Advanced Scopes](advanced-scopes.md) | Partially Done | Medium | Medium | Low | High |
| 4 | [Better Error Messages](better-error-messages.md) | Partially Done | High | High | Low | Very High |
| 6 | [CLI Tools](cli-tools.md) | Not Started | Medium | High | Medium | High |
| 7 | [Query Optimization](query-optimization.md) | Partially Done | High | High | Medium | High |
| 8 | [Benchmark Suite](benchmark-suite.md) | Partially Done | Medium | Medium | Low | High |
| 11 | [Property-Based Testing](property-based-testing.md) | Not Started | Medium | Medium | Medium | Medium |
| 12 | [Snapshot Testing](snapshot-testing.md) | Not Started | Low | Medium | Low | Medium |
| 15 | [Database-Specific Features](database-specific-features.md) | Partially Done | Medium | High | High | Medium |
| 16 | [Event System](completed/event-system.md) | Completed | — | — | — | — |
| 17 | [Validation](validation.md) | Not Started | Medium | High | Medium | High |
| 20 | [Query Analyzer](query-analyzer.md) | Not Started | Medium | High | Medium | High |

### Rejected Features

- **Sequelize Compatibility (#9)**: Sequelize is a JavaScript/Node.js ORM. A compatibility layer in a Go ORM is not meaningful. Django-style sugar methods (`Filter`, `Exclude`) were implemented instead — see `docs/proposals/completed/sugar-methods-for-django-compatibility.md`.
- **TypeORM Compatibility (#10)**: TypeORM is a TypeScript ORM. Same reasoning as above — not applicable to a Go codebase.
- **IDE Support (#5)**: High effort for a Go ORM where `gopls` already provides strong autocomplete. A custom language server is a large undertaking with limited payoff versus core ORM work.
- **Admin Panel (#18)**: High effort with low ROI for a Go ORM. Better suited as a separate project. Existing tools (pgAdmin, Adminer, TablePlus) cover most admin needs.
- **Schema Inspector (#19)**: External tools (DBeaver, TablePlus, dbdiagram.io) already do this well. Medium effort for functionality that overlaps with mature existing tools.
- **Migration Guides (#13)**: Covered by the 18 side-by-side comparison pages in `docs/comparison/` and the completed sugar-methods proposals.
- **Interactive Examples (#14)**: Covered by the 27 runnable examples in `examples/`.

---

## Implementation Roadmap

Updated August 10, 2026. Completed and rejected items are excluded.

### Phase 1 (Quick Wins - Low Effort, High Impact)
- [Better Error Messages](better-error-messages.md) (remaining: SQL context in errors, execution time, stack traces)
- [Snapshot Testing](snapshot-testing.md)
- [Advanced Scopes](advanced-scopes.md) (remaining: global scopes on model)

### Phase 2 (Core Features - Medium Effort, High Impact)
- [ManyToMany Relationships](many-to-many-relationships.md)
- [CLI Tools](cli-tools.md)
- [Query Optimization](query-optimization.md) (remaining: N+1 batching, join optimization)
- [Validation](validation.md)
- [Query Analyzer](query-analyzer.md)

### Phase 3 (Enhanced Experience - Medium/High Effort)
- [Query Caching](query-caching.md)
- [Benchmark Suite](benchmark-suite.md) (remaining: GORM comparison, CI regression detection)
- [Database-Specific Features](database-specific-features.md) (remaining: JSONB arrays, FTS5, spatial)
- [Property-Based Testing](property-based-testing.md)

---

## Open Questions

1. Should we prioritize ManyToMany over Query Caching? — **Still open. ManyToMany remains the highest-impact missing feature.**
2. Should CLI tools be part of core or a separate package? — **Still open.**
3. What validation library should we use or build our own? — **Still open.**
4. Should we support multiple cache backends out of the box? — **Still open.**

---

## References

- GORM Documentation: https://gorm.io/docs/
- Django ORM Documentation: https://docs.djangoproject.com/en/stable/topics/db/queries/
- Laravel Eloquent Documentation: https://laravel.com/docs/eloquent
