# CLI Tools

**Date**: June 1, 2026
**Last Reviewed**: August 25, 2026
**Status**: Not Started
**Priority**: Medium
**Impact**: High
**Effort**: Medium
**ROI**: High

## Description

Create command-line tools for common ORM operations.

## Rationale

- Streamlines development workflow
- Consistent with Laravel Artisan
- Reduces manual setup

## Proposed Commands

```bash
# Migrations
neat migrate:status
neat migrate:run
neat migrate:rollback
neat migrate:refresh
neat migrate:fresh

# Seeders
neat seed
neat seed:run UserSeeder
neat seed:rollback

# Database
neat db:reset
neat db:drop
neat db:create

# Models
neat make:model User
neat make:migration create_users_table
neat make:seeder UserSeeder
```

## Implementation Details

- Cobra-based CLI
- Interactive mode
- Configuration file support
- Environment variable support

## References

- Extracted from `docs/proposals/feature-requests.md` (item #6)
