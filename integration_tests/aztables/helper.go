//go:build integration

package aztables

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/dracory/aztablessql"
	"github.com/dracory/neat"
	contractsdb "github.com/dracory/neat/contracts/database"
	"github.com/dracory/neat/contracts/log"
	"github.com/dracory/neat/database"
)

// azuriteConnStr is the well-known Azurite dev-store connection string,
// overridden by AZTABLES_TEST_CONNSTR if set.
const azuriteConnStr = "DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;TableEndpoint=http://127.0.0.1:10002/devstoreaccount1;"

// testConnStr returns the Azure Table Storage connection string from the
// AZTABLES_TEST_CONNSTR env var, falling back to the Azurite dev-store default.
func testConnStr() string {
	if v := os.Getenv("AZTABLES_TEST_CONNSTR"); v != "" {
		return v
	}
	return azuriteConnStr
}

// GetAztablesConfig returns a neat.DBConfig for Azure Table Storage via
// the aztablessql driver, pointing at Azurite (or AZTABLES_TEST_CONNSTR).
func GetAztablesConfig() neat.DBConfig {
	return neat.DBConfig{
		Default: "aztables",
		Connections: map[string]neat.ConnectionConfig{
			"aztables": {
				Driver: contractsdb.DriverAztables,
				Dsn:    testConnStr(),
			},
		},
		Pool: neat.PoolConfig{
			MaxIdleConns:    5,
			MaxOpenConns:    10,
			ConnMaxLifetime: time.Hour,
			ConnMaxIdleTime: time.Hour,
		},
	}
}

// SetupAztablesConnection creates a database connection to Azurite without
// setting up tables. The caller is responsible for table lifecycle.
func SetupAztablesConnection(t *testing.T) *database.Database {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := GetAztablesConfig()
	db, err := neat.New(config, database.WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to connect to Azure Table Storage (Azurite): %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

// uniqueTableName returns a unique table name per test invocation to avoid
// cross-test interference. Azure Table Storage table names must start with
// a letter and be 3-63 chars, alphanumeric only.
func uniqueTableName(prefix string) string {
	return fmt.Sprintf("%s%d", prefix, time.Now().UnixNano()%1000000)
}

// createTable creates a Table Storage table via raw SQL, tolerating the
// "already exists" case.
func createTable(t *testing.T, db *database.Database, table string) {
	t.Helper()
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("createTable: DB(): %v", err)
	}
	_, err = sqlDB.Exec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s", table))
	if err != nil {
		t.Fatalf("createTable %s: %v", table, err)
	}
}

// dropTable drops a Table Storage table via raw SQL, tolerating the
// "not found" case.
func dropTable(t *testing.T, db *database.Database, table string) {
	t.Helper()
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("dropTable: DB(): %v", err)
	}
	_, _ = sqlDB.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table))
}

// insertEntity inserts a single entity with the given partition/row key and
// properties. props is a map of column-name → value (excluding PartitionKey
// and RowKey, which are passed separately).
func insertEntity(t *testing.T, db *database.Database, table, partition, rowKey string, props map[string]any) {
	t.Helper()
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("insertEntity: DB(): %v", err)
	}

	cols := []string{"PartitionKey", "RowKey"}
	vals := []any{partition, rowKey}
	for col, val := range props {
		cols = append(cols, col)
		vals = append(vals, val)
	}

	placeholders := make([]string, len(vals))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	_, err = sqlDB.Exec(
		fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			table, strings.Join(cols, ", "), strings.Join(placeholders, ", ")),
		vals...,
	)
	if err != nil {
		t.Fatalf("insertEntity %s: %v", table, err)
	}
}

// scanOneRow reads exactly one row from rows and returns it as a column→value
// map. It closes rows and fails the test if there is not exactly one row.
func scanOneRow(t *testing.T, rows *sql.Rows) map[string]any {
	t.Helper()
	defer rows.Close()
	cols, _ := rows.Columns()
	if !rows.Next() {
		t.Fatal("expected 1 row, got 0")
	}
	vals := make([]any, len(cols))
	ptrs := make([]any, len(cols))
	for i := range vals {
		ptrs[i] = &vals[i]
	}
	if err := rows.Scan(ptrs...); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	m := make(map[string]any, len(cols))
	for i, c := range cols {
		m[c] = vals[i]
	}
	if rows.Next() {
		t.Fatal("expected 1 row, got >1")
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows: %v", err)
	}
	return m
}

// collectRows reads all rows from rows and returns them as a slice of
// column→value maps. It closes rows.
func collectRows(t *testing.T, rows *sql.Rows) []map[string]any {
	t.Helper()
	defer rows.Close()
	cols, _ := rows.Columns()
	var result []map[string]any
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		m := make(map[string]any, len(cols))
		for i, c := range cols {
			m[c] = vals[i]
		}
		result = append(result, m)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows: %v", err)
	}
	return result
}
