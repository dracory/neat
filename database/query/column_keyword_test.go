package query

import (
	"strings"
	"testing"
)

// TestWhereKeywordColumn verifies a reserved-word column in Where() is quoted.
func TestWhereKeywordColumn(t *testing.T) {
	tests := []struct {
		name     string
		build    func(q *Query)
		expected string
	}{
		{"shorthand equals", func(q *Query) { q.Where("group", 1) }, `"group" = ?`},
		{"explicit operator", func(q *Query) { q.Where("group > ?", 1) }, `"group" > ?`},
		{"between", func(q *Query) { q.Where("group BETWEEN ? AND ?", 1, 2) }, `"group" BETWEEN`},
		{"or where", func(q *Query) { q.Where("id = ?", 1).OrWhere("order", 2) }, `"order" = ?`},
		{"where in", func(q *Query) { q.WhereIn("group", []any{1, 2}) }, `"group" IN`},
		{"where between", func(q *Query) { q.WhereBetween("order", 1, 2) }, `"order" BETWEEN`},
		{"where null", func(q *Query) { q.WhereNull("group") }, `"group" IS NULL`},
		{"where like", func(q *Query) { q.WhereLike("group", "%x%") }, `"group" LIKE ?`},
		{"where column", func(q *Query) { q.WhereColumn("name", "=", "group") }, `"name" = "group"`},
		{"dotted", func(q *Query) { q.Where("orders.group = ?", 1) }, `"orders"."group" = ?`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := newTableTestQuery("sqlite")
			q.table = "users"
			tt.build(q)
			if q.buildError != nil {
				t.Fatalf("unexpected buildError: %v", q.buildError)
			}
			sql, _ := NewBuilder(q).BuildSelect()
			if !strings.Contains(sql, tt.expected) {
				t.Errorf("expected %q in SQL, got: %s", tt.expected, sql)
			}
		})
	}
}

// TestWhereKeywordColumnMySQL verifies backtick quoting of keyword columns.
func TestWhereKeywordColumnMySQL(t *testing.T) {
	q := newTableTestQuery("mysql")
	q.table = "users"
	q.Where("group", 1)
	sql, _ := NewBuilder(q).BuildSelect()
	if !strings.Contains(sql, "`group` = ?") {
		t.Errorf("expected backtick-quoted column, got: %s", sql)
	}
}

// TestWhereAnyAllNoneKeywordColumns verifies keyword columns are accepted
// and quoted in the multi-column where helpers.
func TestWhereAnyAllNoneKeywordColumns(t *testing.T) {
	tests := []struct {
		name     string
		build    func(q *Query)
		expected string
	}{
		{"any", func(q *Query) { q.WhereAny([]string{"group", "order"}, "=", 1) }, `"group" = ?`},
		{"all", func(q *Query) { q.WhereAll([]string{"group", "order"}, "=", 1) }, `"order" = ?`},
		{"none", func(q *Query) { q.WhereNone([]string{"group"}, "=", 1) }, `NOT ("group" = ?)`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := newTableTestQuery("sqlite")
			q.table = "users"
			tt.build(q)
			if q.buildError != nil {
				t.Fatalf("unexpected buildError: %v", q.buildError)
			}
			sql, _ := NewBuilder(q).BuildSelect()
			if !strings.Contains(sql, tt.expected) {
				t.Errorf("expected %q in SQL, got: %s", tt.expected, sql)
			}
		})
	}
}

// TestGroupOrderDistinctKeywordColumns verifies GROUP BY / ORDER BY / DISTINCT
// accept and quote keyword columns.
func TestGroupOrderDistinctKeywordColumns(t *testing.T) {
	t.Run("group by", func(t *testing.T) {
		q := newTableTestQuery("sqlite")
		q.table = "users"
		q.Group("group")
		if q.buildError != nil {
			t.Fatalf("unexpected buildError: %v", q.buildError)
		}
		sql, _ := NewBuilder(q).BuildSelect()
		if !strings.Contains(sql, `GROUP BY "group"`) {
			t.Errorf("expected quoted GROUP BY, got: %s", sql)
		}
	})

	t.Run("group by dotted", func(t *testing.T) {
		q := newTableTestQuery("sqlite")
		q.table = "users"
		q.Group("orders.group")
		sql, _ := NewBuilder(q).BuildSelect()
		if !strings.Contains(sql, `GROUP BY "orders"."group"`) {
			t.Errorf("expected quoted dotted GROUP BY, got: %s", sql)
		}
	})

	t.Run("order by", func(t *testing.T) {
		q := newTableTestQuery("sqlite")
		q.table = "users"
		q.OrderBy("order")
		sql, _ := NewBuilder(q).BuildSelect()
		if !strings.Contains(sql, `ORDER BY "order" asc`) {
			t.Errorf("expected quoted ORDER BY, got: %s", sql)
		}
	})

	t.Run("order by desc", func(t *testing.T) {
		q := newTableTestQuery("sqlite")
		q.table = "users"
		q.OrderByDesc("group")
		sql, _ := NewBuilder(q).BuildSelect()
		if !strings.Contains(sql, `ORDER BY "group" desc`) {
			t.Errorf("expected quoted ORDER BY, got: %s", sql)
		}
	})

	t.Run("order string", func(t *testing.T) {
		q := newTableTestQuery("sqlite")
		q.table = "users"
		q.Order("group DESC")
		sql, _ := NewBuilder(q).BuildSelect()
		if !strings.Contains(sql, `ORDER BY "group" desc`) {
			t.Errorf("expected quoted ORDER BY, got: %s", sql)
		}
	})

	t.Run("count distinct", func(t *testing.T) {
		q := newTableTestQuery("sqlite")
		q.table = "users"
		q.distinct = true
		q.aggregate = "COUNT"
		q.Distinct("group")
		sql, _ := NewBuilder(q).BuildSelect()
		if !strings.Contains(sql, `COUNT(DISTINCT "group")`) {
			t.Errorf("expected quoted COUNT(DISTINCT ...), got: %s", sql)
		}
	})
}

