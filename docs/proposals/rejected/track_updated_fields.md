# Proposal: Track Updated Fields (Dirty Tracking)

**Date**: June 15, 2026
**Status**: Rejected
**Priority**: Nice to Have
**Author**: Neat ORM Team

This proposal outlines how to implement "dirty tracking" in the Neat ORM, allowing efficient updates by sending only modified fields to the database.

**Rejection Reason**: Very low impact - existing methods (`Update()`, `Select()`) already solve all dirty tracking use cases effectively. The complexity of implementation outweighs the benefits.

## Problem Statement

The current `Save` method updates all non-zero fields of a struct. This works for most cases but can be optimized:

```go
user := User{Name: "John", Age: 30, Email: "john@example.com"}
db.First(&user, 1)

user.Name = "Jane"
db.Save(&user)
// Current: UPDATE users SET name = 'Jane', age = 30, email = 'john@example.com' WHERE id = 1
// Optimized: UPDATE users SET name = 'Jane' WHERE id = 1
```

With dirty tracking:
- Reduced database traffic
- Better handling of zero values (empty strings, 0, false)
- Cleaner UPDATE statements

## Priority: Nice to Have

**This is a convenience feature, not essential functionality.** Neat ORM already handles all dirty tracking use cases through existing methods.

### Current Workarounds (Work Today)

| Use Case | Current Solution | Example |
|----------|------------------|---------|
| Update specific fields | `Update()` with map | `db.Where("id = ?", 1).Update(map[string]any{"name": "Jane"})` |
| Update with zero value | `Update()` with map | `Update(map[string]any{"count": 0})` |
| Partial update via Save | `Select()` + `Save()` | `db.Select("name").Save(&user)` |

### When This Feature Helps

Dirty tracking becomes convenient when:
- Models have **10+ fields** and manual `Select()` field lists become tedious
- You modify fields across **multiple code paths** and tracking changes is complex
- You want `Save()` to "just work" without specifying which fields changed

### When This Feature Is Unnecessary

Skip this implementation if:
- Most models have **3-5 fields** (easy to track manually)
- You're comfortable with explicit `Update()` or `Select()` calls
- You prefer keeping the ORM **simple and minimal**

---

## Inspiration: `dracory/dataobject`

