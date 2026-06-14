package schema

import (
	"fmt"
	"strings"

	ormcontract "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/database/schema/constants"
	"github.com/dracory/neat/support/convert"
)

// Blueprint represents a database table schema definition for migrations.
// It provides a fluent interface for defining table structure, including columns,
// indexes, foreign keys, and other schema elements. The blueprint is used by
// schema builders to generate database-specific SQL statements.
type Blueprint struct {
	columns         []*ColumnDefinition
	commands        []*schema.Command
	prefix          string
	schema          schema.Schema
	skipTransaction bool
	table           string
}

// NewBlueprint creates a new Blueprint instance for the specified schema, prefix, and table.
// The schema parameter provides database connection and grammar capabilities.
// The prefix is used to prefix table names (useful for multi-tenancy or namespacing).
// The table parameter specifies the name of the table to create or modify.
func NewBlueprint(schema schema.Schema, prefix, table string) *Blueprint {
	return &Blueprint{
		prefix: prefix,
		schema: schema,
		table:  table,
	}
}

// BigIncrements creates a new auto-incrementing big integer (bigint) column on the table.
// This is typically used as a primary key for tables that may have many records.
// The column will be unsigned and set to auto-increment.
func (r *Blueprint) BigIncrements(column string) schema.ColumnDefinition {
	return r.UnsignedBigInteger(column).AutoIncrement()
}

// BigInteger creates a new big integer (bigint) column on the table.
// Big integers can store values from -9,223,372,036,854,775,808 to 9,223,372,036,854,775,807.
func (r *Blueprint) BigInteger(column string) schema.ColumnDefinition {
	return r.createAndAddColumn("bigInteger", column)
}

// Boolean creates a new boolean column on the table.
// Boolean columns typically store true/false values, often represented as 0/1 in the database.
func (r *Blueprint) Boolean(column string) schema.ColumnDefinition {
	return r.createAndAddColumn("boolean", column)
}

// Change modifies the most recently added column definition.
// This is used in table modification operations to alter an existing column's properties.
// Returns nil if no columns have been added to the blueprint.
func (r *Blueprint) Change() schema.ColumnDefinition {
	result := r.modifyColumn()
	if result == nil {
		return nil
	}
	return result
}

// Build executes the schema blueprint against the database using the provided query and grammar.
// It converts all commands into SQL statements and executes them in sequence.
// Returns an error if any SQL statement fails to execute.
func (r *Blueprint) Build(query ormcontract.Query, grammar schema.Grammar) error {
	statements, err := r.ToSql(grammar)
	if err != nil {
		return err
	}
	for _, sql := range statements {
		if sql == "" {
			continue
		}
		if _, err := query.Exec(sql); err != nil {
			return err
		}
	}

	return nil
}

// Char creates a new fixed-length character column on the table.
// The length parameter is optional; if not provided, it defaults to the DefaultStringLength constant.
// Char columns are faster than varchar but use fixed storage regardless of actual content length.
func (r *Blueprint) Char(column string, length ...int) schema.ColumnDefinition {
	defaultLength := constants.DefaultStringLength
	if len(length) > 0 {
		defaultLength = length[0]
	}

	columnImpl := r.createAndAddColumn("char", column)
	columnImpl.length = &defaultLength

	return columnImpl
}

// Column creates a new column of the specified type on the table.
// This is a generic method that allows adding any column type by specifying its type string.
// The ttype parameter should match a type recognized by the database grammar.
func (r *Blueprint) Column(column, ttype string) schema.ColumnDefinition {
	return r.createAndAddColumn(ttype, column)
}

// Create indicates that the blueprint should create a new table.
// This command is used when defining a new table structure in a migration.
func (r *Blueprint) Create() {
	r.addCommand(&schema.Command{
		Name: constants.CommandCreate,
	})
}

// Decimal creates a new decimal column on the table.
// Decimal columns are used for precise numeric calculations, such as monetary values.
// Use the methods on the returned ColumnDefinition to set precision and scale.
func (r *Blueprint) Decimal(column string) schema.ColumnDefinition {
	return r.createAndAddColumn("decimal", column)
}

// Date creates a new date column on the table.
// Date columns store calendar dates (year, month, day) without time components.
func (r *Blueprint) Date(column string) schema.ColumnDefinition {
	return r.createAndAddColumn("date", column)
}

// DateTime creates a new datetime column on the table.
// The precision parameter is optional and specifies fractional seconds precision (0-6).
// DateTime columns store both date and time information.
func (r *Blueprint) DateTime(column string, precision ...int) schema.ColumnDefinition {
	columnImpl := r.createAndAddColumn("dateTime", column)
	if len(precision) > 0 {
		columnImpl.precision = &precision[0]
	}

	return columnImpl
}

