package sqlite

import (
	"testing"

	"github.com/dracory/neat/database"
	"github.com/dracory/neat/integration_tests/models"
)

func seedJsonTestData(t *testing.T, db *database.Database) []models.JsonData {
	query := db.Query()
	data := []models.JsonData{
		{Data: `{"name":"json1", "tags":["tag1", "tag2"], "meta":{"id":1, "active":true}}`},
		{Data: `{"name":"json2", "tags":["tag2", "tag3"], "meta":{"id":2, "active":false}}`},
		{Data: `{"name":"json3", "tags":["tag1", "tag3"], "meta":{"id":3, "active":true}}`},
	}
	if err := query.Model(&models.JsonData{}).Create(&data); err != nil {
		t.Fatalf("Failed to create JSON data: %v", err)
	}
	return data
}

func TestSQLiteIntegrationQueryJsonWhereJsonContains(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	data := seedJsonTestData(t, db)
	query := db.Query()

	var foundData []models.JsonData
	err := query.Model(&models.JsonData{}).WhereJsonContains("data->name", "json1").Find(&foundData)
	if err != nil {
		t.Errorf("WhereJsonContains failed: %v", err)
	}
	if len(foundData) != 1 {
		t.Errorf("Expected 1 result, got %d", len(foundData))
	}
	if len(foundData) > 0 && foundData[0].Data != data[0].Data {
		t.Errorf("Expected data %s, got %s", data[0].Data, foundData[0].Data)
	}
}

func TestSQLiteIntegrationQueryJsonOrWhereJsonContains(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	data := seedJsonTestData(t, db)
	query := db.Query()

	var foundData []models.JsonData
	err := query.Model(&models.JsonData{}).Where("id = ?", -1).OrWhereJsonContains("data->name", "json2").Find(&foundData)
	if err != nil {
		t.Errorf("OrWhereJsonContains failed: %v", err)
	}
	if len(foundData) != 1 {
		t.Errorf("Expected 1 result, got %d", len(foundData))
	}
	if len(foundData) > 0 && foundData[0].Data != data[1].Data {
		t.Errorf("Expected data %s, got %s", data[1].Data, foundData[0].Data)
	}
}

func TestSQLiteIntegrationQueryJsonWhereJsonDoesntContain(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	seedJsonTestData(t, db)
	query := db.Query()

	var foundData []models.JsonData
	err := query.Model(&models.JsonData{}).WhereJsonDoesntContain("data->name", "json1").Find(&foundData)
	if err != nil {
		t.Errorf("WhereJsonDoesntContain failed: %v", err)
	}
	if len(foundData) != 2 {
		t.Errorf("Expected 2 results, got %d", len(foundData))
	}
}

func TestSQLiteIntegrationQueryJsonWhereJsonContainsKey(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	seedJsonTestData(t, db)
	query := db.Query()

	var foundData []models.JsonData
	err := query.Model(&models.JsonData{}).WhereJsonContainsKey("data->meta->active").Find(&foundData)
	if err != nil {
		t.Errorf("WhereJsonContainsKey failed: %v", err)
	}
	if len(foundData) != 3 {
		t.Errorf("Expected 3 results, got %d", len(foundData))
	}
}

func TestSQLiteIntegrationQueryJsonWhereJsonDoesntContainKey(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	seedJsonTestData(t, db)
	query := db.Query()

	var foundData []models.JsonData
	err := query.Model(&models.JsonData{}).WhereJsonDoesntContainKey("data->meta->nonexistent").Find(&foundData)
	if err != nil {
		t.Errorf("WhereJsonDoesntContainKey failed: %v", err)
	}
	if len(foundData) != 3 {
		t.Errorf("Expected 3 results, got %d", len(foundData))
	}
}

func TestSQLiteIntegrationQueryJsonWhereJsonLength(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	seedJsonTestData(t, db)
	query := db.Query()

	var foundData []models.JsonData
	err := query.Model(&models.JsonData{}).WhereJsonLength("data->tags", "=", 2).Find(&foundData)
	if err != nil {
		t.Errorf("WhereJsonLength failed: %v", err)
	}
	if len(foundData) != 3 {
		t.Errorf("Expected 3 results, got %d", len(foundData))
	}
}

func TestSQLiteIntegrationQueryJsonWhereJsonLengthInvalidOperator(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	query := db.Query()

	var foundData []models.JsonData
	err := query.Model(&models.JsonData{}).WhereJsonLength("data->tags", "INVALID", 2).Find(&foundData)
	if err == nil {
		t.Error("Expected error for invalid operator in WhereJsonLength, got nil")
	}
}

func TestSQLiteIntegrationQueryJsonUpdateWithPath(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	seedJsonTestData(t, db)
	query := db.Query()

	// Test Update with JSON path - update all records since IDs are zero
	result, err := query.Model(&models.JsonData{}).Update("data->name", "updated_name")
	if err != nil {
		t.Errorf("Update with JSON path failed: %v", err)
	}
	if result != nil && result.RowsAffected != 3 {
		t.Errorf("Expected 3 rows affected, got %d", result.RowsAffected)
	}
}
