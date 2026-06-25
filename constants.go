package neat

import "github.com/dracory/neat/contracts/database/orm"

// This file re-exports project-wide constants so that consumers can reference
// them from the root neat package without importing an internal sub-package.
//
// Example:
//
//	import "github.com/dracory/neat"
//
//	db.Table("users").OrderBy("created_at", neat.SortDesc)
//	blueprint.Timestamp("soft_deleted_at").Default(neat.MaxDateTime)

// Sort directions accepted by orm.Query.OrderBy.
// These are aliases for orm.SortAsc and orm.SortDesc — the canonical source
// of truth lives in contracts/database/orm.
const (
	SortAsc  = orm.SortAsc
	SortDesc = orm.SortDesc
)

// Sentinel date/time values for use as column defaults, sentinel values,
// and soft-delete strategies.
//
// NullDate / NullDateTime represent the earliest valid date in the Gregorian
// calendar (1 AD — there is no year 0). Use these as NOT NULL sentinels for
// "no value" instead of NULL.
//
// MaxDate / MaxDateTime represent the latest representable date/time.
// Use these as NOT NULL sentinels for "not deleted" in max-date soft-delete
// strategies.
const (
	NullDate     = "0002-01-01"
	NullDateTime = "0002-01-01 00:00:00"
	MaxDate      = "9999-12-31"
	MaxDateTime  = "9999-12-31 23:59:59"
)

// Common string constants for yes/no values.
const (
	Yes = "yes"
	No  = "no"
)
