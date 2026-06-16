# Remove Legacy Migration Methods from Schema Contract

**Date**: June 16, 2026
**Status**: Implemented
**Priority**: Medium
**Author**: Neat ORM Team

## Overview

Remove the legacy migration registration surface from the `Schema` contract and its implementation. Specifically, delete `Register([]MigrationInterface)` and `Migrations() []MigrationInterface` from `contracts/database/schema/schema.go` and the supporting field from `database/schema/schema.go`. The `schemer` package has completely superseded this pattern.

## Motivation

### Why Remove Now?

The `schemer` package was introduced (see [Schemer Package](completed/schemer-package.md)) to decouple migration management from schema building. The legacy methods on `Schema` were kept temporarily for backward compatibility. After several releases and full feature parity in `schemer`, the time has come to clean up the contract.

### Problems with Keeping Them

1. **Conceptual Leakage**: `Schema` is a schema builder. Holding a slice of `MigrationInterface` and providing registration is not a schema-building concern.
2. **Hidden State**: The `migrations` field on `Schema` is mutable state that is never used by the library itself anymore (only by user code and old examples).
3. **API Confusion**: New users see `schema.Register()` and assume it is the correct way to run migrations, when `schemer.NewSchemer(db).AddMigration(...)` is the intended path.
4. **Constructor Bloat**: `NewSchema` accepts a `migrations []contractsschema.MigrationInterface` parameter that is always passed as `nil` in internal code.

## Current State

### Contract (`contracts/database/schema/schema.go`)

```go
type Schema interface {
    // ... schema methods ...

    // Migrations Get the migrations.
    Migrations() []MigrationInterface

    // Register migrations.
    Register([]MigrationInterface)

    // ... more schema methods ...
}
```

### Implementation (`database/schema/schema.go`)

```go
type Schema struct {
    // ... other fields ...
    migrations []contractsschema.MigrationInterface
    // ... other fields ...
}

func NewSchema(
    config config.Config,
    log log.Log,
    orm contractsorm.Orm,
    migrations []contractsschema.MigrationInterface, // always nil internally
) (*Schema, error) {
    // ...
}

func (r *Schema) Migrations() []contractsschema.MigrationInterface {
    return r.migrations
}

func (r *Schema) Register(migrations []contractsschema.MigrationInterface) {
    for _, migration := range migrations {
        migration.SetSchema(r)
    }
    r.migrations = migrations
}
```

### Internal Callers

| File | Usage |
|------|-------|
| `database/db.go:165` | `schema.NewSchema(..., nil)` |
| `database/db.go:291` | `schema.NewSchema(..., nil)` |
| `database/schema/schema.go:118` | `NewSchema(..., r.migrations)` (connection switching) |

### External / Example Callers

| File | Usage |
|------|-------|
| `examples/schemer-migrations/main_test.go` | `schema.Register(migrations)` in 4 test functions |
| `database/schema/migration_manager.go` | Already deprecated; should be deleted together |

## Proposed Changes

### 1. Remove from Contract

Delete these two methods from `contracts/database/schema/schema.go`:

```diff
-    // Migrations Get the migrations.
-    Migrations() []MigrationInterface
-
-    // Register migrations.
-    Register([]MigrationInterface)
```

### 2. Remove from Implementation

In `database/schema/schema.go`:

- Remove `migrations []contractsschema.MigrationInterface` field from `Schema` struct.
- Remove `migrations` parameter from `NewSchema` constructor.
- Remove `Migrations()` method.
- Remove `Register()` method.
- Update `Connection()` to pass `nil` removal (no longer needed).

### 3. Update Internal Callers

Update `database/db.go` to remove the `nil` argument:

```diff
-    s, err := schema.NewSchema(database.config, database.logger, database.ormInstance, nil)
+    s, err := schema.NewSchema(database.config, database.logger, database.ormInstance)
```

### 4. Delete Legacy Migration Manager

`database/schema/migration_manager.go` is already marked deprecated. Delete it alongside this cleanup. It references `Schema.Migrations()` indirectly but is unused in production code.

### 5. Update Examples and Tests

The `examples/schemer-migrations/main_test.go` tests call `schema.Register()` to inject the schema into migrations manually. Replace with explicit `SetSchema` calls:

**Before:**
```go
schema := db.Schema()
schema.Register(migrations)
```

**After:**
```go
s := db.Schema()
for _, m := range migrations {
    m.SetSchema(s)
}
```

Or, better, rewrite the examples to use `schemer`:

```go
sc := schemer.NewSchemer(db)
sc.AddMigrations(migrations)
_ = sc.Up(ctx)
```

## Migration Path for Users

### If You Were Using `schema.Register()`

**Before:**
```go
schema := db.Schema()
schema.Register(migrations)
for _, m := range schema.Migrations() {
    _ = m.Up()
}
```

**After:**
```go
sc := schemer.NewSchemer(db)
sc.AddMigrations(migrations)
_ = sc.Up(ctx)
```

### If You Need Manual Schema Injection

`MigrationInterface` still exposes `SetSchema(Schema)`. Call it directly:

```go
for _, m := range migrations {
    m.SetSchema(db.Schema())
}
```

## Benefits

1. **Smaller Contract**: `Schema` interface focuses purely on schema operations.
2. **Less Hidden State**: No mutable `migrations` slice on the schema struct.
3. **Clearer Guidance**: Users are nudged toward the `schemer` package, which provides tracking, rollback, transactions, and batch numbering.
4. **Simpler Constructor**: `NewSchema` loses an always-`nil` parameter.

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Breaking change for users calling `schema.Register()` | Document in CHANGELOG; provide migration path above; this is a major/minor version bump. |
| Breaking change for users calling `schema.Migrations()` | Same as above; `schemer.Status()` provides equivalent observability. |
| `examples/schemer-migrations/main_test.go` tests break | Update tests to use `schemer` or manual `SetSchema` as outlined above. |

## Acceptance Criteria

- [x] `Register` and `Migrations` removed from `contracts/database/schema/schema.go`
- [x] `migrations` field, `Register`, `Migrations`, and parameter removed from `database/schema/schema.go`
- [x] `NewSchema` signature updated in all call sites (`database/db.go`)
- [x] `database/schema/migration_manager.go` deleted
- [x] Examples and tests updated to compile without `schema.Register()`
- [x] CHANGELOG entry added documenting the removal and migration path

## Related Proposals

- [Schemer Package](completed/schemer-package.md)
- [Port Migrator Features to Schemer](completed/port-migrator-features-to-schemer.md)
- [Enhanced Migration System](completed/enhanced-migration-system.md)
