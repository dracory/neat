package orm

import (
	"database/sql"
	"time"
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

// SoftDeletes represents a soft delete trait.
type SoftDeletes struct {
	DeletedAt sql.NullTime `json:"deleted_at,omitempty"`
}

// DeletedAtColumn returns the soft delete column name used in database queries.
// Implements the SoftDeleteColumnNamer interface.
func (sd *SoftDeletes) DeletedAtColumn() string {
	return "deleted_at"
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
