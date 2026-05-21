//go:build integration

package mysql

import (
	"testing"

	"github.com/dracory/neat/database"
	"github.com/dracory/neat/contracts/database/schema"
)

func TestMySQLIntegrationSpatial(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)

	tableName := "test_mysql_spatial"
	_ = db.Schema().DropIfExists(tableName)

	err := db.Schema().Create(tableName, func(table schema.Blueprint) {
		table.ID()
		table.Column("location", "point")
	})
	if err != nil {
		t.Fatalf("Failed to create spatial table: %v", err)
	}

	// Test Insert with spatial data
	// Using ST_PointFromText or ST_GeomFromText
	err = db.Query().Table(tableName).Create(map[string]any{
		"location": db.Query().Raw("ST_PointFromText(?)", "POINT(1 1)"),
	})
	if err != nil {
		t.Errorf("Failed to insert spatial data: %v", err)
	}

	err = db.Query().Table(tableName).Create(map[string]any{
		"location": db.Query().Raw("ST_PointFromText(?)", "POINT(10 10)"),
	})
	if err != nil {
		t.Errorf("Failed to insert second spatial data: %v", err)
	}

	// Test Query using Spatial functions
	// Find points within a certain distance or area
	var results []map[string]any
	// ST_Distance_Sphere returns distance in meters
	err = db.Query().Table(tableName).
		Where("ST_Distance_Sphere(location, ST_PointFromText(?)) < ?", "POINT(1.00001 1.00001)", 1000).
		Find(&results)
	if err != nil {
		t.Errorf("Spatial query failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	// Test ST_X and ST_Y
	var coords struct {
		X float64 `gorm:"column:x"`
		Y float64 `gorm:"column:y"`
	}
	err = db.Query().Table(tableName).
		Select("ST_X(location) as x, ST_Y(location) as y").
		Where("ST_X(location) = ?", 10).
		First(&coords)
	if err != nil {
		t.Errorf("ST_X/ST_Y query failed: %v", err)
	}
	if coords.X != 10.0 {
		t.Errorf("Expected X 10.0, got %f", coords.X)
	}
	if coords.Y != 10.0 {
		t.Errorf("Expected Y 10.0, got %f", coords.Y)
	}

	_ = db.Schema().Drop(tableName)
}
