//go:build integration

package sqlite

import (
	"testing"

	"github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/integration_tests/models"
)

func TestSQLiteIntegrationQueryScopes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	query := db.Query()

	// Seed data
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

	// Define scopes
	activeScope := func(query orm.Query) orm.Query {
		return query.Where("avatar = ?", "active")
	}

	nameScope := func(name string) func(orm.Query) orm.Query {
		return func(query orm.Query) orm.Query {
			return query.Where("name = ?", name)
		}
	}

	t.Run("Local scope without parameters", func(t *testing.T) {
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
	})

	t.Run("Local scope with parameters", func(t *testing.T) {
		var foundUsers []models.User
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
	})

	t.Run("Multiple scopes chaining", func(t *testing.T) {
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

		// Chain with inactive
		foundUsers = nil
		err = query.Model(&models.User{}).Scopes(activeScope, nameScope("scope_user_3")).Find(&foundUsers)
		if err != nil {
			t.Errorf("Multiple scopes with inactive failed: %v", err)
		}
		if len(foundUsers) != 0 {
			t.Errorf("Expected 0 users, got %d", len(foundUsers))
		}
	})
}
