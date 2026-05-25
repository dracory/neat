//go:build integration

package mysql

import (
	"testing"
	"time"

	"github.com/dracory/neat"
	contractsschema "github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/contracts/log"
	"github.com/dracory/neat/database"
)

func TestMySQLIntegrationConnectionSwitch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := GetMySQLConfig()
	config.Connections["mysql2"] = config.Connections["mysql"]

	db, err := neat.New(config, database.WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	conn1 := db
	conn2, err := db.Connection("mysql2")
	if err != nil {
		t.Fatalf("Failed to get mysql2 connection: %v", err)
	}

	tableName1 := "users_conn1"
	tableName2 := "users_conn2"

	_ = conn1.Schema().Drop(tableName1)
	err = conn1.Schema().Create(tableName1, func(table contractsschema.Blueprint) {
		table.ID()
		table.String("name")
	})
	if err != nil {
		t.Fatalf("Failed to create table %s: %v", tableName1, err)
	}

	_ = conn2.Schema().Drop(tableName2)
	err = conn2.Schema().Create(tableName2, func(table contractsschema.Blueprint) {
		table.ID()
		table.String("name")
	})
	if err != nil {
		t.Fatalf("Failed to create table %s: %v", tableName2, err)
	}

	defer func() {
		_ = conn1.Schema().Drop(tableName1)
		_ = conn2.Schema().Drop(tableName2)
	}()

	err = conn1.Query().Table(tableName1).Create(map[string]any{"name": "user1"})
	if err != nil {
		t.Errorf("Insert into conn1 failed: %v", err)
	}

	err = conn2.Query().Table(tableName2).Create(map[string]any{"name": "user2"})
	if err != nil {
		t.Errorf("Insert into conn2 failed: %v", err)
	}

	var count int64
	err = conn1.Query().Table(tableName1).Count(&count)
	if err != nil || count != 1 {
		t.Errorf("Expected count=1 on conn1, got %d, err=%v", count, err)
	}

	err = conn2.Query().Table(tableName2).Count(&count)
	if err != nil || count != 1 {
		t.Errorf("Expected count=1 on conn2, got %d, err=%v", count, err)
	}
}

func TestMySQLIntegrationConnectionDefaultName(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := GetMySQLConfig()

	db, err := neat.New(config, database.WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	conn, err := db.Connection("")
	if err != nil {
		t.Fatalf("Failed to get default connection: %v", err)
	}
	expectedConn := "mysql"
	if conn.Name() != expectedConn {
		t.Errorf("Expected connection name '%s', got '%s'", expectedConn, conn.Name())
	}
}

func TestMySQLIntegrationConnectionNonExistent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := GetMySQLConfig()

	db, err := neat.New(config, database.WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	_, err = db.Connection("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent connection")
	}
}

func TestMySQLIntegrationReadWriteSeparation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	baseConn := GetMySQLConfig().Connections["mysql"]

	config := neat.DBConfig{
		Default: "rw",
		Connections: map[string]neat.ConnectionConfig{
			"rw": {
				Driver: baseConn.Driver,
				Read: []neat.ReplicaConfig{
					{Host: baseConn.Host, Port: baseConn.Port, Database: baseConn.Database, Username: baseConn.Username, Password: baseConn.Password},
				},
				Write: []neat.ReplicaConfig{
					{Host: baseConn.Host, Port: baseConn.Port, Database: baseConn.Database, Username: baseConn.Username, Password: baseConn.Password},
				},
				Charset: baseConn.Charset,
				Loc:     baseConn.Loc,
			},
		},
	}

	db, err := neat.New(config, database.WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create DB with read/write config: %v", err)
	}
	defer db.Close()

	tableName := "rw_sep_test"
	_ = db.Schema().Drop(tableName)
	err = db.Schema().Create(tableName, func(table contractsschema.Blueprint) {
		table.ID()
		table.String("name")
	})
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}
	defer db.Schema().Drop(tableName)

	err = db.Query().Table(tableName).Create(map[string]any{"name": "writer"})
	if err != nil {
		t.Errorf("Write failed: %v", err)
	}

	var count int64
	err = db.Query().Table(tableName).Count(&count)
	if err != nil || count != 1 {
		t.Errorf("Expected count=1 after write, got %d, err=%v", count, err)
	}
}

func TestMySQLIntegrationConnectionPoolSettings(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := GetMySQLConfig()
	config.Pool = neat.PoolConfig{
		MaxIdleConns:    7,
		MaxOpenConns:    13,
		ConnMaxLifetime: 10 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}

	db, err := neat.New(config, database.WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create database with pool settings: %v", err)
	}
	defer db.Close()

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}
	if sqlDB == nil {
		t.Error("Expected non-nil sql.DB")
	}
}
