package grammars

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cast"

	"github.com/dracory/neat/contracts/database"
	"github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/database/schema/constants"
)

type Sqlserver struct {
	attributeCommands []string
	modifiers         []func(schema.Blueprint, schema.ColumnDefinition) string
	serials           []string
	wrap              *Wrap
}

func NewSqlserver(tablePrefix string) *Sqlserver {
	sqlserver := &Sqlserver{
		attributeCommands: []string{constants.CommandComment},
		serials:           []string{"bigInteger", "integer", "mediumInteger", "smallInteger", "tinyInteger"},
		wrap:              NewWrap(database.DriverSqlserver, tablePrefix),
	}
	sqlserver.modifiers = []func(schema.Blueprint, schema.ColumnDefinition) string{
		sqlserver.ModifyDefault,
		sqlserver.ModifyIncrement,
		sqlserver.ModifyNullable,
	}

	return sqlserver
}

func (r *Sqlserver) CompileAdd(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return "", err
	}
	column, err := r.getColumn(blueprint, command.Column)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("alter table %s add %s", table, column), nil
}

func (r *Sqlserver) CompileChange(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return "", err
	}
	column, err := r.getColumn(blueprint, command.Column)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("alter table %s alter column %s", table, column), nil
}

func (r *Sqlserver) CompileColumns(schema, table string) string {
	newSchema := "schema_name()"
	if schema != "" {
		newSchema = r.wrap.Quote(schema)
	}

	return fmt.Sprintf(
		"select col.name, type.name as type_name, "+
			"col.max_length as length, col.precision as precision, col.scale as places, "+
			"col.is_nullable as nullable, def.definition as [default], "+
			"col.is_identity as autoincrement, col.collation_name as collation, "+
			"com.definition as [expression], is_persisted as [persisted], "+
			"cast(prop.value as nvarchar(max)) as comment "+
			"from sys.columns as col "+
			"join sys.types as type on col.user_type_id = type.user_type_id "+
			"join sys.objects as obj on col.object_id = obj.object_id "+
			"join sys.schemas as scm on obj.schema_id = scm.schema_id "+
			"left join sys.default_constraints def on col.default_object_id = def.object_id and col.object_id = def.parent_object_id "+
			"left join sys.extended_properties as prop on obj.object_id = prop.major_id and col.column_id = prop.minor_id and prop.name = 'MS_Description' "+
			"left join sys.computed_columns as com on col.column_id = com.column_id and col.object_id = com.object_id "+
			"where obj.type in ('U', 'V') and obj.name = %s and scm.name = %s "+
			"order by col.column_id", r.wrap.Quote(table), newSchema)
}

func (r *Sqlserver) CompileComment(_ schema.Blueprint, _ *schema.Command) (string, error) {
	return "", nil
}

func (r *Sqlserver) CompileCreate(blueprint schema.Blueprint) (string, error) {
	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return "", err
	}
	columns, err := r.getColumns(blueprint)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("create table %s (%s)", table, strings.Join(columns, ", ")), nil
}

func (r *Sqlserver) CompileDrop(blueprint schema.Blueprint) (string, error) {
	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("drop table %s", table), nil
}

func (r *Sqlserver) CompileDropAllDomains(_ []string) (string, error) {
	return "", nil
}

func (r *Sqlserver) CompileDropAllForeignKeys() string {
	return `DECLARE @sql NVARCHAR(MAX) = N'';
            SELECT @sql += 'ALTER TABLE '
                + QUOTENAME(OBJECT_SCHEMA_NAME(parent_object_id)) + '.' + + QUOTENAME(OBJECT_NAME(parent_object_id))
                + ' DROP CONSTRAINT ' + QUOTENAME(name) + ';'
            FROM sys.foreign_keys;

            EXEC sp_executesql @sql;`
}

func (r *Sqlserver) CompileDropAllTables(_ []string) (string, error) {
	return "EXEC sp_msforeachtable 'DROP TABLE ?'", nil
}

func (r *Sqlserver) CompileDropAllTypes(_ []string) (string, error) {
	return "", nil
}

func (r *Sqlserver) CompileDropAllViews(_ []string) (string, error) {
	return `DECLARE @sql NVARCHAR(MAX) = N'';
	SELECT @sql += 'DROP VIEW ' + QUOTENAME(OBJECT_SCHEMA_NAME(object_id)) + '.' + QUOTENAME(name) + ';'
	FROM sys.views;

	EXEC sp_executesql @sql;`, nil
}

func (r *Sqlserver) CompileDropColumn(blueprint schema.Blueprint, command *schema.Command) ([]string, error) {
	columns, err := r.wrap.Columns(command.Columns)
	if err != nil {
		return nil, err
	}

	dropExistingConstraintsSql, err := r.CompileDropDefaultConstraint(blueprint, command)
	if err != nil {
		return nil, err
	}

	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return nil, err
	}

	return []string{
		fmt.Sprintf("%s alter table %s drop column %s", dropExistingConstraintsSql, table, strings.Join(columns, ", ")),
	}, nil
}

