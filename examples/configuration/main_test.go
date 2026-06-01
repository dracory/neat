package main_test

import (
	"testing"
	"time"

	"github.com/dracory/neat"
	"github.com/dracory/neat/contracts/database/schema"
	mainpkg "github.com/dracory/neat/examples/configuration"
)

func TestRunExample(t *testing.T) {
	err := mainpkg.RunExample()
	if err != nil {
		t.Fatalf("RunExample failed: %v", err)
	}
}

func TestConfiguration_DSN_Functional(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("NewFromDSN failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Verify the connection is actually usable
	err = db.Schema().Create("ping", func(bp schema.Blueprint) {
		bp.ID()
		bp.String("value")
	})
	if err != nil {
		t.Fatalf("Schema.Create after NewFromDSN failed: %v", err)
	}
	if !db.Schema().HasTable("ping") {
		t.Error("expected 'ping' table to exist")
	}
}

func TestConfiguration_PoolConfig_Functional(t *testing.T) {
	cfg := neat.DBConfig{
		Default: "sqlite",
		Connections: map[string]neat.ConnectionConfig{
			"sqlite": {
				Driver:   "sqlite",
				Database: ":memory:",
			},
		},
		Pool: neat.PoolConfig{
			MaxIdleConns:    5,
			MaxOpenConns:    10,
			ConnMaxLifetime: time.Hour,
			ConnMaxIdleTime: time.Hour,
		},
	}

	db, err := neat.New(cfg)
	if err != nil {
		t.Fatalf("New with pool config failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Verify the connection is functional with pool settings applied
	err = db.Schema().Create("items", func(bp schema.Blueprint) {
		bp.ID()
		bp.String("name")
	})
	if err != nil {
		t.Fatalf("Schema.Create with pool config failed: %v", err)
	}
	if !db.Schema().HasTable("items") {
		t.Error("expected 'items' table to exist")
	}
}

func TestConfiguration_DebugMode_Functional(t *testing.T) {
	cfg := neat.DBConfig{
		Default: "sqlite",
		Connections: map[string]neat.ConnectionConfig{
			"sqlite": {
				Driver:   "sqlite",
				Database: ":memory:",
			},
		},
		Debug: true,
	}

	db, err := neat.New(cfg)
	if err != nil {
		t.Fatalf("New with debug config failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Verify debug mode doesn't prevent normal operation
	err = db.Schema().Create("logs", func(bp schema.Blueprint) {
		bp.ID()
		bp.String("message")
	})
	if err != nil {
		t.Fatalf("Schema.Create in debug mode failed: %v", err)
	}
	if !db.Schema().HasTable("logs") {
		t.Error("expected 'logs' table to exist in debug mode")
	}
}
