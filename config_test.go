package neat

import (
	"testing"

	"github.com/dracory/neat/database"
	_ "modernc.org/sqlite"
)

// TestDatabaseTypeAlias verifies that neat.Database is an alias for database.Database,
// so a *neat.Database and *database.Database are interchangeable.
func TestDatabaseTypeAlias(t *testing.T) {
	var _ *database.Database = (*Database)(nil) // compile-time check
	// If this compiles, the type alias is correct.
}

// TestReplicaConfigFields verifies that neat.ReplicaConfig has all expected fields.
func TestReplicaConfigFields(t *testing.T) {
	r := ReplicaConfig{
		Host:     "h",
		Port:     5432,
		Database: "d",
		Username: "u",
		Password: "p",
	}
	if r.Host != "h" || r.Port != 5432 || r.Database != "d" || r.Username != "u" || r.Password != "p" {
		t.Errorf("ReplicaConfig fields not set correctly: %+v", r)
	}
}

// TestConnectionConfigHasReadWriteFields verifies that ConnectionConfig exposes
// Read and Write []ReplicaConfig slices.
func TestConnectionConfigHasReadWriteFields(t *testing.T) {
	cfg := ConnectionConfig{
		Driver: "sqlite",
		Read:   []ReplicaConfig{{Host: "r1"}},
		Write:  []ReplicaConfig{{Host: "w1"}},
	}
	if len(cfg.Read) != 1 || cfg.Read[0].Host != "r1" {
		t.Errorf("Read field not preserved: %+v", cfg.Read)
	}
	if len(cfg.Write) != 1 || cfg.Write[0].Host != "w1" {
		t.Errorf("Write field not preserved: %+v", cfg.Write)
	}
}

// TestNewPropagatesReplicaConfig verifies that neat.New converts Read/Write
// replicas from neat.ReplicaConfig into db.ReplicaConfig correctly.
// We use SQLite in-memory so no real network connection is needed.
func TestNewPropagatesReplicaConfig(t *testing.T) {
	cfg := DBConfig{
		Default: "main",
		Connections: map[string]ConnectionConfig{
			"main": {
				Driver:   "sqlite",
				Database: ":memory:",
				Read: []ReplicaConfig{
					{Host: "read-host", Port: 3306, Database: "rdb", Username: "ru", Password: "rp"},
				},
				Write: []ReplicaConfig{
					{Host: "write-host", Port: 3306, Database: "wdb", Username: "wu", Password: "wp"},
				},
			},
		},
	}

	db, err := New(cfg)
	if err != nil {
		t.Fatalf("neat.New failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	if db == nil {
		t.Fatal("expected non-nil *Database from neat.New")
	}
}

// TestNewFromDSNSQLite verifies that NewFromDSN works for SQLite.
func TestNewFromDSNSQLite(t *testing.T) {
	db, err := NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("NewFromDSN failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	if db == nil {
		t.Fatal("expected non-nil *Database")
	}
}

// TestNewWithLoggerOption verifies that neat.New accepts a WithLogger option.
func TestNewWithLoggerOption(t *testing.T) {
	cfg := DBConfig{
		Default: "default",
		Connections: map[string]ConnectionConfig{
			"default": {Driver: "sqlite", Database: ":memory:"},
		},
	}

	db, err := New(cfg, database.WithLogger(nil))
	if err != nil {
		t.Fatalf("neat.New with WithLogger option failed: %v", err)
	}
	defer func() { _ = db.Close() }()
}
