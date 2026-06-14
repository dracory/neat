# Neat ORM vs Goravel Eloquent Developer Experience Comparison

**Date**: June 1, 2026

## Overview

This document compares the developer experience, ease of use, and features of Neat ORM with Goravel's ORM layer (which uses GORM as its underlying implementation).

## Why Neat Wins?

**Neat ORM wins when you need a standalone ORM library without framework dependencies and want a focused database solution rather than a full-stack framework.** Neat provides a clean, focused ORM that can be used in any Go project without pulling in an entire framework's worth of dependencies. While Goravel offers a complete web framework with auth, routing, and more, Neat offers the flexibility to integrate with any existing project or framework. If you value modularity, minimal dependencies, and the ability to use the ORM independently over an all-in-one framework solution, Neat is the better choice.

## Key Distinction

**Important**: Goravel is not an ORM itself - it's a full-stack web framework that uses **GORM** as its underlying ORM layer. This comparison is between Neat ORM (standalone) and Goravel's GORM-based ORM implementation.

## Neat ORM

### Architecture
- Standalone ORM library built from scratch
- No GORM dependencies
- Designed for feature parity with Laravel's Eloquent ORM
- Clean, focused API design

### Developer Experience
- **Laravel Eloquent-like API**: Familiar syntax for developers coming from PHP/Laravel
- **Clean fluent interface**: Intuitive method chaining
- **Focused feature set**: Essential ORM features without bloat
- **Clear documentation**: Comprehensive docs with examples
- **Quick learning curve**: Easy to get started for Eloquent users
- **No framework dependencies**: Use as standalone library

### Key Features
- **Query Builder**: Fluent Eloquent-like API
- **ORM Features**: Models, relationships, migrations
- **Database Support**: MySQL, PostgreSQL, SQLite, SQL Server, Turso
- **Migrations**: Built-in migration system
- **Associations**: BelongsTo, HasMany, HasOne
- **Soft Deletes**: Built-in soft delete support
- **Transactions**: Transaction support
- **Observers**: Model lifecycle observers
- **Factories/Seeders**: Built-in factories and seeders
- **Connection Pooling**: Configurable pooling
- **Context Support**: Full context.Context support

## Goravel (GORM)

### Architecture
- Full-stack web framework
- Uses GORM as underlying ORM layer
- Inherits GORM's features and ecosystem
- Laravel-inspired framework for Go

### Developer Experience
- **Full-stack framework**: ORM, routing, auth, and more
- **Mature ecosystem**: Extensive GORM community and plugins
- **Laravel-like**: Similar patterns to Laravel framework
- **Comprehensive documentation**: Framework-wide documentation
- **Large community**: Active development and support
- **Production-ready**: Battle-tested in production

### Key Features
- **GORM ORM**: Mature, feature-rich ORM
- **Query Builder**: Chainable GORM API
- **ORM Features**: Models, relationships, hooks
- **Database Support**: Wide database support via GORM
- **Migrations**: Auto-migration and third-party tools
- **Associations**: All relationship types including ManyToMany
- **Soft Deletes**: Built-in soft delete support
- **Transactions**: Transaction support
- **Auth System**: Built-in authentication
- **Routing**: HTTP routing
- **Middleware**: Middleware support
- **Validation**: Request validation

## Comparison Summary

| Aspect | Neat ORM | Goravel (GORM) |
|--------|----------|----------------|
| **Type** | Standalone ORM | Full framework with GORM ORM |
| **Scope** | Database operations only | Full-stack web framework |
| **API Style** | Laravel Eloquent-like | Laravel framework-like |
| **Query Builder** | Fluent Eloquent-like API | GORM chainable API |
| **Database Support** | MySQL, PostgreSQL, SQLite, SQL Server, Turso, Oracle | Wide support via GORM |
| **Migrations** | Built-in migration system | Auto-migration and tools |
| **Associations** | BelongsTo, HasMany, HasOne, PolymorphicBelongsTo, PolymorphicHasMany | All relationship types |
| **Factories/Seeders** | Built-in factories and seeders | Available via plugins |
| **Auth System** | Not included | Built-in authentication |
| **Routing** | Not included | Built-in routing |
| **Learning Curve** | Easy for Eloquent users | Steeper (full framework) |
| **Dependencies** | Minimal | Framework dependencies |

## Strengths

### Neat ORM
- Laravel Eloquent-like API (familiar to PHP developers)
- Clean, focused feature set
- Easy learning curve for Eloquent users
- Built-in migrations, factories, and seeders
- Comprehensive documentation
- No framework dependencies (use anywhere)
- Lightweight and focused

### Goravel (GORM)
- Full-stack solution (everything included)
- Mature GORM ORM with large ecosystem
- Laravel-inspired framework patterns
- Built-in authentication and routing
- Large community and production adoption
- Active development and maintenance
- Comprehensive feature set

## Weaknesses

### Neat ORM
- Newer codebase (less battle-tested)
- Smaller community (fewer eyes on code)
- Limited to database operations only
- Fewer relationship types (no ManyToMany yet)
- No built-in auth or routing

### Goravel (GORM)
- Larger codebase (more complexity)
- Steeper learning curve (full framework)
- Framework dependencies
- Overkill for simple projects
- More complex setup

## Use Case Recommendations

### Choose Neat ORM If:
- You need a standalone ORM library
- You prefer Laravel Eloquent-like API in Go
- You want a clean, focused ORM without bloat
- You need built-in migrations, factories, and seeders
- You're building a microservice or API
- You want to use the ORM in existing projects
- You prefer minimal dependencies

### Choose Goravel If:
- You need a full-stack web framework
- You want Laravel-like framework experience in Go
- You need built-in authentication and routing
- You're building a complete web application
- You want GORM's mature ecosystem
- You prefer an all-in-one solution
- You're starting a new project from scratch

## Conclusion

Neat ORM and Goravel serve different purposes:

**Neat ORM**: Standalone ORM with Laravel Eloquent-like API. Ideal for developers who want a focused ORM library they can use in any Go project without framework dependencies.

**Goravel**: Full-stack web framework with GORM ORM. Ideal for developers building complete web applications who want an all-in-one solution with auth, routing, and more.

The choice depends on your project needs:
- **Standalone ORM**: Neat ORM offers focused database operations with familiar Eloquent API
- **Full framework**: Goravel provides complete web framework with GORM ORM and additional features

Both provide excellent developer experiences but for different use cases (standalone ORM vs full-stack framework).

## References

- Neat ORM Documentation: See `docs/` directory
- Goravel Documentation: https://www.goravel.dev/
- GORM Documentation: https://gorm.io/docs/
- Laravel Eloquent Documentation: https://laravel.com/docs/eloquent
