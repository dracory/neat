package association

import (
	"fmt"
	"regexp"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
)

// Association represents a model association.
// This is the base type for all relationship types (belongs-to, has-many, has-one, polymorphic).
// It provides common functionality for managing relationships between models.
//
// Example:
//
//	type User struct {
//	    ID       uint
//	    Name     string
//	    Profile  *Profile
//	    Posts    []Post
//	}
//
//	// The Profile and Posts fields would be managed through Association instances
type Association struct {
	query       contractsorm.Query // The query builder for database operations
	model       any                // The model instance this association belongs to
	association string             // The name of the association (e.g., "profile", "posts")
}

// NewAssociation creates a new Association instance.
// The query parameter provides the query builder for database operations.
// The model parameter is the model instance this association belongs to.
// The association parameter is the name of the association (e.g., "profile", "posts").
//
// Example:
//
//	user := User{ID: 1, Name: "John"}
//	assoc := NewAssociation(db.Query(), &user, "profile")
func NewAssociation(query contractsorm.Query, model any, association string) *Association {
	return &Association{
		query:       query,
		model:       model,
		association: association,
	}
}

// Find finds records that match given conditions.
// The out parameter must be a pointer to a struct or slice for the results.
// The conds parameter provides optional WHERE conditions for the query.
// This is a base implementation - specific relationship types (belongs-to, has-many, has-one) override this.
//
// Example:
//
//	var profile Profile
//	err := assoc.Find(&profile)
//
//	var posts []Post
//	err := assoc.Find(&posts, "published = ?", true)
func (a *Association) Find(out any, conds ...any) error {
	// This is a base implementation - specific relationship types
	// (belongs-to, has-many, has-one) will override this
	return fmt.Errorf("association type not specified, use specific association type")
}

// Append appends a model to the association.
// The values parameter provides the model(s) to append to the association.
// This is a base implementation - specific relationship types override this.
//
// Example:
//
//	err := assoc.Append(&Post{Title: "New Post"})
func (a *Association) Append(values ...any) error {
	return fmt.Errorf("association type not specified, use specific association type")
}

// Replace replaces the association with the given value.
// The values parameter provides the new model(s) for the association.
// This is a base implementation - specific relationship types override this.
//
// Example:
//
//	err := assoc.Replace(&Post{Title: "Updated Post"})
func (a *Association) Replace(values ...any) error {
	return fmt.Errorf("association type not specified, use specific association type")
}

// Delete deletes the given value from the association.
// The values parameter provides the model(s) to remove from the association.
// This is a base implementation - specific relationship types override this.
//
// Example:
//
//	err := assoc.Delete(&Post{ID: 1})
func (a *Association) Delete(values ...any) error {
	return fmt.Errorf("association type not specified, use specific association type")
}

// Clear clears the association by removing all related models.
// This is a base implementation - specific relationship types override this.
//
// Example:
//
//	err := assoc.Clear()
func (a *Association) Clear() error {
	return fmt.Errorf("association type not specified, use specific association type")
}

// Count returns the number of records in the association.
// This is a base implementation - specific relationship types override this.
//
// Example:
//
//	count := assoc.Count()
func (a *Association) Count() int64 {
	return 0
}

// Query returns the underlying query instance.
// This allows direct access to the query builder for custom operations.
//
// Example:
//
//	query := assoc.Query()
//	err := query.Where("status = ?", "active").Get(&results)
func (a *Association) Query() contractsorm.Query {
	return a.query
}

// Model returns the model instance this association belongs to.
//
// Example:
//
//	user := assoc.Model().(*User)
func (a *Association) Model() any {
	return a.model
}

// AssociationName returns the association name.
//
// Example:
//
//	name := assoc.AssociationName() // "profile", "posts", etc.
func (a *Association) AssociationName() string {
	return a.association
}

// isValidIdentifier validates that a string is a valid SQL identifier.
// This prevents SQL injection when concatenating identifiers into SQL queries.
// SQL identifiers must start with a letter or underscore, followed by letters, digits, or underscores.
//
// Example:
//
//	isValidIdentifier("user_id") // true
//	isValidIdentifier("123_invalid") // false
//	isValidIdentifier("invalid; DROP TABLE") // false
func isValidIdentifier(s string) bool {
	// SQL identifiers must start with a letter or underscore, followed by letters, digits, or underscores
	// This regex matches common SQL identifier patterns
	// Use pre-compiled regex for efficiency and to avoid error handling
	var identifierRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	return identifierRegex.MatchString(s)
}
