package processors

import (
	"testing"

	"github.com/dracory/neat/contracts/database/schema"
)

// ---- helpers ----

func strPtr(s string) *string { return &s }

// ---- utils / processIndexes ----

func TestProcessIndexesEmpty(t *testing.T) {
	result := processIndexes(nil)
	if len(result) != 0 {
		t.Errorf("expected empty, got %v", result)
	}
}

func TestProcessIndexes(t *testing.T) {
	input := []schema.DBIndex{
		{Columns: "id", Name: "PRIMARY", Type: "BTREE", Primary: true, Unique: true},
		{Columns: "email", Name: "idx_email", Type: "HASH", Primary: false, Unique: true},
	}
	got := processIndexes(input)
	if len(got) != 2 {
		t.Fatalf("expected 2, got %d", len(got))
	}
	if got[0].Name != "primary" || !got[0].Primary || !got[0].Unique {
		t.Errorf("unexpected index[0]: %+v", got[0])
	}
	if got[1].Name != "idx_email" || got[1].Primary || !got[1].Unique {
		t.Errorf("unexpected index[1]: %+v", got[1])
	}
	if got[1].Type != "hash" {
		t.Errorf("Type should be lower-cased, got %q", got[1].Type)
	}
}

// ---- MySQL ----

func TestMysqlProcessColumns(t *testing.T) {
	collation := "utf8_general_ci"
	comment := "user name"
	def := "anonymous"
	input := []schema.DBColumn{
		{
			Name:      "name",
			Type:      "varchar(255)",
			TypeName:  "varchar",
			Nullable:  "YES",
			Extra:     "",
			Collation: &collation,
			Comment:   &comment,
			Default:   &def,
		},
		{
			Name:     "id",
			Type:     "int",
			TypeName: "int",
			Nullable: "NO",
			Extra:    "auto_increment",
		},
	}
	p := NewMysql()
	got := p.ProcessColumns(input)
	if len(got) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(got))
	}
	if !got[0].Nullable {
		t.Error("name column should be nullable")
	}
	if got[0].Collation != collation {
		t.Errorf("collation = %q, want %q", got[0].Collation, collation)
	}
	if got[0].Comment != comment {
		t.Errorf("comment = %q, want %q", got[0].Comment, comment)
	}
	if got[0].Default != def {
		t.Errorf("default = %q, want %q", got[0].Default, def)
	}
	if got[1].Autoincrement != true {
		t.Error("id column should have auto_increment")
	}
	if got[1].Nullable {
		t.Error("id column should not be nullable")
	}
}

func TestMysqlProcessColumnsNilPointers(t *testing.T) {
	input := []schema.DBColumn{{Name: "x", Type: "int", Nullable: "NO"}}
	p := NewMysql()
	got := p.ProcessColumns(input)
	if len(got) != 1 {
		t.Fatalf("expected 1 column")
	}
	if got[0].Collation != "" || got[0].Comment != "" {
		t.Errorf("nil pointer fields should yield empty strings: %+v", got[0])
	}
}

func TestMysqlProcessForeignKeys(t *testing.T) {
	input := []schema.DBForeignKey{
		{
			Name:           "fk_user",
			Columns:        "user_id",
			ForeignSchema:  "public",
			ForeignTable:   "users",
			ForeignColumns: "id",
			OnUpdate:       "CASCADE",
			OnDelete:       "SET NULL",
		},
	}
	p := NewMysql()
	got := p.ProcessForeignKeys(input)
	if len(got) != 1 {
		t.Fatalf("expected 1 fk")
	}
	if got[0].OnUpdate != "cascade" || got[0].OnDelete != "set null" {
		t.Errorf("actions should be lower-cased: %+v", got[0])
	}
	if got[0].ForeignTable != "users" {
		t.Errorf("ForeignTable = %q", got[0].ForeignTable)
	}
}

