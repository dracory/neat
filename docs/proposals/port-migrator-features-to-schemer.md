# Port Migrator Features to Schemer

**Date**: June 15, 2026
**Status**: In Progress
**Priority**: Medium
**Author**: Neat ORM Team

## Overview

The `database/migrator` package is deprecated in favor of the new interface-based `database/schemer` package. Before we delete `migrator`, we should port its valuable features into `schemer`. This proposal identifies the features worth saving and defines the work required to port them.

## Background

### Why Port?

`migrator` accumulated several robust features over its lifetime that `schemer` currently lacks or implements incompletely:

- **Security**: table name validation to prevent SQL injection
- **Validation**: migration signature format validation (datetime, date, unix, custom)
- **Observability**: structured logging of migration runs with duration tracking
- **Robustness**: repository schema upgrades, reset safety limits, driver-aware table existence checks
- **Configurability**: configurable migration tracker table name

Dropping `migrator` without porting these would be a functional regression.

## Features to Port

### 1. Configurable Tracker Table Name ✅

**Status**: Implemented

**Current state in schemer**: hardcoded `"migration_tracker"`.
**Migrator approach**: reads from `config.GetString("database.migrations.table", "migrations")`.

**Implementation**: Added `defaultTableName` constant, `tableName` field to `SchemerImplementation`, and `SetTableName(name string) error` method with validation. All hardcoded references replaced with `s.tableName`.

**API**:
```go
schemer := schemer.NewSchemer(db)
schemer.SetTableName("my_migrations") // optional, validated
```

### 2. Table Name SQL Injection Guard ✅

**Status**: Implemented (together with Feature 1)

**Current state in schemer**: none.
**Migrator approach**: `isValidMigrationTableName()` rejects SQL keywords, enforces alphanumeric + underscore only, and checks length.

**Implementation**: Ported `isValidMigrationTableName` from `migrator` as `isValidTableName` and `ValidateTableName` in `database/schemer/validation.go`. Applied in `SetTableName` method. Rejects empty names, names starting with digits, non-alphanumeric/underscore characters, and SQL keywords.

### 3. Migration Signature Format Validation ✅

**Status**: Implemented

**Current state in schemer**: only checks empty string and `>255` characters.
**Migrator approach**: `ValidateMigrationID()` supports datetime, date, unix, and custom formats; validates calendar dates, times, sequences, and descriptions.

**Implementation**: Ported `ValidateMigrationID` as `ValidateMigrationSignature` in `database/schemer/signature.go`. Added `SignatureFormat` type and constants (`DateTime`, `Date`, `Unix`, `Custom`). Added `SetSignatureValidation(enabled bool, format SignatureFormat)` to `SchemerInterface`. Default is disabled (opt-in, no breaking change). When enabled, signatures are validated before migration execution in `runUp`.

**API**:
```go
schemer.SetSignatureValidation(true, schemer.SignatureFormatDateTime)
```

### 4. Repository Schema Upgrades ✅

**Status**: Implemented

**Current state in schemer**: `runUp` creates the table if `!HasTable`, but never evolves an existing schema.
**Migrator approach**: `upgradeRepositorySchema()` checks for missing columns (`description`, `started_at`, `completed_at`) and adds them via `schema.Table`.

**Implementation**: Extracted `ensureMigrationTracker(schema)` helper in `schemer.go`. It creates the table if missing, then checks each expected column with `schema.HasColumn` and adds missing ones via `schema.Table`. Called at the start of every `runUp`. Each column addition is independent — if one fails, the others are still attempted.

### 5. Driver-Aware `getAllTables()` for `Fresh()` ✅

**Status**: Implemented

**Current state in schemer**: `getAllTables()` is a stub returning `[]string{}`, so `Fresh()` drops nothing.
**Migrator approach**: `Fresh()` uses `schema.GetTableListing()` and driver-specific `information_schema` / `sqlite_master` checks.

