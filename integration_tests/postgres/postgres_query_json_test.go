package postgres_test

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
	err := query.Model(&models.JsonData{}).Create(&data)
	if err != nil {
		t.Fatalf("Failed to create JSON data: %v", err)
	}
	return data
}

func TestPostgresIntegrationQueryJsonWhereJsonContains(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()
	seedJsonTestData(t, db)

	var foundData []models.JsonData
	err := query.Model(&models.JsonData{}).WhereJsonContains("data->name", "json1").Find(&foundData)
	if err != nil {
		t.Errorf("WhereJsonContains failed: %v", err)
	}
	if len(foundData) != 1 {
		t.Errorf("Expected 1 result, got %d", len(foundData))
	}
	if len(foundData) > 0 && foundData[0].Data == "" {
		t.Errorf("Expected non-empty data")
	}
}

func TestPostgresIntegrationQueryJsonWhereJsonContainsArray(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
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

func TestPostgresIntegrationQueryJsonOrWhereJsonContains(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()
	seedJsonTestData(t, db)

	var foundData []models.JsonData
	err := query.Model(&models.JsonData{}).Where("id = ?", -1).OrWhereJsonContains("data->name", "json2").Find(&foundData)
	if err != nil {
		t.Errorf("OrWhereJsonContains failed: %v", err)
	}
	if len(foundData) != 1 {
		t.Errorf("Expected 1 result, got %d", len(foundData))
	}
	if len(foundData) > 0 && foundData[0].Data == "" {
		t.Errorf("Expected non-empty data")
	}
}

func TestPostgresIntegrationQueryJsonWhereJsonDoesntContain(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
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

func TestPostgresIntegrationQueryJsonWhereJsonContainsKey(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
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

func TestPostgresIntegrationQueryJsonWhereJsonDoesntContainKey(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
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

func TestPostgresIntegrationQueryJsonWhereJsonLength(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
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

func TestPostgresIntegrationQueryJsonWhereJsonLengthInvalidOperator(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	var foundData []models.JsonData
	err := query.Model(&models.JsonData{}).WhereJsonLength("data->tags", "INVALID", 2).Find(&foundData)
	if err == nil {
		t.Error("Expected error for invalid operator in WhereJsonLength, got nil")
	}
}

func TestPostgresIntegrationQueryJsonInvalidPathSegment(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	var foundData []models.JsonData
	// Test with an invalid path segment containing special characters
	err := query.Model(&models.JsonData{}).WhereJsonContains("data->invalid'path", "value").Find(&foundData)
	if err == nil {
		t.Error("Expected error for invalid JSON path segment, got nil")
	}
}

func TestPostgresIntegrationQueryJsonArrayIndexing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
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
