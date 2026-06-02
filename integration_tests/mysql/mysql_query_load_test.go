package mysql_test

import (
	"testing"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
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

	// Load the Books relationship
	err = query.Load(&foundUser, "Books")
	if err != nil {
		t.Errorf("Load failed: %v", err)
	}

	if foundUser.Books == nil {
		t.Error("Books should be loaded after Load()")
	}
	if len(foundUser.Books) != 2 {
		t.Errorf("Expected 2 books, got %d", len(foundUser.Books))
	}
}

func TestMySQLIntegrationQueryLoadWithConstraints(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	query := db.Query()
	userID := seedLoadTestData(t, db)

	var foundUser models.User
	err := query.Model(&models.User{}).Where("id = ?", userID).First(&foundUser)
	if err != nil {
		t.Errorf("Find failed: %v", err)
	}

	// Load the Books relationship with a constraint (only load books with name containing "1")
	err = query.Load(&foundUser, "Books", func(q contractsorm.Query) contractsorm.Query {
		return q.Where("name LIKE ?", "%1%")
	})
	if err != nil {
		t.Errorf("Load with constraints failed: %v", err)
	}

	if foundUser.Books == nil {
		t.Error("Books should be loaded after Load()")
	}
	if len(foundUser.Books) != 1 {
		t.Errorf("Expected 1 book with constraint, got %d", len(foundUser.Books))
	}
	if foundUser.Books[0].Name != "book 1" {
		t.Errorf("Expected 'book 1', got '%s'", foundUser.Books[0].Name)
	}
}

func TestMySQLIntegrationQueryLoadMissing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	query := db.Query()
	userID := seedLoadTestData(t, db)

	var foundUser models.User
	err := query.Model(&models.User{}).Where("id = ?", userID).First(&foundUser)
	if err != nil {
		t.Errorf("Find failed: %v", err)
	}

	// LoadMissing should load the relationship since it's not loaded yet
	err = query.LoadMissing(&foundUser, "Books")
	if err != nil {
		t.Errorf("LoadMissing failed: %v", err)
	}

	if foundUser.Books == nil {
		t.Error("Books should be loaded after LoadMissing()")
	}
	if len(foundUser.Books) != 2 {
		t.Errorf("Expected 2 books, got %d", len(foundUser.Books))
	}

	// Call LoadMissing again - should skip since already loaded
	err = query.LoadMissing(&foundUser, "Books")
	if err != nil {
		t.Errorf("LoadMissing on already loaded relation failed: %v", err)
	}

	// Count should still be 2 (not reloaded)
	if len(foundUser.Books) != 2 {
		t.Errorf("Expected 2 books after second LoadMissing, got %d", len(foundUser.Books))
	}
}
