# Proposal: Track Updated Fields (Dirty Tracking)

This proposal outlines several ways to implement "Dirty Tracking" in the Neat ORM, allowing for efficient updates by only sending modified fields to the database.

## Overview

Currently, Neat's `Save` method updates all non-zero fields of a struct. While this works for most cases, it can lead to:
1. Unnecessary database traffic.
2. Accidental overwrites of data changed by other processes.
3. Inability to "unset" fields (e.g., setting a string to empty) if the ORM treats empty as "don't update".

Inspired by `dataobject`, we propose three potential implementations.

---

## Proposal A: Automatic Snapshotting (Transparent)

The ORM automatically takes a snapshot of the model when it is loaded from the database and compares it with the current state during the update.

### Mechanism
- When `Scan` is called, the ORM stores a copy of the retrieved data.
- This snapshot can be stored in a hidden field in the model (if it implements a specific interface) or in a registry associated with the query/transaction.
- Upon `Save`, the ORM uses reflection to compare the current struct fields with the snapshot.

### Example Usage
```go
user := models.User{}
db.Model(&user).First(&user, 1)

user.Name = "John Doe"
// user.Age remains 30

db.Save(&user)
// SQL: UPDATE users SET name = 'John Doe', updated_at = '...' WHERE id = 1
```

### Breaking Changes
- **None**. Existing code remains exactly the same.

### Method Signatures
- No public signature changes.
- Internal changes to `database/query/query_scan.go` and `database/query/query_save.go`.

### Pros
- **Zero Boilerplate**: Works "out of the box" for all models.
- **Developer Experience**: Feels like standard Go; no need to use getters/setters.

### Cons
- **Memory Overhead**: Storing a snapshot for every loaded model.
- **CPU Overhead**: Reflection-based comparison on every save.
- **Complexity**: Managing the lifecycle of snapshots (especially when models are passed across different query instances).

---

## Proposal B: Explicit Tracking (Getter/Setter)

Following the `dataobject` pattern, models use private fields and public methods to interact with data. The setter method explicitly marks the field as "dirty".

### Mechanism
- Models embed a `orm.Trackable` struct.
- Fields are kept private.
- Setters call `o.MarkDirty("field_name")`.
- `Save` only processes fields present in the "dirty" map.

### Example Usage
```go
type User struct {
    orm.Model
    orm.Trackable
    name string
}

func (u *User) SetName(name string) {
    u.name = name
    u.MarkDirty("name")
}

// ... in application code
user := models.NewUser()
db.First(user, 1)
user.SetName("John Doe")
db.Save(user)
```

### Breaking Changes
- **High**. Requires users to rewrite their models if they want to use this feature.
- Method signatures for `Save` might need to check for a `Trackable` interface.

### Method Signatures
- New interface `contractsorm.Trackable`.
- New base struct `orm.Trackable` with `MarkDirty`, `IsDirty`, `GetDirty`, etc.

### Pros
- **Performance**: Extremely fast. No comparison needed; the ORM knows exactly what changed.
- **Explicit**: No magic; the developer has full control.
- **Functionality**: Easily allows setting fields to their zero values (e.g., `""` or `0`).

### Cons
- **Boilerplate**: Significant amount of code needed for each model (Getters/Setters).
- **Not "Idiomatic" Go**: Many Go developers prefer public fields on structs for simple data models.

---

## Proposal C: Manual Marking / Selective Update

A middle-ground where the developer manually specifies which fields have changed, or uses a fluent API to restrict the update.

### Mechanism
- Add a fluent method `OnlyColumns(...string)` or `UpdateOnly(&model, "field1", "field2")`.

### Example Usage
```go
user.Name = "John Doe"
db.Model(&user).Only("name").Save(&user)
```

### Breaking Changes
- **None**.

### Method Signatures
- Add `Only(columns ...string) Query` to the `Query` interface.

### Pros
- **Simple to implement**.
- **No overhead** when not used.

### Cons
- **Error-Prone**: Easy to forget to include a column.
- **Tedious**: Developer must manually keep track of what they changed.

---

## Recommendation

We recommend **Proposal A** as the primary implementation because it aligns with the Neat philosophy of being developer-friendly and "magical" in a good way.

However, to address the performance concerns, we can implement it as an **Opt-in** feature:
1. If a model embeds `orm.DirtyTracking`, the ORM will perform snapshotting.
2. If it doesn't, it falls back to the current "update all non-zero" behavior.

We can also provide the methods from **Proposal B** (`MarkDirty`) as a way for developers to manually override or optimize the process.
