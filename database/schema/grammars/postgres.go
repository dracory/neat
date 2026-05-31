package grammars

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cast"

	contractsdatabase "github.com/dracory/neat/contracts/database"
	"github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/database/schema/constants"
)

type Postgres struct {
	attributeCommands []string
	modifiers         []func(schema.Blueprint, schema.ColumnDefinition) string
	serials           []string
	wrap              *Wrap
}

func NewPostgres(tablePrefix string) *Postgres {
	postgres := &Postgres{
		attributeCommands: []string{constants.CommandComment},
		serials:           []string{"bigInteger", "integer", "mediumInteger", "smallInteger", "tinyInteger"},
		wrap:              NewWrap(contractsdatabase.DriverPostgres, tablePrefix),
	}
	postgres.modifiers = []func(schema.Blueprint, schema.ColumnDefinition) string{
		postgres.ModifyCollation,
		postgres.ModifyDefault,
		postgres.ModifyIncrement,
		postgres.ModifyNullable,
	}

	return postgres
}

func (r *Postgres) CompileAdd(blueprint schema.Blueprint, command *schema.Command) (string, error) {
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

func (r *Postgres) CompileChange(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return "", err
	}
	columnName, err := r.wrap.Column(command.Column.GetName())
	if err != nil {
		return "", err
	}

	var statements []string

	// Type change
	statements = append(statements, fmt.Sprintf("alter table %s alter column %s type %s", table, columnName, getType(r, command.Column)))

	// Nullable change
	if command.Column.GetNullable() {
		statements = append(statements, fmt.Sprintf("alter table %s alter column %s drop not null", table, columnName))
	} else {
		statements = append(statements, fmt.Sprintf("alter table %s alter column %s set not null", table, columnName))
	}

	// Default change
	if command.Column.GetDefault() != nil {
		statements = append(statements, fmt.Sprintf("alter table %s alter column %s set default %s", table, columnName, getDefaultValue(command.Column.GetDefault())))
	} else {
		statements = append(statements, fmt.Sprintf("alter table %s alter column %s drop default", table, columnName))
	}

	return strings.Join(statements, "; "), nil
}

func (r *Postgres) CompileColumns(schema, table string) string {
	return fmt.Sprintf(
		"select a.attname as name, t.typname as type_name, format_type(a.atttypid, a.atttypmod) as type, "+
			"(select tc.collcollate from pg_catalog.pg_collation tc where tc.oid = a.attcollation) as collation, "+
			"not a.attnotnull as nullable, "+
			"(select pg_get_expr(adbin, adrelid) from pg_attrdef where c.oid = pg_attrdef.adrelid and pg_attrdef.adnum = a.attnum) as default, "+
			"col_description(c.oid, a.attnum) as comment "+
			"from pg_attribute a, pg_class c, pg_type t, pg_namespace n "+
			"where c.relname = %s and n.nspname = %s and a.attnum > 0 and a.attrelid = c.oid and a.atttypid = t.oid and n.oid = c.relnamespace "+
			"order by a.attnum", r.wrap.Quote(table), r.wrap.Quote(schema))
}

func (r *Postgres) CompileComment(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	comment := "NULL"
	if command.Column.IsSetComment() {
		comment = r.wrap.Quote(strings.ReplaceAll(command.Column.GetComment(), "'", "''"))
	}

	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return "", err
	}
	column, err := r.wrap.Column(command.Column.GetName())
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("comment on column %s.%s is %s", table, column, comment), nil
}

