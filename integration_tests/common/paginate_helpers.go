package common

import (
	"testing"

	contractsdb "github.com/dracory/neat/contracts/database"
	"github.com/dracory/neat/database"
	"github.com/dracory/neat/integration_tests/models"
)

// SeedPaginateTestData creates 15 test users for pagination tests
func SeedPaginateTestData(t *testing.T, db *database.Database) {
	query := db.Query()
	// Clean up existing test data first
	// Oracle has issues with Delete() method when using WHERE clauses, so use raw SQL for Oracle
	if query.Driver() == contractsdb.DriverOracle {
		sqlDB, err := db.DB()
		if err != nil {
			t.Fatalf("Failed to get DB connection: %v", err)
		}
		_, err = sqlDB.Exec(`DELETE FROM USERS WHERE NAME LIKE 'paginate_user_%'`)
		if err != nil {
			t.Logf("Warning: failed to cleanup test data with raw SQL: %v", err)
		}
	} else {
		if _, err := query.Model(&models.User{}).Where("name LIKE ?", "paginate_user_%").Delete(); err != nil {
			t.Logf("Warning: failed to cleanup test data: %v", err)
		}
	}

	for i := 1; i <= 15; i++ {
		user := models.User{
			Name:   "paginate_user_" + string(rune(64+i)),
			Avatar: "avatar",
		}
		if err := query.Model(&models.User{}).Create(&user); err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}
}

// TestPaginateFirstPage tests pagination of the first page
func TestPaginateFirstPage(t *testing.T, db *database.Database) {
	SeedPaginateTestData(t, db)

	var users []models.User
	var total int64
	err := db.Query().Model(&models.User{}).OrderBy("name", "asc").Paginate(1, 5, &users, &total)
	if err != nil {
		t.Errorf("Paginate failed: %v", err)
	}
	if total != 15 {
		t.Errorf("Expected total 15, got %d", total)
	}
	if len(users) != 5 {
		t.Errorf("Expected 5 users, got %d", len(users))
	}
	if len(users) >= 5 {
		if users[0].Name != "paginate_user_A" {
			t.Errorf("Expected 'paginate_user_A', got '%s'", users[0].Name)
		}
		if users[4].Name != "paginate_user_E" {
			t.Errorf("Expected 'paginate_user_E', got '%s'", users[4].Name)
		}
	}
}

// TestPaginateSecondPage tests pagination of the second page
func TestPaginateSecondPage(t *testing.T, db *database.Database) {
	SeedPaginateTestData(t, db)

	var users []models.User
	var total int64
	err := db.Query().Model(&models.User{}).OrderBy("name", "asc").Paginate(2, 5, &users, &total)
	if err != nil {
		t.Errorf("Paginate failed: %v", err)
	}
	if total != 15 {
		t.Errorf("Expected total 15, got %d", total)
	}
	if len(users) != 5 {
		t.Errorf("Expected 5 users, got %d", len(users))
	}
	if len(users) >= 5 {
		if users[0].Name != "paginate_user_F" {
			t.Errorf("Expected 'paginate_user_F', got '%s'", users[0].Name)
		}
		if users[4].Name != "paginate_user_J" {
			t.Errorf("Expected 'paginate_user_J', got '%s'", users[4].Name)
		}
	}
}

// TestPaginateWithConditions tests pagination with WHERE conditions
func TestPaginateWithConditions(t *testing.T, db *database.Database) {
	SeedPaginateTestData(t, db)

	var users []models.User
	var total int64
	err := db.Query().Model(&models.User{}).Where("name <= ?", "paginate_user_E").OrderBy("name", "asc").Paginate(1, 3, &users, &total)
	if err != nil {
		t.Errorf("Paginate with conditions failed: %v", err)
	}
	if total != 5 {
		t.Errorf("Expected total 5, got %d", total)
	}
	if len(users) != 3 {
		t.Errorf("Expected 3 users, got %d", len(users))
	}
}

// TestPaginateWithSelectAliases tests pagination with SELECT aliases
func TestPaginateWithSelectAliases(t *testing.T, db *database.Database) {
	SeedPaginateTestData(t, db)

	var results []struct {
		UserName string
	}
	var total int64
	err := db.Query().Table("users").Select("name as user_name").OrderBy("name", "asc").Paginate(1, 5, &results, &total)
	if err != nil {
		t.Errorf("Paginate with Select aliases failed: %v", err)
	}
	if total != 15 {
		t.Errorf("Expected total 15, got %d", total)
	}
	if len(results) != 5 {
		t.Errorf("Expected 5 results, got %d", len(results))
	}
	if len(results) > 0 && results[0].UserName != "paginate_user_A" {
		t.Errorf("Expected 'paginate_user_A', got '%s'", results[0].UserName)
	}
}

// TestPaginateLastPage tests pagination of the last page
func TestPaginateLastPage(t *testing.T, db *database.Database) {
	SeedPaginateTestData(t, db)

	var users []models.User
	var total int64
	err := db.Query().Model(&models.User{}).OrderBy("name", "asc").Paginate(3, 5, &users, &total)
	if err != nil {
		t.Errorf("Paginate last page failed: %v", err)
	}
	if total != 15 {
		t.Errorf("Expected total 15, got %d", total)
	}
	if len(users) != 5 {
		t.Errorf("Expected 5 users, got %d", len(users))
	}
	if len(users) >= 5 {
		if users[0].Name != "paginate_user_K" {
			t.Errorf("Expected 'paginate_user_K', got '%s'", users[0].Name)
		}
		if users[4].Name != "paginate_user_O" {
			t.Errorf("Expected 'paginate_user_O', got '%s'", users[4].Name)
		}
	}
}

// TestPaginatePageBeyondBounds tests pagination with a page number beyond available data
func TestPaginatePageBeyondBounds(t *testing.T, db *database.Database) {
	SeedPaginateTestData(t, db)

	var users []models.User
	var total int64
	err := db.Query().Model(&models.User{}).Paginate(10, 5, &users, &total)
	if err != nil {
		t.Errorf("Pagination page beyond bounds failed: %v", err)
	}
	if total != 15 {
		t.Errorf("Expected total 15, got %d", total)
	}
	if len(users) != 0 {
		t.Errorf("Expected 0 users, got %d", len(users))
	}
}

// TestPaginateEmptyResults tests pagination with a condition that returns no results
func TestPaginateEmptyResults(t *testing.T, db *database.Database) {
	SeedPaginateTestData(t, db)

	var users []models.User
	var total int64
	err := db.Query().Model(&models.User{}).Where("name = ?", "non_existent").Paginate(1, 5, &users, &total)
	if err != nil {
		t.Errorf("Pagination empty results failed: %v", err)
	}
	if total != 0 {
		t.Errorf("Expected total 0, got %d", total)
	}
	if len(users) != 0 {
		t.Errorf("Expected 0 users, got %d", len(users))
	}
}

// TestCountWithSelectAlias tests COUNT with SELECT alias
func TestCountWithSelectAlias(t *testing.T, db *database.Database) {
	SeedPaginateTestData(t, db)

	var count int64
	err := db.Query().Table("users").Select("name as user_name").Count(&count)
	if err != nil {
		t.Errorf("Count with Select alias failed: %v", err)
	}
	if count != 15 {
		t.Errorf("Expected count 15, got %d", count)
	}
}
