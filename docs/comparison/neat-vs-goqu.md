# Neat ORM vs goqu Developer Experience Comparison

**Date**: June 1, 2026

## Overview

This document compares the developer experience, ease of use, and features of Neat ORM with goqu, a SQL builder and query library for Go.

## Why Neat Wins?

**Neat ORM wins when you need a full ORM with associations, migrations, factories, and seeders rather than just a SQL query builder.** goqu explicitly states it's not a full ORM and lacks standard ORM features like relationships, migrations, and testing tools. Neat provides a complete ORM solution with BelongsTo, HasMany, HasOne associations, built-in migrations, factories, and seeders. If you need a full-featured ORM with all standard capabilities rather than a lightweight query builder, Neat is the clear winner.

## goqu Overview

goqu is a SQL builder and query library for Go. It's important to note that goqu is **not intended to be a full ORM** - it's a query builder with some ORM-like features. It lacks common ORM features like associations and hooks.

## Architecture Comparison

### Neat ORM
- **Full ORM**: Complete ORM with models, relationships, migrations
- **Runtime query builder**: Fluent interface for building queries at runtime
- **ORM pattern**: Object-relational mapping with models
- **Reflection-based**: Uses Go reflection for model mapping
- **Feature-rich**: Associations, observers, migrations, factories, seeders

### goqu
- **SQL builder**: Query building library (not a full ORM)
- **No ORM features**: Lacks associations, hooks, migrations
- **Query-focused**: Focused on SQL query construction
- **Explicit SQL**: SQL-first approach with some abstraction
- **Lightweight**: Minimal feature set focused on queries

## Developer Experience

### Neat ORM
- **Laravel Eloquent-like API**: Familiar syntax for developers coming from PHP/Laravel
- **Full ORM experience**: Complete ORM with all standard features
- **Quick setup**: Simple configuration and immediate usage
- **Flexible queries**: Dynamic query construction at runtime
- **Clear documentation**: Comprehensive docs with examples
- **ORM familiarity**: Traditional ORM experience

### goqu
- **SQL builder API**: Fluent SQL building interface
- **Lightweight**: Minimal setup and dependencies
- **SQL-focused**: Focus on query construction
- **Explicit control**: Clear SQL construction
- **Simple learning curve**: Easy to learn for SQL developers
- **Not an ORM**: No ORM features to learn

## Feature Comparison

| Feature | Neat ORM | goqu |
|---------|----------|------|
| **Type** | Full ORM | SQL builder (not ORM) |
| **Query Builder** | Fluent Eloquent-like API | Fluent SQL builder API |
| **ORM Features** | Models, relationships, hooks | None (query builder only) |
| **Associations** | BelongsTo, HasMany, HasOne, PolymorphicBelongsTo, PolymorphicHasMany | Not supported |
| **Migrations** | Built-in migration system | Not included |
| **Soft Deletes** | Built-in soft delete support | Not included |
| **Observers** | Model lifecycle observers | Not included |
| **Factories/Seeders** | Built-in factories and seeders | Not included |
| **Database Support** | MySQL, PostgreSQL, SQLite, SQL Server, Turso, Oracle | Multiple databases |
| **Learning Curve** | Easy for ORM users | Easy for SQL developers |
| **Scope** | Complete ORM solution | Query building only |

## Query Building

### Neat ORM
- **Fluent interface**: Chainable query builder methods
- **Runtime construction**: Queries built at runtime
- **Dynamic conditions**: Conditions can be added dynamically
- **Laravel-like API**: Similar to Laravel Eloquent
- **ORM integration**: Queries work with models
- **Easy to learn**: Familiar pattern for many developers

### goqu
- **Fluent SQL builder**: Chainable SQL building methods
- **SQL construction**: SQL built explicitly
- **Dynamic conditions**: Conditions can be added dynamically
- **SQL-focused**: SQL construction is primary focus
- **No ORM integration**: Works with structs but not models
- **SQL expertise**: Requires SQL knowledge

## Model Support

