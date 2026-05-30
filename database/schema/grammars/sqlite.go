package grammars

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cast"

	contractsdatabase "github.com/dracory/neat/contracts/database"
	"github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/contracts/log"
	"github.com/dracory/neat/support/collect"
)

type Sqlite struct {
	attributeCommands []string
	log               log.Log
	modifiers         []func(schema.Blueprint, schema.ColumnDefinition) string
	serials           []string
	tablePrefix       string
	wrap              *Wrap
}

func NewSqlite(log log.Log, tablePrefix string) *Sqlite {
	sqlite := &Sqlite{
		attributeCommands: []string{},
		log:               log,
		serials:           []string{"bigInteger", "integer", "mediumInteger", "smallInteger", "tinyInteger"},
		tablePrefix:       tablePrefix,
		wrap:              NewWrap(contractsdatabase.DriverSqlite, tablePrefix),
	}
	sqlite.modifiers = []func(schema.Blueprint, schema.ColumnDefinition) string{
		sqlite.ModifyCollation,
		sqlite.ModifyDefault,
		sqlite.ModifyIncrement,
		sqlite.ModifyNullable,
	}

	return sqlite
}

func (r *Sqlite) CompileAdd(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return "", err
	}
	column, err := r.getColumn(blueprint, command.Column)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("alter table %s add column %s", table, column), nil
}

func (r *Sqlite) CompileChange(_ schema.Blueprint, _ *schema.Command) (string, error) {
	// SQLite doesn't support ALTER COLUMN directly
	// Column changes require recreating the table
	return "", nil
}

func (r *Sqlite) CompileColumns(schema, table string) string {
	return fmt.Sprintf(
		`select name, type, not "notnull" as "nullable", dflt_value as "default", pk as "primary", hidden as "extra", `+
			`"" as collation `+
			"from pragma_table_xinfo(%s) order by cid asc",
		r.wrap.Quote(strings.ReplaceAll(table, ".", "__")))
}

func (r *Sqlite) CompileComment(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	return "", nil
}

func (r *Sqlite) CompileCreate(blueprint schema.Blueprint) (string, error) {
	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return "", err
	}
	columns, err := r.getColumns(blueprint)
	if err != nil {
		return "", err
	}
	foreignKeys, err := r.addForeignKeys(getCommandsByName(blueprint.GetCommands(), "foreign"))
	if err != nil {
		return "", err
	}
	primaryKeys, err := r.addPrimaryKeys(getCommandByName(blueprint.GetCommands(), "primary"))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("create table %s (%s%s%s)",
		table,
		strings.Join(columns, ", "),
		foreignKeys,
		primaryKeys), nil
}

func (r *Sqlite) CompileDisableWriteableSchema() string {
	return r.pragma("writable_schema", "0")
}

func (r *Sqlite) CompileDrop(blueprint schema.Blueprint) (string, error) {
	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("drop table %s", table), nil
}

func (r *Sqlite) CompileDropAllDomains(domains []string) (string, error) {
	return "", nil
}

func (r *Sqlite) CompileDropAllTables(tables []string) (string, error) {
	return "delete from sqlite_master where type in ('table', 'index', 'trigger')", nil
}

func (r *Sqlite) CompileDropAllTypes(types []string) (string, error) {
	return "", nil
}

func (r *Sqlite) CompileDropAllViews(views []string) (string, error) {
	return "delete from sqlite_master where type in ('view')", nil
}

func (r *Sqlite) CompileDropColumn(blueprint schema.Blueprint, command *schema.Command) ([]string, error) {
	// Requires SQLite 3.35+ for DROP COLUMN support
	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return nil, err
	}
	columns, err := r.wrap.Columns(command.Columns)
	if err != nil {
		return nil, err
	}
	prefixed := r.wrap.PrefixArray("drop column", columns)

	return collect.Map(prefixed, func(column string, _ int) string {
		return fmt.Sprintf("alter table %s %s", table, column)
	}), nil
}

func (r *Sqlite) CompileDropForeign(_ schema.Blueprint, _ *schema.Command) (string, error) {
	return "", nil
}

func (r *Sqlite) CompileDropFullText(_ schema.Blueprint, _ *schema.Command) (string, error) {
	return "", nil
}

func (r *Sqlite) CompileDropIfExists(blueprint schema.Blueprint) (string, error) {
	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("drop table if exists %s", table), nil
}

func (r *Sqlite) CompileDropIndex(_ schema.Blueprint, command *schema.Command) (string, error) {
	column, err := r.wrap.Column(command.Index)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("drop index %s", column), nil
}

func (r *Sqlite) CompileDropPrimary(_ schema.Blueprint, _ *schema.Command) (string, error) {
	return "", nil
}

func (r *Sqlite) CompileDropUnique(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	return r.CompileDropIndex(blueprint, command)
}

func (r *Sqlite) CompileEnableWriteableSchema() string {
	return r.pragma("writable_schema", "1")
}

