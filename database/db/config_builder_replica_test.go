package db

import (
	"testing"
)

// TestConnectionConfigReadFieldSet verifies that Read replicas can be stored
// and retrieved from ConnectionConfig.
func TestConnectionConfigReadFieldSet(t *testing.T) {
	cfg := ConnectionConfig{
		Driver: "mysql",
		Read: []ReplicaConfig{
			{Host: "replica1", Port: 3306, Database: "db", Username: "u", Password: "p"},
			{Host: "replica2", Port: 3306, Database: "db", Username: "u", Password: "p"},
		},
	}
	if len(cfg.Read) != 2 {
		t.Errorf("expected 2 Read replicas, got %d", len(cfg.Read))
	}
	if cfg.Read[0].Host != "replica1" {
		t.Errorf("expected first replica host 'replica1', got %q", cfg.Read[0].Host)
	}
}

// TestConnectionConfigWriteFieldSet verifies that Write primaries can be stored
// and retrieved from ConnectionConfig.
func TestConnectionConfigWriteFieldSet(t *testing.T) {
	cfg := ConnectionConfig{
		Driver: "mysql",
		Write: []ReplicaConfig{
			{Host: "primary1", Port: 3306, Database: "db", Username: "u", Password: "p"},
		},
	}
	if len(cfg.Write) != 1 {
		t.Errorf("expected 1 Write entry, got %d", len(cfg.Write))
	}
	if cfg.Write[0].Host != "primary1" {
		t.Errorf("expected write host 'primary1', got %q", cfg.Write[0].Host)
	}
}

// TestReplicaConfigFields verifies that ReplicaConfig has all five required fields.
func TestReplicaConfigFields(t *testing.T) {
	r := ReplicaConfig{
		Host:     "h",
		Port:     5432,
		Database: "d",
		Username: "u",
		Password: "p",
	}
	if r.Host != "h" {
		t.Errorf("Host field: got %q", r.Host)
	}
	if r.Port != 5432 {
		t.Errorf("Port field: got %d", r.Port)
	}
	if r.Database != "d" {
		t.Errorf("Database field: got %q", r.Database)
	}
	if r.Username != "u" {
		t.Errorf("Username field: got %q", r.Username)
	}
	if r.Password != "p" {
		t.Errorf("Password field: got %q", r.Password)
	}
}

// TestReplicaConfigRoundTrip verifies that Read replicas survive being stored
// in a DBConfig and retrieved.
func TestReplicaConfigRoundTrip(t *testing.T) {
	cfg := &DBConfig{
		Default: "default",
		Connections: map[string]ConnectionConfig{
			"default": {
				Driver: "sqlite",
				Database: ":memory:",
				Read: []ReplicaConfig{
					{Host: "read-host", Port: 3306},
				},
				Write: []ReplicaConfig{
					{Host: "write-host", Port: 3306},
				},
			},
		},
	}
	conn := cfg.Connections["default"]
	if len(conn.Read) != 1 || conn.Read[0].Host != "read-host" {
		t.Errorf("Read replica not preserved in DBConfig, got: %+v", conn.Read)
	}
	if len(conn.Write) != 1 || conn.Write[0].Host != "write-host" {
		t.Errorf("Write primary not preserved in DBConfig, got: %+v", conn.Write)
	}
}
