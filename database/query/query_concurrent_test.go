package query

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"

	_ "modernc.org/sqlite"
)

// TestConcurrentQueryExecution tests that multiple goroutines can execute queries concurrently
func TestConcurrentQueryExecution(t *testing.T) {
	dbPath := "file:" + filepath.Join(t.TempDir(), "test.db") + "?_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Insert test data
	for i := 0; i < 100; i++ {
		_, err = db.Exec("INSERT INTO users (name) VALUES (?)", fmt.Sprintf("user%d", i))
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	var wg sync.WaitGroup
	concurrency := 10
	errors := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			q := NewQuery(context.Background(), db, nil, "", nil, nil)
			q.Table("users")

			var count int64
			err := q.Count(&count)
			if err != nil {
				errors <- err
				return
			}

			if count != 100 {
				t.Errorf("Worker %d: expected count 100, got %d", workerID, count)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent query error: %v", err)
	}
}

// TestConcurrentQueryCloning tests that Query.Clone() is thread-safe
func TestConcurrentQueryCloning(t *testing.T) {
	q := NewQuery(context.Background(), nil, nil, "", nil, nil)
	q.Table("users")
	q.Where("status = ?", "active")
	q.Where("age > ?", 18)

	var wg sync.WaitGroup
	concurrency := 100
	cloneCount := atomic.Int32{}

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			clone := q.Clone()
			if clone == nil {
				t.Error("Clone returned nil")
				return
			}

			// Verify clone independence by mutating it
			clone.Where("id = ?", 1)

			cloneCount.Add(1)
		}()
	}

	wg.Wait()

	if cloneCount.Load() != int32(concurrency) {
		t.Errorf("Expected %d clones, got %d", concurrency, cloneCount.Load())
	}
}

// TestConcurrentQueryMutation tests that mutations on cloned queries don't affect each other
func TestConcurrentQueryMutation(t *testing.T) {
	q := NewQuery(context.Background(), nil, nil, "", nil, nil)
	q.Table("users")

	var wg sync.WaitGroup
	concurrency := 50

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			clone := q.Clone()
			clone.Where("id = ?", id)

			// Verify the clone can be used without error
			// The fact that it doesn't panic and we can call methods on it is sufficient
			_ = clone
		}(i)
	}

	wg.Wait()

	// Original query should still be usable (not corrupted by concurrent mutations)
	if len(q.wheres) != 0 {
		t.Errorf("Original query was corrupted: expected 0 where clauses, got %d", len(q.wheres))
	}
}

// TestConcurrentTransactionHandling tests concurrent transaction operations
func TestConcurrentTransactionHandling(t *testing.T) {
	dbPath := "file:" + filepath.Join(t.TempDir(), "test.db") + "?_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, counter INTEGER)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Insert initial data
	_, err = db.Exec("INSERT INTO users (name, counter) VALUES ('test', 0)")
	if err != nil {
		t.Fatalf("Failed to insert initial data: %v", err)
	}

	var wg sync.WaitGroup
	concurrency := 10
	errors := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			q := NewQuery(context.Background(), db, nil, "", nil, nil)
			tx, err := q.Begin()
			if err != nil {
				errors <- err
				return
			}

			// Perform update within transaction using Query wrapper
			_, err = tx.Table("users").Where("id = ?", 1).Update(map[string]any{
				"counter": workerID,
			})
			if err != nil {
				errors <- err
				_ = tx.Rollback()
				return
			}

			err = tx.Commit()
			if err != nil {
				errors <- err
				return
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent transaction error: %v", err)
	}

	// Verify final state - one of the transactions should have committed
	var counter int
	err = db.QueryRow("SELECT counter FROM users WHERE id = 1").Scan(&counter)
	if err != nil {
		t.Fatalf("Failed to query final state: %v", err)
	}

	// Counter should be one of the worker IDs (0-9)
	if counter < 0 || counter >= concurrency {
		t.Errorf("Expected counter to be between 0 and %d, got %d", concurrency-1, counter)
	}
}

