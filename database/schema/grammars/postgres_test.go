package grammars_test

import (
	"strings"
	"testing"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	contractsschema "github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/database/schema/constants"
	"github.com/dracory/neat/database/schema/grammars"
)

func newPostgresGrammar() *grammars.Postgres {
	return grammars.NewPostgres("")
}

// stubColumn is a minimal schema.ColumnDefinition for use in grammar tests.
type stubColumn struct {
	name          string
	ttype         string
	autoIncrement bool
	nullable      bool
	length        int
	precision     int
	places        int
	total         int
	collation     string
	comment       string
	commentSet    bool
	def           any
	useCurrent    bool
	allowed       []any
}

func (c *stubColumn) After(_ string) contractsschema.ColumnDefinition { return c }
func (c *stubColumn) AutoIncrement() contractsschema.ColumnDefinition {
	c.autoIncrement = true
	return c
}
func (c *stubColumn) Change() contractsschema.ColumnDefinition            { return c }
func (c *stubColumn) Collation(v string) contractsschema.ColumnDefinition { c.collation = v; return c }
func (c *stubColumn) Comment(v string) contractsschema.ColumnDefinition {
	c.comment = v
	c.commentSet = true
	return c
}
func (c *stubColumn) Default(v any) contractsschema.ColumnDefinition { c.def = v; return c }
func (c *stubColumn) First() contractsschema.ColumnDefinition        { return c }
func (c *stubColumn) GetAfter() string                               { return "" }
func (c *stubColumn) GetAllowed() []any                              { return c.allowed }
func (c *stubColumn) GetAutoIncrement() bool                         { return c.autoIncrement }
func (c *stubColumn) GetChange() bool                                { return false }
func (c *stubColumn) GetCollation() string                           { return c.collation }
func (c *stubColumn) GetComment() string                             { return c.comment }
func (c *stubColumn) GetDefault() any                                { return c.def }
func (c *stubColumn) GetFirst() bool                                 { return false }
func (c *stubColumn) GetLength() int                                 { return c.length }
func (c *stubColumn) GetName() string                                { return c.name }
func (c *stubColumn) GetNullable() bool                              { return c.nullable }
func (c *stubColumn) GetOnUpdate() any                               { return nil }
func (c *stubColumn) GetPlaces() int {
	if c.places == 0 {
		return 2
	}
	return c.places
}
func (c *stubColumn) GetPrecision() int { return c.precision }
func (c *stubColumn) GetSrid() int      { return 0 }
func (c *stubColumn) GetTotal() int {
	if c.total == 0 {
		return 8
	}
	return c.total
}
func (c *stubColumn) GetType() string                                      { return c.ttype }
func (c *stubColumn) GetUnsigned() bool                                    { return false }
func (c *stubColumn) GetUseCurrent() bool                                  { return c.useCurrent }
func (c *stubColumn) GetUseCurrentOnUpdate() bool                          { return false }
func (c *stubColumn) IsSetComment() bool                                   { return c.commentSet }
func (c *stubColumn) Nullable() contractsschema.ColumnDefinition           { c.nullable = true; return c }
func (c *stubColumn) OnUpdate(_ any) contractsschema.ColumnDefinition      { return c }
func (c *stubColumn) Srid(_ int) contractsschema.ColumnDefinition          { return c }
func (c *stubColumn) Places(v int) contractsschema.ColumnDefinition        { c.places = v; return c }
func (c *stubColumn) Total(v int) contractsschema.ColumnDefinition         { c.total = v; return c }
func (c *stubColumn) Unsigned() contractsschema.ColumnDefinition           { return c }
func (c *stubColumn) UseCurrent() contractsschema.ColumnDefinition         { c.useCurrent = true; return c }
func (c *stubColumn) UseCurrentOnUpdate() contractsschema.ColumnDefinition { return c }

// stubBlueprint is a minimal schema.Blueprint stub for grammar tests.
type stubBlueprint struct {
	table   string
	columns []contractsschema.ColumnDefinition
}

func newBlueprint(table string) *stubBlueprint { return &stubBlueprint{table: table} }

func (b *stubBlueprint) col(name, ttype string) contractsschema.ColumnDefinition {
	c := &stubColumn{name: name, ttype: ttype}
	b.columns = append(b.columns, c)
	return c
}

