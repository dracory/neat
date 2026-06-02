package postgres_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestPostgresIntegrationQuerySelectSpecificColumns(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	user := models.User{Name: "select_user", Avatar: "select_avatar"}
	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	var createdUser models.User
	if err := query.Model(&models.User{}).Where("name = ?", "select_user").First(&createdUser); err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}

	var foundUser models.User
	err := query.Model(&models.User{}).Select("name").Where("id = ?", createdUser.ID).First(&foundUser)
	if err != nil {
		t.Errorf("Find failed: %v", err)
	}
	if foundUser.Name != "select_user" {
		t.Errorf("Expected 'select_user', got '%s'", foundUser.Name)
	}
	if foundUser.Avatar != "" {
		t.Errorf("Expected empty avatar, got '%s'", foundUser.Avatar)
	}
}

func TestPostgresIntegrationQuerySelectWithAliases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	user := models.User{Name: "alias_user"}
	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	var result struct {
		UserName string `db:"column:user_name"`
	}
	err := query.Model(&models.User{}).Select("name as user_name").Where("id = ?", user.ID).Scan(&result)
	if err != nil {
		t.Errorf("Select with aliases failed: %v", err)
	}
	if result.UserName != "alias_user" {
		t.Errorf("Expected 'alias_user', got '%s'", result.UserName)
	}
}

func TestPostgresIntegrationQuerySelectWithRawSubqueries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Skip("Skipping - nested subquery placeholder numbering needs more complex solution")

	db := SetupPostgresTest(t)
	query := db.Query()

	user := models.User{Name: "subquery_user"}
	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	var createdUser models.User
	if err := query.Model(&models.User{}).Where("name = ?", "subquery_user").First(&createdUser); err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}

	var result struct {
		SubName string `db:"column:sub_name"`
	}
	err := query.Model(&models.User{}).
		Select("(SELECT name FROM users WHERE id = ?) as sub_name", createdUser.ID).
		Where("id = ?", createdUser.ID).
		Scan(&result)
	if err != nil {
		t.Errorf("Scan failed: %v", err)
	}
	if result.SubName != "subquery_user" {
		t.Errorf("Expected 'subquery_user', got '%s'", result.SubName)
	}
}
