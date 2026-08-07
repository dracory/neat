package grammars_test

import (
	"strings"
	"testing"

	contractsschema "github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/database/schema/grammars"
)

func newMysqlGrammar() *grammars.Mysql {
	return grammars.NewMysql("")
}

func TestMysqlCompileCreateView(t *testing.T) {
	g := newMysqlGrammar()

	// Plain name
	sql, err := g.CompileCreateView(contractsschema.View{Name: "user_view", Definition: "select * from users"})
	if err != nil {
		t.Fatalf("CompileCreateView returned error: %v", err)
	}
	if !strings.HasPrefix(sql, "create view ") {
		t.Errorf("Expected 'create view' prefix, got: %s", sql)
	}
	if !strings.Contains(sql, "`user_view`") {
		t.Errorf("Expected backtick-quoted view name, got: %s", sql)
	}
	if !strings.Contains(sql, "as select * from users") {
		t.Errorf("Expected view definition, got: %s", sql)
	}

	// Schema-qualified name (database.table on MySQL)
	sql, err = g.CompileCreateView(contractsschema.View{Name: "mydb.my_view", Definition: "select 1"})
	if err != nil {
		t.Fatalf("CompileCreateView returned error for schema-qualified name: %v", err)
	}
	if !strings.Contains(sql, "`mydb`.`my_view`") {
		t.Errorf("Expected schema-quoted name, got: %s", sql)
	}
}

func TestMysqlCompileDropView(t *testing.T) {
	g := newMysqlGrammar()

	sql, err := g.CompileDropView("user_view")
	if err != nil {
		t.Fatalf("CompileDropView returned error: %v", err)
	}
	if !strings.Contains(sql, "drop view") {
		t.Errorf("Expected 'drop view', got: %s", sql)
	}
	if !strings.Contains(sql, "`user_view`") {
		t.Errorf("Expected backtick-quoted view name, got: %s", sql)
	}
}

func TestMysqlCompileDropViewIfExists(t *testing.T) {
	g := newMysqlGrammar()

	// Plain name
	sql, err := g.CompileDropViewIfExists("user_view")
	if err != nil {
		t.Fatalf("CompileDropViewIfExists returned error: %v", err)
	}
	if !strings.Contains(sql, "drop view if exists") {
		t.Errorf("Expected 'drop view if exists', got: %s", sql)
	}
	if !strings.Contains(sql, "`user_view`") {
		t.Errorf("Expected backtick-quoted view name, got: %s", sql)
	}

	// Schema-qualified name
	sql, err = g.CompileDropViewIfExists("mydb.my_view")
	if err != nil {
		t.Fatalf("CompileDropViewIfExists returned error for schema-qualified name: %v", err)
	}
	if !strings.Contains(sql, "`mydb`.`my_view`") {
		t.Errorf("Expected schema-quoted name, got: %s", sql)
	}
}

func TestMysqlCompileCreateViewInjection(t *testing.T) {
	g := newMysqlGrammar()

	_, err := g.CompileCreateView(contractsschema.View{Name: "users; DROP TABLE users; --", Definition: "select 1"})
	if err == nil {
		t.Error("Expected error for invalid identifier, got nil")
	}
}
