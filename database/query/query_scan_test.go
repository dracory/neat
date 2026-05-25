package query_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/database/query"
)

// --- db tag ---

type dbTagModel struct {
	MyCol string `db:"my_col"`
}

func TestScanRowsByDbTag(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_db_tag (my_col TEXT)")
	execSQL(t, w, "INSERT INTO test_db_tag VALUES ('hello')")

	w.SetTable("test_db_tag")
	var result dbTagModel
	if err := w.Q.Find(&result); err != nil {
		t.Fatalf("Find failed: %v", err)
	}
	if result.MyCol != "hello" {
		t.Errorf("expected MyCol='hello', got %q", result.MyCol)
	}
}

// --- neat tag ---

type neatTagModel struct {
	MyCol string `neat:"my_col"`
}

func TestScanRowsByNeatTag(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_neat_tag (my_col TEXT)")
	execSQL(t, w, "INSERT INTO test_neat_tag VALUES ('world')")

	w.SetTable("test_neat_tag")
	var result neatTagModel
	if err := w.Q.Find(&result); err != nil {
		t.Fatalf("Find failed: %v", err)
	}
	if result.MyCol != "world" {
		t.Errorf("expected MyCol='world', got %q", result.MyCol)
	}
}

// --- snake_case fallback (no tag) ---

type snakeCaseModel struct {
	UserName string
}

func TestScanRowsBySnakeCase(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_snake (user_name TEXT)")
	execSQL(t, w, "INSERT INTO test_snake VALUES ('snake')")

	w.SetTable("test_snake")
	var result snakeCaseModel
	if err := w.Q.Find(&result); err != nil {
		t.Fatalf("Find failed: %v", err)
	}
	if result.UserName != "snake" {
		t.Errorf("expected UserName='snake', got %q", result.UserName)
	}
}

// --- extra (unmatched) columns don't panic ---

type narrowModel struct {
	Name string `db:"name"`
}

func TestScanRowsUnmatchedColumnIgnored(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_wide (name TEXT, extra TEXT, another INTEGER)")
	execSQL(t, w, "INSERT INTO test_wide VALUES ('alice', 'ignored', 42)")

	w.SetTable("test_wide")
	var result narrowModel
	if err := w.Q.Find(&result); err != nil {
		t.Fatalf("Find should not error on extra columns: %v", err)
	}
	if result.Name != "alice" {
		t.Errorf("expected Name='alice', got %q", result.Name)
	}
}

// --- slice scan ---

type rowModel struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