func (b *stubBlueprint) BigIncrements(col string) contractsschema.ColumnDefinition {
	return b.col(col, "bigInteger")
}
func (b *stubBlueprint) BigInteger(col string) contractsschema.ColumnDefinition {
	return b.col(col, "bigInteger")
}
func (b *stubBlueprint) Boolean(col string) contractsschema.ColumnDefinition {
	return b.col(col, "boolean")
}
func (b *stubBlueprint) Build(_ contractsorm.Query, _ contractsschema.Grammar) error { return nil }
func (b *stubBlueprint) Char(col string, _ ...int) contractsschema.ColumnDefinition {
	return b.col(col, "char")
}
func (b *stubBlueprint) Column(col, t string) contractsschema.ColumnDefinition { return b.col(col, t) }
func (b *stubBlueprint) Create()                                               {}
func (b *stubBlueprint) Date(col string) contractsschema.ColumnDefinition      { return b.col(col, "date") }
func (b *stubBlueprint) DateTime(col string, _ ...int) contractsschema.ColumnDefinition {
	return b.col(col, "dateTime")
}
func (b *stubBlueprint) DateTimeTz(col string, _ ...int) contractsschema.ColumnDefinition {
	return b.col(col, "dateTimeTz")
}
func (b *stubBlueprint) Decimal(col string) contractsschema.ColumnDefinition {
	return b.col(col, "decimal")
}
func (b *stubBlueprint) Double(col string) contractsschema.ColumnDefinition {
	return b.col(col, "double")
}
func (b *stubBlueprint) Drop()                         {}
func (b *stubBlueprint) DropColumn(_ ...string)        {}
func (b *stubBlueprint) DropForeign(_ ...string)       {}
func (b *stubBlueprint) DropForeignByName(_ string)    {}
func (b *stubBlueprint) DropFullText(_ ...string)      {}
func (b *stubBlueprint) DropFullTextByName(_ string)   {}
func (b *stubBlueprint) DropIfExists()                 {}
func (b *stubBlueprint) DropIndex(_ ...string)         {}
func (b *stubBlueprint) DropIndexByName(_ string)      {}
func (b *stubBlueprint) DropPrimary(_ ...string)       {}
func (b *stubBlueprint) DropSoftDeletes(_ ...string)   {}
func (b *stubBlueprint) DropSoftDeletesTz(_ ...string) {}
func (b *stubBlueprint) DropTimestamps()               {}
func (b *stubBlueprint) DropTimestampsTz()             {}
func (b *stubBlueprint) DropUnique(_ ...string)        {}
func (b *stubBlueprint) DropUniqueByName(_ string)     {}
func (b *stubBlueprint) Enum(col string, allowed []any) contractsschema.ColumnDefinition {
	c := b.col(col, "enum")
	c.(*stubColumn).allowed = allowed
	return c
}
func (b *stubBlueprint) Float(col string, _ ...int) contractsschema.ColumnDefinition {
	return b.col(col, "float")
}
func (b *stubBlueprint) Foreign(_ ...string) contractsschema.ForeignKeyDefinition { return nil }
func (b *stubBlueprint) FullText(_ ...string) contractsschema.IndexDefinition     { return nil }
func (b *stubBlueprint) Geometry(col string) contractsschema.ColumnDefinition     { return b.col(col, "geometry") }
func (b *stubBlueprint) GeometryCollection(col string) contractsschema.ColumnDefinition {
	return b.col(col, "geometryCollection")
}
func (b *stubBlueprint) GetAddedColumns() []contractsschema.ColumnDefinition { return b.columns }
func (b *stubBlueprint) GetCommands() []*contractsschema.Command                  { return nil }
func (b *stubBlueprint) GetTableName() string                                     { return b.table }
func (b *stubBlueprint) HasCommand(_ string) bool                                 { return false }
func (b *stubBlueprint) ID(_ ...string) contractsschema.ColumnDefinition {
	return b.col("id", "bigInteger")
}
func (b *stubBlueprint) Increments(col string) contractsschema.ColumnDefinition {
	return b.col(col, "integer")
}
func (b *stubBlueprint) Index(_ ...string) contractsschema.IndexDefinition { return nil }
func (b *stubBlueprint) Integer(col string) contractsschema.ColumnDefinition {
	return b.col(col, "integer")
}
func (b *stubBlueprint) IntegerIncrements(col string) contractsschema.ColumnDefinition {
	return b.col(col, "integer")
}
func (b *stubBlueprint) Json(col string) contractsschema.ColumnDefinition { return b.col(col, "json") }
func (b *stubBlueprint) Jsonb(col string) contractsschema.ColumnDefinition {
	return b.col(col, "jsonb")
}
func (b *stubBlueprint) LineString(col string) contractsschema.ColumnDefinition {
	return b.col(col, "lineString")
}
func (b *stubBlueprint) LongText(col string) contractsschema.ColumnDefinition {
	return b.col(col, "longText")
}
func (b *stubBlueprint) MediumIncrements(col string) contractsschema.ColumnDefinition {
	return b.col(col, "mediumInteger")
}
func (b *stubBlueprint) MediumInteger(col string) contractsschema.ColumnDefinition {
	return b.col(col, "mediumInteger")
}
func (b *stubBlueprint) MediumText(col string) contractsschema.ColumnDefinition {
	return b.col(col, "mediumText")
}
func (b *stubBlueprint) MultiLineString(col string) contractsschema.ColumnDefinition {
	return b.col(col, "multiLineString")
}
func (b *stubBlueprint) MultiPoint(col string) contractsschema.ColumnDefinition {
	return b.col(col, "multiPoint")
}
func (b *stubBlueprint) MultiPolygon(col string) contractsschema.ColumnDefinition {
	return b.col(col, "multiPolygon")
}
func (b *stubBlueprint) Point(col string) contractsschema.ColumnDefinition   { return b.col(col, "point") }
func (b *stubBlueprint) Polygon(col string) contractsschema.ColumnDefinition { return b.col(col, "polygon") }
func (b *stubBlueprint) Primary(_ ...string)                                 {}
func (b *stubBlueprint) Rename(_ string)          {}
func (b *stubBlueprint) RenameColumn(_, _ string) {}
func (b *stubBlueprint) RenameIndex(_, _ string)  {}
func (b *stubBlueprint) SetTable(_ string)        {}
func (b *stubBlueprint) ShortID(_ ...string) contractsschema.ColumnDefinition {
	return b.col("id", "string")
}
func (b *stubBlueprint) SmallIncrements(col string) contractsschema.ColumnDefinition {
	return b.col(col, "smallInteger")
}
func (b *stubBlueprint) SmallInteger(col string) contractsschema.ColumnDefinition {
	return b.col(col, "smallInteger")
}
func (b *stubBlueprint) SoftDeletes(_ ...string) contractsschema.ColumnDefinition {
	return b.col(constants.DeletedAtColumnName, "timestamp")
}
func (b *stubBlueprint) SoftDeletesTz(_ ...string) contractsschema.ColumnDefinition {
	return b.col(constants.DeletedAtColumnName, "timestampTz")
}
func (b *stubBlueprint) String(col string, length ...int) contractsschema.ColumnDefinition {
	c := b.col(col, "string")
	if len(length) > 0 {
		c.(*stubColumn).length = length[0]
	}
	return c
}
func (b *stubBlueprint) Text(col string) contractsschema.ColumnDefinition { return b.col(col, "text") }
func (b *stubBlueprint) Time(col string, _ ...int) contractsschema.ColumnDefinition {
	return b.col(col, "time")
}
func (b *stubBlueprint) TimeTz(col string, _ ...int) contractsschema.ColumnDefinition {
	return b.col(col, "timeTz")
}
func (b *stubBlueprint) Timestamp(col string, _ ...int) contractsschema.ColumnDefinition {
	return b.col(col, "timestamp")
}
func (b *stubBlueprint) Timestamps(_ ...int)   {}
func (b *stubBlueprint) TimestampsTz(_ ...int) {}
func (b *stubBlueprint) TimestampTz(col string, _ ...int) contractsschema.ColumnDefinition {
	return b.col(col, "timestampTz")
}
func (b *stubBlueprint) TinyIncrements(col string) contractsschema.ColumnDefinition {
	return b.col(col, "tinyInteger")
}
func (b *stubBlueprint) TinyInteger(col string) contractsschema.ColumnDefinition {
	return b.col(col, "tinyInteger")
}
func (b *stubBlueprint) TinyText(col string) contractsschema.ColumnDefinition {
	return b.col(col, "tinyText")
}
func (b *stubBlueprint) ToSql(_ contractsschema.Grammar) ([]string, error)  { return nil, nil }
func (b *stubBlueprint) Unique(_ ...string) contractsschema.IndexDefinition { return nil }
func (b *stubBlueprint) UnsignedBigInteger(col string) contractsschema.ColumnDefinition {
	return b.col(col, "bigInteger")
}
func (b *stubBlueprint) UnsignedInteger(col string) contractsschema.ColumnDefinition {
	return b.col(col, "integer")
}
func (b *stubBlueprint) UnsignedMediumInteger(col string) contractsschema.ColumnDefinition {
	return b.col(col, "mediumInteger")
}
func (b *stubBlueprint) UnsignedSmallInteger(col string) contractsschema.ColumnDefinition {
	return b.col(col, "smallInteger")
}
func (b *stubBlueprint) UnsignedTinyInteger(col string) contractsschema.ColumnDefinition {
	return b.col(col, "tinyInteger")
}

