//go:build integration

package aztables

import (
	"fmt"
	"testing"
)

// TestAztablesConnection verifies that neat can connect to Azurite via the
// aztablessql driver and execute a simple SHOW TABLES query.
func TestAztablesConnection(t *testing.T) {
	db := SetupAztablesConnection(t)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("DB(): %v", err)
	}

	rows, err := sqlDB.Query("SHOW TABLES")
	if err != nil {
		t.Fatalf("SHOW TABLES: %v", err)
	}
	defer rows.Close()

	// Just verify we can iterate without error; the table list may be empty.
	for rows.Next() {
		// drain
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err: %v", err)
	}
}

// TestAztablesCreateAndDropTable verifies CREATE TABLE and DROP TABLE work
// through the neat → aztablessql → Azurite pipeline.
func TestAztablesCreateAndDropTable(t *testing.T) {
	db := SetupAztablesConnection(t)
	table := uniqueTableName("crudtbl")
	createTable(t, db, table)
	dropTable(t, db, table)
}

// TestAztablesInsertSelectPointRead verifies a full INSERT → point-read
// SELECT cycle. A point read is WHERE PartitionKey = ? AND RowKey = ?,
// which aztablessql maps to GetEntity (the most efficient read path).
func TestAztablesInsertSelectPointRead(t *testing.T) {
	db := SetupAztablesConnection(t)
	table := uniqueTableName("insrd")
	createTable(t, db, table)
	t.Cleanup(func() { dropTable(t, db, table) })

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("DB(): %v", err)
	}

	// INSERT
	_, err = sqlDB.Exec(
		fmt.Sprintf("INSERT INTO %s (PartitionKey, RowKey, Name, Age) VALUES (?, ?, ?, ?)", table),
		"pk1", "rk1", "Ada Lovelace", int64(36),
	)
	if err != nil {
		t.Fatalf("INSERT: %v", err)
	}

	// Point read via SELECT * WHERE PartitionKey = ? AND RowKey = ?
	rows, err := sqlDB.Query(
		fmt.Sprintf("SELECT * FROM %s WHERE PartitionKey = ? AND RowKey = ?", table),
		"pk1", "rk1",
	)
	if err != nil {
		t.Fatalf("SELECT: %v", err)
	}
	row := scanOneRow(t, rows)

	if row["PartitionKey"] != "pk1" {
		t.Errorf("PartitionKey = %v, want pk1", row["PartitionKey"])
	}
	if row["RowKey"] != "rk1" {
		t.Errorf("RowKey = %v, want rk1", row["RowKey"])
	}
	if row["Name"] != "Ada Lovelace" {
		t.Errorf("Name = %v, want Ada Lovelace", row["Name"])
	}
}

// TestAztablesUpdate verifies UPDATE with merge semantics (only SET columns
// are touched; existing properties are preserved).
func TestAztablesUpdate(t *testing.T) {
	db := SetupAztablesConnection(t)
	table := uniqueTableName("upd")
	createTable(t, db, table)
	t.Cleanup(func() { dropTable(t, db, table) })

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("DB(): %v", err)
	}

	// Insert initial entity
	_, err = sqlDB.Exec(
		fmt.Sprintf("INSERT INTO %s (PartitionKey, RowKey, Name, Age) VALUES (?, ?, ?, ?)", table),
		"pk1", "rk1", "Ada", int64(36),
	)
	if err != nil {
		t.Fatalf("INSERT: %v", err)
	}

	// UPDATE — merge semantics, only Age is touched, Name is preserved
	_, err = sqlDB.Exec(
		fmt.Sprintf("UPDATE %s SET Age = ? WHERE PartitionKey = ? AND RowKey = ?", table),
		int64(37), "pk1", "rk1",
	)
	if err != nil {
		t.Fatalf("UPDATE: %v", err)
	}

	// Verify
	rows, err := sqlDB.Query(
		fmt.Sprintf("SELECT * FROM %s WHERE PartitionKey = ? AND RowKey = ?", table),
		"pk1", "rk1",
	)
	if err != nil {
		t.Fatalf("SELECT: %v", err)
	}
	row := scanOneRow(t, rows)

	if row["Name"] != "Ada" {
		t.Errorf("Name = %v, want Ada (merge should preserve Name)", row["Name"])
	}
	// Age comes back as float64 from JSON decode
	if age, ok := toFloat64(row["Age"]); !ok || age != 37 {
		t.Errorf("Age = %v, want 37", row["Age"])
	}
}

