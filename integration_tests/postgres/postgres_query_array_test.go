//go:build integration

package postgres

import (
	"testing"

	"github.com/dracory/neat/database"
	"github.com/dracory/neat/contracts/database/schema"
)

func TestPostgresIntegrationArray(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)

	tableName := "test_postgres_arrays"
	_ = db.Schema().DropIfExists(tableName)

	err := db.Schema().Create(tableName, func(table schema.Blueprint) {
		table.ID()
		table.Column("tags", "text[]")
		table.Column("numbers", "integer[]")
	})
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Test Insert with arrays
	err = db.Query().Table(tableName).Create(map[string]any{
		"tags":    []string{"tag1", "tag2"},
		"numbers": []int{1, 2, 3},
	})
	if err != nil {
		t.Errorf("Failed to insert first array: %v", err)
	}

	err = db.Query().Table(tableName).Create(map[string]any{
		"tags":    []string{"tag2", "tag3"},
		"numbers": []int{4, 5, 6},
	})
	if err != nil {
		t.Errorf("Failed to insert second array: %v", err)
	}

	// Test Query using PostgreSQL array operators
	// @> means "contains"
	var results []map[string]any
	err = db.Query().Table(tableName).Where("tags @> ?", "{tag1}").Get(&results)
	if err != nil {
		t.Errorf("Failed to query with @> operator: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	// Test ANY operator
	results = nil
	err = db.Query().Table(tableName).Where("? = ANY(tags)", "tag2").Get(&results)
	if err != nil {
		t.Errorf("Failed to query with ANY operator: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Test Array Overlap &&
	results = nil
	err = db.Query().Table(tableName).Where("numbers && ?", "{3,4}").Get(&results)
	if err != nil {
		t.Errorf("Failed to query with && operator: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	_ = db.Schema().Drop(tableName)
}
