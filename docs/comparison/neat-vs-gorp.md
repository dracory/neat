# Neat ORM vs GORP Developer Experience Comparison

**Date**: June 1, 2026

## Overview

This document compares the developer experience, ease of use, and features of Neat ORM with GORP (Go Relational Persistence), an ORM-ish library for Go.

**Important Note**: The GORP author explicitly states they "hesitate to call gorp an ORM" as it doesn't have full ORM features like associations.

## Why Neat Wins?

**Neat ORM wins when you need a full ORM with associations, migrations, factories, and seeders rather than a basic ORM-ish library.** GORP explicitly states it's not a full ORM and lacks standard ORM features like associations, migrations, and testing tools. Neat provides a complete ORM solution with BelongsTo, HasMany, HasOne associations, built-in migrations, factories, and seeders. If you need a full-featured ORM with all standard capabilities rather than a basic ORM-ish library, Neat is the clear winner.

## GORP Overview

GORP is Go Relational Persistence, described as an "ORM-ish library." It takes a code-first approach using tags for specifications and uses reflection for constructing SQL queries. It lacks full ORM features like associations.

## Architecture Comparison

### Neat ORM
- **Full ORM**: Complete ORM with models, relationships, migrations
- **Laravel Eloquent-like API**: Familiar syntax for developers coming from PHP/Laravel
- **Runtime query builder**: Fluent interface for building queries at runtime
- **Feature-rich**: Associations, observers, migrations, factories, seeders
- **Modern design**: Recent development with security focus

### GORP
- **ORM-ish library**: Not a full ORM (explicitly stated by author)
- **Code-first approach**: Uses tags for specifications
- **Reflection-based**: Uses reflection for SQL queries
- **Limited features**: Lacks associations and full ORM features
- **Mature codebase**: Established but simpler

## Developer Experience

### Neat ORM
- **Laravel Eloquent-like API**: Familiar to PHP/Laravel developers
- **Full ORM experience**: Complete ORM with all standard features
- **Clean fluent interface**: Intuitive method chaining
- **Active development**: Regular updates and improvements
- **Comprehensive documentation**: Docs with examples
- **Testing tools**: Built-in factories and seeders

### GORP
- **Simple API**: Code-first approach with tags
- **Reflection-based**: Automatic SQL construction
- **Limited scope**: Not a full ORM
- **Mature codebase**: Established library
- **Simple learning curve**: Basic features
- **No associations**: Limited relationship support

## Feature Comparison

| Feature | Neat ORM | GORP |
|---------|----------|------|
| **Type** | Full ORM | ORM-ish library (not full ORM) |
| **Query Builder** | Fluent Eloquent-like API | Code-first with tags |
| **Associations** | BelongsTo, HasMany, HasOne | Not supported |
| **Migrations** | Built-in migration system | Not included |
| **Soft Deletes** | Built-in soft delete support | Not included |
| **Observers** | Model lifecycle observers | Not included |
| **Factories/Seeders** | Built-in factories and seeders | Not included |
| **Reflection** | Uses reflection for mapping | Uses reflection for SQL |
| **Scope** | Complete ORM solution | Limited ORM features |

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

### GORP
- Simple code-first approach
- Reflection-based SQL construction
- Mature codebase
- Lightweight
- Simple learning curve
- Tag-based configuration

## Weaknesses

### Neat ORM
- More complex (full ORM)
- Runtime overhead from features
- Newer codebase
- Smaller community

### GORP
- **Not a full ORM** (explicitly stated)
- No associations or relationships
- No migrations
- No factories or seeders
- No observers or hooks
- No soft deletes
- Limited feature set
- Less active development

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

### Choose GORP If:
- You only need basic ORM-ish features
- You prefer code-first approach with tags
- You don't need associations
- You want a simple, lightweight solution
- You don't need migrations
- You don't need testing tools
- You're okay with limited features

**Important**: GORP explicitly states it's not a full ORM. If you need ORM features like associations, Neat ORM is the appropriate choice.

## Conclusion

Neat ORM and GORP serve different purposes:

**Neat ORM**: Full-featured ORM with Laravel Eloquent-like API and complete ORM features. Ideal for developers building full applications who need a comprehensive ORM.

**GORP**: ORM-ish library with basic features and code-first approach. Ideal for developers who only need basic persistence without full ORM features.

The choice depends on your needs:
- **Full ORM with all features**: Neat ORM
- **Basic ORM-ish features**: GORP

**Important**: GORP is explicitly not a full ORM. If you need standard ORM features, Neat ORM is the better choice.

## References

- Neat ORM Documentation: See `docs/` directory
- GORP GitHub: https://github.com/go-gorp/gorp
- Laravel Eloquent Documentation: https://laravel.com/docs/eloquent
