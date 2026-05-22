package query

import (
	"context"
	"database/sql"
	"testing"

	"github.com/dracory/neat/database/db"
	"github.com/dracory/neat/database/driver"
	_ "modernc.org/sqlite"
)

// newSentinelDB opens a real in-memory SQLite connection for pointer-identity tests.
func newSentinelDB() *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		panic("newSentinelDB: " + err.Error())
	}
	return db
}

func testDBConfig() *db.DBConfig {
	return &db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {Driver: "sqlite", Database: ":memory:"},
		},
	}
}

// --- readConn / writeConn fallback ---

func TestReadConnFallsBackToPrimary(t *testing.T) {
	primary := newSentinelDB()
	defer primary.Close()

	q := NewQuery(context.Background(), primary, nil, "", testDBConfig(), nil)
	// readDB is nil — must return primary
	if got := q.readConn(); got != primary {
		t.Errorf("readConn() should return primary when readDB is nil")
	}
}

func TestReadConnUsesReplicaWhenSet(t *testing.T) {
	primary := newSentinelDB()
	replica := newSentinelDB()
	defer primary.Close()
	defer replica.Close()

	q := NewQuery(context.Background(), primary, nil, "", testDBConfig(), nil)
	q.readDB = replica

	if got := q.readConn(); got != replica {
		t.Errorf("readConn() should return readDB when set")
	}
}

func TestWriteConnFallsBackToPrimary(t *testing.T) {
	primary := newSentinelDB()
	defer primary.Close()

	q := NewQuery(context.Background(), primary, nil, "", testDBConfig(), nil)
	if got := q.writeConn(); got != primary {
		t.Errorf("writeConn() should return primary when writeDB is nil")
	}
}

func TestWriteConnUsesWriteWhenSet(t *testing.T) {
	primary := newSentinelDB()
	write := newSentinelDB()
	defer primary.Close()
	defer write.Close()

	q := NewQuery(context.Background(), primary, nil, "", testDBConfig(), nil)
	q.writeDB = write

	if got := q.writeConn(); got != write {
		t.Errorf("writeConn() should return writeDB when set")
	}
}

// --- NewQueryWithReplicas ---

func TestNewQueryWithReplicasSetsFields(t *testing.T) {
	primary := newSentinelDB()
	readReplica := newSentinelDB()
	defer primary.Close()
	defer readReplica.Close()

	drv := driver.NewSQLite()
	q := NewQueryWithReplicas(context.Background(), primary, readReplica, drv, "default", testDBConfig(), nil)

	if q.readDB != readReplica {
		t.Errorf("NewQueryWithReplicas: readDB not set correctly")
	}
	if q.writeDB != primary {
		t.Errorf("NewQueryWithReplicas: writeDB should equal writeConn arg")
	}
	if q.db != primary {
		t.Errorf("NewQueryWithReplicas: primary db field not set")
	}
}

// --- Clone propagates replicas ---

func TestClonePropagatesReplicas(t *testing.T) {
	primary := newSentinelDB()
	readReplica := newSentinelDB()
	write := newSentinelDB()
	defer primary.Close()
	defer readReplica.Close()
	defer write.Close()

	q := NewQuery(context.Background(), primary, nil, "", testDBConfig(), nil)
	q.readDB = readReplica
	q.writeDB = write

	clone := q.Clone().(*Query)

	if clone.readDB != readReplica {
		t.Errorf("Clone() did not propagate readDB")
	}
	if clone.writeDB != write {
		t.Errorf("Clone() did not propagate writeDB")
	}
}

// --- DB() / ReadDB() error during transaction ---

func TestDBErrorsDuringTransaction(t *testing.T) {
	primary := newSentinelDB()
	defer primary.Close()

	q := NewQuery(context.Background(), primary, nil, "", testDBConfig(), nil)
	q.tx = &sql.Tx{} // non-nil tx signals active transaction

	_, err := q.DB()
	if err == nil {
		t.Errorf("DB() should return error during active transaction")
	}
}

func TestReadDBErrorsDuringTransaction(t *testing.T) {
	primary := newSentinelDB()
	defer primary.Close()

	q := NewQuery(context.Background(), primary, nil, "", testDBConfig(), nil)
	q.tx = &sql.Tx{}

	_, err := q.ReadDB()
	if err == nil {
		t.Errorf("ReadDB() should return error during active transaction")
	}
}

// --- DB() / ReadDB() return correct connections when no tx ---

func TestDBReturnsWriteConn(t *testing.T) {
	primary := newSentinelDB()
	write := newSentinelDB()
	defer primary.Close()
	defer write.Close()

	q := NewQuery(context.Background(), primary, nil, "", testDBConfig(), nil)
	q.writeDB = write

	got, err := q.DB()
	if err != nil {
		t.Fatalf("DB() unexpected error: %v", err)
	}
	if got != write {
		t.Errorf("DB() should return writeDB when set")
	}
}

func TestReadDBReturnsReadConn(t *testing.T) {
	primary := newSentinelDB()
	replica := newSentinelDB()
	defer primary.Close()
	defer replica.Close()

	q := NewQuery(context.Background(), primary, nil, "", testDBConfig(), nil)
	q.readDB = replica

	got, err := q.ReadDB()
	if err != nil {
		t.Fatalf("ReadDB() unexpected error: %v", err)
	}
	if got != replica {
		t.Errorf("ReadDB() should return readDB when set")
	}
}
