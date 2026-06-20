package query

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	_ "modernc.org/sqlite"
)

// TestConcurrentQueryContextCancellation tests that context cancellation works correctly with concurrent queries
func TestConcurrentQueryContextCancellation(t *testing.T) {
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
	concurrency := 5
	cancelledCount := atomic.Int32{}

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			ctx, cancel := context.WithCancel(context.Background())

			// Cancel context after a short delay for some workers
			if workerID%2 == 0 {
				time.Sleep(10 * time.Millisecond)
				cancel()
				cancelledCount.Add(1)
				defer cancel()
			} else {
				defer cancel()
			}

			q := NewQuery(ctx, db, nil, "", nil, nil)
			q.Table("users")

			var users []map[string]any
			err := q.Get(&users)

			// Cancelled queries should fail
			if workerID%2 == 0 {
				if err == nil {
					t.Errorf("Worker %d: expected error for cancelled context, got nil", workerID)
				}
			} else {
				if err != nil {
					t.Errorf("Worker %d: unexpected error: %v", workerID, err)
				}
			}
		}(i)
	}

	wg.Wait()

	if cancelledCount.Load() != int32((concurrency+1)/2) {
		t.Errorf("Expected %d cancelled queries, got %d", (concurrency+1)/2, cancelledCount.Load())
	}
}

// TestConcurrentQueryWithTimeout tests that query timeouts work correctly under concurrency
func TestConcurrentQueryWithTimeout(t *testing.T) {
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
	for i := 0; i < 10; i++ {
		_, err = db.Exec("INSERT INTO users (name) VALUES (?)", fmt.Sprintf("user%d", i))
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	var wg sync.WaitGroup
	concurrency := 5

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			q := NewQuery(ctx, db, nil, "", nil, nil)
			q.Table("users")

			var count int64
			err := q.Count(&count)
			if err != nil {
				t.Errorf("Query with timeout failed: %v", err)
			}

			if count != 10 {
				t.Errorf("Expected count 10, got %d", count)
			}
		}()
	}

	wg.Wait()
}

// TestConcurrentQueryWithReplicas tests concurrent queries with read replicas
func TestConcurrentQueryWithReplicas(t *testing.T) {
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

			// Use NewQueryWithReplicas to test replica functionality
			q := NewQueryWithReplicas(context.Background(), db, db, nil, "", nil, nil)
			q.Table("users")

			var count int64
			err := q.Count(&count)
			if err != nil {
				errors <- err
				return
			}

			if count != 50 {
				t.Errorf("Expected count 50, got %d", count)
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent replica query error: %v", err)
	}
}

// TestConcurrentQueryWithObservers tests concurrent queries with observers
func TestConcurrentQueryWithObservers(t *testing.T) {
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
	concurrency := 10
	errors := make(chan error, concurrency)
	observer := &testObserver{}

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			q := NewQuery(context.Background(), db, nil, "", nil, nil)
			q.Table("users")

			// Register a simple observer with a model type
			type TestUser struct {
				ID   int
				Name string
			}
			q.Observe(&TestUser{}, observer)

			// Create using the struct type to trigger observer
			user := TestUser{Name: fmt.Sprintf("user%d", workerID)}
			err := q.Create(&user)
			if err != nil {
				errors <- err
				return
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent observer query error: %v", err)
	}

	// Verify observers were called
	if observer.callCount.Load() != int32(concurrency) {
		t.Errorf("Expected %d observer calls, got %d", concurrency, observer.callCount.Load())
	}
}

// TestConcurrentQueryWithScopes tests concurrent queries with scopes
func TestConcurrentQueryWithScopes(t *testing.T) {
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

	// Verify data insertion
	var activeCount int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE status = ?", "active").Scan(&activeCount)
	if err != nil {
		t.Fatalf("Failed to verify test data: %v", err)
	}
	if activeCount != 25 {
		t.Fatalf("Expected 25 active users in database, got %d", activeCount)
	}

	var wg sync.WaitGroup
	concurrency := 10
	errors := make(chan error, concurrency)

	activeScope := func(q contractsorm.Query) contractsorm.Query {
		return q.Where("status = ?", "active")
	}

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			q := NewQuery(context.Background(), db, nil, "", nil, nil)
			q.Table("users")

			var users []map[string]any
			err := q.Scopes(activeScope).Get(&users)
			if err != nil {
				errors <- err
				return
			}

			// Should be 25 active users
			if len(users) != 25 {
				t.Errorf("Expected 25 active users, got %d", len(users))
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent scope query error: %v", err)
	}
}

