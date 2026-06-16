# Rename Schemer Package to Migrator

**Date**: June 16, 2026
**Status**: Implemented
**Priority**: Medium
**Author**: Neat ORM Team

## Overview

Rename the `database/schemer` package and all of its public identifiers to `database/migrator` and `Migrator*` respectively. `schemer` is a non-standard, invented name that hurts discoverability. `migrator` is the industry-standard term for a migration runner and aligns with developer expectations.

The legacy `database/migrator` package has already been deleted, so the name is available to reclaim.

## Motivation

### The Naming Problem

- **`schemer`** is not a word developers search for when looking for migration tools. It appears nowhere in the lexicon of popular ORMs or migration frameworks.
- **`migrator`** is immediately recognizable. It matches terms used in Laravel (`Migrator`), Rails (`ActiveRecord::Migration`), Django (`migrations`), and virtually every other migration system.
- **Discoverability**: `go get` users hunting for migration support will grep for `migrator`, not `schemer`.
- **API Friction**: `schemer.NewSchemer(db)` reads awkwardly. `migrator.New(db)` or `migrator.NewMigrator(db)` is self-describing.

### Why Now

- The `database/migrator` legacy package was deleted during the schemer porting effort, leaving the package path free.
- `schemer` has a manageable footprint: 9 source files, 3 example directories, and a handful of documentation references.
- The package has stabilized; no major API changes are planned, making this the right time to fix the name before wider adoption.

## Current State

### Package Files (`database/schemer/`)

| File | Purpose |
|------|---------|
| `schemer.go` | `SchemerInterface`, `SchemerImplementation`, `NewSchemer` |
| `tracker.go` | `MigrationTracker`, `MigrationStatus` |
| `signature.go` | `SignatureFormat`, `ValidateMigrationSignature` |
| `validation.go` | `isValidTableName`, `ValidateTableName` |
| `schemer_test.go` | Unit tests for core schemer logic |
| `tracker_test.go` | Tests for tracker types |
| `signature_test.go` | Tests for signature validation |
| `validation_test.go` | Tests for table name validation |
| `transaction_verification_test.go` | Transaction behavior tests |

### Public API Surface

```go
package schemer

type SchemerInterface interface { ... }
type SchemerImplementation struct { ... }
type MigrationTracker struct { ... }
type MigrationStatus struct { ... }
type SignatureFormat string

func NewSchemer(db *database.Database) SchemerInterface
func ValidateMigrationSignature(signature string, format SignatureFormat) error
func ValidateTableName(name string) error
```

### Import Sites

| Path | Count | Type |
|------|-------|------|
| `examples/schemer-migrations/` | 1 | Example + tests |
| `examples/schemer-transactions/` | 1 | Example |
| `examples/schemer-transaction-failure/` | 1 | Example |
| `docs/migrations.html` | 2 | HTML docs |
| `docs/api-reference.html` | 1 | HTML docs |
| `database/schemer/README.md` | 3 | Package README |
| `docs/proposals/completed/` | 9 | Historical proposals |

## Proposed Changes

### 1. Directory and Package Rename

```
database/
├── schemer/           # RENAME TO
├── migrator/          # NEW
```

All files move from `database/schemer/*` to `database/migrator/*`.

### 2. Package Declaration

```diff
- package schemer
+ package migrator
```

### 3. Public Identifier Renames

| Old | New |
|-----|-----|
| `SchemerInterface` | `MigratorInterface` |
| `SchemerImplementation` | `Migrator` |
| `NewSchemer` | `NewMigrator` |
| `defaultTableName` | `defaultTableName` (private, no change) |
| `SignatureFormat` | `SignatureFormat` (no change) |
| `MigrationTracker` | `MigrationTracker` (no change) |
| `MigrationStatus` | `MigrationStatus` (no change) |
| `ValidateMigrationSignature` | `ValidateMigrationSignature` (no change) |
| `ValidateTableName` | `ValidateTableName` (no change) |

**Rationale for `Migrator` (not `MigratorImplementation`)**: The struct is unexported fields + exported methods. In Go, the concrete type name `Migrator` is idiomatic when paired with `MigratorInterface`. It also avoids the stutter of `migrator.NewMigrator` if we choose to keep `NewMigrator`, or allows `migrator.New` if we prefer brevity.

### 4. Constructor Options

**Option A — `NewMigrator` (explicit, verbose):**
```go
m := migrator.NewMigrator(db)
```

**Option B — `New` (idiomatic Go, package disambiguates):**
```go
m := migrator.New(db)
```

**Recommendation**: Use `NewMigrator` to avoid collisions with `neat.New` and to be grep-friendly, but document that `migrator.New` is an acceptable alias if desired in a future iteration.