func (r *Sqlite) CompileForeign(_ schema.Blueprint, _ *schema.Command) (string, error) {
	return "", nil
}

func (r *Sqlite) CompileForeignKeys(_, table string) string {
	return fmt.Sprintf(
		`SELECT 
			GROUP_CONCAT("from") AS columns, 
			"table" AS foreign_table, 
			GROUP_CONCAT("to") AS foreign_columns, 
			on_update, 
			on_delete 
		FROM (
			SELECT * FROM pragma_foreign_key_list(%s) 
			ORDER BY id DESC, seq
		) 
		GROUP BY id, "table", on_update, on_delete`,
		r.wrap.Quote(strings.ReplaceAll(table, ".", "__")),
	)
}

func (r *Sqlite) CompileFullText(_ schema.Blueprint, _ *schema.Command) (string, error) {
	return "", nil
}

func (r *Sqlite) CompileIndex(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	index, err := r.wrap.Column(command.Index)
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
	return fmt.Sprintf("create index %s on %s (%s)", index, table, columns), nil
}

func (r *Sqlite) CompileIndexes(_, table string) string {
	quotedTable := r.wrap.Quote(strings.ReplaceAll(table, ".", "__"))

	return fmt.Sprintf(
		`select 'primary' as name, group_concat(col) as columns, 1 as "unique", 1 as "primary" `+
			`from (select name as col from pragma_table_info(%s) where pk > 0 order by pk, cid) group by name `+
			`union select name, group_concat(col) as columns, "unique", origin = 'pk' as "primary" `+
			`from (select il.*, ii.name as col from pragma_index_list(%s) il, pragma_index_info(il.name) ii order by il.seq, ii.seqno) `+
			`group by name, "unique", "primary"`,
		quotedTable,
		r.wrap.Quote(table),
	)
}

func (r *Sqlite) CompilePrimary(_ schema.Blueprint, _ *schema.Command) (string, error) {
	return "", nil
}

func (r *Sqlite) CompileRebuild() string {
	return "vacuum"
}

func (r *Sqlite) CompileRename(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return "", err
	}
	to, err := r.wrap.Table(command.To)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("alter table %s rename to %s", table, to), nil
}

