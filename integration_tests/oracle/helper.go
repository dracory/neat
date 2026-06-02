package oracle

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/dracory/neat"
	"github.com/dracory/neat/integration_tests/common"
)

// GetOracleConfig returns an Oracle connection config from environment variables
func GetOracleConfig() neat.DBConfig {
	host := common.GetEnv("ORACLE_HOST", "127.0.0.1")
	port := common.GetEnvInt("ORACLE_PORT", 1521)
	dbName := common.GetEnv("ORACLE_DATABASE", "XE")
	username := common.GetEnv("ORACLE_USER", "system")
	password := common.GetEnv("ORACLE_PASS", "oracle")

	return neat.DBConfig{
		Default: "oracle",
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
		Pool: neat.PoolConfig{
			MaxIdleConns:    5,
			MaxOpenConns:    10,
			ConnMaxLifetime: time.Hour,
			ConnMaxIdleTime: time.Hour,
		},
	}
}

// SetupOracleConnection creates a database connection without setting up tables
func SetupOracleConnection(t *testing.T) *sql.DB {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	host := common.GetEnv("ORACLE_HOST", "127.0.0.1")
	port := common.GetEnvInt("ORACLE_PORT", 1521)
	dbName := common.GetEnv("ORACLE_DATABASE", "XE")
	username := common.GetEnv("ORACLE_USER", "system")
	password := common.GetEnv("ORACLE_PASS", "oracle")

	// Oracle connection string format: oracle://user:pass@host:port/service
	dsn := fmt.Sprintf("oracle://%s:%s@%s:%d/%s",
		username, password, host, port, dbName)

	// Use the local Oracle driver directly
	oracleDriver := NewOracle()
	sqlDB, err := oracleDriver.Open(dsn)
	if err != nil {
		t.Fatalf("Failed to connect to Oracle: %v", err)
	}

	t.Cleanup(func() {
		_ = oracleDriver.Close(sqlDB)
	})

	return sqlDB
}
