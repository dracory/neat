package oracle_test

import (
	"database/sql"
	"errors"
	"testing"
)

func TestOracleMapScanNormalization(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get DB: %v", err)
	}
	_, err = sqlDB.Exec("INSERT INTO USERS (NAME, AVATAR, BIO, VOTES) VALUES ('John Doe', 'avatar_url', 'This is a bio description', 10)")
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}
	_, err = sqlDB.Exec("INSERT INTO USERS (NAME, AVATAR, BIO, VOTES) VALUES ('Jane Doe', 'avatar_url', NULL, 15)")
	if err != nil {
		t.Fatalf("failed to insert test user with NULL bio: %v", err)
	}

	// 1. Test []any via slice of interface{} path
	{
		var results []any
		err := db.Query().Table("USERS").Where("NAME", "John Doe").Get(&results)
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
		val, ok := m["BIO"]
		if !ok {
			t.Fatal("expected key 'BIO' in map")
		}
		if s, ok := val.(string); !ok || s != "This is a bio description" {
			t.Errorf("expected string 'This is a bio description', got %T (%v)", val, val)
		}

		// Test NULL column is scanned as nil
		var resultsNull []any
		err = db.Query().Table("USERS").Where("NAME", "Jane Doe").Get(&resultsNull)
		if err != nil {
			t.Fatalf("Get for NULL user failed: %v", err)
		}
		if len(resultsNull) != 1 {
			t.Fatalf("expected 1 result, got %d", len(resultsNull))
		}
		mNull, ok := resultsNull[0].(map[string]any)
		if !ok {
			t.Fatalf("expected map[string]any, got %T", resultsNull[0])
		}
		if mNull["BIO"] != nil {
			t.Errorf("expected bio to be nil for NULL column, got %T (%v)", mNull["BIO"], mNull["BIO"])
		}
	}

	// 2. Test slice of map destination ([]map[string]any)
	{
		var results []map[string]any
		err := db.Query().Table("USERS").Where("NAME", "John Doe").Get(&results)
		if err != nil {
			t.Fatalf("Get with []map failed: %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("expected 1 result, got %d", len(results))
		}
		val := results[0]["BIO"]
		if s, ok := val.(string); !ok || s != "This is a bio description" {
			t.Errorf("expected string 'This is a bio description' in []map, got %T (%v)", val, val)
		}

		// Test NULL column in []map[string]any
		var resultsNull []map[string]any
		err = db.Query().Table("USERS").Where("NAME", "Jane Doe").Get(&resultsNull)
		if err != nil {
			t.Fatalf("Get with []map for NULL user failed: %v", err)
		}
		if len(resultsNull) != 1 {
			t.Fatalf("expected 1 result, got %d", len(resultsNull))
		}
		if resultsNull[0]["BIO"] != nil {
			t.Errorf("expected bio to be nil in []map for NULL column, got %T (%v)", resultsNull[0]["BIO"], resultsNull[0]["BIO"])
		}
	}

	// 3. Test single map destination (*map[string]any)
	{
		var result map[string]any
		err := db.Query().Table("USERS").Where("NAME", "John Doe").First(&result)
		if err != nil {
			t.Fatalf("First with map failed: %v", err)
		}
		val := result["BIO"]
		if s, ok := val.(string); !ok || s != "This is a bio description" {
			t.Errorf("expected string 'This is a bio description' in single map, got %T (%v)", val, val)
		}

		// Test NULL column in single map
		var resultNull map[string]any
		err = db.Query().Table("USERS").Where("NAME", "Jane Doe").First(&resultNull)
		if err != nil {
			t.Fatalf("First with map for NULL user failed: %v", err)
		}
		if resultNull["BIO"] != nil {
			t.Errorf("expected bio to be nil in single map for NULL column, got %T (%v)", resultNull["BIO"], resultNull["BIO"])
		}

		// Test ErrNoRows returned when no rows match
		var noResult map[string]any
		err = db.Query().Table("USERS").Where("NAME", "non-existent").First(&noResult)
		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("expected First with map to return sql.ErrNoRows, got %v", err)
		}
	}
}
