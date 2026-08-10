//go:build integration

package csvdb

import (
	"testing"
)

func TestCSVDBIntegrationBasicQuery(t *testing.T) {
	db := SetupCSVDBTest(t)

	var users []csvdbUser
	err := db.Query().Model(&csvdbUser{}).OrderBy("id", "asc").Get(&users)
	if err != nil {
		t.Fatalf("failed to query users: %v", err)
	}
	if len(users) != 3 {
		t.Fatalf("expected 3 users, got %d", len(users))
	}
	if users[0].Name != "Alice" {
		t.Errorf("expected first user 'Alice', got '%s'", users[0].Name)
	}
	if users[1].Name != "Bob" {
		t.Errorf("expected second user 'Bob', got '%s'", users[1].Name)
	}
	if users[2].Name != "Charlie" {
		t.Errorf("expected third user 'Charlie', got '%s'", users[2].Name)
	}
}

func TestCSVDBIntegrationWhereEquals(t *testing.T) {
	db := SetupCSVDBTest(t)

	var users []csvdbUser
	err := db.Query().Model(&csvdbUser{}).Where("name = ?", "Bob").Get(&users)
	if err != nil {
		t.Fatalf("failed to query: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}
	if users[0].ID != 2 {
		t.Errorf("expected user ID 2, got %d", users[0].ID)
	}
}

func TestCSVDBIntegrationWhereBool(t *testing.T) {
	db := SetupCSVDBTest(t)

	var users []csvdbUser
	err := db.Query().
		Model(&csvdbUser{}).
		Where("active = ?", true).
		OrderBy("name", "asc").
		Get(&users)
	if err != nil {
		t.Fatalf("failed to query active users: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 active users, got %d", len(users))
	}
	if users[0].Name != "Alice" {
		t.Errorf("expected first active user 'Alice', got '%s'", users[0].Name)
	}
	if users[1].Name != "Charlie" {
		t.Errorf("expected second active user 'Charlie', got '%s'", users[1].Name)
	}
}

func TestCSVDBIntegrationWhereInActiveFalse(t *testing.T) {
	db := SetupCSVDBTest(t)

	var users []csvdbUser
	err := db.Query().
		Model(&csvdbUser{}).
		Where("active = ?", false).
		Get(&users)
	if err != nil {
		t.Fatalf("failed to query inactive users: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 inactive user, got %d", len(users))
	}
	if users[0].Name != "Bob" {
		t.Errorf("expected inactive user 'Bob', got '%s'", users[0].Name)
	}
}

func TestCSVDBIntegrationOrderByAsc(t *testing.T) {
	db := SetupCSVDBTest(t)

	var users []csvdbUser
	err := db.Query().Model(&csvdbUser{}).OrderBy("name", "asc").Get(&users)
	if err != nil {
		t.Fatalf("failed to query: %v", err)
	}
	if len(users) != 3 {
		t.Fatalf("expected 3 users, got %d", len(users))
	}
	if users[0].Name != "Alice" {
		t.Errorf("expected first 'Alice', got '%s'", users[0].Name)
	}
	if users[1].Name != "Bob" {
		t.Errorf("expected second 'Bob', got '%s'", users[1].Name)
	}
	if users[2].Name != "Charlie" {
		t.Errorf("expected third 'Charlie', got '%s'", users[2].Name)
	}
}

func TestCSVDBIntegrationOrderByDesc(t *testing.T) {
	db := SetupCSVDBTest(t)

	var users []csvdbUser
	err := db.Query().Model(&csvdbUser{}).OrderBy("name", "desc").Get(&users)
	if err != nil {
		t.Fatalf("failed to query: %v", err)
	}
	if len(users) != 3 {
		t.Fatalf("expected 3 users, got %d", len(users))
	}
	if users[0].Name != "Charlie" {
		t.Errorf("expected first 'Charlie', got '%s'", users[0].Name)
	}
	if users[2].Name != "Alice" {
		t.Errorf("expected last 'Alice', got '%s'", users[2].Name)
	}
}

func TestCSVDBIntegrationLimit(t *testing.T) {
	db := SetupCSVDBTest(t)

	var products []csvdbProduct
	err := db.Query().
		Model(&csvdbProduct{}).
		OrderBy("price", "desc").
		Limit(2).
		Get(&products)
	if err != nil {
		t.Fatalf("failed to query products: %v", err)
	}
	if len(products) != 2 {
		t.Fatalf("expected 2 products, got %d", len(products))
	}
	if products[0].Name != "Gizmo" {
		t.Errorf("expected most expensive 'Gizmo', got '%s'", products[0].Name)
	}
	if products[1].Name != "Gadget" {
		t.Errorf("expected second 'Gadget', got '%s'", products[1].Name)
	}
}

func TestCSVDBIntegrationLimitOffset(t *testing.T) {
	db := SetupCSVDBTest(t)

	var products []csvdbProduct
	err := db.Query().
		Model(&csvdbProduct{}).
		OrderBy("price", "asc").
		Limit(2).
		Offset(1).
		Get(&products)
	if err != nil {
		t.Fatalf("failed to query products: %v", err)
	}
	if len(products) != 2 {
		t.Fatalf("expected 2 products, got %d", len(products))
	}
	// prices sorted asc: 19.99, 49.99, 99.99 → offset 1 → 49.99, 99.99
	if products[0].Name != "Gadget" {
		t.Errorf("expected first 'Gadget', got '%s'", products[0].Name)
	}
	if products[1].Name != "Gizmo" {
		t.Errorf("expected second 'Gizmo', got '%s'", products[1].Name)
	}
}

func TestCSVDBIntegrationWhereGreaterThan(t *testing.T) {
	db := SetupCSVDBTest(t)

	var products []csvdbProduct
	err := db.Query().
		Model(&csvdbProduct{}).
		Where("price > ?", 50).
		OrderBy("price", "desc").
		Get(&products)
	if err != nil {
		t.Fatalf("failed to query products: %v", err)
	}
	if len(products) != 1 {
		t.Fatalf("expected 1 product over $50, got %d", len(products))
	}
	if products[0].Name != "Gizmo" {
		t.Errorf("expected 'Gizmo', got '%s'", products[0].Name)
	}
}

func TestCSVDBIntegrationFirst(t *testing.T) {
	db := SetupCSVDBTest(t)

	var user csvdbUser
	err := db.Query().Model(&csvdbUser{}).Where("name = ?", "Charlie").First(&user)
	if err != nil {
		t.Fatalf("failed to get first: %v", err)
	}
	if user.ID != 3 {
		t.Errorf("expected ID 3, got %d", user.ID)
	}
	if !user.Active {
		t.Errorf("expected active=true, got false")
	}
}

func TestCSVDBIntegrationFind(t *testing.T) {
	db := SetupCSVDBTest(t)

	var users []csvdbUser
	err := db.Query().Model(&csvdbUser{}).Find(&users, "id = ?", 2)
	if err != nil {
		t.Fatalf("failed to find: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}
	if users[0].Name != "Bob" {
		t.Errorf("expected 'Bob', got '%s'", users[0].Name)
	}
}

func TestCSVDBIntegrationCount(t *testing.T) {
	db := SetupCSVDBTest(t)

	var count int64
	err := db.Query().Model(&csvdbUser{}).Count(&count)
	if err != nil {
		t.Fatalf("failed to count: %v", err)
	}
	if count != 3 {
		t.Errorf("expected count 3, got %d", count)
	}
}
