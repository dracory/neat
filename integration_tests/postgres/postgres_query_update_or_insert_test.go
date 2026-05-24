//go:build disabled

package postgres

import (
	"testing"

	"github.com/dracory/neat/database"
	"github.com/dracory/neat/integration_tests/models"
)

func TestPostgresIntegrationUpdateOrInsert(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)

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
		t.Errorf("Expected name 'insert_map', got '%s'", user.Name)
	}
	if user.Avatar != "avatar_map" {
		t.Errorf("Expected avatar 'avatar_map', got '%s'", user.Avatar)
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
		t.Errorf("Expected avatar 'avatar_map_updated', got '%s'", user2.Avatar)
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
		t.Errorf("Expected name 'insert_struct', got '%s'", user3.Name)
	}
	if user3.Avatar != "avatar_struct" {
		t.Errorf("Expected avatar 'avatar_struct', got '%s'", user3.Avatar)
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
		t.Errorf("Expected avatar 'avatar_struct_updated', got '%s'", user4.Avatar)
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
		t.Errorf("Expected bio 'bio_updated', got '%s'", *user5.Bio)
	}
}
