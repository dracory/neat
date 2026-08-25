# ManyToMany Relationships

**Date**: June 1, 2026
**Last Reviewed**: August 25, 2026
**Status**: Not Started
**Priority**: High
**Impact**: High
**Effort**: High
**ROI**: High

## Description

Add support for many-to-many relationships with pivot tables.

## Rationale

- Currently missing from Neat (GORM has it)
- Critical for many real-world applications (e.g., user roles, tags, categories)
- Common requirement in most applications
- The `database/association/` package provides `BelongsTo`, `HasMany`, `HasOne`, and polymorphic variants, but no `many_to_many.go` exists

## Notes

- The association package (`database/association/`) has 8 files covering `belongs_to`, `has_many`, `has_one`, `polymorphic_belongs_to`, `polymorphic_has_many`, and a `field_resolver` — adding `many_to_many` would follow the same pattern.
- Pivot table operations (`Attach`, `Detach`, `Sync`) would likely live on a relationship struct returned by a method like `user.Roles()`.
- Eager loading via `With("Roles")` already works for HasMany/HasOne — many-to-many would extend the same `With()` mechanism with a join table resolver.

## Proposed API

```go
type User struct {
    ID    uint
    Name  string
    Roles []Role `neat:"many_to_many:user_roles"`
}

type Role struct {
    ID          uint
    Name        string
    Permissions []Permission `neat:"many_to_many:role_permissions"`
}

// Usage
user.Roles().Attach(roleID)
user.Roles().Detach(roleID)
user.Roles().Sync([]uint{roleID1, roleID2})
db.Query().Model(&User{}).With("Roles").Find(&users)
```

## Implementation Details

- Automatic pivot table creation
- Sync, attach, detach operations
- Eager loading with `With()`
- Pivot table with additional columns (timestamps, custom fields)

## References

- Extracted from `docs/proposals/feature-requests.md` (item #1)
- GORM many-to-many: https://gorm.io/docs/many_to_many.html
