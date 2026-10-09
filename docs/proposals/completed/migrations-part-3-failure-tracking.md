# Enhanced Schema Migration Interface - Part 3: Failure Tracking

**Date**: October 1, 2026
**Status**: Completed
**Completed**: October 9, 2026
**Priority**: High
**Author**: Neat ORM Team

> **Implementation note**: shipped as an append-only history (one row per
> attempt) rather than the single-row-per-migration design written here.
> The "Out of Scope" history item became the core model; all statuses and
> `Status()` semantics below are implemented as specified.

## Overview

This proposal (Part 3) extends the migration tracking system introduced in
Part 2 to record failed migration attempts in the `migration_tracker` table,
including the error message and execution timestamps. Today a failed migration
leaves no trace in the tracker, making deployment failures invisible to
`Status()` and to anyone inspecting the database.

## Motivation

### The Problem

When `Migrator.Up()` fails:

1. The tracker row is only written *after* `migration.Up()` succeeds
   (`migrator.go:238-247`), so a failed migration gets no row.
2. With transactions enabled (the default), the rollback also removes tracker
   rows written for earlier migrations in the same batch.
3. `Status()` reports the failed migration as `"pending"` — indistinguishable
   from a migration that was never attempted.

In a deployment pipeline this means the tracker gives no signal that a
migration was attempted and failed, and the underlying driver error is only
visible in the returned Go error — lost once the process exits.

### Goals

- Record every migration *attempt* in `migration_tracker` with a status
  (`running`, `completed`, `failed`).
- Persist the error message for failed attempts.
- Preserve `started_at` and `completed_at` on failure — `completed_at`
  records when the attempt finished, not just successful completions.
- Keep `Status()` honest: failed migrations report `"failed"` with the error,
  not `"pending"`.
- Retries must still work: a failed migration is re-attempted on the next
  `Up()` call.

## Schema Changes

### New Columns on `migration_tracker`

| Column          | Type           | Nullable | Purpose                                    |
|-----------------|----------------|----------|--------------------------------------------|
| `status`        | VARCHAR(16)    | yes      | `running`, `completed`, or `failed`        |
| `error_message` | TEXT           | yes      | Error text for failed attempts             |

Both columns are added via the existing `columnsToAdd` upgrade loop in
`ensureMigrationTracker` (`migrator.go:596`), so fresh tables get them at
creation and existing deployments are upgraded in place.

**Nullable by design**: adding a non-nullable column to a populated table
fails on several supported dialects without a default. Reads treat a
NULL/empty `status` as `completed` for backward compatibility with rows
written by older versions.

### Updated `MigrationTracker`

```go
type MigrationTracker struct {
    ID           string    // The migration signature
    Batch        int       // Groups the run
    Description  string
    Status       string    // "running", "completed", "failed" (empty = completed)
    ErrorMessage string    // Error text when status = "failed"
    StartedAt    time.Time
    CompletedAt  time.Time // Set on both success and failure
}
```

`MigrationStatus` is renamed to `MigrationStatusResponse` (it is a DTO, not
state) and gains an `Error` field so `Status()` can surface it. This is a
breaking rename — no alias is kept.

## Execution Flow

### `runUp` — per-migration lifecycle

```
1. Upsert tracker row: status=running, started_at=now, batch=N
2. migration.Up()
   ├─ success → status=completed, completed_at=now, error_message=NULL
   └─ failure → abort loop; transaction rolls back (removing the running row);
               AFTER rollback, write the row outside the transaction:
               status=failed, error_message=<err>, started_at/completed_at set
```

### Why the failure record is written *after* the rollback

Writing `status=failed` inside the transaction is pointless — the rollback
deletes it. `Up()` must therefore:

1. Run the transactional loop as today, capturing the failed signature,
   error, and timestamps.
2. On error, perform a second, non-transactional write to the tracker with
   the failure record, then return the original error unchanged.

When `SetTransactionsEnabled(false)` is used, the failure row is written
inline before returning — no second write needed.

### Retry semantics

- `getRanMigrations`/`isMigrationRan` only count rows with
  `status='completed'` (or empty, for legacy rows). `failed` and `running`
  rows do not block retries.
- A retry **updates the existing row** (signature is the PK — no duplicate
  insert, no history of attempts). One row per migration, always reflecting
  the latest attempt.

### Rollback path

`rollbackMigration`: if `Down()` fails, the tracker row is marked
`status='rollback_failed'` with the error. A failed rollback means the
schema change is *still applied*, so the row must keep counting as "ran"
for `Up()` (otherwise the next `Up` re-applies it) while remaining eligible
for a `Down()` retry. This is a distinct status from `failed`.

### `Status()`

- `completed` → `"completed"`
- `failed` → `"failed"` with `Error` populated
- `rollback_failed` → `"rollback_failed"` with `Error` populated
- `running` → `"running"` — signals a crashed/interrupted deployment
- Registered but absent → `"pending"` (unchanged)

Rollback-candidate queries (`getLastMigrations`, `getMigrationsByBatch`,
`runReset`) only consider rows whose migration was actually applied
(`completed`, legacy empty/NULL, `rollback_failed`) — a `failed` or
`running` Up() attempt must not be rolled back.

## Edge Cases

- **Crash mid-migration** leaves a `running` row (transactional drivers roll
  it back; non-transactional drivers may not). It does not block retries and
  is visible via `Status()` as `"running"`.
- **`Up()` succeeds but tracker write fails** — the row may show `failed`
  for a migration whose schema change was applied. Acceptable: the error is
  still returned to the caller and the drift is visible.
- **Implicit DDL commits** (MySQL `DROP TABLE`, etc.) may leave partial schema
  even when the tracker records `failed`. The status reflects the *attempt
  outcome*, not a guarantee of clean rollback — documented limitation.

## API Surface

```go
// New constants
const (
    MigrationTrackerStatusRunning        = "running"
    MigrationTrackerStatusCompleted      = "completed"
    MigrationTrackerStatusFailed         = "failed"
    MigrationTrackerStatusRollbackFailed = "rollback_failed"
)

// MigrationStatus renamed to MigrationStatusResponse (DTO), gains Error:
type MigrationStatusResponse struct {
    // ... existing fields ...
    Error string `json:"error,omitempty"` // populated when State == "failed"
}
```

No changes to `MigratorInterface` — this is internal behavior plus richer
`Status()` output.

## Implementation Notes

Files touched:

- `database/migrator/tracker.go` — new fields, status constants
- `database/migrator/migrator.go` — `ensureMigrationTracker` columns,
  `runUp` write-timing, `rollbackMigration` failure marking,
  `getRanMigrations` status filter, `Status()` error field
- `database/migrator/migrator_test.go` — tests for: failed attempt recorded
  with error + timestamps (tx on and off), retry overwrites failed row,
  `Status()` reports `failed`, rollback failure marking, legacy rows
  (empty status) treated as completed

## Out of Scope

- Per-attempt history (multiple rows per signature) — would require a
  composite/surrogate PK; can be revisited later if audit granularity is
  needed.
- Surfacing failures in a CLI — no CLI exists yet (see `cli-tools.md`).