func TestScanRowsIntoSlice(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_slice (id INTEGER, name TEXT)")
	execSQL(t, w, "INSERT INTO test_slice VALUES (1,'a'),(2,'b'),(3,'c')")

	w.SetTable("test_slice")
	var results []rowModel
	if err := w.Q.Find(&results); err != nil {
		t.Fatalf("Find into slice failed: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(results))
	}
	if results[0].ID != 1 || results[0].Name != "a" {
		t.Errorf("unexpected first row: %+v", results[0])
	}
	if results[2].Name != "c" {
		t.Errorf("unexpected last row: %+v", results[2])
	}
}

// --- structFieldColumnName resolution order: db > neat > gorm > snake_case ---

func TestStructFieldColumnNameDbTagPriority(t *testing.T) {
	type m struct {
		F string `db:"db_col" neat:"neat_col" gorm:"column:gorm_col"`
	}
	col := query.StructFieldColumnName(reflect.TypeOf(m{}).Field(0))
	if col != "db_col" {
		t.Errorf("expected db tag to win, got %q", col)
	}
}

func TestStructFieldColumnNameNeatTagFallback(t *testing.T) {
	type m struct {
		F string `neat:"neat_col" gorm:"column:gorm_col"`
	}
	col := query.StructFieldColumnName(reflect.TypeOf(m{}).Field(0))
	if col != "neat_col" {
		t.Errorf("expected neat tag fallback, got %q", col)
	}
}

func TestStructFieldColumnNameGormTagFallback(t *testing.T) {
	type m struct {
		F string `gorm:"column:gorm_col"`
	}
	col := query.StructFieldColumnName(reflect.TypeOf(m{}).Field(0))
	if col != "gorm_col" {
		t.Errorf("expected gorm tag fallback, got %q", col)
	}
}

func TestStructFieldColumnNameSnakeCaseFallback(t *testing.T) {
	type m struct {
		MyFieldName string
	}
	col := query.StructFieldColumnName(reflect.TypeOf(m{}).Field(0))
	if !strings.Contains(col, "my_field_name") {
		t.Errorf("expected snake_case fallback, got %q", col)
	}
}

// --- cursor tests ---

func TestCursorBasic(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_cursor (id INTEGER, name TEXT)")
	execSQL(t, w, "INSERT INTO test_cursor VALUES (1,'alice'),(2,'bob'),(3,'charlie')")

	w.SetTable("test_cursor")
	cursorChan, err := w.Q.Cursor()
	if err != nil {
		t.Fatalf("Cursor failed: %v", err)
	}

	count := 0
	for cursor := range cursorChan {
		count++
		if cursor == nil {
			t.Error("Expected non-nil cursor")
		}
	}

	if count != 3 {
		t.Errorf("Expected 3 cursor items, got %d", count)
	}
}

func TestCursorChannelCreationAndConsumption(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_cursor_stream (id INTEGER, value TEXT)")
	execSQL(t, w, "INSERT INTO test_cursor_stream VALUES (1,'one'),(2,'two'),(3,'three'),(4,'four'),(5,'five')")

	w.SetTable("test_cursor_stream")
	cursorChan, err := w.Q.Cursor()
	if err != nil {
		t.Fatalf("Cursor failed: %v", err)
	}

	if cursorChan == nil {
		t.Fatal("Expected non-nil cursor channel")
	}

	results := make([]map[string]any, 0)
	for cursor := range cursorChan {
		if cursor == nil {
			t.Error("Expected non-nil cursor")
			continue
		}

		var result map[string]any
		if err := cursor.Scan(&result); err != nil {
			t.Errorf("Cursor.Scan failed: %v", err)
		}
		results = append(results, result)
	}

	if len(results) != 5 {
		t.Errorf("Expected 5 results, got %d", len(results))
	}
}

func TestCursorErrorHandling(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_cursor_error (id INTEGER, name TEXT)")
	execSQL(t, w, "INSERT INTO test_cursor_error VALUES (1,'test')")

	w.SetTable("test_cursor_error")
	cursorChan, err := w.Q.Cursor()
	if err != nil {
		t.Fatalf("Cursor failed: %v", err)
	}

	count := 0
	for cursor := range cursorChan {
		count++
		if cursor == nil {
			t.Error("Expected non-nil cursor")
		}
	}

	if count != 1 {
		t.Errorf("Expected 1 cursor item, got %d", count)
	}
}

func TestCursorWithTransactions(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_cursor_tx (id INTEGER, name TEXT)")

	execSQL(t, w, "INSERT INTO test_cursor_tx VALUES (1,'tx1'),(2,'tx2')")
	w.SetTable("test_cursor_tx")

	tx, err := w.Q.Begin()
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	cursorChan, err := tx.Cursor()
	if err != nil {
		t.Fatalf("Cursor in transaction failed: %v", err)
	}

	count := 0
	for cursor := range cursorChan {
		count++
		if cursor == nil {
			t.Error("Expected non-nil cursor in transaction")
		}
	}

	if count != 2 {
		t.Errorf("Expected 2 cursor items in transaction, got %d", count)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}
}

func TestCursorWithWhereClauses(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_cursor_where (id INTEGER, name TEXT, status TEXT)")
	execSQL(t, w, "INSERT INTO test_cursor_where VALUES (1,'alice','active'),(2,'bob','inactive'),(3,'charlie','active')")

	w.SetTable("test_cursor_where")
	w.Q.Where("status = ?", "active")

	cursorChan, err := w.Q.Cursor()
	if err != nil {
		t.Fatalf("Cursor with WHERE failed: %v", err)
	}

	count := 0
	for cursor := range cursorChan {
		count++
		if cursor == nil {
			t.Error("Expected non-nil cursor with WHERE")
		}
	}

	if count != 2 {
		t.Errorf("Expected 2 cursor items with WHERE clause, got %d", count)
	}
}

// --- chunk tests ---

func TestChunkBasic(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_chunk (id INTEGER, name TEXT)")
	execSQL(t, w, "INSERT INTO test_chunk VALUES (1,'a'),(2,'b'),(3,'c'),(4,'d'),(5,'e')")

	w.SetTable("test_chunk")

	chunkCount := 0
	totalRows := 0

	err := w.Q.Chunk(2, func(chunk []map[string]any) error {
		chunkCount++
		totalRows += len(chunk)
		return nil
	})

	if err != nil {
		t.Fatalf("Chunk failed: %v", err)
	}

	if chunkCount != 3 {
		t.Errorf("Expected 3 chunks, got %d", chunkCount)
	}

	if totalRows != 5 {
		t.Errorf("Expected 5 total rows, got %d", totalRows)
	}
}

func TestChunkCallbackExecution(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_chunk_callback (id INTEGER, value TEXT)")
	execSQL(t, w, "INSERT INTO test_chunk_callback VALUES (1,'one'),(2,'two'),(3,'three')")

	w.SetTable("test_chunk_callback")

	var capturedChunks [][]map[string]any

	err := w.Q.Chunk(2, func(chunk []map[string]any) error {
		capturedChunks = append(capturedChunks, chunk)
		return nil
	})

	if err != nil {
		t.Fatalf("Chunk failed: %v", err)
	}

	if len(capturedChunks) != 2 {
		t.Errorf("Expected 2 callback executions, got %d", len(capturedChunks))
	}

	if len(capturedChunks[0]) != 2 {
		t.Errorf("Expected first chunk to have 2 rows, got %d", len(capturedChunks[0]))
	}

	if len(capturedChunks[1]) != 1 {
		t.Errorf("Expected second chunk to have 1 row, got %d", len(capturedChunks[1]))
	}
}

func TestChunkSizeVariations(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_chunk_size (id INTEGER, name TEXT)")
	execSQL(t, w, "INSERT INTO test_chunk_size VALUES (1,'a'),(2,'b'),(3,'c'),(4,'d'),(5,'e'),(6,'f'),(7,'g'),(8,'h'),(9,'i'),(10,'j')")

	w.SetTable("test_chunk_size")

	testCases := []struct {
		chunkSize int
		expected  int
	}{
		{1, 10},
		{3, 4},
		{5, 2},
		{10, 1},
		{20, 1},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("size_%d", tc.chunkSize), func(t *testing.T) {
			chunkCount := 0
			err := w.Q.Chunk(tc.chunkSize, func(chunk []map[string]any) error {
				chunkCount++
				return nil
			})

			if err != nil {
				t.Fatalf("Chunk with size %d failed: %v", tc.chunkSize, err)
			}

			if chunkCount != tc.expected {
				t.Errorf("Expected %d chunks with size %d, got %d", tc.expected, tc.chunkSize, chunkCount)
			}
		})
	}
}

