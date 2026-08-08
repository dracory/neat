package common

import (
	"testing"
	"time"

	"github.com/dracory/neat/database"
	"github.com/dracory/neat/integration_tests/models"
)

// SeedDottedColumnData creates users and addresses for dotted column tests.
// Returns the IDs of user1 (with address) and user3 (no address).
func SeedDottedColumnData(t *testing.T, db *database.Database) (uint, uint, uint) {
	now := time.Now()

	user1 := models.User{Name: "dotted_alice", CreatedAt: now, UpdatedAt: now}
	user2 := models.User{Name: "dotted_bob", CreatedAt: now, UpdatedAt: now}
	user3 := models.User{Name: "dotted_carol", CreatedAt: now, UpdatedAt: now}
	if err := db.Query().Model(&models.User{}).Create(&user1); err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}
	if err := db.Query().Model(&models.User{}).Create(&user2); err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}
	if err := db.Query().Model(&models.User{}).Create(&user3); err != nil {
		t.Fatalf("Failed to create user3: %v", err)
	}

	addr1 := models.Address{Name: "dotted_addr1", UserID: user1.ID, CreatedAt: now, UpdatedAt: now}
	addr2 := models.Address{Name: "dotted_addr2", UserID: user2.ID, CreatedAt: now, UpdatedAt: now}
	if err := db.Query().Model(&models.Address{}).Create(&addr1); err != nil {
		t.Fatalf("Failed to create addr1: %v", err)
	}
	if err := db.Query().Model(&models.Address{}).Create(&addr2); err != nil {
		t.Fatalf("Failed to create addr2: %v", err)
	}

	return user1.ID, user2.ID, user3.ID
}

// SeedDottedColumnGroupData creates users with multiple addresses for
// GROUP BY + HAVING tests. user1 gets 1 address, user2 gets 3 addresses.
func SeedDottedColumnGroupData(t *testing.T, db *database.Database) (uint, uint) {
	now := time.Now()

	user1 := models.User{Name: "dotted_group_a", CreatedAt: now, UpdatedAt: now}
	user2 := models.User{Name: "dotted_group_b", CreatedAt: now, UpdatedAt: now}
	if err := db.Query().Model(&models.User{}).Create(&user1); err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}
	if err := db.Query().Model(&models.User{}).Create(&user2); err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	// 1 address for user1
	addr1 := models.Address{Name: "ga1", UserID: user1.ID, CreatedAt: now, UpdatedAt: now}
	if err := db.Query().Model(&models.Address{}).Create(&addr1); err != nil {
		t.Fatalf("Failed to create addr1: %v", err)
	}

	// 3 addresses for user2
	for i := 0; i < 3; i++ {
		addr := models.Address{Name: "gb", UserID: user2.ID, CreatedAt: now, UpdatedAt: now}
		if err := db.Query().Model(&models.Address{}).Create(&addr); err != nil {
			t.Fatalf("Failed to create address: %v", err)
		}
	}

	return user1.ID, user2.ID
}

// TestDottedColumnOrderBy tests ORDER BY with a table.column reference
// in a JOIN query. Verifies results are sorted by the dotted column.
func TestDottedColumnOrderBy(t *testing.T, db *database.Database) {
	SeedDottedColumnData(t, db)

	type Result struct {
		UserName    string `db:"user_name"`
		AddressName string `db:"address_name"`
	}

	var results []Result
	err := db.Query().Table("users").
		Join("addresses ON addresses.user_id = users.id").
		Select("users.name as user_name, addresses.name as address_name").
		OrderBy("users.name", "asc").
		Scan(&results)
	if err != nil {
		t.Errorf("OrderBy dotted column failed: %v", err)
		return
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
		return
	}
	if results[0].UserName != "dotted_alice" {
		t.Errorf("Expected first row 'dotted_alice', got '%s'", results[0].UserName)
	}
	if results[1].UserName != "dotted_bob" {
		t.Errorf("Expected second row 'dotted_bob', got '%s'", results[1].UserName)
	}
}

// TestDottedColumnOrderByDesc tests ORDER BY DESC with a table.column reference.
func TestDottedColumnOrderByDesc(t *testing.T, db *database.Database) {
	SeedDottedColumnData(t, db)

	type Result struct {
		UserName    string `db:"user_name"`
		AddressName string `db:"address_name"`
	}

	var results []Result
	err := db.Query().Table("users").
		Join("addresses ON addresses.user_id = users.id").
		Select("users.name as user_name, addresses.name as address_name").
		OrderByDesc("users.name").
		Scan(&results)
	if err != nil {
		t.Errorf("OrderByDesc dotted column failed: %v", err)
		return
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
		return
	}
	// Descending: bob before alice
	if results[0].UserName != "dotted_bob" {
		t.Errorf("Expected first row 'dotted_bob' (desc), got '%s'", results[0].UserName)
	}
}

