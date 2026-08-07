package grammars_test

import (
	"strings"
	"testing"

	contractsschema "github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/database/schema/grammars"
)

func newSqlserverGrammar() *grammars.Sqlserver {
	return grammars.NewSqlserver("")
}

func TestSqlserverCompileCreateView(t *testing.T) {
	g := newSqlserverGrammar()

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
	sql, err = g.CompileCreateView(contractsschema.View{Name: "dbo.my_view", Definition: "select 1"})
	if err != nil {
		t.Fatalf("CompileCreateView returned error for schema-qualified name: %v", err)
	}
	if !strings.Contains(sql, `"dbo"."my_view"`) {
		t.Errorf("Expected schema-quoted name, got: %s", sql)
	}
}

func TestSqlserverCompileDropView(t *testing.T) {
	g := newSqlserverGrammar()

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

func TestSqlserverCompileDropViewIfExists(t *testing.T) {
	g := newSqlserverGrammar()

	// Plain name
	sql, err := g.CompileDropViewIfExists("user_view")
	if err != nil {
		t.Fatalf("CompileDropViewIfExists returned error: %v", err)
	}
	if !strings.Contains(sql, "if object_id(") {
		t.Errorf("Expected 'if object_id(' guard, got: %s", sql)
	}
	if !strings.Contains(sql, "'V'") {
		t.Errorf("Expected 'V' type filter, got: %s", sql)
	}
	if !strings.Contains(sql, "drop view") {
		t.Errorf("Expected 'drop view', got: %s", sql)
	}

	// Schema-qualified name
	sql, err = g.CompileDropViewIfExists("dbo.my_view")
	if err != nil {
		t.Fatalf("CompileDropViewIfExists returned error for schema-qualified name: %v", err)
	}
	if !strings.Contains(sql, `"dbo"."my_view"`) {
		t.Errorf("Expected schema-quoted name, got: %s", sql)
	}
}

func TestSqlserverCompileCreateViewInjection(t *testing.T) {
	g := newSqlserverGrammar()

	_, err := g.CompileCreateView(contractsschema.View{Name: "users; DROP TABLE users; --", Definition: "select 1"})
	if err == nil {
		t.Error("Expected error for invalid identifier, got nil")
	}
}
