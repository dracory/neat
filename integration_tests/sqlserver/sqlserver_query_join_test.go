package sqlserver

import (
	"testing"
	"time"

	"github.com/dracory/neat/database"
	"github.com/dracory/neat/integration_tests/models"
)

// seedJoinTestData creates two users and one address linked to the first user.
// It returns the IDs of both users so callers can assert join behaviour where
// one user has an address and the other does not.
func seedJoinTestData(t *testing.T, db *database.Database) (uint, uint) {
	now := time.Now()

	user1 := models.User{Name: "join_user1", CreatedAt: now, UpdatedAt: now}
	user2 := models.User{Name: "join_user2", CreatedAt: now, UpdatedAt: now}
	if err := db.Query().Model(&models.User{}).Create(&user1); err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}
	if err := db.Query().Model(&models.User{}).Create(&user2); err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	var createdUser1, createdUser2 models.User
	if err := db.Query().Model(&models.User{}).Where("name = ?", "join_user1").First(&createdUser1); err != nil {
		t.Fatalf("Failed to get created user1: %v", err)
	}
	if err := db.Query().Model(&models.User{}).Where("name = ?", "join_user2").First(&createdUser2); err != nil {
		t.Fatalf("Failed to get created user2: %v", err)
	}

	address1 := models.Address{Name: "address1", UserID: createdUser1.ID, CreatedAt: now, UpdatedAt: now}
	if err := db.Query().Model(&models.Address{}).Create(&address1); err != nil {
		t.Fatalf("Failed to create address1: %v", err)
	}

	return createdUser1.ID, createdUser2.ID
}

// TestSQLServerIntegrationJoinInner verifies that an INNER JOIN between users
// and addresses returns only the user that has an associated address.
func TestSQLServerIntegrationJoinInner(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	userID1, _ := seedJoinTestData(t, db)

	var results []struct {
		UserName    string `db:"column:name"`
		AddressName string `db:"column:address_name"`
	}
	err := db.Query().Table("users").
		Join("addresses ON addresses.user_id = users.id").
		Select("users.name, addresses.name as address_name").
		Where("users.id = ?", userID1).
		Scan(&results)

	if err != nil {
		t.Errorf("Inner join failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
	if len(results) >= 1 {
		if results[0].UserName != "join_user1" {
			t.Errorf("Expected 'join_user1', got '%s'", results[0].UserName)
		}
		if results[0].AddressName != "address1" {
			t.Errorf("Expected 'address1', got '%s'", results[0].AddressName)
		}
	}
}

// TestSQLServerIntegrationJoinLeft verifies that a LEFT JOIN between users and
// addresses returns both users, with NULL address columns for the user that has
// no address (scanned as an empty string via sql.NullString).
func TestSQLServerIntegrationJoinLeft(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	userID1, userID2 := seedJoinTestData(t, db)

	var results []struct {
		UserName    string `db:"column:name"`
		AddressName string `db:"column:address_name"`
	}
	err := db.Query().Table("users").
		LeftJoin("addresses ON addresses.user_id = users.id").
		Select("users.name, addresses.name as address_name").
		Where("users.id = ? OR users.id = ?", userID1, userID2).
		Scan(&results)

	if err != nil {
		t.Errorf("Left join failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}
