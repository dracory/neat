# Migration System Improvements

**Date**: June 16, 2026
**Status**: Proposed
**Priority**: Medium
**Author**: Neat ORM Team

## Overview

This proposal identifies targeted improvements to the `database/migrator` package based on real operational pain points, not feature-parity chasing. Each item below answers the question: *what makes the migration system safer or easier to use for Neat ORM developers?*

## Problem Statement

The `database/migrator` package works well for most use cases, but several friction points have emerged from real usage:

1. **Registration order bugs**: Migrations run in registration order, which is error-prone when migrations are registered from multiple packages or files
2. **Silent duplicates**: Adding the same migration twice succeeds silently, leading to confusing status output and hard-to-debug behavior
3. **Incomplete status**: `Status()` only shows applied migrations, making it impossible to see what's pending without external tracking
4. **Verbose configuration**: Setting up a migrator requires multiple setter calls, leading to boilerplate

## What We Are NOT Changing

After review, the following items from the standalone `dracory/migrate` package are intentionally **not** being adopted because they do not add product value to an ORM-integrated migration system:

| Standalone Feature | Why Not Adopted |
|-------------------|-----------------|
| **Per-migration transactions** | Batch transactions (current default) are the correct behavior for deployments. A failed migration batch should roll back entirely, leaving the database in a known pre-deployment state |
| **Logging integration** | Neat is a library, not a CLI tool. Callers already receive errors and can log themselves. Adding a logger interface adds API surface with no net benefit |
| **`GetHistory()` API** | `Status()` already returns tracker records with timing data. A separate history method would be redundant |
| **Environment variables** | Libraries should not read env vars. Applications read configuration and pass it to the library. This is the caller's responsibility |

## Proposed Changes

### 1. Lexicographical Ordering ✅

**Status**: Implemented

Migrations are now sorted by signature before execution, regardless of registration order. This eliminates a class of ordering bugs when migrations are registered from multiple packages.

```go
migrator.AddMigration(&CreatePostsTable{})      // 2024_06_15_120100_create_posts_table
migrator.AddMigration(&CreateUsersTable{})      // 2024_06_15_120000_create_users_table

// Users table runs first, then posts table, regardless of registration order
```

**Opt-out available**:
```go
migrator.SetLexicographicalOrdering(false)
```

---

### 2. Duplicate Detection

**Status**: Implemented

**Current Behavior**: `AddMigration()` and `AddMigrations()` silently append duplicates.

**Real Problem**: In a real codebase, migrations might be registered from multiple initialization paths (e.g., package `init()` functions, test setup, multiple feature branches merged together). When the same migration is registered twice, the current behavior silently accepts it. This leads to:
- Confusing `Status()` output showing the same migration ID
- Unclear whether the migration will run once or twice
- Difficult debugging when a migration unexpectedly fails on the second (duplicate) run

**Solution**: `AddMigration()` returns an error when a duplicate signature is detected.

```go
func (s *Migrator) AddMigration(migration MigrationInterface) error {
    signature := migration.Signature()
    if signature == "" {
        return fmt.Errorf("migration signature cannot be empty")
    }

    for _, existing := range s.migrations {
        if existing.Signature() == signature {
            return fmt.Errorf("duplicate migration signature: %s", signature)
        }
    }

    s.migrations = append(s.migrations, migration)
    return nil
}
```

**Impact**: This is a **behavior change** that may break code that was accidentally relying on duplicate registration. However, duplicate registration is a bug, and failing fast is the correct behavior.

**Benefits**:
- Catches real bugs at setup time instead of at migration time
- Eliminates confusing status output
- Prevents unexpected re-execution

---

### 3. Complete Status Reporting

**Status**: Implemented

**Current Behavior**: `Status()` only returns records from the `migration_tracker` table. If you have registered migrations that have not yet run, there is no way to see them through the migrator API.

**Real Problem**: When building deployment tooling or health checks, you need to answer: "what migrations are pending?" Currently you must compare `migrator.Status()` output against your registered migration list yourself.

**Solution**: `Status()` returns both applied and pending migrations, with a `State` field indicating which is which.

```go
type MigrationStatus struct {
    ID          string    `json:"id"`
    Description string    `json:"description"`
    Batch       int       `json:"batch"`
    StartedAt   time.Time `json:"started_at"`
    CompletedAt time.Time `json:"completed_at"`
    State       string    `json:"state"` // "completed" or "pending"
}
```

**Implementation**:
```go
func (s *Migrator) Status() ([]MigrationStatus, error) {
    var statuses []MigrationStatus

    // Get completed migrations from tracker
    ranMigrations := make(map[string]bool)
    if s.db.Schema().HasTable(s.tableName) {
        trackers, err := s.getMigrations()
        if err != nil {
            return nil, err
        }
        for _, t := range trackers {
            ranMigrations[t.ID] = true
            statuses = append(statuses, MigrationStatus{
                ID:          t.ID,
                Description: t.Description,
                Batch:       t.Batch,
                StartedAt:   t.StartedAt,
                CompletedAt: t.CompletedAt,
                State:       "completed",
            })
        }
    }

    // Add pending migrations from registered list
    for _, migration := range s.migrations {
        sig := migration.Signature()
        if !ranMigrations[sig] {
            statuses = append(statuses, MigrationStatus{
                ID:          sig,
                Description: migration.Description(),
                State:       "pending",
            })
        }
    }

    // Sort by signature for consistent output
    sort.Slice(statuses, func(i, j int) bool {
        return statuses[i].ID < statuses[j].ID
    })

    return statuses, nil
}
```

