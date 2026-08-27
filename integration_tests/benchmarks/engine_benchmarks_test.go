//go:build integration

// Package benchmarks contains per-engine integration benchmarks that measure
// real database round-trip performance for each supported backend.
//
// Each engine benchmark follows the same pattern: create a connection, set up
// a benchmark table, populate it with N rows, then measure SELECT, WHERE,
// INSERT, UPDATE, and DELETE operations against that engine.
//
// Run with:
//
//	go test -tags=integration -run=^$ -bench=. -benchmem ./integration_tests/benchmarks/...
package benchmarks

import (
	"fmt"
	"testing"
	"time"

	"github.com/dracory/neat"
	"github.com/dracory/neat/contracts/log"
	"github.com/dracory/neat/database"
	"github.com/dracory/neat/integration_tests/common"
	_ "github.com/go-sql-driver/mysql"
)

// benchModel is the model used for all per-engine benchmarks.
type benchModel struct {
	ID        uint      `db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	Age       int       `db:"age"`
	Active    bool      `db:"active"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (benchModel) TableName() string {
	return "bench_models"
}

// benchRecordCount is the number of rows inserted before measuring.
const benchRecordCount = 1000

// --- MySQL ---

func BenchmarkMySQL_Select(b *testing.B) {
	db := setupEngineBench(b, "mysql")
	q := db.Query().Model(benchModel{})

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		var results []benchModel
		if err := q.Limit(100).Get(&results); err != nil {
			b.Fatalf("Select failed: %v", err)
		}
	}
}

func BenchmarkMySQL_Where(b *testing.B) {
	db := setupEngineBench(b, "mysql")
	q := db.Query().Model(benchModel{})

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		var results []benchModel
		if err := q.Where("age", ">", 30).Where("active", true).Limit(100).Get(&results); err != nil {
			b.Fatalf("Where failed: %v", err)
		}
	}
}

func BenchmarkMySQL_Insert(b *testing.B) {
	db := setupEngineBench(b, "mysql")
	q := db.Query().Model(benchModel{})

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		m := benchModel{
			Name:   fmt.Sprintf("bench_%d", b.N),
			Email:  fmt.Sprintf("bench_%d@example.com", b.N),
			Age:    25,
			Active: true,
		}
		if err := q.Create(&m); err != nil {
			b.Fatalf("Insert failed: %v", err)
		}
	}
}

func BenchmarkMySQL_Update(b *testing.B) {
	db := setupEngineBench(b, "mysql")
	q := db.Query().Model(benchModel{})

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		if _, err := q.Where("id", "<=", 100).Update("age", 99); err != nil {
			b.Fatalf("Update failed: %v", err)
		}
	}
}

func BenchmarkMySQL_ToSql(b *testing.B) {
	db := setupEngineBench(b, "mysql")
	q := db.Query().Model(benchModel{}).
		Where("age", ">=", 25).
		Where("active", true).
		OrderBy("id", "desc").
		Limit(50)

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = q.ToSql().Get(nil)
	}
}

// --- PostgreSQL ---

func BenchmarkPostgres_Select(b *testing.B) {
	db := setupEngineBench(b, "postgres")
	q := db.Query().Model(benchModel{})

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		var results []benchModel
		if err := q.Limit(100).Get(&results); err != nil {
			b.Fatalf("Select failed: %v", err)
		}
	}
}

func BenchmarkPostgres_Where(b *testing.B) {
	db := setupEngineBench(b, "postgres")
	q := db.Query().Model(benchModel{})

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		var results []benchModel
		if err := q.Where("age", ">", 30).Where("active", true).Limit(100).Get(&results); err != nil {
			b.Fatalf("Where failed: %v", err)
		}
	}
}

func BenchmarkPostgres_Insert(b *testing.B) {
	db := setupEngineBench(b, "postgres")
	q := db.Query().Model(benchModel{})

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		m := benchModel{
			Name:   fmt.Sprintf("bench_%d", b.N),
			Email:  fmt.Sprintf("bench_%d@example.com", b.N),
			Age:    25,
			Active: true,
		}
		if err := q.Create(&m); err != nil {
			b.Fatalf("Insert failed: %v", err)
		}
	}
}

func BenchmarkPostgres_Update(b *testing.B) {
	db := setupEngineBench(b, "postgres")
	q := db.Query().Model(benchModel{})

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		if _, err := q.Where("id", "<=", 100).Update("age", 99); err != nil {
			b.Fatalf("Update failed: %v", err)
		}
	}
}