func (r *Sqlserver) CompileDropDefaultConstraint(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	columns := fmt.Sprintf("'%s'", strings.Join(command.Columns, "','"))
	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return "", err
	}
	tableName := r.wrap.Quote(table)

	return fmt.Sprintf("DECLARE @sql NVARCHAR(MAX) = '';"+
		"SELECT @sql += 'ALTER TABLE %s DROP CONSTRAINT ' + OBJECT_NAME([default_object_id]) + ';' "+
		"FROM sys.columns "+
		"WHERE [object_id] = OBJECT_ID(%s) AND [name] in (%s) AND [default_object_id] <> 0;"+
		"EXEC(@sql);", table, tableName, columns), nil
}

func (r *Sqlserver) CompileDropForeign(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return "", err
	}
	column, err := r.wrap.Column(command.Index)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("alter table %s drop constraint %s", table, column), nil
}

func (r *Sqlserver) CompileDropFullText(_ schema.Blueprint, _ *schema.Command) (string, error) {
	return "", nil
}

func (r *Sqlserver) CompileDropIfExists(blueprint schema.Blueprint) (string, error) {
	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("if object_id(%s, 'U') is not null drop table %s", r.wrap.Quote(table), table), nil
}

func (r *Sqlserver) CompileDropIndex(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	column, err := r.wrap.Column(command.Index)
	if err != nil {
		return "", err
	}
	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("drop index %s on %s", column, table), nil
}

func (r *Sqlserver) CompileDropPrimary(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return "", err
	}
	column, err := r.wrap.Column(command.Index)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("alter table %s drop constraint %s", table, column), nil
}

func (r *Sqlserver) CompileDropUnique(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	return r.CompileDropIndex(blueprint, command)
}

func (r *Sqlserver) CompileForeign(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return "", err
	}
	column, err := r.wrap.Column(command.Index)
	if err != nil {
		return "", err
	}
	columns, err := r.wrap.Columnize(command.Columns)
	if err != nil {
		return "", err
	}
	on, err := r.wrap.Table(command.On)
	if err != nil {
		return "", err
	}
	references, err := r.wrap.Columnize(command.References)
	if err != nil {
		return "", err
	}
	sql := fmt.Sprintf("alter table %s add constraint %s foreign key (%s) references %s (%s)",
		table, column, columns, on, references)
	if command.OnDelete != "" && r.wrap.IsValidAction(command.OnDelete) {
		sql += " on delete " + command.OnDelete
	}
	if command.OnUpdate != "" && r.wrap.IsValidAction(command.OnUpdate) {
		sql += " on update " + command.OnUpdate
	}

	return sql, nil
}

func (r *Sqlserver) CompileForeignKeys(schema, table string) string {
	newSchema := "schema_name()"
	if schema != "" {
		newSchema = r.wrap.Quote(schema)
	}

	return fmt.Sprintf(
		`SELECT 
			fk.name AS name, 
			string_agg(lc.name, ',') WITHIN GROUP (ORDER BY fkc.constraint_column_id) AS columns, 
			fs.name AS foreign_schema, 
			ft.name AS foreign_table, 
			string_agg(fc.name, ',') WITHIN GROUP (ORDER BY fkc.constraint_column_id) AS foreign_columns, 
			fk.update_referential_action_desc AS on_update, 
			fk.delete_referential_action_desc AS on_delete 
		FROM sys.foreign_keys AS fk 
		JOIN sys.foreign_key_columns AS fkc ON fkc.constraint_object_id = fk.object_id 
		JOIN sys.tables AS lt ON lt.object_id = fk.parent_object_id 
		JOIN sys.schemas AS ls ON lt.schema_id = ls.schema_id 
		JOIN sys.columns AS lc ON fkc.parent_object_id = lc.object_id AND fkc.parent_column_id = lc.column_id 
		JOIN sys.tables AS ft ON ft.object_id = fk.referenced_object_id 
		JOIN sys.schemas AS fs ON ft.schema_id = fs.schema_id 
		JOIN sys.columns AS fc ON fkc.referenced_object_id = fc.object_id AND fkc.referenced_column_id = fc.column_id 
		WHERE lt.name = %s AND ls.name = %s 
		GROUP BY fk.name, fs.name, ft.name, fk.update_referential_action_desc, fk.delete_referential_action_desc`,
		r.wrap.Quote(table),
		newSchema,
	)
}

func (r *Sqlserver) CompileFullText(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	return "", nil
}

