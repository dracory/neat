package database

import (
	"context"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/dracory/neat/contracts/log"
	"github.com/dracory/neat/database/db"
)

// createTempDB creates a temporary SQLite database file and returns its path
func createTempDB(t *testing.T) string {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	return dbPath
}

func TestPool_MaxOpenConns(t *testing.T) {
	tests := []struct {
		name         string
		maxOpenConns int
	}{
		{
			name:         "MaxOpenConns set to 5",
			maxOpenConns: 5,
		},
		{
			name:         "MaxOpenConns set to 10",
			maxOpenConns: 10,
		},
		{
			name:         "MaxOpenConns set to 1",
			maxOpenConns: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbPath := createTempDB(t)

			poolConfig := db.PoolConfig{
				MaxOpenConns:    tt.maxOpenConns,
				MaxIdleConns:    2,
				ConnMaxLifetime: 3600,
				ConnMaxIdleTime: 3600,
			}

			config := db.DBConfig{
				Default: "default",
				Connections: map[string]db.ConnectionConfig{
					"default": {
						Driver:   "sqlite",
						Database: dbPath,
					},
				},
				Pool: poolConfig,
			}

			db, err := New(config, WithLogger(log.NewNoopLogger()))
			if err != nil {
				t.Fatalf("Failed to create database: %v", err)
			}
			defer db.Close()

			sqlDB, err := db.DB()
			if err != nil {
				t.Fatalf("Failed to get sql.DB: %v", err)
			}

			stats := sqlDB.Stats()
			if stats.MaxOpenConnections != tt.maxOpenConns {
				t.Errorf("Expected MaxOpenConnections to be %d, got %d", tt.maxOpenConns, stats.MaxOpenConnections)
			}
		})
	}
}

func TestPool_MaxIdleConns(t *testing.T) {
	tests := []struct {
		name         string
		maxIdleConns int
	}{
		{
			name:         "MaxIdleConns set to 2",
			maxIdleConns: 2,
		},
		{
			name:         "MaxIdleConns set to 5",
			maxIdleConns: 5,
		},
		{
			name:         "MaxIdleConns set to 0",
			maxIdleConns: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbPath := createTempDB(t)

			poolConfig := db.PoolConfig{
				MaxOpenConns:    10,
				MaxIdleConns:    tt.maxIdleConns,
				ConnMaxLifetime: 3600,
				ConnMaxIdleTime: 3600,
			}

			config := db.DBConfig{
				Default: "default",
				Connections: map[string]db.ConnectionConfig{
					"default": {
						Driver:   "sqlite",
						Database: dbPath,
					},
				},
				Pool: poolConfig,
			}

			db, err := New(config, WithLogger(log.NewNoopLogger()))
			if err != nil {
				t.Fatalf("Failed to create database: %v", err)
			}
			defer db.Close()

			sqlDB, err := db.DB()
			if err != nil {
				t.Fatalf("Failed to get sql.DB: %v", err)
			}

			stats := sqlDB.Stats()
			if stats.Idle != tt.maxIdleConns {
				// Idle may be 0 initially, so we need to open some connections
				// Open a connection to populate the pool
				conn, err := sqlDB.Conn(context.Background())
				if err != nil {
					t.Fatalf("Failed to get connection: %v", err)
				}
				conn.Close()

				stats = sqlDB.Stats()
				// After opening a connection, idle should be at most maxIdleConns
				if stats.Idle > tt.maxIdleConns {
					t.Errorf("Expected Idle connections to be at most %d, got %d", tt.maxIdleConns, stats.Idle)
				}
			}
		})
	}
}