// DateTimeTz creates a new datetime with timezone column on the table.
// The precision parameter is optional and specifies fractional seconds precision (0-6).
// DateTimeTz columns store date, time, and timezone information.
func (r *Blueprint) DateTimeTz(column string, precision ...int) schema.ColumnDefinition {
	columnImpl := r.createAndAddColumn("dateTimeTz", column)
	if len(precision) > 0 {
		columnImpl.precision = &precision[0]
	}

	return columnImpl
}

// Double creates a new double-precision floating-point column on the table.
// Double columns provide high precision for floating-point numbers.
func (r *Blueprint) Double(column string) schema.ColumnDefinition {
	return r.createAndAddColumn("double", column)
}

// Drop indicates that the blueprint should drop the table.
// This command permanently removes the table and all its data from the database.
func (r *Blueprint) Drop() {
	r.addCommand(&schema.Command{
		Name: constants.CommandDrop,
	})
}

// DropColumn removes one or more columns from the table.
// Multiple column names can be specified to drop several columns at once.
func (r *Blueprint) DropColumn(column ...string) {
	r.addCommand(&schema.Command{
		Name:    constants.CommandDropColumn,
		Columns: column,
	})
}

// DropForeign removes a foreign key constraint from the table.
// The column parameter specifies the column(s) that the foreign key constraint is based on.
// The foreign key name is automatically generated if not explicitly set.
func (r *Blueprint) DropForeign(column ...string) {
	r.indexCommand(constants.CommandDropForeign, column, schema.IndexConfig{
		Name: r.createIndexName(constants.CommandForeign, column),
	})
}

// DropForeignByName removes a foreign key constraint by its name.
// This is useful when you need to drop a foreign key with a specific name.
func (r *Blueprint) DropForeignByName(name string) {
	r.indexCommand(constants.CommandDropForeign, nil, schema.IndexConfig{
		Name: name,
	})
}

// DropFullText removes a full-text index from the table.
// The column parameter specifies the column(s) that the full-text index is based on.
// The index name is automatically generated if not explicitly set.
func (r *Blueprint) DropFullText(column ...string) {
	r.indexCommand(constants.CommandDropFullText, column, schema.IndexConfig{
		Name: r.createIndexName(constants.CommandFullText, column),
	})
}

// DropFullTextByName removes a full-text index by its name.
// This is useful when you need to drop a full-text index with a specific name.
func (r *Blueprint) DropFullTextByName(name string) {
	r.indexCommand(constants.CommandDropFullText, nil, schema.IndexConfig{
		Name: name,
	})
}

// DropIfExists indicates that the blueprint should drop the table if it exists.
// This command is safer than Drop as it won't fail if the table doesn't exist.
func (r *Blueprint) DropIfExists() {
	r.addCommand(&schema.Command{
		Name: constants.CommandDropIfExists,
	})
}

// DropIndex removes a regular index from the table.
// The column parameter specifies the column(s) that the index is based on.
// The index name is automatically generated if not explicitly set.
func (r *Blueprint) DropIndex(column ...string) {
	r.indexCommand(constants.CommandDropIndex, column, schema.IndexConfig{
		Name: r.createIndexName(constants.CommandIndex, column),
	})
}

// DropIndexByName removes an index by its name.
// This is useful when you need to drop an index with a specific name.
func (r *Blueprint) DropIndexByName(name string) {
	r.indexCommand(constants.CommandDropIndex, nil, schema.IndexConfig{
		Name: name,
	})
}

// DropPrimary removes the primary key constraint from the table.
// The column parameter specifies the column(s) that make up the primary key.
func (r *Blueprint) DropPrimary(column ...string) {
	r.indexCommand(constants.CommandDropPrimary, column, schema.IndexConfig{
		Name: r.createIndexName(constants.CommandPrimary, column),
	})
}

// DropSoftDeletes removes the soft delete column from the table.
// If no column name is specified, it defaults to "soft_deleted_at".
// This is used to remove soft delete functionality from a table.
func (r *Blueprint) DropSoftDeletes(column ...string) {
	if len(column) > 0 {
		r.DropColumn(column[0])
	} else {
		r.DropColumn(constants.SoftDeleteAtColumn)
	}
}

// DropSoftDeletesTz removes the soft delete with timezone column from the table.
// This is an alias for DropSoftDeletes as timezone handling is database-specific.
func (r *Blueprint) DropSoftDeletesTz(column ...string) {
	r.DropSoftDeletes(column...)
}

// DropTimestamps removes the created_at and updated_at timestamp columns from the table.
// This is used to remove automatic timestamp tracking from a table.
func (r *Blueprint) DropTimestamps() {
	r.DropColumn(constants.DefaultCreatedAtColumn, constants.DefaultUpdatedAtColumn)
}

