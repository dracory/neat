//go:build integration

package godb

import (
	"testing"
)

func TestGODB_Query_Basic(t *testing.T) {
	db := SetupGODBTest(t)

	// Test basic Find
	var users []godbUser
	err := db.Query().Table("users").OrderBy("id", "asc").Get(&users)
	if err != nil {
		t.Fatalf("failed to get users: %v", err)
	}

	if len(users) != 3 {
		t.Fatalf("expected 3 users, got %d", len(users))
	}
	if users[0].Name != "Alice" || users[1].Name != "Bob" || users[2].Name != "Charlie" {
		t.Errorf("unexpected users data: %v", users)
	}
}

func TestGODB_Query_Where(t *testing.T) {
	db := SetupGODBTest(t)

	// Test Where equals
	var users []godbUser
	err := db.Query().Table("users").Where("name = ?", "Alice").Get(&users)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if len(users) != 1 || users[0].ID != 1 {
		t.Errorf("expected only Alice, got: %v", users)
	}

	// Test Where boolean
	var activeUsers []godbUser
	err = db.Query().Table("users").Where("active = ?", true).OrderBy("id", "asc").Get(&activeUsers)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if len(activeUsers) != 2 || activeUsers[0].ID != 1 || activeUsers[1].ID != 3 {
		t.Errorf("expected Alice and Charlie, got: %v", activeUsers)
	}
}

func TestGODB_Query_OrderBy_Limit_Offset(t *testing.T) {
	db := SetupGODBTest(t)

	// OrderBy Desc + Limit + Offset
	var products []godbProduct
	err := db.Query().Table("products").OrderBy("price", "desc").Limit(2).Offset(1).Get(&products)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}

	if len(products) != 2 {
		t.Fatalf("expected 2 products, got %d", len(products))
	}
	// Prices: Gizmo (99.99), Gadget (49.99), Widget (19.99)
	// Price desc: Gizmo, Gadget, Widget
	// Limit 2 Offset 1: Gadget, Widget
	if products[0].Name != "Gadget" || products[1].Name != "Widget" {
		t.Errorf("expected Gadget and Widget, got: %v", products)
	}
}

func TestGODB_Query_First(t *testing.T) {
	db := SetupGODBTest(t)

	var u godbUser
	err := db.Query().Table("users").Where("id = ?", 3).First(&u)
	if err != nil {
		t.Fatalf("First err: %v", err)
	}
	if u.Name != "Charlie" {
		t.Errorf("expected Charlie, got %v", u)
	}
}

func TestGODB_Query_Count(t *testing.T) {
	db := SetupGODBTest(t)

	var count int64
	err := db.Query().Table("products").Where("category = ?", "Electronics").Count(&count)
	if err != nil {
		t.Fatalf("Count err: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 electronics, got %d", count)
	}
}