// TestConcurrentQueryWithSoftDelete tests concurrent queries with soft delete
func TestConcurrentQueryWithSoftDelete(t *testing.T) {
	dbPath := "file:" + filepath.Join(t.TempDir(), "test.db") + "?_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, deleted_at TEXT)")
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
		go func(workerID int) {
			defer wg.Done()

			q := NewQuery(context.Background(), db, nil, "", nil, nil)
			q.Table("users")

			// Soft delete some records
			if workerID%2 == 0 {
				_, err := q.Where("id = ?", workerID+1).Delete()
				if err != nil {
					errors <- err
					return
				}
			} else {
				// Count non-deleted records
				var count int64
				err := q.Count(&count)
				if err != nil {
					errors <- err
					return
				}

				// Count should be >= 25 (at least half the records)
				if count < 25 {
					t.Errorf("Expected at least 25 non-deleted records, got %d", count)
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent soft delete error: %v", err)
	}
}

// TestConcurrentQueryWithLocks tests concurrent queries with row locks
func TestConcurrentQueryWithLocks(t *testing.T) {
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
	concurrency := 5
	errors := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			q := NewQuery(context.Background(), db, nil, "", nil, nil)
			q.Table("users")

			// Note: SQLite doesn't support FOR UPDATE, but we test the method chain
			q.LockForUpdate()

			var users []map[string]any
			err := q.Where("id = ?", 1).Get(&users)
			if err != nil {
				errors <- err
				return
			}

			if len(users) != 1 {
				t.Errorf("Expected 1 user, got %d", len(users))
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent lock query error: %v", err)
	}
}

// TestConcurrentQueryWithPagination tests concurrent pagination queries
func TestConcurrentQueryWithPagination(t *testing.T) {
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
		go func(page int) {
			defer wg.Done()

			q := NewQuery(context.Background(), db, nil, "", nil, nil)
			q.Table("users")

			var users []map[string]any
			var total int64

			err := q.Paginate(page, 10, &users, &total)
			if err != nil {
				errors <- err
				return
			}

			if total != 100 {
				t.Errorf("Expected total 100, got %d", total)
			}

			// Each page should have 10 users except possibly the last
			expectedUsers := 10
			if page > 10 {
				expectedUsers = 0 // Pages beyond 10 should be empty
			}
			if len(users) != expectedUsers {
				t.Errorf("Page %d: expected %d users, got %d", page, expectedUsers, len(users))
			}
		}(i + 1)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent pagination error: %v", err)
	}
}

// TestConcurrentQueryWithChunking tests concurrent chunk processing
func TestConcurrentQueryWithChunking(t *testing.T) {
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
	concurrency := 5
	errors := make(chan error, concurrency)
	chunkCount := atomic.Int32{}

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			q := NewQuery(context.Background(), db, nil, "", nil, nil)
			q.Table("users")

			err := q.Chunk(10, func(users []map[string]any) error {
				chunkCount.Add(1)
				return nil
			})
			if err != nil {
				errors <- err
				return
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent chunking error: %v", err)
	}

	// Each worker should process 10 chunks (100 users / 10 per chunk)
	expectedChunks := int32(concurrency * 10)
	if chunkCount.Load() != expectedChunks {
		t.Errorf("Expected %d chunks, got %d", expectedChunks, chunkCount.Load())
	}
}

// TestConcurrentQueryWithCursor tests concurrent cursor streaming
func TestConcurrentQueryWithCursor(t *testing.T) {
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
	concurrency := 5
	errors := make(chan error, concurrency)
	rowCount := atomic.Int32{}

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			q := NewQuery(context.Background(), db, nil, "", nil, nil)
			q.Table("users")

			cursor, err := q.Cursor()
			if err != nil {
				errors <- err
				return
			}

			for range cursor {
				rowCount.Add(1)
			}
			// Cursor channel closes automatically when done
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent cursor error: %v", err)
	}

	// Each worker should read 50 rows
	expectedRows := int32(concurrency * 50)
	if rowCount.Load() != expectedRows {
		t.Errorf("Expected %d rows, got %d", expectedRows, rowCount.Load())
	}
}

// testObserver is a simple test observer implementation
type testObserver struct {
	callCount atomic.Int32
}

func (o *testObserver) Created(event contractsorm.Event) error {
	o.callCount.Add(1)
	return nil
}

func (o *testObserver) Updated(event contractsorm.Event) error {
	return nil
}

func (o *testObserver) Deleted(event contractsorm.Event) error {
	return nil
}

func (o *testObserver) ForceDeleted(event contractsorm.Event) error {
	return nil
}
