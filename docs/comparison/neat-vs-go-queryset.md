# Neat ORM vs go-queryset Developer Experience Comparison

**Date**: June 1, 2026

## Overview

This document compares the developer experience, ease of use, and features of Neat ORM with go-queryset, a 100% type-safe ORM for Go that implements the Django QuerySet pattern.

## Why Neat Wins?

**Neat ORM wins when you prefer Laravel Eloquent-like API and need built-in migrations, factories, and seeders rather than Django QuerySet patterns.** While go-queryset offers 100% type safety and Django QuerySet patterns, it lacks built-in migrations, factories, seeders, and observers. Neat provides Laravel-like patterns, comprehensive testing infrastructure, and model lifecycle observers. If you value Laravel familiarity, built-in testing tools, and observer patterns over Django QuerySet patterns, Neat is the better choice.

## go-queryset Overview

go-queryset is a 100% type-safe ORM for Go that implements the Django QuerySet pattern. It allows query reuse by defining custom methods on QuerySets and supports all DBMS that GORM supports. Performance is similar to GORM.

## Architecture Comparison

### Neat ORM
- **Laravel Eloquent-like API**: Familiar syntax for developers coming from PHP/Laravel
- **Runtime query builder**: Fluent interface for building queries at runtime
- **Feature-rich**: Associations, observers, migrations, factories, seeders
- **Reflection-based**: Uses Go reflection for model mapping
- **Modern design**: Recent development with security focus

### go-queryset
- **Django QuerySet pattern**: Implements Django's QuerySet API
- **100% type-safe**: Compile-time type safety
- **Query reuse**: Custom methods on QuerySets
- **GORM-compatible**: Supports same DBMS as GORM
- **Reflection-based**: Uses reflection like GORM

## Developer Experience

### Neat ORM
- **Laravel Eloquent-like API**: Familiar to PHP/Laravel developers
- **Clean fluent interface**: Intuitive method chaining
- **Active development**: Regular updates and improvements
- **Comprehensive documentation**: Docs with examples
- **Testing tools**: Built-in factories and seeders
- **Runtime flexibility**: Dynamic query construction

### go-queryset
- **Django QuerySet pattern**: Familiar to Python/Django developers
- **100% type-safe**: Compile-time type checking
- **Query reuse**: Custom QuerySet methods
- **GORM compatibility**: Same database support
- **Performance**: Similar to GORM
- **Django familiarity**: Familiar to Django developers

## Feature Comparison

| Feature | Neat ORM | go-queryset |
|---------|----------|-------------|
| **Query Builder** | Fluent Eloquent-like API | Django QuerySet pattern |
| **Type Safety** | Runtime type checking | 100% type-safe (compile-time) |
| **Database Support** | MySQL, PostgreSQL, SQLite, SQL Server, Turso, Oracle | Same as GORM (MySQL, PostgreSQL, SQLite3, SQL Server) |
| **Associations** | BelongsTo, HasMany, HasOne, PolymorphicBelongsTo, PolymorphicHasMany | Supports associations |
| **Migrations** | Built-in migration system | Not included |
| **Soft Deletes** | Built-in soft delete support | Not included |
| **Observers** | Model lifecycle observers | Not included |
| **Factories/Seeders** | Built-in factories and seeders | Not included |
| **Query Reuse** | Limited | Custom QuerySet methods |
| **API Influence** | Laravel Eloquent | Django QuerySet |

## Strengths

### Neat ORM
- Laravel Eloquent-like API (familiar to many)
- Built-in migrations, factories, and seeders
- Model lifecycle observers
- Soft deletes support
- Active development
- Recent security hardening
- Runtime flexibility
- Comprehensive documentation

### go-queryset
- 100% type-safe (compile-time)
- Django QuerySet pattern (familiar to Django developers)
- Query reuse with custom methods
- GORM-compatible database support
- Similar performance to GORM
- Type-safe operations
- Django familiarity

## Weaknesses

### Neat ORM
- Runtime type checking (less compile-time safety)
- Newer codebase
- Smaller community
- No query reuse pattern

### go-queryset
- No built-in migrations
- No factories or seeders
- No observers
- No soft deletes
- Uses reflection (like GORM)
- Less familiar to non-Django developers
- Less recent development

## Use Case Recommendations

### Choose Neat ORM If:
- You prefer Laravel Eloquent-like API
- You need built-in migrations, factories, and seeders
- You need model lifecycle observers
- You need soft deletes
- You prefer runtime flexibility
- You need dynamic query construction
- Recent security hardening is important
- You're coming from PHP/Laravel background

### Choose go-queryset If:
- You prefer Django QuerySet pattern
- You need 100% type safety
- You want query reuse with custom methods
- You're familiar with Django ORM
- You need GORM-compatible database support
- You value compile-time type safety
- You're coming from Python/Django background
- You want Django-like query patterns

## Conclusion

Neat ORM and go-queryset represent different API patterns:

**Neat ORM**: Laravel Eloquent-like ORM with comprehensive features and testing tools. Ideal for developers who want familiar Laravel ORM patterns with built-in migrations and factories.

**go-queryset**: Django QuerySet pattern ORM with 100% type safety. Ideal for developers familiar with Django who want compile-time type safety and query reuse patterns.

The choice depends on your background:
- **Laravel Eloquent patterns**: Neat ORM
- **Django QuerySet patterns**: go-queryset

## References

- Neat ORM Documentation: See `docs/` directory
- go-queryset GitHub: https://github.com/jirfag/go-queryset
- Laravel Eloquent Documentation: https://laravel.com/docs/eloquent
- Django QuerySet Documentation: https://docs.djangoproject.com/en/stable/topics/db/queries/