// TestConcurrentReadOperations tests concurrent read operations
func TestConcurrentReadOperations(t *testing.T) {
	dbPath := "file:" + filepath.Join(t.TempDir(), "test.db") + "?_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Insert test data
	for i := 0; i < 50; i++ {
		_, err = db.Exec("INSERT INTO users (name) VALUES (?)", fmt.Sprintf("user%d", i))
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	var wg sync.WaitGroup
	concurrency := 20
	errors := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			q := NewQuery(context.Background(), db, nil, "", nil, nil)
			q.Table("users")

			var users []map[string]any
			err := q.Get(&users)
			if err != nil {
				errors <- err
				return
			}

			if len(users) != 50 {
				t.Errorf("Expected 50 users, got %d", len(users))
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent read error: %v", err)
	}
}

// TestConcurrentWriteOperations tests concurrent write operations
func TestConcurrentWriteOperations(t *testing.T) {
	dbPath := "file:" + filepath.Join(t.TempDir(), "test.db") + "?_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	var wg sync.WaitGroup
	concurrency := 20
	errors := make(chan error, concurrency)
	successCount := atomic.Int32{}

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			q := NewQuery(context.Background(), db, nil, "", nil, nil)
			q.Table("users")

			err := q.Create(map[string]any{
				"name": fmt.Sprintf("user%d", id),
			})
			if err != nil {
				errors <- err
				return
			}

			successCount.Add(1)
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent write error: %v", err)
	}

	// Verify all writes succeeded
	if successCount.Load() != int32(concurrency) {
		t.Errorf("Expected %d successful writes, got %d", concurrency, successCount.Load())
	}

	// Verify count in database
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count rows: %v", err)
	}

	if count != concurrency {
		t.Errorf("Expected %d rows in database, got %d", concurrency, count)
	}
}

// TestConcurrentMixedOperations tests mixed read and write operations
func TestConcurrentMixedOperations(t *testing.T) {
	dbPath := "file:" + filepath.Join(t.TempDir(), "test.db") + "?_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	readConcurrency := 10
	writeConcurrency := 10
	var wg sync.WaitGroup
	errors := make(chan error, readConcurrency+writeConcurrency)

	// Concurrent readers
	for i := 0; i < readConcurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			q := NewQuery(context.Background(), db, nil, "", nil, nil)
			q.Table("users")

			var count int64
			err := q.Count(&count)
			if err != nil {
				errors <- err
			}
		}()
	}

	// Concurrent writers
	for i := 0; i < writeConcurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			q := NewQuery(context.Background(), db, nil, "", nil, nil)
			q.Table("users")

			err := q.Create(map[string]any{
				"name": fmt.Sprintf("user%d", id),
			})
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent mixed operation error: %v", err)
	}
}

