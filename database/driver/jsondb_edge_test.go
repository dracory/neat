package driver

import (
	"testing"
)

func TestJSONDBEmptyObjectsSkipsTable(t *testing.T) {
	dir := mkdirTempJSONDB(t)
	writeJSONFile(t, dir, "empty_objs.json", "[{}]")

	d := NewJSONDB()
	db, err := d.Open(dir)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='empty_objs'").Scan(&count); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected [{}] to be skipped (no columns = no table), got count=%d", count)
	}
}

func TestJSONDBNullObjectsSkipsTable(t *testing.T) {
	dir := mkdirTempJSONDB(t)
	writeJSONFile(t, dir, "null_objs.json", "[null]")

	d := NewJSONDB()
	db, err := d.Open(dir)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='null_objs'").Scan(&count); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected [null] to be skipped (no columns = no table), got count=%d", count)
	}
}
