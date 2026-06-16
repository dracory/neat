package migrator

import "time"

// MigrationTracker represents a migration record stored in the migration_tracker table
// This is the database model/entity used for persistence
type MigrationTracker struct {
	ID          string    // The migration signature (e.g., "2024_06_15_120000_create_users_table")
	Batch       int       // Timestamp ID (YYYYMMDDHHMMSS). Groups the run
	Description string    // The migration description from Description() method
	StartedAt   time.Time // When the migration started
	CompletedAt time.Time // When the migration finished
}

// MigrationStatus represents the status of a migration returned to users
// This is a DTO/response type derived from MigrationTracker data
type MigrationStatus struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Batch       int       `json:"batch"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
	State       string    `json:"state"` // "pending", "completed", "failed"
}
