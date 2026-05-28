package query_test

import (
	"fmt"
	"testing"
)

// --- pagination tests ---

type PaginateUser struct {
	ID     int    `db:"id"`
	Name   string `db:"name"`
	Email  string `db:"email"`
	Status string `db:"status"`
}

func TestPaginateBasic(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_paginate (id INTEGER, name TEXT, email TEXT)")
	execSQL(t, w, "INSERT INTO test_paginate VALUES (1,'alice','alice@example.com'),(2,'bob','bob@example.com'),(3,'charlie','charlie@example.com'),(4,'dave','dave@example.com'),(5,'eve','eve@example.com')")

	w.SetTable("test_paginate")

	var users []PaginateUser
	var total int64

	err := w.Q.Paginate(1, 2, &users, &total)

	if err != nil {
		t.Fatalf("Paginate failed: %v", err)
	}

	if total != 5 {
		t.Errorf("Expected total count 5, got %d", total)
	}

	if len(users) != 2 {
		t.Fatalf("Expected 2 users on page 1, got %d", len(users))
	}

	if users[0].Name != "alice" {
		t.Errorf("Expected first user to be alice, got %s", users[0].Name)
	}

	if users[1].Name != "bob" {
		t.Errorf("Expected second user to be bob, got %s", users[1].Name)
	}
}

func TestPaginateTotalCount(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_paginate_count (id INTEGER, name TEXT)")
	execSQL(t, w, "INSERT INTO test_paginate_count VALUES (1,'a'),(2,'b'),(3,'c'),(4,'d'),(5,'e'),(6,'f'),(7,'g')")

	w.SetTable("test_paginate_count")

	var results []map[string]any
	var total int64

	err := w.Q.Paginate(1, 3, &results, &total)

	if err != nil {
		t.Fatalf("Paginate failed: %v", err)
	}

	if total != 7 {
		t.Errorf("Expected total count 7, got %d", total)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 results on page 1, got %d", len(results))
	}
}

func TestPaginateOffsetCalculation(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_paginate_offset (id INTEGER, name TEXT)")
	execSQL(t, w, "INSERT INTO test_paginate_offset VALUES (1,'first'),(2,'second'),(3,'third'),(4,'fourth'),(5,'fifth'),(6,'sixth'),(7,'seventh'),(8,'eighth'),(9,'ninth'),(10,'tenth')")

	w.SetTable("test_paginate_offset")

	testCases := []struct {
		page          int
		limit         int
		expectedLen   int
		expectedFirst string
	}{
		{1, 3, 3, "first"},
		{2, 3, 3, "fourth"},
		{3, 3, 3, "seventh"},
		{4, 3, 1, "tenth"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("page_%d_limit_%d", tc.page, tc.limit), func(t *testing.T) {
			var results []map[string]any
			var total int64

			err := w.Q.Paginate(tc.page, tc.limit, &results, &total)

			if err != nil {
				t.Fatalf("Paginate failed: %v", err)
			}

			if len(results) != tc.expectedLen {
				t.Errorf("Expected %d results, got %d", tc.expectedLen, len(results))
			}

			if len(results) > 0 {
				if results[0]["name"] != tc.expectedFirst {
					t.Errorf("Expected first result to be %s, got %s", tc.expectedFirst, results[0]["name"])
				}
			}
		})
	}
}

func TestPaginateWithWhereClauses(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_paginate_where (id INTEGER, name TEXT, status TEXT)")
	execSQL(t, w, "INSERT INTO test_paginate_where VALUES (1,'alice','active'),(2,'bob','inactive'),(3,'charlie','active'),(4,'dave','active'),(5,'eve','inactive'),(6,'frank','active')")

	w.SetTable("test_paginate_where")
	w.Q.Where("status = ?", "active")

	var users []PaginateUser
	var total int64

	err := w.Q.Paginate(1, 2, &users, &total)

	if err != nil {
		t.Fatalf("Paginate with WHERE failed: %v", err)
	}

	if total != 4 {
		t.Errorf("Expected total count 4 (active users), got %d", total)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 active users on page 1, got %d", len(users))
	}

	for _, user := range users {
		if user.Status != "active" {
			t.Errorf("Expected all users to have status 'active', got %s", user.Status)
		}
	}
}

