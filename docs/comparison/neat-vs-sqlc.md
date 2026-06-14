# Neat ORM vs sqlc Developer Experience Comparison

**Date**: June 1, 2026

## Overview

This document compares the developer experience, ease of use, and features of Neat ORM with sqlc, a SQL compiler that generates type-safe Go code from SQL queries.

## Why Neat Wins?

**Neat ORM wins when you prefer a traditional ORM abstraction over writing raw SQL and want built-in migrations, factories, and seeders without code generation complexity.** Neat provides a familiar ORM experience with Laravel Eloquent-like API that abstracts SQL away while still offering dynamic query construction. While sqlc offers superior compile-time type safety and excellent IDE support, Neat offers faster development with no build step and comprehensive testing tools out of the box. If you value ORM abstraction, dynamic queries, and built-in testing infrastructure over writing raw SQL and code generation, Neat is the better choice.

## sqlc Overview

sqlc is not a traditional ORM - it's a SQL compiler that takes SQL queries as input and generates type-safe Go code. It emphasizes compile-time safety and explicit SQL control. You write SQL, sqlc generates type-safe Go code.

## Architecture Comparison

### Neat ORM
- **Runtime query builder**: Fluent interface for building queries at runtime
- **Dynamic queries**: Queries constructed programmatically
- **ORM pattern**: Object-relational mapping with models
- **Reflection-based**: Uses Go reflection for model mapping
- **No code generation**: Direct usage without build step

### sqlc
- **SQL compiler**: Compiles SQL queries to Go code
- **Code generation**: Generates type-safe Go code from SQL
- **Explicit SQL**: Write SQL directly, get type-safe Go code
- **No runtime query building**: All queries defined at compile time
- **Compile-time safety**: Errors caught at compile time

## Developer Experience

### Neat ORM
- **Laravel Eloquent-like API**: Familiar syntax for developers coming from PHP/Laravel
- **No code generation**: Direct usage without build complexity
- **Quick setup**: Simple configuration and immediate usage
- **Flexible queries**: Dynamic query construction at runtime
- **Clear documentation**: Comprehensive docs with examples
- **ORM familiarity**: Traditional ORM experience

### sqlc
- **SQL-first**: Write SQL directly (familiar to SQL developers)
- **Type-safe generation**: Compile-time type safety
- **IDE support**: Excellent autocomplete and type hints
- **Explicit control**: Full control over SQL queries
- **No magic**: What you write is what you get
- **SQL expertise**: Requires strong SQL knowledge

## Feature Comparison

| Feature | Neat ORM | sqlc |
|---------|----------|------|
| **Query Building** | Fluent Eloquent-like API | Write SQL directly |
| **Type Safety** | Runtime type checking | Compile-time type safety |
| **Code Generation** | Not required | Required (generates code) |
| **SQL Control** | Abstracted (query builder) | Full SQL control |
| **Migrations** | Built-in migration system | Not included |
| **Relationships** | BelongsTo, HasMany, HasOne, PolymorphicBelongsTo, PolymorphicHasMany | Manual SQL joins |
| **IDE Support** | Basic autocomplete | Excellent type hints |
| **Learning Curve** | Easy for ORM users | Requires SQL expertise |
| **Build Complexity** | Simple | Requires code generation |
| **Database Support** | MySQL, PostgreSQL, SQLite, SQL Server, Turso, Oracle | Multiple databases |

## Query Building

### Neat ORM
- **Fluent interface**: Chainable query builder methods
- **Runtime construction**: Queries built at runtime
- **Dynamic conditions**: Conditions can be added dynamically
- **Laravel-like API**: Similar to Laravel Eloquent
- **Abstraction**: SQL abstracted away
- **Easy to learn**: Familiar pattern for many developers

### sqlc
- **Write SQL directly**: Full SQL control
- **Type-safe generation**: Generated code is type-safe
- **Static queries**: Queries defined at compile time
- **Explicit control**: No abstraction over SQL
- **SQL expertise**: Requires strong SQL knowledge
- **Compile-time errors**: SQL errors caught at compile time

## Type Safety

### Neat ORM
- **Runtime type checking**: Types validated at query execution
- **Reflection-based**: Uses Go reflection for model mapping
- **Dynamic queries**: Query structure determined at runtime
- **Flexibility**: More flexible but less compile-time safety

### sqlc
- **Compile-time type safety**: Types validated at compile time
- **Code generation**: Type-safe code generated from SQL
- **Static analysis**: IDE support and static analysis tools
- **Strong typing**: Strict type enforcement prevents many errors
- **No runtime type errors**: Type errors caught at compile time

## Schema Management