func BenchmarkPostgres_ToSql(b *testing.B) {
	db := setupEngineBench(b, "postgres")
	q := db.Query().Model(benchModel{}).
		Where("age", ">=", 25).
		Where("active", true).
		OrderBy("id", "desc").
		Limit(50)

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = q.ToSql().Get(nil)
	}
}

// --- SQLite ---

func BenchmarkSQLite_Select(b *testing.B) {
	db := setupEngineBench(b, "sqlite")
	q := db.Query().Model(benchModel{})

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		var results []benchModel
		if err := q.Limit(100).Get(&results); err != nil {
			b.Fatalf("Select failed: %v", err)
		}
	}
}

func BenchmarkSQLite_Where(b *testing.B) {
	db := setupEngineBench(b, "sqlite")
	q := db.Query().Model(benchModel{})

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		var results []benchModel
		if err := q.Where("age", ">", 30).Where("active", true).Limit(100).Get(&results); err != nil {
			b.Fatalf("Where failed: %v", err)
		}
	}
}

func BenchmarkSQLite_Insert(b *testing.B) {
	db := setupEngineBench(b, "sqlite")
	q := db.Query().Model(benchModel{})

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		m := benchModel{
			Name:   fmt.Sprintf("bench_%d", b.N),
			Email:  fmt.Sprintf("bench_%d@example.com", b.N),
			Age:    25,
			Active: true,
		}
		if err := q.Create(&m); err != nil {
			b.Fatalf("Insert failed: %v", err)
		}
	}
}

func BenchmarkSQLite_Update(b *testing.B) {
	db := setupEngineBench(b, "sqlite")
	q := db.Query().Model(benchModel{})

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		if _, err := q.Where("id", "<=", 100).Update("age", 99); err != nil {
			b.Fatalf("Update failed: %v", err)
		}
	}
}

func BenchmarkSQLite_ToSql(b *testing.B) {
	db := setupEngineBench(b, "sqlite")
	q := db.Query().Model(benchModel{}).
		Where("age", ">=", 25).
		Where("active", true).
		OrderBy("id", "desc").
		Limit(50)

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = q.ToSql().Get(nil)
	}
}

// --- SQL Server ---

func BenchmarkSQLServer_Select(b *testing.B) {
	db := setupEngineBench(b, "sqlserver")
	q := db.Query().Model(benchModel{})

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		var results []benchModel
		if err := q.Limit(100).Get(&results); err != nil {
			b.Fatalf("Select failed: %v", err)
		}
	}
}

func BenchmarkSQLServer_Where(b *testing.B) {
	db := setupEngineBench(b, "sqlserver")
	q := db.Query().Model(benchModel{})

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		var results []benchModel
		if err := q.Where("age", ">", 30).Where("active", true).Limit(100).Get(&results); err != nil {
			b.Fatalf("Where failed: %v", err)
		}
	}
}

func BenchmarkSQLServer_Insert(b *testing.B) {
	db := setupEngineBench(b, "sqlserver")
	q := db.Query().Model(benchModel{})

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		m := benchModel{
			Name:   fmt.Sprintf("bench_%d", b.N),
			Email:  fmt.Sprintf("bench_%d@example.com", b.N),
			Age:    25,
			Active: true,
		}
		if err := q.Create(&m); err != nil {
			b.Fatalf("Insert failed: %v", err)
		}
	}
}

func BenchmarkSQLServer_Update(b *testing.B) {
	db := setupEngineBench(b, "sqlserver")
	q := db.Query().Model(benchModel{})

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		if _, err := q.Where("id", "<=", 100).Update("age", 99); err != nil {
			b.Fatalf("Update failed: %v", err)
		}
	}
}

func BenchmarkSQLServer_ToSql(b *testing.B) {
	db := setupEngineBench(b, "sqlserver")
	q := db.Query().Model(benchModel{}).
		Where("age", ">=", 25).
		Where("active", true).
		OrderBy("id", "desc").
		Limit(50)

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = q.ToSql().Get(nil)
	}
}

// --- setupEngineBench ---

// setupEngineBench creates a connection to the given engine, creates the
// benchmark table, and populates it with benchRecordCount rows.
// It returns the *database.Database for use in benchmarks.
func setupEngineBench(b *testing.B, engine string) *database.Database {
	b.Helper()
	if testing.Short() {
		b.Skip("Skipping integration benchmark in short mode")
	}

	dsn := engineDSN(b, engine)
	db, err := neat.NewFromDSN(dsn, database.WithLogger(log.NewNoopLogger()))
	if err != nil {
		b.Fatalf("Failed to connect to %s: %v", engine, err)
	}

	createBenchTable(b, db, engine)
	populateBenchTable(b, db)

	b.Cleanup(func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			_, _ = sqlDB.Exec("DROP TABLE IF EXISTS bench_models")
		}
		_ = db.Close()
	})

	return db
}