// TestAztablesDelete verifies DELETE removes an entity.
func TestAztablesDelete(t *testing.T) {
	db := SetupAztablesConnection(t)
	table := uniqueTableName("del")
	createTable(t, db, table)
	t.Cleanup(func() { dropTable(t, db, table) })

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("DB(): %v", err)
	}

	// Insert
	_, err = sqlDB.Exec(
		fmt.Sprintf("INSERT INTO %s (PartitionKey, RowKey, Name) VALUES (?, ?, ?)", table),
		"pk1", "rk1", "To Delete",
	)
	if err != nil {
		t.Fatalf("INSERT: %v", err)
	}

	// DELETE
	_, err = sqlDB.Exec(
		fmt.Sprintf("DELETE FROM %s WHERE PartitionKey = ? AND RowKey = ?", table),
		"pk1", "rk1",
	)
	if err != nil {
		t.Fatalf("DELETE: %v", err)
	}

	// Verify entity is gone
	rows, err := sqlDB.Query(
		fmt.Sprintf("SELECT * FROM %s WHERE PartitionKey = ? AND RowKey = ?", table),
		"pk1", "rk1",
	)
	if err != nil {
		t.Fatalf("SELECT: %v", err)
	}
	defer rows.Close()
	if rows.Next() {
		t.Error("expected 0 rows after DELETE, got at least 1")
	}
}

// TestAztablesUpsertReplace verifies INSERT OR REPLACE semantics — the
// entity is fully replaced, properties not in the column list are dropped.
func TestAztablesUpsertReplace(t *testing.T) {
	db := SetupAztablesConnection(t)
	table := uniqueTableName("upsr")
	createTable(t, db, table)
	t.Cleanup(func() { dropTable(t, db, table) })

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("DB(): %v", err)
	}

	// Insert initial entity with Name and Age
	_, err = sqlDB.Exec(
		fmt.Sprintf("INSERT INTO %s (PartitionKey, RowKey, Name, Age) VALUES (?, ?, ?, ?)", table),
		"pk1", "rk1", "Ada", 36,
	)
	if err != nil {
		t.Fatalf("INSERT: %v", err)
	}

	// INSERT OR REPLACE — only Name, Age should be dropped (replace semantics)
	_, err = sqlDB.Exec(
		fmt.Sprintf("INSERT OR REPLACE INTO %s (PartitionKey, RowKey, Name) VALUES (?, ?, ?)", table),
		"pk1", "rk1", "Ada Updated",
	)
	if err != nil {
		t.Fatalf("INSERT OR REPLACE: %v", err)
	}

	// Verify: Name is updated, Age should be gone (replaced)
	rows, err := sqlDB.Query(
		fmt.Sprintf("SELECT * FROM %s WHERE PartitionKey = ? AND RowKey = ?", table),
		"pk1", "rk1",
	)
	if err != nil {
		t.Fatalf("SELECT: %v", err)
	}
	row := scanOneRow(t, rows)

	if row["Name"] != "Ada Updated" {
		t.Errorf("Name = %v, want Ada Updated", row["Name"])
	}
	if _, exists := row["Age"]; exists {
		t.Errorf("Age = %v, expected property to be absent after INSERT OR REPLACE", row["Age"])
	}
}

// toFloat64 converts a value (as returned by database/sql scanning an
// aztablessql result, which decodes JSON numbers as float64) to float64.
func toFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int64:
		return float64(n), true
	case int32:
		return float64(n), true
	case int:
		return float64(n), true
	case []byte:
		// aztablessql may return values as []byte depending on the scan path
		s := string(n)
		var f float64
		_, err := fmt.Sscanf(s, "%g", &f)
		if err != nil {
			return 0, false
		}
		return f, true
	case string:
		var f float64
		_, err := fmt.Sscanf(n, "%g", &f)
		if err != nil {
			return 0, false
		}
		return f, true
	case nil:
		return 0, false
	default:
		return 0, false
	}
}
