package main_test

import (
	"testing"

	"github.com/dracory/neat"
	"github.com/dracory/neat/contracts/database/schema"
	mainpkg "github.com/dracory/neat/examples/advanced-queries"
)

func TestRunExample(t *testing.T) {
	// Use in-memory SQLite for testing
	err := mainpkg.RunExample("sqlite://:memory:")
	if err != nil {
		t.Fatalf("RunExample failed: %v", err)
	}
}

func setupAdvancedDB(t *testing.T) *neat.Database {
	t.Helper()
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}

	err = db.Schema().Create("users", func(bp schema.Blueprint) {
		bp.ID()
		bp.String("name")
		bp.String("email")
		bp.Integer("age")
		bp.String("status")
		bp.Timestamp("created_at")
		bp.Timestamp("deleted_at").Nullable()
	})
	if err != nil {
		t.Fatalf("failed to create users: %v", err)
	}

	err = db.Schema().Create("orders", func(bp schema.Blueprint) {
		bp.ID()
		bp.Integer("user_id")
		bp.Decimal("amount")
	})
	if err != nil {
		t.Fatalf("failed to create orders: %v", err)
	}

	users := []map[string]any{
		{"name": "Alice", "email": "alice@example.com", "age": 25, "status": "active", "created_at": "2026-01-01 00:00:00"},
		{"name": "Bob", "email": "bob@example.com", "age": 30, "status": "active", "created_at": "2026-01-02 00:00:00"},
		{"name": "Charlie", "email": "charlie@example.com", "age": 16, "status": "pending", "created_at": "2026-01-03 00:00:00"},
		{"name": "Dave", "email": "dave@example.com", "age": 40, "status": "inactive", "created_at": "2026-01-04 00:00:00"},
	}
	for _, u := range users {
		if err = db.Query().Table("users").Create(u); err != nil {
			t.Fatalf("failed to seed user: %v", err)
		}
	}

	orders := []map[string]any{
		{"user_id": 1, "amount": 100.0},
		{"user_id": 2, "amount": 200.0},
		{"user_id": 1, "amount": 50.0},
	}
	for _, o := range orders {
		if err = db.Query().Table("orders").Create(o); err != nil {
			t.Fatalf("failed to seed order: %v", err)
		}
	}

	return db
}

func TestAdvancedQueries_OrWhere(t *testing.T) {
	db := setupAdvancedDB(t)
	defer db.Close()

	// active OR pending = Alice, Bob, Charlie
	var results []map[string]any
	err := db.Query().Table("users").
		Where("status = ?", "active").
		OrWhere("status = ?", "pending").
		Get(&results)
	if err != nil {
		t.Fatalf("OrWhere query failed: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 (active+pending), got %d", len(results))
	}
}

func TestAdvancedQueries_WhereIn(t *testing.T) {
	db := setupAdvancedDB(t)
	defer db.Close()

	var results []map[string]any
	err := db.Query().Table("users").WhereIn("id", []any{1, 2}).Get(&results)
	if err != nil {
		t.Fatalf("WhereIn failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 users with id in (1,2), got %d", len(results))
	}
}

func TestAdvancedQueries_GroupAndCount(t *testing.T) {
	db := setupAdvancedDB(t)
	defer db.Close()

	var groupResults []map[string]any
	err := db.Query().Table("users").
		Select("status", "COUNT(*) as count").
		Group("status").
		Having("COUNT(*) > 0").
		Get(&groupResults)
	if err != nil {
		t.Fatalf("Group query failed: %v", err)
	}
	// 3 distinct status values: active, pending, inactive
	if len(groupResults) != 3 {
		t.Errorf("expected 3 status groups, got %d", len(groupResults))
	}
}

func TestAdvancedQueries_CountAggregation(t *testing.T) {
	db := setupAdvancedDB(t)
	defer db.Close()

	var count int64
	err := db.Query().Table("users").Count(&count)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 4 {
		t.Errorf("expected 4 users, got %d", count)
	}
}

func TestAdvancedQueries_SumAggregation(t *testing.T) {
	db := setupAdvancedDB(t)
	defer db.Close()

	var sum float64
	err := db.Query().Table("orders").Sum("amount", &sum)
	if err != nil {
		t.Fatalf("Sum failed: %v", err)
	}
	if sum != 350.0 {
		t.Errorf("expected sum 350.0, got %f", sum)
	}
}

func TestAdvancedQueries_Pagination(t *testing.T) {
	db := setupAdvancedDB(t)
	defer db.Close()

	// Page 1: first 2 records
	var page1 []map[string]any
	err := db.Query().Table("users").Limit(2).Offset(0).Get(&page1)
	if err != nil {
		t.Fatalf("page 1 query failed: %v", err)
	}
	if len(page1) != 2 {
		t.Errorf("expected 2 records on page 1, got %d", len(page1))
	}

	// Page 2: next 2 records
	var page2 []map[string]any
	err = db.Query().Table("users").Limit(2).Offset(2).Get(&page2)
	if err != nil {
		t.Fatalf("page 2 query failed: %v", err)
	}
	if len(page2) != 2 {
		t.Errorf("expected 2 records on page 2, got %d", len(page2))
	}

	// Page 3: beyond total, should be empty
	var page3 []map[string]any
	err = db.Query().Table("users").Limit(2).Offset(4).Get(&page3)
	if err != nil {
		t.Fatalf("page 3 query failed: %v", err)
	}
	if len(page3) != 0 {
		t.Errorf("expected 0 records on page 3, got %d", len(page3))
	}
}
