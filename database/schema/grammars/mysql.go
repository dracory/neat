package grammars

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cast"

	contractsdatabase "github.com/dracory/neat/contracts/database"
	"github.com/dracory/neat/contracts/database/schema"
)

type Mysql struct {
	attributeCommands []string
	modifiers         []func(schema.Blueprint, schema.ColumnDefinition) string
	serials           []string
	wrap              *Wrap
}

func NewMysql(tablePrefix string) *Mysql {
	mysql := &Mysql{
		attributeCommands: []string{},
		serials:           []string{"bigInteger", "integer", "mediumInteger", "smallInteger", "tinyInteger"},
		wrap:              NewWrap(contractsdatabase.DriverMysql, tablePrefix),
	}
	mysql.modifiers = []func(schema.Blueprint, schema.ColumnDefinition) string{
		// The sort should not be changed, it effects the SQL output
		mysql.ModifyUnsigned,
		mysql.ModifyCollation,
		mysql.ModifyNullable,
		mysql.ModifyDefault,
		mysql.ModifyOnUpdate,
		mysql.ModifyIncrement,
		mysql.ModifyComment,
		mysql.ModifyAfter,
		mysql.ModifyFirst,
	}

	return mysql
}

func (r *Mysql) CompileAdd(blueprint schema.Blueprint, command *schema.Command) (string, error) {
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

func (r *Mysql) CompileChange(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return "", err
	}
	column, err := r.getColumn(blueprint, command.Column)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("alter table %s modify %s", table, column), nil
}

func (r *Mysql) CompileColumns(schema, table string) string {
	return fmt.Sprintf(
		"select column_name as `name`, data_type as `type_name`, column_type as `type`, "+
			"coalesce(collation_name, '') as `collation`, is_nullable as `nullable`, "+
			"column_default as `default`, column_comment as `comment`, "+
			"generation_expression as `expression`, extra as `extra` "+
			"from information_schema.columns where table_schema = %s and table_name = %s "+
			"order by ordinal_position asc", r.wrap.Quote(schema), r.wrap.Quote(table))
}

func (r *Mysql) CompileComment(_ schema.Blueprint, _ *schema.Command) (string, error) {
	return "", nil
}

func (r *Mysql) CompileCreate(blueprint schema.Blueprint) (string, error) {
	columns, err := r.getColumns(blueprint)
	if err != nil {
		return "", err
	}
	primaryCommand := getCommandByName(blueprint.GetCommands(), "primary")
	if primaryCommand != nil {
		var algorithm string
		if primaryCommand.Algorithm != "" {
			algorithm = "using " + primaryCommand.Algorithm
		}
		columnized, err := r.wrap.Columnize(primaryCommand.Columns)
		if err != nil {
			return "", err
		}
		columns = append(columns, fmt.Sprintf("primary key %s(%s)", algorithm, columnized))

		primaryCommand.ShouldBeSkipped = true
	}

	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("create table %s (%s)", table, strings.Join(columns, ", ")), nil
}

func (r *Mysql) CompileDisableForeignKeyConstraints() string {
	return "SET FOREIGN_KEY_CHECKS=0;"
}

func (r *Mysql) CompileDrop(blueprint schema.Blueprint) (string, error) {
	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("drop table %s", table), nil
}

func (r *Mysql) CompileDropAllDomains(_ []string) (string, error) {
	return "", nil
}

func (r *Mysql) CompileDropAllTables(tables []string) (string, error) {
	columnized, err := r.wrap.Columnize(tables)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("drop table %s", columnized), nil
}

func (r *Mysql) CompileDropAllTypes(_ []string) (string, error) {
	return "", nil
}

func (r *Mysql) CompileDropAllViews(views []string) (string, error) {
	columnized, err := r.wrap.Columnize(views)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("drop view %s", columnized), nil
}

func (r *Mysql) CompileDropColumn(blueprint schema.Blueprint, command *schema.Command) ([]string, error) {
	columns, err := r.wrap.Columns(command.Columns)
	if err != nil {
		return nil, err
	}
	prefixed := r.wrap.PrefixArray("drop", columns)

	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return nil, err
	}

	return []string{
		fmt.Sprintf("alter table %s %s", table, strings.Join(prefixed, ", ")),
	}, nil
}

func (r *Mysql) CompileDropForeign(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return "", err
	}
	column, err := r.wrap.Column(command.Index)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("alter table %s drop foreign key %s", table, column), nil
}

