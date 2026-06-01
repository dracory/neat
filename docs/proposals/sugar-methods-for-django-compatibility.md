# Proposal: Add Sugar Methods for Django Compatibility

**Status**: Proposed
**Date**: June 1, 2026
**Author**: Neat ORM Team

## Summary

Add sugar methods to Neat ORM to improve developer experience for developers coming from Django backgrounds. This proposal suggests adding method aliases like `Filter()` and `Exclude()` as alternatives to existing `Where()` and `WhereNot()` methods.

## Motivation

Currently, Neat ORM follows Laravel Eloquent's API patterns with methods like `Where()`, `WhereNot()`, etc. While this is excellent for Laravel developers, it can be less familiar for developers coming from Django/Python backgrounds who are used to Django's QuerySet API with methods like `Filter()` and `Exclude()`.

By adding these sugar methods, Neat can be more approachable to a broader audience without breaking existing code or changing the core API philosophy.

## Proposed Changes

### 1. Add `Filter()` Method

Add `Filter()` as an alias for `Where()`:

```go
func (q *Query) Filter(condition string, args ...interface{}) *Query {
    return q.Where(condition, args...)
}
```

**Usage:**
```go
// Existing Laravel-style (still works)
User.Where("age >= ?", 18).Get()

// New Django-style alias
User.Filter("age >= ?", 18).Get()
```

### 2. Add `Exclude()` Method

Add `Exclude()` as an alias for `WhereNot()`:

```go
func (q *Query) Exclude(condition string, args ...interface{}) *Query {
    return q.WhereNot(condition, args...)
}
```

**Usage:**
```go
// Existing Laravel-style (still works)
User.WhereNot("is_active = ?", false).Get()

// New Django-style alias
User.Exclude("is_active = ?", false).Get()
```

### 3. Consider `All()` Method

Add `All()` as an alias for `Get()` for Django familiarity:

```go
func (q *Query) All() ([]interface{}, error) {
    return q.Get()
}
```

**Usage:**
```go
// Existing Laravel-style (still works)
User.Where("age >= ?", 18).Get()

// New Django-style alias
User.Filter("age >= ?", 18).All()
```

## Benefits

1. **Lower barrier to entry**: Django developers can use familiar method names
2. **No breaking changes**: Existing code continues to work unchanged
3. **Minimal implementation cost**: Simple wrapper methods
4. **Better documentation**: Can show both Laravel and Django examples
5. **Competitive advantage**: More appealing to developers from both PHP and Python backgrounds

## Drawbacks

1. **API surface expansion**: More methods to maintain
2. **Potential confusion**: Two ways to do the same thing
3. **Documentation complexity**: Need to document both approaches

## Alternatives Considered

### Alternative 1: Full Django Operator Support
Parse Django-style `__` operators (e.g., `age__gte`) and convert to SQL.

**Pros**: More authentic Django experience
**Cons**: Complex implementation, potential for bugs, harder to debug

**Decision**: Rejected in favor of simple method aliases

### Alternative 2: No Changes
Keep current Laravel-only API.

**Pros**: Simpler API, less maintenance
**Cons**: Less approachable to Django developers

**Decision**: Rejected to improve developer experience

## Implementation Plan

1. Add `Filter()` method to query builder
2. Add `Exclude()` method to query builder
3. Add `All()` method to query builder
4. Update documentation to show both Laravel and Django examples
5. Add tests for new methods
6. Update comparison docs to mention Django compatibility

## Open Questions

1. Should we add more Django aliases for methods with different names (e.g., Django `get()` vs Neat `First()` for single record)?
2. Should we support Django's `__` operator syntax in the future?
3. Should we add a "compatibility mode" or just always support both?

## References

- Django ORM Documentation: https://docs.djangoproject.com/en/stable/topics/db/queries/
- Laravel Eloquent Documentation: https://laravel.com/docs/eloquent
- Current Neat ORM Query Builder: `database/query/query_where.go`
