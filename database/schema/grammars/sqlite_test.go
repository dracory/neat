package grammars_test

import (
	"strings"
	"testing"

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
