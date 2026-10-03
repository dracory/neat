package query

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

func newUpdateTestBuilder(dialect string) *Builder {
	return NewBuilder(NewQuery(context.TODO(), nil, &FakeDriver{DialectName: dialect}, "users", nil, nil))
}

func TestBuildUpdateSetMapIsSortedAndSkipsOmitted(t *testing.T) {
	b := newUpdateTestBuilder("postgres")
	b.query.omitColumns = []string{"secret"}

	set := b.buildUpdateSet(map[string]any{"b": 2, "a": 1, "secret": "x", "c": 3}, nil)

	if len(set.exprs) != 3 {
		t.Fatalf("Expected 3 expressions, got %v", set.exprs)
	}
	for i, want := range []string{"a", "b", "c"} {
		if !strings.Contains(set.exprs[i], want) {
			t.Errorf("Expected expr %d to target %q, got %s", i, want, set.exprs[i])
		}
	}
	if !strings.HasSuffix(set.exprs[2], "$3") {
		t.Errorf("Expected last placeholder $3, got %s", set.exprs[2])
	}
	if !reflect.DeepEqual(set.args, []any{1, 2, 3}) {
		t.Errorf("Expected args [1 2 3], got %v", set.args)
	}
	if set.index != 4 {
		t.Errorf("Expected next index 4, got %d", set.index)
	}
}

func TestBuildUpdateSetRawExpressionInMap(t *testing.T) {
	b := newUpdateTestBuilder("postgres")

	set := b.buildUpdateSet(map[string]any{"a": 1, "score": RawExpr("score + ? * ?", 2, 3)}, nil)

	if len(set.exprs) != 2 || !strings.HasSuffix(set.exprs[1], "score + $2 * $3") {
		t.Errorf("Expected raw expression with $2 and $3, got %v", set.exprs)
	}
	if !reflect.DeepEqual(set.args, []any{1, 2, 3}) {
		t.Errorf("Expected args [1 2 3], got %v", set.args)
	}
}

func TestBuildUpdateSetSingleColumn(t *testing.T) {
	b := newUpdateTestBuilder("sqlite")

	set := b.buildUpdateSet("name", []any{"Ann"})

	if len(set.exprs) != 1 || !strings.Contains(set.exprs[0], "name") || !strings.HasSuffix(set.exprs[0], "= ?") {
		t.Errorf("Unexpected expressions: %v", set.exprs)
	}
	if !reflect.DeepEqual(set.args, []any{"Ann"}) {
		t.Errorf("Unexpected args: %v", set.args)
	}
}

func TestBuildUpdateSetExpression(t *testing.T) {
	b := newUpdateTestBuilder("postgres")

	set := b.buildUpdateSet("views = views + ?", []any{5})

	if len(set.exprs) != 1 || set.exprs[0] != "views = views + $1" {
		t.Errorf("Unexpected expressions: %v", set.exprs)
	}
	if set.index != 2 {
		t.Errorf("Expected next index 2, got %d", set.index)
	}
}

func TestBuildUpdateSetJSONPath(t *testing.T) {
	tests := []struct {
		dialect string
		want    string
	}{
		{"mysql", "JSON_SET("},
		{"sqlite", "json_set("},
	}
	for _, tt := range tests {
		t.Run(tt.dialect, func(t *testing.T) {
			set := newUpdateTestBuilder(tt.dialect).buildUpdateSet("data->meta->active", []any{true})
			if len(set.exprs) != 1 || !strings.Contains(set.exprs[0], tt.want) || !strings.Contains(set.exprs[0], "'$.meta.active'") {
				t.Errorf("Unexpected expressions: %v", set.exprs)
			}
			if !reflect.DeepEqual(set.args, []any{true}) {
				t.Errorf("Unexpected args: %v", set.args)
			}
		})
	}
}

func TestBuildUpdateSetJSONPathFallbackForOtherDialects(t *testing.T) {
	set := newUpdateTestBuilder("postgres").buildUpdateSet("data->name", []any{"x"})

	if len(set.exprs) != 1 || strings.Contains(strings.ToLower(set.exprs[0]), "json_set") || !strings.HasSuffix(set.exprs[0], "= $1") {
		t.Errorf("Expected plain assignment fallback, got %v", set.exprs)
	}
}

func TestBuildUpdateSetStruct(t *testing.T) {
	type user struct {
		Name     string
		Password string
	}
	b := newUpdateTestBuilder("sqlite")
	b.query.omitColumns = []string{"password"}

	set := b.buildUpdateSet(&user{Name: "Ann", Password: "pw"}, nil)

	if len(set.exprs) != 1 || !strings.Contains(set.exprs[0], "name") {
		t.Errorf("Expected only name to be set, got %v", set.exprs)
	}
	if !reflect.DeepEqual(set.args, []any{"Ann"}) {
		t.Errorf("Unexpected args: %v", set.args)
	}
}

