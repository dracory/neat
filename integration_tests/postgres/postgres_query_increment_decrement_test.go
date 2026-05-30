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

	user := models.User{Name: "increment_user", Avatar: "group1"}
	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	originalID := user.ID
	res, err := db.Query().Table("users").Where("id = ?", originalID).Increment("id")
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
	if updatedUser.ID != originalID+1 {
		t.Errorf("Expected ID %d, got %d", originalID+1, updatedUser.ID)
	}
}

func TestPostgresIntegrationQueryIncrementByAmount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	user := models.User{Name: "increment_user_by_amount", Avatar: "group1"}
	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	originalID := user.ID
	res, err := db.Query().Table("users").Where("name = ?", "increment_user_by_amount").Increment("id", 5)
	if err != nil {
		t.Errorf("Increment by amount failed: %v", err)
	}
	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	var updatedUser models.User
	err = db.Query().Table("users").Where("name = ?", "increment_user_by_amount").First(&updatedUser)
	if err != nil {
		t.Errorf("Failed to find updated user: %v", err)
	}
	if updatedUser.ID != originalID+5 {
		t.Errorf("Expected ID %d, got %d", originalID+5, updatedUser.ID)
	}
}

func TestPostgresIntegrationQueryDecrement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	user := models.User{Name: "decrement_user", Avatar: "group1"}
	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	originalID := user.ID
	res, err := db.Query().Table("users").Where("name = ?", "decrement_user").Decrement("id")
	if err != nil {
		t.Errorf("Decrement failed: %v", err)
	}
	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	var updatedUser models.User
	err = db.Query().Table("users").Where("name = ?", "decrement_user").First(&updatedUser)
	if err != nil {
		t.Errorf("Failed to find updated user: %v", err)
	}
	if updatedUser.ID != originalID-1 {
		t.Errorf("Expected ID %d, got %d", originalID-1, updatedUser.ID)
	}
}

func TestPostgresIntegrationQueryDecrementByAmount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	user := models.User{Name: "decrement_user_by_amount", Avatar: "group1"}
	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	originalID := user.ID
	res, err := db.Query().Table("users").Where("name = ?", "decrement_user_by_amount").Decrement("id", 3)
	if err != nil {
		t.Errorf("Decrement by amount failed: %v", err)
	}
	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	var updatedUser models.User
	err = db.Query().Table("users").Where("name = ?", "decrement_user_by_amount").First(&updatedUser)
	if err != nil {
		t.Errorf("Failed to find updated user: %v", err)
	}
	if updatedUser.ID != originalID-3 {
		t.Errorf("Expected ID %d, got %d", originalID-3, updatedUser.ID)
	}
}

func TestPostgresIntegrationQueryIncrementDecrementWithWhereConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	user := models.User{Name: "increment_where_user", Avatar: "group1"}
	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	user2 := models.User{Name: "other_where_user", Avatar: "group2"}
	if err := db.Query().Model(&models.User{}).Create(&user2); err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	originalID := user2.ID
	res, err := db.Query().Table("users").Where("avatar = ?", "group2").Increment("id", 10)
	if err != nil {
		t.Errorf("Increment with where failed: %v", err)
	}
	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	var updatedUser models.User
	err = db.Query().Table("users").Where("name = ?", "other_where_user").First(&updatedUser)
	if err != nil {
		t.Errorf("Failed to find updated user: %v", err)
	}
	if updatedUser.ID != originalID+10 {
		t.Errorf("Expected ID %d, got %d", originalID+10, updatedUser.ID)
	}

	var firstUser models.User
	err = db.Query().Table("users").Where("name = ?", "increment_where_user").First(&firstUser)
	if err != nil {
		t.Errorf("Failed to find first user: %v", err)
	}
	if firstUser.ID != user.ID {
		t.Errorf("Expected ID %d, got %d", user.ID, firstUser.ID)
	}
}

func TestPostgresIntegrationQueryIncrementWithExtraColumns(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Skipping Increment with extra columns - feature not implemented")
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
