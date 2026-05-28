package turso

import (
	"testing"

	"github.com/dracory/neat/database"
	"github.com/dracory/neat/integration_tests/models"
)

type joinTestData struct {
	user1 models.User
	user2 models.User
}

func seedJoinTestData(t *testing.T, db *database.Database) joinTestData {
	user1 := models.User{Name: "join_user1"}
	user2 := models.User{Name: "join_user2"}
	if err := db.Query().Model(&models.User{}).Create(&user1); err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}
	if err := db.Query().Model(&models.User{}).Create(&user2); err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	var dbUser1 models.User
	if err := db.Query().Model(&models.User{}).Where("name = ?", "join_user1").First(&dbUser1); err != nil {
		t.Fatalf("Failed to get user1 ID: %v", err)
	}
	user1 = dbUser1

	address1 := models.Address{Name: "address1", UserID: user1.ID}
	if err := db.Query().Table("addresses").Create(&address1); err != nil {
		t.Fatalf("Failed to create address1: %v", err)
	}

	return joinTestData{user1: user1, user2: user2}
}

func TestTursoIntegrationJoinInner(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	data := seedJoinTestData(t, db)

	results := make([]struct {
		UserName    string `gorm:"column:name"`
		AddressName string `gorm:"column:address_name"`
	}, 0)
	err := db.Query().Table("users").
		Join("addresses ON addresses.user_id = users.id").
		Select("users.name, addresses.name as address_name").
		Where("users.id = ?", data.user1.ID).
		Scan(&results)

	if err != nil {
		t.Errorf("Inner Join failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
	if results[0].UserName != "join_user1" {
		t.Errorf("Expected 'join_user1', got '%s'", results[0].UserName)
	}
	if results[0].AddressName != "address1" {
		t.Errorf("Expected 'address1', got '%s'", results[0].AddressName)
	}
}

func TestTursoIntegrationJoinInnerWithConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	seedJoinTestData(t, db)

	results := make([]struct {
		UserName string `gorm:"column:name"`
	}, 0)
	err := db.Query().Table("users").
		Join("addresses ON addresses.user_id = users.id AND addresses.name = ?", "address1").
		Select("users.name").
		Scan(&results)

	if err != nil {
		t.Errorf("Inner Join with conditions failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

func TestTursoIntegrationJoinLeft(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	data := seedJoinTestData(t, db)

	results := make([]struct {
		UserName    string  `gorm:"column:name"`
		AddressName *string `gorm:"column:address_name"`
	}, 0)
	err := db.Query().Table("users").
		LeftJoin("addresses ON addresses.user_id = users.id").
		Select("users.name, addresses.name as address_name").
		Where("users.id = ?", data.user2.ID).
		Scan(&results)

	if err != nil {
		t.Errorf("Left Join failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
	if results[0].UserName != "join_user2" {
		t.Errorf("Expected 'join_user2', got '%s'", results[0].UserName)
	}
	if results[0].AddressName != nil {
		t.Errorf("Expected nil address_name for user without address, got '%s'", *results[0].AddressName)
	}
}

func TestTursoIntegrationJoinMultiple(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	data := seedJoinTestData(t, db)

	// Add a book for user1
	book1 := models.Book{Name: "book1", UserID: data.user1.ID}
	if err := db.Query().Table("books").Create(&book1); err != nil {
		t.Fatalf("Failed to create book1: %v", err)
	}

	results := make([]struct {
		UserName    string `gorm:"column:name"`
		AddressName string `gorm:"column:address_name"`
		BookName    string `gorm:"column:book_name"`
	}, 0)
	err := db.Query().Table("users").
		Join("addresses ON addresses.user_id = users.id").
		Join("books ON books.user_id = users.id").
		Select("users.name, addresses.name as address_name, books.name as book_name").
		Where("users.id = ?", data.user1.ID).
		Scan(&results)

	if err != nil {
		t.Errorf("Multiple Join failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
	if results[0].UserName != "join_user1" {
		t.Errorf("Expected 'join_user1', got '%s'", results[0].UserName)
	}
	if results[0].AddressName != "address1" {
		t.Errorf("Expected 'address1', got '%s'", results[0].AddressName)
	}
	if results[0].BookName != "book1" {
		t.Errorf("Expected 'book1', got '%s'", results[0].BookName)
	}
}
