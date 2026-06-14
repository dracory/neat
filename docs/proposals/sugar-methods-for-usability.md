# Sugar Methods for API Usability

**Date**: June 14, 2026
**Status**: Open for Discussion
**Priority**: Medium

## Overview

This proposal introduces "sugar methods" to improve the developer experience of Neat ORM while maintaining the current performance-oriented API. The current pointer-based approach, while efficient, requires more verbose code:

```go
// Current (verbose)
var count int64
err := db.Query().Count(&count)

var user User
err := db.Query().FindOne(&user)

var users []User
err := db.Query().Find(&users)
```

This proposal adds convenience methods that return values directly, reducing boilerplate while keeping the existing methods for performance-critical scenarios.

## Problem Statement

### Current API Pain Points

1. **Verbosity**: Requires 2 lines instead of 1 for common operations
2. **Inconsistent with Go idioms**: Most Go functions return values directly
3. **Developer friction**: Extra cognitive load for simple operations
4. **Competitive disadvantage**: GORM and other ORMs use return-value approach

### Affected Methods

The pointer-based pattern is used extensively:

- **Aggregations**: `Count()`, `Sum()`, `Avg()`, `Min()`, `Max()`
- **Retrieval**: `First()`, `FindOne()`, `Get()`, `Find()`
- **Transactions**: `Transaction()` (in some cases)

## Proposed Solution

### Dual API Approach

Keep existing methods (performance-oriented) and add sugar methods (usability-oriented):

```go
// Existing API (performance, no allocation)
var count int64
err := db.Query().Count(&count)

// New sugar API (convenience, minor allocation)
count, err := db.Query().CountOne()
```

### Naming Convention

Sugar methods use descriptive suffixes to distinguish from base methods:

**Option A: WithVar suffix (Recommended)**
- `Count()` → `CountWithVar()` (returns int64)
- `Sum()` → `SumWithVar()` (returns float64)
- `Avg()` → `AvgWithVar()` (returns float64)
- `Min()` → `MinWithVar()` (returns float64)
- `Max()` → `MaxWithVar()` (returns float64)
- `First()` → `FirstWithVar()` (returns model)
- `FindOne()` → `FindOneWithVar()` (returns model)
- `Get()` → `GetWithVar()` (returns slice)
- `Find()` → `FindWithVar()` (returns slice)

**Rationale**: Self-documenting - "we handle the variable declaration for you". Makes the trade-off between performance (you manage var) and usability (we manage var) explicit.

**Option B: Result suffix**
- `Count()` → `CountResult()` (returns int64)
- `Sum()` → `SumResult()` (returns float64)
- `Avg()` → `AvgResult()` (returns float64)
- `Min()` → `MinResult()` (returns float64)
- `Max()` → `MaxResult()` (returns float64)
- `First()` → `FirstResult()` (returns model)
- `FindOne()` → `FindOneResult()` (returns model)
- `Get()` → `GetResults()` (returns slice)
- `Find()` → `FindResults()` (returns slice)

**Option C: WithResult suffix**
- `Count()` → `CountWithResult()` (returns int64)
- `Sum()` → `SumWithResult()` (returns float64)
- `Avg()` → `AvgWithResult()` (returns float64)
- `Min()` → `MinWithResult()` (returns float64)
- `Max()` → `MaxWithResult()` (returns float64)
- `First()` → `FirstWithResult()` (returns model)
- `FindOne()` → `FindOneWithResult()` (returns model)
- `Get()` → `GetWithResults()` (returns slice)
- `Find()` → `FindWithResults()` (returns slice)

**Option D: OrError suffix**
- `Count()` → `CountOrError()` (returns int64, error)
- `Sum()` → `SumOrError()` (returns float64, error)
- `Avg()` → `AvgOrError()` (returns float64, error)
- `Min()` → `MinOrError()` (returns float64, error)
- `Max()` → `MaxOrError()` (returns float64, error)
- `First()` → `FirstOrError()` (returns model, error)
- `FindOne()` → `FindOneOrError()` (returns model, error)
- `Get()` → `GetOrError()` (returns slice, error)
- `Find()` → `FindOrError()` (returns slice, error)

