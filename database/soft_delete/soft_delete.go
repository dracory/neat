package soft_delete

import (
	"time"

	"github.com/dracory/neat/database/schema/constants"
)

// Re-export constants for backward compatibility
const (
	// SoftDeleteAtColumn is the default column name used for soft deletes.
	// The default is "soft_deleted_at" — semantically explicit about what the column tracks.
	// Use DeletedAtColumnName for "deleted_at" (Laravel-style) compatibility.
	// Deprecated: Use constants.SoftDeleteAtColumn instead.
	SoftDeleteAtColumn = constants.SoftDeleteAtColumn

	// DeletedAtColumnName is the column name used by the DeletedAt embed (Laravel-compatible).
	// Deprecated: Use constants.DeletedAtColumnName instead.
	DeletedAtColumnName = constants.DeletedAtColumnName

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
	return constants.SoftDeleteAtColumn
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
	return constants.SoftDeleteAtColumn
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
	return constants.DeletedAtColumnName
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

// MaxSoftDeletedAt is the sentinel "not deleted" value for the max-date strategy.
// Records are considered active when soft_deleted_at > NOW(), and deleted when soft_deleted_at <= NOW().
// Using a far-future date (9999-12-31 23:59:59 UTC) instead of NULL allows:
//   - NOT NULL column constraints
//   - Better partial index support (range scans vs IS NULL)
//   - Simpler query conditions without NULL handling
var MaxSoftDeletedAt = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)

// SoftDeletesMaxDate provides soft delete functionality using a max-date sentinel
// instead of NULL. The column used is "soft_deleted_at".
// Embed this in models where the schema enforces NOT NULL on timestamp columns.
//
// Records are soft-deleted when soft_deleted_at is in the past (<= NOW()),
// and active when soft_deleted_at is in the future (e.g., MaxSoftDeletedAt).
//
// Example:
//
//	type User struct {
//	    soft_delete.SoftDeletesMaxDate  // uses "soft_deleted_at" with sentinel
//	    ID   uint
//	    Name string
//	}
type SoftDeletesMaxDate struct {
	SoftDeletedAt time.Time `json:"soft_deleted_at,omitempty" db:"soft_deleted_at"`
}

// SoftDeletedAtColumn returns the soft delete column name used in database queries.
// Implements the SoftDeleteColumnNamer interface.
func (s *SoftDeletesMaxDate) SoftDeletedAtColumn() string {
	return constants.SoftDeleteAtColumn
}

// DeletedAtColumn returns the soft delete column name used in database queries.
//
// Deprecated: Use SoftDeletedAtColumn() instead.
func (s *SoftDeletesMaxDate) DeletedAtColumn() string {
	return s.SoftDeletedAtColumn()
}

// IsSoftDeleted returns true if the model has been soft deleted.
// A record is soft-deleted when soft_deleted_at <= NOW().
func (s *SoftDeletesMaxDate) IsSoftDeleted() bool {
	return !s.SoftDeletedAt.After(time.Now())
}

// IsDeleted returns true if the model has been soft deleted.
//
// Deprecated: Use IsSoftDeleted() instead.
func (s *SoftDeletesMaxDate) IsDeleted() bool {
	return s.IsSoftDeleted()
}

// SoftDelete marks the model as soft-deleted by setting soft_deleted_at to NOW().
func (s *SoftDeletesMaxDate) SoftDelete() {
	s.SoftDeletedAt = time.Now()
}

// Delete marks the model as soft-deleted by setting soft_deleted_at to NOW().
//
// Deprecated: Use SoftDelete() instead.
func (s *SoftDeletesMaxDate) Delete() {
	s.SoftDelete()
}

// RestoreSoftDeleted marks the model as not soft-deleted by setting soft_deleted_at to MaxSoftDeletedAt.
func (s *SoftDeletesMaxDate) RestoreSoftDeleted() {
	s.SoftDeletedAt = MaxSoftDeletedAt
}

// Restore marks the model as not soft-deleted by setting soft_deleted_at to MaxSoftDeletedAt.
//
// Deprecated: Use RestoreSoftDeleted() instead.
func (s *SoftDeletesMaxDate) Restore() {
	s.RestoreSoftDeleted()
}

// GetSoftDeletedAt returns the soft_deleted_at timestamp.
func (s *SoftDeletesMaxDate) GetSoftDeletedAt() time.Time {
	return s.SoftDeletedAt
}

// GetDeletedAt returns the soft_deleted_at timestamp.
//
// Deprecated: Use GetSoftDeletedAt() instead.
func (s *SoftDeletesMaxDate) GetDeletedAt() time.Time {
	return s.GetSoftDeletedAt()
}

// SoftDeleteValue returns the value to write on soft delete.
// Implements the SoftDeleteStrategy interface.
func (s *SoftDeletesMaxDate) SoftDeleteValue() any {
	return time.Now()
}

// RestoreValue returns the value to write on restore.
// Implements the SoftDeleteStrategy interface.
func (s *SoftDeletesMaxDate) RestoreValue() any {
	return MaxSoftDeletedAt
}

