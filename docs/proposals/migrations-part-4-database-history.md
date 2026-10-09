# Migrations Part 4: `database_history` — One Append-Only Table for Migrations and Seeders

**Date**: October 9, 2026
**Status**: Proposed
**Priority**: Medium
**Author**: Neat ORM Team

## Overview

The `migration_tracker` table introduced in Parts 2–3 has outgrown its name.
It is now an append-only history of schema *and* data changes — migrations,
rollbacks, and (after this proposal) seeders. Rename it to `database_history`
and let `seeder.Runner` write into the same table, so every durable change to
the database appears in one ordered timeline.

## Motivation

Seeders already follow the migration pattern in practice: a `Signature()`,
an idempotent `Run()`, and a "did this run already?" question. Today
`Runner.CallOnce` answers that question with an in-memory map — durable
tracking would let seeders behave exactly like run-once migrations:

- A seeder that inserts reference data must run **after** the migration that
  creates the table. With one table and timestamp-based signatures, the two
  interleave naturally in a single ordered history.
- Locking, batching, status, and failure recording are implemented once, not
  duplicated per kind of change.
- `Status()` answers "what has happened to this database?" completely —
  including which seed data was loaded and which attempts failed.

This is the camp-1 design used by Django (`RunPython` data migrations in
`django_migrations`), EF Core (`HasData` seeds inside `__EFMigrationsHistory`),
Flyway (versioned + repeatable scripts in one history), and Liquibase
(changesets in `DATABASECHANGELOG`). Camp 2 (Laravel/Rails/Prisma-style
untracked, always-run seeders) stays possible by simply not registering
seeders for tracking.

## Goals

- Rename the tracking table to `database_history`, with a safe upgrade path
  for existing deployments.
- Add a `kind` column distinguishing `migration` and `seed` rows.
- `Runner.Call`/`CallOnce` record durable history rows; `CallOnce` becomes
  durable across process restarts (reads the table, not an in-memory map).
- Rollback operations (`Down`, `RollbackSteps`, `RollbackToBatch`, `Reset`,
  `Fresh`) skip `kind = 'seed'` rows — most seeds have no meaningful `Down()`,
  and recording a fake `rolled_back` row would be noise.

## Non-Goals

- Repeatable/re-runnable seeds (Flyway `R__`-style checksum-triggered re-runs).
  Run-once semantics only; editing a seeder after it ran means writing a new
  seeder, same as editing a migration. Revisit with a `checksum` column only
  if a real need emerges.
- Per-environment filtering (Liquibase-style contexts). Dev fixtures vs.
  production reference data is handled by the caller registering dev seeders
  only when `env == dev`.
- Removing `Seeder`-only behaviors that don't fit the history model.

## Schema Changes

### Table rename: `migration_tracker` → `database_history`

`defaultTableName` becomes `"database_history"`. `SetTableName` and
`Options.TableName` continue to allow a custom name.

Upgrade path inside `ensureMigrationTracker` (renamed
`ensureDatabaseHistory`):

1. If `database_history` exists → run the existing `columnsToAdd` upgrade loop.
2. Else if `migration_tracker` exists → rename it via `schema.Rename`
   (driver-supported `RENAME TABLE`; fallback: create `database_history`,
   copy all rows ordered by `started_at`/`id`, drop the old table), then run
   `columnsToAdd`.
3. Else → create `database_history` fresh.

The append-only model makes the copy fallback lossless: every row is
immutable history.

### New column

| Column | Type        | Nullable | Purpose                                  |
|--------|-------------|----------|------------------------------------------|
| `kind` | VARCHAR(16) | yes      | `migration` or `seed` (empty/NULL = migration, legacy rows) |

Added via the existing `columnsToAdd` mechanism — nullable by design, same
rationale as `status`/`error_message` in Part 3. Empty `kind` is treated as
`migration` at the read boundary (same legacy-empty pattern, same removal
policy).

### Column rename: `migration` → `signature`

The column now holds seeder signatures too, so `migration` is a misnomer.
Rename during the table-upgrade step (add `signature`, backfill from
`migration`, drop the old column where the dialect supports it; otherwise
keep reading both with `signature` preferred). This is optional — keeping the
`migration` column name is acceptable if the rename proves painful across
dialects.