### Neat ORM
- **Full ORM models**: Complete model support with relationships
- **Struct tags**: Schema defined via Go struct tags
- **Relationships**: BelongsTo, HasMany, HasOne associations
- **Soft deletes**: Built-in soft delete support
- **Observers**: Model lifecycle observers
- **Timestamps**: Automatic timestamp management

### goqu
- **Struct scanning**: Can scan rows into structs
- **No relationships**: No association support
- **No lifecycle hooks**: No observer pattern
- **Manual schema**: Manual schema management
- **Basic models**: Basic struct support only
- **No ORM features**: Explicitly not an ORM

## Migration Support

### Neat ORM
- **Built-in migrations**: Schema builder for migrations
- **ORM driver**: Migration system with ORM support
- **Schema builder**: Fluent API for schema changes
- **Rollback support**: Migration rollback capabilities
- **Integrated**: Migrations integrated with ORM

### goqu
- **No migrations**: Not included
- **External tools**: Requires separate migration tools
- **Manual schema**: Manual schema management
- **No integration**: No migration integration

## Testing Support

### Neat ORM
- **Factories**: Built-in factory pattern for test data
- **Seeders**: Database seeding for test data
- **Integration tests**: Docker Compose setup for testing
- **Transaction rollback**: Test transaction support

### goqu
- **No factories**: Not included
- **No seeders**: Not included
- **Manual testing**: Manual test setup required
- **Transaction support**: Basic transaction support

## Performance

### Neat ORM
- **Runtime overhead**: Reflection and validation at runtime
- **Query builder**: Fluent interface overhead
- **Dynamic queries**: Flexibility with some performance cost
- **Recent optimization**: Performance considerations in design

### goqu
- **Minimal overhead**: Lightweight query building
- **SQL-focused**: Optimized for SQL construction
- **No reflection**: No reflection overhead
- **Performance-focused**: Designed for performance

## Strengths

### Neat ORM
- Laravel Eloquent-like API (familiar to many)
- Full ORM with all standard features
- Built-in migrations, factories, and seeders
- Associations and relationships
- Model lifecycle observers
- Soft deletes support
- Comprehensive documentation
- Complete ORM solution

### goqu
- Lightweight and focused
- Excellent SQL builder
- Minimal dependencies
- Fast performance
- Simple learning curve
- SQL expertise utilization
- Explicit and predictable
- No ORM overhead

## Weaknesses

### Neat ORM
- More complex (full ORM)
- Runtime overhead from features
- Reflection overhead
- Newer codebase (less battle-tested)
- Smaller community

### goqu
- Not a full ORM (limited features)
- No associations or relationships
- No migrations
- No factories or seeders
- No observers or hooks
- Manual schema management
- Requires manual setup for ORM features
- Limited to query building

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
- You're building a full application

### Choose goqu If:
- You only need a SQL query builder
- You don't need ORM features
- You want minimal dependencies
- You prefer explicit SQL construction
- You need high performance
- You're building a simple data access layer
- You don't need associations
- You prefer lightweight solutions
- You're comfortable with SQL
- You want to avoid ORM overhead

## Conclusion

Neat ORM and goqu serve different purposes:

**Neat ORM**: Full-featured ORM with Laravel Eloquent-like API. Offers complete ORM solution with models, relationships, migrations, factories, and more. Ideal for developers building full applications who need a comprehensive ORM.

**goqu**: SQL builder and query library (not an ORM). Offers lightweight, focused SQL query building without ORM features. Ideal for developers who only need query building and want to avoid ORM overhead.

The choice depends on your needs:
- **Full ORM with all features**: Neat ORM
- **Lightweight SQL builder**: goqu

**Important**: goqu explicitly states it's not intended to be an ORM. If you need ORM features like associations, hooks, or migrations, Neat ORM is the appropriate choice.

## References

- Neat ORM Documentation: See `docs/` directory
- goqu Documentation: http://doug-martin.github.io/goqu/
- goqu GitHub: https://github.com/doug-martin/goqu
- goqu Package: https://pkg.go.dev/github.com/doug-martin/goqu/v9
- Laravel Eloquent Documentation: https://laravel.com/docs/eloquent
