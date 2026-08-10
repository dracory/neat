# Validation

**Date**: June 1, 2026
**Last Reviewed**: August 10, 2026
**Status**: Not Started
**Priority**: Medium
**Impact**: High
**Effort**: Medium
**ROI**: High

## Description

Add built-in model validation.

## Rationale

- Data integrity at model level
- Consistent validation logic
- User-friendly error messages

## Proposed API

```go
type User struct {
    ID    uint
    Name  string `validate:"required,min=3,max=50"`
    Email string `validate:"required,email"`
    Age   int    `validate:"min=18,max=120"`
}

user := User{Name: "John", Email: "invalid"}
if err := db.Query().Model(&user).Validate(); err != nil {
    // Validation errors
}
```

## Implementation Details

- Struct tag validation
- Custom validation rules
- Localization support
- Validation groups

## References

- Extracted from `docs/proposals/feature-requests.md` (item #17)