func TestBuildUpdateSetUnsupportedInputIsEmpty(t *testing.T) {
	set := newUpdateTestBuilder("sqlite").buildUpdateSet(42, nil)

	if len(set.exprs) != 0 || len(set.args) != 0 {
		t.Errorf("Expected empty SET, got %v / %v", set.exprs, set.args)
	}
}

func TestIsOmitted(t *testing.T) {
	b := newUpdateTestBuilder("sqlite")
	b.query.omitColumns = []string{"password"}

	if !b.isOmitted("password") || b.isOmitted("name") {
		t.Error("isOmitted returned wrong result")
	}
}

func TestIsSoftDeleteUpdate(t *testing.T) {
	b := newUpdateTestBuilder("sqlite")
	b.query.model = &updateTestSoftModel{}

	if !b.isSoftDeleteUpdate(map[string]any{"deleted_at": nil}) {
		t.Error("Expected map containing the soft-delete column to be a soft-delete update")
	}
	if b.isSoftDeleteUpdate(map[string]any{"name": "x"}) {
		t.Error("Expected map without the soft-delete column not to be a soft-delete update")
	}
	if b.isSoftDeleteUpdate("deleted_at") {
		t.Error("Expected non-map input not to be a soft-delete update")
	}
}

func TestAppendUpdateWhereAndLimit(t *testing.T) {
	limit := 5
	tests := []struct {
		name      string
		dialect   string
		limit     *int
		where     string
		wantParts []string
		wantArgs  int
	}{
		{"no limit", "mysql", nil, "id = ?", []string{"UPDATE", "WHERE id = ?"}, 1},
		{"no where", "mysql", nil, "", []string{"UPDATE"}, 0},
		{"mysql limit", "mysql", &limit, "id = ?", []string{"UPDATE", "WHERE id = ?", "LIMIT 5"}, 1},
		{"postgres limit", "postgres", &limit, "id = ?", []string{"UPDATE", "WHERE id = ?", "LIMIT 5"}, 1},
		{"sqlserver top", "sqlserver", &limit, "id = ?", []string{"UPDATE TOP (5)", "WHERE id = ?"}, 1},
		{"unknown dialect ignores limit", "oracle", &limit, "id = ?", []string{"UPDATE", "WHERE id = ?"}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newUpdateTestBuilder(tt.dialect)
			b.query.limit = tt.limit
			var whereArgs []any
			if tt.where != "" {
				whereArgs = []any{1}
			}

			parts, args := b.appendUpdateWhereAndLimit([]string{"UPDATE"}, nil, tt.where, whereArgs)

			if !reflect.DeepEqual(parts, tt.wantParts) {
				t.Errorf("Expected parts %v, got %v", tt.wantParts, parts)
			}
			if len(args) != tt.wantArgs {
				t.Errorf("Expected %d args, got %d", tt.wantArgs, len(args))
			}
		})
	}
}

func TestAppendUpdateWhereAndLimitSQLite(t *testing.T) {
	limit := 3
	b := newUpdateTestBuilder("sqlite")
	b.query.limit = &limit
	b.query.orders = []orderClause{{column: "id", direction: "DESC"}}

	parts, args := b.appendUpdateWhereAndLimit([]string{"UPDATE"}, nil, "", nil)

	got := strings.Join(parts, " ")
	for _, want := range []string{"WHERE rowid IN (SELECT rowid FROM", "WHERE 1=1", "ORDER BY", "DESC LIMIT 3)"} {
		if !strings.Contains(got, want) {
			t.Errorf("Expected %q in %s", want, got)
		}
	}
	if len(args) != 0 {
		t.Errorf("Expected no args, got %v", args)
	}
}

func TestBuildUpdateOrderClause(t *testing.T) {
	b := newUpdateTestBuilder("sqlite")
	if got := b.buildUpdateOrderClause(); got != "" {
		t.Errorf("Expected empty clause, got %q", got)
	}

	b.query.orders = []orderClause{{column: "a", direction: "ASC"}, {column: "b", direction: "DESC"}}
	got := b.buildUpdateOrderClause()
	if !strings.HasPrefix(got, " ORDER BY ") || !strings.Contains(got, "ASC, ") || !strings.HasSuffix(got, "DESC") {
		t.Errorf("Unexpected order clause: %q", got)
	}
}
