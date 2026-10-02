package migrator

import "time"

// Migration tracker status values stored in the status column.
const (
	MigrationTrackerStatusRunning   = "running"
	MigrationTrackerStatusCompleted = "completed"
	MigrationTrackerStatusFailed    = "failed"
	// RollbackFailed marks a migration whose Down() failed. The schema change
	// is still applied, so the migration counts as ran for Up() but remains
	// eligible for a rollback retry.
	MigrationTrackerStatusRollbackFailed = "rollback_failed"
)

// MigrationTracker represents a migration record stored in the migration_tracker table
// This is the database model/entity used for persistence
type MigrationTracker struct {
	ID           string    // The migration signature (e.g., "2024_06_15_120000_create_users_table")
	Batch        int       // Incrementing batch number grouping one Up() run
	Description  string    // The migration description from Description() method
	Status       string    // "running", "completed", "failed" or "rollback_failed" (empty = completed, legacy rows)
	ErrorMessage string    // The error text when Status is "failed"
	StartedAt    time.Time // When the migration started
	CompletedAt  time.Time // When the migration attempt finished (success or failure)
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
	Error       string    `json:"error,omitempty"` // Error message when State is "failed"
}