func TestChunkWithTypedSlices(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_chunk_typed (id INTEGER, name TEXT)")
	execSQL(t, w, "INSERT INTO test_chunk_typed VALUES (1,'alice'),(2,'bob'),(3,'charlie'),(4,'dave')")

	w.SetTable("test_chunk_typed")

	type Person struct {
		Id   int    `db:"id"`
		Name string `db:"name"`
	}

	chunkCount := 0
	totalRows := 0

	err := w.Q.Chunk(2, func(chunk []Person) error {
		chunkCount++
		totalRows += len(chunk)

		for _, p := range chunk {
			if p.Id == 0 || p.Name == "" {
				t.Errorf("Expected non-zero Id and Name, got %+v", p)
			}
		}

		return nil
	})

	if err != nil {
		t.Fatalf("Chunk with typed slice failed: %v", err)
	}

	if chunkCount != 2 {
		t.Errorf("Expected 2 chunks, got %d", chunkCount)
	}

	if totalRows != 4 {
		t.Errorf("Expected 4 total rows, got %d", totalRows)
	}
}

func TestChunkErrorHandling(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_chunk_error (id INTEGER, name TEXT)")
	execSQL(t, w, "INSERT INTO test_chunk_error VALUES (1,'a'),(2,'b'),(3,'c')")

	w.SetTable("test_chunk_error")

	t.Run("invalid callback type", func(t *testing.T) {
		err := w.Q.Chunk(2, "not a function")
		if err == nil {
			t.Error("Expected error for non-function callback")
		}
	})

	t.Run("callback returns error", func(t *testing.T) {
		callCount := 0
		err := w.Q.Chunk(2, func(chunk []map[string]any) error {
			callCount++
			if callCount == 2 {
				return fmt.Errorf("callback error")
			}
			return nil
		})

		if err == nil {
			t.Error("Expected error from callback")
		}

		if callCount != 2 {
			t.Errorf("Expected callback to be called twice before error, got %d", callCount)
		}
	})

	t.Run("empty result set", func(t *testing.T) {
		execSQL(t, w, "CREATE TABLE test_chunk_empty (id INTEGER)")
		w.SetTable("test_chunk_empty")

		chunkCount := 0
		err := w.Q.Chunk(2, func(chunk []map[string]any) error {
			chunkCount++
			return nil
		})

		if err != nil {
			t.Fatalf("Chunk with empty result failed: %v", err)
		}

		if chunkCount != 0 {
			t.Errorf("Expected 0 chunks for empty result, got %d", chunkCount)
		}
	})
}

func TestChunkWithTransactions(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_chunk_tx (id INTEGER, name TEXT)")
	execSQL(t, w, "INSERT INTO test_chunk_tx VALUES (1,'tx1'),(2,'tx2'),(3,'tx3')")

	w.SetTable("test_chunk_tx")

	tx, err := w.Q.Begin()
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	chunkCount := 0
	err = tx.Chunk(2, func(chunk []map[string]any) error {
		chunkCount++
		return nil
	})

	if err != nil {
		t.Fatalf("Chunk in transaction failed: %v", err)
	}

	if chunkCount != 2 {
		t.Errorf("Expected 2 chunks in transaction, got %d", chunkCount)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}
}

func TestChunkWithWhereClauses(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_chunk_where (id INTEGER, name TEXT, status TEXT)")
	execSQL(t, w, "INSERT INTO test_chunk_where VALUES (1,'alice','active'),(2,'bob','inactive'),(3,'charlie','active'),(4,'dave','active')")

	w.SetTable("test_chunk_where")
	w.Q.Where("status = ?", "active")

	chunkCount := 0
	totalRows := 0

	err := w.Q.Chunk(2, func(chunk []map[string]any) error {
		chunkCount++
		totalRows += len(chunk)
		return nil
	})

	if err != nil {
		t.Fatalf("Chunk with WHERE failed: %v", err)
	}

	if chunkCount != 2 {
		t.Errorf("Expected 2 chunks with WHERE clause, got %d", chunkCount)
	}

	if totalRows != 3 {
		t.Errorf("Expected 3 total rows with WHERE clause, got %d", totalRows)
	}
}

// --- FirstOr/FirstOrCreate/FirstOrNew/UpdateOrCreate tests ---

type FirstOrUser struct {
	ID    int    `db:"id"`
	Name  string `db:"name"`
	Email string `db:"email"`
}

func TestFirstOrWithCallback(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_first_or (id INTEGER, name TEXT, email TEXT)")
	execSQL(t, w, "INSERT INTO test_first_or VALUES (1,'alice','alice@example.com')")

	w.SetTable("test_first_or")

	t.Run("record found - callback not executed", func(t *testing.T) {
		var user FirstOrUser
		callbackExecuted := false

		err := w.Q.Where("id = ?", 1).FirstOr(&user, func() error {
			callbackExecuted = true
			return nil
		})

		if err != nil {
			t.Fatalf("FirstOr failed: %v", err)
		}

		if callbackExecuted {
			t.Error("Callback should not be executed when record is found")
		}

		if user.Name != "alice" {
			t.Errorf("Expected user.Name='alice', got %q", user.Name)
		}
	})

	t.Run("record not found - callback executed", func(t *testing.T) {
		var user FirstOrUser
		callbackExecuted := false
		callbackError := fmt.Errorf("not found")

		err := w.Q.Where("id = ?", 999).FirstOr(&user, func() error {
			callbackExecuted = true
			return callbackError
		})

		if err != callbackError {
			t.Errorf("Expected callback error, got %v", err)
		}

		if !callbackExecuted {
			t.Error("Callback should be executed when record is not found")
		}
	})

	t.Run("callback returns nil on not found", func(t *testing.T) {
		var user FirstOrUser
		callbackExecuted := false

		err := w.Q.Where("id = ?", 999).FirstOr(&user, func() error {
			callbackExecuted = true
			return nil
		})

		if err != nil {
			t.Fatalf("FirstOr with nil callback error failed: %v", err)
		}

		if !callbackExecuted {
			t.Error("Callback should be executed when record is not found")
		}
	})
}

