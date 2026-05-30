package query

import (
	"context"
	"strings"
	"testing"
)

func TestClauseType(t *testing.T) {
	clauseTypes := []ClauseType{
		WhereClause,
		SelectClause,
		JoinClause,
		GroupClause,
		HavingClause,
		OrderClause,
		LimitClause,
		OffsetClause,
	}

	for i, ct := range clauseTypes {
		if int(ct) != i {
			t.Errorf("Expected ClauseType %d to have value %d, got %d", i, i, int(ct))
		}
	}
}

func TestWhereClauseBuilder(t *testing.T) {
	w := &WhereClauseBuilder{
		_type: "and",
		query: "id = ?",
		args:  []any{1},
	}

	if w.Type() != WhereClause {
		t.Errorf("Expected WhereClause, got %v", w.Type())
	}

	sql, args := w.ToSQL()
	if sql != "id = ?" {
		t.Errorf("Expected 'id = ?', got %q", sql)
	}
	if len(args) != 1 || args[0] != 1 {
		t.Errorf("Expected args [1], got %v", args)
	}
}

func TestSelectClauseBuilder(t *testing.T) {
	s := &SelectClauseBuilder{
		columns: []any{"name", "age"},
	}

	if s.Type() != SelectClause {
		t.Errorf("Expected SelectClause, got %v", s.Type())
	}

	sql, args := s.ToSQL()
	if sql != "name, age" {
		t.Errorf("Expected 'name, age', got %q", sql)
	}
	if len(args) != 0 {
		t.Errorf("Expected no args, got %v", args)
	}
}

func TestJoinClauseBuilder(t *testing.T) {
	j := &JoinClauseBuilder{
		_type: "LEFT JOIN",
		query: "orders ON orders.user_id = users.id",
		args:  nil,
	}

	if j.Type() != JoinClause {
		t.Errorf("Expected JoinClause, got %v", j.Type())
	}

	sql, args := j.ToSQL()
	if sql != "LEFT JOIN orders ON orders.user_id = users.id" {
		t.Errorf("Unexpected SQL: %q", sql)
	}
	if len(args) != 0 {
		t.Errorf("Expected no args, got %v", args)
	}
}

func TestJoinClauseBuilderWithArgs(t *testing.T) {
	j := &JoinClauseBuilder{
		_type: "JOIN",
		query: "orders ON orders.status = ?",
		args:  []any{"active"},
	}

	sql, args := j.ToSQL()
	if sql != "JOIN orders ON orders.status = ?" {
		t.Errorf("Unexpected SQL: %q", sql)
	}
	if len(args) != 1 || args[0] != "active" {
		t.Errorf("Expected args [\"active\"], got %v", args)
	}
}

func TestOrderClauseBuilder(t *testing.T) {
	o := &OrderClauseBuilder{
		column:    "created_at",
		direction: "desc",
	}

	if o.Type() != OrderClause {
		t.Errorf("Expected OrderClause, got %v", o.Type())
	}

	sql, args := o.ToSQL()
	if sql != "created_at desc" {
		t.Errorf("Expected 'created_at desc', got %q", sql)
	}
	if len(args) != 0 {
		t.Errorf("Expected no args, got %v", args)
	}
}

func TestOrderClauseBuilderAsc(t *testing.T) {
	o := &OrderClauseBuilder{
		column:    "name",
		direction: "asc",
	}

	sql, _ := o.ToSQL()
	if sql != "name asc" {
		t.Errorf("Expected 'name asc', got %q", sql)
	}
}

func TestLimitClauseBuilder(t *testing.T) {
	l := &LimitClauseBuilder{limit: 25}

	if l.Type() != LimitClause {
		t.Errorf("Expected LimitClause, got %v", l.Type())
	}

	sql, args := l.ToSQL()
	if sql != "LIMIT 25" {
		t.Errorf("Expected 'LIMIT 25', got %q", sql)
	}
	if len(args) != 0 {
		t.Errorf("Expected no args, got %v", args)
	}
}

func TestOffsetClauseBuilder(t *testing.T) {
	o := &OffsetClauseBuilder{offset: 50}

	if o.Type() != OffsetClause {
		t.Errorf("Expected OffsetClause, got %v", o.Type())
	}

	sql, args := o.ToSQL()
	if sql != "OFFSET 50" {
		t.Errorf("Expected 'OFFSET 50', got %q", sql)
	}
	if len(args) != 0 {
		t.Errorf("Expected no args, got %v", args)
	}
}

func TestWhereClauseBuilderOrType(t *testing.T) {
	w := &WhereClauseBuilder{
		_type: "or",
		query: "status = ?",
		args:  []any{"inactive"},
	}

	sql, args := w.ToSQL()
	if sql != "status = ?" {
		t.Errorf("Expected 'status = ?', got %q", sql)
	}
	if len(args) != 1 || args[0] != "inactive" {
		t.Errorf("Expected args [\"inactive\"], got %v", args)
	}
}

func TestClauseOrderingInSQL(t *testing.T) {
	// Verifies that BuildSelect emits clauses in the correct SQL order:
	// SELECT … FROM … JOIN … WHERE … GROUP BY … HAVING … ORDER BY … LIMIT … OFFSET
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.table = "users"
	q.selects = []selectClause{{expr: "id"}}
	q.joins = []joinClause{{_type: "LEFT JOIN", query: "orders ON orders.user_id = users.id"}}
	q.wheres = []whereClause{{_type: "and", query: "users.active = ?", args: []any{true}}}
	q.groups = []string{"users.id"}
	q.havings = []havingClause{{query: "COUNT(orders.id) > ?", args: []any{0}}}
	q.orders = []orderClause{{column: "users.id", direction: "asc"}}
	limit, offset := 10, 20
	q.limit = &limit
	q.offset = &offset

	sql, _ := NewBuilder(q).BuildSelect()

	clauses := []string{"SELECT", "FROM", "LEFT JOIN", "WHERE", "GROUP BY", "HAVING", "ORDER BY", "LIMIT", "OFFSET"}
	pos := -1
	for _, clause := range clauses {
		next := strings.Index(sql, clause)
		if next == -1 {
			t.Errorf("Expected SQL to contain %q, got: %s", clause, sql)
			continue
		}
		if next <= pos {
			t.Errorf("Clause %q appears out of order in SQL: %s", clause, sql)
		}
		pos = next
	}
}