func TestPool_ConnMaxLifetime(t *testing.T) {
	tests := []struct {
		name            string
		connMaxLifetime int
	}{
		{
			name:            "ConnMaxLifetime set to 1 second",
			connMaxLifetime: 1,
		},
		{
			name:            "ConnMaxLifetime set to 5 seconds",
			connMaxLifetime: 5,
		},
		{
			name:            "ConnMaxLifetime set to 0 (no limit)",
			connMaxLifetime: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbPath := createTempDB(t)

			poolConfig := db.PoolConfig{
				MaxOpenConns:    10,
				MaxIdleConns:    2,
				ConnMaxLifetime: tt.connMaxLifetime,
				ConnMaxIdleTime: 3600,
			}

			config := db.DBConfig{
				Default: "default",
				Connections: map[string]db.ConnectionConfig{
					"default": {
						Driver:   "sqlite",
						Database: dbPath,
					},
				},
				Pool: poolConfig,
			}

			db, err := New(config, WithLogger(log.NewNoopLogger()))
			if err != nil {
				t.Fatalf("Failed to create database: %v", err)
			}
			defer db.Close()

			sqlDB, err := db.DB()
			if err != nil {
				t.Fatalf("Failed to get sql.DB: %v", err)
			}

			// Open a connection to verify the pool is configured
			conn, err := sqlDB.Conn(context.Background())
			if err != nil {
				t.Fatalf("Failed to get connection: %v", err)
			}
			conn.Close()

			// The configuration is set, but we can't easily test the actual lifetime behavior
			// without waiting for the duration. We verify the configuration was accepted.
			stats := sqlDB.Stats()
			if stats.MaxOpenConnections != 10 {
				t.Errorf("Expected pool to be configured with MaxOpenConns 10, got %d", stats.MaxOpenConnections)
			}
		})
	}
}

func TestPool_ConnMaxIdleTime(t *testing.T) {
	tests := []struct {
		name            string
		connMaxIdleTime int
	}{
		{
			name:            "ConnMaxIdleTime set to 1 second",
			connMaxIdleTime: 1,
		},
		{
			name:            "ConnMaxIdleTime set to 5 seconds",
			connMaxIdleTime: 5,
		},
		{
			name:            "ConnMaxIdleTime set to 0 (no limit)",
			connMaxIdleTime: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbPath := createTempDB(t)

			poolConfig := db.PoolConfig{
				MaxOpenConns:    10,
				MaxIdleConns:    2,
				ConnMaxLifetime: 3600,
				ConnMaxIdleTime: tt.connMaxIdleTime,
			}

			config := db.DBConfig{
				Default: "default",
				Connections: map[string]db.ConnectionConfig{
					"default": {
						Driver:   "sqlite",
						Database: dbPath,
					},
				},
				Pool: poolConfig,
			}

			db, err := New(config, WithLogger(log.NewNoopLogger()))
			if err != nil {
				t.Fatalf("Failed to create database: %v", err)
			}
			defer db.Close()

			sqlDB, err := db.DB()
			if err != nil {
				t.Fatalf("Failed to get sql.DB: %v", err)
			}

			// Open a connection to verify the pool is configured
			conn, err := sqlDB.Conn(context.Background())
			if err != nil {
				t.Fatalf("Failed to get connection: %v", err)
			}
			conn.Close()

			// The configuration is set, but we can't easily test the actual idle time behavior
			// without waiting for the duration. We verify the configuration was accepted.
			stats := sqlDB.Stats()
			if stats.MaxOpenConnections != 10 {
				t.Errorf("Expected pool to be configured with MaxOpenConns 10, got %d", stats.MaxOpenConnections)
			}
		})
	}
}

func TestPool_WithPoolOption(t *testing.T) {
	dbPath := createTempDB(t)

	poolConfig := db.PoolConfig{
		MaxOpenConns:    15,
		MaxIdleConns:    5,
		ConnMaxLifetime: 1800,
		ConnMaxIdleTime: 900,
	}

	config := db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: dbPath,
			},
		},
	}

	db, err := New(config, WithPool(poolConfig), WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}

	// Open a connection to verify the pool is configured
	conn, err := sqlDB.Conn(context.Background())
	if err != nil {
		t.Fatalf("Failed to get connection: %v", err)
	}
	conn.Close()

	stats := sqlDB.Stats()
	if stats.MaxOpenConnections != 15 {
		t.Errorf("Expected MaxOpenConnections to be 15, got %d", stats.MaxOpenConnections)
	}
}

func TestPool_DefaultConfiguration(t *testing.T) {
	dbPath := createTempDB(t)

	// Test that when no pool config is provided, defaults are used
	config := db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: dbPath,
			},
		},
	}

	db, err := New(config, WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}

	// Open a connection to verify the pool is configured
	conn, err := sqlDB.Conn(context.Background())
	if err != nil {
		t.Fatalf("Failed to get connection: %v", err)
	}
	conn.Close()

	stats := sqlDB.Stats()
	t.Logf("Pool stats with no config: MaxOpen=%d, Open=%d, InUse=%d, Idle=%d",
		stats.MaxOpenConnections, stats.OpenConnections, stats.InUse, stats.Idle)
	// When no pool config is provided, the ORM uses zero values which means unlimited
	// SetMaxOpenConns(0) means unlimited connections
	if stats.MaxOpenConnections != 0 {
		t.Errorf("Expected MaxOpenConnections to be 0 (unlimited), got %d", stats.MaxOpenConnections)
	}
}

