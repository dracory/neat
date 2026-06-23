package orm

import (
	"context"
	"database/sql"
)

// ArraySource is implemented by any "model" that wants array-backed storage.
type ArraySource interface {
	TableName() string
	Rows() ([]map[string]any, error)
}

// ArraySchema is an optional interface for empty-dataset or type-ambiguous cases.
type ArraySchema interface {
	Schema() map[string]string // column -> type ("string", "int", "float", "bool", "time")
}

// ArrayPopulator is implemented by drivers that support array-backed storage.
type ArrayPopulator interface {
	Populate(ctx context.Context, db *sql.DB, source ArraySource) error
}
