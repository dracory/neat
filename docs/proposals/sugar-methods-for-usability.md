# Sugar Methods for API Usability

**Date**: June 14, 2026
**Status**: Implemented
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

**Option A: AsVar suffix (Recommended)**
- `Count()` → `CountAsVar()` (returns int64)
- `Sum()` → `SumAsVar()` (returns float64)
- `Avg()` → `AvgAsVar()` (returns float64)
- `Min()` → `MinAsVar()` (returns float64)
- `Max()` → `MaxAsVar()` (returns float64)
- `First()` → `FirstAsVar()` (returns model)
- `FindOne()` → `FindOneAsVar()` (returns model)
- `Get()` → `GetAsVar()` (returns slice)
- `Find()` → `FindAsVar()` (returns slice)

**Rationale**: Self-documenting - "we handle the variable declaration for you". Makes the trade-off between performance (you manage var) and usability (we manage var) explicit.

**Option G: AsVar suffix (Implemented)**
- `Count()` → `CountAsVar()` (returns int64)
- `Sum()` → `SumAsVar()` (returns float64)
- `Avg()` → `AvgAsVar()` (returns float64)
- `Min()` → `MinAsVar()` (returns float64)
- `Max()` → `MaxAsVar()` (returns float64)
- `First()` → `FirstAsVar()` (returns model)
- `FindOne()` → `FindOneAsVar()` (returns model)
- `Get()` → `GetAsVar()` (returns slice)
- `Find()` → `FindAsVar()` (returns slice)

**Rationale**: "AsVar" implies "give me the result as a variable" rather than "with a variable alongside it". More idiomatic in Go (like `errors.As`, type assertions) and clearly conveys the intent of materializing the result into a variable. This was chosen over the original recommendation for better Go idiomaticity.

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

// New sugar methods (using Option G: AsVar suffix)
func (q *Query) CountAsVar() (int64, error)
func (q *Query) SumAsVar(column string) (float64, error)
func (q *Query) AvgAsVar(column string) (float64, error)
func (q *Query) MinAsVar(column string) (float64, error)
func (q *Query) MaxAsVar(column string) (float64, error)
```

**Usage Examples**:

```go
// Current
var count int64
err := db.Query().Count(&count)

// Sugar
count, err := db.Query().CountAsVar()

// Current
var total float64
err := db.Query().Sum("price", &total)

// Sugar
total, err := db.Query().SumAsVar("price")
```

### Retrieval Methods

```go
// Existing (keep)
func (q *Query) First(dest any) error
func (q *Query) FindOne(dest any) error
func (q *Query) Get(dest any) error
func (q *Query) Find(dest any) error

// New sugar methods (using Option A: AsVar suffix)
func (q *Query) FirstAsVar() (T, error)
func (q *Query) FindOneAsVar[T any]() (T, error)
func (q *Query) GetAsVar[T any]() ([]T, error)
func (q *Query) FindAsVar[T any]() ([]T, error)
```

**Usage Examples**:

```go
// Current
var user User
err := db.Query().First(&user)

// Sugar
user, err := db.Query().FirstAsVar()

// Current
var users []User
err := db.Query().Find(&users)

// Sugar
users, err := db.Query().FindAsVar()
```

### Generic Implementation

The sugar methods use Go generics to maintain type safety:

```go
func (q *Query) FirstAsVar() (T, error) {
    var result T
    err := q.First(&result)
    return result, err
}

