package schema

type ColumnDefinition interface {
	// After set the column as after another column (Mysql only)
	After(column string) ColumnDefinition
	// AutoIncrement set the column as auto increment
	AutoIncrement() ColumnDefinition
	// Change mark the column as a change operation
	Change() ColumnDefinition
	// Collation set the column collation
	Collation(collation string) ColumnDefinition
	// Comment sets the comment value
	Comment(comment string) ColumnDefinition
	// Default set the default value
	Default(def any) ColumnDefinition
	// First set the column as first (Mysql only)
	First() ColumnDefinition
	// GetAfter returns the after value
	GetAfter() string
	// GetAllowed returns the allowed value
	GetAllowed() []any
	// GetAutoIncrement returns the autoIncrement value
	GetAutoIncrement() bool
	// GetChange returns the change value
	GetChange() bool
	// GetCollation returns the collation value
	GetCollation() string
	// GetComment returns the comment value
	GetComment() (comment string)
	// GetDefault returns the default value
	GetDefault() any
	// GetFirst returns the first value
	GetFirst() bool
	// GetLength returns the length value
	GetLength() int
	// GetName returns the name value
	GetName() string
	// GetNullable returns the nullable value
	GetNullable() bool
	// GetOnUpdate returns the onUpdate value
	GetOnUpdate() any
	// GetPlaces returns the places value
	GetPlaces() int
	// GetPrecision returns the precision value
	GetPrecision() int
	// GetSrid returns the SRID value
	GetSrid() int
	// GetTotal returns the total value
	GetTotal() int
	// GetType returns the type value
	GetType() string
	// GetUnsigned returns the unsigned value
	GetUnsigned() bool
	// GetUseCurrent returns the useCurrent value
	GetUseCurrent() bool
	// GetUseCurrentOnUpdate returns the useCurrentOnUpdate value
	GetUseCurrentOnUpdate() bool
	// IsSetComment returns true if the comment value is set
	IsSetComment() bool
	// OnUpdate sets the column to use the value on update (Mysql only)
	OnUpdate(value any) ColumnDefinition
	// Srid set the SRID (Spatial only)
	Srid(srid int) ColumnDefinition
	// Places set the decimal places
	Places(places int) ColumnDefinition
	// Total set the decimal total
	Total(total int) ColumnDefinition
	// Nullable allow NULL values to be inserted into the column
	Nullable() ColumnDefinition
	// Unsigned set the column as unsigned
	Unsigned() ColumnDefinition
	// UseCurrent set the column to use the current timestamp
	UseCurrent() ColumnDefinition
	// UseCurrentOnUpdate set the column to use the current timestamp on update (Mysql only)
	UseCurrentOnUpdate() ColumnDefinition
}

type Column struct {
	Autoincrement bool
	Collation     string
	Comment       string
	Default       string
	Name          string
	Nullable      bool
	Type          string
	TypeName      string
}
