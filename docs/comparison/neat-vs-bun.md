# Neat ORM vs Bun Developer Experience Comparison

**Date**: June 1, 2026

## Overview

This document compares the developer experience, ease of use, and features of Neat ORM with Bun, a SQL-first Golang ORM that emphasizes performance and minimal overhead.

## Why Neat Wins?

**Neat ORM wins when you prefer Laravel Eloquent-like API and need multi-database support including SQLite, SQL Server, and Turso rather than SQL-first PostgreSQL/MySQL focus.** While Bun offers excellent performance and compile-time type safety, it's limited to PostgreSQL and MySQL. Neat provides a more familiar Laravel-like API, supports more databases including SQL Server and Turso, and includes built-in migrations, factories, and seeders. If you value Laravel patterns, database flexibility, and comprehensive testing tools over SQL-first performance optimization, Neat is the better choice.

## Bun Overview

Bun is a SQL-first Golang ORM for PostgreSQL and MySQL that is built on database/sql APIs. It emphasizes minimal overhead over raw SQL, type-safe operations, and rich relationships. It's the recommended successor to go-pg.

## Architecture Comparison

### Neat ORM
- **Laravel Eloquent-like API**: Familiar syntax for developers coming from PHP/Laravel
- **Runtime query builder**: Fluent interface for building queries at runtime
- **Multi-database**: MySQL, PostgreSQL, SQLite, SQL Server, Turso
- **Feature-rich**: Associations, observers, migrations, factories, seeders
- **ORM-focused**: Traditional ORM pattern

### Bun
- **SQL-first**: Doesn't hide SQL, embraces it
- **Performance-optimized**: Minimal overhead over raw SQL
- **PostgreSQL and MySQL**: Focused on these two databases
- **Type-safe operations**: Leverages Go's static typing
- **Production-ready**: Migrations, fixtures, soft deletes, OpenTelemetry

## Developer Experience

### Neat ORM
- **Laravel Eloquent-like API**: Familiar to PHP/Laravel developers
- **Multi-database support**: Works with multiple databases
- **Clean fluent interface**: Intuitive method chaining
- **Active development**: Regular updates and improvements
- **Comprehensive documentation**: Docs with examples
- **Testing tools**: Built-in factories and seeders

### Bun
- **SQL-first approach**: Embraces SQL rather than hiding it
- **Type-safe**: Compile-time type safety
- **Performance-focused**: Minimal overhead
- **Rich relationships**: Complex table relationships
- **Production features**: Migrations, fixtures, soft deletes
- **OpenTelemetry support**: Observability built-in

## Feature Comparison

| Feature | Neat ORM | Bun |
|---------|----------|-----|
| **Query Builder** | Fluent Eloquent-like API | SQL-first API |
| **Database Support** | MySQL, PostgreSQL, SQLite, SQL Server, Turso | PostgreSQL, MySQL |
| **Type Safety** | Runtime type checking | Compile-time type safety |
| **Performance** | Good with minimal overhead | Excellent (minimal overhead) |
| **Migrations** | Built-in migration system | SQL-based migrations |
| **Associations** | BelongsTo, HasMany, HasOne | Rich relationships |
| **Soft Deletes** | Built-in soft delete support | Built-in soft delete support |
| **Observers** | Model lifecycle observers | Query hooks |
| **Factories/Seeders** | Built-in factories and seeders | Fixtures support |
| **OpenTelemetry** | Not included | Built-in support |
| **SQL Approach** | ORM abstraction | SQL-first |

## Strengths

### Neat ORM
- Laravel Eloquent-like API (familiar to many)
- Multi-database support
- Built-in migrations, factories, and seeders
- Model lifecycle observers
- Active development
- Recent security hardening
- Comprehensive documentation

### Bun
- SQL-first approach (embraces SQL)
- Excellent performance (minimal overhead)
- Compile-time type safety
- Rich relationships
- Production-ready features
- OpenTelemetry support
- Successor to go-pg (mature codebase)

## Weaknesses

### Neat ORM
- Runtime type checking (less compile-time safety)
- Less performance optimization focus
- Newer codebase
- Smaller community
- No OpenTelemetry support

### Bun
- Limited database support (PostgreSQL, MySQL only)
- SQL-first may be less familiar to ORM users
- No factories/seeders (has fixtures)
- No observers (has query hooks)
- Less Laravel-like API

## Use Case Recommendations

### Choose Neat ORM If:
- You prefer Laravel Eloquent-like API
- You need multi-database support (SQLite, SQL Server, Turso)
- You need built-in factories and seeders
- You need model lifecycle observers
- You're coming from PHP/Laravel background
- You prefer traditional ORM abstraction
- Recent security hardening is important

### Choose Bun If:
- You prefer SQL-first approach
- You need excellent performance
- You need compile-time type safety
- You only use PostgreSQL or MySQL
- You need OpenTelemetry support
- You want rich relationships
- You're migrating from go-pg
- You value minimal SQL overhead

## Conclusion

Neat ORM and Bun represent different philosophies:

**Neat ORM**: Traditional ORM with Laravel Eloquent-like API and multi-database support. Ideal for developers who want familiar ORM abstraction and database flexibility.

**Bun**: SQL-first ORM with excellent performance and compile-time type safety. Ideal for developers who embrace SQL and need performance optimization for PostgreSQL/MySQL.

The choice depends on your philosophy:
- **ORM abstraction with database flexibility**: Neat ORM
- **SQL-first with performance focus**: Bun

Both are excellent choices but for different development styles and needs.

## References

- Neat ORM Documentation: See `docs/` directory
- Bun Documentation: https://bun.uptrace.dev/
- Bun GitHub: https://github.com/uptrace/bun
- Laravel Eloquent Documentation: https://laravel.com/docs/eloquent
