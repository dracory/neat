//go:build integration

package postgres

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestPostgresIntegrationQueryIncrement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	user := models.User{Name: "increment_user", ID: 1, Avatar: "group1"}
	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	res, err := db.Query().Table("users").Where("id = ?", 1).Increment("id")
	if err != nil {
		t.Errorf("Increment failed: %v", err)
	}
	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	var updatedUser models.User
	err = db.Query().Table("users").Where("name = ?", "increment_user").First(&updatedUser)
	if err != nil {
		t.Errorf("Failed to find updated user: %v", err)
	}
	if updatedUser.ID != 2 {
		t.Errorf("Expected ID 2, got %d", updatedUser.ID)
	}
}

func TestPostgresIntegrationQueryIncrementByAmount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	user := models.User{Name: "increment_user", ID: 1, Avatar: "group1"}
	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	res, err := db.Query().Table("users").Where("name = ?", "increment_user").Increment("id", 5)
	if err != nil {
		t.Errorf("Increment by amount failed: %v", err)
	}
	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	var updatedUser models.User
	err = db.Query().Table("users").Where("name = ?", "increment_user").First(&updatedUser)
	if err != nil {
		t.Errorf("Failed to find updated user: %v", err)
	}
	if updatedUser.ID != 6 {
		t.Errorf("Expected ID 6, got %d", updatedUser.ID)
	}
}

func TestPostgresIntegrationQueryDecrement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	user := models.User{Name: "increment_user", ID: 10, Avatar: "group1"}
	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	res, err := db.Query().Table("users").Where("name = ?", "increment_user").Decrement("id")
	if err != nil {
		t.Errorf("Decrement failed: %v", err)
	}
	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	var updatedUser models.User
	err = db.Query().Table("users").Where("name = ?", "increment_user").First(&updatedUser)
	if err != nil {
		t.Errorf("Failed to find updated user: %v", err)
	}
	if updatedUser.ID != 9 {
		t.Errorf("Expected ID 9, got %d", updatedUser.ID)
	}
}

func TestPostgresIntegrationQueryDecrementByAmount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	user := models.User{Name: "increment_user", ID: 10, Avatar: "group1"}
	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	res, err := db.Query().Table("users").Where("name = ?", "increment_user").Decrement("id", 3)
	if err != nil {
		t.Errorf("Decrement by amount failed: %v", err)
	}
	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	var updatedUser models.User
	err = db.Query().Table("users").Where("name = ?", "increment_user").First(&updatedUser)
	if err != nil {
		t.Errorf("Failed to find updated user: %v", err)
	}
	if updatedUser.ID != 7 {
		t.Errorf("Expected ID 7, got %d", updatedUser.ID)
	}
}

func TestPostgresIntegrationQueryIncrementDecrementWithWhereConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	user := models.User{Name: "increment_user", ID: 1, Avatar: "group1"}
	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	user2 := models.User{Name: "other_user", ID: 10, Avatar: "group2"}
	if err := db.Query().Model(&models.User{}).Create(&user2); err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	res, err := db.Query().Table("users").Where("avatar = ?", "group2").Increment("id", 10)
	if err != nil {
		t.Errorf("Increment with where failed: %v", err)
	}
	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	var updatedUser models.User
	err = db.Query().Table("users").Where("name = ?", "other_user").First(&updatedUser)
	if err != nil {
		t.Errorf("Failed to find updated user: %v", err)
	}
	if updatedUser.ID != 20 {
		t.Errorf("Expected ID 20, got %d", updatedUser.ID)
	}

	var firstUser models.User
	err = db.Query().Table("users").Where("name = ?", "increment_user").First(&firstUser)
	if err != nil {
		t.Errorf("Failed to find first user: %v", err)
	}
	if firstUser.ID != 1 {
		t.Errorf("Expected ID 1, got %d", firstUser.ID)
	}
}

func TestPostgresIntegrationQueryIncrementWithExtraColumns(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	user := models.User{Name: "increment_user", ID: 3, Avatar: "group1"}
	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	res, err := db.Query().Table("users").Where("name = ?", "increment_user").Increment("id", 1, map[string]any{"avatar": "updated_group"})
	if err != nil {
		t.Errorf("Increment with extra columns failed: %v", err)
	}
	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	var updatedUser models.User
	err = db.Query().Table("users").Where("name = ?", "increment_user").First(&updatedUser)
	if err != nil {
		t.Errorf("Failed to find updated user: %v", err)
	}
	if updatedUser.ID != 4 {
		t.Errorf("Expected ID 4, got %d", updatedUser.ID)
	}
	if updatedUser.Avatar != "updated_group" {
		t.Errorf("Expected avatar 'updated_group', got '%s'", updatedUser.Avatar)
	}
}

func TestPostgresIntegrationQueryIncrementDecrementInvalidColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)

	_, err := db.Query().Table("users").Increment("invalid; column")
	if err == nil {
		t.Error("Expected error for invalid column")
	}
}
