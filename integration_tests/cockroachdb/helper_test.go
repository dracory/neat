//go:build integration

package cockroachdb_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/dracory/neat"
	"github.com/dracory/neat/contracts/log"
	"github.com/dracory/neat/database"
	"github.com/dracory/neat/integration_tests/common"
	_ "github.com/lib/pq"
)

// CockroachDB-specific behavior notes
// --------------------------------
// CockroachDB is a distributed SQL database with PostgreSQL wire-protocol
// compatibility. It accepts the lib/pq driver, so Neat's postgres driver
// works with CockroachDB out of the box. The differences below are why
// these integration tests exist:
//
//  1. Savepoints:   CockroachDB supports SAVEPOINT but with different
//                   semantics than PostgreSQL. Nested transaction behavior
//                   may differ.
//  2. Auto-increment: CockroachDB's SERIAL/BIGSERIAL uses gen_random_uuid()
//                   or ordered sequences depending on version. IDs are not
//                   guaranteed to be sequential. Tests must not assert on
//                   sequential IDs, only on uniqueness and non-zero values.
//  3. JSONB:        CockroachDB supports JSONB but with a subset of
//                   PostgreSQL's JSON functions. JSON tests focus on basic
//                   store/retrieve/query.
//  4. Locking:      CockroachDB supports SELECT FOR UPDATE but locking
//                   semantics differ from PostgreSQL due to distributed
//                   architecture.
//  5. Indexing:     CockroachDB automatically creates indexes for primary
//                   keys and unique constraints. Index naming may differ.
//  6. Schema changes: CockroachDB schema changes are online but may have
//                   different transactional behavior than PostgreSQL.

