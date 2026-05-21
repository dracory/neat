package association

import (
	"fmt"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
)

// Association represents a model association.
type Association struct {
	query       contractsorm.Query
	model       any
	association string
}

// NewAssociation creates a new Association instance.
func NewAssociation(query contractsorm.Query, model any, association string) *Association {
	return &Association{
		query:       query,
		model:       model,
		association: association,
	}
}

// Find finds records that match given conditions.
func (a *Association) Find(out any, conds ...any) error {
	// This is a base implementation - specific relationship types
	// (belongs-to, has-many, has-one) will override this
	return fmt.Errorf("association type not specified, use specific association type")
}

// Append appending a model to the association.
func (a *Association) Append(values ...any) error {
	return fmt.Errorf("association type not specified, use specific association type")
}

// Replace replaces the association with the given value.
func (a *Association) Replace(values ...any) error {
	return fmt.Errorf("association type not specified, use specific association type")
}

// Delete deletes the given value from the association.
func (a *Association) Delete(values ...any) error {
	return fmt.Errorf("association type not specified, use specific association type")
}

// Clear clears the association.
func (a *Association) Clear() error {
	return fmt.Errorf("association type not specified, use specific association type")
}

// Count returns the number of records in the association.
func (a *Association) Count() int64 {
	return 0
}

// Query returns the underlying query instance.
func (a *Association) Query() contractsorm.Query {
	return a.query
}

// Model returns the model instance.
func (a *Association) Model() any {
	return a.model
}

// AssociationName returns the association name.
func (a *Association) AssociationName() string {
	return a.association
}
