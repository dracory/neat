package jsondb

import (
	"encoding/json"
	"testing"
	"time"
)

func TestJSONDBIntegrationTypeInferenceInteger(t *testing.T) {
	db := SetupJSONDBTest(t)

	var rows []jsondbTypedRow
	err := db.Query().Model(&jsondbTypedRow{}).Where("count = ?", 100).Get(&rows)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("Expected 1 row, got %d", len(rows))
	}
	if rows[0].Count != 100 {
		t.Errorf("Expected count 100, got %d", rows[0].Count)
	}
}

func TestJSONDBIntegrationTypeInferenceFloat(t *testing.T) {
	db := SetupJSONDBTest(t)

	var rows []jsondbTypedRow
	err := db.Query().Model(&jsondbTypedRow{}).Where("price = ?", 19.99).Get(&rows)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("Expected 1 row, got %d", len(rows))
	}
	if rows[0].Price != 19.99 {
		t.Errorf("Expected price 19.99, got %v", rows[0].Price)
	}
}

func TestJSONDBIntegrationTypeInferenceBool(t *testing.T) {
	db := SetupJSONDBTest(t)

	var rows []jsondbTypedRow
	err := db.Query().Model(&jsondbTypedRow{}).Where("active = ?", true).Get(&rows)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(rows))
	}
}

func TestJSONDBIntegrationTypeInferenceTime(t *testing.T) {
	db := SetupJSONDBTest(t)

	var rows []jsondbTypedRow
	err := db.Query().Model(&jsondbTypedRow{}).Where("created = ?", time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)).Get(&rows)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("Expected 1 row, got %d", len(rows))
	}
	if !rows[0].Created.Equal(time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)) {
		t.Errorf("Expected time 2024-01-15T10:30:00Z, got %v", rows[0].Created)
	}
}

func TestJSONDBIntegrationTypeInferenceNestedObject(t *testing.T) {
	db := SetupJSONDBTest(t)

	var rows []jsondbTypedRow
	err := db.Query().Model(&jsondbTypedRow{}).Where("id = ?", 1).Get(&rows)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("Expected 1 row, got %d", len(rows))
	}

	metaStr := rows[0].Meta
	if metaStr == "" {
		t.Fatal("Expected meta not to be empty")
	}

	var meta map[string]any
	if err := json.Unmarshal([]byte(metaStr), &meta); err != nil {
		t.Fatalf("Failed to parse meta JSON string: %v", err)
	}

	if meta["city"] != "NYC" || meta["zip"] != "10001" {
		t.Errorf("Unexpected meta content: %v", meta)
	}

	// Verify we can query using SQLite JSON functions!
	var rowsByCity []jsondbTypedRow
	err = db.Query().
		Model(&jsondbTypedRow{}).
		Where("json_extract(meta, '$.city') = ?", "SF").
		Get(&rowsByCity)

	if err != nil {
		t.Fatalf("JSON extract query failed: %v", err)
	}

	if len(rowsByCity) != 1 {
		t.Fatalf("Expected 1 row with city SF, got %d", len(rowsByCity))
	}
	if rowsByCity[0].ID != 2 {
		t.Errorf("Expected row ID 2, got %d", rowsByCity[0].ID)
	}
}

func TestJSONDBIntegrationTypeInferenceNestedArray(t *testing.T) {
	db := SetupJSONDBTest(t)

	var rows []jsondbTypedRow
	err := db.Query().Model(&jsondbTypedRow{}).Where("id = ?", 1).Get(&rows)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("Expected 1 row, got %d", len(rows))
	}

	tagsStr := rows[0].Tags
	var tags []string
	if err := json.Unmarshal([]byte(tagsStr), &tags); err != nil {
		t.Fatalf("Failed to parse tags JSON string: %v", err)
	}

	if len(tags) != 2 || tags[0] != "red" || tags[1] != "blue" {
		t.Errorf("Unexpected tags: %v", tags)
	}
}