func TestPool_NewFromDSN_DefaultPool(t *testing.T) {
	// Test that NewFromDSN applies default pool configuration
	dbPath := createTempDB(t)
	db, err := NewFromDSN("sqlite://"+dbPath, WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create database from DSN: %v", err)
	}
	defer db.Close()

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}

	stats := sqlDB.Stats()
	// Default pool config from NewFromDSN: MaxOpenConns=100, MaxIdleConns=10
	if stats.MaxOpenConnections != 100 {
		t.Errorf("Expected MaxOpenConnections to be 100 (default), got %d", stats.MaxOpenConnections)
	}
}

func TestPool_NewFromDSN_CustomPool(t *testing.T) {
	// Test that NewFromDSN accepts custom pool configuration
	dbPath := createTempDB(t)
	poolConfig := db.PoolConfig{
		MaxOpenConns:    25,
		MaxIdleConns:    8,
		ConnMaxLifetime: 7200,
		ConnMaxIdleTime: 1800,
	}

	db, err := NewFromDSN("sqlite://"+dbPath, WithPool(poolConfig), WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create database from DSN: %v", err)
	}
	defer db.Close()

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}

	stats := sqlDB.Stats()
	if stats.MaxOpenConnections != 25 {
		t.Errorf("Expected MaxOpenConnections to be 25, got %d", stats.MaxOpenConnections)
	}
}

func TestPool_ExhaustionBehavior(t *testing.T) {
	// Test pool exhaustion with a very small pool
	dbPath := createTempDB(t)
	poolConfig := db.PoolConfig{
		MaxOpenConns:    2,
		MaxIdleConns:    1,
		ConnMaxLifetime: 3600,
		ConnMaxIdleTime: 3600,
	}

	config := db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: dbPath,
			},
		},
		Pool: poolConfig,
	}

	db, err := New(config, WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}

	// Simulate pool exhaustion by holding connections
	var wg sync.WaitGroup
	errors := make(chan error, 5)

	// Try to acquire more connections than the pool allows
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			conn, err := sqlDB.Conn(ctx)
			if err != nil {
				errors <- err
				return
			}
			defer conn.Close()

			// Hold the connection briefly
			time.Sleep(100 * time.Millisecond)
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check if any connections failed due to pool exhaustion
	errorCount := 0
	for err := range errors {
		errorCount++
		t.Logf("Connection attempt failed: %v", err)
	}

	// With a small pool and concurrent access, some requests may timeout
	// This is expected behavior for pool exhaustion
	stats := sqlDB.Stats()
	t.Logf("Pool stats after exhaustion test: Open=%d, InUse=%d, Idle=%d, MaxOpen=%d",
		stats.OpenConnections, stats.InUse, stats.Idle, stats.MaxOpenConnections)

	// Verify pool was actually limited
	if stats.MaxOpenConnections != 2 {
		t.Errorf("Expected MaxOpenConnections to be 2, got %d", stats.MaxOpenConnections)
	}
}

func TestPool_Stats(t *testing.T) {
	dbPath := createTempDB(t)
	poolConfig := db.PoolConfig{
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 3600,
		ConnMaxIdleTime: 3600,
	}

	config := db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: dbPath,
			},
		},
		Pool: poolConfig,
	}

	db, err := New(config, WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}

	// Get initial stats
	stats := sqlDB.Stats()
	t.Logf("Initial pool stats: Open=%d, InUse=%d, Idle=%d, MaxOpen=%d",
		stats.OpenConnections, stats.InUse, stats.Idle, stats.MaxOpenConnections)

	// Open a connection
	conn, err := sqlDB.Conn(context.Background())
	if err != nil {
		t.Fatalf("Failed to get connection: %v", err)
	}

	// Check stats with connection in use
	stats = sqlDB.Stats()
	if stats.InUse < 1 {
		t.Errorf("Expected at least 1 connection in use, got %d", stats.InUse)
	}

	// Close the connection
	conn.Close()

	// Check stats after closing
	stats = sqlDB.Stats()
	t.Logf("Pool stats after close: Open=%d, InUse=%d, Idle=%d, MaxOpen=%d",
		stats.OpenConnections, stats.InUse, stats.Idle, stats.MaxOpenConnections)
}