// TestDottedColumnGroupBy tests GROUP BY with a table.column reference
// and COUNT(*) aggregate in a JOIN query.
func TestDottedColumnGroupBy(t *testing.T, db *database.Database) {
	SeedDottedColumnGroupData(t, db)

	type Result struct {
		UserName string `db:"user_name"`
		Count    int64  `db:"addr_count"`
	}

	var results []Result
	err := db.Query().Table("addresses").
		Join("users ON users.id = addresses.user_id").
		Select("users.name as user_name, COUNT(*) as addr_count").
		Group("users.name").
		OrderBy("users.name", "asc").
		Scan(&results)
	if err != nil {
		t.Errorf("GroupBy dotted column failed: %v", err)
		return
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(results))
		return
	}
	if results[0].UserName != "dotted_group_a" || results[0].Count != 1 {
		t.Errorf("Expected dotted_group_a with 1, got %s with %d", results[0].UserName, results[0].Count)
	}
	if results[1].UserName != "dotted_group_b" || results[1].Count != 3 {
		t.Errorf("Expected dotted_group_b with 3, got %s with %d", results[1].UserName, results[1].Count)
	}
}

// TestDottedColumnWhereColumn tests WhereColumn with table.column references
// on both sides of the comparison in a JOIN query.
func TestDottedColumnWhereColumn(t *testing.T, db *database.Database) {
	user1ID, _, _ := SeedDottedColumnData(t, db)

	type Result struct {
		UserName    string `db:"user_name"`
		AddressName string `db:"address_name"`
	}

	var results []Result
	err := db.Query().Table("users").
		Join("addresses ON addresses.user_id = users.id").
		WhereColumn("users.id", "=", "addresses.user_id").
		Where("users.id = ?", user1ID).
		Select("users.name as user_name, addresses.name as address_name").
		OrderBy("users.name", "asc").
		Scan(&results)
	if err != nil {
		t.Errorf("WhereColumn with dotted columns failed: %v", err)
		return
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
		return
	}
	if results[0].UserName != "dotted_alice" {
		t.Errorf("Expected 'dotted_alice', got '%s'", results[0].UserName)
	}
}

// TestDottedColumnDistinct tests Distinct with a table.column reference
// in a COUNT(DISTINCT col) aggregate.
func TestDottedColumnDistinct(t *testing.T, db *database.Database) {
	SeedDottedColumnData(t, db)

	// COUNT(DISTINCT users.name) — 2 users have addresses, 1 doesn't
	var count int64
	err := db.Query().Table("users").
		Join("addresses ON addresses.user_id = users.id").
		Distinct("users.name").
		Count(&count)
	if err != nil {
		t.Errorf("Distinct with dotted column failed: %v", err)
		return
	}

	if count != 2 {
		t.Errorf("Expected 2 distinct users with addresses, got %d", count)
	}
}

// TestDottedColumnGroupByHaving tests GROUP BY with a table.column reference
// combined with a HAVING clause to filter groups.
func TestDottedColumnGroupByHaving(t *testing.T, db *database.Database) {
	SeedDottedColumnGroupData(t, db)

	type Result struct {
		UserName string `db:"user_name"`
		Count    int64  `db:"addr_count"`
	}

	var results []Result
	err := db.Query().Table("addresses").
		Join("users ON users.id = addresses.user_id").
		Select("users.name as user_name, COUNT(*) as addr_count").
		Group("users.name").
		Having("COUNT(*) > ?", 2).
		OrderBy("users.name", "asc").
		Scan(&results)
	if err != nil {
		t.Errorf("GroupBy + Having with dotted column failed: %v", err)
		return
	}

	// user1 has 1 address (filtered out), user2 has 3 (passes HAVING > 2)
	if len(results) != 1 {
		t.Errorf("Expected 1 group passing HAVING > 2, got %d", len(results))
		return
	}
	if results[0].UserName != "dotted_group_b" {
		t.Errorf("Expected 'dotted_group_b', got '%s'", results[0].UserName)
	}
	if results[0].Count != 3 {
		t.Errorf("Expected count 3, got %d", results[0].Count)
	}
}