func (r *Sqlite) CompileRenameColumn(blueprint schema.Blueprint, command *schema.Command) (string, error) {
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

func (r *Sqlite) CompileRenameIndex(s schema.Schema, blueprint schema.Blueprint, command *schema.Command) ([]string, error) {
	indexes, err := s.GetIndexes(blueprint.GetTableName())
	if err != nil {
		r.log.Errorf("failed to get %s indexes: %v", blueprint.GetTableName(), err)
		return nil, err
	}

	indexes = collect.Filter(indexes, func(index schema.Index, _ int) bool {
		return strings.EqualFold(index.Name, command.From)
	})

	if len(indexes) == 0 {
		return nil, fmt.Errorf("index %s does not exist", command.From)
	}
	if indexes[0].Primary {
		return nil, fmt.Errorf("SQLite does not support altering primary keys")
	}
	if indexes[0].Unique {
		dropUnique, err := r.CompileDropUnique(blueprint, &schema.Command{
			Index: indexes[0].Name,
		})
		if err != nil {
			return nil, err
		}
		addUnique, err := r.CompileUnique(blueprint, &schema.Command{
			Index:   command.To,
			Columns: indexes[0].Columns,
		})
		if err != nil {
			return nil, err
		}
		return []string{dropUnique, addUnique}, nil
	}

	dropIndex, err := r.CompileDropIndex(blueprint, &schema.Command{
		Index: indexes[0].Name,
	})
	if err != nil {
		return nil, err
	}
	addIndex, err := r.CompileIndex(blueprint, &schema.Command{
		Index:   command.To,
		Columns: indexes[0].Columns,
	})
	if err != nil {
		return nil, err
	}
	return []string{dropIndex, addIndex}, nil
}

func (r *Sqlite) CompileUnique(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	return r.CompileIndex(blueprint, command)
}

func (r *Sqlite) CompileTypes() string {
	return ""
}

func (r *Sqlite) CompileViews(database string) string {
	return "select name, sql as definition from sqlite_master where type = 'view' order by name"
}

func (r *Sqlite) CompileTables(database string) string {
	return "select name from sqlite_master where type = 'table' and name not like 'sqlite_%' order by name"
}

func (r *Sqlite) GetAttributeCommands() []string {
	return r.attributeCommands
}

func (r *Sqlite) GetModifiers() []func(blueprint schema.Blueprint, column schema.ColumnDefinition) string {
	return r.modifiers
}

func (r *Sqlite) ModifyCollation(_ schema.Blueprint, column schema.ColumnDefinition) string {
	if collation := column.GetCollation(); collation != "" {
		return " collate " + collation
	}

	return ""
}

func (r *Sqlite) ModifyDefault(blueprint schema.Blueprint, column schema.ColumnDefinition) string {
	if column.GetDefault() != nil {
		return fmt.Sprintf(" default %s", getDefaultValue(column.GetDefault()))
	}

	return ""
}

func (r *Sqlite) ModifyNullable(blueprint schema.Blueprint, column schema.ColumnDefinition) string {
	if column.GetNullable() {
		return " null"
	} else {
		return " not null"
	}
}

func (r *Sqlite) ModifyIncrement(blueprint schema.Blueprint, column schema.ColumnDefinition) string {
	if slices.Contains(r.serials, column.GetType()) && column.GetAutoIncrement() {
		return " primary key autoincrement"
	}

	return ""
}

func (r *Sqlite) TypeBigInteger(column schema.ColumnDefinition) string {
	return "integer"
}

func (r *Sqlite) TypeBoolean(_ schema.ColumnDefinition) string {
	return "tinyint(1)"
}

func (r *Sqlite) TypeChar(column schema.ColumnDefinition) string {
	return "varchar"
}

func (r *Sqlite) TypeDate(column schema.ColumnDefinition) string {
	return "date"
}

func (r *Sqlite) TypeDateTime(column schema.ColumnDefinition) string {
	return r.TypeTimestamp(column)
}

func (r *Sqlite) TypeDateTimeTz(column schema.ColumnDefinition) string {
	return r.TypeDateTime(column)
}

func (r *Sqlite) TypeDecimal(column schema.ColumnDefinition) string {
	return "numeric"
}

func (r *Sqlite) TypeDouble(column schema.ColumnDefinition) string {
	return "double"
}

func (r *Sqlite) TypeEnum(column schema.ColumnDefinition) string {
	return fmt.Sprintf(`varchar check ("%s" in (%s))`, column.GetName(), strings.Join(r.wrap.Quotes(cast.ToStringSlice(column.GetAllowed())), ", "))
}

func (r *Sqlite) TypeFloat(column schema.ColumnDefinition) string {
	return "float"
}

func (r *Sqlite) TypeInteger(column schema.ColumnDefinition) string {
	return "integer"
}

func (r *Sqlite) TypeJson(column schema.ColumnDefinition) string {
	return "text"
}

func (r *Sqlite) TypeJsonb(column schema.ColumnDefinition) string {
	return "text"
}

func (r *Sqlite) TypeLongText(column schema.ColumnDefinition) string {
	return "text"
}

func (r *Sqlite) TypeMediumInteger(column schema.ColumnDefinition) string {
	return "integer"
}

func (r *Sqlite) TypeMediumText(column schema.ColumnDefinition) string {
	return "text"
}

func (r *Sqlite) TypeSmallInteger(column schema.ColumnDefinition) string {
	return "integer"
}

func (r *Sqlite) TypeString(column schema.ColumnDefinition) string {
	return "varchar"
}

func (r *Sqlite) TypeText(column schema.ColumnDefinition) string {
	return "text"
}

func (r *Sqlite) TypeTime(column schema.ColumnDefinition) string {
	return "time"
}

func (r *Sqlite) TypeTimeTz(column schema.ColumnDefinition) string {
	return r.TypeTime(column)
}

func (r *Sqlite) TypeTimestamp(column schema.ColumnDefinition) string {
	if column.GetUseCurrent() {
		column.Default(Expression("CURRENT_TIMESTAMP"))
	}

	return "datetime"
}

func (r *Sqlite) TypeTimestampTz(column schema.ColumnDefinition) string {
	return r.TypeTimestamp(column)
}

func (r *Sqlite) TypeTinyInteger(column schema.ColumnDefinition) string {
	return "integer"
}

func (r *Sqlite) TypeTinyText(column schema.ColumnDefinition) string {
	return "text"
}

func (r *Sqlite) addForeignKeys(commands []*schema.Command) (string, error) {
	var sql string

	for _, command := range commands {
		fk, err := r.getForeignKey(command)
		if err != nil {
			return "", err
		}
		sql += fk
	}

	return sql, nil
}

func (r *Sqlite) addPrimaryKeys(command *schema.Command) (string, error) {
	if command == nil {
		return "", nil
	}

	columns, err := r.wrap.Columnize(command.Columns)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(", primary key (%s)", columns), nil
}

func (r *Sqlite) getColumns(blueprint schema.Blueprint) ([]string, error) {
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

func (r *Sqlite) getColumn(blueprint schema.Blueprint, column schema.ColumnDefinition) (string, error) {
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

func (r *Sqlite) getForeignKey(command *schema.Command) (string, error) {
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

	sql := fmt.Sprintf(", foreign key(%s) references %s(%s)", columns, on, references)

	if command.OnDelete != "" && r.wrap.IsValidAction(command.OnDelete) {
		sql += " on delete " + command.OnDelete
	}
	if command.OnUpdate != "" && r.wrap.IsValidAction(command.OnUpdate) {
		sql += " on update " + command.OnUpdate
	}

	return sql, nil
}

func (r *Sqlite) pragma(name, value string) string {
	return fmt.Sprintf("pragma %s = %s", name, value)
}
