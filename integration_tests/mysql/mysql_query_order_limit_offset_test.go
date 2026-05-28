//go:build integration

package mysql

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func seedOrderLimitOffsetTestData(t *testing.T, db interface{}) {
	var query interface {
		Model(interface{}) interface{ Create(interface{}) error }
	}
	switch v := db.(type) {
	case interface {
		Query() interface {
			Model(interface{}) interface{ Create(interface{}) error }
		}
	}:
		query = v.Query()
	default:
		query = db
	}

	users := []models.User{
		{Name: "user_b", Avatar: "avatar2"},
		{Name: "user_a", Avatar: "avatar1"},
		{Name: "user_c", Avatar: "avatar3"},
		{Name: "user_d", Avatar: "avatar4"},
		{Name: "user_e", Avatar: "avatar5"},
	}
	if err := query.Model(&models.User{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}
}

func TestMySQLIntegrationOrderByAscending(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	seedOrderLimitOffsetTestData(t, db)

	var results []models.User
	err := db.Query().Model(&models.User{}).Where("name LIKE ?", "user_%").OrderBy("name", "asc").Find(&results)
	if err != nil {
		t.Errorf("OrderBy ascending failed: %v", err)
	}
	if len(results) != 5 {
		t.Errorf("Expected 5 results, got %d", len(results))
	}
	if results[0].Name != "user_a" {
		t.Errorf("Expected 'user_a', got '%s'", results[0].Name)
	}
	if results[1].Name != "user_b" {
		t.Errorf("Expected 'user_b', got '%s'", results[1].Name)
	}
	if results[2].Name != "user_c" {
		t.Errorf("Expected 'user_c', got '%s'", results[2].Name)
	}
}

func TestMySQLIntegrationOrderByDescending(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	seedOrderLimitOffsetTestData(t, db)

	var results []models.User
	err := db.Query().Model(&models.User{}).Where("name LIKE ?", "user_%").OrderBy("name", "desc").Find(&results)
	if err != nil {
		t.Errorf("OrderBy descending failed: %v", err)
	}
	if len(results) != 5 {
		t.Errorf("Expected 5 results, got %d", len(results))
	}
	if results[0].Name != "user_e" {
		t.Errorf("Expected 'user_e', got '%s'", results[0].Name)
	}
	if results[1].Name != "user_d" {
		t.Errorf("Expected 'user_d', got '%s'", results[1].Name)
	}
	if results[2].Name != "user_c" {
		t.Errorf("Expected 'user_c', got '%s'", results[2].Name)
	}
}

func TestMySQLIntegrationOrderByDescMethod(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	seedOrderLimitOffsetTestData(t, db)

	var results []models.User
	err := db.Query().Model(&models.User{}).Where("name LIKE ?", "user_%").OrderByDesc("name").Find(&results)
	if err != nil {
		t.Errorf("OrderByDesc failed: %v", err)
	}
	if len(results) != 5 {
		t.Errorf("Expected 5 results, got %d", len(results))
	}
	if results[0].Name != "user_e" {
		t.Errorf("Expected 'user_e', got '%s'", results[0].Name)
	}
}

func TestMySQLIntegrationMultipleOrderByClauses(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	query := db.Query()
	seedOrderLimitOffsetTestData(t, db)

	if err := query.Model(&models.User{}).Create(&models.User{Name: "user_a", Avatar: "avatar0"}); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	var results []models.User
	err := db.Query().Model(&models.User{}).Where("name LIKE ?", "user_%").OrderBy("name", "asc").OrderBy("avatar", "asc").Find(&results)
	if err != nil {
		t.Errorf("Multiple OrderBy failed: %v", err)
	}

	var userAEntries []models.User
	for _, u := range results {
		if u.Name == "user_a" {
			userAEntries = append(userAEntries, u)
		}
	}
	if len(userAEntries) != 2 {
		t.Errorf("Expected 2 user_a entries, got %d", len(userAEntries))
	}
	if userAEntries[0].Avatar != "avatar0" {
		t.Errorf("Expected 'avatar0', got '%s'", userAEntries[0].Avatar)
	}
	if userAEntries[1].Avatar != "avatar1" {
		t.Errorf("Expected 'avatar1', got '%s'", userAEntries[1].Avatar)
	}
}

func TestMySQLIntegrationLimitClause(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	seedOrderLimitOffsetTestData(t, db)

	var results []models.User
	err := db.Query().Model(&models.User{}).Where("name LIKE ?", "user_%").Limit(2).Find(&results)
	if err != nil {
		t.Errorf("Limit failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

func TestMySQLIntegrationLimitWithOrderBy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	seedOrderLimitOffsetTestData(t, db)

	var results []models.User
	err := db.Query().Model(&models.User{}).Where("name LIKE ?", "user_%").OrderBy("name", "asc").Limit(2).Find(&results)
	if err != nil {
		t.Errorf("Limit with OrderBy failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
	if results[0].Name != "user_a" {
		t.Errorf("Expected 'user_a', got '%s'", results[0].Name)
	}
	if results[1].Name != "user_a" {
		t.Errorf("Expected 'user_a', got '%s'", results[1].Name)
	}
}

func TestMySQLIntegrationOffsetWithLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	seedOrderLimitOffsetTestData(t, db)

	var results []models.User
	err := db.Query().Model(&models.User{}).Where("name LIKE ?", "user_%").OrderBy("name", "asc").Offset(2).Limit(2).Find(&results)
	if err != nil {
		t.Errorf("Offset with Limit failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
	if results[0].Name != "user_b" {
		t.Errorf("Expected 'user_b', got '%s'", results[0].Name)
	}
	if results[1].Name != "user_c" {
		t.Errorf("Expected 'user_c', got '%s'", results[1].Name)
	}
}