// engineDSN returns a DSN for the given engine based on environment variables.
func engineDSN(b *testing.B, engine string) string {
	b.Helper()
	switch engine {
	case "mysql":
		host := common.GetEnv("MYSQL_HOST", "127.0.0.1")
		port := common.GetEnvInt("MYSQL_PORT", 3306)
		dbName := common.GetEnv("MYSQL_DATABASE", "test")
		user := common.GetEnv("MYSQL_USER", "root")
		pass := common.GetEnv("MYSQL_PASS", "root")
		return fmt.Sprintf("mysql://%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local",
			user, pass, host, port, dbName)

	case "postgres":
		host := common.GetEnv("POSTGRES_HOST", "127.0.0.1")
		port := common.GetEnvInt("POSTGRES_PORT", 5432)
		dbName := common.GetEnv("POSTGRES_DATABASE", "test")
		user := common.GetEnv("POSTGRES_USER", "test")
		pass := common.GetEnv("POSTGRES_PASS", "test")
		sslmode := common.GetEnv("POSTGRES_SSLMODE", "disable")
		return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			user, pass, host, port, dbName, sslmode)

	case "sqlite":
		return "sqlite://:memory:?multi_stmts=true"

	case "sqlserver":
		host := common.GetEnv("SQLSERVER_HOST", "127.0.0.1")
		port := common.GetEnvInt("SQLSERVER_PORT", 1433)
		dbName := common.GetEnv("SQLSERVER_DATABASE", "test")
		user := common.GetEnv("SQLSERVER_USER", "sa")
		pass := common.GetEnv("SQLSERVER_PASS", "YourStrong@Passw0rd")
		return fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
			user, pass, host, port, dbName)

	default:
		b.Fatalf("Unknown engine: %s", engine)
		return ""
	}
}

// createBenchTable creates the benchmark table with engine-appropriate SQL.
func createBenchTable(b *testing.B, db *database.Database, engine string) {
	b.Helper()
	sqlDB, err := db.DB()
	if err != nil {
		b.Fatalf("createBenchTable: DB(): %v", err)
	}

	var stmt string
	switch engine {
	case "mysql":
		stmt = `CREATE TABLE IF NOT EXISTS bench_models (
			id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
			name       VARCHAR(255) NOT NULL DEFAULT '',
			email      VARCHAR(255) NOT NULL DEFAULT '',
			age        INT NOT NULL DEFAULT 0,
			active     BOOLEAN NOT NULL DEFAULT TRUE,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`
	case "postgres":
		stmt = `CREATE TABLE IF NOT EXISTS bench_models (
			id         BIGSERIAL PRIMARY KEY,
			name       VARCHAR(255) NOT NULL DEFAULT '',
			email      VARCHAR(255) NOT NULL DEFAULT '',
			age        INT NOT NULL DEFAULT 0,
			active     BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`
	case "sqlite":
		stmt = `CREATE TABLE IF NOT EXISTS bench_models (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			name       TEXT NOT NULL DEFAULT '',
			email      TEXT NOT NULL DEFAULT '',
			age        INTEGER NOT NULL DEFAULT 0,
			active     INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`
	case "sqlserver":
		stmt = `CREATE TABLE IF NOT EXISTS bench_models (
			id         BIGINT IDENTITY(1,1) PRIMARY KEY,
			name       NVARCHAR(255) NOT NULL DEFAULT '',
			email      NVARCHAR(255) NOT NULL DEFAULT '',
			age        INT NOT NULL DEFAULT 0,
			active     BIT NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL DEFAULT GETDATE(),
			updated_at DATETIME NOT NULL DEFAULT GETDATE()
		)`
	}

	if _, err := sqlDB.Exec(stmt); err != nil {
		b.Fatalf("createBenchTable (%s): %v", engine, err)
	}

	// Clean any existing data
	_, _ = sqlDB.Exec("DELETE FROM bench_models")
}

// populateBenchTable inserts benchRecordCount rows into the benchmark table.
func populateBenchTable(b *testing.B, db *database.Database) {
	b.Helper()
	q := db.Query().Model(benchModel{})

	records := make([]benchModel, benchRecordCount)
	for i := 0; i < benchRecordCount; i++ {
		records[i] = benchModel{
			Name:   fmt.Sprintf("user_%d", i),
			Email:  fmt.Sprintf("user_%d@example.com", i),
			Age:    20 + (i % 50),
			Active: i%2 == 0,
		}
	}

	if err := q.Create(&records); err != nil {
		b.Fatalf("populateBenchTable: %v", err)
	}
}
