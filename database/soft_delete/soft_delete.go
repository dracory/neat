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

// DeletedAtColumn returns the soft delete column name used in database queries.
// Implements the SoftDeleteColumnNamer interface.
func (sd *SoftDeletes) DeletedAtColumn() string {
	return "deleted_at"
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

// SoftDeletesAlt provides soft delete functionality using the "soft_deleted_at" column name.
// Use this instead of SoftDeletes when your schema uses "soft_deleted_at" for semantic clarity.
//
// Example:
//
//	type User struct {
//	    soft_delete.SoftDeletesAlt
//	    ID   uint
//	    Name string
//	}
type SoftDeletesAlt struct {
	SoftDeletedAt *time.Time `json:"soft_deleted_at,omitempty" db:"soft_deleted_at"`
}

// DeletedAtColumn returns the soft delete column name used in database queries.
// Implements the SoftDeleteColumnNamer interface.
func (sd *SoftDeletesAlt) DeletedAtColumn() string {
	return "soft_deleted_at"
}

// IsDeleted returns true if the model has been soft deleted.
func (sd *SoftDeletesAlt) IsDeleted() bool {
	return sd.SoftDeletedAt != nil
}

// Delete marks the model as deleted by setting the soft_deleted_at timestamp.
func (sd *SoftDeletesAlt) Delete() {
	now := time.Now()
	sd.SoftDeletedAt = &now
}

// Restore marks the model as not deleted by setting soft_deleted_at to nil.
func (sd *SoftDeletesAlt) Restore() {
	sd.SoftDeletedAt = nil
}

// GetDeletedAt returns the soft_deleted_at timestamp.
func (sd *SoftDeletesAlt) GetDeletedAt() *time.Time {
	return sd.SoftDeletedAt
}
