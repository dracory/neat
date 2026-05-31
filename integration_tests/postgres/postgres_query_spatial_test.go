
package postgres

import (
	"testing"

	neatcontracts "github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/database/query"
)

func TestPostgresIntegrationSpatial(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)

	// Check if PostGIS is installed
	var postgisVersion string
	err := db.Query().Select(query.RawExpr("PostGIS_Version()")).Scan(&postgisVersion)
	if err != nil {
		t.Skip("PostGIS not installed, skipping spatial tests")
	}

	// Create table with spatial column
	err = db.Schema().Create("spatial_models", func(blueprint neatcontracts.Blueprint) {
		blueprint.ID()
		blueprint.String("name")
		blueprint.Point("location")
	})
	if err != nil {
		t.Fatalf("Failed to create spatial table: %v", err)
	}
	defer db.Schema().Drop("spatial_models")

	// Insert data using ST_GeomFromText
	queryBuilder := db.Query()
	err = queryBuilder.Table("spatial_models").Create(map[string]any{
		"name":     "Point 1",
		"location": query.RawExpr("ST_GeomFromText(?)", "POINT(1 1)"),
	})
	if err != nil {
		t.Fatalf("Failed to insert spatial data: %v", err)
	}

	// Query data using ST_AsText
	var results []map[string]any
	err = db.Query().Table("spatial_models").
		Select(query.RawExpr("name, ST_AsText(location) as location_text")).
		Find(&results)

	if err != nil {
		t.Fatalf("Failed to query spatial data: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	} else {
		name := results[0]["name"].(string)
		locationText := results[0]["location_text"].(string)

		if name != "Point 1" {
			t.Errorf("Expected name 'Point 1', got %s", name)
		}
		if locationText != "POINT(1 1)" {
			t.Errorf("Expected location 'POINT(1 1)', got %s", locationText)
		}
	}
}
