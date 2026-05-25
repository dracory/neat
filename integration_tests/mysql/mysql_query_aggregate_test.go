//go:build integration

package mysql

import (
	"testing"
	"time"

	"github.com/dracory/neat/database"
	"github.com/dracory/neat/integration_tests/models"
)

func seedAggregateTestData(t *testing.T, db *database.Database) []models.User {
	query := db.Query()
	now := time.Now()

	users := []models.User{
		{Name: "aggregate_user_1", Avatar: "group1", CreatedAt: now, UpdatedAt: now},
		{Name: "aggregate_user_2", Avatar: "group1", CreatedAt: now, UpdatedAt: now},
		{Name: "aggregate_user_3", Avatar: "group2", CreatedAt: now, UpdatedAt: now},
		{Name: "aggregate_user_4", Avatar: "group2", CreatedAt: now, UpdatedAt: now},
	}

	for _, user := range users {
		if err := query.Model(&models.User{}).Create(&user); err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	var createdUsers []models.User
	if err := query.Model(&models.User{}).Where("name LIKE ?", "aggregate_user_%").Find(&createdUsers); err != nil {
		t.Fatalf("Failed to get created users: %v", err)
	}

	return createdUsers
}

func TestMySQLIntegrationQueryAggregateSum(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	query := db.Query()
	createdUsers := seedAggregateTestData(t, db)

	var sum int64
	err := query.Table("users").Where("name LIKE ?", "aggregate_user_%").Sum("id", &sum)
	if err != nil {
		t.Errorf("Sum failed: %v", err)
	}

	expectedSum := int64(0)
	for _, user := range createdUsers {
		expectedSum += int64(user.ID)
	}
	if sum != expectedSum {
		t.Errorf("Expected sum %d, got %d", expectedSum, sum)
	}
}

func TestMySQLIntegrationQueryAggregateSumWithWhere(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	query := db.Query()
	createdUsers := seedAggregateTestData(t, db)

	var sum int64
	err := query.Table("users").Where("avatar = ?", "group1").Sum("id", &sum)
	if err != nil {
		t.Errorf("Sum with where failed: %v", err)
	}

	expectedSum := int64(0)
	for _, user := range createdUsers {
		if user.Avatar == "group1" {
			expectedSum += int64(user.ID)
		}
	}
	if sum != expectedSum {
		t.Errorf("Expected sum %d, got %d", expectedSum, sum)
	}
}

func TestMySQLIntegrationQueryAggregateAvg(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	query := db.Query()
	seedAggregateTestData(t, db)

	var avg float64
	err := query.Table("users").Where("name LIKE ?", "aggregate_user_%").Avg("id", &avg)
	if err != nil {
		t.Errorf("Avg failed: %v", err)
	}
	if avg == 0 {
		t.Error("Avg should not be zero")
	}
}

func TestMySQLIntegrationQueryAggregateMax(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	query := db.Query()
	seedAggregateTestData(t, db)

	var max int64
	err := query.Table("users").Where("name LIKE ?", "aggregate_user_%").Max("id", &max)
	if err != nil {
		t.Errorf("Max failed: %v", err)
	}
	if max == 0 {
		t.Error("Max should not be zero")
	}
}

func TestMySQLIntegrationQueryAggregateMin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	query := db.Query()
	seedAggregateTestData(t, db)

	var min int64
	err := query.Table("users").Where("name LIKE ?", "aggregate_user_%").Min("id", &min)
	if err != nil {
		t.Errorf("Min failed: %v", err)
	}
	if min == 0 {
		t.Error("Min should not be zero")
	}
}

func TestMySQLIntegrationQueryAggregateGroupBy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	query := db.Query()
	seedAggregateTestData(t, db)

	type Result struct {
		Avatar string
		Total  int64
	}
	var results []Result
	err := query.Table("users").Where("name LIKE ?", "aggregate_user_%").
		Select("avatar, SUM(id) as total").Group("avatar").Scan(&results)
	if err != nil {
		t.Errorf("GroupBy failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

func TestMySQLIntegrationQueryAggregateInvalidColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	query := db.Query()

	var sum int64
	err := query.Table("users").Sum("invalid; column", &sum)
	if err == nil {
		t.Error("Expected error for invalid column")
	}
}

func TestMySQLIntegrationQueryAggregateNilPointer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	query := db.Query()

	err := query.Table("users").Sum("id", nil)
	if err == nil {
		t.Error("Expected error for nil pointer")
	}
}

func TestMySQLIntegrationQueryAggregateEmptyResult(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	query := db.Query()

	var sum *int64
	err := query.Table("users").Where("name = ?", "non_existent").Sum("id", &sum)
	if err != nil {
		t.Errorf("Sum failed: %v", err)
	}
	if sum != nil {
		t.Error("Expected nil for empty result set")
	}

	var avg *float64
	err = query.Table("users").Where("name = ?", "non_existent").Avg("id", &avg)
	if err != nil {
		t.Errorf("Avg failed: %v", err)
	}
	if avg != nil {
		t.Error("Expected nil for empty result set")
	}
}

func TestMySQLIntegrationQueryAggregateNonNumericColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	query := db.Query()
	seedAggregateTestData(t, db)

	var max string
	err := query.Table("users").Where("name LIKE ?", "aggregate_user_%").Max("name", &max)
	if err != nil {
		t.Errorf("Max on string failed: %v", err)
	}
	if max == "" {
		t.Error("Max should not be empty")
	}
}
