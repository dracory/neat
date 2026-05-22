package sqlite

import (
	"strings"
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestSQLiteIntegrationQueryJson(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	query := db.Query()

	// Create JSON data
	data := []models.JsonData{
		{Data: `{"name":"json1", "tags":["tag1", "tag2"], "meta":{"id":1, "active":true}}`},
		{Data: `{"name":"json2", "tags":["tag2", "tag3"], "meta":{"id":2, "active":false}}`},
		{Data: `{"name":"json3", "tags":["tag1", "tag3"], "meta":{"id":3, "active":true}}`},
	}
	if err := query.Model(&models.JsonData{}).Create(&data); err != nil {
		t.Fatalf("Failed to create JSON data: %v", err)
	}

	// Test WhereJsonContains (SQLite uses = for simple field comparison in my implementation)
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

	// Test OrWhereJsonContains
	foundData = nil
	err = query.Model(&models.JsonData{}).Where("id = ?", -1).OrWhereJsonContains("data->name", "json2").Find(&foundData)
	if err != nil {
		t.Errorf("OrWhereJsonContains failed: %v", err)
	}
	if len(foundData) != 1 {
		t.Errorf("Expected 1 result, got %d", len(foundData))
	}
	if len(foundData) > 0 && foundData[0].ID != data[1].ID {
		t.Errorf("Expected ID %d, got %d", data[1].ID, foundData[0].ID)
	}

	// Test WhereJsonDoesntContain
	foundData = nil
	err = query.Model(&models.JsonData{}).WhereJsonDoesntContain("data->name", "json1").Find(&foundData)
	if err != nil {
		t.Errorf("WhereJsonDoesntContain failed: %v", err)
	}
	if len(foundData) != 2 {
		t.Errorf("Expected 2 results, got %d", len(foundData))
	}

	// Test WhereJsonContainsKey
	foundData = nil
	err = query.Model(&models.JsonData{}).WhereJsonContainsKey("data->meta->active").Find(&foundData)
	if err != nil {
		t.Errorf("WhereJsonContainsKey failed: %v", err)
	}
	if len(foundData) != 3 {
		t.Errorf("Expected 3 results, got %d", len(foundData))
	}

	// Test WhereJsonDoesntContainKey
	foundData = nil
	err = query.Model(&models.JsonData{}).WhereJsonDoesntContainKey("data->meta->nonexistent").Find(&foundData)
	if err != nil {
		t.Errorf("WhereJsonDoesntContainKey failed: %v", err)
	}
	if len(foundData) != 3 {
		t.Errorf("Expected 3 results, got %d", len(foundData))
	}

	// Test WhereJsonLength
	foundData = nil
	err = query.Model(&models.JsonData{}).WhereJsonLength("data->tags", "=", 2).Find(&foundData)
	if err != nil {
		t.Errorf("WhereJsonLength failed: %v", err)
	}
	if len(foundData) != 3 {
		t.Errorf("Expected 3 results, got %d", len(foundData))
	}

	err = query.Model(&models.JsonData{}).WhereJsonLength("data->tags", ">", 2).Find(&foundData)
	if err != nil {
		t.Errorf("WhereJsonLength greater than failed: %v", err)
	}
	if len(foundData) != 0 {
		t.Errorf("Expected 0 results, got %d", len(foundData))
	}

	// Test array indexing
	foundData = nil
	err = query.Model(&models.JsonData{}).WhereJsonContains("data->tags->0", "tag1").Find(&foundData)
	if err != nil {
		t.Errorf("WhereJsonContains with array indexing failed: %v", err)
	}
	if len(foundData) != 2 { // json1 and json3 have tag1 at index 0
		t.Errorf("Expected 2 results, got %d", len(foundData))
	}

	// Test Update with JSON path
	result, err := query.Model(&models.JsonData{}).Where("id = ?", data[0].ID).Update("data->name", "updated_name")
	if err != nil {
		t.Errorf("Update with JSON path failed: %v", err)
	}
	if result.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", result.RowsAffected)
	}

	var updatedData models.JsonData
	err = query.Model(&models.JsonData{}).Find(&updatedData, data[0].ID)
	if err != nil {
		t.Errorf("Failed to find updated data: %v", err)
	}
	if !strings.Contains(updatedData.Data, "updated_name") {
		t.Error("Expected updated_name in data")
	}
}