func (r *Mysql) CompileDropFullText(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	return r.CompileDropIndex(blueprint, command)
}

func (r *Mysql) CompileDropIfExists(blueprint schema.Blueprint) (string, error) {
	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("drop table if exists %s", table), nil
}

func (r *Mysql) CompileDropIndex(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return "", err
	}
	column, err := r.wrap.Column(command.Index)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("alter table %s drop index %s", table, column), nil
}

func (r *Mysql) CompileDropPrimary(blueprint schema.Blueprint, _ *schema.Command) (string, error) {
	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("alter table %s drop primary key", table), nil
}

func (r *Mysql) CompileDropUnique(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	return r.CompileDropIndex(blueprint, command)
}

func (r *Mysql) CompileEnableForeignKeyConstraints() string {
	return "SET FOREIGN_KEY_CHECKS=1;"
}

func (r *Mysql) CompileForeign(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return "", err
	}
	index, err := r.wrap.Column(command.Index)
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
		table, index, columns, on, references)
	if command.OnDelete != "" && r.wrap.IsValidAction(command.OnDelete) {
		sql += " on delete " + command.OnDelete
	}
	if command.OnUpdate != "" && r.wrap.IsValidAction(command.OnUpdate) {
		sql += " on update " + command.OnUpdate
	}

	return sql, nil
}

func (r *Mysql) CompileForeignKeys(schema, table string) string {
	return fmt.Sprintf(
		`SELECT 
			kc.constraint_name AS name, 
			GROUP_CONCAT(kc.column_name ORDER BY kc.ordinal_position) AS columns, 
			kc.referenced_table_schema AS foreign_schema, 
			kc.referenced_table_name AS foreign_table, 
			GROUP_CONCAT(kc.referenced_column_name ORDER BY kc.ordinal_position) AS foreign_columns, 
			rc.update_rule AS on_update, 
			rc.delete_rule AS on_delete 
		FROM information_schema.key_column_usage kc 
		JOIN information_schema.referential_constraints rc 
			ON kc.constraint_schema = rc.constraint_schema 
			AND kc.constraint_name = rc.constraint_name 
		WHERE kc.table_schema = %s 
			AND kc.table_name = %s 
			AND kc.referenced_table_name IS NOT NULL 
		GROUP BY 
			kc.constraint_name, 
			kc.referenced_table_schema, 
			kc.referenced_table_name, 
			rc.update_rule, 
			rc.delete_rule`,
		r.wrap.Quote(schema),
		r.wrap.Quote(table),
	)
}

func (r *Mysql) CompileFullText(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	return r.compileKey(blueprint, command, "fulltext")
}

func (r *Mysql) CompileIndex(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	var algorithm string
	if command.Algorithm != "" {
		algorithm = " using " + command.Algorithm
	}

	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return "", err
	}
	index, err := r.wrap.Column(command.Index)
	if err != nil {
		return "", err
	}
	columns, err := r.wrap.Columnize(command.Columns)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("alter table %s add %s %s%s(%s)",
		table,
		"index",
		index,
		algorithm,
		columns,
	), nil
}

func (r *Mysql) CompileIndexes(schema, table string) string {
	return fmt.Sprintf(
		"select index_name as `name`, group_concat(column_name order by seq_in_index) as `columns`, "+
			"index_type as `type`, not non_unique as `unique` "+
			"from information_schema.statistics where table_schema = %s and table_name = %s "+
			"group by index_name, index_type, non_unique",
		r.wrap.Quote(schema),
		r.wrap.Quote(table),
	)
}

func (r *Mysql) CompilePrimary(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	var algorithm string
	if command.Algorithm != "" {
		algorithm = "using " + command.Algorithm
	}

	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return "", err
	}
	columns, err := r.wrap.Columnize(command.Columns)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("alter table %s add primary key %s(%s)", table, algorithm, columns), nil
}

func (r *Mysql) CompileRename(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return "", err
	}
	to, err := r.wrap.Table(command.To)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("rename table %s to %s", table, to), nil
}

func (r *Mysql) CompileRenameColumn(blueprint schema.Blueprint, command *schema.Command) (string, error) {
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

	return fmt.Sprintf("alter table %s rename column %s to %s", table, from, to), nil
}