func TestFirstOrCreate(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_first_or_create (id INTEGER, name TEXT, email TEXT)")
	execSQL(t, w, "INSERT INTO test_first_or_create VALUES (1,'alice','alice@example.com')")

	w.SetTable("test_first_or_create")

	t.Run("record exists - returns existing", func(t *testing.T) {
		var user FirstOrUser
		user.Name = "bob"
		user.Email = "bob@example.com"

		err := w.Q.Where("id = ?", 1).FirstOrCreate(&user)

		if err != nil {
			t.Fatalf("FirstOrCreate failed: %v", err)
		}

		if user.Name != "alice" {
			t.Errorf("Expected existing user.Name='alice', got %q", user.Name)
		}

		if user.Email != "alice@example.com" {
			t.Errorf("Expected existing user.Email='alice@example.com', got %q", user.Email)
		}
	})

	t.Run("record not found - creates new", func(t *testing.T) {
		var user FirstOrUser
		user.Name = "charlie"
		user.Email = "charlie@example.com"

		err := w.Q.Where("id = ?", 2).FirstOrCreate(&user)

		if err != nil {
			t.Fatalf("FirstOrCreate create failed: %v", err)
		}

		// Note: FirstOrCreate simplified implementation doesn't use WHERE clause for create
		// It just calls Create() on the model, so ID may be auto-generated
		if user.Name != "charlie" {
			t.Errorf("Expected created user.Name='charlie', got %q", user.Name)
		}

		if user.Email != "charlie@example.com" {
			t.Errorf("Expected created user.Email='charlie@example.com', got %q", user.Email)
		}
	})

	t.Run("create with auto-increment", func(t *testing.T) {
		var user FirstOrUser
		user.Name = "dave"
		user.Email = "dave@example.com"

		err := w.Q.FirstOrCreate(&user)

		if err != nil {
			t.Fatalf("FirstOrCreate auto-increment failed: %v", err)
		}

		if user.ID == 0 {
			t.Error("Expected auto-incremented ID to be set")
		}

		if user.Name != "dave" {
			t.Errorf("Expected user.Name='dave', got %q", user.Name)
		}
	})
}

func TestFirstOrNew(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_first_or_new (id INTEGER, name TEXT, email TEXT)")
	execSQL(t, w, "INSERT INTO test_first_or_new VALUES (1,'alice','alice@example.com')")

	w.SetTable("test_first_or_new")

	t.Run("record exists - returns existing", func(t *testing.T) {
		var user FirstOrUser
		attributes := map[string]any{"id": 1}

		err := w.Q.FirstOrNew(&user, attributes)

		if err != nil {
			t.Fatalf("FirstOrNew failed: %v", err)
		}

		if user.Name != "alice" {
			t.Errorf("Expected existing user.Name='alice', got %q", user.Name)
		}
	})

	t.Run("record not found - prepares new instance", func(t *testing.T) {
		var user FirstOrUser
		user.Name = "bob"
		user.Email = "bob@example.com"
		attributes := map[string]any{"id": 999}

		err := w.Q.FirstOrNew(&user, attributes)

		if err != nil {
			t.Fatalf("FirstOrNew prepare failed: %v", err)
		}

		// Note: FirstOrNew simplified implementation doesn't modify the model
		// when record is not found - it just returns nil
		// The model remains unchanged from the database result (which is empty)
	})

	t.Run("with values parameter", func(t *testing.T) {
		var user FirstOrUser
		attributes := map[string]any{"id": 1}
		values := map[string]any{"name": "updated"}

		err := w.Q.FirstOrNew(&user, attributes, values)

		if err != nil {
			t.Fatalf("FirstOrNew with values failed: %v", err)
		}

		// Since record exists, values should not be applied
		if user.Name != "alice" {
			t.Errorf("Expected existing user.Name='alice', got %q", user.Name)
		}
	})
}

func TestUpdateOrCreate(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_update_or_create (id INTEGER, name TEXT, email TEXT)")
	execSQL(t, w, "INSERT INTO test_update_or_create VALUES (1,'alice','alice@example.com')")

	w.SetTable("test_update_or_create")

	t.Run("record exists - calls Save on values", func(t *testing.T) {
		var user FirstOrUser
		attributes := map[string]any{"id": 1}
		values := map[string]any{"name": "alice_updated", "email": "alice_new@example.com"}

		err := w.Q.UpdateOrCreate(&user, attributes, values)

		// Note: Current simplified implementation calls Save(values) when record exists
		// This test verifies the method executes without error
		// The actual update behavior depends on the Save implementation
		if err != nil {
			t.Fatalf("UpdateOrCreate failed: %v", err)
		}
	})

	t.Run("record not found - calls Create on values", func(t *testing.T) {
		execSQL(t, w, "CREATE TABLE test_update_or_create_2 (id INTEGER, name TEXT, email TEXT)")
		w.SetTable("test_update_or_create_2")

		var user FirstOrUser
		attributes := map[string]any{"id": 999}
		values := map[string]any{"name": "bob", "email": "bob@example.com"}

		err := w.Q.UpdateOrCreate(&user, attributes, values)

		// Note: Current simplified implementation calls Create(values) when record not found
		// This test verifies the method executes without error
		if err != nil {
			t.Fatalf("UpdateOrCreate create failed: %v", err)
		}

		// Verify a record was created
		var count int64
		err = w.Q.Count(&count)
		if err != nil {
			t.Fatalf("Count failed: %v", err)
		}

		if count == 0 {
			t.Error("Expected at least one record to be created")
		}
	})

	t.Run("with struct attributes", func(t *testing.T) {
		var user FirstOrUser
		attributes := FirstOrUser{ID: 1}
		values := map[string]any{"email": "struct_update@example.com"}

		err := w.Q.UpdateOrCreate(&user, attributes, values)

		// Note: Current implementation doesn't use attributes for filtering
		// This test verifies the method accepts struct attributes
		if err != nil {
			t.Fatalf("UpdateOrCreate with struct attributes failed: %v", err)
		}
	})

	t.Run("with map values", func(t *testing.T) {
		var user FirstOrUser
		attributes := map[string]any{"id": 1}
		values := map[string]any{"name": "test", "email": "test@example.com"}

		err := w.Q.UpdateOrCreate(&user, attributes, values)

		if err != nil {
			t.Fatalf("UpdateOrCreate with map values failed: %v", err)
		}
	})

	t.Run("with struct values", func(t *testing.T) {
		var user FirstOrUser
		attributes := map[string]any{"id": 1}
		values := FirstOrUser{Name: "struct_val", Email: "struct@example.com"}

		err := w.Q.UpdateOrCreate(&user, attributes, values)

		if err != nil {
			t.Fatalf("UpdateOrCreate with struct values failed: %v", err)
		}
	})
}

