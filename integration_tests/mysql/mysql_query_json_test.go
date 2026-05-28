//go:build integration

package mysql

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func seedJsonTestData(t *testing.T, db interface{}) []models.JsonData {
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

	data := []models.JsonData{
		{Data: `{"name":"json1", "tags":["tag1", "tag2"], "meta":{"id":1, "active":true}}`},
		{Data: `{"name":"json2", "tags":["tag2", "tag3"], "meta":{"id":2, "active":false}}`},
		{Data: `{"name":"json3", "tags":["tag1", "tag3"], "meta":{"id":3, "active":true}}`},
	}
	err := query.Model(&models.JsonData{}).Create(&data)
	if err != nil {
		t.Fatalf("Failed to create JSON data: %v", err)
	}
	return data
}

func TestMySQLIntegrationQueryJsonWhereJsonContains(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	query := db.Query()
	data := seedJsonTestData(t, db)

	var foundData []models.JsonData
	err := query.Model(&models.JsonData{}).WhereJsonContains("data->name", "json1").Find(&foundData)
	if err != nil {
		t.Errorf("WhereJsonContains failed: %v", err)
	}
	if len(foundData) != 1 {
		t.Errorf("Expected 1 result, got %d", len(foundData))
	}
	if len(foundData) > 0 && foundData[0].ID != data[0].ID {
		t.Errorf("Expected ID %d, got %d", data[0].ID, foundData[0].ID)
	}
}

func TestMySQLIntegrationQueryJsonWhereJsonContainsArray(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	query := db.Query()
	seedJsonTestData(t, db)

	var foundData []models.JsonData
	err := query.Model(&models.JsonData{}).WhereJsonContains("data->tags", "tag1").Find(&foundData)
	if err != nil {
		t.Errorf("WhereJsonContains for array failed: %v", err)
	}
	if len(foundData) != 2 {
		t.Errorf("Expected 2 results, got %d", len(foundData))
	}
}

func TestMySQLIntegrationQueryJsonOrWhereJsonContains(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	query := db.Query()
	data := seedJsonTestData(t, db)

	var foundData []models.JsonData
	err := query.Model(&models.JsonData{}).Where("id = ?", -1).OrWhereJsonContains("data->name", "json2").Find(&foundData)
	if err != nil {
		t.Errorf("OrWhereJsonContains failed: %v", err)
	}
	if len(foundData) != 1 {
		t.Errorf("Expected 1 result, got %d", len(foundData))
	}
	if len(foundData) > 0 && foundData[0].ID != data[1].ID {
		t.Errorf("Expected ID %d, got %d", data[1].ID, foundData[0].ID)
	}
}

func TestMySQLIntegrationQueryJsonWhereJsonDoesntContain(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	query := db.Query()
	seedJsonTestData(t, db)

	var foundData []models.JsonData
	err := query.Model(&models.JsonData{}).WhereJsonDoesntContain("data->name", "json1").Find(&foundData)
	if err != nil {
		t.Errorf("WhereJsonDoesntContain failed: %v", err)
	}
	if len(foundData) != 2 {
		t.Errorf("Expected 2 results, got %d", len(foundData))
	}
}

func TestMySQLIntegrationQueryJsonWhereJsonContainsKey(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	query := db.Query()
	seedJsonTestData(t, db)

	var foundData []models.JsonData
	err := query.Model(&models.JsonData{}).WhereJsonContainsKey("data->meta->active").Find(&foundData)
	if err != nil {
		t.Errorf("WhereJsonContainsKey failed: %v", err)
	}
	if len(foundData) != 3 {
		t.Errorf("Expected 3 results, got %d", len(foundData))
	}
}

func TestMySQLIntegrationQueryJsonWhereJsonDoesntContainKey(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	query := db.Query()
	seedJsonTestData(t, db)

	var foundData []models.JsonData
	err := query.Model(&models.JsonData{}).WhereJsonDoesntContainKey("data->meta->nonexistent").Find(&foundData)
	if err != nil {
		t.Errorf("WhereJsonDoesntContainKey failed: %v", err)
	}
	if len(foundData) != 3 {
		t.Errorf("Expected 3 results, got %d", len(foundData))
	}
}

func TestMySQLIntegrationQueryJsonWhereJsonLength(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	query := db.Query()
	seedJsonTestData(t, db)

	var foundData []models.JsonData
	err := query.Model(&models.JsonData{}).WhereJsonLength("data->tags", "=", 2).Find(&foundData)
	if err != nil {
		t.Errorf("WhereJsonLength failed: %v", err)
	}
	if len(foundData) != 3 {
		t.Errorf("Expected 3 results, got %d", len(foundData))
	}
}

func TestMySQLIntegrationQueryJsonArrayIndexing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	query := db.Query()
	seedJsonTestData(t, db)

	var foundData []models.JsonData
	err := query.Model(&models.JsonData{}).WhereJsonContains("data->tags->0", "tag1").Find(&foundData)
	if err != nil {
		t.Errorf("WhereJsonContains with array indexing failed: %v", err)
	}
	if len(foundData) != 2 {
		t.Errorf("Expected 2 results, got %d", len(foundData))
	}
}

func TestMySQLIntegrationQueryJsonUpdateWithPath(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	query := db.Query()
	data := seedJsonTestData(t, db)

	result, err := query.Model(&models.JsonData{}).Where("id = ?", data[0].ID).Update("data->name", "updated_name")
	if err != nil {
		t.Errorf("Update with JSON path failed: %v", err)
	}
	if result.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", result.RowsAffected)
	}

	var updatedData models.JsonData
	err = query.Model(&models.JsonData{}).Where("id = ?", data[0].ID).First(&updatedData)
	if err != nil {
		t.Errorf("Failed to find updated data: %v", err)
	}
	if updatedData.Data == "" {
		t.Error("Data should not be empty")
	}
}
