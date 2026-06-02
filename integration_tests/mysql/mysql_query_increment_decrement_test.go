package mysql_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestMySQLIntegrationQueryIncrement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	query := db.Query()

	user := models.User{Name: "increment_user", Avatar: "group1", Votes: 10}
	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	var createdUser models.User
	if err := query.Model(&models.User{}).Where("name = ?", "increment_user").First(&createdUser); err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}

	// Increment votes column (valid use case)
	res, err := query.Model(&models.User{}).Where("id = ?", createdUser.ID).Increment("votes", 5)
	if err != nil {
		t.Fatalf("Failed to increment votes: %v", err)
	}

	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	var updatedUser models.User
	if err := query.Model(&models.User{}).Where("id = ?", createdUser.ID).First(&updatedUser); err != nil {
		t.Fatalf("Failed to get updated user: %v", err)
	}

	if updatedUser.Votes != 15 {
		t.Errorf("Expected votes to be 15, got %d", updatedUser.Votes)
	}
}

func TestMySQLIntegrationQueryDecrement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	query := db.Query()

	user := models.User{Name: "decrement_user", Avatar: "group1", Votes: 20}
	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	var createdUser models.User
	if err := query.Model(&models.User{}).Where("name = ?", "decrement_user").First(&createdUser); err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}

	// Decrement votes column (valid use case)
	res, err := query.Model(&models.User{}).Where("id = ?", createdUser.ID).Decrement("votes", 3)
	if err != nil {
		t.Fatalf("Failed to decrement votes: %v", err)
	}

	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	var updatedUser models.User
	if err := query.Model(&models.User{}).Where("id = ?", createdUser.ID).First(&updatedUser); err != nil {
		t.Fatalf("Failed to get updated user: %v", err)
	}

	if updatedUser.Votes != 17 {
		t.Errorf("Expected votes to be 17, got %d", updatedUser.Votes)
	}
}

func TestMySQLIntegrationQueryIncrementDecrementWithWhere(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	query := db.Query()

	user := models.User{Name: "group1_user", Avatar: "group1", Votes: 10}
	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	user2 := models.User{Name: "group2_user", Avatar: "group2", Votes: 20}
	if err := db.Query().Model(&models.User{}).Create(&user2); err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	// Increment votes for users in group1
	res, err := query.Model(&models.User{}).Where("avatar = ?", "group1").Increment("votes", 5)
	if err != nil {
		t.Fatalf("Failed to increment votes: %v", err)
	}

	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	var updatedUser models.User
	if err := query.Model(&models.User{}).Where("name = ?", "group1_user").First(&updatedUser); err != nil {
		t.Fatalf("Failed to get updated user: %v", err)
	}

	if updatedUser.Votes != 15 {
		t.Errorf("Expected votes to be 15, got %d", updatedUser.Votes)
	}

	// Decrement votes for users in group2
	res, err = query.Model(&models.User{}).Where("avatar = ?", "group2").Decrement("votes", 3)
	if err != nil {
		t.Fatalf("Failed to decrement votes: %v", err)
	}

	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	var updatedUser2 models.User
	if err := query.Model(&models.User{}).Where("name = ?", "group2_user").First(&updatedUser2); err != nil {
		t.Fatalf("Failed to get updated user2: %v", err)
	}

	if updatedUser2.Votes != 17 {
		t.Errorf("Expected votes to be 17, got %d", updatedUser2.Votes)
	}
}
