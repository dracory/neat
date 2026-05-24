//go:build disabled

package postgres

import (
	"strings"
	"testing"

	"github.com/dracory/neat/database"
	"github.com/dracory/neat/integration_tests/models"
)

func TestPostgresIntegrationQueryValue(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	// Seed data
	users := []models.User{
		{Name: "value_user_1", Avatar: "avatar_1"},
		{Name: "value_user_2", Avatar: "avatar_2"},
	}
	for _, user := range users {
		if err := query.Model(&models.User{}).Create(&user); err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	t.Run("Value basic", func(t *testing.T) {
		var name string
		err := db.Query().Model(&models.User{}).OrderBy("id", "asc").Value("name", &name)
		if err != nil {
			t.Errorf("Value basic failed: %v", err)
		}
		if name != "value_user_1" {
			t.Errorf("Expected 'value_user_1', got '%s'", name)
		}
	})

	t.Run("Value with where", func(t *testing.T) {
		var name string
		err := db.Query().Model(&models.User{}).Where("avatar = ?", "avatar_2").Value("name", &name)
		if err != nil {
			t.Errorf("Value with where failed: %v", err)
		}
		if name != "value_user_2" {
			t.Errorf("Expected 'value_user_2', got '%s'", name)
		}
	})

	t.Run("ToSql Value", func(t *testing.T) {
		var name string
		sql := db.Query().Model(&models.User{}).Where("id = ?", 1).ToSql().Value("name", &name)
		if !strings.Contains(sql, "SELECT \"name\" FROM \"users\"") {
			t.Error("Expected SQL to contain 'SELECT \"name\" FROM \"users\"'")
		}
		if !strings.Contains(sql, "WHERE \"id\" = $1") {
			t.Error("Expected SQL to contain 'WHERE \"id\" = $1'")
		}
		if !strings.Contains(sql, "LIMIT 1") {
			t.Error("Expected SQL to contain 'LIMIT 1'")
		}
	})
}
