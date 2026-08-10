//go:build integration

package jsondb

import (
	"testing"
)

func TestJSONDBIntegrationJoinLeft(t *testing.T) {
	db := SetupJSONDBTest(t)

	var results []jsondbOrderWithUser
	err := db.Query().
		Table("orders AS o").
		LeftJoin("users AS u ON o.user_id = u.id").
		Select("o.id, u.name AS user_name, o.total").
		OrderBy("o.id", "asc").
		Get(&results)

	if err != nil {
		t.Fatalf("Left join failed: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	// 1. Order 1: User 1 (Alice), product 2, qty 3, total 149.97
	if results[0].UserName != "Alice" {
		t.Errorf("Expected Alice for order 1, got %s", results[0].UserName)
	}
	if results[0].Total != 149.97 {
		t.Errorf("Expected 149.97, got %v", results[0].Total)
	}

	// 2. Order 2: User 3 (Charlie), product 1, qty 5, total 99.95
	if results[1].UserName != "Charlie" {
		t.Errorf("Expected Charlie for order 2, got %s", results[1].UserName)
	}

	// 3. Order 3: User 1 (Alice), product 3, qty 1, total 99.99
	if results[2].UserName != "Alice" {
		t.Errorf("Expected Alice for order 3, got %s", results[2].UserName)
	}
}

func TestJSONDBIntegrationJoinInner(t *testing.T) {
	db := SetupJSONDBTest(t)

	var results []jsondbOrderWithUser
	err := db.Query().
		Table("orders AS o").
		Join("users AS u ON o.user_id = u.id").
		Select("o.id, u.name AS user_name, o.total").
		Where("u.active = ?", true).
		OrderBy("o.id", "asc").
		Get(&results)

	if err != nil {
		t.Fatalf("Inner join failed: %v", err)
	}

	// Active users are Alice (ID 1) and Charlie (ID 3). All orders (1, 2, 3) are by active users.
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}
}

func TestJSONDBIntegrationJoinThreeTables(t *testing.T) {
	db := SetupJSONDBTest(t)

	type OrderDetails struct {
		ID          int     `db:"id"`
		UserName    string  `db:"user_name"`
		ProductName string  `db:"product_name"`
		Total       float64 `db:"total"`
	}

	var results []OrderDetails
	err := db.Query().
		Table("orders AS o").
		Join("users AS u ON o.user_id = u.id").
		Join("products AS p ON o.product_id = p.id").
		Select("o.id, u.name AS user_name, p.name AS product_name, o.total").
		OrderBy("o.id", "asc").
		Get(&results)

	if err != nil {
		t.Fatalf("3-table join failed: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	// Order 1: User 1 (Alice), product 2 (Gadget), total 149.97
	if results[0].UserName != "Alice" || results[0].ProductName != "Gadget" {
		t.Errorf("Unexpected result at row 0: %+v", results[0])
	}

	// Order 2: User 3 (Charlie), product 1 (Widget), total 99.95
	if results[1].UserName != "Charlie" || results[1].ProductName != "Widget" {
		t.Errorf("Unexpected result at row 1: %+v", results[1])
	}
}
