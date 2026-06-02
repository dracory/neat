package oracle_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/dracory/neat"
	"github.com/dracory/neat/contracts/log"
	"github.com/dracory/neat/database"
	"github.com/dracory/neat/integration_tests/common"
	_ "github.com/sijms/go-ora/v2"
)

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

// GetOracleConfig returns an Oracle connection config from environment variables
func GetOracleConfig() neat.DBConfig {
	host := common.GetEnv("ORACLE_HOST", "127.0.0.1")
	port := common.GetEnvInt("ORACLE_PORT", 1521)
	database := common.GetEnv("ORACLE_DATABASE", "XE")
	username := common.GetEnv("ORACLE_USER", "system")
	password := common.GetEnv("ORACLE_PASS", "oracle")

	return neat.DBConfig{
		Default: "oracle",
		Connections: map[string]neat.ConnectionConfig{
			"oracle": {
				Driver:   "oracle",
				Host:     host,
				Port:     port,
				Database: database,
				Username: username,
				Password: password,
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

// SetupOracleTest creates a database connection and sets up test tables for Oracle
func SetupOracleTest(t *testing.T) *database.Database {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	host := common.GetEnv("ORACLE_HOST", "127.0.0.1")
	port := common.GetEnvInt("ORACLE_PORT", 1521)
	dbName := common.GetEnv("ORACLE_DATABASE", "XE")
	username := common.GetEnv("ORACLE_USER", "system")
	password := common.GetEnv("ORACLE_PASS", "oracle")

	config := neat.DBConfig{
		Default: "oracle",
		Debug:   true, // Enable debug mode to see actual SQL errors
		Connections: map[string]neat.ConnectionConfig{
			"oracle": {
				Driver:   "oracle",
				Host:     host,
				Port:     port,
				Database: dbName,
				Username: username,
				Password: password,
			},
		},
	}

	db, err := neat.New(config)
	if err != nil {
		t.Fatalf("Failed to connect to Oracle: %v", err)
	}

	// Enable query logging
	db.EnableQueryLog()

	createOracleTestTables(t, db)
	// Clean up any existing data before each test
	cleanupOracleTestData(t, db)

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

// cleanupOracleTestData removes all data from test tables
func cleanupOracleTestData(t *testing.T, db *database.Database) {
	t.Helper()
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("cleanupOracleTestData: DB(): %v", err)
	}
	stmts := []string{
		`BEGIN EXECUTE IMMEDIATE 'TRUNCATE TABLE USERS CASCADE'; EXCEPTION WHEN OTHERS THEN NULL; END;`,
		`BEGIN EXECUTE IMMEDIATE 'TRUNCATE TABLE ADDRESSES CASCADE'; EXCEPTION WHEN OTHERS THEN NULL; END;`,
		`BEGIN EXECUTE IMMEDIATE 'TRUNCATE TABLE BOOKS CASCADE'; EXCEPTION WHEN OTHERS THEN NULL; END;`,
		`BEGIN EXECUTE IMMEDIATE 'TRUNCATE TABLE PEOPLES CASCADE'; EXCEPTION WHEN OTHERS THEN NULL; END;`,
		`BEGIN EXECUTE IMMEDIATE 'TRUNCATE TABLE JSON_DATAS CASCADE'; EXCEPTION WHEN OTHERS THEN NULL; END;`,
		`BEGIN EXECUTE IMMEDIATE 'TRUNCATE TABLE POSTS CASCADE'; EXCEPTION WHEN OTHERS THEN NULL; END;`,
		`BEGIN EXECUTE IMMEDIATE 'TRUNCATE TABLE VIDEOS CASCADE'; EXCEPTION WHEN OTHERS THEN NULL; END;`,
		`BEGIN EXECUTE IMMEDIATE 'TRUNCATE TABLE COMMENTS CASCADE'; EXCEPTION WHEN OTHERS THEN NULL; END;`,
	}
	for _, stmt := range stmts {
		if _, err := sqlDB.Exec(stmt); err != nil {
			// Ignore errors if table doesn't exist
			continue
		}
	}
}

// createOracleTestTables creates all tables required by the integration test models.
func createOracleTestTables(t *testing.T, db *database.Database) {
	t.Helper()
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("createOracleTestTables: DB(): %v", err)
	}

	// Check if USERS table already exists
	var tableExists int
	if err := sqlDB.QueryRow("SELECT COUNT(*) FROM ALL_TABLES WHERE TABLE_NAME = 'USERS'").Scan(&tableExists); err == nil && tableExists > 0 {
		// Tables already exist, just cleanup data
		cleanupOracleTestData(t, db)
		return
	}

	stmts := []string{
		`BEGIN EXECUTE IMMEDIATE 'DROP TABLE BOOKS CASCADE CONSTRAINTS'; EXCEPTION WHEN OTHERS THEN NULL; END;`,
		`BEGIN EXECUTE IMMEDIATE 'DROP TABLE ADDRESSES CASCADE CONSTRAINTS'; EXCEPTION WHEN OTHERS THEN NULL; END;`,
		`BEGIN EXECUTE IMMEDIATE 'DROP TABLE USERS CASCADE CONSTRAINTS'; EXCEPTION WHEN OTHERS THEN NULL; END;`,
		`BEGIN EXECUTE IMMEDIATE 'DROP TABLE PEOPLES CASCADE CONSTRAINTS'; EXCEPTION WHEN OTHERS THEN NULL; END;`,
		`BEGIN EXECUTE IMMEDIATE 'DROP TABLE JSON_DATAS CASCADE CONSTRAINTS'; EXCEPTION WHEN OTHERS THEN NULL; END;`,
		`BEGIN EXECUTE IMMEDIATE 'DROP TABLE POSTS CASCADE CONSTRAINTS'; EXCEPTION WHEN OTHERS THEN NULL; END;`,
		`BEGIN EXECUTE IMMEDIATE 'DROP TABLE VIDEOS CASCADE CONSTRAINTS'; EXCEPTION WHEN OTHERS THEN NULL; END;`,
		`BEGIN EXECUTE IMMEDIATE 'DROP TABLE COMMENTS CASCADE CONSTRAINTS'; EXCEPTION WHEN OTHERS THEN NULL; END;`,
		`BEGIN EXECUTE IMMEDIATE 'DROP SEQUENCE BOOKS_SEQ'; EXCEPTION WHEN OTHERS THEN NULL; END;`,
		`BEGIN EXECUTE IMMEDIATE 'DROP SEQUENCE ADDRESSES_SEQ'; EXCEPTION WHEN OTHERS THEN NULL; END;`,
		`BEGIN EXECUTE IMMEDIATE 'DROP SEQUENCE USERS_SEQ'; EXCEPTION WHEN OTHERS THEN NULL; END;`,
		`BEGIN EXECUTE IMMEDIATE 'DROP SEQUENCE PEOPLES_SEQ'; EXCEPTION WHEN OTHERS THEN NULL; END;`,
		`BEGIN EXECUTE IMMEDIATE 'DROP SEQUENCE JSON_DATAS_SEQ'; EXCEPTION WHEN OTHERS THEN NULL; END;`,
		`BEGIN EXECUTE IMMEDIATE 'DROP SEQUENCE POSTS_SEQ'; EXCEPTION WHEN OTHERS THEN NULL; END;`,
		`BEGIN EXECUTE IMMEDIATE 'DROP SEQUENCE VIDEOS_SEQ'; EXCEPTION WHEN OTHERS THEN NULL; END;`,
		`BEGIN EXECUTE IMMEDIATE 'DROP SEQUENCE COMMENTS_SEQ'; EXCEPTION WHEN OTHERS THEN NULL; END;`,
		`CREATE TABLE USERS (
			ID         NUMBER(20) GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
			NAME       VARCHAR2(255) DEFAULT '',
			AVATAR     VARCHAR2(255) DEFAULT '',
			BIO        CLOB,
			VOTES      NUMBER(10) DEFAULT 0,
			DELETED_AT TIMESTAMP,
			CREATED_AT TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UPDATED_AT TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE ADDRESSES (
			ID         NUMBER(20) GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
			NAME       VARCHAR2(255) DEFAULT '',
			USER_ID    NUMBER(20),
			CREATED_AT TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UPDATED_AT TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE BOOKS (
			ID         NUMBER(20) GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
			NAME       VARCHAR2(255) DEFAULT '',
			USER_ID    NUMBER(20),
			CREATED_AT TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UPDATED_AT TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE PEOPLES (
			ID         NUMBER(20) GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
			BODY       CLOB NOT NULL,
			CREATED_AT TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UPDATED_AT TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE JSON_DATAS (
			ID         NUMBER(20) GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
			DATA       CLOB NOT NULL,
			CREATED_AT TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UPDATED_AT TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
	}
	for _, stmt := range stmts {
		if _, err := sqlDB.Exec(stmt); err != nil {
			t.Fatalf("createOracleTestTables: %v", err)
		}
	}
}

// SetupOracleConnection creates a database connection without setting up tables
func SetupOracleConnection(t *testing.T) *database.Database {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	host := common.GetEnv("ORACLE_HOST", "127.0.0.1")
	port := common.GetEnvInt("ORACLE_PORT", 1521)
	database := common.GetEnv("ORACLE_DATABASE", "XE")
	username := common.GetEnv("ORACLE_USER", "system")
	password := common.GetEnv("ORACLE_PASS", "oracle")
	dsn := fmt.Sprintf("oracle://%s:%s@%s:%d/%s",
		username, password, host, port, database)

	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		t.Fatalf("Failed to connect to Oracle: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}