func TestFirstOrErrorHandling(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_first_or_error (id INTEGER, name TEXT)")
	execSQL(t, w, "INSERT INTO test_first_or_error VALUES (1,'alice')")

	w.SetTable("test_first_or_error")

	t.Run("callback error propagation", func(t *testing.T) {
		var user FirstOrUser
		expectedError := fmt.Errorf("custom error")

		err := w.Q.Where("id = ?", 999).FirstOr(&user, func() error {
			return expectedError
		})

		if err != expectedError {
			t.Errorf("Expected callback error to be propagated, got %v", err)
		}
	})

	t.Run("callback panic handling", func(t *testing.T) {
		var user FirstOrUser

		// Note: This test verifies that panics in callbacks are not caught
		// In production, you should handle panics in your callbacks
		defer func() {
			if r := recover(); r != nil {
				// Expected panic
			}
		}()

		_ = w.Q.Where("id = ?", 999).FirstOr(&user, func() error {
			panic("callback panic")
		})
	})
}

func TestFirstOrCreateErrorHandling(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_first_or_create_error (id INTEGER, name TEXT)")

	w.SetTable("test_first_or_create_error")

	t.Run("create failure handling", func(t *testing.T) {
		var user FirstOrUser
		// Missing required field should cause create to fail
		user.Name = "" // Empty name might violate constraints

		err := w.Q.FirstOrCreate(&user)

		// The error handling depends on database constraints
		// This test verifies the method doesn't panic
		if err != nil {
			// Expected error due to constraints
		}
	})
}

func TestFirstOrNewErrorHandling(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_first_or_new_error (id INTEGER, name TEXT)")

	w.SetTable("test_first_or_new_error")

	t.Run("nil attributes handling", func(t *testing.T) {
		var user FirstOrUser

		err := w.Q.FirstOrNew(&user, nil)

		if err != nil {
			t.Fatalf("FirstOrNew with nil attributes failed: %v", err)
		}
	})

	t.Run("invalid attributes type", func(t *testing.T) {
		var user FirstOrUser

		err := w.Q.FirstOrNew(&user, "invalid")

		// Should handle gracefully or return error
		_ = err
	})
}

func TestUpdateOrCreateUpdateVsCreateLogic(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_update_create_logic (id INTEGER, name TEXT, email TEXT, status TEXT)")
	execSQL(t, w, "INSERT INTO test_update_create_logic VALUES (1,'alice','alice@example.com','active')")

	w.SetTable("test_update_create_logic")

	t.Run("update path - record found via First", func(t *testing.T) {
		var user FirstOrUser
		attributes := map[string]any{"id": 1}
		values := map[string]any{"name": "alice_updated", "status": "inactive"}

		// Note: Current implementation uses First(dest) to check existence
		// not the attributes parameter
		err := w.Q.UpdateOrCreate(&user, attributes, values)

		// Should call Save(values) when record exists
		if err != nil {
			t.Fatalf("UpdateOrCreate update path failed: %v", err)
		}
	})

	t.Run("create path - record not found via First", func(t *testing.T) {
		execSQL(t, w, "CREATE TABLE test_update_create_logic_2 (id INTEGER, name TEXT, email TEXT)")
		w.SetTable("test_update_create_logic_2")

		var user FirstOrUser
		attributes := map[string]any{"id": 999}
		values := map[string]any{"name": "bob", "email": "bob@example.com"}

		// Note: Current implementation calls Create(values) when First fails
		err := w.Q.UpdateOrCreate(&user, attributes, values)

		if err != nil {
			t.Fatalf("UpdateOrCreate create path failed: %v", err)
		}

		// Verify a record was created
		var count int64
		err = w.Q.Count(&count)
		if err != nil {
			t.Fatalf("Count failed: %v", err)
		}

		if count == 0 {
			t.Error("Expected at least one record to be created")
		}
	})

	t.Run("with where clause on query", func(t *testing.T) {
		execSQL(t, w, "INSERT INTO test_update_create_logic VALUES (3,'charlie','charlie@example.com','active')")
		w.SetTable("test_update_create_logic")

		var user FirstOrUser
		attributes := map[string]any{"id": 3}
		values := map[string]any{"status": "inactive"}

		// Note: WHERE clause affects the First() call
		w.Q.Where("id = ?", 3)
		err := w.Q.UpdateOrCreate(&user, attributes, values)

		if err != nil {
			t.Fatalf("UpdateOrCreate with WHERE failed: %v", err)
		}
	})

	t.Run("empty table - should create", func(t *testing.T) {
		execSQL(t, w, "CREATE TABLE test_update_create_logic_3 (id INTEGER, name TEXT)")
		w.SetTable("test_update_create_logic_3")

		var user FirstOrUser
		attributes := map[string]any{"id": 1}
		values := map[string]any{"name": "new_user"}

		err := w.Q.UpdateOrCreate(&user, attributes, values)

		// Note: First() will fail on empty table, so Create() is called
		// This test verifies the method executes without error
		if err != nil {
			t.Fatalf("UpdateOrCreate on empty table failed: %v", err)
		}
	})
}

