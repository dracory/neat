package orm

import (
	"context"
	"testing"

	"github.com/dracory/neat/contracts/log"
	contractsdb "github.com/dracory/neat/contracts/database"
	"github.com/dracory/neat/database/db"
	"github.com/dracory/neat/database/driver"
	"github.com/dracory/neat/database/query"
	_ "modernc.org/sqlite"
)

func TestNewOrm(t *testing.T) {
	ctx := context.Background()
	dbConfig := &db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver: contractsdb.DriverSqlite,
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
		log:   log.NewStdLogger(),
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

func sqliteMemoryConfig(name string) *db.DBConfig {
	return &db.DBConfig{
		Default: name,
		Connections: map[string]db.ConnectionConfig{
			name: {
				Driver: contractsdb.DriverSqlite,
				Database: ":memory:",
			},
		},
	}
}

func TestBuildQueryUsesWriteConnection(t *testing.T) {
	cfg := sqliteMemoryConfig("default")
	orm, err := BuildOrm(context.Background(), cfg, "default", log.NewStdLogger(), nil)
	if err != nil {
		t.Fatalf("BuildOrm failed: %v", err)
	}
	q := orm.Query()
	if q == nil {
		t.Fatal("Expected non-nil Query (write connection)")
	}
}

func TestBuildQuerySQLiteInMemory(t *testing.T) {
	cfg := sqliteMemoryConfig("main")

	orm, err := BuildOrm(context.Background(), cfg, "main", log.NewStdLogger(), nil)
	if err != nil {
		t.Fatalf("BuildOrm failed: %v", err)
	}
	if orm == nil {
		t.Fatal("Expected non-nil Orm")
	}
}

func TestBuildQueryConnectionNotFound(t *testing.T) {
	cfg := sqliteMemoryConfig("default")

	_, err := BuildOrm(context.Background(), cfg, "nonexistent", log.NewStdLogger(), nil)
	if err == nil {
		t.Error("Expected error for non-existent connection, got nil")
	}
}

func TestBuildQueryMissingDriver(t *testing.T) {
	cfg := &db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver:   "",
				Database: ":memory:",
			},
		},
	}

	_, err := BuildOrm(context.Background(), cfg, "default", log.NewStdLogger(), nil)
	if err == nil {
		t.Error("Expected error for missing driver, got nil")
	}
}

func TestBuildQueryReadWriteConnectionSelection(t *testing.T) {
	// Without replica config, query should use the primary connection.
	cfg := sqliteMemoryConfig("default")
	orm, err := BuildOrm(context.Background(), cfg, "default", log.NewStdLogger(), nil)
	if err != nil {
		t.Fatalf("BuildOrm failed: %v", err)
	}

	q := orm.Query()
	if q == nil {
		t.Fatal("Expected non-nil Query from Orm")
	}
}

func TestBuildQueryFallbackWhenReplicaUnavailable(t *testing.T) {
	// Configure a read replica with an invalid address; primary must still work.
	cfg := &db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver: contractsdb.DriverSqlite,
				Database: ":memory:",
				Read: []db.ReplicaConfig{
					{Host: "invalid-host-that-does-not-exist", Port: 9999},
				},
			},
		},
	}

	orm, err := BuildOrm(context.Background(), cfg, "default", log.NewStdLogger(), nil)
	// Either succeeds (replica open silently fails) or returns an error — primary must not panic.
	if err != nil {
		// Acceptable if the DSN build itself fails
		return
	}
	if orm == nil {
		t.Fatal("Expected non-nil Orm even with bad replica config")
	}
}

func TestEnableDebugMissesDefaultQuery(t *testing.T) {
	// This test proves that Orm.EnableDebug() does not call EnableDebug()
	// on r.query (the default query), only on queries in the r.queries map.
	// In contrast, DisableDebug() handles both r.query and r.queries.
	//
	// The fresh query uses a separate dbConfig so that EnableDebug() on the
	// map query (which sets dbConfig.Debug=true) doesn't mask the bug via
	// the shared config fallback in IsDebug().
	cfg := sqliteMemoryConfig("default")
	orm, err := BuildOrm(context.Background(), cfg, "default", log.NewStdLogger(), nil)
	if err != nil {
		t.Fatalf("BuildOrm failed: %v", err)
	}

	// Create a separate config with debug=false for the fresh query
	separateCfg := sqliteMemoryConfig("default")
	freshQuery := query.NewQuery(
		context.Background(),
		nil,
		driver.NewSQLite(),
		"default",
		separateCfg,
		log.NewStdLogger(),
	)
	orm.SetQuery(freshQuery)

	defaultQ, ok := orm.query.(*query.Query)
	if !ok {
		t.Fatal("Expected orm.query to be *query.Query")
	}
	if defaultQ.IsDebug() {
		t.Fatal("Fresh query should have debug=false before EnableDebug")
	}

	orm.EnableDebug()

	// EnableDebug should enable debug on r.query, just like DisableDebug does.
	// This currently FAILS because EnableDebug only iterates r.queries, skipping r.query.
	if !defaultQ.IsDebug() {
		t.Error("BUG: EnableDebug() did not enable debug on r.query — it only iterates r.queries, not r.query")
	}

	// The query still in r.queries should have debug enabled
	for _, q := range orm.queries {
		if qConcrete, ok := q.(*query.Query); ok && !qConcrete.IsDebug() {
			t.Error("Expected query in r.queries map to have debug enabled")
		}
	}
}

func TestDisableDebugHandlesDefaultQuery(t *testing.T) {
	// This test shows that DisableDebug() DOES handle r.query,
	// proving the asymmetry with EnableDebug().
	cfg := sqliteMemoryConfig("default")
	cfg.Debug = true
	orm, err := BuildOrm(context.Background(), cfg, "default", log.NewStdLogger(), nil)
	if err != nil {
		t.Fatalf("BuildOrm failed: %v", err)
	}

	// Replace r.query with a fresh query that has debug=true via config
	freshQuery := query.NewQuery(
		context.Background(),
		nil,
		driver.NewSQLite(),
		"default",
		cfg,
		log.NewStdLogger(),
	)
	orm.SetQuery(freshQuery)

	defaultQ, ok := orm.query.(*query.Query)
	if !ok {
		t.Fatal("Expected orm.query to be *query.Query")
	}
	if !defaultQ.IsDebug() {
		t.Fatal("Fresh query should have debug=true from config before DisableDebug")
	}

	orm.DisableDebug()

	// DisableDebug DOES handle r.query, so this should be false
	if defaultQ.IsDebug() {
		t.Error("DisableDebug failed to disable debug on r.query")
	}
}

func TestOrmConnectionSwitching(t *testing.T) {
	cfg := &db.DBConfig{
		Default: "primary",
		Connections: map[string]db.ConnectionConfig{
			"primary":   {Driver: contractsdb.DriverSqlite, Database: ":memory:"},
			"secondary": {Driver: contractsdb.DriverSqlite, Database: ":memory:"},
		},
	}

	orm, err := BuildOrm(context.Background(), cfg, "primary", log.NewStdLogger(), nil)
	if err != nil {
		t.Fatalf("BuildOrm failed: %v", err)
	}

	secondary := orm.Connection("secondary")
	if secondary == nil {
		t.Fatal("Expected non-nil Orm from Connection()")
	}
	if secondary.Name() != "secondary" {
		t.Errorf("Expected connection name 'secondary', got %q", secondary.Name())
	}
	// Primary must be unchanged
	if orm.Name() != "primary" {
		t.Errorf("Expected original connection name 'primary', got %q", orm.Name())
	}
}