func TestPostgresCompileChange(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("users")
	col := bp.String("nullable_col", 200).Nullable()

	sql, err := g.CompileChange(bp, &contractsschema.Command{Column: col})
	if err != nil {
		t.Fatalf("CompileChange returned error: %v", err)
	}
	// Column name must be double-quoted identifier, not single-quoted string literal
	if strings.Contains(sql, `'nullable_col'`) {
		t.Errorf("CompileChange must not use single-quoted column name, got: %s", sql)
	}
	if !strings.Contains(sql, `"nullable_col"`) {
		t.Errorf("CompileChange must use double-quoted column name, got: %s", sql)
	}
	// Type must be bare (varchar(200)), not full DDL repeating the column name
	if strings.Count(sql, `"nullable_col"`) > 3 {
		t.Errorf("Column name appears too many times — likely embedding full DDL as type: %s", sql)
	}
	if !strings.Contains(sql, "varchar(200)") {
		t.Errorf("Expected 'varchar(200)' as the type, got: %s", sql)
	}
	// Nullable change: nullable column should produce DROP NOT NULL
	if !strings.Contains(sql, "drop not null") {
		t.Errorf("Expected 'drop not null' for nullable column, got: %s", sql)
	}
}

func TestPostgresCompileChangeNotNullable(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("users")
	col := bp.String("required_col", 100)

	sql, err := g.CompileChange(bp, &contractsschema.Command{Column: col})
	if err != nil {
		t.Fatalf("CompileChange returned error: %v", err)
	}
	if !strings.Contains(sql, "set not null") {
		t.Errorf("Expected 'set not null' for non-nullable column, got: %s", sql)
	}
}

