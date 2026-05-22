package query_test

import (
	"database/sql"
	"testing"

	"github.com/dracory/neat/database/driver"
	"github.com/dracory/neat/database/query"
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

// --- readConn / writeConn fallback ---

func TestReadConnFallsBackToPrimary(t *testing.T) {
	primary := newSentinelDB()
	defer primary.Close()

	w := query.WrapQuery(query.NewTestQuery(primary, nil, query.MakeDBConfig(), nil))
	if got := w.ReadConn(); got != primary {
		t.Errorf("ReadConn() should return primary when readDB is nil")
	}
}

func TestReadConnUsesReplicaWhenSet(t *testing.T) {
	primary := newSentinelDB()
	replica := newSentinelDB()
	defer primary.Close()
	defer replica.Close()

	w := query.WrapQuery(query.NewTestQuery(primary, nil, query.MakeDBConfig(), nil))
	w.SetReadDB(replica)
	if got := w.ReadConn(); got != replica {
		t.Errorf("ReadConn() should return readDB when set")
	}
}

func TestWriteConnFallsBackToPrimary(t *testing.T) {
	primary := newSentinelDB()
	defer primary.Close()

	w := query.WrapQuery(query.NewTestQuery(primary, nil, query.MakeDBConfig(), nil))
	if got := w.WriteConn(); got != primary {
		t.Errorf("WriteConn() should return primary when writeDB is nil")
	}
}

func TestWriteConnUsesWriteWhenSet(t *testing.T) {
	primary := newSentinelDB()
	write := newSentinelDB()
	defer primary.Close()
	defer write.Close()

	w := query.WrapQuery(query.NewTestQuery(primary, nil, query.MakeDBConfig(), nil))
	w.SetWriteDB(write)
	if got := w.WriteConn(); got != write {
		t.Errorf("WriteConn() should return writeDB when set")
	}
}

// --- NewTestQueryWithReplicas ---

func TestNewQueryWithReplicasSetsFields(t *testing.T) {
	primary := newSentinelDB()
	readReplica := newSentinelDB()
	defer primary.Close()
	defer readReplica.Close()

	drv := driver.NewSQLite()
	q := query.NewTestQueryWithReplicas(primary, readReplica, drv, query.MakeDBConfig())
	w := query.WrapQuery(q)

	if w.ReadDB() != readReplica {
		t.Errorf("NewTestQueryWithReplicas: ReadDB not set correctly")
	}
	if w.WriteDB() != primary {
		t.Errorf("NewTestQueryWithReplicas: WriteDB should equal write arg")
	}
	if w.PrimaryDB() != primary {
		t.Errorf("NewTestQueryWithReplicas: PrimaryDB not set")
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

	w := query.WrapQuery(query.NewTestQuery(primary, nil, query.MakeDBConfig(), nil))
	w.SetReadDB(readReplica)
	w.SetWriteDB(write)

	cloneQ := w.Q.Clone()
	cloneW := query.WrapQuery(cloneQ.(*query.Query))

	if cloneW.ReadDB() != readReplica {
		t.Errorf("Clone() did not propagate readDB")
	}
	if cloneW.WriteDB() != write {
		t.Errorf("Clone() did not propagate writeDB")
	}
}

// --- DB() / ReadDB() error during transaction ---

func TestDBErrorsDuringTransaction(t *testing.T) {
	primary := newSentinelDB()
	defer primary.Close()

	w := query.WrapQuery(query.NewTestQuery(primary, nil, query.MakeDBConfig(), nil))
	w.SetTx(&sql.Tx{})

	_, err := w.Q.DB()
	if err == nil {
		t.Errorf("DB() should return error during active transaction")
	}
}

func TestReadDBErrorsDuringTransaction(t *testing.T) {
	primary := newSentinelDB()
	defer primary.Close()

	w := query.WrapQuery(query.NewTestQuery(primary, nil, query.MakeDBConfig(), nil))
	w.SetTx(&sql.Tx{})

	_, err := w.Q.ReadDB()
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

	w := query.WrapQuery(query.NewTestQuery(primary, nil, query.MakeDBConfig(), nil))
	w.SetWriteDB(write)

	got, err := w.Q.DB()
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

	w := query.WrapQuery(query.NewTestQuery(primary, nil, query.MakeDBConfig(), nil))
	w.SetReadDB(replica)

	got, err := w.Q.ReadDB()
	if err != nil {
		t.Fatalf("ReadDB() unexpected error: %v", err)
	}
	if got != replica {
		t.Errorf("ReadDB() should return readDB when set")
	}
}
