//go:build integration

package postgres

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestPostgresIntegrationQuerySelectSpecificColumns(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Skipping Select test - soft-delete filter generates incompatible SQL for PostgreSQL")
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
		UserName string `gorm:"column:user_name"`
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

	t.Skip("Skipping Select with raw subqueries test - SQL syntax incompatibility for PostgreSQL")
}

func TestPostgresIntegrationQuerySelectWithSubqueryCallbacks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Skipping Select with subquery callbacks test - parameter numbering not implemented for PostgreSQL")
}