func (r *Postgres) CompileCreate(blueprint schema.Blueprint) (string, error) {
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

func (r *Postgres) CompileDrop(blueprint schema.Blueprint) (string, error) {
	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("drop table %s", table), nil
}

func (r *Postgres) CompileDropAllDomains(domains []string) (string, error) {
	escaped, err := r.EscapeNames(domains)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("drop domain %s cascade", strings.Join(escaped, ", ")), nil
}

func (r *Postgres) CompileDropAllTables(tables []string) (string, error) {
	columnized, err := r.wrap.Columnize(tables)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("drop table %s cascade", columnized), nil
}

func (r *Postgres) CompileDropAllTypes(types []string) (string, error) {
	escaped, err := r.EscapeNames(types)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("drop type %s cascade", strings.Join(escaped, ", ")), nil
}

func (r *Postgres) CompileDropAllViews(views []string) (string, error) {
	columnized, err := r.wrap.Columnize(views)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("drop view %s cascade", columnized), nil
}

func (r *Postgres) CompileDropColumn(blueprint schema.Blueprint, command *schema.Command) ([]string, error) {
	columns, err := r.wrap.Columns(command.Columns)
	if err != nil {
		return nil, err
	}
	prefixed := r.wrap.PrefixArray("drop column", columns)

	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return nil, err
	}

	return []string{
		fmt.Sprintf("alter table %s %s", table, strings.Join(prefixed, ", ")),
	}, nil
}