**Impact**: This is a **behavioral change** -- previously `Status()` only returned tracker records. Now it also returns pending migrations. This is more correct and useful, but code that assumed `Status()` only returned applied migrations may need adjustment.

**Benefits**:
- Single call answers "what's the full migration state?"
- Enables deployment health checks and dashboards
- Consistent with how migration tools should behave

---

### 4. Configurable Constructor

**Status**: Implemented

**Current Behavior**: `NewMigrator(db)` returns a migrator with defaults. All customization requires subsequent setter calls.

**Real Problem**: Setting up a migrator with non-default options requires boilerplate:

```go
// Current verbose setup
migrator := migrator.NewMigrator(db)
migrator.SetTableName("my_migrations")
migrator.SetSignatureValidation(true, migrator.SignatureFormatDateTime)
migrator.SetTransactionIsolationLevel("SERIALIZABLE")
```

**Solution**: Add an `Options` struct and a single-call constructor.

```go
type Options struct {
    TableName                  string
    TransactionIsolationLevel  string
    LexicographicalOrdering    bool
    SignatureValidationEnabled bool
    SignatureValidationFormat  SignatureFormat
}

func NewMigratorWithOptions(db *database.Database, opts *Options) (MigratorInterface, error) {
    m := NewMigrator(db)

    if opts == nil {
        return m, nil
    }

    if opts.TableName != "" {
        if err := m.SetTableName(opts.TableName); err != nil {
            return nil, err
        }
    }

    if opts.TransactionIsolationLevel != "" {
        m.SetTransactionIsolationLevel(opts.TransactionIsolationLevel)
    }

    m.SetLexicographicalOrdering(opts.LexicographicalOrdering)

    if opts.SignatureValidationEnabled {
        m.SetSignatureValidation(true, opts.SignatureValidationFormat)
    }

    return m, nil
}
```

**Usage**:
```go
migrator, err := migrator.NewMigratorWithOptions(db, &migrator.Options{
    TableName:               "my_migrations",
    LexicographicalOrdering: true,
})
```

**Benefits**:
- Reduces boilerplate
- Enables configuration from application config structs
- Works naturally with dependency injection
- `NewMigrator(db)` continues to work unchanged

## Migration Path

### Phase 1: Non-Breaking Additions

- [x] **Lexicographical Ordering** -- *Implemented June 16, 2026*
- [x] **Configurable Constructor** -- `NewMigratorWithOptions()` and `Options` struct

### Phase 2: Behavior Changes

- [x] **Duplicate Detection** -- `AddMigration()` returns error on duplicate signature
- [x] **Complete Status Reporting** -- `Status()` returns both `completed` and `pending` migrations

### Phase 3: Documentation

- [x] Update `database/migrator/README.md`
- [ ] Update examples (no examples need changes -- `NewMigrator` still works)
- [ ] Update API reference (HTML docs to be regenerated from source)

## Backward Compatibility

| Change | Breaking? | Notes |
|--------|-----------|-------|
| Lexicographical Ordering | No | `SetLexicographicalOrdering(false)` preserves old behavior |
| Configurable Constructor | No | New method; `NewMigrator()` unchanged |
| Duplicate Detection | **Yes** | Code with duplicate registrations will now error. This is a bug fix |
| Complete Status | **Yes** | `Status()` returns additional pending records. Code assuming only completed records may need adjustment |

## Acceptance Criteria

### Duplicate Detection

- [x] `AddMigration()` returns error when signature already exists
- [x] `AddMigrations()` returns error if any signature is duplicate
- [x] Empty signature returns error
- [x] First registration succeeds, second registration of same migration fails

### Complete Status

- [x] `Status()` returns both completed and pending migrations
- [x] Completed migrations have `state: "completed"`, `batch`, `started_at`, `completed_at`
- [x] Pending migrations have `state: "pending"`, no timing fields
- [x] Results are sorted by signature
- [x] Empty tracker table returns all registered migrations as pending

### Configurable Constructor

- [x] `NewMigratorWithOptions(db, nil)` is equivalent to `NewMigrator(db)`
- [x] `NewMigratorWithOptions(db, &Options{})` applies all defaults
- [x] All Options fields are applied correctly
- [x] Invalid `TableName` returns error
- [x] `NewMigrator(db)` continues to work unchanged

## Why These Four?

| Feature | Solves Real Problem? | Adds Maintenance Burden? | Worth It? |
|---------|---------------------|--------------------------|-----------|
| Lexicographical Ordering | Ordering bugs across packages | Low | Yes |
| Duplicate Detection | Silent failures, confusing status | Low | Yes |
| Complete Status | Can't see pending migrations | Low | Yes |
| Configurable Constructor | Boilerplate in every project | Low | Yes |
| Logging | Caller can already log | Medium (new interface) | No |
| Context in Up/Down | Top-level ctx is sufficient | Medium (new interface) | No |
| History API | Redundant with Status() | Low | No |
| Env Vars | Library shouldn't read env | Low | No |
| Per-migration tx | Batch tx is correct default | High (new modes, testing) | No |
