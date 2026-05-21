package orm

import (
	"context"
	"testing"

	"github.com/dracory/neat/contracts/log"
	"github.com/dracory/neat/database/db"
)

func TestNewOrm(t *testing.T) {
	ctx := context.Background()
	dbConfig := &db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: ":memory:",
			},
		},
	}
	connection := "default"
	logger := log.NewStdLogger()

	orm := NewOrm(
		ctx,
		dbConfig,
		connection,
		nil,
		nil,
		logger,
		nil,
		nil,
		nil,
		nil,
	)

	if orm == nil {
		t.Fatal("Expected orm to be created")
	}

	if orm.ctx != ctx {
		t.Error("Expected ctx to be set")
	}

	if orm.dbConfig != dbConfig {
		t.Error("Expected dbConfig to be set")
	}

	if orm.connection != connection {
		t.Error("Expected connection to be set")
	}

	if orm.log != logger {
		t.Error("Expected log to be set")
	}
}

func TestOrmConnection(t *testing.T) {
	orm := &Orm{
		connection: "default",
	}

	if orm.Name() != "default" {
		t.Errorf("Expected connection name to be default, got %s", orm.Name())
	}
}

func TestOrmQuery(t *testing.T) {
	orm := &Orm{
		query: nil,
	}

	query := orm.Query()
	// Query may be nil if not initialized, just verify the method doesn't panic
	_ = query
}

func TestOrmDB(t *testing.T) {
	orm := &Orm{
		dbConnections: nil,
	}

	db, err := orm.DB()
	if err == nil {
		t.Error("Expected error when no DB connections are available")
	}
	if db != nil {
		t.Error("Expected nil DB when no connections are available")
	}
}