// GetCockroachDBConfig returns a CockroachDB connection config from environment variables.
// CockroachDB speaks the PostgreSQL wire protocol, so we use the postgres driver.
func GetCockroachDBConfig() neat.DBConfig {
	host := common.GetEnv("COCKROACHDB_HOST", "127.0.0.1")
	port := common.GetEnvInt("COCKROACHDB_PORT", 26257)
	database := common.GetEnv("COCKROACHDB_DATABASE", "test")
	username := common.GetEnv("COCKROACHDB_USER", "root")
	password := common.GetEnv("COCKROACHDB_PASS", "")

	sslmode := common.GetEnv("COCKROACHDB_SSLMODE", "disable")

	return neat.DBConfig{
		Default: "cockroachdb",
		Connections: map[string]neat.ConnectionConfig{
			"cockroachdb": {
				Driver:   "postgres", // CockroachDB uses the postgres driver
				Host:     host,
				Port:     port,
				Database: database,
				Username: username,
				Password: password,
				SSLMode:  sslmode,
				Schema:   "public",
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

// SetupCockroachDBTest creates a database connection and sets up test tables for CockroachDB
func SetupCockroachDBTest(t *testing.T) *database.Database {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	host := common.GetEnv("COCKROACHDB_HOST", "127.0.0.1")
	port := common.GetEnvInt("COCKROACHDB_PORT", 26257)
	dbName := common.GetEnv("COCKROACHDB_DATABASE", "test")
	username := common.GetEnv("COCKROACHDB_USER", "root")
	password := common.GetEnv("COCKROACHDB_PASS", "")

	var dsn string
	if password != "" {
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			username, password, host, port, dbName)
	} else {
		dsn = fmt.Sprintf("postgres://%s@%s:%d/%s?sslmode=disable",
			username, host, port, dbName)
	}

	db, err := neat.NewFromDSN(dsn, database.WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to connect to CockroachDB: %v", err)
	}

	createCockroachDBTestTables(t, db)
	cleanupCockroachDBTestData(t, db)

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

// cleanupCockroachDBTestData removes all data from test tables
func cleanupCockroachDBTestData(t *testing.T, db *database.Database) {
	t.Helper()
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("cleanupCockroachDBTestData: DB(): %v", err)
	}
	stmts := []string{
		`DELETE FROM users`,
		`DELETE FROM addresses`,
		`DELETE FROM books`,
		`DELETE FROM peoples`,
		`DELETE FROM json_datas`,
		`DELETE FROM bigserial_users`,
		`DELETE FROM comments`,
		`DELETE FROM posts`,
		`DELETE FROM videos`,
	}
	for _, stmt := range stmts {
		if _, err := sqlDB.Exec(stmt); err != nil {
			continue
		}
	}
}

// createCockroachDBTestTables creates all tables required by the integration test models.
// The SQL is PostgreSQL-compatible, since CockroachDB speaks the PostgreSQL protocol.
func createCockroachDBTestTables(t *testing.T, db *database.Database) {
	t.Helper()
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("createCockroachDBTestTables: DB(): %v", err)
	}
	stmts := []string{
		`DROP TABLE IF EXISTS comments`,
		`DROP TABLE IF EXISTS videos`,
		`DROP TABLE IF EXISTS posts`,
		`DROP TABLE IF EXISTS books`,
		`DROP TABLE IF EXISTS addresses`,
		`DROP TABLE IF EXISTS users`,
		`DROP TABLE IF EXISTS peoples`,
		`DROP TABLE IF EXISTS json_datas`,
		`DROP TABLE IF EXISTS bigserial_users`,
		`CREATE TABLE IF NOT EXISTS users (
			id         BIGSERIAL PRIMARY KEY,
			name       VARCHAR(255) NOT NULL DEFAULT '',
			avatar     VARCHAR(255) NOT NULL DEFAULT '',
			bio        TEXT,
			votes      INT NOT NULL DEFAULT 0,
			soft_deleted_at TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS addresses (
			id         BIGSERIAL PRIMARY KEY,
			name       VARCHAR(255) NOT NULL DEFAULT '',
			user_id    BIGINT,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS books (
			id         BIGSERIAL PRIMARY KEY,
			name       VARCHAR(255) NOT NULL DEFAULT '',
			user_id    BIGINT,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS peoples (
			id         BIGSERIAL PRIMARY KEY,
			body       TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS json_datas (
			id         BIGSERIAL PRIMARY KEY,
			data       JSONB NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS bigserial_users (
			id         BIGSERIAL PRIMARY KEY,
			name       VARCHAR(255) NOT NULL DEFAULT '',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS posts (
			id         BIGSERIAL PRIMARY KEY,
			title      VARCHAR(255) NOT NULL DEFAULT '',
			content    TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS videos (
			id         BIGSERIAL PRIMARY KEY,
			title      VARCHAR(255) NOT NULL DEFAULT '',
			url        VARCHAR(255) NOT NULL DEFAULT '',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS comments (
			id              BIGSERIAL PRIMARY KEY,
			body            TEXT NOT NULL DEFAULT '',
			commentable_id  BIGINT NOT NULL DEFAULT 0,
			commentable_type VARCHAR(255) NOT NULL DEFAULT '',
			created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
	}
	for _, stmt := range stmts {
		if _, err := sqlDB.Exec(stmt); err != nil {
			t.Fatalf("createCockroachDBTestTables: %v (stmt: %s)", err, stmt)
		}
	}
}

// SetupCockroachDBConnection creates a database connection without setting up tables
func SetupCockroachDBConnection(t *testing.T) *database.Database {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	host := common.GetEnv("COCKROACHDB_HOST", "127.0.0.1")
	port := common.GetEnvInt("COCKROACHDB_PORT", 26257)
	dbName := common.GetEnv("COCKROACHDB_DATABASE", "test")
	username := common.GetEnv("COCKROACHDB_USER", "root")
	password := common.GetEnv("COCKROACHDB_PASS", "")

	var dsn string
	if password != "" {
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			username, password, host, port, dbName)
	} else {
		dsn = fmt.Sprintf("postgres://%s@%s:%d/%s?sslmode=disable",
			username, host, port, dbName)
	}

	db, err := neat.NewFromDSN(dsn, database.WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to connect to CockroachDB: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}
