//go:build integration

package query

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/dracory/neat/contracts/log"
)

// PerformanceTestModel represents a model for performance testing
type PerformanceTestModel struct {
	ID        uint      `db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	Age       int       `db:"age"`
	Score     float64   `db:"score"`
	Active    bool      `db:"active"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (PerformanceTestModel) TableName() string {
	return "performance_test_models"
}

// setupPerformanceTestDB creates and populates a test database with large datasets
func setupPerformanceTestDB(t *testing.T, recordCount int) *Query {
	t.Helper()

	ctx := context.Background()
	query := NewQuery(ctx, nil, nil, "default", nil, log.NewNoopLogger())

	// Create table
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS performance_test_models (
			id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
			name       VARCHAR(255) NOT NULL,
			email      VARCHAR(255) NOT NULL,
			age        INT NOT NULL,
			score      DOUBLE NOT NULL,
			active     BOOLEAN NOT NULL DEFAULT TRUE,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`

	// Execute table creation
	if err := query.Raw(createTableSQL).Exec(); err != nil {
		t.Fatalf("Failed to create performance test table: %v", err)
	}

	// Drop existing data
	query.Raw("TRUNCATE TABLE performance_test_models").Exec()

	// Bulk insert test data
	batchSize := 1000
	for i := 0; i < recordCount; i += batchSize {
		end := i + batchSize
		if end > recordCount {
			end = recordCount
		}

		records := make([]PerformanceTestModel, end-i)
		for j := i; j < end; j++ {
			records[j-i] = PerformanceTestModel{
				Name:  fmt.Sprintf("User_%d", j),
				Email: fmt.Sprintf("user_%d@example.com", j),
				Age:   20 + (j % 50),
				Score: float64(j % 100),
				Active: (j % 2) == 0,
			}
		}

		if err := query.Create(&records); err != nil {
			t.Fatalf("Failed to insert batch %d: %v", i/batchSize, err)
		}
	}

	t.Cleanup(func() {
		query.Raw("DROP TABLE IF EXISTS performance_test_models").Exec()
	})

	return query
}

// BenchmarkLargeDatasetSelect benchmarks SELECT operations on large datasets
func BenchmarkLargeDatasetSelect_10K(b *testing.B) {
	benchmarkLargeDatasetSelect(b, 10000)
}

func BenchmarkLargeDatasetSelect_100K(b *testing.B) {
	benchmarkLargeDatasetSelect(b, 100000)
}

func benchmarkLargeDatasetSelect(b *testing.B, recordCount int) {
	if testing.Short() {
		b.Skip("Skipping performance benchmark in short mode")
	}

	query := setupPerformanceTestDB(b, recordCount)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var results []PerformanceTestModel
		if err := query.Where("active", true).Limit(100).Get(&results); err != nil {
			b.Fatalf("Query failed: %v", err)
		}
	}
}

// BenchmarkLargeDatasetWhere benchmarks WHERE clause performance on large datasets
func BenchmarkLargeDatasetWhere_10K(b *testing.B) {
	benchmarkLargeDatasetWhere(b, 10000)
}

func BenchmarkLargeDatasetWhere_100K(b *testing.B) {
	benchmarkLargeDatasetWhere(b, 100000)
}

func benchmarkLargeDatasetWhere(b *testing.B, recordCount int) {
	if testing.Short() {
		b.Skip("Skipping performance benchmark in short mode")
	}

	query := setupPerformanceTestDB(b, recordCount)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var results []PerformanceTestModel
		if err := query.Where("age", ">", 30).Where("active", true).Get(&results); err != nil {
			b.Fatalf("Query failed: %v", err)
		}
	}
}

// BenchmarkCursorStreaming benchmarks cursor streaming performance
func BenchmarkCursorStreaming_10K(b *testing.B) {
	benchmarkCursorStreaming(b, 10000)
}

func BenchmarkCursorStreaming_100K(b *testing.B) {
	benchmarkCursorStreaming(b, 100000)
}

func benchmarkCursorStreaming(b *testing.B, recordCount int) {
	if testing.Short() {
		b.Skip("Skipping performance benchmark in short mode")
	}

	query := setupPerformanceTestDB(b, recordCount)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cursor, err := query.Cursor()
		if err != nil {
			b.Fatalf("Cursor failed: %v", err)
		}

		count := 0
		for range cursor {
			count++
			if count >= 1000 {
				break
			}
		}
	}
}

// BenchmarkChunkProcessing benchmarks chunk processing performance
func BenchmarkChunkProcessing_10K(b *testing.B) {
	benchmarkChunkProcessing(b, 10000)
}

func BenchmarkChunkProcessing_100K(b *testing.B) {
	benchmarkChunkProcessing(b, 100000)
}

func benchmarkChunkProcessing(b *testing.B, recordCount int) {
	if testing.Short() {
		b.Skip("Skipping performance benchmark in short mode")
	}

	query := setupPerformanceTestDB(b, recordCount)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count := 0
		err := query.Chunk(100, func(results []PerformanceTestModel) bool {
			count += len(results)
			return count < 1000
		})
		if err != nil {
			b.Fatalf("Chunk failed: %v", err)
		}
	}
}

// BenchmarkMemoryUsage measures memory allocation patterns for streaming operations
func BenchmarkMemoryUsage_Cursor(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping memory benchmark in short mode")
	}

	query := setupPerformanceTestDB(b, 10000)

	b.ResetTimer()
	allocs := b.AllocsPerRun(5, func() {
		cursor, err := query.Cursor()
		if err != nil {
			b.Fatalf("Cursor failed: %v", err)
		}

		count := 0
		for range cursor {
			count++
			if count >= 1000 {
				break
			}
		}
	})

	b.ReportMetric(allocs, "allocs/op")
}

// BenchmarkMemoryUsage_Chunk measures memory allocation patterns for chunk operations
func BenchmarkMemoryUsage_Chunk(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping memory benchmark in short mode")
	}

	query := setupPerformanceTestDB(b, 10000)

	b.ResetTimer()
	allocs := b.AllocsPerRun(5, func() {
		count := 0
		err := query.Chunk(100, func(results []PerformanceTestModel) bool {
			count += len(results)
			return count < 1000
		})
		if err != nil {
			b.Fatalf("Chunk failed: %v", err)
		}
	})

	b.ReportMetric(allocs, "allocs/op")
}

// BenchmarkMemoryUsage_Get measures memory allocation patterns for Get operations
func BenchmarkMemoryUsage_Get(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping memory benchmark in short mode")
	}

	query := setupPerformanceTestDB(b, 10000)

	b.ResetTimer()
	allocs := b.AllocsPerRun(5, func() {
		var results []PerformanceTestModel
		if err := query.Limit(1000).Get(&results); err != nil {
			b.Fatalf("Get failed: %v", err)
		}
	})

	b.ReportMetric(allocs, "allocs/op")
}

// BenchmarkComplexQuery benchmarks complex query performance with joins and aggregations
func BenchmarkComplexQuery_10K(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping performance benchmark in short mode")
	}

	query := setupPerformanceTestDB(b, 10000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var results []PerformanceTestModel
		if err := query.
			Where("age", ">=", 25).
			Where("active", true).
			Where("score", ">", 50).
			OrderBy("created_at", "desc").
			Limit(100).
			Get(&results); err != nil {
			b.Fatalf("Complex query failed: %v", err)
		}
	}
}

// BenchmarkPagination benchmarks pagination performance
func BenchmarkPagination_10K(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping performance benchmark in short mode")
	}

	query := setupPerformanceTestDB(b, 10000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		page := (i % 100) + 1
		var results []PerformanceTestModel
		if err := query.Paginate(page, 50, &results); err != nil {
			b.Fatalf("Pagination failed: %v", err)
		}
	}
}

// BenchmarkUpdate benchmarks update performance on large datasets
func BenchmarkUpdate_10K(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping performance benchmark in short mode")
	}

	query := setupPerformanceTestDB(b, 10000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := query.Where("id", "<=", 100).Update("score", 99.9); err != nil {
			b.Fatalf("Update failed: %v", err)
		}
	}
}

// BenchmarkDelete benchmarks delete performance on large datasets
func BenchmarkDelete_10K(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping performance benchmark in short mode")
	}

	query := setupPerformanceTestDB(b, 10000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Re-insert data for each iteration
		records := make([]PerformanceTestModel, 10)
		for j := 0; j < 10; j++ {
			records[j] = PerformanceTestModel{
				Name:  fmt.Sprintf("Temp_%d_%d", i, j),
				Email: fmt.Sprintf("temp_%d_%d@example.com", i, j),
				Age:   20,
				Score: 50.0,
				Active: true,
			}
		}
		query.Create(&records)

		// Delete the records
		if err := query.Where("name", "like", "Temp_%").Delete(); err != nil {
			b.Fatalf("Delete failed: %v", err)
		}
	}
}

// TestPerformanceBaseline establishes baseline performance metrics
func TestPerformanceBaseline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping baseline test in short mode")
	}

	recordCount := 10000
	query := setupPerformanceTestDB(t, recordCount)

	t.Run("Select_1000_records", func(t *testing.T) {
		start := time.Now()
		var results []PerformanceTestModel
		err := query.Limit(1000).Get(&results)
		duration := time.Since(start)

		if err != nil {
			t.Fatalf("Select failed: %v", err)
		}

		t.Logf("Selected %d records in %v", len(results), duration)
		t.Logf("Performance baseline: %v per 1000 records", duration)
	})

	t.Run("Cursor_stream_1000_records", func(t *testing.T) {
		start := time.Now()
		cursor, err := query.Cursor()
		if err != nil {
			t.Fatalf("Cursor failed: %v", err)
		}

		count := 0
		for range cursor {
			count++
			if count >= 1000 {
				break
			}
		}
		duration := time.Since(start)

		t.Logf("Streamed %d records via cursor in %v", count, duration)
	})

	t.Run("Chunk_1000_records", func(t *testing.T) {
		start := time.Now()
		count := 0
		err := query.Chunk(100, func(results []PerformanceTestModel) bool {
			count += len(results)
			return count < 1000
		})
		duration := time.Since(start)

		if err != nil {
			t.Fatalf("Chunk failed: %v", err)
		}

		t.Logf("Chunked %d records in %v", count, duration)
	})

	t.Run("Memory_Stats", func(t *testing.T) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		t.Logf("Memory stats before operations:")
		t.Logf("  Alloc: %v bytes", m.Alloc)
		t.Logf("  TotalAlloc: %v bytes", m.TotalAlloc)
		t.Logf("  Sys: %v bytes", m.Sys)
		t.Logf("  NumGC: %d", m.NumGC)
	})
}
