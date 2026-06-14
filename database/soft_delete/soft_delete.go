package soft_delete

import (
	"time"
)

const (
	// SoftDeleteAtColumn is the default column name used for soft deletes.
	// The default is "soft_deleted_at" — semantically explicit about what the column tracks.
	// Use DeletedAt embed for "deleted_at" (Laravel-style) compatibility.
	SoftDeleteAtColumn = "soft_deleted_at"

	// DeletedAtColumnName is the column name used by the DeletedAt embed (Laravel-compatible).
	DeletedAtColumnName = "deleted_at"

	// DeletedAtColumn is an alias for SoftDeleteAtColumn.
	//
	// Deprecated: Use SoftDeleteAtColumn instead.
	DeletedAtColumn = SoftDeleteAtColumn
)

// SoftDeletes provides soft delete functionality for models using the "soft_deleted_at" column.
// This is the default embed for new projects — the column name is semantically explicit.
//
// To use the Laravel-compatible "deleted_at" column, embed DeletedAt instead.
//
// Example:
//
//	type User struct {
//	    soft_delete.SoftDeletes  // uses "soft_deleted_at"
//	    ID   uint
//	    Name string
//	}
type SoftDeletes struct {
	SoftDeletedAt *time.Time `json:"soft_deleted_at,omitempty" db:"soft_deleted_at"`
}

// SoftDeletedAtColumn returns the soft delete column name used in database queries.
// Implements the SoftDeleteColumnNamer interface.
func (sd *SoftDeletes) SoftDeletedAtColumn() string {
	return "soft_deleted_at"
}

// DeletedAtColumn returns the soft delete column name used in database queries.
//
// Deprecated: Use SoftDeletedAtColumn() instead.
func (sd *SoftDeletes) DeletedAtColumn() string {
	return sd.SoftDeletedAtColumn()
}

// IsSoftDeleted returns true if the model has been soft deleted.
func (sd *SoftDeletes) IsSoftDeleted() bool {
	return sd.SoftDeletedAt != nil
}

// IsDeleted returns true if the model has been soft deleted.
//
// Deprecated: Use IsSoftDeleted() instead.
func (sd *SoftDeletes) IsDeleted() bool {
	return sd.IsSoftDeleted()
}

// SoftDelete marks the model as soft-deleted by setting the soft_deleted_at timestamp.
func (sd *SoftDeletes) SoftDelete() {
	now := time.Now()
	sd.SoftDeletedAt = &now
}

// Delete marks the model as soft-deleted by setting the soft_deleted_at timestamp.
//
// Deprecated: Use SoftDelete() instead.
func (sd *SoftDeletes) Delete() {
	sd.SoftDelete()
}

// RestoreSoftDeleted marks the model as not soft-deleted by setting soft_deleted_at to nil.
func (sd *SoftDeletes) RestoreSoftDeleted() {
	sd.SoftDeletedAt = nil
}

// Restore marks the model as not soft-deleted by setting soft_deleted_at to nil.
//
// Deprecated: Use RestoreSoftDeleted() instead.
func (sd *SoftDeletes) Restore() {
	sd.RestoreSoftDeleted()
}

// GetSoftDeletedAt returns the soft_deleted_at timestamp.
func (sd *SoftDeletes) GetSoftDeletedAt() *time.Time {
	return sd.SoftDeletedAt
}

// GetDeletedAt returns the soft_deleted_at timestamp.
//
// Deprecated: Use GetSoftDeletedAt() instead.
func (sd *SoftDeletes) GetDeletedAt() *time.Time {
	return sd.GetSoftDeletedAt()
}

// SoftDeletedAt provides soft delete functionality using the "soft_deleted_at" column name.
// Identical to SoftDeletes — provided for explicit naming consistency with CreatedAt/UpdatedAt.
//
// Example:
//
//	type User struct {
//	    soft_delete.SoftDeletedAt  // uses "soft_deleted_at"
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

// DeletedAt provides soft delete functionality using the "deleted_at" column name.
// Use this embed for Laravel-compatible schemas or any existing schema that uses "deleted_at".
//
// Example:
//
//	type User struct {
//	    soft_delete.DeletedAt  // uses "deleted_at" (Laravel-compatible)
//	    ID   uint
//	    Name string
//	}
type DeletedAt struct {
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// SoftDeletedAtColumn returns the soft delete column name used in database queries.
// Implements the SoftDeleteColumnNamer interface.
func (sd *DeletedAt) SoftDeletedAtColumn() string {
	return "deleted_at"
}

// IsSoftDeleted returns true if the model has been soft deleted.
func (sd *DeletedAt) IsSoftDeleted() bool {
	return sd.DeletedAt != nil
}

// IsDeleted returns true if the model has been soft deleted.
//
// Deprecated: Use IsSoftDeleted() instead.
func (sd *DeletedAt) IsDeleted() bool {
	return sd.IsSoftDeleted()
}

// SoftDelete marks the model as soft-deleted by setting the deleted_at timestamp.
func (sd *DeletedAt) SoftDelete() {
	now := time.Now()
	sd.DeletedAt = &now
}

// Delete marks the model as soft-deleted by setting the deleted_at timestamp.
//
// Deprecated: Use SoftDelete() instead.
func (sd *DeletedAt) Delete() {
	sd.SoftDelete()
}

// RestoreSoftDeleted marks the model as not soft-deleted by setting deleted_at to nil.
func (sd *DeletedAt) RestoreSoftDeleted() {
	sd.DeletedAt = nil
}

// Restore marks the model as not soft-deleted by setting deleted_at to nil.
//
// Deprecated: Use RestoreSoftDeleted() instead.
func (sd *DeletedAt) Restore() {
	sd.RestoreSoftDeleted()
}

// GetSoftDeletedAt returns the deleted_at timestamp.
func (sd *DeletedAt) GetSoftDeletedAt() *time.Time {
	return sd.DeletedAt
}

// GetDeletedAt returns the deleted_at timestamp.
//
// Deprecated: Use GetSoftDeletedAt() instead.
func (sd *DeletedAt) GetDeletedAt() *time.Time {
	return sd.GetSoftDeletedAt()
}
