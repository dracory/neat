//go:build integration

package godb

import (
	"testing"
)

func TestGODB_Aggregate_Sum_Avg_MinMax(t *testing.T) {
	db := SetupGODBTest(t)

	// SUM total of all orders
	var sum float64
	err := db.Query().Table("orders").Sum("total", &sum)
	if err != nil {
		t.Fatalf("Sum failed: %v", err)
	}
	if sum < 349.90 || sum > 349.92 {
		t.Errorf("expected SUM ~349.91, got %f", sum)
	}

	// AVG price of products
	var avg float64
	err = db.Query().Table("products").Avg("price", &avg)
	if err != nil {
		t.Fatalf("Avg failed: %v", err)
	}
	if avg < 56.65 || avg > 56.66 {
		t.Errorf("expected AVG ~56.66, got %f", avg)
	}

	// MIN price of products
	var min float64
	err = db.Query().Table("products").Min("price", &min)
	if err != nil {
		t.Fatalf("Min failed: %v", err)
	}
	if min != 19.99 {
		t.Errorf("expected MIN 19.99, got %f", min)
	}

	// MAX price of products
	var max float64
	err = db.Query().Table("products").Max("price", &max)
	if err != nil {
		t.Fatalf("Max failed: %v", err)
	}
	if max != 99.99 {
		t.Errorf("expected MAX 99.99, got %f", max)
	}
}

func TestGODB_Aggregate_GroupBy(t *testing.T) {
	db := SetupGODBTest(t)

	type CategoryCount struct {
		Category string `db:"category"`
		Cnt      int    `db:"cnt"`
	}

	var results []CategoryCount
	err := db.Query().
		Table("products").
		Group("category").
		Select("category", "count(*) AS cnt").
		OrderBy("category", "asc").
		Get(&results)

	if err != nil {
		t.Fatalf("Group failed: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(results))
	}

	// Electronics (2), Hardware (1)
	if results[0].Category != "Electronics" || results[0].Cnt != 2 {
		t.Errorf("unexpected first group: %+v", results[0])
	}
	if results[1].Category != "Hardware" || results[1].Cnt != 1 {
		t.Errorf("unexpected second group: %+v", results[1])
	}
}