func TestPostgresCompileChangeWithDefault(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("users")
	col := bp.String("status", 50).Default("active")

	sql, err := g.CompileChange(bp, &contractsschema.Command{Column: col})
	if err != nil {
		t.Fatalf("CompileChange returned error: %v", err)
	}
	if !strings.Contains(sql, "set default") {
		t.Errorf("Expected 'set default' for column with default value, got: %s", sql)
	}
}

func TestPostgresCompileChangeDropDefault(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("users")
	col := bp.String("status", 50)

	sql, err := g.CompileChange(bp, &contractsschema.Command{Column: col})
	if err != nil {
		t.Fatalf("CompileChange returned error: %v", err)
	}
	if !strings.Contains(sql, "drop default") {
		t.Errorf("Expected 'drop default' for column without default, got: %s", sql)
	}
}

func TestPostgresCompileCreate(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("users")
	bp.String("name")

	sql, err := g.CompileCreate(bp)
	if err != nil {
		t.Fatalf("CompileCreate returned error: %v", err)
	}
	if !strings.Contains(sql, "create table") {
		t.Errorf("Expected 'create table', got: %s", sql)
	}
	if !strings.Contains(sql, `"users"`) {
		t.Errorf("Expected quoted table name, got: %s", sql)
	}
	if !strings.Contains(sql, `"name"`) {
		t.Errorf("Expected column 'name', got: %s", sql)
	}
}

func TestPostgresCompileDrop(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("users")

	sql, err := g.CompileDrop(bp)
	if err != nil {
		t.Fatalf("CompileDrop returned error: %v", err)
	}
	if !strings.Contains(sql, "drop table") {
		t.Errorf("Expected 'drop table', got: %s", sql)
	}
	if !strings.Contains(sql, `"users"`) {
		t.Errorf("Expected quoted table name, got: %s", sql)
	}
}

func TestPostgresCompileDropIfExists(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("users")

	sql, err := g.CompileDropIfExists(bp)
	if err != nil {
		t.Fatalf("CompileDropIfExists returned error: %v", err)
	}
	if !strings.Contains(sql, "drop table if exists") {
		t.Errorf("Expected 'drop table if exists', got: %s", sql)
	}
}