// DropTimestampsTz removes the timestamp with timezone columns from the table.
// This is an alias for DropTimestamps as timezone handling is database-specific.
func (r *Blueprint) DropTimestampsTz() {
	r.DropTimestamps()
}

// DropUnique removes a unique index from the table.
// The column parameter specifies the column(s) that the unique index is based on.
// The index name is automatically generated if not explicitly set.
func (r *Blueprint) DropUnique(column ...string) {
	r.indexCommand(constants.CommandDropUnique, column, schema.IndexConfig{
		Name: r.createIndexName(constants.CommandUnique, column),
	})
}

// DropUniqueByName removes a unique index by its name.
// This is useful when you need to drop a unique index with a specific name.
func (r *Blueprint) DropUniqueByName(name string) {
	r.indexCommand(constants.CommandDropUnique, nil, schema.IndexConfig{
		Name: name,
	})
}

// Enum creates a new enum column on the table.
// The allowed parameter specifies the valid values for the enum.
// Enum columns restrict values to a predefined set of options.
func (r *Blueprint) Enum(column string, allowed []any) schema.ColumnDefinition {
	columnImpl := r.createAndAddColumn("enum", column)
	columnImpl.allowed = allowed

	return columnImpl
}

// Float creates a new float column on the table.
// The precision parameter is optional; if not provided, it defaults to 53 bits.
// Float columns are used for single-precision floating-point numbers.
func (r *Blueprint) Float(column string, precision ...int) schema.ColumnDefinition {
	columnImpl := r.createAndAddColumn("float", column)
	columnImpl.precision = convert.Pointer(53)

	if len(precision) > 0 {
		columnImpl.precision = &precision[0]
	}

	return columnImpl
}

// Foreign creates a new foreign key constraint on the table.
// The column parameter specifies the column(s) that reference another table.
// Returns a ForeignKeyDefinition that can be used to configure the constraint further.
func (r *Blueprint) Foreign(column ...string) schema.ForeignKeyDefinition {
	command := r.indexCommand(constants.CommandForeign, column)

	return NewForeignKeyDefinition(command)
}

// FullText creates a new full-text index on the table.
// The column parameter specifies the column(s) to include in the full-text index.
// Full-text indexes enable efficient text searching capabilities.
func (r *Blueprint) FullText(column ...string) schema.IndexDefinition {
	command := r.indexCommand(constants.CommandFullText, column)

	return NewIndexDefinition(command)
}

// Geometry creates a new geometry column on the table.
// Geometry columns store spatial data such as points, lines, and polygons.
func (r *Blueprint) Geometry(column string) schema.ColumnDefinition {
	return r.createAndAddColumn("geometry", column)
}

// GeometryCollection creates a new geometry collection column on the table.
// Geometry collections can store multiple geometry objects in a single column.
func (r *Blueprint) GeometryCollection(column string) schema.ColumnDefinition {
	return r.createAndAddColumn("geometryCollection", column)
}

// GetAddedColumns returns all column definitions that have been added to this blueprint.
// This is useful for introspection or validation of the schema definition.
func (r *Blueprint) GetAddedColumns() []schema.ColumnDefinition {
	var columns []schema.ColumnDefinition
	for _, column := range r.columns {
		columns = append(columns, column)
	}

	return columns
}

// GetCommands returns all schema commands that have been added to this blueprint.
// This includes create, drop, alter, and other schema operations.
func (r *Blueprint) GetCommands() []*schema.Command {
	return r.commands
}

// GetTableName returns the name of the table for this blueprint.
func (r *Blueprint) GetTableName() string {
	return r.table
}

// HasCommand checks if a command of the specified type exists in this blueprint.
// This is useful for conditional logic based on the blueprint's state.
func (r *Blueprint) HasCommand(command string) bool {
	for _, c := range r.commands {
		if c.Name == command {
			return true
		}
	}

	return false
}

// ID creates a new auto-incrementing big integer primary key column on the table.
// If no column name is specified, it defaults to "id".
// This is a convenience method for creating a standard primary key.
func (r *Blueprint) ID(column ...string) schema.ColumnDefinition {
	if len(column) > 0 {
		return r.BigIncrements(column[0])
	}

	return r.BigIncrements("id")
}

// Increments creates a new auto-incrementing integer column on the table.
// This is an alias for IntegerIncrements.
func (r *Blueprint) Increments(column string) schema.ColumnDefinition {
	return r.IntegerIncrements(column)
}

