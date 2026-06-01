# Neat ORM vs upper/db Developer Experience Comparison

**Date**: June 1, 2026

## Overview

This document compares the developer experience, ease of use, and features of Neat ORM with upper/db, a productive data access layer (DAL) for Go that provides database-agnostic tools.

## Why Neat Wins?

**Neat ORM wins when you need a full ORM with associations, migrations, factories, and seeders rather than a data access layer with limited ORM features.** upper/db is a DAL with ORM convenience but lacks full ORM features like associations, migrations, and testing tools. Neat provides a complete ORM solution with BelongsTo, HasMany, HasOne associations, built-in migrations, factories, and seeders. If you need a full-featured ORM with all standard capabilities rather than a DAL with limited ORM features, Neat is the better choice.

## upper/db Overview

upper/db is a data access layer (DAL) for Go that provides agnostic tools to work with different data sources. It aims to provide the convenience of an ORM while allowing SQL usage when needed. It's described as a DAL rather than a full ORM.

## Architecture Comparison

### Neat ORM
- **Full ORM**: Complete ORM with models, relationships, migrations
- **Laravel Eloquent-like API**: Familiar syntax for developers coming from PHP/Laravel
- **Runtime query builder**: Fluent interface for building queries at runtime
- **Feature-rich**: Associations, observers, migrations, factories, seeders
- **Modern design**: Recent development with security focus

### upper/db
- **Data Access Layer**: Productive DAL (not a full ORM)
- **Database-agnostic**: Works with different data sources
- **ORM convenience**: ORM-like convenience with SQL flexibility
- **Multi-database**: Support for PostgreSQL, MySQL, SQLite, and more
- **Mature codebase**: Established library

## Developer Experience

### Neat ORM
- **Laravel Eloquent-like API**: Familiar to PHP/Laravel developers
- **Full ORM experience**: Complete ORM with all standard features
- **Clean fluent interface**: Intuitive method chaining
- **Active development**: Regular updates and improvements
- **Comprehensive documentation**: Docs with examples
- **Testing tools**: Built-in factories and seeders

### upper/db
- **Productive DAL**: Focus on business logic without manual SQL
- **Database-agnostic**: Works with multiple databases
- **ORM convenience**: ORM-like features
- **SQL flexibility**: Can use SQL when needed
- **Mature codebase**: Established library
- **Simpler learning curve**: DAL approach

## Feature Comparison

| Feature | Neat ORM | upper/db |
|---------|----------|----------|
| **Type** | Full ORM | Data Access Layer (DAL) |
| **Query Builder** | Fluent Eloquent-like API | Database-agnostic API |
| **Database Support** | MySQL, PostgreSQL, SQLite, SQL Server, Turso | PostgreSQL, MySQL, SQLite, and more |
| **Associations** | BelongsTo, HasMany, HasOne | Limited association support |
| **Migrations** | Built-in migration system | Not included |
| **Soft Deletes** | Built-in soft delete support | Not included |
| **Observers** | Model lifecycle observers | Not included |
| **Factories/Seeders** | Built-in factories and seeders | Not included |
| **SQL Flexibility** | RawExpr for raw SQL | Can use SQL when needed |
| **Scope** | Complete ORM solution | DAL with ORM convenience |

## Strengths

### Neat ORM
- Laravel Eloquent-like API (familiar to many)
- Full ORM with all standard features
- Built-in migrations, factories, and seeders
- Associations and relationships
- Model lifecycle observers
- Soft deletes support
- Active development
- Recent security hardening

### upper/db
- Database-agnostic (works with many databases)
- ORM convenience with SQL flexibility
- Productive (focus on business logic)
- Mature codebase
- Can use SQL when needed
- Simpler than full ORM
- Good for multiple data sources

## Weaknesses

### Neat ORM
- More complex (full ORM)
- Runtime overhead from features
- Newer codebase
- Smaller community

### upper/db
- Not a full ORM (DAL approach)
- Limited association support
- No migrations
- No factories or seeders
- No observers or hooks
- No soft deletes
- Less ORM features

## Use Case Recommendations

### Choose Neat ORM If:
- You need a full ORM with all standard features
- You need associations and relationships
- You need built-in migrations
- You need factories and seeders for testing
- You prefer Laravel Eloquent-like API
- You need model lifecycle observers
- You need soft deletes
- You want a complete ORM solution

### Choose upper/db If:
- You need database-agnostic data access
- You want ORM convenience with SQL flexibility
- You work with multiple data sources
- You prefer a DAL over full ORM
- You don't need full ORM features
- You want to focus on business logic
- You need to use SQL occasionally

## Conclusion

Neat ORM and upper/db serve different purposes:

**Neat ORM**: Full-featured ORM with Laravel Eloquent-like API and complete ORM features. Ideal for developers building full applications who need a comprehensive ORM.

**upper/db**: Database-agnostic data access layer with ORM convenience and SQL flexibility. Ideal for developers who need productive data access without full ORM complexity.

The choice depends on your needs:
- **Full ORM with all features**: Neat ORM
- **Database-agnostic DAL with ORM convenience**: upper/db

## References

- Neat ORM Documentation: See `docs/` directory
- upper/db Documentation: https://upper.io/
- upper/db GitHub: https://github.com/upper/db
- Laravel Eloquent Documentation: https://laravel.com/docs/eloquent