func TestPostgresCompileAdd(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("users")
	col := bp.String("email")

	sql, err := g.CompileAdd(bp, &contractsschema.Command{Column: col})
	if err != nil {
		t.Fatalf("CompileAdd returned error: %v", err)
	}
	if !strings.Contains(sql, "alter table") {
		t.Errorf("Expected 'alter table', got: %s", sql)
	}
	if !strings.Contains(sql, "add column") {
		t.Errorf("Expected 'add column', got: %s", sql)
	}
	if !strings.Contains(sql, `"email"`) {
		t.Errorf("Expected column 'email', got: %s", sql)
	}
}

func TestPostgresCompileDropColumn(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("users")

	sqls, err := g.CompileDropColumn(bp, &contractsschema.Command{Columns: []string{"email", "avatar"}})
	if err != nil {
		t.Fatalf("CompileDropColumn returned error: %v", err)
	}
	if len(sqls) == 0 {
		t.Fatal("Expected at least one SQL statement")
	}
	if !strings.Contains(sqls[0], "alter table") {
		t.Errorf("Expected 'alter table', got: %s", sqls[0])
	}
	if !strings.Contains(sqls[0], "drop column") {
		t.Errorf("Expected 'drop column', got: %s", sqls[0])
	}
}

func TestPostgresCompileRename(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("users")

	sql, err := g.CompileRename(bp, &contractsschema.Command{To: "members"})
	if err != nil {
		t.Fatalf("CompileRename returned error: %v", err)
	}
	if !strings.Contains(sql, "alter table") {
		t.Errorf("Expected 'alter table', got: %s", sql)
	}
	if !strings.Contains(sql, "rename to") {
		t.Errorf("Expected 'rename to', got: %s", sql)
	}
	if !strings.Contains(sql, `"members"`) {
		t.Errorf("Expected new table name 'members', got: %s", sql)
	}
}

func TestPostgresCompileRenameColumn(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("users")

	sql, err := g.CompileRenameColumn(bp, &contractsschema.Command{From: "name", To: "full_name"})
	if err != nil {
		t.Fatalf("CompileRenameColumn returned error: %v", err)
	}
	if !strings.Contains(sql, "rename column") {
		t.Errorf("Expected 'rename column', got: %s", sql)
	}
	if !strings.Contains(sql, `"name"`) {
		t.Errorf("Expected old column name, got: %s", sql)
	}
	if !strings.Contains(sql, `"full_name"`) {
		t.Errorf("Expected new column name, got: %s", sql)
	}
}

func TestPostgresCompilePrimary(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("users")

	sql, err := g.CompilePrimary(bp, &contractsschema.Command{Columns: []string{"id"}})
	if err != nil {
		t.Fatalf("CompilePrimary returned error: %v", err)
	}
	if !strings.Contains(sql, "add primary key") {
		t.Errorf("Expected 'add primary key', got: %s", sql)
	}
}

func TestPostgresCompileUnique(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("users")

	sql, err := g.CompileUnique(bp, &contractsschema.Command{Index: "users_email_unique", Columns: []string{"email"}})
	if err != nil {
		t.Fatalf("CompileUnique returned error: %v", err)
	}
	if !strings.Contains(sql, "add constraint") {
		t.Errorf("Expected 'add constraint', got: %s", sql)
	}
	if !strings.Contains(sql, "unique") {
		t.Errorf("Expected 'unique', got: %s", sql)
	}
}

func TestPostgresCompileIndex(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("users")

	sql, err := g.CompileIndex(bp, &contractsschema.Command{Index: "users_name_index", Columns: []string{"name"}})
	if err != nil {
		t.Fatalf("CompileIndex returned error: %v", err)
	}
	if !strings.Contains(sql, "create index") {
		t.Errorf("Expected 'create index', got: %s", sql)
	}
}

func TestPostgresCompileDropIndex(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("users")

	sql, err := g.CompileDropIndex(bp, &contractsschema.Command{Index: "users_name_index"})
	if err != nil {
		t.Fatalf("CompileDropIndex returned error: %v", err)
	}
	if !strings.Contains(sql, "drop index") {
		t.Errorf("Expected 'drop index', got: %s", sql)
	}
}

func TestPostgresCompileDropUnique(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("users")

	sql, err := g.CompileDropUnique(bp, &contractsschema.Command{Index: "users_email_unique"})
	if err != nil {
		t.Fatalf("CompileDropUnique returned error: %v", err)
	}
	if !strings.Contains(sql, "drop constraint") {
		t.Errorf("Expected 'drop constraint', got: %s", sql)
	}
}