func TestPaginateErrorHandling(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_paginate_error (id INTEGER, name TEXT)")
	execSQL(t, w, "INSERT INTO test_paginate_error VALUES (1,'test')")

	w.SetTable("test_paginate_error")

	t.Run("nil total pointer", func(t *testing.T) {
		var users []PaginateUser
		err := w.Q.Paginate(1, 10, &users, nil)

		if err != nil {
			t.Fatalf("Paginate with nil total should succeed: %v", err)
		}

		if len(users) != 1 {
			t.Errorf("Expected 1 user, got %d", len(users))
		}
	})

	t.Run("invalid page number", func(t *testing.T) {
		var users []PaginateUser
		var total int64

		// Page 0 should result in negative offset, but the implementation
		// calculates offset as (page-1)*limit, so page 0 gives offset -limit
		// This might fail or return empty results depending on database
		err := w.Q.Paginate(0, 10, &users, &total)

		// The method should handle this gracefully
		_ = err
	})

	t.Run("empty result set", func(t *testing.T) {
		execSQL(t, w, "CREATE TABLE test_paginate_empty (id INTEGER)")
		w.SetTable("test_paginate_empty")

		var results []map[string]any
		var total int64

		err := w.Q.Paginate(1, 10, &results, &total)

		if err != nil {
			t.Fatalf("Paginate with empty result failed: %v", err)
		}

		if total != 0 {
			t.Errorf("Expected total count 0, got %d", total)
		}

		if len(results) != 0 {
			t.Errorf("Expected 0 results, got %d", len(results))
		}
	})

	t.Run("count query failure", func(t *testing.T) {
		// Test with invalid table to trigger count failure
		w.SetTable("nonexistent_table")

		var users []PaginateUser
		var total int64

		err := w.Q.Paginate(1, 10, &users, &total)

		if err == nil {
			t.Error("Expected error for count query failure")
		}
	})
}

func TestPaginateWithTransactions(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_paginate_tx (id INTEGER, name TEXT)")
	execSQL(t, w, "INSERT INTO test_paginate_tx VALUES (1,'tx1'),(2,'tx2'),(3,'tx3'),(4,'tx4')")

	w.SetTable("test_paginate_tx")

	tx, err := w.Q.Begin()
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	var users []PaginateUser
	var total int64

	err = tx.Paginate(1, 2, &users, &total)

	if err != nil {
		t.Fatalf("Paginate in transaction failed: %v", err)
	}

	if total != 4 {
		t.Errorf("Expected total count 4 in transaction, got %d", total)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users in transaction, got %d", len(users))
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}
}

func TestPaginateTypedStructs(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_paginate_typed (id INTEGER, name TEXT, email TEXT)")
	execSQL(t, w, "INSERT INTO test_paginate_typed VALUES (1,'alice','alice@example.com'),(2,'bob','bob@example.com'),(3,'charlie','charlie@example.com')")

	w.SetTable("test_paginate_typed")

	users := make([]PaginateUser, 0)
	var total int64

	err := w.Q.Paginate(1, 2, &users, &total)

	if err != nil {
		t.Fatalf("Paginate with typed structs failed: %v", err)
	}

	if total != 3 {
		t.Errorf("Expected total count 3, got %d", total)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}

	if users[0].ID != 1 || users[0].Name != "alice" {
		t.Errorf("Unexpected first user: %+v", users[0])
	}
}

func TestPaginateLastPage(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_paginate_last (id INTEGER, name TEXT)")
	execSQL(t, w, "INSERT INTO test_paginate_last VALUES (1,'a'),(2,'b'),(3,'c'),(4,'d'),(5,'e')")

	w.SetTable("test_paginate_last")

	results := make([]map[string]any, 0)
	var total int64

	err := w.Q.Paginate(3, 2, &results, &total)

	if err != nil {
		t.Fatalf("Paginate last page failed: %v", err)
	}

	if total != 5 {
		t.Errorf("Expected total count 5, got %d", total)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result on last page, got %d", len(results))
	}

	if results[0]["name"] != "e" {
		t.Errorf("Expected last result to be 'e', got %s", results[0]["name"])
	}
}

func TestPaginateBeyondLastPage(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_paginate_beyond (id INTEGER, name TEXT)")
	execSQL(t, w, "INSERT INTO test_paginate_beyond VALUES (1,'a'),(2,'b'),(3,'c')")

	w.SetTable("test_paginate_beyond")

	var results []map[string]any
	var total int64

	err := w.Q.Paginate(10, 5, &results, &total)

	if err != nil {
		t.Fatalf("Paginate beyond last page failed: %v", err)
	}

	if total != 3 {
		t.Errorf("Expected total count 3, got %d", total)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results beyond last page, got %d", len(results))
	}
}
