package sqlite

import (
	"testing"

	"github.com/dracory/neat/database"
	"github.com/dracory/neat/integration_tests/models"
)

type aggregateTestData struct {
	expectedSum int64
	group1Sum   int64
	group2Sum   int64
	minID       int64
	maxID       int64
	expectedAvg float64
}

func seedAggregateTestData(t *testing.T, db *database.Database) aggregateTestData {
	users := []models.User{
		{Name: "aggregate_user_1", Avatar: "group1"},
		{Name: "aggregate_user_2", Avatar: "group1"},
		{Name: "aggregate_user_3", Avatar: "group2"},
		{Name: "aggregate_user_4", Avatar: "group2"},
	}
	for _, user := range users {
		if err := db.Query().Model(&models.User{}).Create(&user); err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	var insertedUsers []models.User
	if err := db.Query().Model(&models.User{}).Where("name LIKE ?", "aggregate_user_%").Find(&insertedUsers); err != nil {
		t.Fatalf("Failed to query inserted users: %v", err)
	}
	if len(insertedUsers) != 4 {
		t.Fatalf("Expected 4 inserted users, got %d", len(insertedUsers))
	}

	var expectedSum int64
	var group1Sum, group2Sum int64
	var minID, maxID int64 = int64(insertedUsers[0].ID), int64(insertedUsers[0].ID)
	for _, u := range insertedUsers {
		expectedSum += int64(u.ID)
		if u.Avatar == "group1" {
			group1Sum += int64(u.ID)
		} else {
			group2Sum += int64(u.ID)
		}
		if int64(u.ID) < minID {
			minID = int64(u.ID)
		}
		if int64(u.ID) > maxID {
			maxID = int64(u.ID)
		}
	}
	expectedAvg := float64(expectedSum) / float64(len(insertedUsers))

	return aggregateTestData{
		expectedSum: expectedSum,
		group1Sum:   group1Sum,
		group2Sum:   group2Sum,
		minID:       minID,
		maxID:       maxID,
		expectedAvg: expectedAvg,
	}
}

func TestSQLiteIntegrationQueryAggregateSum(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	data := seedAggregateTestData(t, db)

	var sum int64
	err := db.Query().Table("users").Where("name LIKE ?", "aggregate_user_%").Sum("id", &sum)
	if err != nil {
		t.Errorf("Sum failed: %v", err)
	}
	if sum != data.expectedSum {
		t.Errorf("Expected sum %d, got %d", data.expectedSum, sum)
	}
}

func TestSQLiteIntegrationQueryAggregateSumWithWhere(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	data := seedAggregateTestData(t, db)

	var sum int64
	err := db.Query().Table("users").Where("avatar = ?", "group1").Sum("id", &sum)
	if err != nil {
		t.Errorf("Sum with where failed: %v", err)
	}
	if sum != data.group1Sum {
		t.Errorf("Expected sum %d for group1, got %d", data.group1Sum, sum)
	}
}

func TestSQLiteIntegrationQueryAggregateAvg(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	data := seedAggregateTestData(t, db)

	var avg float64
	err := db.Query().Table("users").Where("name LIKE ?", "aggregate_user_%").Avg("id", &avg)
	if err != nil {
		t.Errorf("Avg failed: %v", err)
	}
	if avg != data.expectedAvg {
		t.Errorf("Expected avg %f, got %f", data.expectedAvg, avg)
	}
}

func TestSQLiteIntegrationQueryAggregateMax(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	data := seedAggregateTestData(t, db)

	var max int64
	err := db.Query().Table("users").Where("name LIKE ?", "aggregate_user_%").Max("id", &max)
	if err != nil {
		t.Errorf("Max failed: %v", err)
	}
	if max != data.maxID {
		t.Errorf("Expected max %d, got %d", data.maxID, max)
	}
}

func TestSQLiteIntegrationQueryAggregateMin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	data := seedAggregateTestData(t, db)

	var min int64
	err := db.Query().Table("users").Where("name LIKE ?", "aggregate_user_%").Min("id", &min)
	if err != nil {
		t.Errorf("Min failed: %v", err)
	}
	if min != data.minID {
		t.Errorf("Expected min %d, got %d", data.minID, min)
	}
}