func TestPostgresCompileDropPrimary(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("users")

	sql, err := g.CompileDropPrimary(bp, &contractsschema.Command{})
	if err != nil {
		t.Fatalf("CompileDropPrimary returned error: %v", err)
	}
	if !strings.Contains(sql, "drop constraint") {
		t.Errorf("Expected 'drop constraint', got: %s", sql)
	}
	if !strings.Contains(sql, "pkey") {
		t.Errorf("Expected '_pkey' suffix in constraint name, got: %s", sql)
	}
}

func TestPostgresCompileForeign(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("posts")

	sql, err := g.CompileForeign(bp, &contractsschema.Command{
		Index:      "posts_user_id_foreign",
		Columns:    []string{"user_id"},
		On:         "users",
		References: []string{"id"},
	})
	if err != nil {
		t.Fatalf("CompileForeign returned error: %v", err)
	}
	if !strings.Contains(sql, "foreign key") {
		t.Errorf("Expected 'foreign key', got: %s", sql)
	}
	if !strings.Contains(sql, "references") {
		t.Errorf("Expected 'references', got: %s", sql)
	}
}

func TestPostgresCompileDropForeign(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("posts")

	sql, err := g.CompileDropForeign(bp, &contractsschema.Command{Index: "posts_user_id_foreign"})
	if err != nil {
		t.Fatalf("CompileDropForeign returned error: %v", err)
	}
	if !strings.Contains(sql, "drop constraint") {
		t.Errorf("Expected 'drop constraint', got: %s", sql)
	}
}

func TestPostgresCompileDropAllTables(t *testing.T) {
	g := newPostgresGrammar()

	sql, err := g.CompileDropAllTables([]string{"users", "posts"})
	if err != nil {
		t.Fatalf("CompileDropAllTables returned error: %v", err)
	}
	if !strings.Contains(sql, "drop table") {
		t.Errorf("Expected 'drop table', got: %s", sql)
	}
	if !strings.Contains(sql, "cascade") {
		t.Errorf("Expected 'cascade', got: %s", sql)
	}
}

func TestPostgresCompileDropAllViews(t *testing.T) {
	g := newPostgresGrammar()

	sql, err := g.CompileDropAllViews([]string{"user_view"})
	if err != nil {
		t.Fatalf("CompileDropAllViews returned error: %v", err)
	}
	if !strings.Contains(sql, "drop view") {
		t.Errorf("Expected 'drop view', got: %s", sql)
	}
	if !strings.Contains(sql, "cascade") {
		t.Errorf("Expected 'cascade', got: %s", sql)
	}
}

func TestPostgresCompileDropAllDomains(t *testing.T) {
	g := newPostgresGrammar()

	sql, err := g.CompileDropAllDomains([]string{"email_domain"})
	if err != nil {
		t.Fatalf("CompileDropAllDomains returned error: %v", err)
	}
	if !strings.Contains(sql, "drop domain") {
		t.Errorf("Expected 'drop domain', got: %s", sql)
	}
}

func TestPostgresCompileDropAllTypes(t *testing.T) {
	g := newPostgresGrammar()

	sql, err := g.CompileDropAllTypes([]string{"mood"})
	if err != nil {
		t.Fatalf("CompileDropAllTypes returned error: %v", err)
	}
	if !strings.Contains(sql, "drop type") {
		t.Errorf("Expected 'drop type', got: %s", sql)
	}
}

func TestPostgresCompileTables(t *testing.T) {
	g := newPostgresGrammar()
	sql := g.CompileTables("")
	if !strings.Contains(sql, "pg_class") {
		t.Errorf("Expected pg_class in CompileTables, got: %s", sql)
	}
}

func TestPostgresCompileColumns(t *testing.T) {
	g := newPostgresGrammar()
	sql := g.CompileColumns("public", "users")
	if !strings.Contains(sql, "pg_attribute") {
		t.Errorf("Expected pg_attribute in CompileColumns, got: %s", sql)
	}
}

func TestPostgresCompileIndexes(t *testing.T) {
	g := newPostgresGrammar()
	sql := g.CompileIndexes("public", "users")
	if !strings.Contains(sql, "pg_index") {
		t.Errorf("Expected pg_index in CompileIndexes, got: %s", sql)
	}
}

func TestPostgresCompileForeignKeys(t *testing.T) {
	g := newPostgresGrammar()
	sql := g.CompileForeignKeys("public", "posts")
	if !strings.Contains(sql, "pg_constraint") {
		t.Errorf("Expected pg_constraint in CompileForeignKeys, got: %s", sql)
	}
}

