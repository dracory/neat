package csvdb

import (
	"testing"
)

func TestCSVDBIntegrationJoinLeft(t *testing.T) {
	db := SetupCSVDBTest(t)

	var results []csvdbOrderWithUser
	err := db.Query().
		Table("orders AS o").
		LeftJoin("users AS u ON o.user_id = u.id").
		Select("o.id, u.name AS user_name, o.total").
		OrderBy("o.id", "asc").
		Get(&results)
	if err != nil {
		t.Fatalf("failed to query join: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 joined rows, got %d", len(results))
	}
	// Order 1: user_id=1 → Alice, total=149.97
	if results[0].ID != 1 {
		t.Errorf("expected order ID 1, got %d", results[0].ID)
	}
	if results[0].UserName != "Alice" {
		t.Errorf("expected first order's user 'Alice', got '%s'", results[0].UserName)
	}
	if results[0].Total != 149.97 {
		t.Errorf("expected first order total 149.97, got %f", results[0].Total)
	}
	// Order 2: user_id=3 → Charlie, total=99.95
	if results[1].UserName != "Charlie" {
		t.Errorf("expected second order's user 'Charlie', got '%s'", results[1].UserName)
	}
	// Order 3: user_id=1 → Alice, total=99.99
	if results[2].UserName != "Alice" {
		t.Errorf("expected third order's user 'Alice', got '%s'", results[2].UserName)
	}
}

func TestCSVDBIntegrationJoinInner(t *testing.T) {
	db := SetupCSVDBTest(t)

	var results []csvdbOrderWithUser
	err := db.Query().
		Table("orders AS o").
		Join("users AS u ON o.user_id = u.id").
		Select("o.id, u.name AS user_name, o.total").
		OrderBy("o.id", "asc").
		Get(&results)
	if err != nil {
		t.Fatalf("failed to query inner join: %v", err)
	}
	// All orders have matching users, so inner join returns all 3
	if len(results) != 3 {
		t.Fatalf("expected 3 joined rows, got %d", len(results))
	}
}

func TestCSVDBIntegrationJoinThreeTables(t *testing.T) {
	db := SetupCSVDBTest(t)

	type orderDetail struct {
		ID          int     `db:"id"`
		UserName    string  `db:"user_name"`
		ProductName string  `db:"product_name"`
		Quantity    int     `db:"quantity"`
		Total       float64 `db:"total"`
	}

	var results []orderDetail
	err := db.Query().
		Table("orders AS o").
		LeftJoin("users AS u ON o.user_id = u.id").
		LeftJoin("products AS p ON o.product_id = p.id").
		Select("o.id, u.name AS user_name, p.name AS product_name, o.quantity, o.total").
		OrderBy("o.id", "asc").
		Get(&results)
	if err != nil {
		t.Fatalf("failed to query 3-table join: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(results))
	}
	// Order 1: user_id=1 → Alice, product_id=2 → Gadget, qty=3, total=149.97
	if results[0].UserName != "Alice" {
		t.Errorf("expected user 'Alice', got '%s'", results[0].UserName)
	}
	if results[0].ProductName != "Gadget" {
		t.Errorf("expected product 'Gadget', got '%s'", results[0].ProductName)
	}
	if results[0].Quantity != 3 {
		t.Errorf("expected quantity 3, got %d", results[0].Quantity)
	}
}

func TestCSVDBIntegrationJoinWithWhere(t *testing.T) {
	db := SetupCSVDBTest(t)

	var results []csvdbOrderWithUser
	err := db.Query().
		Table("orders AS o").
		LeftJoin("users AS u ON o.user_id = u.id").
		Select("o.id, u.name AS user_name, o.total").
		Where("u.active = ?", true).
		OrderBy("o.id", "asc").
		Get(&results)
	if err != nil {
		t.Fatalf("failed to query join with where: %v", err)
	}
	// Alice (active=true) has orders 1 and 3; Charlie (active=true) has order 2
	// Bob (active=false) has no orders
	if len(results) != 3 {
		t.Fatalf("expected 3 rows (all orders belong to active users), got %d", len(results))
	}
	for _, r := range results {
		if r.UserName == "Bob" {
			t.Errorf("Bob should not appear (inactive user)")
		}
	}
}