func TestMysqlProcessIndexes(t *testing.T) {
	input := []schema.DBIndex{
		{Columns: "id", Name: "PRIMARY", Type: "BTREE", Primary: false, Unique: false},
	}
	p := NewMysql()
	got := p.ProcessIndexes(input)
	if len(got) != 1 {
		t.Fatalf("expected 1 index")
	}
	if got[0].Name != "primary" {
		t.Errorf("index name = %q, want primary", got[0].Name)
	}
	if !got[0].Primary {
		t.Error("index named 'primary' should have Primary=true")
	}
}

func TestMysqlProcessTables(t *testing.T) {
	input := []schema.Table{{Name: "users"}, {Name: "orders"}}
	p := NewMysql()
	got := p.ProcessTables(input)
	if len(got) != 2 {
		t.Errorf("ProcessTables should return input unchanged, got %d", len(got))
	}
}

// ---- Postgres ----

func TestPostgresProcessColumnsAutoincrement(t *testing.T) {
	seqDefault := "nextval('users_id_seq'::regclass)"
	otherDefault := "0"
	input := []schema.DBColumn{
		{Name: "id", Type: "integer", TypeName: "int4", Default: &seqDefault},
		{Name: "age", Type: "integer", TypeName: "int4", Default: &otherDefault},
	}
	p := NewPostgres()
	got := p.ProcessColumns(input)
	if !got[0].Autoincrement {
		t.Error("id column with nextval default should be autoincrement")
	}
	if got[1].Autoincrement {
		t.Error("age column should not be autoincrement")
	}
}

func TestPostgresProcessColumnsNilPointers(t *testing.T) {
	input := []schema.DBColumn{{Name: "x", Type: "int"}}
	p := NewPostgres()
	got := p.ProcessColumns(input)
	if got[0].Collation != "" || got[0].Comment != "" {
		t.Errorf("nil pointers should yield empty strings: %+v", got[0])
	}
}

func TestPostgresProcessForeignKeysShortCodes(t *testing.T) {
	input := []schema.DBForeignKey{
		{Name: "fk1", Columns: "a", ForeignTable: "b", ForeignColumns: "id", OnUpdate: "a", OnDelete: "c"},
		{Name: "fk2", Columns: "x", ForeignTable: "y", ForeignColumns: "id", OnUpdate: "r", OnDelete: "n"},
		{Name: "fk3", Columns: "x", ForeignTable: "y", ForeignColumns: "id", OnUpdate: "UNKNOWN", OnDelete: "UNKNOWN"},
	}
	p := NewPostgres()
	got := p.ProcessForeignKeys(input)
	if got[0].OnUpdate != "no action" || got[0].OnDelete != "cascade" {
		t.Errorf("short codes a/c: %+v", got[0])
	}
	if got[1].OnUpdate != "restrict" || got[1].OnDelete != "set null" {
		t.Errorf("short codes r/n: %+v", got[1])
	}
	if got[2].OnUpdate != "unknown" || got[2].OnDelete != "unknown" {
		t.Errorf("unknown codes should fall through to lower-case: %+v", got[2])
	}
}

func TestPostgresProcessIndexes(t *testing.T) {
	input := []schema.DBIndex{{Columns: "email", Name: "IDX_EMAIL", Type: "btree", Unique: true}}
	p := NewPostgres()
	got := p.ProcessIndexes(input)
	if len(got) != 1 || got[0].Name != "idx_email" {
		t.Errorf("unexpected: %+v", got)
	}
}

func TestPostgresProcessTypes(t *testing.T) {
	input := []schema.Type{
		{Type: "b", Category: "n"},
		{Type: "e", Category: "e"},
		{Type: "z", Category: "z"},
	}
	p := NewPostgres()
	got := p.ProcessTypes(input)
	if got[0].Type != "base" || got[0].Category != "numeric" {
		t.Errorf("type b/n: %+v", got[0])
	}
	if got[1].Type != "enum" || got[1].Category != "enum" {
		t.Errorf("type e/e: %+v", got[1])
	}
	if got[2].Type != "" || got[2].Category != "internal_use" {
		t.Errorf("unknown type code should map to empty, z-cat to internal_use: %+v", got[2])
	}
}

