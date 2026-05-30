
package mysql

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

func TestMySQLIntegrationQueryScopesWithoutParameters(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	query := db.Query()
	seedScopesTestData(t, db)

	activeScope := func(query contractsorm.Query) contractsorm.Query {
		return query.Where("avatar = ?", "active")
	}

	var activeUsers []models.User
	err := query.Model(&models.User{}).Scopes(activeScope).Find(&activeUsers)
	if err != nil {
		t.Errorf("Scope without parameters failed: %v", err)
	}
	if len(activeUsers) != 2 {
		t.Errorf("Expected 2 active users, got %d", len(activeUsers))
	}
	for _, u := range activeUsers {
		if u.Avatar != "active" {
			t.Errorf("Expected avatar 'active', got '%s'", u.Avatar)
		}
	}
}

func TestMySQLIntegrationQueryScopesWithParameters(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	query := db.Query()
	seedScopesTestData(t, db)

	nameScope := func(name string) func(contractsorm.Query) contractsorm.Query {
		return func(query contractsorm.Query) contractsorm.Query {
			return query.Where("name = ?", name)
		}
	}

	var foundUsers []models.User
	err := query.Model(&models.User{}).Scopes(nameScope("scope_user_1")).Find(&foundUsers)
	if err != nil {
		t.Errorf("Scope with parameters failed: %v", err)
	}
	if len(foundUsers) != 1 {
		t.Errorf("Expected 1 user, got %d", len(foundUsers))
	}
	if foundUsers[0].Name != "scope_user_1" {
		t.Errorf("Expected 'scope_user_1', got '%s'", foundUsers[0].Name)
	}
}

func TestMySQLIntegrationQueryScopesMultipleChaining(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	query := db.Query()
	seedScopesTestData(t, db)

	activeScope := func(query contractsorm.Query) contractsorm.Query {
		return query.Where("avatar = ?", "active")
	}

	nameScope := func(name string) func(contractsorm.Query) contractsorm.Query {
		return func(query contractsorm.Query) contractsorm.Query {
			return query.Where("name = ?", name)
		}
	}

	var foundUsers []models.User
	err := query.Model(&models.User{}).Scopes(activeScope, nameScope("scope_user_1")).Find(&foundUsers)
	if err != nil {
		t.Errorf("Multiple scopes chaining failed: %v", err)
	}
	if len(foundUsers) != 1 {
		t.Errorf("Expected 1 user, got %d", len(foundUsers))
	}
	if foundUsers[0].Name != "scope_user_1" {
		t.Errorf("Expected 'scope_user_1', got '%s'", foundUsers[0].Name)
	}
	if foundUsers[0].Avatar != "active" {
		t.Errorf("Expected avatar 'active', got '%s'", foundUsers[0].Avatar)
	}

	foundUsers = nil
	err = query.Model(&models.User{}).Scopes(activeScope, nameScope("scope_user_3")).Find(&foundUsers)
	if err != nil {
		t.Errorf("Multiple scopes with inactive failed: %v", err)
	}
	if len(foundUsers) != 0 {
		t.Errorf("Expected 0 users, got %d", len(foundUsers))
	}
}