// Index creates a new regular index on the table.
// The column parameter specifies the column(s) to include in the index.
// Returns an IndexDefinition that can be used to configure the index further.
func (r *Blueprint) Index(column ...string) schema.IndexDefinition {
	command := r.indexCommand(constants.CommandIndex, column)

	return NewIndexDefinition(command)
}

// Integer creates a new integer column on the table.
// Integer columns can store values from -2,147,483,648 to 2,147,483,647.
func (r *Blueprint) Integer(column string) schema.ColumnDefinition {
	return r.createAndAddColumn("integer", column)
}

// IntegerIncrements creates a new auto-incrementing integer column on the table.
// The column will be unsigned and set to auto-increment.
func (r *Blueprint) IntegerIncrements(column string) schema.ColumnDefinition {
	return r.UnsignedInteger(column).AutoIncrement()
}

// Json creates a new JSON column on the table.
// JSON columns store JSON data and provide JSON validation and querying capabilities.
func (r *Blueprint) Json(column string) schema.ColumnDefinition {
	return r.createAndAddColumn("json", column)
}

// Jsonb creates a new JSONB column on the table.
// JSONB is a binary JSON format that provides better performance for querying and indexing.
// This is primarily used in PostgreSQL databases.
func (r *Blueprint) Jsonb(column string) schema.ColumnDefinition {
	return r.createAndAddColumn("jsonb", column)
}

// LineString creates a new linestring column on the table.
// Linestrings store a sequence of points forming a line.
func (r *Blueprint) LineString(column string) schema.ColumnDefinition {
	return r.createAndAddColumn("lineString", column)
}

// LongText creates a new long text column on the table.
// Long text columns can store large amounts of text data (up to 4GB in some databases).
func (r *Blueprint) LongText(column string) schema.ColumnDefinition {
	return r.createAndAddColumn("longText", column)
}

// MediumIncrements creates a new auto-incrementing medium integer column on the table.
// The column will be unsigned and set to auto-increment.
func (r *Blueprint) MediumIncrements(column string) schema.ColumnDefinition {
	return r.UnsignedMediumInteger(column).AutoIncrement()
}

// MediumInteger creates a new medium integer column on the table.
// Medium integers can store values from -8,388,608 to 8,388,607.
func (r *Blueprint) MediumInteger(column string) schema.ColumnDefinition {
	return r.createAndAddColumn("mediumInteger", column)
}

// MediumText creates a new medium text column on the table.
// Medium text columns can store moderate amounts of text data (up to 16MB in some databases).
func (r *Blueprint) MediumText(column string) schema.ColumnDefinition {
	return r.createAndAddColumn("mediumText", column)
}

// MultiLineString creates a new multilinestring column on the table.
// Multilinestrings store a collection of linestrings.
func (r *Blueprint) MultiLineString(column string) schema.ColumnDefinition {
	return r.createAndAddColumn("multiLineString", column)
}

// MultiPoint creates a new multipoint column on the table.
// Multipoints store a collection of points.
func (r *Blueprint) MultiPoint(column string) schema.ColumnDefinition {
	return r.createAndAddColumn("multiPoint", column)
}

// MultiPolygon creates a new multipolygon column on the table.
// Multipolygons store a collection of polygons.
func (r *Blueprint) MultiPolygon(column string) schema.ColumnDefinition {
	return r.createAndAddColumn("multiPolygon", column)
}

// Point creates a new point column on the table.
// Points store a single coordinate (x, y) in 2D space.
func (r *Blueprint) Point(column string) schema.ColumnDefinition {
	return r.createAndAddColumn("point", column)
}

// Polygon creates a new polygon column on the table.
// Polygons store a closed shape defined by a sequence of points.
func (r *Blueprint) Polygon(column string) schema.ColumnDefinition {
	return r.createAndAddColumn("polygon", column)
}

// Primary creates a new primary key constraint on the table.
// The column parameter specifies the column(s) that make up the primary key.
// If no columns are specified, the most recently added column will be used.
func (r *Blueprint) Primary(column ...string) {
	r.indexCommand(constants.CommandPrimary, column)
}

// Rename indicates that the table should be renamed to the specified name.
// The to parameter specifies the new table name.
func (r *Blueprint) Rename(to string) {
	command := &schema.Command{
		Name: constants.CommandRename,
		To:   to,
	}

	r.addCommand(command)
}

// RenameColumn renames a column from one name to another.
// The from parameter specifies the current column name.
// The to parameter specifies the new column name.
func (r *Blueprint) RenameColumn(from, to string) {
	command := &schema.Command{
		Name: constants.CommandRenameColumn,
		From: from,
		To:   to,
	}

	r.addCommand(command)
}