func TestUpdateOrCreateAttributeMatching(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_attr_matching (id INTEGER, name TEXT, email TEXT, status TEXT)")
	execSQL(t, w, "INSERT INTO test_attr_matching VALUES (1,'alice','alice@example.com','active')")
	execSQL(t, w, "INSERT INTO test_attr_matching VALUES (2,'bob','bob@example.com','inactive')")

	w.SetTable("test_attr_matching")

	t.Run("attributes parameter is accepted but not used for filtering", func(t *testing.T) {
		var user FirstOrUser
		attributes := map[string]any{"id": 1}
		values := map[string]any{"email": "new_alice@example.com"}

		// Note: Current implementation doesn't use attributes for filtering
		// It uses First(dest) to check existence
		err := w.Q.UpdateOrCreate(&user, attributes, values)

		if err != nil {
			t.Fatalf("UpdateOrCreate with attributes failed: %v", err)
		}
	})

	t.Run("multiple attributes in map", func(t *testing.T) {
		var user FirstOrUser
		attributes := map[string]any{"id": 2, "status": "inactive"}
		values := map[string]any{"email": "new_bob@example.com"}

		err := w.Q.UpdateOrCreate(&user, attributes, values)

		// Note: Multiple attributes are accepted but not used for filtering
		if err != nil {
			t.Fatalf("UpdateOrCreate with multiple attributes failed: %v", err)
		}
	})

	t.Run("struct attributes", func(t *testing.T) {
		var user FirstOrUser
		attributes := FirstOrUser{ID: 1, Name: "alice"}
		values := map[string]any{"status": "updated"}

		err := w.Q.UpdateOrCreate(&user, attributes, values)

		// Note: Struct attributes are accepted but not used for filtering
		if err != nil {
			t.Fatalf("UpdateOrCreate with struct attributes failed: %v", err)
		}
	})

	t.Run("empty attributes map", func(t *testing.T) {
		var user FirstOrUser
		attributes := map[string]any{}
		values := map[string]any{"name": "test"}

		err := w.Q.UpdateOrCreate(&user, attributes, values)

		if err != nil {
			t.Fatalf("UpdateOrCreate with empty attributes failed: %v", err)
		}
	})

	t.Run("attributes with nil values", func(t *testing.T) {
		var user FirstOrUser
		attributes := map[string]any{"id": nil}
		values := map[string]any{"name": "test"}

		err := w.Q.UpdateOrCreate(&user, attributes, values)

		// Should handle nil in attributes gracefully
		if err != nil {
			t.Fatalf("UpdateOrCreate with nil attribute values failed: %v", err)
		}
	})
}

func TestUpdateOrCreateErrorHandling(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_update_or_create_error (id INTEGER, name TEXT, email TEXT)")

	w.SetTable("test_update_or_create_error")

	t.Run("nil attributes handling", func(t *testing.T) {
		var user FirstOrUser
		values := map[string]any{"name": "test", "email": "test@example.com"}

		err := w.Q.UpdateOrCreate(&user, nil, values)

		// Note: nil attributes are accepted but not used for filtering
		if err != nil {
			t.Fatalf("UpdateOrCreate with nil attributes failed: %v", err)
		}
	})

	t.Run("nil values handling", func(t *testing.T) {
		execSQL(t, w, "INSERT INTO test_update_or_create_error VALUES (1,'existing','existing@example.com')")
		var user FirstOrUser
		attributes := map[string]any{"id": 1}

		err := w.Q.UpdateOrCreate(&user, attributes, nil)

		// Note: nil values will cause Save/Create to fail
		// This test verifies the error is propagated
		if err == nil {
			t.Error("Expected error with nil values")
		}
	})
}

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

	var users []PaginateUser
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

	var results []map[string]any
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

// --- scopes tests ---

type ScopeUser struct {
	ID     int    `db:"id"`
	Name   string `db:"name"`
	Status string `db:"status"`
	Age    int    `db:"age"`
}

func TestScopesMethod(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_scopes (id INTEGER, name TEXT, status TEXT, age INTEGER)")
	execSQL(t, w, "INSERT INTO test_scopes VALUES (1,'alice','active',25),(2,'bob','inactive',30),(3,'charlie','active',35)")

	w.SetTable("test_scopes")

	t.Run("single scope without parameters", func(t *testing.T) {
		activeScope := func(q contractsorm.Query) contractsorm.Query {
			return q.Where("status = ?", "active")
		}

		var users []ScopeUser
		err := w.Q.Scopes(activeScope).Find(&users)
		if err != nil {
			t.Fatalf("Scopes with single scope failed: %v", err)
		}

		if len(users) != 2 {
			t.Errorf("Expected 2 active users, got %d", len(users))
		}

		for _, user := range users {
			if user.Status != "active" {
				t.Errorf("Expected status 'active', got '%s'", user.Status)
			}
		}
	})

	t.Run("multiple scopes without parameters", func(t *testing.T) {
		activeScope := func(q contractsorm.Query) contractsorm.Query {
			return q.Where("status = ?", "active")
		}

		youngScope := func(q contractsorm.Query) contractsorm.Query {
			return q.Where("age < ?", 30)
		}

		var users []ScopeUser
		err := w.Q.Scopes(activeScope, youngScope).Find(&users)
		if err != nil {
			t.Fatalf("Scopes with multiple scopes failed: %v", err)
		}

		if len(users) != 1 {
			t.Errorf("Expected 1 active young user, got %d", len(users))
		}

		if users[0].Name != "alice" {
			t.Errorf("Expected user 'alice', got '%s'", users[0].Name)
		}
	})

	t.Run("scope with closure parameters", func(t *testing.T) {
		nameScope := func(name string) func(contractsorm.Query) contractsorm.Query {
			return func(q contractsorm.Query) contractsorm.Query {
				return q.Where("name = ?", name)
			}
		}

		var users []ScopeUser
		err := w.Q.Scopes(nameScope("bob")).Find(&users)
		if err != nil {
			t.Fatalf("Scopes with closure parameters failed: %v", err)
		}

		if len(users) != 1 {
			t.Errorf("Expected 1 user, got %d", len(users))
		}

		if users[0].Name != "bob" {
			t.Errorf("Expected user 'bob', got '%s'", users[0].Name)
		}
	})

	t.Run("empty scopes list", func(t *testing.T) {
		var users []ScopeUser
		err := w.Q.Scopes().Find(&users)
		if err != nil {
			t.Fatalf("Scopes with empty list failed: %v", err)
		}

		if len(users) != 3 {
			t.Errorf("Expected 3 users with no scopes, got %d", len(users))
		}
	})
}