### Neat ORM
- **Struct tags**: Schema defined via Go struct tags
- **Migration support**: Built-in migration system
- **Flexible structure**: Schema can change at runtime
- **Reflection-based**: Uses Go reflection for schema discovery
- **Quick changes**: No regeneration needed

### sqlc
- **SQL schema**: Schema defined in SQL
- **No built-in migrations**: Migration tools required separately
- **Manual schema management**: Schema changes manual
- **Regeneration required**: SQL changes need code regeneration
- **Explicit control**: Full control over schema

## Setup and Configuration

### Neat ORM
- **Simple setup**: Import and use immediately
- **No build step**: Direct usage without code generation
- **DSN configuration**: Simple connection string
- **Quick start**: Get started in minutes
- **Minimal configuration**: Basic configuration options

### sqlc
- **SQL setup**: Define SQL queries in files
- **Code generation setup**: Configure sqlc generation
- **Build step**: Must run code generation
- **Initial complexity**: More setup required
- **Configuration**: sqlc.yaml configuration file

## Error Handling

### Neat ORM
- **Runtime errors**: Errors caught at query execution
- **Debug mode**: Detailed errors in development
- **Context preservation**: Context errors not suppressed
- **Clear error messages**: Easy to debug
- **SQL errors**: SQL errors at runtime

### sqlc
- **Compile-time errors**: SQL errors caught at compile time
- **Type errors**: Type errors caught at compile time
- **Fewer runtime errors**: Compile-time safety prevents many issues
- **SQL validation**: SQL validated at generation time
- **Generated error handling**: Consistent error patterns

## Performance

### Neat ORM
- **Runtime overhead**: Reflection and validation at runtime
- **Query builder**: Fluent interface overhead
- **Dynamic queries**: Flexibility with some performance cost
- **Recent optimization**: Performance considerations in design

### sqlc
- **No runtime overhead**: Generated code is efficient
- **Compile-time optimization**: Queries optimized at generation time
- **Direct SQL**: No abstraction overhead
- **Optimized generation**: Generated code is optimized

## Strengths

### Neat ORM
- Laravel Eloquent-like API (familiar to many)
- No code generation required (simple setup)
- Flexible runtime query building
- Easy learning curve for ORM users
- Quick to get started
- Dynamic query construction
- Built-in migrations, factories, seeders

### sqlc
- Compile-time type safety (catch errors early)
- Full SQL control (no abstraction)
- Excellent IDE support and autocomplete
- Type-safe operations by design
- No runtime query building overhead
- SQL expertise utilization
- Explicit and predictable

## Weaknesses

### Neat ORM
- Runtime type checking (less compile-time safety)
- Reflection overhead
- SQL abstraction (less control)
- Newer codebase (less battle-tested)
- Smaller community
- Fewer IDE type hints

### sqlc
- Code generation required (build complexity)
- Requires SQL expertise
- Less flexible runtime queries
- Schema changes require regeneration
- More complex setup
- No built-in migrations
- Manual relationship management
- Verbose for simple queries

## Use Case Recommendations

### Choose Neat ORM If:
- You prefer Laravel Eloquent-like API
- You need flexible runtime query building
- You want to avoid code generation
- You need dynamic query construction
- You prefer simple setup and quick start
- You're coming from PHP/Laravel background
- You want traditional ORM experience
- You need built-in migrations and factories
- You prefer abstraction over SQL

### Choose sqlc If:
- You prefer writing SQL directly
- Compile-time type safety is critical
- You want full control over SQL queries
- You have strong SQL expertise
- You prefer explicit over implicit
- You want excellent IDE support
- You're building complex queries
- You value compile-time error detection
- You want to avoid runtime query building overhead

## Conclusion

Neat ORM and sqlc represent fundamentally different approaches to database operations:

**Neat ORM**: Traditional runtime query builder with Laravel Eloquent-like API. Offers flexibility and quick setup without code generation. Ideal for developers who want a familiar ORM experience with SQL abstraction.

**sqlc**: SQL compiler with compile-time type safety. Provides type-safe operations through code generation with full SQL control. Ideal for developers who prefer writing SQL directly and value compile-time safety.

The choice depends on your philosophy:
- **ORM abstraction with runtime flexibility**: Neat ORM
- **SQL-first with compile-time safety**: sqlc

Both provide excellent developer experiences but through different approaches (runtime query builder vs SQL compilation).

## References

- Neat ORM Documentation: See `docs/` directory
- sqlc Documentation: https://sqlc.dev/
- sqlc GitHub: https://github.com/sqlc-dev/sqlc
- Laravel Eloquent Documentation: https://laravel.com/docs/eloquent
