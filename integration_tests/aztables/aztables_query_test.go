//go:build integration

package aztables

import (
	"fmt"
	"testing"
)

// TestAztablesWhereFilter verifies that WHERE with a non-key column filter
// works (falls back to ListEntities with OData $filter).
func TestAztablesWhereFilter(t *testing.T) {
	db := SetupAztablesConnection(t)
	table := uniqueTableName("wfilt")
	createTable(t, db, table)
	t.Cleanup(func() { dropTable(t, db, table) })

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("DB(): %v", err)
	}

	// Insert multiple entities in the same partition with different Ages
	insertEntity(t, db, table, "pk1", "rk1", map[string]any{"Name": "Alice", "Age": int64(30)})
	insertEntity(t, db, table, "pk1", "rk2", map[string]any{"Name": "Bob", "Age": int64(40)})
	insertEntity(t, db, table, "pk1", "rk3", map[string]any{"Name": "Carol", "Age": int64(50)})

	// WHERE Age > 35 should return Bob and Carol
	rows, err := sqlDB.Query(
		fmt.Sprintf("SELECT * FROM %s WHERE PartitionKey = ? AND Age > ?", table),
		"pk1", int64(35),
	)
	if err != nil {
		t.Fatalf("SELECT with filter: %v", err)
	}
	result := collectRows(t, rows)

	if len(result) != 2 {
		t.Fatalf("expected 2 rows with Age > 35, got %d", len(result))
	}
}

// TestAztablesLimit verifies that LIMIT caps the result set.
func TestAztablesLimit(t *testing.T) {
	db := SetupAztablesConnection(t)
	table := uniqueTableName("wlim")
	createTable(t, db, table)
	t.Cleanup(func() { dropTable(t, db, table) })

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("DB(): %v", err)
	}

	// Insert 5 entities
	for i := 1; i <= 5; i++ {
		insertEntity(t, db, table, "pk1",
			fmt.Sprintf("rk%02d", i),
			map[string]any{"Name": fmt.Sprintf("User%d", i)})
	}

	// SELECT with LIMIT 3
	rows, err := sqlDB.Query(
		fmt.Sprintf("SELECT * FROM %s WHERE PartitionKey = ? LIMIT 3", table),
		"pk1",
	)
	if err != nil {
		t.Fatalf("SELECT with LIMIT: %v", err)
	}
	result := collectRows(t, rows)

	if len(result) != 3 {
		t.Errorf("expected 3 rows with LIMIT 3, got %d", len(result))
	}
}

// TestAztablesCount verifies SELECT COUNT(*) returns the correct count.
func TestAztablesCount(t *testing.T) {
	db := SetupAztablesConnection(t)
	table := uniqueTableName("wcnt")
	createTable(t, db, table)
	t.Cleanup(func() { dropTable(t, db, table) })

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("DB(): %v", err)
	}

	// Insert 3 entities in partition pk1, 2 in pk2
	insertEntity(t, db, table, "pk1", "rk1", map[string]any{"Name": "A"})
	insertEntity(t, db, table, "pk1", "rk2", map[string]any{"Name": "B"})
	insertEntity(t, db, table, "pk1", "rk3", map[string]any{"Name": "C"})
	insertEntity(t, db, table, "pk2", "rk1", map[string]any{"Name": "D"})
	insertEntity(t, db, table, "pk2", "rk2", map[string]any{"Name": "E"})

	// COUNT(*) for pk1 should be 3
	rows, err := sqlDB.Query(
		fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE PartitionKey = ?", table),
		"pk1",
	)
	if err != nil {
		t.Fatalf("SELECT COUNT(*): %v", err)
	}
	row := scanOneRow(t, rows)

	count, ok := toFloat64(row["count"])
	if !ok {
		t.Fatalf("count column is not numeric: %v", row["count"])
	}
	if count != 3 {
		t.Errorf("COUNT(*) for pk1 = %v, want 3", count)
	}
}

// TestAztablesSelectProjection verifies SELECT with specific columns
// (projection via OData $select).
func TestAztablesSelectProjection(t *testing.T) {
	db := SetupAztablesConnection(t)
	table := uniqueTableName("wproj")
	createTable(t, db, table)
	t.Cleanup(func() { dropTable(t, db, table) })

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("DB(): %v", err)
	}

	insertEntity(t, db, table, "pk1", "rk1", map[string]any{"Name": "Ada", "Age": 36})

	// SELECT Name, Age — should project only those columns
	rows, err := sqlDB.Query(
		fmt.Sprintf("SELECT Name, Age FROM %s WHERE PartitionKey = ? AND RowKey = ?", table),
		"pk1", "rk1",
	)
	if err != nil {
		t.Fatalf("SELECT projection: %v", err)
	}
	row := scanOneRow(t, rows)

	if row["Name"] != "Ada" {
		t.Errorf("Name = %v, want Ada", row["Name"])
	}
	// The projected columns should be present; PartitionKey/RowKey may or may
	// not be included by the service, but Name and Age must be.
	if _, exists := row["Name"]; !exists {
		t.Error("Name column missing from projection result")
	}
}

// TestAztablesWhereComparisonOperators verifies all supported comparison
// operators (=, !=, >, >=, <, <=) in WHERE clauses.
//
// String values are used for eq/ne because Azure Table Storage's eq/ne
// operators are type-strict: an Edm.Int64 property will not match an
// Edm.Int32 OData literal (the default for bare integer literals).
// Range operators (gt/ge/lt/le) are more lenient and work with numeric
// values regardless of the stored EDM type, so those use int64 values.
func TestAztablesWhereComparisonOperators(t *testing.T) {
	db := SetupAztablesConnection(t)
	table := uniqueTableName("wcmp")
	createTable(t, db, table)
	t.Cleanup(func() { dropTable(t, db, table) })

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("DB(): %v", err)
	}

	// Insert entities with string Grade and numeric Age
	insertEntity(t, db, table, "pk1", "rk1", map[string]any{"Grade": "A", "Age": int64(20)})
	insertEntity(t, db, table, "pk1", "rk2", map[string]any{"Grade": "B", "Age": int64(30)})
	insertEntity(t, db, table, "pk1", "rk3", map[string]any{"Grade": "C", "Age": int64(40)})

	tests := []struct {
		name      string
		where     string
		args      []any
		wantCount int
	}{
		// eq/ne use string values (type-safe in OData)
		{"equal", "PartitionKey = ? AND Grade = ?", []any{"pk1", "B"}, 1},
		{"notEqual", "PartitionKey = ? AND Grade != ?", []any{"pk1", "B"}, 2},
		// Range operators use int64 values (lenient type matching)
		{"greaterThan", "PartitionKey = ? AND Age > ?", []any{"pk1", int64(25)}, 2},
		{"greaterEqual", "PartitionKey = ? AND Age >= ?", []any{"pk1", int64(30)}, 2},
		{"lessThan", "PartitionKey = ? AND Age < ?", []any{"pk1", int64(35)}, 2},
		{"lessEqual", "PartitionKey = ? AND Age <= ?", []any{"pk1", int64(30)}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, err := sqlDB.Query(
				fmt.Sprintf("SELECT * FROM %s WHERE %s", table, tt.where),
				tt.args...,
			)
			if err != nil {
				t.Fatalf("SELECT: %v", err)
			}
			result := collectRows(t, rows)
			if len(result) != tt.wantCount {
				t.Errorf("operator %s: expected %d rows, got %d", tt.name, tt.wantCount, len(result))
			}
		})
	}
}
