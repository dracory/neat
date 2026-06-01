# Neat ORM vs go-pg Developer Experience Comparison

**Date**: June 1, 2026

## Overview

This document compares the developer experience, ease of use, and features of Neat ORM with go-pg, a PostgreSQL-focused ORM for Go.

**Important Note**: go-pg is in maintenance mode and will NOT receive new features. The official recommendation is to use Bun Golang ORM instead.

## Why Neat Wins?

**Neat ORM wins when you need multi-database support, active development, and built-in testing tools rather than PostgreSQL-specific features in maintenance mode.** While go-pg offers excellent PostgreSQL-specific features like arrays and hstore, it's in maintenance mode with no new features. Neat provides active development, multi-database support including SQL Server and Turso, and comprehensive testing infrastructure with factories and seeders. If you value active development, database flexibility, and built-in testing tools over PostgreSQL-specific optimizations, Neat is the better choice.

## go-pg Overview

go-pg is a PostgreSQL client and ORM for Go with focus on PostgreSQL features. It supports PostgreSQL-specific features like arrays, hstore, composite types, and has good performance.

## Architecture Comparison

### Neat ORM
- **Multi-database**: Supports MySQL, PostgreSQL, SQLite, SQL Server, Turso
- **Laravel Eloquent-like API**: Familiar syntax for developers coming from PHP/Laravel
- **Runtime query builder**: Fluent interface for building queries at runtime
- **Feature-rich**: Associations, observers, migrations, factories, seeders
- **Active development**: Recent development with security focus

### go-pg
- **PostgreSQL-only**: Focused exclusively on PostgreSQL
- **PostgreSQL features**: Arrays, hstore, composite types
- **Maintenance mode**: No new features being added
- **Performance-focused**: Optimized for PostgreSQL performance
- **Mature codebase**: Established but in maintenance

## Developer Experience

### Neat ORM
- **Laravel Eloquent-like API**: Familiar to PHP/Laravel developers
- **Multi-database support**: Works with multiple databases
- **Clean fluent interface**: Intuitive method chaining
- **Active development**: Regular updates and improvements
- **Comprehensive documentation**: Docs with examples
- **Testing tools**: Built-in factories and seeders

### go-pg
- **PostgreSQL expertise**: Deep PostgreSQL feature support
- **Performance**: Good performance for PostgreSQL
- **PostgreSQL-specific**: Arrays, hstore, composite types
- **Maintenance mode**: No new features
- **Mature documentation**: Established docs
- **Limited scope**: PostgreSQL only

## Feature Comparison

| Feature | Neat ORM | go-pg |
|---------|----------|-------|
| **Database Support** | MySQL, PostgreSQL, SQLite, SQL Server, Turso | PostgreSQL only |
| **Query Builder** | Fluent Eloquent-like API | PostgreSQL-focused API |
| **PostgreSQL Features** | Basic support | Arrays, hstore, composite types |
| **Migrations** | Built-in migration system | Migrations available |
| **Associations** | BelongsTo, HasMany, HasOne | Has one, belongs to, has many, many to many |
| **Soft Deletes** | Built-in soft delete support | Not included |
| **Observers** | Model lifecycle observers | Not included |
| **Factories/Seeders** | Built-in factories and seeders | Not included |
| **Development Status** | Active development | Maintenance mode |
| **Performance** | Good | Optimized for PostgreSQL |

## Strengths

### Neat ORM
- Multi-database support
- Laravel Eloquent-like API
- Active development
- Built-in migrations, factories, seeders
- Model lifecycle observers
- Recent security hardening
- Comprehensive documentation

### go-pg
- Deep PostgreSQL feature support
- PostgreSQL-specific optimizations
- Good performance
- Mature codebase
- Arrays, hstore, composite types
- PostgreSQL expertise

## Weaknesses

### Neat ORM
- Less PostgreSQL-specific features
- Newer codebase
- Smaller community

### go-pg
- **Maintenance mode** (no new features)
- PostgreSQL only (no multi-database)
- No factories or seeders
- No observers
- No soft deletes
- Official recommendation to use Bun

## Use Case Recommendations

### Choose Neat ORM If:
- You need multi-database support
- You prefer Laravel Eloquent-like API
- You need built-in migrations, factories, and seeders
- You need model lifecycle observers
- You prefer active development
- You might switch databases in the future
- Recent security hardening is important

### Choose go-pg If:
- You're committed to PostgreSQL only
- You need PostgreSQL-specific features (arrays, hstore)
- You need PostgreSQL optimizations
- You're maintaining existing go-pg code
- You don't need new features
- Performance is critical for PostgreSQL

**Important**: If starting a new project, consider Bun (the recommended successor to go-pg) instead of go-pg.

## Conclusion

Neat ORM and go-pg serve different purposes:

**Neat ORM**: Multi-database ORM with Laravel Eloquent-like API and active development. Ideal for projects that need database flexibility and modern ORM features.

**go-pg**: PostgreSQL-focused ORM with deep PostgreSQL feature support but in maintenance mode. Ideal for existing PostgreSQL projects that need PostgreSQL-specific features.

**Recommendation**: For new projects, consider Bun (the official successor to go-pg) rather than go-pg due to its maintenance status.

## References

- Neat ORM Documentation: See `docs/` directory
- go-pg Documentation: https://pg.uptrace.dev/
- go-pg GitHub: https://github.com/go-pg/pg
- Bun ORM (go-pg successor): https://bun.uptrace.dev/
- Laravel Eloquent Documentation: https://laravel.com/docs/eloquent
