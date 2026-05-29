//go:build integration

package mysql

import (
	"testing"

	"github.com/dracory/neat/database"
	"github.com/dracory/neat/integration_tests/models"
)

func seedLoadTestData(t *testing.T, db *database.Database) int64 {
	query := db.Query()
	user := models.User{Name: "load_user"}
	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	var createdUser models.User
	if err := query.Model(&models.User{}).Where("name = ?", "load_user").First(&createdUser); err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}

	book1 := models.Book{Name: "book 1", UserID: createdUser.ID}
	book2 := models.Book{Name: "book 2", UserID: createdUser.ID}
	if err := query.Model(&models.Book{}).Create(&book1); err != nil {
		t.Fatalf("Failed to create book 1: %v", err)
	}
	if err := query.Model(&models.Book{}).Create(&book2); err != nil {
		t.Fatalf("Failed to create book 2: %v", err)
	}

	return int64(createdUser.ID)
}

func TestMySQLIntegrationQueryLoadSingleRelationship(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Skip("Load method not currently supported in neat")

	db := SetupMySQLTest(t)
	query := db.Query()
	userID := seedLoadTestData(t, db)

	var foundUser models.User
	err := query.Model(&models.User{}).Where("id = ?", userID).First(&foundUser)
	if err != nil {
		t.Errorf("Find failed: %v", err)
	}
	if foundUser.Books != nil {
		t.Error("Books should be nil before load")
	}
}

func TestMySQLIntegrationQueryLoadWithConstraints(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Skip("Load method not currently supported in neat")

	db := SetupMySQLTest(t)
	seedLoadTestData(t, db)
}

func TestMySQLIntegrationQueryLoadMissing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Skip("LoadMissing method not currently supported in neat")

	db := SetupMySQLTest(t)
	seedLoadTestData(t, db)
}
