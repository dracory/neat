package orm

import (
	"fmt"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
)

// Factory implements the Factory interface for creating test data.
type Factory struct {
	orm  *Orm
	count int
}

// NewFactory creates a new Factory instance.
func NewFactory(orm *Orm) *Factory {
	return &Factory{
		orm:  orm,
		count: 1,
	}
}

// Count sets the number of models that should be generated.
func (f *Factory) Count(count int) contractsorm.Factory {
	f.count = count
	return f
}

// Create creates a model and persists it to the database.
func (f *Factory) Create(value any, attributes ...map[string]any) error {
	// TODO: Implement factory creation with attributes
	return fmt.Errorf("not implemented")
}

// CreateQuietly creates a model and persists it to the database without firing any model events.
func (f *Factory) CreateQuietly(value any, attributes ...map[string]any) error {
	// TODO: Implement quiet factory creation
	return fmt.Errorf("not implemented")
}

// Make creates a model and returns it, but does not persist it to the database.
func (f *Factory) Make(value any, attributes ...map[string]any) error {
	// TODO: Implement factory make
	return fmt.Errorf("not implemented")
}
