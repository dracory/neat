
package postgres

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestPostgresIntegrationQueryValueBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	users := []models.User{
		{Name: "value_user_1", Avatar: "avatar_1"},
		{Name: "value_user_2", Avatar: "avatar_2"},
	}
	for _, user := range users {
		if err := query.Model(&models.User{}).Create(&user); err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	var name string
	err := db.Query().Model(&models.User{}).OrderBy("id", "asc").Value("name", &name)
	if err != nil {
		t.Errorf("Value basic failed: %v", err)
	}
	if name != "value_user_1" {
		t.Errorf("Expected 'value_user_1', got '%s'", name)
	}
}

func TestPostgresIntegrationQueryValueWithWhere(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	users := []models.User{
		{Name: "value_user_1", Avatar: "avatar_1"},
		{Name: "value_user_2", Avatar: "avatar_2"},
	}
	for _, user := range users {
		if err := query.Model(&models.User{}).Create(&user); err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	var name string
	err := db.Query().Model(&models.User{}).Where("avatar = ?", "avatar_2").Value("name", &name)
	if err != nil {
		t.Errorf("Value with where failed: %v", err)
	}
	if name != "value_user_2" {
		t.Errorf("Expected 'value_user_2', got '%s'", name)
	}
}

func TestPostgresIntegrationQueryToSqlValue(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	users := []models.User{
		{Name: "value_user_1", Avatar: "avatar_1"},
		{Name: "value_user_2", Avatar: "avatar_2"},
	}
	for _, user := range users {
		if err := query.Model(&models.User{}).Create(&user); err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	var name string
	sql := db.Query().Model(&models.User{}).Where("id = ?", 1).ToSql().Value("name", &name)
	if sql == "" {
		t.Error("SQL should not be empty")
	}
}