func (q *Query) FindAsVar[T any]() ([]T, error) {
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
count, err := db.Query().CountAsVar()
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

## Detailed Implementation Plan

### Complete Method Inventory

The following methods in the Query interface use pointer-based parameters and will receive sugar variants:

#### Aggregation Methods (8 methods)
1. `Count(count *int64) error` → `CountAsVar() (int64, error)`
2. `Sum(column string, dest any) error` → `SumAsVar(column string) (float64, error)`
3. `Avg(column string, dest any) error` → `AvgAsVar(column string) (float64, error)`
4. `Min(column string, dest any) error` → `MinAsVar(column string) (float64, error)`
5. `Max(column string, dest any) error` → `MaxAsVar(column string) (float64, error)`
6. `Exists(exists *bool) error` → `ExistsAsVar() (bool, error)`
7. `Pluck(column string, dest any) error` → `PluckAsVar[T any](column string) ([]T, error)`
8. `Value(column string, dest any) error` → `ValueAsVar[T any](column string) (T, error)`

#### Retrieval Methods (6 methods)
1. `First(dest any) error` → `FirstAsVar() (T, error)`
2. `FindOne(dest any) error` → `FindOneAsVar[T any]() (T, error)` (alias for FirstAsVar)
3. `Get(dest any) error` → `GetAsVar[T any]() ([]T, error)`
4. `Find(dest any, conds ...any) error` → `FindAsVar[T any](conds ...any) ([]T, error)`
5. `All(dest any) error` → `AllAsVar[T any]() ([]T, error)` (alias for GetAsVar)
6. `FindAll(dest any) error` → `FindAllAsVar[T any]() ([]T, error)` (alias for GetAsVar)

**Total: 14 sugar methods to implement**

### Implementation Phases

#### Phase 1: Core Numeric Aggregations (5 methods)
**File**: `database/query/query_aggregate.go`

**Methods to add**:
```go
// CountAsVar returns the number of records matching the query.
func (q *Query) CountAsVar() (int64, error) {
    var count int64
    err := q.Count(&count)
    return count, err
}

// SumAsVar returns the sum of the specified column.
func (q *Query) SumAsVar(column string) (float64, error) {
    var result float64
    err := q.Sum(column, &result)
    return result, err
}

// AvgAsVar returns the average of the specified column.
func (q *Query) AvgAsVar(column string) (float64, error) {
    var result float64
    err := q.Avg(column, &result)
    return result, err
}

// MinAsVar returns the minimum value of the specified column.
func (q *Query) MinAsVar(column string) (float64, error) {
    var result float64
    err := q.Min(column, &result)
    return result, err
}

// MaxAsVar returns the maximum value of the specified column.
func (q *Query) MaxAsVar(column string) (float64, error) {
    var result float64
    err := q.Max(column, &result)
    return result, err
}
```

**Tests to add**:
- Test each sugar method returns correct values
- Test error propagation
- Test with empty result sets
- Test with nil DB

**Documentation updates**:
- Update `docs/query-builder.html` with sugar method examples
- Add comparison table showing both APIs

#### Phase 2: Boolean Aggregations (1 method)
**File**: `database/query/query_aggregate.go`

**Methods to add**:
```go
// ExistsAsVar checks if any records match the query.
func (q *Query) ExistsAsVar() (bool, error) {
    var exists bool
    err := q.Exists(&exists)
    return exists, err
}
```

**Tests to add**:
- Test returns true when records exist
- Test returns false when no records exist
- Test error propagation

#### Phase 3: Column-Based Aggregations (2 methods)
**File**: `database/query/query_aggregate.go`

**Methods to add**:
```go
// PluckAsVar retrieves a single column's values from the query results.
func (q *Query) PluckAsVar[T any](column string) ([]T, error) {
    var result []T
    err := q.Pluck(column, &result)
    return result, err
}

// ValueAsVar retrieves a single column's value from the first result.
func (q *Query) ValueAsVar[T any](column string) (T, error) {
    var result T
    err := q.Value(column, &result)
    return result, err
}
```

**Tests to add**:
- Test PluckAsVar with different types (int, string, float64)
- Test ValueAsVar with different types
- Test error propagation
- Test with empty result sets

#### Phase 4: Single Record Retrieval (2 methods)
**File**: `database/query/query_first.go`

**Methods to add**:
```go
// FirstAsVar retrieves the first record matching the query.
func (q *Query) FirstAsVar() (T, error) {
    var result T
    err := q.First(&result)
    return result, err
}

// FindOneAsVar is an alias for FirstAsVar.
func (q *Query) FindOneAsVar[T any]() (T, error) {
    return q.FirstAsVar[T]()
}
```

**Tests to add**:
- Test returns correct model
- Test error propagation
- Test with relations (With)
- Test with empty result sets
- Test alias methods work identically

#### Phase 5: Multiple Record Retrieval (4 methods)
**File**: `database/query/query_get.go` and `database/query/query_find.go`

**Methods to add**:
```go
// GetAsVar retrieves all records matching the query.
func (q *Query) GetAsVar[T any]() ([]T, error) {
    var result []T
    err := q.Get(&result)
    return result, err
}

// AllAsVar is an alias for GetAsVar.
func (q *Query) AllAsVar[T any]() ([]T, error) {
    return q.GetAsVar[T]()
}

// FindAllAsVar is an alias for GetAsVar.
func (q *Query) FindAllAsVar[T any]() ([]T, error) {
    return q.GetAsVar[T]()
}

// FindAsVar retrieves records matching the given conditions.
func (q *Query) FindAsVar[T any](conds ...any) ([]T, error) {
    var result []T
    err := q.Find(&result, conds...)
    return result, err
}
```

**Tests to add**:
- Test returns correct slice of models
- Test error propagation
- Test with relations (With)
- Test with empty result sets
- Test with conditions parameter
- Test alias methods work identically

#### Phase 6: Interface Updates
**File**: `contracts/database/orm/orm.go`

**Add to Query interface**:
```go
// Sugar methods for usability
CountAsVar() (int64, error)
SumAsVar(column string) (float64, error)
AvgAsVar(column string) (float64, error)
MinAsVar(column string) (float64, error)
MaxAsVar(column string) (float64, error)
ExistsAsVar() (bool, error)
PluckAsVar[T any](column string) ([]T, error)
ValueAsVar[T any](column string) (T, error)
FirstAsVar() (T, error)
FindOneAsVar[T any]() (T, error)
GetAsVar[T any]() ([]T, error)
AllAsVar[T any]() ([]T, error)
FindAllAsVar[T any]() ([]T, error)
FindAsVar[T any](conds ...any) ([]T, error)
```

#### Phase 7: Documentation Updates

**Files to update**:
1. `docs/query-builder.html`
   - Add sugar method examples for all 14 methods
   - Add comparison table showing pointer vs sugar API
   - Add performance guidance section
   - Update all existing examples to use sugar methods where appropriate

2. `docs/api-reference.html`
   - Add sugar methods to the comprehensive method list
   - Document signatures and return types

3. `examples/README.md`
   - Add section on sugar methods
   - Show when to use each API

4. Create new example: `examples/sugar-methods/`
   - `main.go` demonstrating all sugar methods
   - `main_test.go` with comprehensive tests
   - `README.md` explaining the sugar method approach

#### Phase 8: Migration Guide

**Create**: `docs/migration-guide-sugar-methods.md`

Content:
- How to migrate existing code to sugar methods
- When to keep existing methods (performance-critical paths)
- Side-by-side comparison examples
- Common migration patterns
- Performance benchmarks

### Testing Strategy

#### Unit Tests
For each sugar method, add tests covering:
- Happy path (successful execution)
- Error propagation (database errors, build errors)
- Empty result sets
- Nil DB handling
- Type safety (for generic methods)

#### Integration Tests
- Test with all supported database drivers (MySQL, PostgreSQL, SQLite)
- Test with real models and relations
- Test in transaction context
- Test with soft deletes

#### Benchmark Tests
Add benchmarks comparing:
- Base method vs sugar method performance
- Allocation differences
- Execution time differences

Example benchmark:
```go
func BenchmarkCount(b *testing.B) {
    var count int64
    for i := 0; i < b.N; i++ {
        err := query.Count(&count)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkCountAsVar(b *testing.B) {
    for i := 0; i < b.N; i++ {
        _, err := query.CountAsVar()
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

### Success Criteria

- [ ] All 14 sugar methods implemented
- [ ] All methods added to Query interface
- [ ] Unit tests pass for all methods
- [ ] Integration tests pass for all database drivers
- [ ] Benchmarks show acceptable performance overhead (< 50ns)
- [ ] Documentation updated with examples
- [ ] Migration guide created
- [ ] Example code demonstrating sugar methods
- [ ] No breaking changes to existing API
- [ ] Code review approved

### Estimated Timeline

- Phase 1: 2-3 days
- Phase 2: 0.5 day
- Phase 3: 1-2 days
- Phase 4: 1-2 days
- Phase 5: 2-3 days
- Phase 6: 0.5 day
- Phase 7: 2-3 days
- Phase 8: 1 day

**Total estimated effort**: 10-15 days

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
count, err := db.Query().CountAsVar()
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
- Use `CountAsVar()` for application logic, API handlers
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
   - Option A: `CountAsVar()`, `FirstAsVar()`, `FindAsVar()` (AsVar suffix - self-documenting, "we handle the var")
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
count, err := db.Query().CountAsVar()

total, err := db.Query().SumAsVar("price")

user, err := db.Query().FirstAsVar()

users, err := db.Query().FindAsVar()
```

### Lines of Code Reduction
- Aggregations: 2 lines → 1 line (50% reduction)
- Single record: 2 lines → 1 line (50% reduction)
- Multiple records: 2 lines → 1 line (50% reduction)

## Implementation Status

**Status**: ? Implemented (June 14, 2026)

### What Was Implemented

All 14 sugar methods have been implemented with the AsVar naming convention:

**Aggregation Methods (8 methods)**
- CountAsVar() (int64, error)
- SumAsVar(column string) (float64, error)
- AvgAsVar(column string) (float64, error)
- MinAsVar(column string) (float64, error)
- MaxAsVar(column string) (float64, error)
- ExistsAsVar() (bool, error)
- PluckAsVar(column string) ([]any, error)
- ValueAsVar(column string) (any, error)

**Retrieval Methods (6 methods)**
- FirstAsVar() (any, error)
- FindOneAsVar() (any, error) (alias for FirstAsVar)
- GetAsVar() ([]any, error)
- AllAsVar() ([]any, error) (alias for GetAsVar)
- FindAllAsVar() ([]any, error) (alias for GetAsVar)
- FindAsVar(conds ...any) ([]any, error)

### Implementation Notes

1. **Type System**: Uses ny return type instead of generics due to Go interface method constraints
2. **Interface Updates**: All methods added to the Query interface in contracts/database/orm/orm.go
3. **Implementation Files**:
   - database/query/query_aggregate.go - Aggregation methods
   - database/query/query_first.go - Single record methods
   - database/query/query_get.go - Multiple record methods
   - database/query/query_find.go - Find method
4. **Testing**: Comprehensive test suite in database/query/query_sugar_test.go
5. **Documentation**: Migration guide and examples created

### Files Created/Modified

**Modified Files**:
- contracts/database/orm/orm.go - Added method signatures to Query interface
- database/query/query_aggregate.go - Added 8 sugar methods
- database/query/query_first.go - Added 2 sugar methods
- database/query/query_get.go - Added 3 sugar methods
- database/query/query_find.go - Added 1 sugar method

**New Files**:
- database/query/query_sugar_test.go - Comprehensive test suite
- docs/migration-guide-sugar-methods.md - Migration guide
- examples/sugar-methods/main.go - Example code
- examples/sugar-methods/README.md - Example documentation

### Backward Compatibility

? 100% backward compatible - all existing methods remain unchanged
