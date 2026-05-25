//go:build disabled

package sqlite

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestSQLiteIntegrationUpdateOrInsertWithMap(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	query := db.Query()

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
}

func TestSQLiteIntegrationUpdateOrInsertUpdateWithMap(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	query := db.Query()

	err := query.Table("users").UpdateOrInsert(
		map[string]any{"name": "insert_map"},
		map[string]any{"avatar": "avatar_map"},
	)
	if err != nil {
		t.Errorf("UpdateOrInsert with map failed: %v", err)
	}

	err = query.Table("users").UpdateOrInsert(
		map[string]any{"name": "insert_map"},
		map[string]any{"avatar": "avatar_map_updated"},
	)
	if err != nil {
		t.Errorf("UpdateOrInsert update with map failed: %v", err)
	}

	var user models.User
	err = query.Table("users").Where("name = ?", "insert_map").First(&user)
	if err != nil {
		t.Errorf("Failed to find updated user: %v", err)
	}
	if user.Avatar != "avatar_map_updated" {
		t.Errorf("Expected 'avatar_map_updated', got '%s'", user.Avatar)
	}
}

func TestSQLiteIntegrationUpdateOrInsertWithStruct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	query := db.Query()

	err := query.Table("users").UpdateOrInsert(
		models.User{Name: "insert_struct"},
		models.User{Avatar: "avatar_struct"},
	)
	if err != nil {
		t.Errorf("UpdateOrInsert with struct failed: %v", err)
	}

	var user models.User
	err = query.Table("users").Where("name = ?", "insert_struct").First(&user)
	if err != nil {
		t.Errorf("Failed to find inserted struct user: %v", err)
	}
	if user.Name != "insert_struct" {
		t.Errorf("Expected 'insert_struct', got '%s'", user.Name)
	}
	if user.Avatar != "avatar_struct" {
		t.Errorf("Expected 'avatar_struct', got '%s'", user.Avatar)
	}
}

func TestSQLiteIntegrationUpdateOrInsertUpdateWithStruct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	query := db.Query()

	err := query.Table("users").UpdateOrInsert(
		models.User{Name: "insert_struct"},
		models.User{Avatar: "avatar_struct"},
	)
	if err != nil {
		t.Errorf("UpdateOrInsert with struct failed: %v", err)
	}

	err = query.Table("users").UpdateOrInsert(
		models.User{Name: "insert_struct"},
		models.User{Avatar: "avatar_struct_updated"},
	)
	if err != nil {
		t.Errorf("UpdateOrInsert update with struct failed: %v", err)
	}

	var user models.User
	err = query.Table("users").Where("name = ?", "insert_struct").First(&user)
	if err != nil {
		t.Errorf("Failed to find updated struct user: %v", err)
	}
	if user.Avatar != "avatar_struct_updated" {
		t.Errorf("Expected 'avatar_struct_updated', got '%s'", user.Avatar)
	}
}

func TestSQLiteIntegrationUpdateOrInsertWithWhereClause(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	query := db.Query()

	err := query.Table("users").UpdateOrInsert(
		map[string]any{"name": "insert_map"},
		map[string]any{"avatar": "avatar_map"},
	)
	if err != nil {
		t.Errorf("UpdateOrInsert with map failed: %v", err)
	}

	err = query.Table("users").Where("name = ?", "insert_map").UpdateOrInsert(
		map[string]any{"name": "insert_map", "avatar": "avatar_map_updated"},
		map[string]any{"bio": "bio_updated"},
	)
	if err != nil {
		t.Errorf("UpdateOrInsert with where clause failed: %v", err)
	}

	var user models.User
	err = query.Table("users").Where("name = ?", "insert_map").First(&user)
	if err != nil {
		t.Errorf("Failed to find user with bio: %v", err)
	}
	if user.Bio == nil {
		t.Error("Bio should be set")
		return
	}
	if *user.Bio != "bio_updated" {
		t.Errorf("Expected 'bio_updated', got '%s'", *user.Bio)
	}
}
