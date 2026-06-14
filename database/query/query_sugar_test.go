package query_test

import (
	"testing"
)

// TestCountAsVar tests the CountAsVar sugar method.
func TestCountAsVar(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, email TEXT, age INTEGER)")
	execSQL(t, w, "INSERT INTO users VALUES (1, 'Alice', 'alice@example.com', 20)")
	execSQL(t, w, "INSERT INTO users VALUES (2, 'Bob', 'bob@example.com', 25)")
	execSQL(t, w, "INSERT INTO users VALUES (3, 'Charlie', 'charlie@example.com', 30)")
	execSQL(t, w, "INSERT INTO users VALUES (4, 'Diana', 'diana@example.com', 35)")
	execSQL(t, w, "INSERT INTO users VALUES (5, 'Eve', 'eve@example.com', 40)")

	w.SetTable("users")

	// Test CountAsVar
	count, err := w.Q.CountAsVar()
	if err != nil {
		t.Fatalf("CountAsVar failed: %v", err)
	}
	if count != 5 {
		t.Errorf("Expected count 5, got %d", count)
	}

	// Test with WHERE clause
	w = openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE users2 (id INTEGER PRIMARY KEY, name TEXT, email TEXT, age INTEGER)")
	execSQL(t, w, "INSERT INTO users2 VALUES (1, 'Alice', 'alice@example.com', 20)")
	execSQL(t, w, "INSERT INTO users2 VALUES (2, 'Bob', 'bob@example.com', 25)")
	execSQL(t, w, "INSERT INTO users2 VALUES (3, 'Charlie', 'charlie@example.com', 30)")
	w.SetTable("users2")
	w.Q.Where("age > ?", 25)
	count, err = w.Q.CountAsVar()
	if err != nil {
		t.Fatalf("CountAsVar with WHERE failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}
}

// TestSumAsVar tests the SumAsVar sugar method.
func TestSumAsVar(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, email TEXT, age INTEGER)")
	execSQL(t, w, "INSERT INTO users VALUES (1, 'Alice', 'alice@example.com', 20)")
	execSQL(t, w, "INSERT INTO users VALUES (2, 'Bob', 'bob@example.com', 25)")
	execSQL(t, w, "INSERT INTO users VALUES (3, 'Charlie', 'charlie@example.com', 30)")
	execSQL(t, w, "INSERT INTO users VALUES (4, 'Diana', 'diana@example.com', 35)")
	execSQL(t, w, "INSERT INTO users VALUES (5, 'Eve', 'eve@example.com', 40)")

	w.SetTable("users")

	// Test SumAsVar
	total, err := w.Q.SumAsVar("age")
	if err != nil {
		t.Fatalf("SumAsVar failed: %v", err)
	}
	expectedTotal := 20 + 25 + 30 + 35 + 40
	if total != float64(expectedTotal) {
		t.Errorf("Expected sum %f, got %f", float64(expectedTotal), total)
	}
}

// TestAvgAsVar tests the AvgAsVar sugar method.
func TestAvgAsVar(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, email TEXT, age INTEGER)")
	execSQL(t, w, "INSERT INTO users VALUES (1, 'Alice', 'alice@example.com', 20)")
	execSQL(t, w, "INSERT INTO users VALUES (2, 'Bob', 'bob@example.com', 25)")
	execSQL(t, w, "INSERT INTO users VALUES (3, 'Charlie', 'charlie@example.com', 30)")
	execSQL(t, w, "INSERT INTO users VALUES (4, 'Diana', 'diana@example.com', 35)")
	execSQL(t, w, "INSERT INTO users VALUES (5, 'Eve', 'eve@example.com', 40)")

	w.SetTable("users")

	// Test AvgAsVar
	avg, err := w.Q.AvgAsVar("age")
	if err != nil {
		t.Fatalf("AvgAsVar failed: %v", err)
	}
	expectedAvg := (20 + 25 + 30 + 35 + 40) / 5.0
	if avg != expectedAvg {
		t.Errorf("Expected avg %f, got %f", expectedAvg, avg)
	}
}

