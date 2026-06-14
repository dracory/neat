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

// SoftDeleteStrategy is implemented by models that use a non-NULL soft delete strategy.
// Models that do NOT implement this interface are assumed to use the NULL strategy.
//
// The SoftDeleteStrategy interface extends SoftDeleteColumnNamer to provide
// complete control over soft delete behavior, including the values used for
// soft delete and restore operations, as well as the SQL conditions for
// filtering active and deleted records.
//
// This is used by the max-date sentinel strategy (SoftDeletesMaxDate) where
// records are soft-deleted when their timestamp is in the past (<= current time),
// and active when it is in the future (e.g., 9999-12-31 23:59:59).
//
// Example:
//
//	func (s *MyModel) SoftDeleteValue() any { return time.Now() }
//	func (s *MyModel) RestoreValue() any     { return MaxSoftDeletedAt }
//	func (s *MyModel) IsDeletedCondition(q func(string) string) (string, []any) {
//	    return q("soft_deleted_at") + " <= ?", []any{time.Now()}
//	}
//	func (s *MyModel) IsActiveCondition(q func(string) string) (string, []any) {
//	    return q("soft_deleted_at") + " > ?", []any{time.Now()}
//	}
type SoftDeleteStrategy interface {
	SoftDeleteColumnNamer
	// SoftDeleteValue returns the value to write on soft delete (e.g. time.Now()).
	SoftDeleteValue() any
	// RestoreValue returns the value to write on restore.
	RestoreValue() any
	// IsDeletedCondition returns the SQL fragment + args for the "only deleted" filter.
	IsDeletedCondition(quoteIdentifier func(string) string) (string, []any)
	// IsActiveCondition returns the SQL fragment + args for the "not deleted" filter.
	IsActiveCondition(quoteIdentifier func(string) string) (string, []any)
}
