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

	t.Skip("Skipping UpdateOrInsert with struct test - soft-delete filter incompatibility")
}

func TestPostgresIntegrationUpdateOrInsertWithStructUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Skipping UpdateOrInsert with struct test - soft-delete filter incompatibility")
}

func TestPostgresIntegrationUpdateOrInsertWithWhereClause(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Skipping UpdateOrInsert with where clause test - soft-delete filter incompatibility")
}