// TestMinAsVar tests the MinAsVar sugar method.
func TestMinAsVar(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, email TEXT, age INTEGER)")
	execSQL(t, w, "INSERT INTO users VALUES (1, 'Alice', 'alice@example.com', 20)")
	execSQL(t, w, "INSERT INTO users VALUES (2, 'Bob', 'bob@example.com', 25)")
	execSQL(t, w, "INSERT INTO users VALUES (3, 'Charlie', 'charlie@example.com', 30)")
	execSQL(t, w, "INSERT INTO users VALUES (4, 'Diana', 'diana@example.com', 35)")
	execSQL(t, w, "INSERT INTO users VALUES (5, 'Eve', 'eve@example.com', 40)")

	w.SetTable("users")

	// Test MinAsVar
	min, err := w.Q.MinAsVar("age")
	if err != nil {
		t.Fatalf("MinAsVar failed: %v", err)
	}
	if min != 20 {
		t.Errorf("Expected min 20, got %f", min)
	}
}

// TestMaxAsVar tests the MaxAsVar sugar method.
func TestMaxAsVar(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, email TEXT, age INTEGER)")
	execSQL(t, w, "INSERT INTO users VALUES (1, 'Alice', 'alice@example.com', 20)")
	execSQL(t, w, "INSERT INTO users VALUES (2, 'Bob', 'bob@example.com', 25)")
	execSQL(t, w, "INSERT INTO users VALUES (3, 'Charlie', 'charlie@example.com', 30)")
	execSQL(t, w, "INSERT INTO users VALUES (4, 'Diana', 'diana@example.com', 35)")
	execSQL(t, w, "INSERT INTO users VALUES (5, 'Eve', 'eve@example.com', 40)")

	w.SetTable("users")

	// Test MaxAsVar
	max, err := w.Q.MaxAsVar("age")
	if err != nil {
		t.Fatalf("MaxAsVar failed: %v", err)
	}
	if max != 40 {
		t.Errorf("Expected max 40, got %f", max)
	}
}

// TestExistsAsVar tests the ExistsAsVar sugar method.
func TestExistsAsVar(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, email TEXT, age INTEGER)")
	execSQL(t, w, "INSERT INTO users VALUES (1, 'Alice', 'alice@example.com', 20)")
	execSQL(t, w, "INSERT INTO users VALUES (2, 'Bob', 'bob@example.com', 25)")
	execSQL(t, w, "INSERT INTO users VALUES (3, 'Charlie', 'charlie@example.com', 30)")
	execSQL(t, w, "INSERT INTO users VALUES (4, 'Diana', 'diana@example.com', 35)")
	execSQL(t, w, "INSERT INTO users VALUES (5, 'Eve', 'eve@example.com', 40)")

	w.SetTable("users")

	// Test ExistsAsVar with existing record
	w.Q.Where("age > ?", 25)
	exists, err := w.Q.ExistsAsVar()
	if err != nil {
		t.Fatalf("ExistsAsVar failed: %v", err)
	}
	if !exists {
		t.Error("Expected exists to be true")
	}

	// Test ExistsAsVar with non-existing record
	w = openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE users2 (id INTEGER PRIMARY KEY, name TEXT, email TEXT, age INTEGER)")
	execSQL(t, w, "INSERT INTO users2 VALUES (1, 'Alice', 'alice@example.com', 20)")
	w.SetTable("users2")
	w.Q.Where("age > ?", 100)
	exists, err = w.Q.ExistsAsVar()
	if err != nil {
		t.Fatalf("ExistsAsVar failed: %v", err)
	}
	if exists {
		t.Error("Expected exists to be false")
	}
}

