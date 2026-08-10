//go:build integration

package jsondb

import (
	"testing"
)

func TestJSONDBIntegrationCountAll(t *testing.T) {
	db := SetupJSONDBTest(t)

	var result []jsondbCountResult
	err := db.Query().
		Table("users").
		Select("COUNT(*) AS cnt").
		Get(&result)
	if err != nil {
		t.Fatalf("failed to count users: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 row, got %d", len(result))
	}
	if result[0].Count != 3 {
		t.Errorf("expected count 3, got %d", result[0].Count)
	}
}

func TestJSONDBIntegrationCountWithWhere(t *testing.T) {
	db := SetupJSONDBTest(t)

	var result []jsondbCountResult
	err := db.Query().
		Table("users").
		Select("COUNT(*) AS cnt").
		Where("active = ?", true).
		Get(&result)
	if err != nil {
		t.Fatalf("failed to count active users: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 row, got %d", len(result))
	}
	if result[0].Count != 2 {
		t.Errorf("expected count 2, got %d", result[0].Count)
	}
}

func TestJSONDBIntegrationSum(t *testing.T) {
	db := SetupJSONDBTest(t)

	var result []jsondbSumResult
	err := db.Query().
		Table("orders").
		Select("SUM(total) AS total_sum").
		Get(&result)
	if err != nil {
		t.Fatalf("failed to sum orders: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 row, got %d", len(result))
	}
	expectedSum := 149.97 + 99.95 + 99.99
	if result[0].Total < expectedSum-0.01 || result[0].Total > expectedSum+0.01 {
		t.Errorf("expected sum ~%.2f, got %.2f", expectedSum, result[0].Total)
	}
}

func TestJSONDBIntegrationAvg(t *testing.T) {
	db := SetupJSONDBTest(t)

	var result []jsondbAvgResult
	err := db.Query().
		Table("products").
		Select("AVG(price) AS price_avg").
		Get(&result)
	if err != nil {
		t.Fatalf("failed to avg products: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 row, got %d", len(result))
	}
	expectedAvg := (19.99 + 49.99 + 99.99) / 3
	if result[0].Avg < expectedAvg-0.01 || result[0].Avg > expectedAvg+0.01 {
		t.Errorf("expected avg ~%.2f, got %.2f", expectedAvg, result[0].Avg)
	}
}

func TestJSONDBIntegrationMinMax(t *testing.T) {
	db := SetupJSONDBTest(t)

	var result []jsondbMinMaxResult
	err := db.Query().
		Table("products").
		Select("MIN(price) AS min_price, MAX(price) AS max_price").
		Get(&result)
	if err != nil {
		t.Fatalf("failed to min/max products: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 row, got %d", len(result))
	}
	if result[0].MinPrice != 19.99 {
		t.Errorf("expected min 19.99, got %f", result[0].MinPrice)
	}
	if result[0].MaxPrice != 99.99 {
		t.Errorf("expected max 99.99, got %f", result[0].MaxPrice)
	}
}

func TestJSONDBIntegrationGroupBy(t *testing.T) {
	db := SetupJSONDBTest(t)

	type categoryCount struct {
		Category string  `db:"category"`
		Count    int     `db:"cnt"`
		AvgPrice float64 `db:"avg_price"`
	}

	var results []categoryCount
	err := db.Query().
		Table("products").
		Select("category, COUNT(*) AS cnt, AVG(price) AS avg_price").
		Group("category").
		OrderBy("category", "asc").
		Get(&results)
	if err != nil {
		t.Fatalf("failed to group by category: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(results))
	}
	if results[0].Category != "Electronics" {
		t.Errorf("expected first category 'Electronics', got '%s'", results[0].Category)
	}
	if Stat := results[0].Count; Stat != 2 {
		t.Errorf("expected Electronics count 2, got %d", Stat)
	}
	expectedAvg := (49.99 + 99.99) / 2
	if results[0].AvgPrice < expectedAvg-0.01 || results[0].AvgPrice > expectedAvg+0.01 {
		t.Errorf("expected Electronics avg price ~%.2f, got %.2f", expectedAvg, results[0].AvgPrice)
	}

	if results[1].Category != "Hardware" {
		t.Errorf("expected second category 'Hardware', got '%s'", results[1].Category)
	}
	if results[1].Count != 1 {
		t.Errorf("expected Hardware count 1, got %d", results[1].Count)
	}
}

func TestJSONDBIntegrationSumWithJoin(t *testing.T) {
	db := SetupJSONDBTest(t)

	type userTotal struct {
		UserName string  `db:"user_name"`
		TotalSum float64 `db:"total_sum"`
	}

	var results []userTotal
	err := db.Query().
		Table("orders AS o").
		LeftJoin("users AS u ON o.user_id = u.id").
		Select("u.name AS user_name, SUM(o.total) AS total_sum").
		Group("u.name").
		OrderBy("u.name", "asc").
		Get(&results)
	if err != nil {
		t.Fatalf("failed to sum with join: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 users with orders, got %d", len(results))
	}
	// Alice has orders 1 (149.97) and 3 (99.99) → 249.96
	if results[0].UserName != "Alice" {
		t.Errorf("expected first user 'Alice', got '%s'", results[0].UserName)
	}
	aliceSum := 149.97 + 99.99
	if results[0].TotalSum < aliceSum-0.01 || results[0].TotalSum > aliceSum+0.01 {
		t.Errorf("expected Alice sum ~%.2f, got %.2f", aliceSum, results[0].TotalSum)
	}
	// Charlie has order 2 (99.95)
	if results[1].UserName != "Charlie" {
		t.Errorf("expected second user 'Charlie', got '%s'", results[1].UserName)
	}
	if results[1].TotalSum != 99.95 {
		t.Errorf("expected Charlie sum 99.95, got %.2f", results[1].TotalSum)
	}
}
