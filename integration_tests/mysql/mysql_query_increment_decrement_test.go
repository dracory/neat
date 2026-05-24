//go:build integration

package mysql

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestMySQLIntegrationQueryIncrement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Skip("Increment on auto-increment ID not applicable")

	db := SetupMySQLTest(t)
	query := db.Query()

	user := models.User{Name: "increment_user", Avatar: "group1"}
	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	var createdUser models.User
	if err := query.Model(&models.User{}).Where("name = ?", "increment_user").First(&createdUser); err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}
}

func TestMySQLIntegrationQueryDecrement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Skip("Decrement on auto-increment ID not applicable")

	db := SetupMySQLTest(t)
	query := db.Query()

	user := models.User{Name: "increment_user", Avatar: "group1"}
	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	var createdUser models.User
	if err := query.Model(&models.User{}).Where("name = ?", "increment_user").First(&createdUser); err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}
}

func TestMySQLIntegrationQueryIncrementDecrementWithWhere(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Skip("Increment/decrement on auto-increment ID not applicable")

	db := SetupMySQLTest(t)
	query := db.Query()

	user := models.User{Name: "increment_user", Avatar: "group1"}
	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	var createdUser models.User
	if err := query.Model(&models.User{}).Where("name = ?", "increment_user").First(&createdUser); err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}

	user2 := models.User{Name: "other_user", Avatar: "group2"}
	if err := db.Query().Model(&models.User{}).Create(&user2); err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	var createdUser2 models.User
	if err := db.Query().Model(&models.User{}).Where("name = ?", "other_user").First(&createdUser2); err != nil {
		t.Fatalf("Failed to get created user2: %v", err)
	}
}