**Option E: As prefix**
- `Count()` → `AsCount()` (returns int64)
- `Sum()` → `AsSum()` (returns float64)
- `Avg()` → `AsAvg()` (returns float64)
- `Min()` → `AsMin()` (returns float64)
- `Max()` → `AsMax()` (returns float64)
- `First()` → `AsFirst()` (returns model)
- `FindOne()` → `AsFindOne()` (returns model)
- `Get()` → `AsGet()` (returns slice)
- `Find()` → `AsFind()` (returns slice)

**Option F: Direct return (no suffix, requires generics)**
- `Count()` → `Count[T int64]()` (returns int64)
- `Sum()` → `Sum[T float64]()` (returns float64)
- `Avg()` → `Avg[T float64]()` (returns float64)
- `Min()` → `Min[T float64]()` (returns float64)
- `Max()` → `Max[T float64]()` (returns float64)
- `First()` → `First[T User]()` (returns T)
- `FindOne()` → `FindOne[T User]()` (returns T)
- `Get()` → `Get[T User]()` (returns []T)
- `Find()` → `Find[T User]()` (returns []T)

## Detailed API Proposal

### Aggregation Methods

```go
// Existing (keep)
func (q *Query) Count(count *int64) error
func (q *Query) Sum(column string, dest any) error
func (q *Query) Avg(column string, dest any) error
func (q *Query) Min(column string, dest any) error
func (q *Query) Max(column string, dest any) error

// New sugar methods (using Option A: WithVar suffix)
func (q *Query) CountWithVar() (int64, error)
func (q *Query) SumWithVar(column string) (float64, error)
func (q *Query) AvgWithVar(column string) (float64, error)
func (q *Query) MinWithVar(column string) (float64, error)
func (q *Query) MaxWithVar(column string) (float64, error)
```

**Usage Examples**:

```go
// Current
var count int64
err := db.Query().Count(&count)

// Sugar
count, err := db.Query().CountWithVar()

// Current
var total float64
err := db.Query().Sum("price", &total)

// Sugar
total, err := db.Query().SumWithVar("price")
```

### Retrieval Methods

```go
// Existing (keep)
func (q *Query) First(dest any) error
func (q *Query) FindOne(dest any) error
func (q *Query) Get(dest any) error
func (q *Query) Find(dest any) error

// New sugar methods (using Option A: WithVar suffix)
func (q *Query) FirstWithVar[T any]() (T, error)
func (q *Query) FindOneWithVar[T any]() (T, error)
func (q *Query) GetWithVar[T any]() ([]T, error)
func (q *Query) FindWithVar[T any]() ([]T, error)
```

**Usage Examples**:

```go
// Current
var user User
err := db.Query().First(&user)

// Sugar
user, err := db.Query().FirstWithVar[User]()

// Current
var users []User
err := db.Query().Find(&users)

// Sugar
users, err := db.Query().FindWithVar[User]()
```

### Generic Implementation

The sugar methods use Go generics to maintain type safety:

```go
func (q *Query) FirstWithVar[T any]() (T, error) {
    var result T
    err := q.First(&result)
    return result, err
}

func (q *Query) FindWithVar[T any]() ([]T, error) {
    var result []T
    err := q.Find(&result)
    return result, err
}
```

## Performance Considerations

### Allocation Impact

Sugar methods introduce minor allocations:

```go
// No allocation (existing)
var count int64
err := db.Query().Count(&count)

// 1 allocation (sugar - for zero value)
count, err := db.Query().CountWithVar()
```

**Benchmark Estimates**:
- Aggregation sugar: ~1-2 ns overhead (negligible)
- Single record sugar: ~5-10 ns overhead (negligible)
- Slice sugar: ~10-20 ns overhead (negligible)

### Use Case Guidance

**Use existing methods when**:
- Performance-critical code paths
- Hot loops
- Writing to existing variables
- Avoiding allocations

**Use sugar methods when**:
- Application logic
- API handlers
- Business logic
- Readability matters more than micro-optimizations

## Implementation Plan

### Phase 1: Core Aggregations
- Add `CountWithVar()`, `SumWithVar()`, `AvgWithVar()`, `MinWithVar()`, `MaxWithVar()`
- Add tests
- Update documentation