// SoftDeletedCondition returns the SQL fragment + args for the "only soft deleted" filter.
// Implements the SoftDeleteStrategy interface.
func (s *SoftDeletesMaxDate) SoftDeletedCondition(quoteIdentifier func(string) string) (string, []any) {
	return quoteIdentifier(constants.SoftDeleteAtColumn) + " <= ?", []any{time.Now()}
}

// NotSoftDeletedCondition returns the SQL fragment + args for the "not soft deleted" filter.
// Implements the SoftDeleteStrategy interface.
func (s *SoftDeletesMaxDate) NotSoftDeletedCondition(quoteIdentifier func(string) string) (string, []any) {
	return quoteIdentifier(constants.SoftDeleteAtColumn) + " > ?", []any{time.Now()}
}

// DeletedAtMaxDate provides soft delete functionality using a max-date sentinel
// with the "deleted_at" column name (Laravel-compatible).
// Use this when your schema uses "deleted_at" and enforces NOT NULL.
//
// Records are soft-deleted when deleted_at is in the past (<= NOW()),
// and active when deleted_at is in the future (e.g., MaxSoftDeletedAt).
//
// Example:
//
//	type Post struct {
//	    soft_delete.DeletedAtMaxDate  // uses "deleted_at" with sentinel
//	    ID    uint
//	    Title string
//	}
type DeletedAtMaxDate struct {
	DeletedAt time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// SoftDeletedAtColumn returns the soft delete column name used in database queries.
// Implements the SoftDeleteColumnNamer interface.
func (s *DeletedAtMaxDate) SoftDeletedAtColumn() string {
	return constants.DeletedAtColumnName
}

// DeletedAtColumn returns the soft delete column name used in database queries.
//
// Deprecated: Use SoftDeletedAtColumn() instead.
func (s *DeletedAtMaxDate) DeletedAtColumn() string {
	return s.SoftDeletedAtColumn()
}

// IsSoftDeleted returns true if the model has been soft deleted.
// A record is soft-deleted when deleted_at <= NOW().
func (s *DeletedAtMaxDate) IsSoftDeleted() bool {
	return !s.DeletedAt.After(time.Now())
}

// IsDeleted returns true if the model has been soft deleted.
//
// Deprecated: Use IsSoftDeleted() instead.
func (s *DeletedAtMaxDate) IsDeleted() bool {
	return s.IsSoftDeleted()
}

// SoftDelete marks the model as soft-deleted by setting deleted_at to NOW().
func (s *DeletedAtMaxDate) SoftDelete() {
	s.DeletedAt = time.Now()
}

// Delete marks the model as soft-deleted by setting deleted_at to NOW().
//
// Deprecated: Use SoftDelete() instead.
func (s *DeletedAtMaxDate) Delete() {
	s.SoftDelete()
}

// RestoreSoftDeleted marks the model as not soft-deleted by setting deleted_at to MaxSoftDeletedAt.
func (s *DeletedAtMaxDate) RestoreSoftDeleted() {
	s.DeletedAt = MaxSoftDeletedAt
}

// Restore marks the model as not soft-deleted by setting deleted_at to MaxSoftDeletedAt.
//
// Deprecated: Use RestoreSoftDeleted() instead.
func (s *DeletedAtMaxDate) Restore() {
	s.RestoreSoftDeleted()
}

// GetSoftDeletedAt returns the deleted_at timestamp.
func (s *DeletedAtMaxDate) GetSoftDeletedAt() time.Time {
	return s.DeletedAt
}

// GetDeletedAt returns the deleted_at timestamp.
//
// Deprecated: Use GetSoftDeletedAt() instead.
func (s *DeletedAtMaxDate) GetDeletedAt() time.Time {
	return s.GetSoftDeletedAt()
}

// SoftDeleteValue returns the value to write on soft delete.
// Implements the SoftDeleteStrategy interface.
func (s *DeletedAtMaxDate) SoftDeleteValue() any {
	return time.Now()
}

// RestoreValue returns the value to write on restore.
// Implements the SoftDeleteStrategy interface.
func (s *DeletedAtMaxDate) RestoreValue() any {
	return MaxSoftDeletedAt
}

// SoftDeletedCondition returns the SQL fragment + args for the "only soft deleted" filter.
// Implements the SoftDeleteStrategy interface.
func (s *DeletedAtMaxDate) SoftDeletedCondition(quoteIdentifier func(string) string) (string, []any) {
	return quoteIdentifier(constants.DeletedAtColumnName) + " <= ?", []any{time.Now()}
}

// NotSoftDeletedCondition returns the SQL fragment + args for the "not soft deleted" filter.
// Implements the SoftDeleteStrategy interface.
func (s *DeletedAtMaxDate) NotSoftDeletedCondition(quoteIdentifier func(string) string) (string, []any) {
	return quoteIdentifier(constants.DeletedAtColumnName) + " > ?", []any{time.Now()}
}
