//go:build disabled

package postgres

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestPostgreSQLIntegrationQueryLoadSingleRelationship(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
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

	var foundUser models.User
	err := query.Model(&models.User{}).Where("id = ?", createdUser.ID).First(&foundUser)
	if err != nil {
		t.Errorf("Find failed: %v", err)
	}
	if foundUser.Books != nil {
		t.Error("Books should be nil before load")
	}

	t.Skip("Load method not currently supported in neat")
}

func TestPostgreSQLIntegrationQueryLoadWithConstraints(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
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

	t.Skip("Load method not currently supported in neat")
}

func TestPostgreSQLIntegrationQueryLoadMissing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
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

	t.Skip("LoadMissing method not currently supported in neat")
}
