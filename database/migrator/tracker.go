package migrator

import "time"

// nullTime is the NOT NULL "no value" sentinel used for columns that have no
// meaningful timestamp yet — e.g. completed_at on a running row. It mirrors
// neat.NullDateTime ("0002-01-01 00:00:00") as a time.Time and sorts before
// any real timestamp.
var nullTime = time.Date(2, 1, 1, 0, 0, 0, 0, time.UTC)

// MigrationStatus is the status value stored in the status column.
// A named type so only the declared constants can be assigned to
// MigrationTracker.Status — a misspelled literal is a compile error.
type MigrationStatus string

// Migration tracker status values stored in the status column.
const (
	MigrationStatusRunning   MigrationStatus = "running"
	MigrationStatusCompleted MigrationStatus = "completed"
	MigrationStatusFailed    MigrationStatus = "failed"
	// RolledBack marks a migration that was rolled back via Down().
	MigrationStatusRolledBack MigrationStatus = "rolled_back"
	// RollbackFailed marks a migration whose Down() failed. The schema change
	// is still applied, so the migration counts as ran for Up() but remains
	// eligible for a rollback retry.
	MigrationStatusRollbackFailed MigrationStatus = "rollback_failed"
)

// MigrationTracker represents a migration record stored in the migration_tracker table
// This is the database model/entity used for persistence. The table is an
// append-only history: each Up or Down attempt adds a row, and a migration's
// current state is its most recent row.
//
// Legacy note: rows written before the status column existed have Status = ""
// and are treated as "completed" at the read sites. This compatibility case
// (Status == "") should be removed after October 2027, once deployments have
// had a year to write a status on every new row.
type MigrationTracker struct {
	ID           string          // Unique, time-ordered record ID for this execution attempt
	Migration    string          // The migration signature (e.g., "2024_06_15_120000_create_users_table")
	Batch        int             // Incrementing batch number grouping one Up() run; rollback rows keep the batch they reverse
	Description  string          // The migration description from Description() method
	Status       MigrationStatus // one of the MigrationStatus* values (empty = completed, legacy rows — see note below)
	StartedAt    time.Time       // When the migration started
	CompletedAt  time.Time       // When the migration attempt finished (success or failure)
	ErrorMessage string          // The error text when Status is "failed" or "rollback_failed"
}

// MigrationName returns the migration signature, falling back to ID if Migration is empty (legacy rows).
func (t MigrationTracker) MigrationName() string {
	if t.Migration != "" {
		return t.Migration
	}
	return t.ID
}

// MigrationStatusResponse is the DTO returned by Migrator.Status(), derived
// from MigrationTracker rows and the registered (pending) migrations.
type MigrationStatusResponse struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Batch       int       `json:"batch"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
	State       string    `json:"state"`           // "pending", "running", "completed", "failed", "rollback_failed"
	Error       string    `json:"error,omitempty"` // Error message when State is "failed" or "rollback_failed"
}