**Implementation**: Replaced the stub with `schema.GetTableListing()` (delegated to driver-specific grammars). Changed `getAllTables()` to accept a `contractsschema.Schema` parameter so it uses the transaction-aware schema when called inside `Fresh()`, avoiding SQLite connection-locking issues. Filters out the migration tracking table. `TestFresh` was un-skipped and now passes.

### 6. Structured Logging / Observability

**Current state in schemer**: completely silent.
**Migrator approach**: `log(level, message, fields)` prints structured output with migration name, batch, duration, and errors.

**Proposal**: Add an optional `Logger` callback or interface to `SchemerImplementation`:

```go
type MigrationLogger interface {
    LogMigrationStarted(signature string, batch int)
    LogMigrationCompleted(signature string, batch int, duration time.Duration)
    LogMigrationFailed(signature string, batch int, err error)
}
```

Or accept a simple callback `SetLogFunc(func(level, message string, fields map[string]any))`. At minimum, emit start, completion (with duration), and failure events.

### 7. `Reset()` Safety Limit

**Current state in schemer**: `Reset()` calls `runReset` which loops over all migrations with no upper bound.
**Migrator approach**: `maxIterations := 1000` guard prevents infinite loops if tracker deletion fails.

**Proposal**: Add `const maxResetIterations = 1000` and enforce it in `runReset`. Return an explicit error if the limit is reached.

### 8. Sequential Batch Numbering

**Current state in schemer**: `getNextBatchNumber()` returns `int(time.Now().Unix())`.
**Migrator approach**: `MAX(batch) + 1` — simple, monotonic, human-friendly.

**Proposal**: Switch to sequential batch numbering (`MAX(batch) + 1`, starting at `1`). Unix timestamps are unwieldy for humans reading migration status. This is a minor behavioral change but improves DX.

## What NOT to Port

| Feature | Reason |
|---------|--------|
| Global migration registry (`RegisterMigration`) | Global state is an anti-pattern; `schemer`'s explicit `AddMigration` is cleaner |
| File-based discovery (`filepath.Walk`) | Intentional design difference — `schemer` is interface-based, not file-based |
| `Create()` file generator | Belongs in a CLI tool or generator, not the core migration runner library |
| Per-migration `Begin()`/`Commit()` | `schemer` already has superior whole-operation transaction wrapping via `WithTransaction` |

## Implementation Plan

| Phase | Feature | Estimated Effort |
|-------|---------|----------------|
| ~~1~~ | ~~Configurable table name + validation~~ | ~~Small~~ | **Done** |
| ~~2~~ | ~~Signature format validation utility~~ | ~~Small~~ | **Done** |
| ~~3~~ | ~~Repository schema upgrades~~ | ~~Small~~ | **Done** |
| ~~4~~ | ~~Driver-aware `getAllTables()` + `Fresh()` fix~~ | ~~Medium~~ | **Done** |
| 5 | Logging / observability hooks | Small |
| 6 | Reset safety limit + sequential batch numbering | Small |
| 7 | Delete `database/migrator` package | Small |

## Acceptance Criteria

- [x] `schemer` supports configurable and validated tracker table names
- [x] `schemer` can validate migration signatures in datetime/date/unix/custom formats
- [x] `schemer` upgrades existing `migration_tracker` schemas on startup (adds missing columns)
- [x] `Fresh()` correctly drops all user tables (not just the tracker) and re-runs migrations
- [ ] `Reset()` has an iteration safety limit
- [ ] Batch numbers are sequential (`1, 2, 3...`) instead of Unix timestamps
- [ ] `schemer` emits structured log events for migration start, completion, and failure
- [ ] `database/migrator` is deleted without functional regression

## Related Proposals

- [Schemer Package](completed/schemer-package.md)
- [Schemer Transaction Support](completed/schemer-transaction-support.md)
- [Enhanced Migration System](completed/enhanced-migration-system.md)