// TestAggregateKeywordColumn verifies SUM("group") is quoted while
// "*" and "1" aggregate columns stay raw.
func TestAggregateKeywordColumn(t *testing.T) {
	q := newTableTestQuery("sqlite")
	q.table = "users"
	q.aggregate = "SUM"
	q.aggregateCol = "group"
	sql, _ := NewBuilder(q).BuildSelect()
	if !strings.Contains(sql, `SUM("group")`) {
		t.Errorf("expected SUM(\"group\"), got: %s", sql)
	}

	q2 := newTableTestQuery("sqlite")
	q2.table = "users"
	q2.aggregate = "COUNT"
	q2.aggregateCol = "1"
	sql2, _ := NewBuilder(q2).BuildSelect()
	if !strings.Contains(sql2, `COUNT(1)`) {
		t.Errorf("expected raw COUNT(1), got: %s", sql2)
	}
}

// TestSelectKeywordColumn verifies a keyword column in Select() is quoted.
func TestSelectKeywordColumn(t *testing.T) {
	q := newTableTestQuery("sqlite")
	q.table = "users"
	q.Select([]string{"group", "name"})
	sql, _ := NewBuilder(q).BuildSelect()
	if !strings.Contains(sql, `SELECT "group", "name"`) {
		t.Errorf("expected quoted select list, got: %s", sql)
	}
}

// TestKeywordColumnInsert verifies a keyword column in an INSERT is quoted.
func TestKeywordColumnInsert(t *testing.T) {
	q := newTableTestQuery("sqlite")
	q.table = "users"
	sql, _ := NewBuilder(q).BuildInsert(map[string]any{"group": "a"})
	if !strings.Contains(sql, `"group"`) {
		t.Errorf("expected quoted column in INSERT, got: %s", sql)
	}
}

// TestInvalidColumnNamesSetBuildError verifies genuinely invalid column names
// set buildError instead of being silently dropped.
func TestInvalidColumnNamesSetBuildError(t *testing.T) {
	tests := []struct {
		name  string
		build func(q *Query)
	}{
		{"group by injection", func(q *Query) { q.Group("a;b") }},
		{"order by injection", func(q *Query) { q.OrderBy("x--") }},
		{"order by desc injection", func(q *Query) { q.OrderByDesc("x' OR '1'='1") }},
		{"order injection", func(q *Query) { q.Order("a b") }},
		{"where like bad column", func(q *Query) { q.WhereLike("a b", "x") }},
		{"where column bad rhs", func(q *Query) { q.WhereColumn("x", "=", "a' OR '1'='1") }},
		{"where any bad col", func(q *Query) { q.WhereAny([]string{"ok", "bad;col"}, "=", 1) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := newTableTestQuery("sqlite")
			q.table = "users"
			tt.build(q)
			if q.buildError == nil {
				t.Errorf("expected buildError for invalid column name")
			}
		})
	}
}

// TestIsValidColumnReferenceKeywords verifies keywords are now valid column refs.
func TestIsValidColumnReferenceKeywords(t *testing.T) {
	valid := []string{"group", "order", "group.id", "orders.group"}
	for _, in := range valid {
		if !isValidColumnReference(in) {
			t.Errorf("isValidColumnReference(%q) = false, want true", in)
		}
	}
	invalid := []string{"a;b", "x--", "a.b.c", "1col", "a' OR '1'='1"}
	for _, in := range invalid {
		if isValidColumnReference(in) {
			t.Errorf("isValidColumnReference(%q) = true, want false", in)
		}
	}
}