// RenameIndex renames an index from one name to another.
// The from parameter specifies the current index name.
// The to parameter specifies the new index name.
// This operation skips transaction wrapping due to SQLite limitations.
func (r *Blueprint) RenameIndex(from, to string) {
	command := &schema.Command{
		Name: constants.CommandRenameIndex,
		From: from,
		To:   to,
	}

	r.addCommand(command)
	r.SkipTransaction()
}

// SetTable changes the table name for this blueprint.
// This is useful when the table name needs to be modified after blueprint creation.
func (r *Blueprint) SetTable(name string) {
	r.table = name
}

// ShortID creates a new short string primary key column on the table.
// Suitable for client-generated short IDs (e.g., Crockford Base32).
func (r *Blueprint) ShortID(column ...string) schema.ColumnDefinition {
	name := constants.DefaultIDColumn
	if len(column) > 0 {
		name = column[0]
	}
	col := r.String(name, 21)
	r.Primary(name)
	return col
}

// ShouldSkipTransaction returns true if the blueprint should skip the transaction wrapper.
// This is used for SQLite-specific operations like RenameIndex that have issues with DDL in transactions.
func (r *Blueprint) ShouldSkipTransaction() bool {
	return r.skipTransaction
}

// SkipTransaction marks the blueprint to skip the transaction wrapper during build.
// This is primarily used for SQLite RenameIndex operations to avoid savepoint timeout issues.
func (r *Blueprint) SkipTransaction() {
	r.skipTransaction = true
}

// SmallIncrements creates a new auto-incrementing small integer column on the table.
// The column will be unsigned and set to auto-increment.
func (r *Blueprint) SmallIncrements(column string) schema.ColumnDefinition {
	return r.UnsignedSmallInteger(column).AutoIncrement()
}

// SmallInteger creates a new small integer column on the table.
// Small integers can store values from -32,768 to 32,767.
func (r *Blueprint) SmallInteger(column string) schema.ColumnDefinition {
	return r.createAndAddColumn("smallInteger", column)
}

// SoftDeletes creates a new nullable timestamp column for soft deletes.
// If no column name is specified, it defaults to "soft_deleted_at".
// Soft deletes mark records as deleted without actually removing them from the database.
func (r *Blueprint) SoftDeletes(column ...string) schema.ColumnDefinition {
	newColumn := constants.SoftDeleteAtColumn
	if len(column) > 0 {
		newColumn = column[0]
	}

	return r.Timestamp(newColumn).Nullable()
}

// SoftDeletesTz creates a new nullable timestamp with timezone column for soft deletes.
// If no column name is specified, it defaults to "soft_deleted_at".
// This is similar to SoftDeletes but includes timezone information.
func (r *Blueprint) SoftDeletesTz(column ...string) schema.ColumnDefinition {
	newColumn := constants.SoftDeleteAtColumn
	if len(column) > 0 {
		newColumn = column[0]
	}

	return r.TimestampTz(newColumn).Nullable()
}

// SoftDeletesMaxDate creates a soft delete column using the max-date sentinel strategy.
// The column is NOT NULL and defaults to "9999-12-31 23:59:59" to indicate active records.
// This is useful when you need NOT NULL constraints or better index performance.
func (r *Blueprint) SoftDeletesMaxDate(column ...string) schema.ColumnDefinition {
	newColumn := constants.SoftDeleteAtColumn
	if len(column) > 0 {
		newColumn = column[0]
	}

	return r.Timestamp(newColumn).Default(constants.MaxSoftDeletedAtDefault)
}

// SoftDeletesMaxDateTz creates a soft delete column with timezone using the max-date sentinel strategy.
// The column is NOT NULL and defaults to "9999-12-31 23:59:59" to indicate active records.
// This is useful when you need NOT NULL constraints or better index performance.
func (r *Blueprint) SoftDeletesMaxDateTz(column ...string) schema.ColumnDefinition {
	newColumn := constants.SoftDeleteAtColumn
	if len(column) > 0 {
		newColumn = column[0]
	}

	return r.TimestampTz(newColumn).Default(constants.MaxSoftDeletedAtDefault)
}

// String creates a new variable-length string column on the table.
// The length parameter is optional; if not provided, it defaults to the DefaultStringLength constant.
// String columns are flexible and use storage proportional to actual content length.
func (r *Blueprint) String(column string, length ...int) schema.ColumnDefinition {
	defaultLength := constants.DefaultStringLength
	if len(length) > 0 {
		defaultLength = length[0]
	}

	columnImpl := r.createAndAddColumn("string", column)
	columnImpl.length = &defaultLength

	return columnImpl
}

// Text creates a new text column on the table.
// Text columns can store variable-length strings with no specified maximum length.
func (r *Blueprint) Text(column string) schema.ColumnDefinition {
	return r.createAndAddColumn("text", column)
}

