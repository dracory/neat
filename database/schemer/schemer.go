// Package schemer is a deprecated compatibility shim for database/migrator.
//
// Deprecated: Use github.com/dracory/neat/database/migrator instead.
// Replace schemer.NewSchemer with migrator.NewMigrator.
package schemer

import (
	"github.com/dracory/neat/database"
	"github.com/dracory/neat/database/migrator"
)

// SchemerInterface is an alias for migrator.MigratorInterface.
//
// Deprecated: Use migrator.MigratorInterface.
type SchemerInterface = migrator.MigratorInterface

// SchemerImplementation is an alias for migrator.Migrator.
//
// Deprecated: Use migrator.Migrator.
type SchemerImplementation = migrator.Migrator

// NewSchemer is an alias for migrator.NewMigrator.
//
// Deprecated: Use migrator.NewMigrator.
func NewSchemer(db *database.Database) SchemerInterface {
	return migrator.NewMigrator(db)
}