## Seeder Integration

`seeder.Runner` gains an optional history writer:

```go
type SeederRunner struct {
    // ...
    history HistoryWriter // nil = untracked (today's behavior)
}

// Contracts
type HistoryWriter interface {
    // Record a durable, append-only entry for a signature.
    Recorded(signature string) bool          // has a completed/failed row
    RecordRun(signature, description string) error // insert run row
}
```

`Call`/`CallOnce` write a `kind = 'seed'` row per execution — with the same
`status`/`error_message`/`started_at`/`completed_at` semantics as migration
rows (a failed seeder leaves a `failed` row with the error). `CallOnce`
checks `history.Recorded(signature)` when a history writer is attached,
making "once" durable across restarts rather than process-local.

Seeders opt in explicitly: `NewRunner()` stays untracked;
`NewRunnerWithHistory(db)` (or `SetHistoryWriter`) enables tracking. Dev-only
seeders are simply not registered on non-dev environments.

## Rollback & Status Semantics

- `activeTrackers` / `latestTrackers` stay unchanged; they operate on
  signatures regardless of `kind`.
- `runRollbackSteps` / `runRollbackToBatch` / `runReset` filter
  `kind = 'migration'` (or empty) before calling `rollbackMigration`. Seed
  rows are never rolled back and never get `rolled_back` rows.
- `Status()` can either include seed rows (with a `Kind` field on
  `MigrationStatusResponse`) or accept a filter. Recommended: add
  `Kind string 'json:"kind"'` to the DTO — consumers see the full timeline
  and can ignore seeds if they only care about schema.
- `Fresh` keeps the same policy: drop all tables, append `rolled_back` rows
  for applied migrations; seed rows are marked `rolled_back` too (their data
  is gone after the drop), then `runUp` replays migrations followed by
  tracked seeders in signature order.

## Open Questions

1. Do seeders participate in `batch` numbering? Simplest: they share the
   sequence (a seeder run after `Up` gets the same or next batch), which keeps
   `RollbackToBatch` meaningful. Alternative: batch 0 / NULL for seeds —
   simpler but loses ordering info for `RollbackToBatch`.
2. Should `Status()` include unregistered-but-tracked seed signatures?
   Consistent with today's unregistered-migration handling, yes.
3. `Rename table` dialect coverage — confirm `schema.Rename` exists across
   all supported grammars; otherwise the copy-fallback becomes the only path.

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Existing deployments have `migration_tracker` | Automatic rename/copy in `ensure*` on first run; old table dropped only after successful copy |
| Renamed public constant `defaultTableName` is internal — `SetTableName` already covers customization | Low blast radius |
| Seed failures poison the history for retries | Same model as `failed` migrations: next call retries; `rollback_failed` analogue unnecessary (seeds don't roll back) |
| Dev fixtures polluting production history | Caller-side: register dev seeders only under `env == dev` (documented in README) |

## Tasks

1. Rename `defaultTableName`, `ensureMigrationTracker` → `ensureDatabaseHistory`,
   add `kind` column to create/upgrade paths.
2. Implement rename-or-copy upgrade from `migration_tracker`.
3. Optional: rename `migration` column → `signature`.
4. `contracts/database/seeder`: add history-aware runner hooks; wire
   `Call`/`CallOnce` to write `kind = 'seed'` rows.
5. `MigrationStatusResponse.Kind`; `Status()` handling for seed rows.
6. Rollback paths skip `kind = 'seed'`; `Fresh` handles seed history.
7. Tests: upgrade path, seeder tracking, durable `CallOnce`, rollback skips
   seeds, mixed timeline ordering.
8. Update `database/migrator/README.md` + `database/seeder` docs.

## Alternative Considered

Keep two tables (`migration_tracker` + `seeder_history`). Rejected: it
duplicates the append-only machinery, splits the timeline (ordering a seed
between two migrations requires cross-table timestamp merging), and the name
`migration_tracker` stays misleading either way — the append-only rewrite in
this branch already made it a history table, not a state table.
