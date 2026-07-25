package mysql_test

import (
	"testing"
)

func TestMySQLMapScanNormalization(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)

	// Insert a test user with some string and text fields
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get DB: %v", err)
	}
	_, err = sqlDB.Exec("INSERT INTO users (name, avatar, bio, votes) VALUES ('John Doe', 'avatar_url', 'This is a bio description', 10)")
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	// 1. Test []any via slice of interface{} path
	{
		var results []any
		err := db.Query().Table("users").Where("name", "John Doe").Get(&results)
		if err != nil {
			t.Fatalf("Get with []any failed: %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("expected 1 result, got %d", len(results))
		}
		m, ok := results[0].(map[string]any)
		if !ok {
			t.Fatalf("expected map[string]any, got %T", results[0])
		}
		val, ok := m["bio"]
		if !ok {
			t.Fatal("expected key 'bio' in map")
		}
		if s, ok := val.(string); !ok || s != "This is a bio description" {
			t.Errorf("expected string 'This is a bio description', got %T (%v)", val, val)
		}
	}

	// 2. Test slice of map destination ([]map[string]any)
	{
		var results []map[string]any
		err := db.Query().Table("users").Where("name", "John Doe").Get(&results)
		if err != nil {
			t.Fatalf("Get with []map failed: %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("expected 1 result, got %d", len(results))
		}
		val := results[0]["bio"]
		if s, ok := val.(string); !ok || s != "This is a bio description" {
			t.Errorf("expected string 'This is a bio description' in []map, got %T (%v)", val, val)
		}
	}

	// 3. Test single map destination (*map[string]any)
	{
		var result map[string]any
		err := db.Query().Table("users").Where("name", "John Doe").First(&result)
		if err != nil {
			t.Fatalf("First with map failed: %v", err)
		}
		val := result["bio"]
		if s, ok := val.(string); !ok || s != "This is a bio description" {
			t.Errorf("expected string 'This is a bio description' in single map, got %T (%v)", val, val)
		}
	}
}
