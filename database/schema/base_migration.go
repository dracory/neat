package schema

import (
	"errors"

	contractsschema "github.com/dracory/neat/contracts/database/schema"
)

// BaseMigration provides common functionality for all migrations
type BaseMigration struct {
	schema contractsschema.Schema
}

// SetSchema sets the schema for this migration
func (b *BaseMigration) SetSchema(schema contractsschema.Schema) {
	b.schema = schema
}

// GetSchema returns the schema for this migration
func (b *BaseMigration) GetSchema() contractsschema.Schema {
	return b.schema
}

func (b *BaseMigration) Signature() string {
	return ""
}

func (b *BaseMigration) Description() string {
	return ""
}

func (b *BaseMigration) Up() error {
	return errors.New("up method not implemented")
}

func (b *BaseMigration) Down() error {
	return errors.New("down method not implemented")
}
