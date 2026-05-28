package turso

import (
	"testing"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/database"
	"github.com/dracory/neat/integration_tests/models"
)

func seedScopesTestData(t *testing.T, db *database.Database) {
	query := db.Query()
	users := []models.User{
		{Name: "scope_user_1", Avatar: "active"},
		{Name: "scope_user_2", Avatar: "active"},
		{Name: "scope_user_3", Avatar: "inactive"},
	}

	for _, user := range users {
		if err := query.Model(&models.User{}).Create(&user); err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}
}

func TestTursoIntegrationQueryScopesLocalWithoutParameters(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	seedScopesTestData(t, db)
	query := db.Query()

	activeScope := func(q contractsorm.Query) contractsorm.Query {
		return q.Where("avatar = ?", "active")
	}

	var activeUsers []models.User
	err := query.Model(&models.User{}).Scopes(activeScope).Find(&activeUsers)
	if err != nil {
		t.Errorf("Local scope without parameters failed: %v", err)
	}
	if len(activeUsers) != 2 {
		t.Errorf("Expected 2 users, got %d", len(activeUsers))
	}
	for _, u := range activeUsers {
		if u.Avatar != "active" {
			t.Errorf("Expected avatar 'active', got '%s'", u.Avatar)
		}
	}
}

func TestTursoIntegrationQueryScopesLocalWithParameters(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	seedScopesTestData(t, db)
	query := db.Query()

	nameScope := func(name string) func(contractsorm.Query) contractsorm.Query {
		return func(q contractsorm.Query) contractsorm.Query {
			return q.Where("name = ?", name)
		}
	}

	foundUsers := make([]models.User, 0)
	err := query.Model(&models.User{}).Scopes(nameScope("scope_user_1")).Find(&foundUsers)
	if err != nil {
		t.Errorf("Local scope with parameters failed: %v", err)
	}
	if len(foundUsers) != 1 {
		t.Errorf("Expected 1 user, got %d", len(foundUsers))
	}
	if foundUsers[0].Name != "scope_user_1" {
		t.Errorf("Expected 'scope_user_1', got '%s'", foundUsers[0].Name)
	}
}
