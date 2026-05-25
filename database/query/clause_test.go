package query

import (
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
