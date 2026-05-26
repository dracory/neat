package sqlite

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestSQLiteIntegrationQueryIncrement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	query := db.Query()

	user := models.User{Name: "increment_user", Votes: 10}
	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	res, err := db.Query().Table("users").Where("name = ?", "increment_user").Increment("votes")
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
	if updatedUser.Votes != 11 {
		t.Errorf("Expected 11 votes, got %d", updatedUser.Votes)
	}
}

func TestSQLiteIntegrationQueryIncrementByAmount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	query := db.Query()

	user := models.User{Name: "increment_user", Votes: 10}
	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	res, err := db.Query().Table("users").Where("name = ?", "increment_user").Increment("votes", int64(5))
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
	if updatedUser.Votes != 15 {
		t.Errorf("Expected 15 votes, got %d", updatedUser.Votes)
	}
}

func TestSQLiteIntegrationQueryDecrement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	query := db.Query()

	user := models.User{Name: "decrement_user", Votes: 10}
	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	res, err := db.Query().Table("users").Where("name = ?", "decrement_user").Decrement("votes")
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
	if updatedUser.Votes != 9 {
		t.Errorf("Expected 9 votes, got %d", updatedUser.Votes)
	}
}

func TestSQLiteIntegrationQueryDecrementByAmount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	query := db.Query()

	user := models.User{Name: "decrement_user", Votes: 10}
	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	res, err := db.Query().Table("users").Where("name = ?", "decrement_user").Decrement("votes", int64(3))
	if err != nil {
		t.Errorf("Decrement by amount failed: %v", err)
	}
	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	var updatedUser models.User
	err = db.Query().Table("users").Where("name = ?", "decrement_user").First(&updatedUser)
	if err != nil {
		t.Errorf("Failed to find updated user: %v", err)
	}
	if updatedUser.Votes != 7 {
		t.Errorf("Expected 7 votes, got %d", updatedUser.Votes)
	}
}

func TestSQLiteIntegrationQueryIncrementDecrementInvalidColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)

	_, err := db.Query().Table("users").Increment("invalid; column")
	if err == nil {
		t.Error("Expected error for invalid column")
	}
}