// TestPluckAsVar tests the PluckAsVar sugar method.
func TestPluckAsVar(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, email TEXT, age INTEGER)")
	execSQL(t, w, "INSERT INTO users VALUES (1, 'Alice', 'alice@example.com', 20)")
	execSQL(t, w, "INSERT INTO users VALUES (2, 'Bob', 'bob@example.com', 25)")
	execSQL(t, w, "INSERT INTO users VALUES (3, 'Charlie', 'charlie@example.com', 30)")
	execSQL(t, w, "INSERT INTO users VALUES (4, 'Diana', 'diana@example.com', 35)")
	execSQL(t, w, "INSERT INTO users VALUES (5, 'Eve', 'eve@example.com', 40)")

	w.SetTable("users")

	// Test PluckAsVar with strings
	emailsAny, err := w.Q.PluckAsVar("email")
	if err != nil {
		t.Fatalf("PluckAsVar failed: %v", err)
	}
	// Verify we got the right number of results
	if len(emailsAny) != 5 {
		t.Errorf("Expected 5 emails, got %d", len(emailsAny))
	}

	// Test PluckAsVar with integers
	agesAny, err := w.Q.PluckAsVar("age")
	if err != nil {
		t.Fatalf("PluckAsVar failed: %v", err)
	}
	// Verify we got the right number of results
	if len(agesAny) != 5 {
		t.Errorf("Expected 5 ages, got %d", len(agesAny))
	}
}

// TestValueAsVar tests the ValueAsVar sugar method.
func TestValueAsVar(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, email TEXT, age INTEGER)")
	execSQL(t, w, "INSERT INTO users VALUES (1, 'Alice', 'alice@example.com', 20)")
	execSQL(t, w, "INSERT INTO users VALUES (2, 'Bob', 'bob@example.com', 25)")
	execSQL(t, w, "INSERT INTO users VALUES (3, 'Charlie', 'charlie@example.com', 30)")
	execSQL(t, w, "INSERT INTO users VALUES (4, 'Diana', 'diana@example.com', 35)")
	execSQL(t, w, "INSERT INTO users VALUES (5, 'Eve', 'eve@example.com', 40)")

	w.SetTable("users")

	// Test ValueAsVar with string
	w.Q.Where("age = ?", 20)
	emailAny, err := w.Q.ValueAsVar("email")
	if err != nil {
		t.Fatalf("ValueAsVar failed: %v", err)
	}
	email := emailAny.(string)
	if email != "alice@example.com" {
		t.Errorf("Expected email alice@example.com, got %v", email)
	}

	// Test ValueAsVar with int
	w = openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE users2 (id INTEGER PRIMARY KEY, name TEXT, email TEXT, age INTEGER)")
	execSQL(t, w, "INSERT INTO users2 VALUES (1, 'Alice', 'alice@example.com', 20)")
	w.SetTable("users2")
	w.Q.Where("email = ?", "alice@example.com")
	ageAny, err := w.Q.ValueAsVar("age")
	if err != nil {
		t.Fatalf("ValueAsVar failed: %v", err)
	}
	age := ageAny.(int64)
	if age != 20 {
		t.Errorf("Expected age 20, got %v", age)
	}
}

// TestFirstAsVar tests the FirstAsVar sugar method.
func TestFirstAsVar(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, email TEXT, age INTEGER)")
	execSQL(t, w, "INSERT INTO users VALUES (1, 'Alice', 'alice@example.com', 20)")
	execSQL(t, w, "INSERT INTO users VALUES (2, 'Bob', 'bob@example.com', 25)")
	execSQL(t, w, "INSERT INTO users VALUES (3, 'Charlie', 'charlie@example.com', 30)")
	execSQL(t, w, "INSERT INTO users VALUES (4, 'Diana', 'diana@example.com', 35)")
	execSQL(t, w, "INSERT INTO users VALUES (5, 'Eve', 'eve@example.com', 40)")

	w.SetTable("users")

	// Test FirstAsVar
	w.Q.Where("age > ?", 25)
	userAny, err := w.Q.FirstAsVar()
	if err != nil {
		t.Fatalf("FirstAsVar failed: %v", err)
	}
	// Verify we got a result
	if userAny == nil {
		t.Error("Expected user to not be nil")
	}

	// Test with no results
	w = openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE users2 (id INTEGER PRIMARY KEY, name TEXT, email TEXT, age INTEGER)")
	w.SetTable("users2")
	w.Q.Where("age > ?", 100)
	_, err = w.Q.FirstAsVar()
	if err == nil {
		t.Error("Expected error when no results found")
	}
}