func (r *Postgres) CompileDropForeign(blueprint schema.Blueprint, command *schema.Command) (string, error) {
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

func (r *Postgres) CompileDropFullText(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	return r.CompileDropIndex(blueprint, command)
}

func (r *Postgres) CompileDropIfExists(blueprint schema.Blueprint) (string, error) {
	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("drop table if exists %s", table), nil
}

func (r *Postgres) CompileDropIndex(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	column, err := r.wrap.Column(command.Index)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("drop index %s", column), nil
}

func (r *Postgres) CompileDropPrimary(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	tableName := blueprint.GetTableName()
	index, err := r.wrap.Column(fmt.Sprintf("%s%s_pkey", r.wrap.GetPrefix(), tableName))
	if err != nil {
		return "", err
	}
	table, err := r.wrap.Table(tableName)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("alter table %s drop constraint %s", table, index), nil
}

func (r *Postgres) CompileDropUnique(blueprint schema.Blueprint, command *schema.Command) (string, error) {
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

func (r *Postgres) CompileForeign(blueprint schema.Blueprint, command *schema.Command) (string, error) {
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

func (r *Postgres) addSrid(ttype string, column schema.ColumnDefinition) string {
	if column.GetSrid() > 0 {
		if strings.Contains(ttype, "(") {
			return strings.Replace(ttype, ")", fmt.Sprintf(",%d)", column.GetSrid()), 1)
		}
		return fmt.Sprintf("%s(Geometry,%d)", ttype, column.GetSrid())
	}

	return ttype
}

func (r *Postgres) CompileForeignKeys(schema, table string) string {
	return fmt.Sprintf(
		`SELECT 
			c.conname AS name, 
			string_agg(la.attname, ',' ORDER BY conseq.ord) AS columns, 
			fn.nspname AS foreign_schema, 
			fc.relname AS foreign_table, 
			string_agg(fa.attname, ',' ORDER BY conseq.ord) AS foreign_columns, 
			c.confupdtype AS on_update, 
			c.confdeltype AS on_delete 
		FROM pg_constraint c 
		JOIN pg_class tc ON c.conrelid = tc.oid 
		JOIN pg_namespace tn ON tn.oid = tc.relnamespace 
		JOIN pg_class fc ON c.confrelid = fc.oid 
		JOIN pg_namespace fn ON fn.oid = fc.relnamespace 
		JOIN LATERAL unnest(c.conkey) WITH ORDINALITY AS conseq(num, ord) ON TRUE 
		JOIN pg_attribute la ON la.attrelid = c.conrelid AND la.attnum = conseq.num 
		JOIN pg_attribute fa ON fa.attrelid = c.confrelid AND fa.attnum = c.confkey[conseq.ord] 
		WHERE c.contype = 'f' AND tc.relname = %s AND tn.nspname = %s 
		GROUP BY c.conname, fn.nspname, fc.relname, c.confupdtype, c.confdeltype`,
		r.wrap.Quote(table),
		r.wrap.Quote(schema),
	)
}

func (r *Postgres) CompileFullText(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	language := "english"
	if command.Language != "" {
		language = command.Language
	}

	var columns []string
	for _, column := range command.Columns {
		col, err := r.wrap.Column(column)
		if err != nil {
			return "", err
		}
		columns = append(columns, fmt.Sprintf("to_tsvector(%s, %s)", r.wrap.Quote(language), col))
	}

	index, err := r.wrap.Column(command.Index)
	if err != nil {
		return "", err
	}
	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("create index %s on %s using gin(%s)", index, table, strings.Join(columns, " || ")), nil
}

func (r *Postgres) CompileIndex(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	var algorithm string
	if command.Algorithm != "" {
		algorithm = " using " + command.Algorithm
	}

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

	return fmt.Sprintf("create index %s on %s%s (%s)",
		index,
		table,
		algorithm,
		columns,
	), nil
}

func (r *Postgres) CompileIndexes(schema, table string) string {
	return fmt.Sprintf(
		"select ic.relname as name, string_agg(a.attname, ',' order by indseq.ord) as columns, "+
			"am.amname as \"type\", i.indisunique as \"unique\", i.indisprimary as \"primary\" "+
			"from pg_index i "+
			"join pg_class tc on tc.oid = i.indrelid "+
			"join pg_namespace tn on tn.oid = tc.relnamespace "+
			"join pg_class ic on ic.oid = i.indexrelid "+
			"join pg_am am on am.oid = ic.relam "+
			"join lateral unnest(i.indkey) with ordinality as indseq(num, ord) on true "+
			"left join pg_attribute a on a.attrelid = i.indrelid and a.attnum = indseq.num "+
			"where tc.relname = %s and tn.nspname = %s "+
			"group by ic.relname, am.amname, i.indisunique, i.indisprimary",
		r.wrap.Quote(table),
		r.wrap.Quote(schema),
	)
}

func (r *Postgres) CompilePrimary(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	table, err := r.wrap.Table(blueprint.GetTableName())
	if err != nil {
		return "", err
	}
	columns, err := r.wrap.Columnize(command.Columns)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("alter table %s add primary key (%s)", table, columns), nil
}

func (r *Postgres) CompileRename(blueprint schema.Blueprint, command *schema.Command) (string, error) {
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

func (r *Postgres) CompileRenameColumn(blueprint schema.Blueprint, command *schema.Command) (string, error) {
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

func (r *Postgres) CompileRenameIndex(_ schema.Schema, _ schema.Blueprint, command *schema.Command) ([]string, error) {
	from, err := r.wrap.Column(command.From)
	if err != nil {
		return nil, err
	}
	to, err := r.wrap.Column(command.To)
	if err != nil {
		return nil, err
	}
	return []string{
		fmt.Sprintf("alter index %s rename to %s", from, to),
	}, nil
}

func (r *Postgres) CompileTables(_ string) string {
	return "select c.relname as name, n.nspname as schema, pg_total_relation_size(c.oid) as size, " +
		"obj_description(c.oid, 'pg_class') as comment from pg_class c, pg_namespace n " +
		"where c.relkind in ('r', 'p') and n.oid = c.relnamespace and n.nspname not in ('pg_catalog', 'information_schema') " +
		"order by c.relname"
}

func (r *Postgres) CompileTypes() string {
	return `select t.typname as name, n.nspname as schema, t.typtype as type, t.typcategory as category, 
		((t.typinput = 'array_in'::regproc and t.typoutput = 'array_out'::regproc) or t.typtype = 'm') as implicit 
		from pg_type t 
		join pg_namespace n on n.oid = t.typnamespace 
		left join pg_class c on c.oid = t.typrelid 
		left join pg_type el on el.oid = t.typelem 
		left join pg_class ce on ce.oid = el.typrelid 
		where ((t.typrelid = 0 and (ce.relkind = 'c' or ce.relkind is null)) or c.relkind = 'c') 
		and not exists (select 1 from pg_depend d where d.objid in (t.oid, t.typelem) and d.deptype = 'e') 
		and n.nspname not in ('pg_catalog', 'information_schema')`
}

func (r *Postgres) CompileUnique(blueprint schema.Blueprint, command *schema.Command) (string, error) {
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

	sql := fmt.Sprintf("alter table %s add constraint %s unique (%s)", table, index, columns)

	if command.Deferrable != nil {
		if *command.Deferrable {
			sql += " deferrable"
		} else {
			sql += " not deferrable"
		}
	}
	if command.Deferrable != nil && command.InitiallyImmediate != nil {
		if *command.InitiallyImmediate {
			sql += " initially immediate"
		} else {
			sql += " initially deferred"
		}
	}

	return sql, nil
}

func (r *Postgres) CompileViews(database string) string {
	return "select viewname as name, schemaname as schema, definition from pg_views where schemaname not in ('pg_catalog', 'information_schema') order by viewname"
}

func (r *Postgres) EscapeNames(names []string) ([]string, error) {
	escapedNames := make([]string, 0, len(names))

	for _, name := range names {
		segments := strings.Split(name, ".")
		for i, segment := range segments {
			segment = strings.Trim(segment, `'"`)
			quoted, err := r.wrap.Value(segment)
			if err != nil {
				return nil, err
			}
			segments[i] = quoted
		}
		escapedName := strings.Join(segments, ".")
		escapedNames = append(escapedNames, escapedName)
	}

	return escapedNames, nil
}

func (r *Postgres) GetAttributeCommands() []string {
	return r.attributeCommands
}

func (r *Postgres) ModifyCollation(_ schema.Blueprint, column schema.ColumnDefinition) string {
	if collation := column.GetCollation(); collation != "" {
		return " collate " + r.wrap.Quote(collation)
	}

	return ""
}

func (r *Postgres) ModifyDefault(blueprint schema.Blueprint, column schema.ColumnDefinition) string {
	if column.GetDefault() != nil {
		return fmt.Sprintf(" default %s", getDefaultValue(column.GetDefault()))
	}

	return ""
}

func (r *Postgres) ModifyNullable(blueprint schema.Blueprint, column schema.ColumnDefinition) string {
	if column.GetNullable() {
		return " null"
	} else {
		return " not null"
	}
}

func (r *Postgres) ModifyIncrement(blueprint schema.Blueprint, column schema.ColumnDefinition) string {
	if !blueprint.HasCommand("primary") && slices.Contains(r.serials, column.GetType()) && column.GetAutoIncrement() {
		return " primary key"
	}

	return ""
}

func (r *Postgres) TypeBigInteger(column schema.ColumnDefinition) string {
	if column.GetAutoIncrement() {
		return "bigserial"
	}

	return "bigint"
}

func (r *Postgres) TypeBoolean(_ schema.ColumnDefinition) string {
	return "boolean"
}

func (r *Postgres) TypeChar(column schema.ColumnDefinition) string {
	length := column.GetLength()
	if length > 0 {
		return fmt.Sprintf("char(%d)", length)
	}

	return "char"
}

func (r *Postgres) TypeDate(column schema.ColumnDefinition) string {
	return "date"
}

func (r *Postgres) TypeDateTime(column schema.ColumnDefinition) string {
	return r.TypeTimestamp(column)
}

func (r *Postgres) TypeDateTimeTz(column schema.ColumnDefinition) string {
	return r.TypeTimestampTz(column)
}

func (r *Postgres) TypeDecimal(column schema.ColumnDefinition) string {
	return fmt.Sprintf("decimal(%d, %d)", column.GetTotal(), column.GetPlaces())
}

func (r *Postgres) TypeDouble(column schema.ColumnDefinition) string {
	return "double precision"
}

func (r *Postgres) TypeEnum(column schema.ColumnDefinition) string {
	return fmt.Sprintf(`varchar(255) check ("%s" in (%s))`, column.GetName(), strings.Join(r.wrap.Quotes(cast.ToStringSlice(column.GetAllowed())), ", "))
}

func (r *Postgres) TypeFloat(column schema.ColumnDefinition) string {
	precision := column.GetPrecision()
	if precision > 0 {
		return fmt.Sprintf("float(%d)", precision)
	}

	return "float"
}

func (r *Postgres) TypeGeometry(column schema.ColumnDefinition) string {
	return r.addSrid("geometry", column)
}

func (r *Postgres) TypeGeometryCollection(column schema.ColumnDefinition) string {
	return r.addSrid("geometry(GeometryCollection)", column)
}

func (r *Postgres) TypeInteger(column schema.ColumnDefinition) string {
	if column.GetAutoIncrement() {
		return "serial"
	}

	return "integer"
}

func (r *Postgres) TypeLineString(column schema.ColumnDefinition) string {
	return r.addSrid("geometry(LineString)", column)
}

func (r *Postgres) TypeJson(column schema.ColumnDefinition) string {
	return "json"
}

func (r *Postgres) TypeJsonb(column schema.ColumnDefinition) string {
	return "jsonb"
}

func (r *Postgres) TypeLongText(column schema.ColumnDefinition) string {
	return "text"
}

func (r *Postgres) TypeMediumInteger(column schema.ColumnDefinition) string {
	return r.TypeInteger(column)
}

func (r *Postgres) TypeMediumText(column schema.ColumnDefinition) string {
	return "text"
}

func (r *Postgres) TypeMultiLineString(_ schema.ColumnDefinition) string {
	return "multilinestring"
}

func (r *Postgres) TypeMultiPoint(_ schema.ColumnDefinition) string {
	return "multipoint"
}

func (r *Postgres) TypeMultiPolygon(_ schema.ColumnDefinition) string {
	return "multipolygon"
}

func (r *Postgres) TypePoint(_ schema.ColumnDefinition) string {
	return "point"
}

func (r *Postgres) TypePolygon(_ schema.ColumnDefinition) string {
	return "polygon"
}

func (r *Postgres) TypeSmallInteger(column schema.ColumnDefinition) string {
	if column.GetAutoIncrement() {
		return "smallserial"
	}

	return "smallint"
}

func (r *Postgres) TypeString(column schema.ColumnDefinition) string {
	length := column.GetLength()
	if length > 0 {
		return fmt.Sprintf("varchar(%d)", length)
	}

	return "varchar"
}

func (r *Postgres) TypeText(column schema.ColumnDefinition) string {
	return "text"
}

func (r *Postgres) TypeTime(column schema.ColumnDefinition) string {
	return fmt.Sprintf("time(%d) without time zone", column.GetPrecision())
}

func (r *Postgres) TypeTimeTz(column schema.ColumnDefinition) string {
	return fmt.Sprintf("time(%d) with time zone", column.GetPrecision())
}

func (r *Postgres) TypeTimestamp(column schema.ColumnDefinition) string {
	if column.GetUseCurrent() {
		column.Default(Expression("CURRENT_TIMESTAMP"))
	}

	return fmt.Sprintf("timestamp(%d) without time zone", column.GetPrecision())
}

func (r *Postgres) TypeTimestampTz(column schema.ColumnDefinition) string {
	if column.GetUseCurrent() {
		column.Default(Expression("CURRENT_TIMESTAMP"))
	}

	return fmt.Sprintf("timestamp(%d) with time zone", column.GetPrecision())
}

func (r *Postgres) TypeTinyInteger(column schema.ColumnDefinition) string {
	return r.TypeSmallInteger(column)
}

func (r *Postgres) TypeTinyText(column schema.ColumnDefinition) string {
	return "varchar(255)"
}

func (r *Postgres) getColumns(blueprint schema.Blueprint) ([]string, error) {
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

func (r *Postgres) getColumn(blueprint schema.Blueprint, column schema.ColumnDefinition) (string, error) {
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