func (r *Sqlserver) CompileIndex(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	column, err := r.wrap.Column(command.Index)
	if err != nil {
		return "", err
	}
	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return "", err
	}
	columns, err := r.wrap.Columnize(command.Columns)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("create index %s on %s (%s)", column, table, columns), nil
}

func (r *Sqlserver) CompileIndexes(schema, table string) string {
	newSchema := "schema_name()"
	if schema != "" {
		newSchema = r.wrap.Quote(schema)
	}

	return fmt.Sprintf(
		"select idx.name as name, string_agg(col.name, ',') within group (order by idxcol.key_ordinal) as columns, "+
			"idx.type_desc as [type], idx.is_unique as [unique], idx.is_primary_key as [primary] "+
			"from sys.indexes as idx "+
			"join sys.tables as tbl on idx.object_id = tbl.object_id "+
			"join sys.schemas as scm on tbl.schema_id = scm.schema_id "+
			"join sys.index_columns as idxcol on idx.object_id = idxcol.object_id and idx.index_id = idxcol.index_id "+
			"join sys.columns as col on idxcol.object_id = col.object_id and idxcol.column_id = col.column_id "+
			"where tbl.name = %s and scm.name = %s "+
			"group by idx.name, idx.type_desc, idx.is_unique, idx.is_primary_key",
		r.wrap.Quote(table),
		newSchema,
	)
}

func (r *Sqlserver) CompilePrimary(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return "", err
	}
	column, err := r.wrap.Column(command.Index)
	if err != nil {
		return "", err
	}
	columns, err := r.wrap.Columnize(command.Columns)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("alter table %s add constraint %s primary key (%s)", table, column, columns), nil
}

func (r *Sqlserver) CompileRename(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return "", err
	}
	to, err := r.wrap.Table(command.To)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("sp_rename %s, %s", r.wrap.Quote(table), to), nil
}

func (r *Sqlserver) CompileRenameColumn(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return "", err
	}
	from, err := r.wrap.Column(command.From)
	if err != nil {
		return "", err
	}
	to, err := r.wrap.Column(command.To)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("sp_rename %s, %s, N'COLUMN'", r.wrap.Quote(table+"."+from), to), nil
}

func (r *Sqlserver) CompileRenameIndex(_ schema.Schema, blueprint schema.Blueprint, command *schema.Command) ([]string, error) {
	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return nil, err
	}
	from, err := r.wrap.Column(command.From)
	if err != nil {
		return nil, err
	}
	to, err := r.wrap.Column(command.To)
	if err != nil {
		return nil, err
	}
	return []string{
		fmt.Sprintf("sp_rename %s, %s, N'INDEX'", r.wrap.Quote(table+"."+from), to),
	}, nil
}

func (r *Sqlserver) CompileTables(_ string) string {
	return "select t.name as name, schema_name(t.schema_id) as [schema], sum(u.total_pages) * 8 * 1024 as size " +
		"from sys.tables as t " +
		"join sys.partitions as p on p.object_id = t.object_id " +
		"join sys.allocation_units as u on u.container_id = p.hobt_id " +
		"group by t.name, t.schema_id " +
		"order by t.name"
}

func (r *Sqlserver) CompileTypes() string {
	return ""
}

func (r *Sqlserver) CompileUnique(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	column, err := r.wrap.Column(command.Index)
	if err != nil {
		return "", err
	}
	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return "", err
	}
	columns, err := r.wrap.Columnize(command.Columns)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("create unique index %s on %s (%s)", column, table, columns), nil
}

func (r *Sqlserver) CompileViews(_ string) string {
	return "select name, schema_name(v.schema_id) as [schema], definition from sys.views as v " +
		"inner join sys.sql_modules as m on v.object_id = m.object_id " +
		"order by name"
}

func (r *Sqlserver) GetAttributeCommands() []string {
	return r.attributeCommands
}

func (r *Sqlserver) ModifyDefault(_ schema.Blueprint, column schema.ColumnDefinition) string {
	if column.GetDefault() != nil {
		return fmt.Sprintf(" default %s", getDefaultValue(column.GetDefault()))
	}

	return ""
}

func (r *Sqlserver) ModifyNullable(_ schema.Blueprint, column schema.ColumnDefinition) string {
	if column.GetNullable() {
		return " null"
	} else {
		return " not null"
	}
}

func (r *Sqlserver) ModifyIncrement(blueprint schema.Blueprint, column schema.ColumnDefinition) string {
	if slices.Contains(r.serials, column.GetType()) && column.GetAutoIncrement() {
		if blueprint.HasCommand("primary") {
			return " identity"
		}
		return " identity primary key"
	}

	return ""
}

func (r *Sqlserver) TypeBigInteger(_ schema.ColumnDefinition) string {
	return "bigint"
}

func (r *Sqlserver) TypeBoolean(_ schema.ColumnDefinition) string {
	return "bit"
}

