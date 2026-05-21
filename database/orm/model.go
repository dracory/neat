package orm

import (
	"database/sql"
	"time"
)

// Model represents a base model with ID and timestamps.
type Model struct {
	ID        uint       `json:"id"`
	Timestamps
}

// SoftDeletes represents a soft delete trait.
type SoftDeletes struct {
	DeletedAt sql.NullTime `json:"deleted_at,omitempty"`
}

// Timestamps represents created and updated timestamp fields.
type Timestamps struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
