package query

import (
	"testing"
)

func TestRawExprFunction(t *testing.T) {
	// Test the RawExpr helper function
	expr := RawExpr("NOW()")
	if expr.SQL != "NOW()" {
		t.Errorf("Expected SQL 'NOW()', got '%s'", expr.SQL)
	}
	if len(expr.Args) != 0 {
		t.Errorf("Expected empty Args, got %v", expr.Args)
	}

	// Test with arguments
	expr2 := RawExpr("DATE_ADD(NOW(), INTERVAL ? DAY)", 7)
	if expr2.SQL != "DATE_ADD(NOW(), INTERVAL ? DAY)" {
		t.Errorf("Unexpected SQL: %s", expr2.SQL)
	}
	if len(expr2.Args) != 1 || expr2.Args[0] != 7 {
		t.Errorf("Unexpected Args: %v", expr2.Args)
	}
}

func TestRawExprFiltersNilArgs(t *testing.T) {
	// Test that nil arguments are filtered out
	expr := RawExpr("NOW()", nil)
	if len(expr.Args) != 0 {
		t.Errorf("Expected 0 args after filtering nil, got %d", len(expr.Args))
	}

	// Test that non-nil arguments are kept
	expr2 := RawExpr("score + ?", 10, nil, 20)
	if len(expr2.Args) != 2 {
		t.Errorf("Expected 2 args after filtering nil, got %d", len(expr2.Args))
	}
	if expr2.Args[0] != 10 {
		t.Errorf("Expected arg[0] to be 10, got %v", expr2.Args[0])
	}
	if expr2.Args[1] != 20 {
		t.Errorf("Expected arg[1] to be 20, got %v", expr2.Args[1])
	}
}
