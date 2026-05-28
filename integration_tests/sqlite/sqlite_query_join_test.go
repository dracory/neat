package sqlite

import (
	"fmt"
	"testing"

	"github.com/dracory/neat"
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

func TestSQLiteIntegrationJoinInner(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
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

func TestSQLiteIntegrationJoinInnerWithConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
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

func TestSQLiteIntegrationJoinInnerWithAliases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	seedJoinTestData(t, db)

	results := make([]struct {
		UserName string `gorm:"column:name"`
	}, 0)
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
}

func TestSQLiteIntegrationJoinLeft(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	seedJoinTestData(t, db)

	results := make([]struct {
		UserName    string  `gorm:"column:name"`
		AddressName *string `gorm:"column:address_name"`
	}, 0)
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
}

func TestSQLiteIntegrationJoinLeftWithConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	seedJoinTestData(t, db)

	results := make([]struct {
		UserName    string  `gorm:"column:name"`
		AddressName *string `gorm:"column:address_name"`
	}, 0)
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
}

func TestSQLiteIntegrationJoinLeftWithAliases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	seedJoinTestData(t, db)

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
}

func TestSQLiteIntegrationJoinRight(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	if !isSQLiteVersionAtLeast(t, db, 3, 39, 0) {
		t.Skip("RIGHT JOIN requires SQLite 3.39.0 or higher")
	}
	seedJoinTestData(t, db)

	results := make([]struct {
		UserName    *string `gorm:"column:name"`
		AddressName string  `gorm:"column:address_name"`
	}, 0)
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
}

func TestSQLiteIntegrationJoinRightWithConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	if !isSQLiteVersionAtLeast(t, db, 3, 39, 0) {
		t.Skip("RIGHT JOIN requires SQLite 3.39.0 or higher")
	}
	seedJoinTestData(t, db)

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
}

func TestSQLiteIntegrationJoinRightWithAliases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	if !isSQLiteVersionAtLeast(t, db, 3, 39, 0) {
		t.Skip("RIGHT JOIN requires SQLite 3.39.0 or higher")
	}
	seedJoinTestData(t, db)

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
}

func TestSQLiteIntegrationJoinCross(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	seedJoinTestData(t, db)

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
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

func TestSQLiteIntegrationJoinCrossWithConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	seedJoinTestData(t, db)

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
}

func TestSQLiteIntegrationJoinCrossWithSelect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	seedJoinTestData(t, db)

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
}

func TestSQLiteIntegrationJoinMultiple(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	data := seedJoinTestData(t, db)

	book1 := models.Book{Name: "book1", UserID: data.user1.ID}
	if err := db.Query().Model(&models.Book{}).Create(&book1); err != nil {
		t.Fatalf("Failed to create book1: %v", err)
	}

	results := make([]struct {
		UserName    string `gorm:"column:user_name"`
		AddressName string `gorm:"column:address_name"`
		BookName    string `gorm:"column:book_name"`
	}, 0)
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
}

func TestSQLiteIntegrationJoinChaining(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	seedJoinTestData(t, db)

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
}

func TestSQLiteIntegrationJoinComplex(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	seedJoinTestData(t, db)

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
}

func isSQLiteVersionAtLeast(t *testing.T, db *neat.Database, major, minor, patch int) bool {
	sqlDB, err := db.DB()
	if err != nil {
		t.Logf("failed to get underlying sql.DB: %v", err)
		return false
	}

	rows, err := sqlDB.Query("SELECT sqlite_version()")
	if err != nil {
		t.Logf("failed to get sqlite version: %v", err)
		return false
	}
	defer rows.Close()

	var versionStr string
	if rows.Next() {
		err = rows.Scan(&versionStr)
		if err != nil {
			t.Logf("failed to scan sqlite version: %v", err)
			return false
		}
	} else {
		t.Log("no rows returned for sqlite_version()")
		return false
	}

	var curMajor, curMinor, curPatch int
	_, err = fmt.Sscanf(versionStr, "%d.%d.%d", &curMajor, &curMinor, &curPatch)
	if err != nil {
		t.Logf("failed to parse sqlite version '%s': %v", versionStr, err)
		return false
	}

	if curMajor > major {
		return true
	}
	if curMajor < major {
		return false
	}
	if curMinor > minor {
		return true
	}
	if curMinor < minor {
		return false
	}
	return curPatch >= patch
}