func (r *Sqlserver) TypeChar(column schema.ColumnDefinition) string {
	return fmt.Sprintf("nchar(%d)", column.GetLength())
}

func (r *Sqlserver) TypeDate(_ schema.ColumnDefinition) string {
	return "date"
}

func (r *Sqlserver) TypeDateTime(column schema.ColumnDefinition) string {
	return r.TypeTimestamp(column)
}

func (r *Sqlserver) TypeDateTimeTz(column schema.ColumnDefinition) string {
	return r.TypeTimestampTz(column)
}

func (r *Sqlserver) TypeDecimal(column schema.ColumnDefinition) string {
	return fmt.Sprintf("decimal(%d, %d)", column.GetTotal(), column.GetPlaces())
}

func (r *Sqlserver) TypeDouble(_ schema.ColumnDefinition) string {
	return "double precision"
}

func (r *Sqlserver) TypeEnum(column schema.ColumnDefinition) string {
	return fmt.Sprintf(`nvarchar(255) check ("%s" in (%s))`, column.GetName(), strings.Join(r.wrap.Quotes(cast.ToStringSlice(column.GetAllowed())), ", "))
}

func (r *Sqlserver) TypeFloat(column schema.ColumnDefinition) string {
	precision := column.GetPrecision()
	if precision > 0 {
		return fmt.Sprintf("float(%d)", precision)
	}

	return "float"
}

func (r *Sqlserver) TypeInteger(_ schema.ColumnDefinition) string {
	return "int"
}

func (r *Sqlserver) TypeJson(_ schema.ColumnDefinition) string {
	return "nvarchar(max)"
}

func (r *Sqlserver) TypeJsonb(_ schema.ColumnDefinition) string {
	return "nvarchar(max)"
}

func (r *Sqlserver) TypeLongText(_ schema.ColumnDefinition) string {
	return "nvarchar(max)"
}

func (r *Sqlserver) TypeMediumInteger(_ schema.ColumnDefinition) string {
	return "int"
}

func (r *Sqlserver) TypeMediumText(_ schema.ColumnDefinition) string {
	return "nvarchar(max)"
}

func (r *Sqlserver) TypeSmallInteger(_ schema.ColumnDefinition) string {
	return "smallint"
}

func (r *Sqlserver) TypeString(column schema.ColumnDefinition) string {
	length := column.GetLength()
	if length > 0 {
		return fmt.Sprintf("nvarchar(%d)", length)
	}

	return "nvarchar(255)"
}

func (r *Sqlserver) TypeText(_ schema.ColumnDefinition) string {
	return "nvarchar(max)"
}

func (r *Sqlserver) TypeTime(column schema.ColumnDefinition) string {
	if column.GetPrecision() > 0 {
		return fmt.Sprintf("time(%d)", column.GetPrecision())
	} else {
		return "time"
	}
}

func (r *Sqlserver) TypeTimeTz(column schema.ColumnDefinition) string {
	return r.TypeTime(column)
}

func (r *Sqlserver) TypeTimestamp(column schema.ColumnDefinition) string {
	if column.GetUseCurrent() {
		column.Default(Expression("CURRENT_TIMESTAMP"))
	}

	// Use datetime2 by default for better compatibility with Go's time.Time
	// datetime has limited range (1753-9999), datetime2 has full range (0001-9999)
	if column.GetPrecision() > 0 {
		return fmt.Sprintf("datetime2(%d)", column.GetPrecision())
	} else {
		return "datetime2"
	}
}

func (r *Sqlserver) TypeTimestampTz(column schema.ColumnDefinition) string {
	if column.GetUseCurrent() {
		column.Default(Expression("CURRENT_TIMESTAMP"))
	}

	if column.GetPrecision() > 0 {
		return fmt.Sprintf("datetimeoffset(%d)", column.GetPrecision())
	} else {
		return "datetimeoffset"
	}
}

func (r *Sqlserver) TypeTinyInteger(_ schema.ColumnDefinition) string {
	return "tinyint"
}

func (r *Sqlserver) TypeTinyText(_ schema.ColumnDefinition) string {
	return "nvarchar(255)"
}

func (r *Sqlserver) getColumn(blueprint schema.Blueprint, column schema.ColumnDefinition) (string, error) {
	col, err := r.wrap.Column(column.GetName())
	if err != nil {
		return "", err
	}
	sql := fmt.Sprintf("%s %s", col, getType(r, column))

	for _, modifier := range r.modifiers {
		sql += modifier(blueprint, column)
	}

	return sql, nil
}

func (r *Sqlserver) getColumns(blueprint schema.Blueprint) ([]string, error) {
	var columns []string
	for _, column := range blueprint.GetAddedColumns() {
		col, err := r.getColumn(blueprint, column)
		if err != nil {
			return nil, err
		}
		columns = append(columns, col)
	}

	return columns, nil
}
