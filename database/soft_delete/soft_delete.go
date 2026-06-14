package soft_delete

import (
	"time"
)

const (
	// SoftDeleteAtColumn is the default column name used for soft deletes.
	SoftDeleteAtColumn = "deleted_at"

	// DeletedAtColumn is the default column name used for soft deletes.
	//
	// Deprecated: Use SoftDeleteAtColumn instead.
	DeletedAtColumn = SoftDeleteAtColumn
)

// SoftDeletes provides soft delete functionality for models using the "deleted_at" column.
// Models that embed this struct will have soft delete capabilities.
//
// Example:
//
//	type User struct {
//	    soft_delete.SoftDeletes
//	    ID   uint
//	    Name string
//	}
type SoftDeletes struct {
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// SoftDeletedAtColumn returns the soft delete column name used in database queries.
// Implements the SoftDeleteColumnNamer interface.
func (sd *SoftDeletes) SoftDeletedAtColumn() string {
	return "deleted_at"
}

// DeletedAtColumn returns the soft delete column name used in database queries.
//
// Deprecated: Use SoftDeletedAtColumn() instead.
func (sd *SoftDeletes) DeletedAtColumn() string {
	return sd.SoftDeletedAtColumn()
}

// IsSoftDeleted returns true if the model has been soft deleted.
func (sd *SoftDeletes) IsSoftDeleted() bool {
	return sd.DeletedAt != nil
}

// IsDeleted returns true if the model has been soft deleted.
//
// Deprecated: Use IsSoftDeleted() instead.
func (sd *SoftDeletes) IsDeleted() bool {
	return sd.IsSoftDeleted()
}

// SoftDelete marks the model as soft-deleted by setting the deleted_at timestamp.
func (sd *SoftDeletes) SoftDelete() {
	now := time.Now()
	sd.DeletedAt = &now
}

// Delete marks the model as soft-deleted by setting the deleted_at timestamp.
//
// Deprecated: Use SoftDelete() instead.
func (sd *SoftDeletes) Delete() {
	sd.SoftDelete()
}

// RestoreSoftDeleted marks the model as not soft-deleted by setting deleted_at to nil.
func (sd *SoftDeletes) RestoreSoftDeleted() {
	sd.DeletedAt = nil
}

// Restore marks the model as not soft-deleted by setting deleted_at to nil.
//
// Deprecated: Use RestoreSoftDeleted() instead.
func (sd *SoftDeletes) Restore() {
	sd.RestoreSoftDeleted()
}

// GetSoftDeletedAt returns the deleted_at timestamp.
func (sd *SoftDeletes) GetSoftDeletedAt() *time.Time {
	return sd.DeletedAt
}

// GetDeletedAt returns the deleted_at timestamp.
//
// Deprecated: Use GetSoftDeletedAt() instead.
func (sd *SoftDeletes) GetDeletedAt() *time.Time {
	return sd.GetSoftDeletedAt()
}

// SoftDeletedAt provides soft delete functionality using the "soft_deleted_at" column name.
// Use this instead of SoftDeletes when your schema uses "soft_deleted_at" for semantic clarity,
// following the same naming convention as the CreatedAt and UpdatedAt embeds.
//
// Example:
//
//	type User struct {
//	    soft_delete.SoftDeletedAt
//	    ID   uint
//	    Name string
//	}
type SoftDeletedAt struct {
	SoftDeletedAt *time.Time `json:"soft_deleted_at,omitempty" db:"soft_deleted_at"`
}

// SoftDeletedAtColumn returns the soft delete column name used in database queries.
// Implements the SoftDeleteColumnNamer interface.
func (sd *SoftDeletedAt) SoftDeletedAtColumn() string {
	return "soft_deleted_at"
}

// DeletedAtColumn returns the soft delete column name used in database queries.
//
// Deprecated: Use SoftDeletedAtColumn() instead.
func (sd *SoftDeletedAt) DeletedAtColumn() string {
	return sd.SoftDeletedAtColumn()
}

// IsSoftDeleted returns true if the model has been soft deleted.
func (sd *SoftDeletedAt) IsSoftDeleted() bool {
	return sd.SoftDeletedAt != nil
}

// IsDeleted returns true if the model has been soft deleted.
//
// Deprecated: Use IsSoftDeleted() instead.
func (sd *SoftDeletedAt) IsDeleted() bool {
	return sd.IsSoftDeleted()
}

// SoftDelete marks the model as soft-deleted by setting the soft_deleted_at timestamp.
func (sd *SoftDeletedAt) SoftDelete() {
	now := time.Now()
	sd.SoftDeletedAt = &now
}

// Delete marks the model as soft-deleted by setting the soft_deleted_at timestamp.
//
// Deprecated: Use SoftDelete() instead.
func (sd *SoftDeletedAt) Delete() {
	sd.SoftDelete()
}

// RestoreSoftDeleted marks the model as not soft-deleted by setting soft_deleted_at to nil.
func (sd *SoftDeletedAt) RestoreSoftDeleted() {
	sd.SoftDeletedAt = nil
}

// Restore marks the model as not soft-deleted by setting soft_deleted_at to nil.
//
// Deprecated: Use RestoreSoftDeleted() instead.
func (sd *SoftDeletedAt) Restore() {
	sd.RestoreSoftDeleted()
}

// GetSoftDeletedAt returns the soft_deleted_at timestamp.
func (sd *SoftDeletedAt) GetSoftDeletedAt() *time.Time {
	return sd.SoftDeletedAt
}

// GetDeletedAt returns the soft_deleted_at timestamp.
//
// Deprecated: Use GetSoftDeletedAt() instead.
func (sd *SoftDeletedAt) GetDeletedAt() *time.Time {
	return sd.GetSoftDeletedAt()
}
