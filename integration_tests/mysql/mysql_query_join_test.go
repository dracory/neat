//go:build integration

package mysql

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestMySQLIntegrationJoin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	if db == nil {
		t.Skip("MySQL not available")
	}
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

	// Get the created users to get their IDs
	var createdUser1, createdUser2 models.User
	if err := query.Model(&models.User{}).Where("name = ?", "join_user1").First(&createdUser1); err != nil {
		t.Fatalf("Failed to get created user1: %v", err)
	}
	if err := query.Model(&models.User{}).Where("name = ?", "join_user2").First(&createdUser2); err != nil {
		t.Fatalf("Failed to get created user2: %v", err)
	}

	address1 := models.Address{Name: "address1", UserID: createdUser1.ID}
	if err := query.Model(&models.Address{}).Create(&address1); err != nil {
		t.Fatalf("Failed to create address1: %v", err)
	}

	t.Run("Inner Join", func(t *testing.T) {
		var results []struct {
			UserName    string `db:"column:name"`
			AddressName string `db:"column:address_name"`
		}
		err := db.Query().Table("users").
			Join("addresses ON addresses.user_id = users.id").
			Select("users.name, addresses.name as address_name").
			Where("users.id = ?", createdUser1.ID).
			Scan(&results)

		if err != nil {
			t.Errorf("Inner join failed: %v", err)
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
			UserName string `db:"column:name"`
		}
		err := db.Query().Table("users").
			Join("addresses ON addresses.user_id = users.id AND addresses.name = ?", "address1").
			Select("users.name").
			Scan(&results)

		if err != nil {
			t.Errorf("Inner join with conditions failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
	})

	t.Run("Inner Join with aliases", func(t *testing.T) {
		var results []struct {
			UserName string `db:"column:name"`
		}
		err := db.Query().Table("users as u").
			Join("addresses as a ON a.user_id = u.id").
			Select("u.name").
			Scan(&results)

		if err != nil {
			t.Errorf("Inner join with aliases failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
	})

	t.Run("Left Join", func(t *testing.T) {
		var results []struct {
			UserName    string  `db:"column:name"`
			AddressName *string `db:"column:address_name"`
		}
		err := db.Query().Table("users").
			LeftJoin("addresses ON addresses.user_id = users.id").
			Select("users.name, addresses.name as address_name").
			OrderBy("users.name", "asc").
			Scan(&results)

		if err != nil {
			t.Errorf("Left join failed: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}
		if results[0].UserName != "join_user1" {
			t.Errorf("Expected 'join_user1', got '%s'", results[0].UserName)
		}
		if results[0].AddressName == nil {
			t.Error("Expected address name to be set")
		}
		if results[1].UserName != "join_user2" {
			t.Errorf("Expected 'join_user2', got '%s'", results[1].UserName)
		}
		if results[1].AddressName != nil {
			t.Error("Expected address name to be nil")
		}
	})

	t.Run("Left Join with conditions", func(t *testing.T) {
		var results []struct {
			UserName    string  `db:"column:name"`
			AddressName *string `db:"column:address_name"`
		}
		err := db.Query().Table("users").
			LeftJoin("addresses ON addresses.user_id = users.id AND addresses.name = ?", "non-existent").
			Select("users.name, addresses.name as address_name").
			OrderBy("users.name", "asc").
			Scan(&results)

		if err != nil {
			t.Errorf("Left join with conditions failed: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}
		if results[0].AddressName != nil {
			t.Error("Expected address name to be nil")
		}
		if results[1].AddressName != nil {
			t.Error("Expected address name to be nil")
		}
	})

	t.Run("Left Join with aliases", func(t *testing.T) {
		var results []struct {
			UserName string `db:"column:name"`
		}
		err := db.Query().Table("users as u").
			LeftJoin("addresses as a ON a.user_id = u.id").
			Select("u.name").
			Scan(&results)

		if err != nil {
			t.Errorf("Left join with aliases failed: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}
	})

	t.Run("Right Join", func(t *testing.T) {
		var results []struct {
			UserName    *string `db:"column:name"`
			AddressName string  `db:"column:address_name"`
		}
		err := db.Query().Table("users").
			RightJoin("addresses ON addresses.user_id = users.id").
			Select("users.name, addresses.name as address_name").
			Scan(&results)

		if err != nil {
			t.Errorf("Right join failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		if results[0].UserName == nil {
			t.Error("Expected user name to be set")
		}
		if *results[0].UserName != "join_user1" {
			t.Errorf("Expected 'join_user1', got '%s'", *results[0].UserName)
		}
		if results[0].AddressName != "address1" {
			t.Errorf("Expected 'address1', got '%s'", results[0].AddressName)
		}
	})

	t.Run("Right Join with conditions", func(t *testing.T) {
		var results []struct {
			AddressName string `db:"column:address_name"`
		}
		err := db.Query().Table("users").
			RightJoin("addresses ON addresses.user_id = users.id AND users.name = ?", "non-existent").
			Select("addresses.name as address_name").
			Scan(&results)

		if err != nil {
			t.Errorf("Right join with conditions failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
	})

	t.Run("Right Join with aliases", func(t *testing.T) {
		var results []struct {
			AddressName string `db:"column:address_name"`
		}
		err := db.Query().Table("users as u").
			RightJoin("addresses as a ON a.user_id = u.id").
			Select("a.name as address_name").
			Scan(&results)

		if err != nil {
			t.Errorf("Right join with aliases failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
	})

	t.Run("Cross Join", func(t *testing.T) {
		var results []struct {
			UserName    string `db:"column:user_name"`
			AddressName string `db:"column:address_name"`
		}
		err := db.Query().Table("users").
			CrossJoin("addresses").
			Select("users.name as user_name, addresses.name as address_name").
			Scan(&results)

		if err != nil {
			t.Errorf("Cross join failed: %v", err)
		}
		// 2 users * 1 address = 2 results
		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}
	})

	t.Run("Cross Join with conditions", func(t *testing.T) {
		var results []struct {
			UserName string `db:"column:user_name"`
		}
		err := db.Query().Table("users").
			CrossJoin("addresses").
			Where("addresses.user_id = users.id").
			Select("users.name as user_name").
			Scan(&results)

		if err != nil {
			t.Errorf("Cross join with conditions failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
	})

	t.Run("Cross Join with Select", func(t *testing.T) {
		var results []struct {
			UserName string `db:"column:name"`
		}
		err := db.Query().Table("users").
			CrossJoin("addresses").
			Select("users.name").
			Scan(&results)

		if err != nil {
			t.Errorf("Cross join with select failed: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}
	})
}