func TestPostgresProcessTables(t *testing.T) {
	input := []schema.Table{{Name: "t1"}}
	p := NewPostgres()
	if got := p.ProcessTables(input); len(got) != 1 {
		t.Error("ProcessTables should return input unchanged")
	}
}

// ---- SQLite ----

func TestSqliteProcessColumnsAutoincrement(t *testing.T) {
	input := []schema.DBColumn{
		{Name: "id", Type: "INTEGER", Primary: true},
		{Name: "name", Type: "TEXT"},
	}
	p := NewSqlite()
	got := p.ProcessColumns(input)
	if !got[0].Autoincrement {
		t.Error("single INTEGER primary key should be autoincrement")
	}
	if got[1].Autoincrement {
		t.Error("TEXT column should not be autoincrement")
	}
	if got[0].Type != "integer" {
		t.Errorf("type should be lower-cased, got %q", got[0].Type)
	}
}

func TestSqliteProcessColumnsMultiplePKNoAutoincrement(t *testing.T) {
	input := []schema.DBColumn{
		{Name: "a", Type: "INTEGER", Primary: true},
		{Name: "b", Type: "INTEGER", Primary: true},
	}
	p := NewSqlite()
	got := p.ProcessColumns(input)
	for _, col := range got {
		if col.Autoincrement {
			t.Errorf("composite PK column %q should not be autoincrement", col.Name)
		}
	}
}

func TestSqliteProcessColumnsTypeNameParsing(t *testing.T) {
	input := []schema.DBColumn{
		{Name: "x", Type: "varchar(255)"},
		{Name: "y", Type: "int"},
	}
	p := NewSqlite()
	got := p.ProcessColumns(input)
	if got[0].TypeName != "varchar" {
		t.Errorf("TypeName from varchar(255) = %q, want varchar", got[0].TypeName)
	}
	if got[1].TypeName != "int" {
		t.Errorf("TypeName from int = %q, want int", got[1].TypeName)
	}
}

func TestSqliteProcessForeignKeys(t *testing.T) {
	input := []schema.DBForeignKey{
		{Name: "fk", Columns: "uid", ForeignTable: "users", ForeignColumns: "id", OnUpdate: "CASCADE", OnDelete: "SET NULL"},
	}
	p := NewSqlite()
	got := p.ProcessForeignKeys(input)
	if len(got) != 1 {
		t.Fatalf("expected 1")
	}
	if got[0].OnUpdate != "cascade" || got[0].OnDelete != "set null" {
		t.Errorf("actions not lower-cased: %+v", got[0])
	}
}

func TestSqliteProcessIndexesSinglePrimary(t *testing.T) {
	input := []schema.DBIndex{
		{Columns: "id", Name: "pk", Primary: true, Unique: true},
		{Columns: "email", Name: "idx_email", Unique: true},
	}
	p := NewSqlite()
	got := p.ProcessIndexes(input)
	if len(got) != 2 {
		t.Errorf("expected 2 indexes for single PK, got %d", len(got))
	}
}

func TestSqliteProcessIndexesMultiplePrimaryFiltered(t *testing.T) {
	input := []schema.DBIndex{
		{Columns: "a", Name: "pk_a", Primary: true},
		{Columns: "b", Name: "pk_b", Primary: true},
		{Columns: "email", Name: "idx_email"},
	}
	p := NewSqlite()
	got := p.ProcessIndexes(input)
	for _, idx := range got {
		if idx.Primary {
			t.Errorf("primary indexes should be filtered out when count>1: %+v", idx)
		}
	}
	if len(got) != 1 || got[0].Name != "idx_email" {
		t.Errorf("only non-primary index should remain: %+v", got)
	}
}

func TestSqliteProcessTables(t *testing.T) {
	input := []schema.Table{{Name: "t"}}
	p := NewSqlite()
	if got := p.ProcessTables(input); len(got) != 1 {
		t.Error("ProcessTables should return input unchanged")
	}
}