The [dracory/dataobject](https://github.com/dracory/dataobject) library demonstrates an elegant dirty tracking pattern using `Set()` and `Hydrate()` methods:

```go
func (do *DataObject) Set(key string, value string) {
    do.data[key] = value
    do.dataChanged[key] = value  // automatic tracking
}

func (do *DataObject) Hydrate(data map[string]string) {
    do.data = data  // load without marking dirty
}
```

**Adapting for Neat ORM**: Since Neat uses native Go structs with typed fields (not `map[string]string`), we can adapt this pattern to work with struct scanning and reflection-based comparison.

---

## Proposal A: Automatic Snapshotting (Recommended)

Models implement a `Snapshotter` interface. The ORM automatically stores a copy when loading and compares on save.

### Implementation

```go
// contracts/database/orm/snapshotter.go
package orm

type Snapshotter interface {
    SetSnapshot(data map[string]any)
    GetSnapshot() map[string]any
    ClearSnapshot()
}
```

### Usage

```go
type User struct {
    orm.Model
    Name  string `db:"name"`
    Email string `db:"email"`
    Age   int    `db:"age"`
    snapshot map[string]any
}

func (u *User) SetSnapshot(data map[string]any) { u.snapshot = data }
func (u *User) GetSnapshot() map[string]any     { return u.snapshot }
func (u *User) ClearSnapshot()                  { u.snapshot = nil }

// Usage - no changes to application code
user := User{}
db.First(&user, 1)  // Snapshot stored automatically
user.Name = "Jane"
db.Save(&user)        // Only name is updated
```

### PROS

| Benefit | Description |
|---------|-------------|
| **Transparent** | Zero changes to application code after interface implementation |
| **Opt-in** | Only models implementing `Snapshotter` get the behavior; no breaking changes |
| **Familiar** | Similar pattern to `json.Marshaler`, `sql.Scanner` |
| **Zero-value support** | Can update fields to `0`, `""`, `false` - the snapshot shows they changed |
| **Automatic** | Works without manual `MarkDirty()` calls in every setter |
| **Type-safe** | Maintains Go's type system (unlike map-based approaches) |

### CONS

| Consideration | Mitigation |
|---------------|------------|
| **Memory overhead** | One extra map per tracked instance; cleared after `Save` |
| **CPU overhead** | Reflection-based comparison on save; limited to `Snapshotter` models |
| **Partial loading** | If `Select()` loads partial fields, snapshot may be incomplete; document or skip snapshotting |
| **Interface boilerplate** | ~10 lines of boilerplate per model; could be code-generated |

### Implementation Notes

Internal changes needed:
- `query_scan.go`: Call `SetSnapshot` after successful scan
- `query_save.go`: Detect `Snapshotter`, compare fields, build partial UPDATE
- `builder_update.go`: Generate UPDATE with only changed columns

### Handling Partial Loading

If `Select()` is used to load partial fields, the ORM can either:
1. Skip snapshotting when partial fields loaded (documented limitation)
2. Store which fields were loaded and only compare those

---

## Proposal B: Hybrid Approach (Manual + Automatic)

Combine automatic snapshotting with optional manual control via embedded struct.

### Implementation

```go
// database/orm/trackable.go
type Trackable struct {
    dirtyFields map[string]bool
}

func (t *Trackable) MarkDirty(fields ...string) {
    for _, f := range fields {
        t.dirtyFields[f] = true
    }
}

func (t *Trackable) IsDirty() bool {
    return len(t.dirtyFields) > 0
}

func (t *Trackable) GetDirtyFields() []string { ... }
func (t *Trackable) ClearDirty() { t.dirtyFields = nil }
```

### Usage

```go
type User struct {
    orm.Model
    orm.Trackable  // embedded
    Name  string
    Email string
}

user := User{}
db.First(&user, 1)

user.Name = "Jane"
user.MarkDirty("name")  // explicit control

db.Save(&user)
```

### PROS

| Benefit | Description |
|---------|-------------|
| **Explicit control** | Developer decides exactly which fields are "dirty" |
| **No reflection** | Slightly faster than snapshot comparison at save time |
| **Partial model friendly** | Works well with `Select()` partial loading |
| **Composable** | Can be combined with `Snapshotter` for hybrid approach |

### CONS

| Consideration | Mitigation |
|---------------|------------|
| **Manual tracking** | Easy to forget `MarkDirty()`, causing silent bugs |
| **Boilerplate** | Must call `MarkDirty()` in every setter or after every field change |
| **String field names** | Prone to typos; no compile-time checking |

### When to Use

- When you want explicit control over which fields are considered "changed"
- When working with partial models where automatic comparison isn't desired
- As a fallback for models that need custom dirty tracking logic

---

## Existing Alternative: `Select()` Method

For ad-hoc partial updates, the existing `Select()` method already works:

```go
user := User{ID: 1, Name: "Jane"}
db.Select("name").Save(&user)
// SQL: UPDATE users SET name = 'Jane' WHERE id = 1
```

### PROS

| Benefit | Description |
|---------|-------------|
| **Already works** | No implementation needed |
| **Explicit** | Clear intent in code |
| **Zero overhead** | No memory or CPU overhead |

### CONS

| Consideration | Description |
|---------------|-------------|
| **Manual** | Must specify fields every time |
| **Not automatic** | Doesn't solve general dirty tracking use case |
| **Easy to forget** | If you change multiple fields, must update `Select()` call |

This is useful for one-off updates but doesn't solve the general dirty tracking use case.

---

## Comparison Matrix

| Aspect | Proposal A (Snapshot) | Proposal B (Trackable) | Existing `Select` |
|--------|------------------------|------------------------|-------------------|
| **Boilerplate** | Interface methods (~10 lines) | Embed + `MarkDirty` calls everywhere | Per-query |
| **Automatic** | Yes | No (manual) | No (manual) |
| **Performance** | Reflection on save | No reflection | Optimal |
| **Zero values** | Supported | Supported | Supported |
| **Partial loading** | Edge case | Works fine | N/A |
| **Forgotten tracking** | Automatic | Silent bug | N/A |

---

## Recommendation

**Implement Proposal A (Snapshotter)** as the primary mechanism.

### Why Snapshotter

1. **Least intrusive**: Models only need ~10 lines of interface implementation
2. **Automatic**: Works transparently once set up
3. **Type-safe**: Maintains Go's type system (unlike `dataobject` map approach)
4. **Flexible**: Opt-in per model via interface

### Implementation Plan

1. **Add `Snapshotter` interface** to `contracts/database/orm/`
2. **Modify `Save()`** to detect and use snapshot comparison
3. **Add helper utilities** for common snapshot operations
4. **Document** the interface and benefits

### Example Migration

```go
// Before - no dirty tracking
type User struct {
    orm.Model
    Name  string `db:"name"`
    Email string `db:"email"`
}

// After - with dirty tracking
type User struct {
    orm.Model
    Name  string `db:"name"`
    Email string `db:"email"`
    snapshot map[string]any
}

func (u *User) SetSnapshot(data map[string]any) { u.snapshot = data }
func (u *User) GetSnapshot() map[string]any     { return u.snapshot }
func (u *User) ClearSnapshot()                  { u.snapshot = nil }
```

Application code requires **no changes**.

---

## Implementation Details

### Save Method Behavior

```go
func (q *Query) Save(value any) error {
    // Check if model implements Snapshotter
    if snapshotter, ok := value.(Snapshotter); ok {
        if snapshot := snapshotter.GetSnapshot(); snapshot != nil {
            // Compare and build partial UPDATE
            changed := getChangedFields(value, snapshot)
            if len(changed) > 0 {
                // Build UPDATE only with changed fields
                return q.updateOnly(value, changed)
            }
            // No changes - optionally update timestamps only
            return nil
        }
    }
    // Fall back to current behavior
    return q.saveAllNonZero(value)
}
```

### Zero Value Updates

With snapshotting, zero values work correctly:

```go
user := User{ID: 1}
db.First(&user)  // user.Count = 5 from DB

user.Count = 0  // Setting to zero
db.Save(&user)  // UPDATE users SET count = 0 WHERE id = 1
// Snapshot shows count changed from 5 to 0
```

---

## Appendix: Related Proposals

- [Alternative Soft Delete Support](alternative-soft-delete-support.md)
- [Alternative Soft Delete Column Naming](alternative-soft-delete-column-naming.md)

## References

1. [dracory/dataobject](https://github.com/dracory/dataobject) - Go data object library with automatic dirty tracking
2. [Doctrine ORM Change Tracking Policies](https://www.doctrine-project.org/projects/doctrine-orm/en/2.17/reference/change-tracking-policies.html)