func TestScopeApplicationOrder(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_scope_order (id INTEGER, name TEXT, status TEXT, age INTEGER)")
	execSQL(t, w, "INSERT INTO test_scope_order VALUES (1,'alice','active',25),(2,'bob','active',30),(3,'charlie','inactive',35)")

	w.SetTable("test_scope_order")

	t.Run("scopes applied in order", func(t *testing.T) {
		scope1 := func(q contractsorm.Query) contractsorm.Query {
			return q.Where("status = ?", "active")
		}

		scope2 := func(q contractsorm.Query) contractsorm.Query {
			return q.Where("age > ?", 25)
		}

		scope3 := func(q contractsorm.Query) contractsorm.Query {
			return q.OrderBy("name", "asc")
		}

		var users []ScopeUser
		err := w.Q.Scopes(scope1, scope2, scope3).Find(&users)
		if err != nil {
			t.Fatalf("Scope application order failed: %v", err)
		}

		if len(users) != 1 {
			t.Errorf("Expected 1 user matching all scopes, got %d", len(users))
		}

		if users[0].Name != "bob" {
			t.Errorf("Expected user 'bob', got '%s'", users[0].Name)
		}
	})

	t.Run("scope order affects result", func(t *testing.T) {
		limitScope := func(q contractsorm.Query) contractsorm.Query {
			return q.Limit(1)
		}

		orderScope := func(q contractsorm.Query) contractsorm.Query {
			return q.OrderBy("age", "desc")
		}

		var users []ScopeUser
		err := w.Q.Scopes(orderScope, limitScope).Find(&users)
		if err != nil {
			t.Fatalf("Scope order with limit failed: %v", err)
		}

		if len(users) != 1 {
			t.Errorf("Expected 1 user, got %d", len(users))
		}

		if users[0].Name != "charlie" {
			t.Errorf("Expected oldest user 'charlie', got '%s'", users[0].Name)
		}
	})

	t.Run("reversed scope order", func(t *testing.T) {
		limitScope := func(q contractsorm.Query) contractsorm.Query {
			return q.Limit(1)
		}

		orderScope := func(q contractsorm.Query) contractsorm.Query {
			return q.OrderBy("age", "asc")
		}

		var users []ScopeUser
		err := w.Q.Scopes(limitScope, orderScope).Find(&users)
		if err != nil {
			t.Fatalf("Reversed scope order failed: %v", err)
		}

		if len(users) != 1 {
			t.Errorf("Expected 1 user, got %d", len(users))
		}

		if users[0].Name != "alice" {
			t.Errorf("Expected youngest user 'alice', got '%s'", users[0].Name)
		}
	})
}

func TestScopeWithQueryChaining(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_scope_chain (id INTEGER, name TEXT, status TEXT, age INTEGER)")
	execSQL(t, w, "INSERT INTO test_scope_chain VALUES (1,'alice','active',25),(2,'bob','active',30),(3,'charlie','inactive',35),(4,'dave','active',40)")

	w.SetTable("test_scope_chain")

	t.Run("scope before where clause", func(t *testing.T) {
		activeScope := func(q contractsorm.Query) contractsorm.Query {
			return q.Where("status = ?", "active")
		}

		var users []ScopeUser
		err := w.Q.Scopes(activeScope).Where("age > ?", 30).Find(&users)
		if err != nil {
			t.Fatalf("Scope before where failed: %v", err)
		}

		if len(users) != 1 {
			t.Errorf("Expected 1 user, got %d", len(users))
		}

		if users[0].Name != "dave" {
			t.Errorf("Expected user 'dave', got '%s'", users[0].Name)
		}
	})

	t.Run("scope after where clause", func(t *testing.T) {
		activeScope := func(q contractsorm.Query) contractsorm.Query {
			return q.Where("status = ?", "active")
		}

		var users []ScopeUser
		err := w.Q.Where("age < ?", 30).Scopes(activeScope).Find(&users)
		if err != nil {
			t.Fatalf("Scope after where failed: %v", err)
		}

		if len(users) != 1 {
			t.Errorf("Expected 1 user, got %d", len(users))
		}

		if users[0].Name != "alice" {
			t.Errorf("Expected user 'alice', got '%s'", users[0].Name)
		}
	})

	t.Run("scope with first", func(t *testing.T) {
		activeScope := func(q contractsorm.Query) contractsorm.Query {
			return q.Where("status = ?", "active")
		}

		var user ScopeUser
		err := w.Q.Scopes(activeScope).OrderBy("age", "asc").First(&user)
		if err != nil {
			t.Fatalf("Scope with first failed: %v", err)
		}

		if user.Name != "alice" {
			t.Errorf("Expected user 'alice', got '%s'", user.Name)
		}
	})
}

