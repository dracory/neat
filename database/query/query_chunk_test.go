package query_test

import (
	"fmt"
	"testing"
)

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

	capturedChunks := make([][]map[string]any, 0)

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