### 5. Example Directory Renames

```
examples/
├── schemer-migrations/         →  migrator-migrations/
├── schemer-transactions/       →  migrator-transactions/
└── schemer-transaction-failure/ → migrator-transaction-failure/
```

All `package main` files inside update their import path:

```diff
- "github.com/dracory/neat/database/schemer"
+ "github.com/dracory/neat/database/migrator"
```

And their variable names:

```diff
- schemerInstance := schemer.NewSchemer(db)
+ migratorInstance := migrator.NewMigrator(db)
```

### 6. Documentation Updates

- `database/schemer/README.md` → `database/migrator/README.md` (all mentions of `schemer` updated)
- `docs/migrations.html` — search/replace `schemer` → `migrator`
- `docs/api-reference.html` — update API reference entries
- `README.md` (root) — update feature bullet and migration example

### 7. Historical Proposals

Completed proposals in `docs/proposals/completed/` that reference `schemer` should be left as-is for historical accuracy. Add a note at the top of each affected proposal:

```markdown
> **Note**: The `schemer` package was renamed to `migrator` in a later release.
> Replace `database/schemer` with `database/migrator` and `schemer.NewSchemer` with `migrator.NewMigrator` when reading this proposal.
```

## Backward Compatibility

### Deprecation Shim (One Release Cycle)

Create a thin compatibility wrapper so existing users have a migration window:

```go
// database/schemer/schemer.go (temporary compatibility package)
// Deprecated: Use database/migrator instead.
package schemer

import (
    "github.com/dracory/neat/database"
    "github.com/dracory/neat/database/migrator"
)

// SchemerInterface is an alias for migrator.MigratorInterface.
// Deprecated: Use migrator.MigratorInterface.
type SchemerInterface = migrator.MigratorInterface

// SchemerImplementation is an alias for migrator.Migrator.
// Deprecated: Use migrator.Migrator.
type SchemerImplementation = migrator.Migrator

// NewSchemer is an alias for migrator.NewMigrator.
// Deprecated: Use migrator.NewMigrator.
func NewSchemer(db *database.Database) SchemerInterface {
    return migrator.NewMigrator(db)
}
```

This gives users a full minor-version cycle to migrate before the shim is deleted.

### Migration Path for Users

**Before:**
```go
import "github.com/dracory/neat/database/schemer"

s := schemer.NewSchemer(db)
s.AddMigration(&CreateUsersTable{})
_ = s.Up(ctx)
```

**After:**
```go
import "github.com/dracory/neat/database/migrator"

m := migrator.NewMigrator(db)
m.AddMigration(&CreateUsersTable{})
_ = m.Up(ctx)
```

## Benefits

1. **Discoverability**: Developers searching for "migration" or "migrator" will find the package immediately.
2. **Industry Alignment**: `migrator` is the conventional name across ORM ecosystems.
3. **Cleaner API**: `migrator.NewMigrator(db)` is more self-documenting than `schemer.NewSchemer(db)`.
4. **Reclaims a Good Name**: The `migrator` path was wasted on a deprecated package. Now it hosts the production-quality system.

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Breaking change for existing users | Provide a `database/schemer` deprecation shim for one release cycle. Document in CHANGELOG. |
| External blog posts / tutorials reference `schemer` | Add a redirect note in the root README and the old package path. Search engines will eventually update. |
| Large refactor burden | Only 9 source files and 3 example directories are affected. The change is mechanical (search/replace). |
| Confusion with old `database/migrator` | The old package was already deleted. Add a note in the new `database/migrator/README.md` explaining this is the successor to both the old `migrator` and `schemer`. |

## Acceptance Criteria

- [ ] Directory renamed: `database/schemer/` → `database/migrator/`
- [ ] Package declaration changed to `package migrator` in all files
- [ ] `SchemerInterface` renamed to `MigratorInterface`
- [ ] `SchemerImplementation` renamed to `Migrator`
- [ ] `NewSchemer` renamed to `NewMigrator`
- [ ] Example directories renamed and imports updated
- [ ] Package README rewritten with `migrator` branding
- [ ] Root README updated to reference `migrator` instead of `schemer`
- [ ] HTML documentation updated
- [ ] Deprecation shim created at `database/schemer/` (alias types + `NewSchemer`)
- [ ] All tests pass after rename
- [ ] CHANGELOG entry added with migration path for users

## Related Proposals

- [Schemer Package](completed/schemer-package.md)
- [Port Migrator Features to Schemer](completed/port-migrator-features-to-schemer.md)
- [Remove Legacy Migration Methods from Schema](remove-legacy-migration-from-schema.md)
