//go:build integration

package mysql

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestMySQLIntegrationUpdateOrInsert(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	query := db.Query()

	// Test Insert with Map
	err := query.Table("users").UpdateOrInsert(
		map[string]any{"name": "insert_map"},
		map[string]any{"avatar": "avatar_map"},
	)
	if err != nil {
		t.Errorf("UpdateOrInsert with map failed: %v", err)
	}

	var user models.User
	err = query.Table("users").Where("name = ?", "insert_map").First(&user)
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
	err = query.Table("users").UpdateOrInsert(
		map[string]any{"name": "insert_map"},
		map[string]any{"avatar": "avatar_map_updated"},
	)
	if err != nil {
		t.Errorf("UpdateOrInsert update with map failed: %v", err)
	}

	var user2 models.User
	err = query.Table("users").Where("name = ?", "insert_map").First(&user2)
	if err != nil {
		t.Errorf("Failed to find updated user: %v", err)
	}
	if user2.Avatar != "avatar_map_updated" {
		t.Errorf("Expected 'avatar_map_updated', got '%s'", user2.Avatar)
	}

	// Test Insert with Struct
	err = query.Table("users").UpdateOrInsert(
		models.User{Name: "insert_struct"},
		models.User{Avatar: "avatar_struct"},
	)
	if err != nil {
		t.Errorf("UpdateOrInsert with struct failed: %v", err)
	}

	var user3 models.User
	err = query.Table("users").Where("name = ?", "insert_struct").First(&user3)
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
	err = query.Table("users").UpdateOrInsert(
		models.User{Name: "insert_struct"},
		models.User{Avatar: "avatar_struct_updated"},
	)
	if err != nil {
		t.Errorf("UpdateOrInsert update with struct failed: %v", err)
	}

	var user4 models.User
	err = query.Table("users").Where("name = ?", "insert_struct").First(&user4)
	if err != nil {
		t.Errorf("Failed to find updated struct user: %v", err)
	}
	if user4.Avatar != "avatar_struct_updated" {
		t.Errorf("Expected 'avatar_struct_updated', got '%s'", user4.Avatar)
	}

	// Test with existing Where clause
	err = query.Table("users").Where("name = ?", "insert_map").UpdateOrInsert(
		map[string]any{"name": "insert_map", "avatar": "avatar_map_updated"},
		map[string]any{"bio": "bio_updated"},
	)
	if err != nil {
		t.Errorf("UpdateOrInsert with where clause failed: %v", err)
	}

	var user5 models.User
	err = query.Table("users").Where("name = ?", "insert_map").First(&user5)
	if err != nil {
		t.Errorf("Failed to find user with bio: %v", err)
	}
	if user5.Bio == nil {
		t.Error("Bio should be set")
		return
	}
	if *user5.Bio != "bio_updated" {
		t.Errorf("Expected 'bio_updated', got '%s'", *user5.Bio)
	}
}
