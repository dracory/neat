package orm

type Factory interface {
	// Count sets the number of models that should be generated.
	Count(count int) Factory
	// Table sets the table name for database operations.
	Table(table string) Factory
	// Create creates a model and persists it to the database, returning the created instance(s).
	Create(value any, attributes ...map[string]any) (any, error)
	// CreateQuietly creates a model and persists it to the database without firing any model events, returning the created instance(s).
	CreateQuietly(value any, attributes ...map[string]any) (any, error)
	// Make creates a model and returns it, but does not persist it to the database.
	Make(value any, attributes ...map[string]any) (any, error)
}
