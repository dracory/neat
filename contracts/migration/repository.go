package migration

import "time"

type File struct {
	ID          uint
	Migration   string
	Batch       int
	Description string
	StartedAt   time.Time
	CompletedAt time.Time
}

type Repository interface {
	// CreateRepository Create the migration repository data store.
	CreateRepository() error
	// Delete Remove a migration from the log.
	Delete(migration string) error
	// DeleteRepository Delete the migration repository data store.
	DeleteRepository() error
	// GetLast Get the last migration batch.
	GetLast() ([]File, error)
	// GetMigrations Get the completed migrations.
	GetMigrations() ([]File, error)
	// GetMigrationsByBatch Get the list of the migrations by batch.
	GetMigrationsByBatch(batch int) ([]File, error)
	// GetMigrationsByStep Get the list of migrations.
	GetMigrationsByStep(steps int) ([]File, error)
	// GetNextBatchNumber Get the next migration batch number.
	GetNextBatchNumber() (int, error)
	// GetRan Get the completed migrations.
	GetRan() ([]string, error)
	// Log that a migration was run.
	Log(file string, batch int, description string, startedAt, completedAt time.Time) error
	// RepositoryExists Determine if the migration repository exists.
	RepositoryExists() bool
	// GetHistory Get the migration execution history from the database.
	GetHistory() ([]File, error)
}
