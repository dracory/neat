# Neat ORM Comparisons

This directory contains developer experience and feature comparisons between Neat ORM and other Go database libraries.

## Available Comparisons

### Major ORMs
- [Neat ORM vs GORM](./neat-vs-gorm.md) - Comparison with the most popular Go ORM
- [Neat ORM vs Ent](./neat-vs-ent.md) - Comparison with Facebook's code-first ORM
- [Neat ORM vs Xorm](./neat-vs-xorm.md) - Comparison with the simple and powerful ORM
- [Neat ORM vs Bun](./neat-vs-bun.md) - Comparison with the SQL-first ORM

### Database-Specific ORMs
- [Neat ORM vs go-pg](./neat-vs-go-pg.md) - Comparison with the PostgreSQL-focused ORM

### Framework ORMs
- [Neat ORM vs Goravel](./neat-vs-goravel-eloquent.md) - Comparison with the Goravel framework's GORM-based ORM
- [Neat ORM vs Beego ORM](./neat-vs-beego-orm.md) - Comparison with the Beego framework ORM

### SQL Compilers & Code Generators
- [Neat ORM vs sqlc](./neat-vs-sqlc.md) - Comparison with the SQL compiler approach
- [Neat ORM vs SQLBoiler](./neat-vs-sqlboiler.md) - Comparison with the database-first code generator

### Query Builders & DALs
- [Neat ORM vs goqu](./neat-vs-goqu.md) - Comparison with the SQL builder library
- [Neat ORM vs upper/db](./neat-vs-upper-io.md) - Comparison with the database-agnostic DAL

### Traditional ORMs
- [Neat ORM vs GORP](./neat-vs-gorp.md) - Comparison with Go Relational Persistence
- [Neat ORM vs Reform](./neat-vs-reform.md) - Comparison with better record management
- [Neat ORM vs REL](./neat-vs-rel.md) - Comparison with the modern clean API ORM
- [Neat ORM vs go-queryset](./neat-vs-go-queryset.md) - Comparison with Django-like querysets
- [Neat ORM vs godb](./neat-vs-godb.md) - Comparison with the simple ORM

### Specialized ORMs
- [Neat ORM vs Zoom](./neat-vs-zoom.md) - Comparison with the Redis-based in-memory ORM

## Comparison Categories

### Traditional ORMs
- **GORM**: Mature, feature-rich ORM with large ecosystem
- **Neat ORM**: Laravel Eloquent-like API with focused feature set

### Code-First ORMs
- **Ent**: Facebook's code-first ORM with compile-time type safety
- **Neat ORM**: Runtime query builder with Laravel-like API

### Full-Stack Frameworks
- **Goravel**: Full-stack framework using GORM
- **Neat ORM**: Standalone ORM library

### SQL Compilers
- **sqlc**: SQL compiler generating type-safe Go code
- **Neat ORM**: Traditional ORM with query builder

## Key Differentiators

### Neat ORM Strengths
- Laravel Eloquent-like API (familiar to PHP developers)
- No code generation required
- Built-in migrations, factories, and seeders
- Clean, focused feature set
- Quick setup and easy learning curve
- Oracle database support
- Advanced soft delete strategies (NULL-based and max-date sentinel)
- Polymorphic associations (PolymorphicBelongsTo, PolymorphicHasMany)
- Multiple query method aliases (Sequelize and Django-style)
- Security hardening with SQL injection prevention
- ToSql interface for SQL generation without execution

### When to Choose Neat ORM
- You prefer Laravel Eloquent-like API
- You want a standalone ORM without framework dependencies
- You need built-in migrations, factories, and seeders
- You're coming from PHP/Laravel background
- You prefer traditional ORM experience

## Additional Resources

- [Neat ORM Documentation](../README.md)
- [API Reference](../api-reference.md)
- [Query Builder](../query-builder.md)
- [Associations](../associations.md)
