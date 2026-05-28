package turso

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestTursoIntegrationQueryFind(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	query := db.Query()

	// Create test data
	user := models.User{Name: "find_user"}
	err := query.Model(&models.User{}).Create(&user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Find all users
	var users []models.User
	err = query.Model(&models.User{}).Find(&users)
	if err != nil {
		t.Errorf("Find failed: %v", err)
	}
	if len(users) == 0 {
		t.Error("Should have found at least one user")
	}
}

func TestTursoIntegrationQueryFirst(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	query := db.Query()

	// Create test data
	user := models.User{Name: "first_user"}
	err := query.Model(&models.User{}).Create(&user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Find first user
	var foundUser models.User
	err = query.Model(&models.User{}).First(&foundUser)
	if err != nil {
		t.Errorf("First failed: %v", err)
	}
	if foundUser.ID == 0 {
		t.Error("ID should be set")
	}
}

func TestTursoIntegrationQueryWhere(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	query := db.Query()

	// Create test data
	user := models.User{Name: "where_user"}
	err := query.Model(&models.User{}).Create(&user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Find user with where clause
	var foundUser models.User
	err = query.Model(&models.User{}).Where("name = ?", "where_user").First(&foundUser)
	if err != nil {
		t.Errorf("Where query failed: %v", err)
	}
	if foundUser.Name != "where_user" {
		t.Errorf("Expected name 'where_user', got '%s'", foundUser.Name)
	}
}
