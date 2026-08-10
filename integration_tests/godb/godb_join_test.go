//go:build integration

package godb

import (
	"testing"
)

func TestGODB_Join_Left(t *testing.T) {
	db := SetupGODBTest(t)

	var results []godbOrderWithUser
	err := db.Query().
		Table("orders").
		LeftJoin("users AS u ON orders.user_id = u.id").
		Select("orders.id", "u.name AS user_name", "orders.total").
		OrderBy("orders.id", "asc").
		Get(&results)

	if err != nil {
		t.Fatalf("JOIN failed: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 join results, got %d", len(results))
	}

	// Order 1: User 1 (Alice), Total 149.97
	// Order 2: User 3 (Charlie), Total 99.95
	// Order 3: User 1 (Alice), Total 99.99
	if results[0].UserName != "Alice" || results[0].Total != 149.97 {
		t.Errorf("unexpected first result: %+v", results[0])
	}
	if results[1].UserName != "Charlie" || results[1].Total != 99.95 {
		t.Errorf("unexpected second result: %+v", results[1])
	}
	if results[2].UserName != "Alice" || results[2].Total != 99.99 {
		t.Errorf("unexpected third result: %+v", results[2])
	}
}

func TestGODB_Join_ThreeTables(t *testing.T) {
	db := SetupGODBTest(t)

	type FullOrderDetail struct {
		OrderID     int     `db:"order_id"`
		UserName    string  `db:"user_name"`
		ProductName string  `db:"product_name"`
		Price       float64 `db:"price"`
		Quantity    int     `db:"quantity"`
		Total       float64 `db:"total"`
	}

	var results []FullOrderDetail
	err := db.Query().
		Table("orders AS o").
		Join("users AS u ON o.user_id = u.id").
		Join("products AS p ON o.product_id = p.id").
		Select("o.id AS order_id", "u.name AS user_name", "p.name AS product_name", "p.price", "o.quantity", "o.total").
		OrderBy("o.id", "asc").
		Get(&results)

	if err != nil {
		t.Fatalf("3-table JOIN failed: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// Order 1: Alice, Gadget (49.99), Quantity 3, Total 149.97
	first := results[0]
	if first.OrderID != 1 || first.UserName != "Alice" || first.ProductName != "Gadget" || first.Price != 49.99 || first.Quantity != 3 || first.Total != 149.97 {
		t.Errorf("unexpected first detail: %+v", first)
	}
}
