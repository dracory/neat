//go:build integration

package sqlite

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestSQLiteIntegrationUpdateOrInsert(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)

	// Test Insert with Map
	err := db.Query().Table("users").UpdateOrInsert(
		map[string]any{"name": "insert_map"},
		map[string]any{"avatar": "avatar_map"},
	)
	if err != nil {
		t.Errorf("UpdateOrInsert with map (insert) failed: %v", err)
	}

	var user models.User
	err = db.Query().Table("users").Where("name = ?", "insert_map").First(&user)
	if err != nil {
		t.Errorf("Failed to find inserted user: %v", err)
	}
	if user.Name != "insert_map" {
		t.Errorf("Expected 'insert_map', got '%s'", user.Name)
	}
	if user.Avatar != "avatar_map" {
		t.Errorf("Expected 'avatar_map', got '%s'", user.Avatar)
	}

	// Test Update with Map
	err = db.Query().Table("users").UpdateOrInsert(
		map[string]any{"name": "insert_map"},
		map[string]any{"avatar": "avatar_map_updated"},
	)
	if err != nil {
		t.Errorf("UpdateOrInsert with map (update) failed: %v", err)
	}

	var user2 models.User
	err = db.Query().Table("users").Where("name = ?", "insert_map").First(&user2)
	if err != nil {
		t.Errorf("Failed to find updated user: %v", err)
	}
	if user2.Avatar != "avatar_map_updated" {
		t.Errorf("Expected 'avatar_map_updated', got '%s'", user2.Avatar)
	}

	// Test Insert with Struct
	err = db.Query().Table("users").UpdateOrInsert(
		models.User{Name: "insert_struct"},
		models.User{Avatar: "avatar_struct"},
	)
	if err != nil {
		t.Errorf("UpdateOrInsert with struct (insert) failed: %v", err)
	}

	var user3 models.User
	err = db.Query().Table("users").Where("name = ?", "insert_struct").First(&user3)
	if err != nil {
		t.Errorf("Failed to find inserted struct user: %v", err)
	}
	if user3.Name != "insert_struct" {
		t.Errorf("Expected 'insert_struct', got '%s'", user3.Name)
	}
	if user3.Avatar != "avatar_struct" {
		t.Errorf("Expected 'avatar_struct', got '%s'", user3.Avatar)
	}

	// Test Update with Struct
	err = db.Query().Table("users").UpdateOrInsert(
		models.User{Name: "insert_struct"},
		models.User{Avatar: "avatar_struct_updated"},
	)
	if err != nil {
		t.Errorf("UpdateOrInsert with struct (update) failed: %v", err)
	}

	var user4 models.User
	err = db.Query().Table("users").Where("name = ?", "insert_struct").First(&user4)
	if err != nil {
		t.Errorf("Failed to find updated struct user: %v", err)
	}
	if user4.Avatar != "avatar_struct_updated" {
		t.Errorf("Expected 'avatar_struct_updated', got '%s'", user4.Avatar)
	}

	// Test with existing Where clause
	err = db.Query().Table("users").Where("name = ?", "insert_map").UpdateOrInsert(
		map[string]any{"name": "insert_map", "avatar": "avatar_map_updated"},
		map[string]any{"bio": "bio_updated"},
	)
	if err != nil {
		t.Errorf("UpdateOrInsert with where clause failed: %v", err)
	}

	var user5 models.User
	err = db.Query().Table("users").Where("name = ?", "insert_map").First(&user5)
	if err != nil {
		t.Errorf("Failed to find user with where clause: %v", err)
	}
	if user5.Bio == nil {
		t.Error("Expected Bio to be non-nil")
	}
	if *user5.Bio != "bio_updated" {
		t.Errorf("Expected 'bio_updated', got '%s'", *user5.Bio)
	}

	// Test interaction with existing Where clause
	// Create another user
	err = db.Query().Table("users").Create(map[string]any{"name": "other_user", "avatar": "other_avatar"})
	if err != nil {
		t.Fatalf("Failed to create other user: %v", err)
	}

	// UpdateOrInsert with a Where clause that should restrict the operation
	err = db.Query().Table("users").Where("name = ?", "other_user").UpdateOrInsert(
		map[string]any{"name": "insert_map"}, // attributes match user5, but Where matches other_user
		map[string]any{"avatar": "restricted_avatar"},
	)
	if err != nil {
		t.Errorf("UpdateOrInsert with restriction failed: %v", err)
	}

	// User5 should NOT be updated because of the Where clause on other_user
	var userCheck models.User
	err = db.Query().Table("users").Where("name = ?", "insert_map").First(&userCheck)
	if err != nil {
		t.Errorf("Failed to find user check: %v", err)
	}
	if userCheck.Avatar != "avatar_map_updated" {
		t.Errorf("Expected 'avatar_map_updated', got '%s'", userCheck.Avatar)
	}

	// other_user SHOULD be updated if the attributes and Where clause combined result in a match
	// Wait, in Goravel/Laravel, attributes are also merged into the Where.
	// If Where("name", "other_user") and attributes is {"name": "insert_map"}, then it looks for name=other_user AND name=insert_map, which is empty.
	// So it should INSERT.

	var userOther models.User
	err = db.Query().Table("users").Where("name = ?", "other_user").First(&userOther)
	if err != nil {
		t.Errorf("Failed to find other user: %v", err)
	}
	if userOther.Avatar != "other_avatar" {
		t.Errorf("Expected 'other_avatar', got '%s'", userOther.Avatar)
	}

	// Check if a new record was inserted
	var count int64
	db.Query().Table("users").Where("avatar = ?", "restricted_avatar").Count(&count)
	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}
}
