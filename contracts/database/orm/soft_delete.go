package orm

// SoftDeleteColumnNamer is implemented by models that customize the soft delete column name.
// By default, neat uses "soft_deleted_at" as the soft delete column. Implement this interface
// to override the column name for a given model.
//
// The three built-in embeds all satisfy this interface automatically:
//
//	soft_delete.SoftDeletes   — column "soft_deleted_at" (default)
//	soft_delete.SoftDeletedAt — column "soft_deleted_at" (explicit naming)
//	soft_delete.DeletedAt     — column "deleted_at"      (Laravel-compatible)
//
// Example — fully custom column name:
//
//	type Order struct {
//	    soft_delete.SoftDeletes
//	    ID uint
//	}
//
//	func (o *Order) SoftDeletedAtColumn() string {
//	    return "removed_at"
//	}
type SoftDeleteColumnNamer interface {
	// SoftDeletedAtColumn returns the database column name used for soft deletes.
	SoftDeletedAtColumn() string
}
