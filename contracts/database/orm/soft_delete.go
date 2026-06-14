package orm

// SoftDeleteColumnNamer is implemented by models that customize the soft delete column name.
// By default, neat uses "deleted_at" as the soft delete column. Implement this interface
// to override the column name for a given model.
//
// Example — using SoftDeletedAt (built-in soft_deleted_at):
//
//	type User struct {
//	    neat.SoftDeletedAt
//	    ID   uint
//	    Name string
//	}
//
// Example — custom column name on an existing model:
//
//	type User struct {
//	    neat.SoftDeletes
//	    ID   uint
//	    Name string
//	}
//
//	func (u *User) DeletedAtColumn() string {
//	    return "removed_at"
//	}
type SoftDeleteColumnNamer interface {
	// DeletedAtColumn returns the database column name used for soft deletes.
	DeletedAtColumn() string
}
