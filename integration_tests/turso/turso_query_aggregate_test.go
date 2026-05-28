package turso

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

func TestTursoIntegrationQueryAggregateSum(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
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

func TestTursoIntegrationQueryAggregateSumWithWhere(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
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

func TestTursoIntegrationQueryAggregateAvg(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
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

func TestTursoIntegrationQueryAggregateMax(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
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

func TestTursoIntegrationQueryAggregateMin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
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

func TestTursoIntegrationQueryAggregateGroupBy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
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