### Phase 2: Retrieval Methods
- Add `FirstWithVar()`, `FindOneWithVar()`, `GetWithVar()`, `FindWithVar()`
- Add tests
- Update documentation

### Phase 3: Additional Convenience
- Consider `Must*` variants (panic on error)
- Consider `Or*` variants (return default on error)
- Update examples

### Phase 4: Documentation & Migration
- Update all documentation to show both APIs
- Add migration guide
- Update examples with sugar methods where appropriate

## Backward Compatibility

This proposal is **100% backward compatible**:
- All existing methods remain unchanged
- No breaking changes
- Existing code continues to work
- Sugar methods are additive only

## Documentation Strategy

### Documentation Guidelines

1. **Primary examples**: Use sugar methods for readability
2. **Performance notes**: Document when to use base methods
3. **Migration guide**: Show how to switch between APIs
4. **Best practices**: Guide on when to use each API

### Example Documentation

```markdown
## Counting Records

### Sugar Method (Recommended for most cases)
```go
count, err := db.Query().CountWithVar()
if err != nil {
    return err
}
```

### Base Method (For performance-critical code)
```go
var count int64
err := db.Query().Count(&count)
if err != nil {
    return err
}
```

### When to use each:
- Use `CountWithVar()` for application logic, API handlers
- Use `Count()` for hot loops, performance-critical paths
```

## Alternative Approaches Considered

### Option 1: Change Existing Methods (Rejected)
**Pros**: Single API, simpler
**Cons**: Breaking change, affects performance everywhere

### Option 2: Configuration-Based (Rejected)
**Pros**: Single API surface
**Cons**: Runtime overhead, confusing behavior, harder to debug

### Option 3: Builder Pattern (Rejected)
**Pros**: Fluent API
**Cons**: More complex, doesn't solve core issue

### Option 4: Wrapper Functions (Rejected)
**Pros**: No changes to core
**Cons**: Separate package, fragmentation, harder to discover

**Chosen Approach**: Dual API with sugar methods - best balance of usability, performance, and backward compatibility.

## Open Questions

1. **Naming Convention**: Which naming scheme should we use?
   - Option A: `CountWithVar()`, `FirstWithVar()`, `FindWithVar()` (WithVar suffix - self-documenting, "we handle the var")
   - Option B: `CountResult()`, `FirstResult()`, `FindResults()` (Result suffix - consistent)
   - Option C: `CountWithResult()`, `FirstWithResult()`, `FindWithResults()` (WithResult suffix)
   - Option D: `CountOrError()`, `FirstOrError()`, `FindOrError()` (OrError suffix - implies return value + error)
   - Option E: `AsCount()`, `AsFirst()`, `AsFind()` (As prefix)
   - Option F: `Count[T int64]()`, `First[T User]()` (Direct return with generics)
2. **Generics**: Should all sugar methods use generics, or only retrieval methods?
3. **Must variants**: Should we add `MustCount()`, `MustFirst()` that panic on error?
4. **Default values**: Should we add `OrDefault*()` variants that return zero values on error?

## Success Metrics

- Adoption rate of sugar methods in new code
- Developer satisfaction surveys
- Reduction in lines of code for common operations
- No performance regression in benchmarks
- Minimal confusion between the two APIs

## References

- GORM API design (return-value approach)
- sqlx API design (pointer-based approach)
- Go idioms and conventions
- Developer feedback from issue tracker

## Appendix: Complete API Comparison

### Current API
```go
var count int64
err := db.Query().Count(&count)

var total float64
err := db.Query().Sum("price", &total)

var user User
err := db.Query().First(&user)

var users []User
err := db.Query().Find(&users)
```

### Proposed Sugar API
```go
count, err := db.Query().CountWithVar()

total, err := db.Query().SumWithVar("price")

user, err := db.Query().FirstWithVar[User]()

users, err := db.Query().FindWithVar[User]()
```

### Lines of Code Reduction
- Aggregations: 2 lines → 1 line (50% reduction)
- Single record: 2 lines → 1 line (50% reduction)
- Multiple records: 2 lines → 1 line (50% reduction)