// Time creates a new time column on the table.
// The precision parameter is optional and specifies fractional seconds precision (0-6).
// Time columns store time of day without date components.
func (r *Blueprint) Time(column string, precision ...int) schema.ColumnDefinition {
	columnImpl := r.createAndAddColumn("time", column)
	if len(precision) > 0 {
		columnImpl.precision = &precision[0]
	}

	return columnImpl
}

// TimeTz creates a new time with timezone column on the table.
// The precision parameter is optional and specifies fractional seconds precision (0-6).
// TimeTz columns store time of day with timezone information.
func (r *Blueprint) TimeTz(column string, precision ...int) schema.ColumnDefinition {
	columnImpl := r.createAndAddColumn("timeTz", column)
	if len(precision) > 0 {
		columnImpl.precision = &precision[0]
	}

	return columnImpl
}

// Timestamp creates a new timestamp column on the table.
// The precision parameter is optional and specifies fractional seconds precision (0-6).
// Timestamp columns store date and time information.
func (r *Blueprint) Timestamp(column string, precision ...int) schema.ColumnDefinition {
	columnImpl := r.createAndAddColumn("timestamp", column)
	if len(precision) > 0 {
		columnImpl.precision = &precision[0]
	}

	return columnImpl
}

// Timestamps adds created_at and updated_at timestamp columns to the table.
// Both columns are nullable and use the same precision if specified.
// This is a convenience method for adding automatic timestamp tracking.
func (r *Blueprint) Timestamps(precision ...int) {
	r.Timestamp(constants.DefaultCreatedAtColumn, precision...).Nullable()
	r.Timestamp(constants.DefaultUpdatedAtColumn, precision...).Nullable()
}

// TimestampsTz adds created_at and updated_at timestamp with timezone columns to the table.
// Both columns are nullable and use the same precision if specified.
// This is similar to Timestamps but includes timezone information.
func (r *Blueprint) TimestampsTz(precision ...int) {
	r.TimestampTz(constants.DefaultCreatedAtColumn, precision...).Nullable()
	r.TimestampTz(constants.DefaultUpdatedAtColumn, precision...).Nullable()
}

// TimestampTz creates a new timestamp with timezone column on the table.
// The precision parameter is optional and specifies fractional seconds precision (0-6).
// TimestampTz columns store date, time, and timezone information.
func (r *Blueprint) TimestampTz(column string, precision ...int) schema.ColumnDefinition {
	columnImpl := r.createAndAddColumn("timestampTz", column)
	if len(precision) > 0 {
		columnImpl.precision = &precision[0]
	}

	return columnImpl
}

// TinyIncrements creates a new auto-incrementing tiny integer column on the table.
// The column will be unsigned and set to auto-increment.
func (r *Blueprint) TinyIncrements(column string) schema.ColumnDefinition {
	return r.UnsignedTinyInteger(column).AutoIncrement()
}

// TinyInteger creates a new tiny integer column on the table.
// Tiny integers can store values from -128 to 127.
func (r *Blueprint) TinyInteger(column string) schema.ColumnDefinition {
	return r.createAndAddColumn("tinyInteger", column)
}

// TinyText creates a new tiny text column on the table.
// Tiny text columns can store small amounts of text data (up to 255 bytes in some databases).
func (r *Blueprint) TinyText(column string) schema.ColumnDefinition {
	return r.createAndAddColumn("tinyText", column)
}