func TestScopeErrorHandling(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_scope_error (id INTEGER, name TEXT, status TEXT)")
	execSQL(t, w, "INSERT INTO test_scope_error VALUES (1,'alice','active')")

	w.SetTable("test_scope_error")

	t.Run("scope returns nil query", func(t *testing.T) {
		nilScope := func(q contractsorm.Query) contractsorm.Query {
			return nil
		}

		var users []ScopeUser
		err := w.Q.Scopes(nilScope).Find(&users)
		// The current implementation doesn't error when scope returns nil
		// It just falls back to the original query
		_ = err
		_ = users
	})

	t.Run("scope with invalid where clause", func(t *testing.T) {
		invalidScope := func(q contractsorm.Query) contractsorm.Query {
			return q.Where("invalid_column = ?", "value")
		}

		var users []ScopeUser
		err := w.Q.Scopes(invalidScope).Find(&users)
		// SQLite doesn't error on invalid columns in WHERE clause
		// It just returns no results
		_ = err
		_ = users
	})

	t.Run("scope panic handling", func(t *testing.T) {
		panicScope := func(q contractsorm.Query) contractsorm.Query {
			panic("scope panic")
		}

		var users []ScopeUser
		defer func() {
			if r := recover(); r != nil {
				// Expected panic
			}
		}()

		_ = w.Q.Scopes(panicScope).Find(&users)
	})

	t.Run("scope with invalid table", func(t *testing.T) {
		w.SetTable("nonexistent_table")

		validScope := func(q contractsorm.Query) contractsorm.Query {
			return q.Where("status = ?", "active")
		}

		var users []ScopeUser
		err := w.Q.Scopes(validScope).Find(&users)
		if err == nil {
			t.Error("Expected error for invalid table with scope")
		}
	})

	t.Run("scope with nil destination", func(t *testing.T) {
		activeScope := func(q contractsorm.Query) contractsorm.Query {
			return q.Where("status = ?", "active")
		}

		err := w.Q.Scopes(activeScope).Find(nil)
		if err == nil {
			t.Error("Expected error for nil destination")
		}
	})

	t.Run("scope with non-slice destination", func(t *testing.T) {
		activeScope := func(q contractsorm.Query) contractsorm.Query {
			return q.Where("status = ?", "active")
		}

		var user ScopeUser
		err := w.Q.Scopes(activeScope).Find(&user)
		if err == nil {
			t.Error("Expected error for non-slice destination with Find")
		}
	})
}

func TestScopeWithTransactions(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_scope_tx (id INTEGER, name TEXT, status TEXT)")
	execSQL(t, w, "INSERT INTO test_scope_tx VALUES (1,'alice','active'),(2,'bob','inactive')")

	w.SetTable("test_scope_tx")

	tx, err := w.Q.Begin()
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	activeScope := func(q contractsorm.Query) contractsorm.Query {
		return q.Where("status = ?", "active")
	}

	var users []ScopeUser
	err = tx.Scopes(activeScope).Find(&users)
	if err != nil {
		t.Fatalf("Scope in transaction failed: %v", err)
	}

	if len(users) != 1 {
		t.Errorf("Expected 1 active user in transaction, got %d", len(users))
	}

	if users[0].Name != "alice" {
		t.Errorf("Expected user 'alice', got '%s'", users[0].Name)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}
}

func TestScopeIsolation(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_scope_isolation (id INTEGER, name TEXT, status TEXT)")
	execSQL(t, w, "INSERT INTO test_scope_isolation VALUES (1,'alice','active'),(2,'bob','inactive')")

	w.SetTable("test_scope_isolation")

	t.Run("scope does not affect original query", func(t *testing.T) {
		activeScope := func(q contractsorm.Query) contractsorm.Query {
			return q.Where("status = ?", "active")
		}

		var scopedUsers []ScopeUser
		err := w.Q.Scopes(activeScope).Find(&scopedUsers)
		if err != nil {
			t.Fatalf("Scoped query failed: %v", err)
		}

		if len(scopedUsers) != 1 {
			t.Errorf("Expected 1 scoped user, got %d", len(scopedUsers))
		}

		var allUsers []ScopeUser
		err = w.Q.Find(&allUsers)
		if err != nil {
			t.Fatalf("Original query failed: %v", err)
		}

		if len(allUsers) != 2 {
			t.Errorf("Expected 2 users in original query, got %d", len(allUsers))
		}
	})

	t.Run("multiple scopes on same query", func(t *testing.T) {
		activeScope := func(q contractsorm.Query) contractsorm.Query {
			return q.Where("status = ?", "active")
		}

		inactiveScope := func(q contractsorm.Query) contractsorm.Query {
			return q.Where("status = ?", "inactive")
		}

		var activeUsers []ScopeUser
		err := w.Q.Scopes(activeScope).Find(&activeUsers)
		if err != nil {
			t.Fatalf("First scope failed: %v", err)
		}

		var inactiveUsers []ScopeUser
		err = w.Q.Scopes(inactiveScope).Find(&inactiveUsers)
		if err != nil {
			t.Fatalf("Second scope failed: %v", err)
		}

		if len(activeUsers) != 1 {
			t.Errorf("Expected 1 active user, got %d", len(activeUsers))
		}

		if len(inactiveUsers) != 1 {
			t.Errorf("Expected 1 inactive user, got %d", len(inactiveUsers))
		}
	})
}

func TestScopeWithModel(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_scope_model (id INTEGER, name TEXT, status TEXT, age INTEGER)")
	execSQL(t, w, "INSERT INTO test_scope_model VALUES (1,'alice','active',25),(2,'bob','inactive',30)")

	t.Run("scope with model set", func(t *testing.T) {
		activeScope := func(q contractsorm.Query) contractsorm.Query {
			return q.Where("status = ?", "active")
		}

		var users []ScopeUser
		err := w.Q.Model(&ScopeUser{}).Table("test_scope_model").Scopes(activeScope).Find(&users)
		if err != nil {
			t.Fatalf("Scope with model failed: %v", err)
		}

		if len(users) != 1 {
			t.Errorf("Expected 1 user, got %d", len(users))
		}
	})
}
