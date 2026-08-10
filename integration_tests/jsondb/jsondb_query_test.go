//go:build integration

package jsondb

import (
	"testing"
)

func TestJSONDBIntegrationBasicQuery(t *testing.T) {
	db := SetupJSONDBTest(t)

	var users []jsondbUser
	err := db.Query().Model(&jsondbUser{}).Get(&users)
	if err != nil {
		t.Fatalf("Failed to query users: %v", err)
	}

	if len(users) != 3 {
		t.Errorf("Expected 3 users, got %d", len(users))
	}

	if users[0].Name != "Alice" || users[1].Name != "Bob" || users[2].Name != "Charlie" {
		t.Errorf("Unexpected user names: %v", users)
	}
}

func TestJSONDBIntegrationProductAndOrderModels(t *testing.T) {
	db := SetupJSONDBTest(t)

	var products []jsondbProduct
	err := db.Query().Model(&jsondbProduct{}).Get(&products)
	if err != nil {
		t.Fatalf("Failed to query products model: %v", err)
	}
	if len(products) != 3 {
		t.Errorf("Expected 3 products, got %d", len(products))
	}

	var orders []jsondbOrder
	err = db.Query().Model(&jsondbOrder{}).Get(&orders)
	if err != nil {
		t.Fatalf("Failed to query orders model: %v", err)
	}
	if len(orders) != 3 {
		t.Errorf("Expected 3 orders, got %d", len(orders))
	}
}

func TestJSONDBIntegrationWhereEquals(t *testing.T) {
	db := SetupJSONDBTest(t)

	var users []jsondbUser
	err := db.Query().Model(&jsondbUser{}).Where("email = ?", "bob@example.com").Get(&users)
	if err != nil {
		t.Fatalf("Failed to query: %v", err)
	}

	if len(users) != 1 {
		t.Fatalf("Expected 1 user, got %d", len(users))
	}
	if users[0].Name != "Bob" {
		t.Errorf("Expected Bob, got %s", users[0].Name)
	}
}

func TestJSONDBIntegrationWhereBool(t *testing.T) {
	db := SetupJSONDBTest(t)

	var activeUsers []jsondbUser
	// active is a bool column, maps to 1 in SQLite
	err := db.Query().Model(&jsondbUser{}).Where("active = ?", true).Get(&activeUsers)
	if err != nil {
		t.Fatalf("Failed to query active users: %v", err)
	}

	if len(activeUsers) != 2 {
		t.Errorf("Expected 2 active users, got %d", len(activeUsers))
	}

	var inactiveUsers []jsondbUser
	err = db.Query().Model(&jsondbUser{}).Where("active = ?", false).Get(&inactiveUsers)
	if err != nil {
		t.Fatalf("Failed to query inactive users: %v", err)
	}

	if len(inactiveUsers) != 1 {
		t.Errorf("Expected 1 inactive user, got %d", len(inactiveUsers))
	}
	if inactiveUsers[0].Name != "Bob" {
		t.Errorf("Expected Bob as inactive user, got %s", inactiveUsers[0].Name)
	}
}

func TestJSONDBIntegrationOrderByAsc(t *testing.T) {
	db := SetupJSONDBTest(t)

	var users []jsondbUser
	err := db.Query().Model(&jsondbUser{}).OrderBy("name", "asc").Get(&users)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(users) != 3 {
		t.Fatalf("Expected 3 users, got %d", len(users))
	}
	if users[0].Name != "Alice" || users[1].Name != "Bob" || users[2].Name != "Charlie" {
		t.Errorf("Unexpected sort order: %v", users)
	}
}

func TestJSONDBIntegrationOrderByDesc(t *testing.T) {
	db := SetupJSONDBTest(t)

	var users []jsondbUser
	err := db.Query().Model(&jsondbUser{}).OrderBy("name", "desc").Get(&users)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(users) != 3 {
		t.Fatalf("Expected 3 users, got %d", len(users))
	}
	if users[0].Name != "Charlie" || users[1].Name != "Bob" || users[2].Name != "Alice" {
		t.Errorf("Unexpected sort order: %v", users)
	}
}

func TestJSONDBIntegrationLimit(t *testing.T) {
	db := SetupJSONDBTest(t)

	var users []jsondbUser
	err := db.Query().Model(&jsondbUser{}).OrderBy("id", "asc").Limit(2).Get(&users)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}
	if users[0].Name != "Alice" || users[1].Name != "Bob" {
		t.Errorf("Unexpected users with limit: %v", users)
	}
}

func TestJSONDBIntegrationLimitOffset(t *testing.T) {
	db := SetupJSONDBTest(t)

	var users []jsondbUser
	err := db.Query().Model(&jsondbUser{}).OrderBy("id", "asc").Limit(1).Offset(1).Get(&users)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(users) != 1 {
		t.Fatalf("Expected 1 user, got %d", len(users))
	}
	if users[0].Name != "Bob" {
		t.Errorf("Expected Bob, got %s", users[0].Name)
	}
}

func TestJSONDBIntegrationFirst(t *testing.T) {
	db := SetupJSONDBTest(t)

	var user jsondbUser
	err := db.Query().Model(&jsondbUser{}).Where("id = ?", 3).First(&user)
	if err != nil {
		t.Fatalf("First failed: %v", err)
	}

	if user.Name != "Charlie" {
		t.Errorf("Expected Charlie, got %s", user.Name)
	}
}

func TestJSONDBIntegrationFind(t *testing.T) {
	db := SetupJSONDBTest(t)

	var users []jsondbUser
	err := db.Query().Model(&jsondbUser{}).Find(&users, "id = ?", 2)
	if err != nil {
		t.Fatalf("Find failed: %v", err)
	}

	if len(users) != 1 {
		t.Fatalf("Expected 1 user, got %d", len(users))
	}
	if users[0].Name != "Bob" {
		t.Errorf("Expected Bob, got %s", users[0].Name)
	}
}

func TestJSONDBIntegrationCount(t *testing.T) {
	db := SetupJSONDBTest(t)

	var count int64
	err := db.Query().Model(&jsondbUser{}).Count(&count)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected count of 3, got %d", count)
	}
}