func TestPostgresTypeBigIntegerAutoIncrement(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("users")
	col := bp.BigInteger("id").AutoIncrement()

	result := g.TypeBigInteger(col.(*stubColumn))
	if result != "bigserial" {
		t.Errorf("Expected 'bigserial' for auto-increment big integer, got: %s", result)
	}
}

func TestPostgresTypeBigIntegerNormal(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("users")
	col := bp.BigInteger("count")

	result := g.TypeBigInteger(col.(*stubColumn))
	if result != "bigint" {
		t.Errorf("Expected 'bigint', got: %s", result)
	}
}

func TestPostgresTypeIntegerAutoIncrement(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("users")
	col := bp.Integer("id").AutoIncrement()

	result := g.TypeInteger(col.(*stubColumn))
	if result != "serial" {
		t.Errorf("Expected 'serial' for auto-increment integer, got: %s", result)
	}
}

func TestPostgresTypeIntegerNormal(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("users")
	col := bp.Integer("age")

	result := g.TypeInteger(col.(*stubColumn))
	if result != "integer" {
		t.Errorf("Expected 'integer', got: %s", result)
	}
}

func TestPostgresTypeSmallIntegerAutoIncrement(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("users")
	col := bp.SmallInteger("id").AutoIncrement()

	result := g.TypeSmallInteger(col.(*stubColumn))
	if result != "smallserial" {
		t.Errorf("Expected 'smallserial', got: %s", result)
	}
}

func TestPostgresTypeString(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("users")
	col := bp.String("name", 100)

	result := g.TypeString(col.(*stubColumn))
	if result != "varchar(100)" {
		t.Errorf("Expected 'varchar(100)', got: %s", result)
	}
}

func TestPostgresTypeStringNoLength(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("users")
	col := bp.String("name")

	result := g.TypeString(col.(*stubColumn))
	if result != "varchar" && !strings.HasPrefix(result, "varchar(") {
		t.Errorf("Expected 'varchar' or 'varchar(N)', got: %s", result)
	}
}

func TestPostgresTypeText(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("posts")
	col := bp.Text("body")

	result := g.TypeText(col.(*stubColumn))
	if result != "text" {
		t.Errorf("Expected 'text', got: %s", result)
	}
}

func TestPostgresTypeLongText(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("posts")
	col := bp.LongText("body")

	result := g.TypeLongText(col.(*stubColumn))
	if result != "text" {
		t.Errorf("Expected 'text', got: %s", result)
	}
}

func TestPostgresTypeBoolean(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("users")
	col := bp.Boolean("active")

	result := g.TypeBoolean(col.(*stubColumn))
	if result != "boolean" {
		t.Errorf("Expected 'boolean', got: %s", result)
	}
}

func TestPostgresTypeJson(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("events")
	col := bp.Json("payload")

	result := g.TypeJson(col.(*stubColumn))
	if result != "json" {
		t.Errorf("Expected 'json', got: %s", result)
	}
}

func TestPostgresTypeJsonb(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("events")
	col := bp.Jsonb("payload")

	result := g.TypeJsonb(col.(*stubColumn))
	if result != "jsonb" {
		t.Errorf("Expected 'jsonb', got: %s", result)
	}
}

func TestPostgresTypeDecimal(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("orders")
	col := bp.Decimal("amount")

	result := g.TypeDecimal(col.(*stubColumn))
	if !strings.HasPrefix(result, "decimal(") {
		t.Errorf("Expected 'decimal(...)', got: %s", result)
	}
}

func TestPostgresTypeDouble(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("metrics")
	col := bp.Double("value")

	result := g.TypeDouble(col.(*stubColumn))
	if result != "double precision" {
		t.Errorf("Expected 'double precision', got: %s", result)
	}
}

func TestPostgresTypeDate(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("events")
	col := bp.Date("occurred_on")

	result := g.TypeDate(col.(*stubColumn))
	if result != "date" {
		t.Errorf("Expected 'date', got: %s", result)
	}
}

func TestPostgresTypeTimestamp(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("events")
	col := bp.Timestamp("created_at")

	result := g.TypeTimestamp(col.(*stubColumn))
	if !strings.Contains(result, "timestamp") {
		t.Errorf("Expected 'timestamp...', got: %s", result)
	}
	if !strings.Contains(result, "without time zone") {
		t.Errorf("Expected 'without time zone', got: %s", result)
	}
}

func TestPostgresTypeTimestampTz(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("events")
	col := bp.TimestampTz("created_at")

	result := g.TypeTimestampTz(col.(*stubColumn))
	if !strings.Contains(result, "with time zone") {
		t.Errorf("Expected 'with time zone', got: %s", result)
	}
}

