package tidb_test

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

// TiDB-specific behavior notes
// --------------------------------
// TiDB is a distributed SQL database with very high MySQL compatibility.
// It speaks the MySQL wire protocol and accepts the go-sql-driver/mysql
// driver, so Neat's MySQL driver works with TiDB out of the box. The
// differences below are why these integration tests exist:
//
//  1. Savepoints:   TiDB does not support the SAVEPOINT syntax. Neat's
//                   transaction layer emulates savepoints, which the
//                   transaction tests verify.
//  2. Foreign keys: TiDB v6.6+ supports foreign keys, but enforcement
//                   behavior may differ from MySQL.
//  3. Auto-increment: TiDB's AUTO_INCREMENT is not guaranteed to be
//                   sequential (distributed allocation). Tests must not
//                   assert on sequential IDs, only on uniqueness and
//                   non-zero values.
//  4. JSON:         TiDB supports a subset of MySQL's JSON functions.
//                   JSON tests focus on basic store/retrieve/query.
//  5. Collation:    TiDB defaults to utf8mb4_bin (case-sensitive), while
//                   MySQL defaults to utf8mb4_0900_ai_ci (case-insensitive).
//                   Tests relying on case-insensitive comparisons may need
//                   adjustment.

// TestModel is a simple model for integration testing
type TestModel struct {
	ID        uint      `db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	Age       int       `db:"age"`
	Active    bool      `db:"active"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (TestModel) TableName() string {
	return "test_models"
}

// GetTiDBConfig returns a TiDB connection config from environment variables.
// TiDB speaks the MySQL wire protocol, so we use the MySQL driver.
func GetTiDBConfig() neat.DBConfig {
	host := common.GetEnv("TIDB_HOST", "127.0.0.1")
	port := common.GetEnvInt("TIDB_PORT", 4000)
	database := common.GetEnv("TIDB_DATABASE", "test")
	username := common.GetEnv("TIDB_USER", "root")
	password := common.GetEnv("TIDB_PASS", "")

	return neat.DBConfig{
		Default: "tidb",
		Connections: map[string]neat.ConnectionConfig{
			"tidb": {
				Driver:   "mysql", // TiDB uses the MySQL driver
				Host:     host,
				Port:     port,
				Database: database,
				Username: username,
				Password: password,
				Charset:  "utf8mb4",
				Loc:      "Local",
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

// SetupTestDB creates a database connection and sets up test tables
func SetupTestDB(config neat.DBConfig) (*database.Database, error) {
	db, err := neat.New(config, database.WithLogger(log.NewNoopLogger()))
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	// Note: Schema builder setup is skipped for now - requires proper blueprint configuration
	// Tests should handle their own table setup as needed

	return db, nil
}

// TeardownTestDB drops test tables and closes the connection
func TeardownTestDB(db *database.Database) error {
	// Drop test table
	if db != nil {
		if err := db.Schema().Drop("test_models"); err != nil {
			return err
		}
		return db.Close()
	}
	return nil
}

// SetupTiDBTest creates a database connection and sets up test tables for TiDB
func SetupTiDBTest(t *testing.T) *database.Database {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	host := common.GetEnv("TIDB_HOST", "127.0.0.1")
	port := common.GetEnvInt("TIDB_PORT", 4000)
	dbName := common.GetEnv("TIDB_DATABASE", "test")
	username := common.GetEnv("TIDB_USER", "root")
	password := common.GetEnv("TIDB_PASS", "")
	dsn := fmt.Sprintf("mysql://%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		username, password, host, port, dbName)

	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		t.Fatalf("Failed to connect to TiDB: %v", err)
	}

	createTiDBTestTables(t, db)
	// Clean up any existing data before each test
	cleanupTiDBTestData(t, db)

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

// cleanupTiDBTestData removes all data from test tables
func cleanupTiDBTestData(t *testing.T, db *database.Database) {
	t.Helper()
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("cleanupTiDBTestData: DB(): %v", err)
	}
	stmts := []string{
		`DELETE IGNORE FROM users`,
		`DELETE IGNORE FROM addresses`,
		`DELETE IGNORE FROM books`,
		`DELETE IGNORE FROM peoples`,
		`DELETE IGNORE FROM json_datas`,
	}
	for _, stmt := range stmts {
		if _, err := sqlDB.Exec(stmt); err != nil {
			// Ignore errors if table doesn't exist
			continue
		}
	}
}

// createTiDBTestTables creates all tables required by the integration test models.
// The SQL is identical to the MySQL helper's SQL, since TiDB is MySQL-compatible.
func createTiDBTestTables(t *testing.T, db *database.Database) {
	t.Helper()
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("createTiDBTestTables: DB(): %v", err)
	}
	stmts := []string{
		`DROP TABLE IF EXISTS books`,
		`DROP TABLE IF EXISTS addresses`,
		`DROP TABLE IF EXISTS users`,
		`DROP TABLE IF EXISTS peoples`,
		`DROP TABLE IF EXISTS json_datas`,
		`CREATE TABLE IF NOT EXISTS users (
			id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
			name       VARCHAR(255) NOT NULL DEFAULT '',
			avatar     VARCHAR(255) NOT NULL DEFAULT '',
			bio        TEXT,
			votes      INT NOT NULL DEFAULT 0,
			soft_deleted_at DATETIME,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS addresses (
			id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
			name       VARCHAR(255) NOT NULL DEFAULT '',
			user_id    BIGINT UNSIGNED,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS books (
			id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
			name       VARCHAR(255) NOT NULL DEFAULT '',
			user_id    BIGINT UNSIGNED,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS peoples (
			id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
			body       TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS json_datas (
			id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
			data       JSON NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}
	for _, stmt := range stmts {
		if _, err := sqlDB.Exec(stmt); err != nil {
			t.Fatalf("createTiDBTestTables: %v", err)
		}
	}
}

// SetupTiDBConnection creates a database connection without setting up tables
func SetupTiDBConnection(t *testing.T) *database.Database {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	host := common.GetEnv("TIDB_HOST", "127.0.0.1")
	port := common.GetEnvInt("TIDB_PORT", 4000)
	database := common.GetEnv("TIDB_DATABASE", "test")
	username := common.GetEnv("TIDB_USER", "root")
	password := common.GetEnv("TIDB_PASS", "")
	dsn := fmt.Sprintf("mysql://%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		username, password, host, port, database)

	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		t.Fatalf("Failed to connect to TiDB: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}
