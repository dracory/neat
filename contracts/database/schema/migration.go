package schema

// MigrationInterface defines the contract for a single migration
type MigrationInterface interface {
	// Signature Get the migration signature.
	Signature() string
	// Up Run the migrations.
	Up() error
	// Down Reverse the migrations.
	Down() error
	// SetSchema sets the schema for this migration
	SetSchema(schema Schema)
	// GetSchema returns the schema for this migration
	GetSchema() Schema
}