func TestPostgresTypeTinyText(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("users")
	col := bp.TinyText("label")

	result := g.TypeTinyText(col.(*stubColumn))
	if result != "varchar(255)" {
		t.Errorf("Expected 'varchar(255)', got: %s", result)
	}
}

func TestPostgresModifyNullable(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("users")
	col := bp.String("name").Nullable()

	result := g.ModifyNullable(bp, col.(*stubColumn))
	if result != " null" {
		t.Errorf("Expected ' null', got: %q", result)
	}
}

func TestPostgresModifyNotNullable(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("users")
	col := bp.String("name")

	result := g.ModifyNullable(bp, col.(*stubColumn))
	if result != " not null" {
		t.Errorf("Expected ' not null', got: %q", result)
	}
}

func TestPostgresModifyDefaultString(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("users")
	col := bp.String("status").Default("active")

	result := g.ModifyDefault(bp, col.(*stubColumn))
	if !strings.Contains(result, "default") {
		t.Errorf("Expected 'default' modifier, got: %q", result)
	}
	if !strings.Contains(result, "active") {
		t.Errorf("Expected default value 'active', got: %q", result)
	}
}

func TestPostgresModifyDefaultNil(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("users")
	col := bp.String("name")

	result := g.ModifyDefault(bp, col.(*stubColumn))
	if result != "" {
		t.Errorf("Expected empty string for no default, got: %q", result)
	}
}

func TestPostgresModifyCollation(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("users")
	col := bp.String("name").Collation("en_US")

	result := g.ModifyCollation(bp, col.(*stubColumn))
	if !strings.Contains(result, "collate") {
		t.Errorf("Expected 'collate' in result, got: %q", result)
	}
	if !strings.Contains(result, "en_US") {
		t.Errorf("Expected collation value, got: %q", result)
	}
}

func TestPostgresModifyIncrementAddsKeyWhenNoPrimaryCommand(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("users")
	col := bp.Integer("id").AutoIncrement()

	result := g.ModifyIncrement(bp, col.(*stubColumn))
	if result != " primary key" {
		t.Errorf("Expected ' primary key', got: %q", result)
	}
}

func TestPostgresCompileRenameIndex(t *testing.T) {
	g := newPostgresGrammar()

	sqls, err := g.CompileRenameIndex(nil, nil, &contractsschema.Command{From: "old_idx", To: "new_idx"})
	if err != nil {
		t.Fatalf("CompileRenameIndex returned error: %v", err)
	}
	if len(sqls) == 0 {
		t.Fatal("Expected at least one SQL statement")
	}
	if !strings.Contains(sqls[0], "alter index") {
		t.Errorf("Expected 'alter index', got: %s", sqls[0])
	}
	if !strings.Contains(sqls[0], "rename to") {
		t.Errorf("Expected 'rename to', got: %s", sqls[0])
	}
}

func TestPostgresCompileFullText(t *testing.T) {
	g := newPostgresGrammar()
	bp := newBlueprint("posts")

	sql, err := g.CompileFullText(bp, &contractsschema.Command{
		Index:   "posts_body_fulltext",
		Columns: []string{"body"},
	})
	if err != nil {
		t.Fatalf("CompileFullText returned error: %v", err)
	}
	if !strings.Contains(sql, "create index") {
		t.Errorf("Expected 'create index', got: %s", sql)
	}
	if !strings.Contains(sql, "using gin") {
		t.Errorf("Expected 'using gin', got: %s", sql)
	}
	if !strings.Contains(sql, "tsvector") {
		t.Errorf("Expected 'tsvector', got: %s", sql)
	}
}

func TestPostgresCompileWithTablePrefix(t *testing.T) {
	g := grammars.NewPostgres("app_")
	bp := newBlueprint("app_users")
	bp.String("name")

	sql, err := g.CompileCreate(bp)
	if err != nil {
		t.Fatalf("CompileCreate with prefix returned error: %v", err)
	}
	if !strings.Contains(sql, "app_users") {
		t.Errorf("Expected table with prefix 'app_users', got: %s", sql)
	}
}

func TestPostgresGetAttributeCommands(t *testing.T) {
	g := newPostgresGrammar()
	cmds := g.GetAttributeCommands()
	if len(cmds) == 0 {
		t.Error("Expected at least one attribute command")
	}
	found := false
	for _, c := range cmds {
		if c == "comment" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected 'comment' in attribute commands, got: %v", cmds)
	}
}
