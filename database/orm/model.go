package orm

import (
	"database/sql"
	"time"

	"github.com/dracory/neat/database/schema/constants"
)

// Model represents a base model with ID and timestamps.
type Model struct {
	ID uint `json:"id"`
	Timestamps
}

// ShortID provides a short string primary key field.
// Embed it in your model to opt into client-generated short IDs.
// No timestamps, no soft deletes — just the ID.
type ShortID struct {
	ID string `json:"id" db:"id"`
}

// SoftDeletes represents a soft delete trait using the "soft_deleted_at" column.
// This is the default — use DeletedAt for Laravel-compatible "deleted_at" column.
type SoftDeletes struct {
	SoftDeletedAt sql.NullTime `json:"soft_deleted_at,omitempty" db:"soft_deleted_at"`
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

// DeletedAt represents a soft delete trait using the "deleted_at" column (Laravel-compatible).
// Use this when your schema follows the Laravel Eloquent convention.
type DeletedAt struct {
	DeletedAt sql.NullTime `json:"deleted_at,omitempty" db:"deleted_at"`
}

// SoftDeletedAtColumn returns the soft delete column name used in database queries.
// Implements the SoftDeleteColumnNamer interface.
func (sd *DeletedAt) SoftDeletedAtColumn() string {
	return constants.DeletedAtColumnName
}

// DeletedAtColumn returns the soft delete column name used in database queries.
//
// Deprecated: Use SoftDeletedAtColumn() instead.
func (sd *DeletedAt) DeletedAtColumn() string {
	return sd.SoftDeletedAtColumn()
}

// CreatedAt provides only the created timestamp for immutable models.
// Embed this when you need only created_at without updated_at (e.g., audit logs, event sourcing).
type CreatedAt struct {
	CreatedAt time.Time `json:"created_at"`
}

// UpdatedAt provides only the updated timestamp.
// Embed this when you need only updated_at without created_at.
type UpdatedAt struct {
	UpdatedAt time.Time `json:"updated_at"`
}

// Timestamps represents both created and updated timestamp fields.
// This is a convenience embed combining CreatedAt and UpdatedAt.
type Timestamps struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