// ToSql converts the blueprint into a slice of SQL statements using the provided grammar.
// It processes all commands and generates the appropriate database-specific SQL.
// Returns an error if any command fails to compile.
func (r *Blueprint) ToSql(grammar schema.Grammar) ([]string, error) {
	r.addImpliedCommands(grammar)

	// Generate CHANGE commands for columns marked as changed
	for _, column := range r.columns {
		if column.GetChange() && !r.hasChangeCommand(column.GetName()) {
			r.addCommand(&schema.Command{
				Name:   constants.CommandChange,
				Column: column,
			})
		}
	}

	var statements []string
	for _, command := range r.commands {
		if command.ShouldBeSkipped {
			continue
		}

		switch command.Name {
		case constants.CommandAdd:
			// Skip ADD command if the column is marked for change
			if command.Column != nil && command.Column.GetChange() {
				continue
			}
			stmt, err := grammar.CompileAdd(r, command)
			if err != nil {
				return nil, err
			}
			statements = append(statements, stmt)
		case constants.CommandChange:
			stmt, err := grammar.CompileChange(r, command)
			if err != nil {
				return nil, err
			}
			if stmt != "" {
				statements = append(statements, stmt)
			}
		case constants.CommandComment:
			stmt, err := grammar.CompileComment(r, command)
			if err != nil {
				return nil, err
			}
			if stmt != "" {
				statements = append(statements, stmt)
			}
		case constants.CommandCreate:
			stmt, err := grammar.CompileCreate(r)
			if err != nil {
				return nil, err
			}
			statements = append(statements, stmt)
		case constants.CommandDrop:
			stmt, err := grammar.CompileDrop(r)
			if err != nil {
				return nil, err
			}
			statements = append(statements, stmt)
		case constants.CommandDropColumn:
			stmts, err := grammar.CompileDropColumn(r, command)
			if err != nil {
				return nil, err
			}
			statements = append(statements, stmts...)
		case constants.CommandDropForeign:
			stmt, err := grammar.CompileDropForeign(r, command)
			if err != nil {
				return nil, err
			}
			statements = append(statements, stmt)
		case constants.CommandDropFullText:
			stmt, err := grammar.CompileDropFullText(r, command)
			if err != nil {
				return nil, err
			}
			statements = append(statements, stmt)
		case constants.CommandDropIfExists:
			stmt, err := grammar.CompileDropIfExists(r)
			if err != nil {
				return nil, err
			}
			statements = append(statements, stmt)
		case constants.CommandDropIndex:
			stmt, err := grammar.CompileDropIndex(r, command)
			if err != nil {
				return nil, err
			}
			statements = append(statements, stmt)
		case constants.CommandDropPrimary:
			stmt, err := grammar.CompileDropPrimary(r, command)
			if err != nil {
				return nil, err
			}
			statements = append(statements, stmt)
		case constants.CommandDropUnique:
			stmt, err := grammar.CompileDropUnique(r, command)
			if err != nil {
				return nil, err
			}
			statements = append(statements, stmt)
		case constants.CommandForeign:
			stmt, err := grammar.CompileForeign(r, command)
			if err != nil {
				return nil, err
			}
			statements = append(statements, stmt)
		case constants.CommandFullText:
			stmt, err := grammar.CompileFullText(r, command)
			if err != nil {
				return nil, err
			}
			statements = append(statements, stmt)
		case constants.CommandIndex:
			stmt, err := grammar.CompileIndex(r, command)
			if err != nil {
				return nil, err
			}
			statements = append(statements, stmt)
		case constants.CommandPrimary:
			stmt, err := grammar.CompilePrimary(r, command)
			if err != nil {
				return nil, err
			}
			statements = append(statements, stmt)
		case constants.CommandRename:
			stmt, err := grammar.CompileRename(r, command)
			if err != nil {
				return nil, err
			}
			statements = append(statements, stmt)
		case constants.CommandRenameColumn:
			stmt, err := grammar.CompileRenameColumn(r, command)
			if err != nil {
				return nil, err
			}
			statements = append(statements, stmt)
		case constants.CommandRenameIndex:
			stmts, err := grammar.CompileRenameIndex(r.schema, r, command)
			if err != nil {
				return nil, err
			}
			statements = append(statements, stmts...)
		case constants.CommandUnique:
			stmt, err := grammar.CompileUnique(r, command)
			if err != nil {
				return nil, err
			}
			statements = append(statements, stmt)
		}
	}

	return statements, nil
}

// Unique creates a new unique index on the table.
// The column parameter specifies the column(s) to include in the unique index.
// Returns an IndexDefinition that can be used to configure the index further.
func (r *Blueprint) Unique(column ...string) schema.IndexDefinition {
	command := r.indexCommand(constants.CommandUnique, column)

	return NewIndexDefinition(command)
}

// UnsignedBigInteger creates a new unsigned big integer column on the table.
// Unsigned big integers can store values from 0 to 18,446,744,073,709,551,615.
func (r *Blueprint) UnsignedBigInteger(column string) schema.ColumnDefinition {
	return r.BigInteger(column).Unsigned()
}

// UnsignedInteger creates a new unsigned integer column on the table.
// Unsigned integers can store values from 0 to 4,294,967,295.
func (r *Blueprint) UnsignedInteger(column string) schema.ColumnDefinition {
	return r.Integer(column).Unsigned()
}

// UnsignedMediumInteger creates a new unsigned medium integer column on the table.
// Unsigned medium integers can store values from 0 to 16,777,215.
func (r *Blueprint) UnsignedMediumInteger(column string) schema.ColumnDefinition {
	return r.MediumInteger(column).Unsigned()
}

// UnsignedSmallInteger creates a new unsigned small integer column on the table.
// Unsigned small integers can store values from 0 to 65,535.
func (r *Blueprint) UnsignedSmallInteger(column string) schema.ColumnDefinition {
	return r.SmallInteger(column).Unsigned()
}

