package grammars_test

import (
	"strings"
	"testing"

	contractsschema "github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/database/schema/grammars"
)

func newOracleGrammar() *grammars.Oracle {
	return grammars.NewOracle("")
}

func TestOracleCompileCreateView(t *testing.T) {
	g := newOracleGrammar()

	// Plain name — Oracle uppercases unquoted identifiers in Wrap.Value
	sql, err := g.CompileCreateView(contractsschema.View{Name: "user_view", Definition: "select * from users"})
	if err != nil {
		t.Fatalf("CompileCreateView returned error: %v", err)
	}
	if !strings.HasPrefix(sql, "create or replace force view ") {
		t.Errorf("Expected 'create or replace force view' prefix, got: %s", sql)
	}
	if !strings.Contains(sql, "as select * from users") {
		t.Errorf("Expected view definition, got: %s", sql)
	}

	// Schema-qualified name
	sql, err = g.CompileCreateView(contractsschema.View{Name: "scott.my_view", Definition: "select 1"})
	if err != nil {
		t.Fatalf("CompileCreateView returned error for schema-qualified name: %v", err)
	}
	if !strings.Contains(sql, `"SCOTT"."MY_VIEW"`) {
		t.Errorf("Expected schema-quoted uppercased name, got: %s", sql)
	}
}

func TestOracleCompileDropView(t *testing.T) {
	g := newOracleGrammar()

	sql, err := g.CompileDropView("user_view")
	if err != nil {
		t.Fatalf("CompileDropView returned error: %v", err)
	}
	if !strings.Contains(sql, "drop view") {
		t.Errorf("Expected 'drop view', got: %s", sql)
	}
}

func TestOracleCompileDropViewIfExists(t *testing.T) {
	g := newOracleGrammar()

	// Plain name
	sql, err := g.CompileDropViewIfExists("user_view")
	if err != nil {
		t.Fatalf("CompileDropViewIfExists returned error: %v", err)
	}
	if !strings.Contains(sql, "BEGIN EXECUTE IMMEDIATE") {
		t.Errorf("Expected PL/SQL block, got: %s", sql)
	}
	if !strings.Contains(sql, "DROP VIEW USER_VIEW") {
		t.Errorf("Expected 'DROP VIEW USER_VIEW' inside block, got: %s", sql)
	}
	if !strings.Contains(sql, "SQLCODE != -942") {
		t.Errorf("Expected SQLCODE -942 guard, got: %s", sql)
	}
	if !strings.HasSuffix(sql, "END;") {
		t.Errorf("Expected to end with 'END;', got: %s", sql)
	}

	// Schema-qualified name — still validated by wrap.Table, uppercased in output
	sql, err = g.CompileDropViewIfExists("scott.my_view")
	if err != nil {
		t.Fatalf("CompileDropViewIfExists returned error for schema-qualified name: %v", err)
	}
	if !strings.Contains(sql, "DROP VIEW SCOTT.MY_VIEW") {
		t.Errorf("Expected uppercased schema-qualified name, got: %s", sql)
	}
}

func TestOracleCompileDropIfExists(t *testing.T) {
	g := newOracleGrammar()

	sql, err := g.CompileDropIfExists(newBlueprint("users"))
	if err != nil {
		t.Fatalf("CompileDropIfExists returned error: %v", err)
	}
	if !strings.Contains(sql, "BEGIN EXECUTE IMMEDIATE") {
		t.Errorf("Expected PL/SQL block, got: %s", sql)
	}
	if !strings.Contains(sql, "DROP TABLE USERS CASCADE CONSTRAINTS") {
		t.Errorf("Expected 'DROP TABLE USERS CASCADE CONSTRAINTS' inside block, got: %s", sql)
	}
	if !strings.HasSuffix(sql, "END;;") {
		t.Errorf("Expected to end with 'END;;', got: %s", sql)
	}

	// Schema-qualified name — still validated by wrap.Table, uppercased in output
	sql, err = g.CompileDropIfExists(newBlueprint("scott.users"))
	if err != nil {
		t.Fatalf("CompileDropIfExists returned error for schema-qualified name: %v", err)
	}
	if !strings.Contains(sql, "DROP TABLE SCOTT.USERS") {
		t.Errorf("Expected uppercased schema-qualified name, got: %s", sql)
	}
}

func TestOracleCompileDropIfExistsInjection(t *testing.T) {
	g := newOracleGrammar()

	// A quote closes the PL/SQL string literal and a semicolon starts a new
	// statement — both must be rejected by identifier validation.
	for _, name := range []string{
		"users'; DROP TABLE users; --",
		"x'; EXECUTE IMMEDIATE 'DROP TABLE users'; --",
		"users; DROP TABLE users",
		"users'--",
		"users\"",
		"users ",
		"",
	} {
		if _, err := g.CompileDropIfExists(newBlueprint(name)); err == nil {
			t.Errorf("Expected error for malicious table name %q, got nil", name)
		}
	}
}

func TestOracleCompileCreateViewInjection(t *testing.T) {
	g := newOracleGrammar()

	_, err := g.CompileCreateView(contractsschema.View{Name: "users; DROP TABLE users; --", Definition: "select 1"})
	if err == nil {
		t.Error("Expected error for invalid identifier, got nil")
	}
}
