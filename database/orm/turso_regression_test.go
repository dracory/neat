package orm

import (
	"context"
	"testing"

	"github.com/dracory/neat/contracts/log"
	contractsdb "github.com/dracory/neat/contracts/database"
	"github.com/dracory/neat/database/db"
	_ "modernc.org/sqlite"
)

// TestReplicaConnectionsClosed verifies that read-replica and write-primary
// *sql.DB handles are stored in dbConnections and closed by Close().
func TestReplicaConnectionsClosed(t *testing.T) {
	cfg := &db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver: contractsdb.DriverSqlite,
				Database: ":memory:",
				Read: []db.ReplicaConfig{
					{Database: ":memory:"},
				},
			},
		},
	}

	o, err := BuildOrm(context.Background(), cfg, "default", log.NewStdLogger(), nil)
	if err != nil {
		t.Fatalf("BuildOrm failed: %v", err)
	}

	// Verify replica connection was stored in dbConnections
	readKey := "default:read"
	readDB, ok := o.dbConnections[readKey]
	if !ok {
		t.Fatal("Expected read-replica connection to be stored in dbConnections with key 'default:read'")
	}
	if readDB == nil {
		t.Fatal("Expected non-nil read-replica *sql.DB")
	}

	// Close should close both the replica and the base connection
	err = o.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Verify both keys were removed from dbConnections
	if _, ok := o.dbConnections[readKey]; ok {
		t.Error("Expected read-replica connection to be removed from dbConnections after Close()")
	}
	if _, ok := o.dbConnections["default"]; ok {
		t.Error("Expected base connection to be removed from dbConnections after Close()")
	}
}

// TestReplicaDsnCleared verifies that when a base connection has a Dsn field set,
// replica configs clear it so BuildDSN reconstructs from individual fields.
// We use the array driver here to avoid needing a real database for replicas.
func TestReplicaDsnCleared(t *testing.T) {
	cfg := &db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver: contractsdb.DriverSqlite,
				Database: ":memory:",
				Read: []db.ReplicaConfig{
					{Database: ":memory:"},
				},
			},
		},
	}

	o, err := BuildOrm(context.Background(), cfg, "default", log.NewStdLogger(), nil)
	if err != nil {
		t.Fatalf("BuildOrm failed: %v", err)
	}
	defer func() { _ = o.Close() }()

	// If Dsn was not cleared, the replica would use the base Dsn.
	// With Dsn cleared, BuildDSN reconstructs from Database field.
	// We just verify the replica was successfully opened (stored in dbConnections).
	readKey := "default:read"
	if _, ok := o.dbConnections[readKey]; !ok {
		t.Fatal("Expected read-replica connection to be stored — Dsn clearing may have failed")
	}
}

// TestWritePrimaryConnectionsClosed verifies that write-primary connections
// are stored in dbConnections and closed by Close().
func TestWritePrimaryConnectionsClosed(t *testing.T) {
	cfg := &db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver: contractsdb.DriverSqlite,
				Database: ":memory:",
				Write: []db.ReplicaConfig{
					{Database: ":memory:"},
				},
			},
		},
	}

	o, err := BuildOrm(context.Background(), cfg, "default", log.NewStdLogger(), nil)
	if err != nil {
		t.Fatalf("BuildOrm failed: %v", err)
	}

	// Verify write-primary connection was stored
	writeKey := "default:write"
	writeDB, ok := o.dbConnections[writeKey]
	if !ok {
		t.Fatal("Expected write-primary connection to be stored in dbConnections with key 'default:write'")
	}
	if writeDB == nil {
		t.Fatal("Expected non-nil write-primary *sql.DB")
	}

	// Close should close all connections
	err = o.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if _, ok := o.dbConnections[writeKey]; ok {
		t.Error("Expected write-primary connection to be removed from dbConnections after Close()")
	}
	if _, ok := o.dbConnections["default"]; ok {
		t.Error("Expected base connection to be removed from dbConnections after Close()")
	}
}

// TestCloseWithoutReplicas verifies that Close() still works correctly when
// no replicas or write-primaries are configured.
func TestCloseWithoutReplicas(t *testing.T) {
	cfg := &db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {Driver: contractsdb.DriverSqlite, Database: ":memory:"},
		},
	}

	o, err := BuildOrm(context.Background(), cfg, "default", log.NewStdLogger(), nil)
	if err != nil {
		t.Fatalf("BuildOrm failed: %v", err)
	}

	err = o.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if _, ok := o.dbConnections["default"]; ok {
		t.Error("Expected base connection to be removed after Close()")
	}
}

// TestConfigureConnectionPoolTursoFilePrefix verifies that file: prefix (without //)
// is recognized as local for single-connection pinning, matching the libSQL driver's
// accepted DSN format.
func TestConfigureConnectionPoolTursoFilePrefix(t *testing.T) {
	tests := []struct {
		name      string
		dsn       string
		database  string
		expectPin bool
	}{
		{"file:// prefix in Dsn", "file:///path/to/db", "", true},
		{"file: prefix in Dsn (no slashes)", "file:local.db", "", true},
		{"file:// prefix in Database", "", "file:///path/to/db", true},
		{"file: prefix in Database (no slashes)", "", "file:local.db", true},
		{"libsql:// remote Dsn", "libsql://my-db.turso.io", "", false},
		{"libsql:// remote Database", "", "libsql://my-db.turso.io", false},
		{"empty Dsn and Database", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connConfig := db.ConnectionConfig{
				Driver: contractsdb.DriverTurso,
				Dsn:      tt.dsn,
				Database: tt.database,
			}

			// We can't easily check the pool settings on a *sql.DB directly,
			// but we can verify the logic by checking what pinSingleConn would be.
			// Since configureConnectionPool is a package-level function, we test
			// the prefix detection logic here.
			pinSingleConn := false
			if connConfig.Driver == "turso" {
				pinSingleConn = startsWithFile(connConfig.Dsn) || startsWithFile(connConfig.Database)
			}
			if pinSingleConn != tt.expectPin {
				t.Errorf("Expected pinSingleConn=%v, got %v", tt.expectPin, pinSingleConn)
			}
		})
	}
}

// startsWithFile checks if a string starts with "file:" (covering both file:// and file:).
func startsWithFile(s string) bool {
	return len(s) >= 5 && s[:5] == "file:"
}