// TestFindOneAsVar tests the FindOneAsVar sugar method (alias for FirstAsVar).
func TestFindOneAsVar(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, email TEXT, age INTEGER)")
	execSQL(t, w, "INSERT INTO users VALUES (1, 'Alice', 'alice@example.com', 20)")
	execSQL(t, w, "INSERT INTO users VALUES (2, 'Bob', 'bob@example.com', 25)")
	execSQL(t, w, "INSERT INTO users VALUES (3, 'Charlie', 'charlie@example.com', 30)")
	execSQL(t, w, "INSERT INTO users VALUES (4, 'Diana', 'diana@example.com', 35)")
	execSQL(t, w, "INSERT INTO users VALUES (5, 'Eve', 'eve@example.com', 40)")

	w.SetTable("users")

	// Test FindOneAsVar
	w.Q.Where("age > ?", 25)
	userAny, err := w.Q.FindOneAsVar()
	if err != nil {
		t.Fatalf("FindOneAsVar failed: %v", err)
	}
	// Verify we got a result
	if userAny == nil {
		t.Error("Expected user to not be nil")
	}
}

// TestGetAsVar tests the GetAsVar sugar method.
func TestGetAsVar(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, email TEXT, age INTEGER)")
	execSQL(t, w, "INSERT INTO users VALUES (1, 'Alice', 'alice@example.com', 20)")
	execSQL(t, w, "INSERT INTO users VALUES (2, 'Bob', 'bob@example.com', 25)")
	execSQL(t, w, "INSERT INTO users VALUES (3, 'Charlie', 'charlie@example.com', 30)")
	execSQL(t, w, "INSERT INTO users VALUES (4, 'Diana', 'diana@example.com', 35)")
	execSQL(t, w, "INSERT INTO users VALUES (5, 'Eve', 'eve@example.com', 40)")

	w.SetTable("users")

	// Test GetAsVar
	w.Q.Where("age > ?", 25)
	usersAny, err := w.Q.GetAsVar()
	if err != nil {
		t.Fatalf("GetAsVar failed: %v", err)
	}
	// Verify we got the right number of results
	if len(usersAny) != 3 {
		t.Errorf("Expected 3 users, got %d", len(usersAny))
	}
}

// TestAllAsVar tests the AllAsVar sugar method (alias for GetAsVar).
func TestAllAsVar(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, email TEXT, age INTEGER)")
	execSQL(t, w, "INSERT INTO users VALUES (1, 'Alice', 'alice@example.com', 20)")
	execSQL(t, w, "INSERT INTO users VALUES (2, 'Bob', 'bob@example.com', 25)")
	execSQL(t, w, "INSERT INTO users VALUES (3, 'Charlie', 'charlie@example.com', 30)")
	execSQL(t, w, "INSERT INTO users VALUES (4, 'Diana', 'diana@example.com', 35)")
	execSQL(t, w, "INSERT INTO users VALUES (5, 'Eve', 'eve@example.com', 40)")

	w.SetTable("users")

	// Test AllAsVar
	usersAny, err := w.Q.AllAsVar()
	if err != nil {
		t.Fatalf("AllAsVar failed: %v", err)
	}
	// Verify we got the right number of results
	if len(usersAny) != 5 {
		t.Errorf("Expected 5 users, got %d", len(usersAny))
	}
}

