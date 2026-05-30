//go:build integration

package postgres

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestPostgresIntegrationUpdateOrInsertWithMapInsert(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)

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
}

func TestPostgresIntegrationUpdateOrInsertWithMapUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)

	err := db.Query().Table("users").UpdateOrInsert(
		map[string]any{"name": "insert_map"},
		map[string]any{"avatar": "avatar_map"},
	)
	if err != nil {
		t.Errorf("UpdateOrInsert with map (insert) failed: %v", err)
	}

	err = db.Query().Table("users").UpdateOrInsert(
		map[string]any{"name": "insert_map"},
		map[string]any{"avatar": "avatar_map_updated"},
	)
	if err != nil {
		t.Errorf("UpdateOrInsert with map (update) failed: %v", err)
	}

	var user models.User
	err = db.Query().Table("users").Where("name = ?", "insert_map").First(&user)
	if err != nil {
		t.Errorf("Failed to find updated user: %v", err)
	}
	if user.Avatar != "avatar_map_updated" {
		t.Errorf("Expected avatar 'avatar_map_updated', got '%s'", user.Avatar)
	}
}

func TestPostgresIntegrationUpdateOrInsertWithStructInsert(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	err := query.Model(&models.User{}).UpdateOrInsert(
		map[string]any{"name": "insert_struct"},
		map[string]any{"avatar": "avatar_struct"},
	)
	if err != nil {
		t.Fatalf("UpdateOrInsert with struct failed: %v", err)
	}

	var user models.User
	err = query.Model(&models.User{}).WithTrashed().Where("name = ?", "insert_struct").First(&user)
	if err != nil {
		t.Fatalf("Failed to find inserted struct user: %v", err)
	}
	if user.Name != "insert_struct" {
		t.Errorf("Expected 'insert_struct', got '%s'", user.Name)
	}
	if user.Avatar != "avatar_struct" {
		t.Errorf("Expected 'avatar_struct', got '%s'", user.Avatar)
	}
}

func TestPostgresIntegrationUpdateOrInsertWithStructUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	// First insert
	err := query.Model(&models.User{}).UpdateOrInsert(
		map[string]any{"name": "insert_struct"},
		map[string]any{"avatar": "avatar_struct"},
	)
	if err != nil {
		t.Fatalf("UpdateOrInsert with struct failed: %v", err)
	}

	// Then update
	err = query.Model(&models.User{}).UpdateOrInsert(
		map[string]any{"name": "insert_struct"},
		map[string]any{"avatar": "avatar_struct_updated"},
	)
	if err != nil {
		t.Errorf("UpdateOrInsert update with struct failed: %v", err)
	}

	var user models.User
	err = query.Model(&models.User{}).Where("name = ?", "insert_struct").First(&user)
	if err != nil {
		t.Errorf("Failed to find updated struct user: %v", err)
	}
	if user.Avatar != "avatar_struct_updated" {
		t.Errorf("Expected 'avatar_struct_updated', got '%s'", user.Avatar)
	}
}

func TestPostgresIntegrationUpdateOrInsertWithWhereClause(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	// First insert
	err := query.Model(&models.User{}).UpdateOrInsert(
		map[string]any{"name": "insert_map"},
		map[string]any{"avatar": "avatar_map"},
	)
	if err != nil {
		t.Errorf("UpdateOrInsert with map (insert) failed: %v", err)
	}

	// Then update with where clause
	err = query.Model(&models.User{}).Where("name = ?", "insert_map").UpdateOrInsert(
		map[string]any{"name": "insert_map"},
		map[string]any{"bio": "bio_updated"},
	)
	if err != nil {
		t.Errorf("UpdateOrInsert with where clause failed: %v", err)
	}

	var user models.User
	err = query.Model(&models.User{}).Where("name = ?", "insert_map").First(&user)
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
