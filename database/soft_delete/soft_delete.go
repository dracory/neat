package soft_delete

import (
	"time"
)

const (
	// DeletedAtColumn is the default column name for soft deletes
	DeletedAtColumn = "deleted_at"
)

// SoftDeletes provides soft delete functionality for models.
// Models that embed this struct will have soft delete capabilities.
type SoftDeletes struct {
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// IsDeleted returns true if the model has been soft deleted.
func (sd *SoftDeletes) IsDeleted() bool {
	return sd.DeletedAt != nil
}

// Delete marks the model as deleted by setting the deleted_at timestamp.
func (sd *SoftDeletes) Delete() {
	now := time.Now()
	sd.DeletedAt = &now
}

// Restore marks the model as not deleted by setting deleted_at to nil.
func (sd *SoftDeletes) Restore() {
	sd.DeletedAt = nil
}

// GetDeletedAt returns the deleted_at timestamp.
func (sd *SoftDeletes) GetDeletedAt() *time.Time {
	return sd.DeletedAt
}