// UnsignedTinyInteger creates a new unsigned tiny integer column on the table.
// Unsigned tiny integers can store values from 0 to 255.
func (r *Blueprint) UnsignedTinyInteger(column string) schema.ColumnDefinition {
	return r.TinyInteger(column).Unsigned()
}

// addAttributeCommands adds attribute commands (like comments) for columns.
// This is called automatically during SQL generation to ensure column attributes are included.
func (r *Blueprint) addAttributeCommands(grammar schema.Grammar) {
	attributeCommands := grammar.GetAttributeCommands()
	for _, column := range r.columns {
		for _, command := range attributeCommands {
			if command == constants.CommandComment && column.comment != nil {
				r.addCommand(&schema.Command{
					Column: column,
					Name:   constants.CommandComment,
				})
			}
		}
	}
}

// addCommand adds a command to the blueprint's command list.
func (r *Blueprint) addCommand(command *schema.Command) {
	r.commands = append(r.commands, command)
}

// addImpliedCommands adds commands that are implied by the blueprint's state.
// This includes attribute commands and other derived commands.
func (r *Blueprint) addImpliedCommands(grammar schema.Grammar) {
	r.addAttributeCommands(grammar)
}

// createAndAddColumn creates a new column definition and adds it to the blueprint.
// If the blueprint is not in create mode, an ADD command is automatically generated.
// Returns the created column definition for further configuration.
func (r *Blueprint) createAndAddColumn(ttype, name string) *ColumnDefinition {
	columnImpl := &ColumnDefinition{
		name:  &name,
		ttype: convert.Pointer(ttype),
	}

	r.columns = append(r.columns, columnImpl)

	if !r.isCreate() {
		r.addCommand(&schema.Command{
			Name:   constants.CommandAdd,
			Column: columnImpl,
		})
	}

	return columnImpl
}

// createIndexName generates a standardized index name based on the table, columns, and index type.
// The name includes the table prefix and replaces special characters with underscores.
func (r *Blueprint) createIndexName(ttype string, columns []string) string {
	var table string
	if strings.Contains(r.table, ".") {
		lastDotIndex := strings.LastIndex(r.table, ".")
		table = r.table[:lastDotIndex+1] + r.prefix + r.table[lastDotIndex+1:]
	} else {
		table = r.prefix + r.table
	}

	index := strings.ToLower(fmt.Sprintf("%s_%s_%s", table, strings.Join(columns, "_"), ttype))

	index = strings.ReplaceAll(index, "-", "_")
	index = strings.ReplaceAll(index, ".", "_")

	return index
}

// indexCommand creates a new index command and adds it to the blueprint.
// The config parameter allows specifying custom index properties like name, algorithm, or language.
// If no config is provided, an index name is automatically generated.
func (r *Blueprint) indexCommand(name string, columns []string, config ...schema.IndexConfig) *schema.Command {
	command := &schema.Command{
		Columns: columns,
		Name:    name,
	}

	if len(config) > 0 {
		command.Algorithm = config[0].Algorithm
		command.Index = config[0].Name
		command.Language = config[0].Language
	} else {
		command.Index = r.createIndexName(name, columns)
	}

	r.addCommand(command)

	return command
}

// isCreate checks if the blueprint is in create mode.
// Returns true if a CREATE command has been added to the blueprint.
func (r *Blueprint) isCreate() bool {
	for _, command := range r.commands {
		if command.Name == constants.CommandCreate {
			return true
		}
	}

	return false
}

// hasChangeCommand checks if a CHANGE command already exists for the given column
func (r *Blueprint) hasChangeCommand(columnName string) bool {
	for _, cmd := range r.commands {
		if cmd.Name == constants.CommandChange && cmd.Column != nil && cmd.Column.GetName() == columnName {
			return true
		}
	}
	return false
}

// modifyColumn marks the most recently added column for modification.
// This is used by the Change method to convert an ADD command into a CHANGE command.
// Returns the modified column definition, or nil if no columns exist.
func (r *Blueprint) modifyColumn() *ColumnDefinition {
	if len(r.columns) == 0 {
		return nil
	}

	column := r.columns[len(r.columns)-1]

	// Mark the column as changed
	column.change = convert.Pointer(true)

	// Remove the ADD command for this specific column
	for i := len(r.commands) - 1; i >= 0; i-- {
		cmd := r.commands[i]
		if cmd.Name == constants.CommandAdd && cmd.Column.GetName() == column.GetName() {
			r.commands = append(r.commands[:i], r.commands[i+1:]...)
			break
		}
	}

	r.addCommand(&schema.Command{
		Name:   constants.CommandChange,
		Column: column,
	})

	return column
}
