//go:build integration

package postgres

import (
	"testing"

	"github.com/dracory/neat/database"
	"github.com/dracory/neat/integration_tests/models"
)

func TestPostgresIntegrationJoin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	// Seed data
	user1 := models.User{Name: "join_user1"}
	user2 := models.User{Name: "join_user2"}
	if err := query.Model(&models.User{}).Create(&user1); err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}
	if err := query.Model(&models.User{}).Create(&user2); err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	address1 := models.Address{Name: "address1", UserID: user1.ID}
	if err := query.Create(&address1); err != nil {
		t.Fatalf("Failed to create address1: %v", err)
	}

	t.Run("Inner Join", func(t *testing.T) {
		var results []struct {
			UserName    string `gorm:"column:name"`
			AddressName string `gorm:"column:address_name"`
		}
		err := db.Query().Table("users").
			Join("addresses ON addresses.user_id = users.id").
			Select("users.name, addresses.name as address_name").
			Where("users.id = ?", user1.ID).
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
	})

	t.Run("Inner Join with conditions", func(t *testing.T) {
		var results []struct {
			UserName string `gorm:"column:name"`
		}
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
	})

	t.Run("Inner Join with aliases", func(t *testing.T) {
		var results []struct {
			UserName string `gorm:"column:name"`
		}
		err := db.Query().Table("users as u").
			Join("addresses as a ON a.user_id = u.id").
			Select("u.name").
			Scan(&results)

		if err != nil {
			t.Errorf("Inner Join with aliases failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
	})

	t.Run("Left Join", func(t *testing.T) {
		var results []struct {
			UserName    string  `gorm:"column:name"`
			AddressName *string `gorm:"column:address_name"`
		}
		err := db.Query().Table("users").
			LeftJoin("addresses ON addresses.user_id = users.id").
			Select("users.name, addresses.name as address_name").
			OrderBy("users.name", "asc").
			Scan(&results)

		if err != nil {
			t.Errorf("Left Join failed: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}
		if results[0].UserName != "join_user1" {
			t.Errorf("Expected 'join_user1', got '%s'", results[0].UserName)
		}
		if results[0].AddressName == nil {
			t.Error("Expected AddressName to be non-nil for join_user1")
		}
		if results[1].UserName != "join_user2" {
			t.Errorf("Expected 'join_user2', got '%s'", results[1].UserName)
		}
		if results[1].AddressName != nil {
			t.Error("Expected AddressName to be nil for join_user2")
		}
	})

	t.Run("Left Join with conditions", func(t *testing.T) {
		var results []struct {
			UserName    string  `gorm:"column:name"`
			AddressName *string `gorm:"column:address_name"`
		}
		err := db.Query().Table("users").
			LeftJoin("addresses ON addresses.user_id = users.id AND addresses.name = ?", "non-existent").
			Select("users.name, addresses.name as address_name").
			OrderBy("users.name", "asc").
			Scan(&results)

		if err != nil {
			t.Errorf("Left Join with conditions failed: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}
		if results[0].AddressName != nil {
			t.Error("Expected AddressName to be nil")
		}
		if results[1].AddressName != nil {
			t.Error("Expected AddressName to be nil")
		}
	})

	t.Run("Left Join with aliases", func(t *testing.T) {
		var results []struct {
			UserName string `gorm:"column:name"`
		}
		err := db.Query().Table("users as u").
			LeftJoin("addresses as a ON a.user_id = u.id").
			Select("u.name").
			Scan(&results)

		if err != nil {
			t.Errorf("Left Join with aliases failed: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}
	})

	t.Run("Right Join", func(t *testing.T) {
		var results []struct {
			UserName    *string `gorm:"column:name"`
			AddressName string  `gorm:"column:address_name"`
		}
		err := db.Query().Table("users").
			RightJoin("addresses ON addresses.user_id = users.id").
			Select("users.name, addresses.name as address_name").
			Scan(&results)

		if err != nil {
			t.Errorf("Right Join failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		if results[0].UserName == nil || *results[0].UserName != "join_user1" {
			t.Errorf("Expected 'join_user1', got '%v'", results[0].UserName)
		}
		if results[0].AddressName != "address1" {
			t.Errorf("Expected 'address1', got '%s'", results[0].AddressName)
		}
	})

	t.Run("Right Join with conditions", func(t *testing.T) {
		var results []struct {
			AddressName string `gorm:"column:address_name"`
		}
		err := db.Query().Table("users").
			RightJoin("addresses ON addresses.user_id = users.id AND users.name = ?", "non-existent").
			Select("addresses.name as address_name").
			Scan(&results)

		if err != nil {
			t.Errorf("Right Join with conditions failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
	})

	t.Run("Right Join with aliases", func(t *testing.T) {
		var results []struct {
			AddressName string `gorm:"column:address_name"`
		}
		err := db.Query().Table("users as u").
			RightJoin("addresses as a ON a.user_id = u.id").
			Select("a.name as address_name").
			Scan(&results)

		if err != nil {
			t.Errorf("Right Join with aliases failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
	})

	t.Run("Cross Join", func(t *testing.T) {
		var results []struct {
			UserName    string `gorm:"column:user_name"`
			AddressName string `gorm:"column:address_name"`
		}
		err := db.Query().Table("users").
			CrossJoin("addresses").
			Select("users.name as user_name, addresses.name as address_name").
			Scan(&results)

		if err != nil {
			t.Errorf("Cross Join failed: %v", err)
		}
		// 2 users * 1 address = 2 results
		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}
	})

	t.Run("Cross Join with conditions", func(t *testing.T) {
		var results []struct {
			UserName string `gorm:"column:user_name"`
		}
		err := db.Query().Table("users").
			CrossJoin("addresses").
			Where("addresses.user_id = users.id").
			Select("users.name as user_name").
			Scan(&results)

		if err != nil {
			t.Errorf("Cross Join with conditions failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
	})

	t.Run("Cross Join with Select", func(t *testing.T) {
		var results []struct {
			UserName string `gorm:"column:name"`
		}
		err := db.Query().Table("users").
			CrossJoin("addresses").
			Select("users.name").
			Scan(&results)

		if err != nil {
			t.Errorf("Cross Join with Select failed: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}
	})

	t.Run("Multiple joins", func(t *testing.T) {
		book1 := models.Book{Name: "book1", UserID: user1.ID}
		if err := query.Create(&book1); err != nil {
			t.Fatalf("Failed to create book1: %v", err)
		}

		var results []struct {
			UserName    string `gorm:"column:user_name"`
			AddressName string `gorm:"column:address_name"`
			BookName    string `gorm:"column:book_name"`
		}
		err := db.Query().Table("users").
			Join("addresses ON addresses.user_id = users.id").
			Join("books ON books.user_id = users.id").
			Select("users.name as user_name, addresses.name as address_name, books.name as book_name").
			Scan(&results)

		if err != nil {
			t.Errorf("Multiple joins failed: %v", err)
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
	})

	t.Run("Join chaining", func(t *testing.T) {
		var results []struct {
			UserName string `gorm:"column:name"`
		}
		err := db.Query().Table("users").
			Join("addresses ON addresses.user_id = users.id").
			LeftJoin("books ON books.user_id = users.id").
			Select("users.name").
			Scan(&results)

		if err != nil {
			t.Errorf("Join chaining failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
	})

	t.Run("Complex join scenarios", func(t *testing.T) {
		var results []struct {
			UserName string `gorm:"column:name"`
		}
		err := db.Query().Table("users").
			Join("addresses ON addresses.user_id = users.id").
			LeftJoin("books ON books.user_id = users.id").
			Where("addresses.name = ?", "address1").
			OrWhere("books.name = ?", "non-existent").
			Select("users.name").
			Scan(&results)

		if err != nil {
			t.Errorf("Complex join scenarios failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
	})
}
