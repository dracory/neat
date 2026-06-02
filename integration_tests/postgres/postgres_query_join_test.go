package postgres_test

import (
	"testing"
	"time"

	"github.com/dracory/neat/integration_tests/common"
	"github.com/dracory/neat/integration_tests/models"
)

func TestPostgresIntegrationInnerJoin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestJoinInner(t, db)
}

func TestPostgresIntegrationInnerJoinWithConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestJoinInnerWithConditions(t, db)
}

func TestPostgresIntegrationInnerJoinWithAliases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestJoinInnerWithAliases(t, db)
}

func TestPostgresIntegrationLeftJoin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestJoinLeft(t, db)
}

func TestPostgresIntegrationLeftJoinWithConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestJoinLeftWithConditions(t, db)
}

func TestPostgresIntegrationLeftJoinWithAliases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestJoinLeftWithAliases(t, db)
}

func TestPostgresIntegrationRightJoin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestJoinRight(t, db)
}

func TestPostgresIntegrationRightJoinWithConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestJoinRightWithConditions(t, db)
}

func TestPostgresIntegrationRightJoinWithAliases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestJoinRightWithAliases(t, db)
}

func TestPostgresIntegrationCrossJoin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestJoinCross(t, db)
}

func TestPostgresIntegrationCrossJoinWithConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestJoinCrossWithConditions(t, db)
}

func TestPostgresIntegrationCrossJoinWithSelect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestJoinCrossWithSelect(t, db)
}

func TestPostgresIntegrationMultipleJoins(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	userID1, _ := common.SeedJoinTestData(t, db)

	now := time.Now()
	book1 := models.Book{Name: "book1", UserID: userID1, CreatedAt: now, UpdatedAt: now}
	if err := db.Query().Model(&models.Book{}).Create(&book1); err != nil {
		t.Fatalf("Failed to create book1: %v", err)
	}

	var results []struct {
		UserName    string `db:"column:user_name"`
		AddressName string `db:"column:address_name"`
		BookName    string `db:"column:book_name"`
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
	if len(results) >= 1 {
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
}

func TestPostgresIntegrationJoinChaining(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.SeedJoinTestData(t, db)

	var results []struct {
		UserName string `db:"column:name"`
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
}

func TestPostgresIntegrationComplexJoinScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.SeedJoinTestData(t, db)

	var results []struct {
		UserName string `db:"column:name"`
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
}