// ---- SQL Server ----

func TestSqlserverGetTypeVariableLength(t *testing.T) {
	cases := []struct {
		col  schema.DBColumn
		want string
	}{
		{schema.DBColumn{TypeName: "varchar", Length: 255}, "varchar(255)"},
		{schema.DBColumn{TypeName: "nvarchar", Length: -1}, "nvarchar(max)"},
		{schema.DBColumn{TypeName: "char", Length: 10}, "char(10)"},
		{schema.DBColumn{TypeName: "varbinary", Length: -1}, "varbinary(max)"},
		{schema.DBColumn{TypeName: "decimal", Precision: 18, Places: 2}, "decimal(18,2)"},
		{schema.DBColumn{TypeName: "numeric", Precision: 10, Places: 4}, "numeric(10,4)"},
		{schema.DBColumn{TypeName: "float", Precision: 53}, "float(53)"},
		{schema.DBColumn{TypeName: "datetime2", Precision: 7}, "datetime2(7)"},
		{schema.DBColumn{TypeName: "int"}, "int"},
		{schema.DBColumn{TypeName: "bigint"}, "bigint"},
	}
	for _, c := range cases {
		got := getType(c.col)
		if got != c.want {
			t.Errorf("getType(%+v) = %q, want %q", c.col, got, c.want)
		}
	}
}

func TestSqlserverProcessColumns(t *testing.T) {
	collation := "SQL_Latin1_General_CP1_CI_AS"
	comment := "primary key"
	def := "1"
	input := []schema.DBColumn{
		{
			Name:          "id",
			TypeName:      "int",
			Autoincrement: true,
			Collation:     &collation,
			Comment:       &comment,
			Default:       &def,
		},
		{Name: "name", TypeName: "varchar", Length: 100},
	}
	p := NewSqlserver()
	got := p.ProcessColumns(input)
	if !got[0].Autoincrement {
		t.Error("id should be autoincrement")
	}
	if got[0].Collation != collation {
		t.Errorf("collation = %q", got[0].Collation)
	}
	if got[0].Comment != comment {
		t.Errorf("comment = %q", got[0].Comment)
	}
	if got[1].Type != "varchar(100)" {
		t.Errorf("type = %q, want varchar(100)", got[1].Type)
	}
}

func TestSqlserverProcessColumnsNilPointers(t *testing.T) {
	input := []schema.DBColumn{{Name: "x", TypeName: "int"}}
	p := NewSqlserver()
	got := p.ProcessColumns(input)
	if got[0].Collation != "" || got[0].Comment != "" {
		t.Errorf("nil pointers should yield empty strings: %+v", got[0])
	}
}

func TestSqlserverProcessForeignKeys(t *testing.T) {
	input := []schema.DBForeignKey{
		{Name: "fk", Columns: "uid", ForeignTable: "users", ForeignColumns: "id", OnUpdate: "NO_ACTION", OnDelete: "SET_NULL"},
	}
	p := NewSqlserver()
	got := p.ProcessForeignKeys(input)
	if got[0].OnUpdate != "no action" || got[0].OnDelete != "set null" {
		t.Errorf("underscores should be replaced and lower-cased: %+v", got[0])
	}
}

func TestSqlserverProcessIndexes(t *testing.T) {
	input := []schema.DBIndex{{Columns: "id", Name: "PK_USERS", Type: "CLUSTERED", Primary: true, Unique: true}}
	p := NewSqlserver()
	got := p.ProcessIndexes(input)
	if len(got) != 1 || got[0].Name != "pk_users" {
		t.Errorf("unexpected: %+v", got)
	}
}

func TestSqlserverProcessTables(t *testing.T) {
	input := []schema.Table{{Name: "dbo.users"}}
	p := NewSqlserver()
	if got := p.ProcessTables(input); len(got) != 1 {
		t.Error("ProcessTables should return input unchanged")
	}
}