func (r *Mysql) CompileRenameIndex(_ schema.Schema, blueprint schema.Blueprint, command *schema.Command) ([]string, error) {
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
		fmt.Sprintf("alter table %s rename index %s to %s", table, from, to),
	}, nil
}

func (r *Mysql) CompileTables(database string) string {
	return fmt.Sprintf("select table_name as `name`, (data_length + index_length) as `size`, "+
		"table_comment as `comment`, engine as `engine`, table_collation as `collation` "+
		"from information_schema.tables where table_schema = %s and table_type in ('BASE TABLE', 'SYSTEM VERSIONED') "+
		"order by table_name", r.wrap.Quote(database))
}

func (r *Mysql) CompileTypes() string {
	return ""
}

func (r *Mysql) CompileUnique(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	return r.compileKey(blueprint, command, "unique")
}

func (r *Mysql) CompileViews(database string) string {
	return fmt.Sprintf("select table_name as `name`, view_definition as `definition` "+
		"from information_schema.views where table_schema = %s "+
		"order by table_name", r.wrap.Quote(database))
}

func (r *Mysql) GetAttributeCommands() []string {
	return r.attributeCommands
}

func (r *Mysql) ModifyAfter(_ schema.Blueprint, column schema.ColumnDefinition) string {
	if after := column.GetAfter(); after != "" {
		return " after " + after
	}

	return ""
}

func (r *Mysql) ModifyCollation(_ schema.Blueprint, column schema.ColumnDefinition) string {
	if collation := column.GetCollation(); collation != "" {
		return " collate " + collation
	}

	return ""
}

func (r *Mysql) ModifyComment(_ schema.Blueprint, column schema.ColumnDefinition) string {
	if comment := column.GetComment(); comment != "" {
		// Escape special characters to prevent SQL injection
		comment = strings.ReplaceAll(comment, "'", "''")
		comment = strings.ReplaceAll(comment, "\\", "\\\\")

		return fmt.Sprintf(" comment '%s'", comment)
	}

	return ""
}

func (r *Mysql) ModifyDefault(_ schema.Blueprint, column schema.ColumnDefinition) string {
	if column.GetDefault() != nil {
		return fmt.Sprintf(" default %s", getDefaultValue(column.GetDefault()))
	}

	return ""
}

func (r *Mysql) ModifyNullable(_ schema.Blueprint, column schema.ColumnDefinition) string {
	if column.GetNullable() {
		return " null"
	} else {
		return " not null"
	}
}

func (r *Mysql) ModifyFirst(_ schema.Blueprint, column schema.ColumnDefinition) string {
	if column.GetFirst() {
		return " first"
	}

	return ""
}

func (r *Mysql) ModifyIncrement(blueprint schema.Blueprint, column schema.ColumnDefinition) string {
	if slices.Contains(r.serials, column.GetType()) && column.GetAutoIncrement() {
		if blueprint.HasCommand("primary") {
			return "auto_increment"
		}
		return " auto_increment primary key"
	}

	return ""
}

func (r *Mysql) ModifyOnUpdate(_ schema.Blueprint, column schema.ColumnDefinition) string {
	onUpdate := column.GetOnUpdate()
	if onUpdate != nil {
		switch value := onUpdate.(type) {
		case Expression:
			return " on update " + string(value)
		case string:
			if onUpdate.(string) != "" {
				return " on update " + value
			}
		}
	}

	return ""
}

func (r *Mysql) ModifyUnsigned(_ schema.Blueprint, column schema.ColumnDefinition) string {
	if column.GetUnsigned() {
		return " unsigned"
	}

	return ""
}

func (r *Mysql) TypeBigInteger(_ schema.ColumnDefinition) string {
	return "bigint"
}

func (r *Mysql) TypeBoolean(_ schema.ColumnDefinition) string {
	return "tinyint(1)"
}

func (r *Mysql) TypeChar(column schema.ColumnDefinition) string {
	return fmt.Sprintf("char(%d)", column.GetLength())
}

func (r *Mysql) TypeDate(_ schema.ColumnDefinition) string {
	return "date"
}

func (r *Mysql) TypeDateTime(column schema.ColumnDefinition) string {
	current := "CURRENT_TIMESTAMP"
	precision := column.GetPrecision()
	if precision > 0 {
		current = fmt.Sprintf("CURRENT_TIMESTAMP(%d)", precision)
	}
	if column.GetUseCurrent() {
		column.Default(Expression(current))
	}
	if column.GetUseCurrentOnUpdate() {
		column.OnUpdate(Expression(current))
	}

	if precision > 0 {
		return fmt.Sprintf("datetime(%d)", precision)
	} else {
		return "datetime"
	}
}