func TestPool_Ping(t *testing.T) {
	dbPath := createTempDB(t)
	poolConfig := db.PoolConfig{
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: 3600,
		ConnMaxIdleTime: 3600,
	}

	config := db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: dbPath,
			},
		},
		Pool: poolConfig,
	}

	db, err := New(config, WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}

	// Ping the database to verify pool is working
	err = sqlDB.Ping()
	if err != nil {
		t.Errorf("Failed to ping database: %v", err)
	}
}

func TestPool_SetMaxOpenConns(t *testing.T) {
	// Test that we can dynamically change MaxOpenConns
	dbPath := createTempDB(t)
	config := db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: dbPath,
			},
		},
		Pool: db.PoolConfig{
			MaxOpenConns:    5,
			MaxIdleConns:    2,
			ConnMaxLifetime: 3600,
			ConnMaxIdleTime: 3600,
		},
	}

	db, err := New(config, WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}

	// Change MaxOpenConns dynamically
	sqlDB.SetMaxOpenConns(10)

	stats := sqlDB.Stats()
	if stats.MaxOpenConnections != 10 {
		t.Errorf("Expected MaxOpenConnections to be 10 after SetMaxOpenConns, got %d", stats.MaxOpenConnections)
	}
}

func TestPool_SetMaxIdleConns(t *testing.T) {
	// Test that we can dynamically change MaxIdleConns
	dbPath := createTempDB(t)
	config := db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: dbPath,
			},
		},
		Pool: db.PoolConfig{
			MaxOpenConns:    10,
			MaxIdleConns:    2,
			ConnMaxLifetime: 3600,
			ConnMaxIdleTime: 3600,
		},
	}

	db, err := New(config, WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}

	// Change MaxIdleConns dynamically
	sqlDB.SetMaxIdleConns(5)

	// Open a connection to populate the pool
	conn, err := sqlDB.Conn(context.Background())
	if err != nil {
		t.Fatalf("Failed to get connection: %v", err)
	}
	conn.Close()

	stats := sqlDB.Stats()
	// Verify the setting was applied (idle should not exceed the new max)
	if stats.Idle > 5 {
		t.Errorf("Expected Idle connections to be at most 5 after SetMaxIdleConns, got %d", stats.Idle)
	}
}

func TestPool_SetConnMaxLifetime(t *testing.T) {
	// Test that we can dynamically change ConnMaxLifetime
	dbPath := createTempDB(t)
	config := db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: dbPath,
			},
		},
		Pool: db.PoolConfig{
			MaxOpenConns:    10,
			MaxIdleConns:    2,
			ConnMaxLifetime: 3600,
			ConnMaxIdleTime: 3600,
		},
	}

	db, err := New(config, WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}

	// Change ConnMaxLifetime dynamically
	newLifetime := 30 * time.Minute
	sqlDB.SetConnMaxLifetime(newLifetime)

	// Verify the setting was applied by checking the pool is still functional
	err = sqlDB.Ping()
	if err != nil {
		t.Errorf("Failed to ping database after SetConnMaxLifetime: %v", err)
	}
}

func TestPool_SetConnMaxIdleTime(t *testing.T) {
	// Test that we can dynamically change ConnMaxIdleTime
	dbPath := createTempDB(t)
	config := db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: dbPath,
			},
		},
		Pool: db.PoolConfig{
			MaxOpenConns:    10,
			MaxIdleConns:    2,
			ConnMaxLifetime: 3600,
			ConnMaxIdleTime: 3600,
		},
	}

	db, err := New(config, WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}

	// Change ConnMaxIdleTime dynamically
	newIdleTime := 5 * time.Minute
	sqlDB.SetConnMaxIdleTime(newIdleTime)

	// Verify the setting was applied by checking the pool is still functional
	err = sqlDB.Ping()
	if err != nil {
		t.Errorf("Failed to ping database after SetConnMaxIdleTime: %v", err)
	}
}