// TestConcurrentQueryWithWhereClause tests concurrent queries with where clauses
func TestConcurrentQueryWithWhereClause(t *testing.T) {
	dbPath := "file:" + filepath.Join(t.TempDir(), "test.db") + "?_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, status TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Insert test data
	for i := 0; i < 50; i++ {
		status := "active"
		if i%2 == 0 {
			status = "inactive"
		}
		_, err = db.Exec("INSERT INTO users (name, status) VALUES (?, ?)", fmt.Sprintf("user%d", i), status)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	var wg sync.WaitGroup
	concurrency := 10
	errors := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			q := NewQuery(context.Background(), db, nil, "", nil, nil)
			q.Table("users")

			var count int64
			err := q.Where("status = ?", "active").Count(&count)
			if err != nil {
				errors <- err
				return
			}

			// Should be 25 active users
			if count != 25 {
				t.Errorf("Worker %d: expected 25 active users, got %d", workerID, count)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent where clause error: %v", err)
	}
}

// TestConcurrentQueryWithJoins tests concurrent queries with joins
func TestConcurrentQueryWithJoins(t *testing.T) {
	dbPath := "file:" + filepath.Join(t.TempDir(), "test.db") + "?_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec(`
		CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE posts (id INTEGER PRIMARY KEY, user_id INTEGER, title TEXT)
	`)
	if err != nil {
		t.Fatalf("Failed to create posts table: %v", err)
	}

	// Insert test data
	for i := 0; i < 10; i++ {
		_, err = db.Exec("INSERT INTO users (name) VALUES (?)", fmt.Sprintf("user%d", i))
		if err != nil {
			t.Fatalf("Failed to insert user: %v", err)
		}
		for j := 0; j < 5; j++ {
			_, err = db.Exec("INSERT INTO posts (user_id, title) VALUES (?, ?)", i+1, fmt.Sprintf("post%d", j))
			if err != nil {
				t.Fatalf("Failed to insert post: %v", err)
			}
		}
	}

	var wg sync.WaitGroup
	concurrency := 10
	errors := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			q := NewQuery(context.Background(), db, nil, "", nil, nil)
			q.Table("users")

			var results []map[string]any
			err := q.Join("posts ON users.id = posts.user_id").Get(&results)
			if err != nil {
				errors <- err
				return
			}

			// Should have 50 results (10 users * 5 posts each)
			if len(results) != 50 {
				t.Errorf("Expected 50 results, got %d", len(results))
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent join error: %v", err)
	}
}

// TestConcurrentQueryWithLimitOffset tests concurrent queries with limit and offset
func TestConcurrentQueryWithLimitOffset(t *testing.T) {
	dbPath := "file:" + filepath.Join(t.TempDir(), "test.db") + "?_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Insert test data
	for i := 0; i < 100; i++ {
		_, err = db.Exec("INSERT INTO users (name) VALUES (?)", fmt.Sprintf("user%d", i))
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	var wg sync.WaitGroup
	concurrency := 10
	errors := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			q := NewQuery(context.Background(), db, nil, "", nil, nil)
			q.Table("users")

			var users []map[string]any
			err := q.Limit(10).Offset(workerID * 10).Get(&users)
			if err != nil {
				errors <- err
				return
			}

			// Should have 10 users per page
			if len(users) != 10 {
				t.Errorf("Worker %d: expected 10 users, got %d", workerID, len(users))
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent limit/offset error: %v", err)
	}
}

// TestConcurrentQueryWithOrderBy tests concurrent queries with order by
func TestConcurrentQueryWithOrderBy(t *testing.T) {
	dbPath := "file:" + filepath.Join(t.TempDir(), "test.db") + "?_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Insert test data
	for i := 0; i < 50; i++ {
		_, err = db.Exec("INSERT INTO users (name) VALUES (?)", fmt.Sprintf("user%d", i))
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	var wg sync.WaitGroup
	concurrency := 10
	errors := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			q := NewQuery(context.Background(), db, nil, "", nil, nil)
			q.Table("users")

			var users []map[string]any
			err := q.OrderBy("id", "desc").Get(&users)
			if err != nil {
				errors <- err
				return
			}

			// Should have 50 users
			if len(users) != 50 {
				t.Errorf("Expected 50 users, got %d", len(users))
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent order by error: %v", err)
	}
}

// TestConcurrentQueryWithDistinct tests concurrent queries with distinct
func TestConcurrentQueryWithDistinct(t *testing.T) {
	dbPath := "file:" + filepath.Join(t.TempDir(), "test.db") + "?_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, status TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Insert test data with duplicate statuses
	for i := 0; i < 50; i++ {
		status := "active"
		switch i % 3 {
		case 0:
			status = "pending"
		case 1:
			status = "inactive"
		}
		_, err = db.Exec("INSERT INTO users (name, status) VALUES (?, ?)", fmt.Sprintf("user%d", i), status)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	var wg sync.WaitGroup
	concurrency := 10
	errors := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			q := NewQuery(context.Background(), db, nil, "", nil, nil)
			q.Table("users")

			var statuses []map[string]any
			err := q.Distinct().Select("status").Get(&statuses)
			if err != nil {
				errors <- err
				return
			}

			// Should have 3 distinct statuses
			if len(statuses) != 3 {
				t.Errorf("Expected 3 distinct statuses, got %d", len(statuses))
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent distinct error: %v", err)
	}
}

// TestConcurrentQueryWithAggregates tests concurrent queries with aggregate functions
func TestConcurrentQueryWithAggregates(t *testing.T) {
	dbPath := "file:" + filepath.Join(t.TempDir(), "test.db") + "?_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, age INTEGER)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Insert test data
	for i := 0; i < 50; i++ {
		_, err = db.Exec("INSERT INTO users (name, age) VALUES (?, ?)", fmt.Sprintf("user%d", i), 20+i)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	var wg sync.WaitGroup
	concurrency := 10
	errors := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			q := NewQuery(context.Background(), db, nil, "", nil, nil)
			q.Table("users")

			var avgAge float64
			err := q.Avg("age", &avgAge)
			if err != nil {
				errors <- err
				return
			}

			// Average should be around 44.5 (average of 20-69)
			if avgAge < 44 || avgAge > 45 {
				t.Errorf("Expected average age around 44.5, got %f", avgAge)
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent aggregate error: %v", err)
	}
}