func (r *Mysql) TypeDateTimeTz(column schema.ColumnDefinition) string {
	return r.TypeDateTime(column)
}

func (r *Mysql) TypeDecimal(column schema.ColumnDefinition) string {
	return fmt.Sprintf("decimal(%d, %d)", column.GetTotal(), column.GetPlaces())
}

func (r *Mysql) TypeDouble(_ schema.ColumnDefinition) string {
	return "double"
}

func (r *Mysql) TypeEnum(column schema.ColumnDefinition) string {
	return fmt.Sprintf(`enum(%s)`, strings.Join(r.wrap.Quotes(cast.ToStringSlice(column.GetAllowed())), ", "))
}

func (r *Mysql) TypeFloat(column schema.ColumnDefinition) string {
	precision := column.GetPrecision()
	if precision > 0 {
		return fmt.Sprintf("float(%d)", precision)
	}

	return "float"
}

func (r *Mysql) TypeInteger(_ schema.ColumnDefinition) string {
	return "int"
}

func (r *Mysql) TypeJson(_ schema.ColumnDefinition) string {
	return "json"
}

func (r *Mysql) TypeJsonb(_ schema.ColumnDefinition) string {
	return "json"
}

func (r *Mysql) TypeLongText(_ schema.ColumnDefinition) string {
	return "longtext"
}

func (r *Mysql) TypeMediumInteger(_ schema.ColumnDefinition) string {
	return "mediumint"
}

func (r *Mysql) TypeMediumText(_ schema.ColumnDefinition) string {
	return "mediumtext"
}

func (r *Mysql) TypeSmallInteger(_ schema.ColumnDefinition) string {
	return "smallint"
}

func (r *Mysql) TypeString(column schema.ColumnDefinition) string {
	length := column.GetLength()
	if length > 0 {
		return fmt.Sprintf("varchar(%d)", length)
	}

	return "varchar(255)"
}

func (r *Mysql) TypeText(_ schema.ColumnDefinition) string {
	return "text"
}

func (r *Mysql) TypeTime(column schema.ColumnDefinition) string {
	if column.GetPrecision() > 0 {
		return fmt.Sprintf("time(%d)", column.GetPrecision())
	} else {
		return "time"
	}
}

func (r *Mysql) TypeTimeTz(column schema.ColumnDefinition) string {
	return r.TypeTime(column)
}

func (r *Mysql) TypeTimestamp(column schema.ColumnDefinition) string {
	current := "CURRENT_TIMESTAMP"
	precision := column.GetPrecision()
	if precision > 0 {
		current = fmt.Sprintf("CURRENT_TIMESTAMP(%d)", precision)
	}
	if column.GetUseCurrent() {
		column.Default(Expression(current))
	}
	if column.GetUseCurrentOnUpdate() {
		column.OnUpdate(Expression(current))
	}

	if precision > 0 {
		return fmt.Sprintf("timestamp(%d)", precision)
	} else {
		return "timestamp"
	}
}

func (r *Mysql) TypeTimestampTz(column schema.ColumnDefinition) string {
	return r.TypeTimestamp(column)
}

func (r *Mysql) TypeTinyInteger(_ schema.ColumnDefinition) string {
	return "tinyint"
}

func (r *Mysql) TypeTinyText(_ schema.ColumnDefinition) string {
	return "tinytext"
}

func (r *Mysql) compileKey(blueprint schema.Blueprint, command *schema.Command, ttype string) (string, error) {
	var algorithm string
	if command.Algorithm != "" {
		algorithm = " using " + command.Algorithm
	}

	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return "", err
	}
	index, err := r.wrap.Column(command.Index)
	if err != nil {
		return "", err
	}
	columns, err := r.wrap.Columnize(command.Columns)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("alter table %s add %s %s%s(%s)",
		table,
		ttype,
		index,
		algorithm,
		columns), nil
}

func (r *Mysql) getColumns(blueprint schema.Blueprint) ([]string, error) {
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

func (r *Mysql) getColumn(blueprint schema.Blueprint, column schema.ColumnDefinition) (string, error) {
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
