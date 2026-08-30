package grammars_test

import (
	"strings"
	"testing"

	contractsschema "github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/contracts/log"
	"github.com/dracory/neat/database/schema/grammars"
)

func newSqliteGrammar() *grammars.Sqlite {
	return grammars.NewSqlite(log.NewNoopLogger(), "")
}

func TestSqliteCompileColumns(t *testing.T) {
	g := newSqliteGrammar()
	sql := g.CompileColumns("", "users")

	if !strings.Contains(sql, "pragma_table_xinfo") {
		t.Errorf("Expected pragma_table_xinfo in CompileColumns, got: %s", sql)
	}

	if !strings.Contains(sql, "'' as collation") {
		t.Errorf("Expected single-quoted empty string '' as collation, got: %s", sql)
	}

	if strings.Contains(sql, `"" as collation`) {
		t.Errorf("CompileColumns must not use double-quoted empty string for collation, got: %s", sql)
	}
}

func TestSqliteCompileCreateView(t *testing.T) {
	g := newSqliteGrammar()

	// Plain name
	sql, err := g.CompileCreateView(contractsschema.View{Name: "user_view", Definition: "select * from users"})
	if err != nil {
		t.Fatalf("CompileCreateView returned error: %v", err)
	}
	if !strings.HasPrefix(sql, "create view ") {
		t.Errorf("Expected 'create view' prefix, got: %s", sql)
	}
	if !strings.Contains(sql, `"user_view"`) {
		t.Errorf("Expected quoted view name, got: %s", sql)
	}
	if !strings.Contains(sql, "as select * from users") {
		t.Errorf("Expected view definition, got: %s", sql)
	}

	// Schema-qualified name
	sql, err = g.CompileCreateView(contractsschema.View{Name: "main.my_view", Definition: "select 1"})
	if err != nil {
		t.Fatalf("CompileCreateView returned error for schema-qualified name: %v", err)
	}
	if !strings.Contains(sql, `"main"."my_view"`) {
		t.Errorf("Expected schema-quoted name, got: %s", sql)
	}
}

func TestSqliteCompileDropView(t *testing.T) {
	g := newSqliteGrammar()

	sql, err := g.CompileDropView("user_view")
	if err != nil {
		t.Fatalf("CompileDropView returned error: %v", err)
	}
	if !strings.Contains(sql, "drop view") {
		t.Errorf("Expected 'drop view', got: %s", sql)
	}
	if !strings.Contains(sql, `"user_view"`) {
		t.Errorf("Expected quoted view name, got: %s", sql)
	}
}

func TestSqliteCompileDropViewIfExists(t *testing.T) {
	g := newSqliteGrammar()

	// Plain name
	sql, err := g.CompileDropViewIfExists("user_view")
	if err != nil {
		t.Fatalf("CompileDropViewIfExists returned error: %v", err)
	}
	if !strings.Contains(sql, "drop view if exists") {
		t.Errorf("Expected 'drop view if exists', got: %s", sql)
	}
	if !strings.Contains(sql, `"user_view"`) {
		t.Errorf("Expected quoted view name, got: %s", sql)
	}

	// Schema-qualified name
	sql, err = g.CompileDropViewIfExists("main.my_view")
	if err != nil {
		t.Fatalf("CompileDropViewIfExists returned error for schema-qualified name: %v", err)
	}
	if !strings.Contains(sql, `"main"."my_view"`) {
		t.Errorf("Expected schema-quoted name, got: %s", sql)
	}
}

func TestSqliteCompileCreateViewInjection(t *testing.T) {
	g := newSqliteGrammar()

	_, err := g.CompileCreateView(contractsschema.View{Name: "users; DROP TABLE users; --", Definition: "select 1"})
	if err == nil {
		t.Error("Expected error for invalid identifier, got nil")
	}
}

func TestSqliteCompileUnique(t *testing.T) {
	g := newSqliteGrammar()
	bp := newBlueprint("users")

	sql, err := g.CompileUnique(bp, &contractsschema.Command{Index: "users_email_unique", Columns: []string{"email"}})
	if err != nil {
		t.Fatalf("CompileUnique returned error: %v", err)
	}
	if !strings.Contains(sql, "create unique index") {
		t.Errorf("Expected 'create unique index', got: %s", sql)
	}
}

func TestSqliteCompileIndex(t *testing.T) {
	g := newSqliteGrammar()
	bp := newBlueprint("users")

	sql, err := g.CompileIndex(bp, &contractsschema.Command{Index: "users_name_index", Columns: []string{"name"}})
	if err != nil {
		t.Fatalf("CompileIndex returned error: %v", err)
	}
	if !strings.Contains(sql, "create index") {
		t.Errorf("Expected 'create index', got: %s", sql)
	}
	if strings.Contains(sql, "unique") {
		t.Errorf("CompileIndex must not include 'unique', got: %s", sql)
	}
}
