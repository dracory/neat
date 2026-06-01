package query

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestSavePoint(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.inTransaction = true
	q.tx, _ = db.Begin()

	err = q.SavePoint("test_savepoint")
	if err != nil {
		t.Errorf("SavePoint failed: %v", err)
	}

	_ = q.tx.Rollback()
}

func TestSavePointNotInTransaction(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)

	err := q.SavePoint("test_savepoint")
	if err == nil {
		t.Error("Expected error when not in transaction")
		return
	}

	if err.Error() != "not in a transaction" {
		t.Errorf("Expected 'not in a transaction' error, got: %v", err)
	}
}

func TestRollbackTo(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.inTransaction = true
	q.tx, _ = db.Begin()

	_, err = q.tx.Exec("INSERT INTO test (name) VALUES ('before_savepoint')")
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	err = q.SavePoint("sp1")
	if err != nil {
		t.Fatalf("SavePoint failed: %v", err)
	}

	_, err = q.tx.Exec("INSERT INTO test (name) VALUES ('after_savepoint')")
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	err = q.RollbackTo("sp1")
	if err != nil {
		t.Errorf("RollbackTo failed: %v", err)
	}

	var count int
	err = q.tx.QueryRow("SELECT COUNT(*) FROM test WHERE name = 'after_savepoint'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 rows after rollback, got %d", count)
	}

	_ = q.tx.Rollback()
}

func TestRollbackToNotInTransaction(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)

	err := q.RollbackTo("test_savepoint")
	if err == nil {
		t.Error("Expected error when not in transaction")
		return
	}

	if err.Error() != "not in a transaction" {
		t.Errorf("Expected 'not in a transaction' error, got: %v", err)
	}
}

func TestSavepointCreationAndRollback(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.inTransaction = true
	q.tx, _ = db.Begin()

	_, err = q.tx.Exec("INSERT INTO test (name) VALUES ('initial')")
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	err = q.SavePoint("sp1")
	if err != nil {
		t.Fatalf("SavePoint failed: %v", err)
	}

	_, err = q.tx.Exec("INSERT INTO test (name) VALUES ('after_sp1')")
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	err = q.RollbackTo("sp1")
	if err != nil {
		t.Fatalf("RollbackTo failed: %v", err)
	}

	var count int
	err = q.tx.QueryRow("SELECT COUNT(*) FROM test").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 row after rollback, got %d", count)
	}

	_ = q.tx.Rollback()
}

func TestNestedSavepointLevels(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.inTransaction = true
	q.tx, _ = db.Begin()

	_, err = q.tx.Exec("INSERT INTO test (name) VALUES ('level0')")
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	err = q.SavePoint("sp1")
	if err != nil {
		t.Fatalf("SavePoint sp1 failed: %v", err)
	}

	_, err = q.tx.Exec("INSERT INTO test (name) VALUES ('level1')")
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	err = q.SavePoint("sp2")
	if err != nil {
		t.Fatalf("SavePoint sp2 failed: %v", err)
	}

	_, err = q.tx.Exec("INSERT INTO test (name) VALUES ('level2')")
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	err = q.RollbackTo("sp1")
	if err != nil {
		t.Fatalf("RollbackTo sp1 failed: %v", err)
	}

	var count int
	err = q.tx.QueryRow("SELECT COUNT(*) FROM test").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 row after rollback to sp1, got %d", count)
	}

	_ = q.tx.Rollback()
}

func TestSavepointErrorHandling(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.inTransaction = true
	q.tx, _ = db.Begin()

	err = q.SavePoint("invalid savepoint name with spaces")
	if err == nil {
		t.Error("Expected error for invalid savepoint name")
	}

	_ = q.tx.Rollback()
}

func TestRollbackToInvalidSavepoint(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.inTransaction = true
	q.tx, _ = db.Begin()

	err = q.RollbackTo("nonexistent_savepoint")
	if err == nil {
		t.Error("Expected error for nonexistent savepoint")
	}

	_ = q.tx.Rollback()
}

func TestBeginCreatesSavepointForNestedTransaction(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.inTransaction = true
	q.tx, _ = db.Begin()
	q.savepointLevel = 0

	nestedQ, err := q.Begin()
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	nestedQuery := nestedQ.(*Query)
	if nestedQuery.savepointName == "" {
		t.Error("Expected savepoint name to be set for nested transaction")
	}

	if nestedQuery.savepointLevel != 1 {
		t.Errorf("Expected savepoint level 1, got %d", nestedQuery.savepointLevel)
	}

	_ = q.tx.Rollback()
}

func TestCommitReleasesSavepoint(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.inTransaction = true
	q.tx, _ = db.Begin()

	_, err = q.tx.Exec("INSERT INTO test (name) VALUES ('test')")
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	err = q.SavePoint("test_sp")
	if err != nil {
		t.Fatalf("SavePoint failed: %v", err)
	}

	q.savepointName = "test_sp"

	err = q.Commit()
	if err != nil {
		t.Errorf("Commit failed: %v", err)
	}

	if q.savepointName != "" {
		t.Error("Expected savepoint name to be cleared after commit")
	}
}

func TestRollbackReleasesSavepoint(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.inTransaction = true
	q.tx, _ = db.Begin()

	_, err = q.tx.Exec("INSERT INTO test (name) VALUES ('test')")
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	err = q.SavePoint("test_sp")
	if err != nil {
		t.Fatalf("SavePoint failed: %v", err)
	}

	q.savepointName = "test_sp"

	err = q.Rollback()
	if err != nil {
		t.Errorf("Rollback failed: %v", err)
	}

	if q.savepointName != "" {
		t.Error("Expected savepoint name to be cleared after rollback")
	}
}
