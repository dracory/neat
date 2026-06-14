package constants

const (
	CommandAdd          = "add"
	CommandChange       = "change"
	CommandComment      = "comment"
	CommandCreate       = "create"
	CommandDrop         = "drop"
	CommandDropColumn   = "dropColumn"
	CommandDropForeign  = "dropForeign"
	CommandDropFullText = "dropFullText"
	CommandDropIfExists = "dropIfExists"
	CommandDropIndex    = "dropIndex"
	CommandDropPrimary  = "dropPrimary"
	CommandDropUnique   = "dropUnique"
	CommandForeign      = "foreign"
	CommandFullText     = "fullText"
	CommandIndex        = "index"
	CommandPrimary      = "primary"
	CommandRename       = "rename"
	CommandRenameColumn = "renameColumn"
	CommandRenameIndex  = "renameIndex"
	CommandUnique       = "unique"
	DefaultStringLength = 255

	// MaxSoftDeletedAtDefault is the default value string for max-date sentinel soft delete columns.
	// Use this with .Default() when creating soft delete columns with the max-date strategy.
	MaxSoftDeletedAtDefault = "9999-12-31 23:59:59"

	// DefaultColumnNames are the default column names used by schema builder helpers.
	DefaultIDColumn            = "id"
	DefaultCreatedAtColumn     = "created_at"
	DefaultUpdatedAtColumn     = "updated_at"
	DefaultSoftDeletedAtColumn = "soft_deleted_at"

	// SoftDeleteColumnNames are the column names used for soft delete functionality.
	// SoftDeleteAtColumn is the default column name used for soft deletes.
	// The default is "soft_deleted_at" — semantically explicit about what the column tracks.
	// Use DeletedAtColumnName for "deleted_at" (Laravel-style) compatibility.
	SoftDeleteAtColumn = DefaultSoftDeletedAtColumn

	// DeletedAtColumnName is the column name used by the DeletedAt embed (Laravel-compatible).
	DeletedAtColumnName = "deleted_at"

	// DeletedAtColumn is an alias for SoftDeleteAtColumn.
	//
	// Deprecated: Use SoftDeleteAtColumn instead.
	DeletedAtColumn = SoftDeleteAtColumn
)