func TestSQLiteIntegrationQueryAggregateGroupBy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	data := seedAggregateTestData(t, db)

	type Result struct {
		Avatar string
		Total  int64
	}
	var results []Result
	err := db.Query().Table("users").Where("name LIKE ?", "aggregate_user_%").
		Select("avatar, SUM(id) as total").Group("avatar").Scan(&results)
	if err != nil {
		t.Errorf("Aggregate with GroupBy failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
	for _, res := range results {
		if res.Avatar == "group1" {
			if res.Total != data.group1Sum {
				t.Errorf("Expected total %d for group1, got %d", data.group1Sum, res.Total)
			}
		} else if res.Avatar == "group2" {
			if res.Total != data.group2Sum {
				t.Errorf("Expected total %d for group2, got %d", data.group2Sum, res.Total)
			}
		}
	}
}

func TestSQLiteIntegrationQueryAggregateInvalidColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)

	var sum int64
	err := db.Query().Table("users").Sum("invalid; column", &sum)
	if err == nil {
		t.Error("Expected error for invalid column")
	}
}

func TestSQLiteIntegrationQueryAggregateSQLInjectionComment(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)

	var sum int64
	err := db.Query().Table("users").Sum("id; --", &sum)
	if err == nil {
		t.Error("Expected error for SQL injection attempt")
	}
}

func TestSQLiteIntegrationQueryAggregateSQLInjectionUnion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)

	var sum int64
	err := db.Query().Table("users").Sum("id) UNION SELECT 1--", &sum)
	if err == nil {
		t.Error("Expected error for SQL injection attempt")
	}
}

func TestSQLiteIntegrationQueryAggregateNilPointer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)

	err := db.Query().Table("users").Sum("id", nil)
	if err == nil {
		t.Error("Expected error for nil destination")
	}
}

func TestSQLiteIntegrationQueryAggregateEmptyResultSet(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)

	var sum *int64
	err := db.Query().Table("users").Where("name = ?", "non_existent").Sum("id", &sum)
	if err != nil {
		t.Errorf("Sum with empty result failed: %v", err)
	}
	if sum != nil {
		t.Error("Expected nil for empty result set")
	}

	var avg *float64
	err = db.Query().Table("users").Where("name = ?", "non_existent").Avg("id", &avg)
	if err != nil {
		t.Errorf("Avg with empty result failed: %v", err)
	}
	if avg != nil {
		t.Error("Expected nil for empty result set")
	}
}

func TestSQLiteIntegrationQueryAggregateNullValues(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)

	u1 := models.User{Name: "null_test_1", Bio: nil}
	bio2 := "some bio"
	u2 := models.User{Name: "null_test_2", Bio: &bio2}

	if err := db.Query().Model(&models.User{}).Create(&u1); err != nil {
		t.Fatalf("Failed to create u1: %v", err)
	}
	if err := db.Query().Model(&models.User{}).Create(&u2); err != nil {
		t.Fatalf("Failed to create u2: %v", err)
	}

	var count int64
	err := db.Query().Table("users").Where("name LIKE ?", "null_test_%").WhereNotNull("bio").Count(&count)
	if err != nil {
		t.Errorf("Count with WhereNotNull failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}

	var nullUsers []models.User
	if err := db.Query().Model(&models.User{}).Where("name LIKE ?", "null_test_%").Find(&nullUsers); err != nil {
		t.Fatalf("Failed to query null test users: %v", err)
	}
	var expectedNullSum int64
	for _, u := range nullUsers {
		expectedNullSum += int64(u.ID)
	}

	var sum int64
	err = db.Query().Table("users").Where("name LIKE ?", "null_test_%").Sum("id", &sum)
	if err != nil {
		t.Errorf("Sum with NULL values failed: %v", err)
	}
	if sum != expectedNullSum {
		t.Errorf("Expected sum %d, got %d", expectedNullSum, sum)
	}
}

func TestSQLiteIntegrationQueryAggregateNonNumericColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	seedAggregateTestData(t, db)

	var sum float64
	err := db.Query().Table("users").Where("name LIKE ?", "aggregate_user_%").Sum("name", &sum)
	if err != nil {
		t.Errorf("Sum on string column failed: %v", err)
	}
	if sum != 0.0 {
		t.Errorf("Expected sum 0.0 for string column, got %f", sum)
	}

	var max string
	err = db.Query().Table("users").Where("name LIKE ?", "aggregate_user_%").Max("name", &max)
	if err != nil {
		t.Errorf("Max on string column failed: %v", err)
	}
	if max != "aggregate_user_4" {
		t.Errorf("Expected 'aggregate_user_4', got '%s'", max)
	}
}