// TestFindAllAsVar tests the FindAllAsVar sugar method (alias for GetAsVar).
func TestFindAllAsVar(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, email TEXT, age INTEGER)")
	execSQL(t, w, "INSERT INTO users VALUES (1, 'Alice', 'alice@example.com', 20)")
	execSQL(t, w, "INSERT INTO users VALUES (2, 'Bob', 'bob@example.com', 25)")
	execSQL(t, w, "INSERT INTO users VALUES (3, 'Charlie', 'charlie@example.com', 30)")
	execSQL(t, w, "INSERT INTO users VALUES (4, 'Diana', 'diana@example.com', 35)")
	execSQL(t, w, "INSERT INTO users VALUES (5, 'Eve', 'eve@example.com', 40)")

	w.SetTable("users")

	// Test FindAllAsVar
	usersAny, err := w.Q.FindAllAsVar()
	if err != nil {
		t.Fatalf("FindAllAsVar failed: %v", err)
	}
	// Verify we got the right number of results
	if len(usersAny) != 5 {
		t.Errorf("Expected 5 users, got %d", len(usersAny))
	}
}

// TestFindAsVar tests the FindAsVar sugar method.
func TestFindAsVar(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, email TEXT, age INTEGER)")
	execSQL(t, w, "INSERT INTO users VALUES (1, 'Alice', 'alice@example.com', 20)")
	execSQL(t, w, "INSERT INTO users VALUES (2, 'Bob', 'bob@example.com', 25)")
	execSQL(t, w, "INSERT INTO users VALUES (3, 'Charlie', 'charlie@example.com', 30)")
	execSQL(t, w, "INSERT INTO users VALUES (4, 'Diana', 'diana@example.com', 35)")
	execSQL(t, w, "INSERT INTO users VALUES (5, 'Eve', 'eve@example.com', 40)")

	w.SetTable("users")

	// Test FindAsVar with conditions
	usersAny, err := w.Q.Where("age > ?", 25).FindAsVar()
	if err != nil {
		t.Fatalf("FindAsVar failed: %v", err)
	}
	// Verify we got the right number of results
	if len(usersAny) != 3 {
		t.Errorf("Expected 3 users, got %d", len(usersAny))
	}

	// Test FindAsVar without conditions
	w = openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE users2 (id INTEGER PRIMARY KEY, name TEXT, email TEXT, age INTEGER)")
	execSQL(t, w, "INSERT INTO users2 VALUES (1, 'Alice', 'alice@example.com', 20)")
	execSQL(t, w, "INSERT INTO users2 VALUES (2, 'Bob', 'bob@example.com', 25)")
	w.SetTable("users2")
	usersAny, err = w.Q.FindAsVar()
	if err != nil {
		t.Fatalf("FindAsVar without conditions failed: %v", err)
	}
	// Verify we got the right number of results
	if len(usersAny) != 2 {
		t.Errorf("Expected 2 users, got %d", len(usersAny))
	}
}

// TestSugarMethodErrorPropagation tests that errors are properly propagated from base methods.
func TestSugarMethodErrorPropagation(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, email TEXT, age INTEGER)")
	w.SetTable("users")

	// Test with invalid column
	_, err := w.Q.SumAsVar("invalid_column")
	if err == nil {
		t.Error("Expected error for invalid column")
	}
}

// TestSugarMethodEmptyResults tests sugar methods with empty result sets.
func TestSugarMethodEmptyResults(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, email TEXT, age INTEGER)")
	w.SetTable("users")

	// Test CountAsVar with no results
	count, err := w.Q.Where("age > ?", 100).CountAsVar()
	if err != nil {
		t.Fatalf("CountAsVar failed: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}

	// Test GetAsVar with no results
	usersAny, err := w.Q.Where("age > ?", 100).GetAsVar()
	if err != nil {
		t.Fatalf("GetAsVar failed: %v", err)
	}
	// Verify we got empty results
	if len(usersAny) != 0 {
		t.Errorf("Expected empty slice, got %d users", len(usersAny))
	}

	// Test PluckAsVar with no results
	emailsAny, err := w.Q.Where("age > ?", 100).PluckAsVar("email")
	if err != nil {
		t.Fatalf("PluckAsVar failed: %v", err)
	}
	// Verify we got empty results
	if len(emailsAny) != 0 {
		t.Errorf("Expected empty slice, got %d emails", len(emailsAny))
	}
}
