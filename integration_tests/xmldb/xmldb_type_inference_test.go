package xmldb

import (
	"encoding/json"
	"testing"
	"time"
)

func TestXMLDBIntegrationTypeInferenceInteger(t *testing.T) {
	db := SetupXMLDBTest(t)

	var rows []xmldbTypedRow
	err := db.Query().Model(&xmldbTypedRow{}).Where("count = ?", 100).Get(&rows)
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

func TestXMLDBIntegrationTypeInferenceFloat(t *testing.T) {
	db := SetupXMLDBTest(t)

	var rows []xmldbTypedRow
	err := db.Query().Model(&xmldbTypedRow{}).Where("price = ?", 19.99).Get(&rows)
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

func TestXMLDBIntegrationTypeInferenceBool(t *testing.T) {
	db := SetupXMLDBTest(t)

	var rows []xmldbTypedRow
	err := db.Query().Model(&xmldbTypedRow{}).Where("active = ?", true).Get(&rows)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(rows))
	}
}

func TestXMLDBIntegrationTypeInferenceTime(t *testing.T) {
	db := SetupXMLDBTest(t)

	var rows []xmldbTypedRow
	err := db.Query().Model(&xmldbTypedRow{}).Where("created = ?", time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)).Get(&rows)
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

func TestXMLDBIntegrationTypeInferenceNestedObject(t *testing.T) {
	db := SetupXMLDBTest(t)

	var rows []xmldbTypedRow
	err := db.Query().Model(&xmldbTypedRow{}).Where("id = ?", 1).Get(&rows)
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

	zipVal := meta["zip"]
	if meta["city"] != "NYC" || (zipVal != "10001" && zipVal != float64(10001) && zipVal != int64(10001)) {
		t.Errorf("Unexpected meta content: %v", meta)
	}

	// Verify we can query using SQLite JSON functions!
	var rowsByCity []xmldbTypedRow
	err = db.Query().
		Model(&xmldbTypedRow{}).
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

func TestXMLDBIntegrationRepeatedElementsAsArray(t *testing.T) {
	// The typed.xml fixture has <tags><tag>red</tag><tag>blue</tag></tags>
	// for row 1. Previously, repeated child elements were silently collapsed
	// (last-wins), losing "red". Now they should be collected into a JSON
	// array: {"tag":["red","blue"]}.
	db := SetupXMLDBTest(t)

	var rows []xmldbTypedRow
	err := db.Query().Model(&xmldbTypedRow{}).Where("id = ?", 1).Get(&rows)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("Expected 1 row, got %d", len(rows))
	}

	tagsStr := rows[0].Tags
	if tagsStr == "" {
		t.Fatal("Expected tags not to be empty")
	}

	var tagsObj map[string]any
	if err := json.Unmarshal([]byte(tagsStr), &tagsObj); err != nil {
		t.Fatalf("Failed to parse tags JSON string: %v (value: %s)", err, tagsStr)
	}

	tagArray, ok := tagsObj["tag"].([]any)
	if !ok {
		t.Fatalf("Expected tags.tag to be a JSON array, got %T (%v)", tagsObj["tag"], tagsObj["tag"])
	}
	if len(tagArray) != 2 {
		t.Fatalf("Expected 2 tag values (red, blue), got %d: %v", len(tagArray), tagArray)
	}
	if tagArray[0] != "red" || tagArray[1] != "blue" {
		t.Errorf("Expected [red, blue], got %v", tagArray)
	}
}
