package query_test

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/contracts/log"
	"github.com/dracory/neat/database/db"
	"github.com/dracory/neat/database/driver"
	"github.com/dracory/neat/database/query"
	_ "modernc.org/sqlite"
)

// --- Connection tests ---

func twoConnectionDBConfig() *db.DBConfig {
	return &db.DBConfig{
		Default: "primary",
		Connections: map[string]db.ConnectionConfig{
			"primary":   {Driver: "sqlite", Database: ":memory:"},
			"secondary": {Driver: "sqlite", Database: ":memory:"},
		},
	}
}

func TestConnectionSwitchUnknownNameReturnsSelf(t *testing.T) {
	w := openSQLiteQuery(t)
	w.SetDBConfig(twoConnectionDBConfig())

	returned := w.Q.Connection("nonexistent")
	if returned != w.Q {
		t.Error("expected Connection('nonexistent') to return the original query")
	}
}

func TestConnectionSwitchReturnsNewQuery(t *testing.T) {
	w := openSQLiteQuery(t)
	w.SetDBConfig(twoConnectionDBConfig())

	newQ := w.Q.Connection("secondary")
	if newQ == nil {
		t.Fatal("expected non-nil query from Connection()")
	}
	if newQ == w.Q {
		t.Error("expected Connection() to return a new Query instance, not the same")
	}
}

func TestConnectionSwitchUsesCorrectDriver(t *testing.T) {
	w := openSQLiteQuery(t)
	w.SetDBConfig(twoConnectionDBConfig())

	newQ := w.Q.Connection("secondary")
	got := string(newQ.Driver())
	if got != "sqlite" {
		t.Errorf("expected driver 'sqlite' for secondary connection, got %q", got)
	}
}

func TestConnectionSwitchEmptyNameReturnsSelf(t *testing.T) {
	w := openSQLiteQuery(t)
	w.SetDBConfig(twoConnectionDBConfig())

	returned := w.Q.Connection("")
	if returned != w.Q {
		t.Error("expected Connection('') to return the original query")
	}
}

// --- Model tests ---

type User struct {
	ID   uint
	Name string
}

type Address struct {
	ID     uint
	Name   string
	UserID uint
}

func (User) TableName() string {
	return "users"
}

func (Address) TableName() string {
	return "addresses"
}

func TestModelAlwaysUpdatesTableName(t *testing.T) {
	w := query.WrapQuery(query.NewTestQuery(nil, nil, query.MakeDBConfig(), nil))

	w.Q.Model(&User{})
	if w.GetTable() != "users" {
		t.Errorf("Expected table 'users', got '%s'", w.GetTable())
	}

	w.Q.Model(&Address{})
	if w.GetTable() != "addresses" {
		t.Errorf("Expected table 'addresses' after second Model() call, got '%s'", w.GetTable())
	}
}

func TestModelWithNilValue(t *testing.T) {
	w := query.WrapQuery(query.NewTestQuery(nil, nil, query.MakeDBConfig(), nil))

	w.Q.Model(nil)
	if w.GetTable() != "" {
		t.Errorf("Expected empty table for nil model, got '%s'", w.GetTable())
	}
}

// --- Query log tests ---

type capturingLogger struct {
	mu       sync.Mutex
	warnings []string
}

func (l *capturingLogger) Debugf(format string, args ...any) {}
func (l *capturingLogger) Infof(format string, args ...any)  {}
func (l *capturingLogger) Errorf(format string, args ...any) {}
func (l *capturingLogger) Warning(args ...any)               {}
func (l *capturingLogger) Warningf(format string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.warnings = append(l.warnings, format)
}
func (l *capturingLogger) hasWarning() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.warnings) > 0
}
func (l *capturingLogger) warningContains(s string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, w := range l.warnings {
		if strings.Contains(w, s) {
			return true
		}
	}
	return false
}

var _ log.Log = (*capturingLogger)(nil)

func openSQLiteQueryWithConfig(t *testing.T, cfg *db.DBConfig, lg log.Log) *query.TestQuery {
	t.Helper()
	return query.WrapQuery(query.NewTestQueryWithConfig(t, cfg, lg))
}

func logTestDBConfig(slowThresholdMs int) *db.DBConfig {
	cfg := query.MakeDBConfig()
	cfg.SlowThreshold = slowThresholdMs
	return cfg
}

func TestEnableQueryLogCapturesEntries(t *testing.T) {
	w := openSQLiteQueryWithConfig(t, logTestDBConfig(0), nil)
	execSQL(t, w, "CREATE TABLE ql_test (id INTEGER)")
	execSQL(t, w, "INSERT INTO ql_test VALUES (1)")

	w.Q.EnableQueryLog()
	w.SetTable("ql_test")
	var result []map[string]any
	_ = w.Q.Get(&result)

	logs := w.Q.GetQueryLog()
	if len(logs) == 0 {
		t.Error("expected at least one query log entry after EnableQueryLog")
	}
}

func TestDisableQueryLogSuppressesEntries(t *testing.T) {
	w := openSQLiteQueryWithConfig(t, logTestDBConfig(0), nil)
	execSQL(t, w, "CREATE TABLE ql_off (id INTEGER)")
	execSQL(t, w, "INSERT INTO ql_off VALUES (1)")

	w.Q.EnableQueryLog()
	w.Q.DisableQueryLog()
	w.SetTable("ql_off")
	var result []map[string]any
	_ = w.Q.Get(&result)

	if len(w.Q.GetQueryLog()) != 0 {
		t.Error("expected no log entries after DisableQueryLog")
	}
}

func TestFlushQueryLogClearsEntries(t *testing.T) {
	w := openSQLiteQueryWithConfig(t, logTestDBConfig(0), nil)
	execSQL(t, w, "CREATE TABLE ql_flush (id INTEGER)")
	execSQL(t, w, "INSERT INTO ql_flush VALUES (1)")

	w.Q.EnableQueryLog()
	w.SetTable("ql_flush")
	var result []map[string]any
	_ = w.Q.Get(&result)

	if len(w.Q.GetQueryLog()) == 0 {
		t.Fatal("precondition: expected log entries before flush")
	}
	w.Q.FlushQueryLog()
	if len(w.Q.GetQueryLog()) != 0 {
		t.Error("expected empty log after FlushQueryLog")
	}
}

func TestLogQueryRecordsDuration(t *testing.T) {
	w := openSQLiteQueryWithConfig(t, logTestDBConfig(0), nil)
	execSQL(t, w, "CREATE TABLE ql_dur (id INTEGER)")
	execSQL(t, w, "INSERT INTO ql_dur VALUES (1)")

	w.Q.EnableQueryLog()
	w.SetTable("ql_dur")
	var result []map[string]any
	_ = w.Q.Get(&result)

	logs := w.Q.GetQueryLog()
	if len(logs) == 0 {
		t.Fatal("expected query log entry")
	}
	if logs[0].Time < 0 {
		t.Errorf("expected Time >= 0, got %v", logs[0].Time)
	}
}

func TestLogQuerySlowThresholdEmitsWarning(t *testing.T) {
	lg := &capturingLogger{}
	w := openSQLiteQueryWithConfig(t, logTestDBConfig(1), lg)

	startTime := time.Now().Add(-time.Millisecond * 2)
	w.LogQuery("SELECT 1", nil, startTime)

	if !lg.hasWarning() {
		t.Error("expected slow-query warning when elapsed >= SlowThreshold")
	}
}

func TestLogQuerySlowThresholdContainsSlow(t *testing.T) {
	lg := &capturingLogger{}
	w := openSQLiteQueryWithConfig(t, logTestDBConfig(1), lg)

	startTime := time.Now().Add(-time.Millisecond * 2)
	w.LogQuery("SELECT 1", nil, startTime)

	if !lg.warningContains("slow query") {
		t.Error("expected warning message to contain 'slow query'")
	}
}

func TestLogQueryNoWarnWhenThresholdNotSet(t *testing.T) {
	lg := &capturingLogger{}
	cfg := &db.DBConfig{
		Default:       "default",
		Connections:   map[string]db.ConnectionConfig{"default": {Driver: "sqlite", Database: ":memory:"}},
		SlowThreshold: -1,
	}
	w := openSQLiteQueryWithConfig(t, cfg, lg)
	execSQL(t, w, "CREATE TABLE ql_nowarn (id INTEGER)")
	execSQL(t, w, "INSERT INTO ql_nowarn VALUES (1)")

	w.Q.EnableQueryLog()
	w.SetTable("ql_nowarn")
	var result []map[string]any
	_ = w.Q.Get(&result)

	if lg.hasWarning() {
		t.Error("expected no warning when SlowThreshold is negative")
	}
}

// --- Query routing tests ---

func newSentinelDB() *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		panic("newSentinelDB: " + err.Error())
	}
	return db
}

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

// --- Context handling tests ---

func TestWithContextReturnsNewQuery(t *testing.T) {
	w := openSQLiteQuery(t)
	ctx := context.Background()

	newQ := w.Q.WithContext(ctx)
	if newQ == nil {
		t.Fatal("expected non-nil query from WithContext()")
	}
	if newQ == w.Q {
		t.Error("expected WithContext() to return a new Query instance, not the same")
	}
}

func TestWithContextSetsContext(t *testing.T) {
	w := openSQLiteQuery(t)
	ctx := context.Background()

	newQ := w.Q.WithContext(ctx)
	wrapped := query.WrapQuery(newQ.(*query.Query))

	if wrapped.Context() != ctx {
		t.Error("expected WithContext() to set the context on the new query")
	}
}

func TestWithContextPreservesOriginalContext(t *testing.T) {
	w := openSQLiteQuery(t)
	ctx1 := context.Background()
	originalCtx := w.Context()

	newQ := w.Q.WithContext(ctx1)
	wrapped := query.WrapQuery(newQ.(*query.Query))

	if wrapped.Context() != ctx1 {
		t.Error("expected new query to have ctx1")
	}

	// Original query should still have its original context
	if w.Context() != originalCtx {
		t.Error("expected original query to retain its original context")
	}
}

func TestContextPropagationToClone(t *testing.T) {
	w := openSQLiteQuery(t)
	ctx := context.Background()
	w.SetContext(ctx)

	cloneQ := w.Q.Clone()
	cloneW := query.WrapQuery(cloneQ.(*query.Query))

	if cloneW.Context() != ctx {
		t.Error("expected Clone() to propagate context")
	}
}

func TestContextCancellationPreventsQuery(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE ctx_cancel (id INTEGER)")
	execSQL(t, w, "INSERT INTO ctx_cancel VALUES (1)")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	w.Q = w.Q.WithContext(ctx).(*query.Query)
	w.SetTable("ctx_cancel")

	var result []map[string]any
	err := w.Q.Get(&result)

	if err == nil {
		t.Error("expected query to fail with cancelled context")
	}
	if !errors.Is(err, context.Canceled) && !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("expected context.Canceled error, got: %v", err)
	}
}

func TestContextWithValue(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE ctx_value (id INTEGER)")
	execSQL(t, w, "INSERT INTO ctx_value VALUES (1)")

	key := "test-key"
	value := "test-value"
	ctx := context.WithValue(context.Background(), key, value)

	w.Q = w.Q.WithContext(ctx).(*query.Query)
	w.SetTable("ctx_value")

	var result []map[string]any
	err := w.Q.Get(&result)

	if err != nil {
		t.Errorf("unexpected error with context value: %v", err)
	}

	wrapped := query.WrapQuery(w.Q)
	if wrapped.Context().Value(key) != value {
		t.Error("expected context value to be preserved")
	}
}

func TestContextWithTransaction(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE ctx_tx (id INTEGER, name TEXT)")

	ctx := context.Background()
	w.Q = w.Q.WithContext(ctx).(*query.Query)

	err := w.Q.Transaction(func(tx contractsorm.Query) error {
		wrapped := query.WrapQuery(tx.(*query.Query))
		if wrapped.Context() != ctx {
			t.Error("expected transaction query to preserve context")
		}

		wrapped.SetTable("ctx_tx")
		return tx.Create(map[string]any{"id": 1, "name": "test"})
	})

	if err != nil {
		t.Errorf("unexpected error in transaction with context: %v", err)
	}
}

func TestContextPropagationThroughChainedMethods(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE ctx_chain (id INTEGER)")
	execSQL(t, w, "INSERT INTO ctx_chain VALUES (1)")

	ctx := context.Background()
	w.Q = w.Q.WithContext(ctx).(*query.Query)
	w.Q = w.Q.Table("ctx_chain").Where("id", 1).(*query.Query)

	wrapped := query.WrapQuery(w.Q)
	if wrapped.Context() != ctx {
		t.Error("expected context to be preserved through chained methods")
	}

	var result map[string]any
	err := w.Q.First(&result)
	if err != nil {
		t.Errorf("unexpected error with chained context: %v", err)
	}
}

// --- Observer registration tests ---

type TestObserver struct {
	CreatedCalled  bool
	UpdatedCalled  bool
	DeletedCalled  bool
	CreatingCalled bool
	UpdatingCalled bool
	DeletingCalled bool
	SavingCalled   bool
	SavedCalled    bool
}

func (o *TestObserver) Created(event contractsorm.Event) error {
	o.CreatedCalled = true
	return nil
}

func (o *TestObserver) Updated(event contractsorm.Event) error {
	o.UpdatedCalled = true
	return nil
}

func (o *TestObserver) Deleted(event contractsorm.Event) error {
	o.DeletedCalled = true
	return nil
}

func (o *TestObserver) Creating(event contractsorm.Event) error {
	o.CreatingCalled = true
	return nil
}

func (o *TestObserver) Updating(event contractsorm.Event) error {
	o.UpdatingCalled = true
	return nil
}

func (o *TestObserver) Deleting(event contractsorm.Event) error {
	o.DeletingCalled = true
	return nil
}

func (o *TestObserver) ForceDeleted(event contractsorm.Event) error {
	return nil
}

func (o *TestObserver) ForceDeleting(event contractsorm.Event) error {
	return nil
}

func (o *TestObserver) Restored(event contractsorm.Event) error {
	return nil
}

func (o *TestObserver) Restoring(event contractsorm.Event) error {
	return nil
}

func (o *TestObserver) Retrieved(event contractsorm.Event) error {
	return nil
}

func (o *TestObserver) Saving(event contractsorm.Event) error {
	o.SavingCalled = true
	return nil
}

func (o *TestObserver) Saved(event contractsorm.Event) error {
	o.SavedCalled = true
	return nil
}

type TestModel struct {
	ID   uint
	Name string
}

func (TestModel) TableName() string {
	return "observer_test"
}

func TestObserveRegistersObserver(t *testing.T) {
	w := openSQLiteQuery(t)
	observer := &TestObserver{}

	w.Q.Observe(&TestModel{}, observer)

	wrapped := query.WrapQuery(w.Q)
	observers := wrapped.GetModelToObserver()

	if len(observers) != 1 {
		t.Errorf("expected 1 observer, got %d", len(observers))
	}

	if observers[0].Observer != observer {
		t.Error("observer not registered correctly")
	}
}

func TestObserveMultipleObservers(t *testing.T) {
	w := openSQLiteQuery(t)
	observer1 := &TestObserver{}
	observer2 := &TestObserver{}

	w.Q.Observe(&TestModel{}, observer1)
	w.Q.Observe(&TestModel{}, observer2)

	wrapped := query.WrapQuery(w.Q)
	observers := wrapped.GetModelToObserver()

	if len(observers) != 2 {
		t.Errorf("expected 2 observers, got %d", len(observers))
	}
}

func TestObserveWithDifferentModels(t *testing.T) {
	w := openSQLiteQuery(t)
	observer1 := &TestObserver{}
	observer2 := &TestObserver{}

	type AnotherModel struct {
		ID uint
	}

	w.Q.Observe(&TestModel{}, observer1)
	w.Q.Observe(&AnotherModel{}, observer2)

	wrapped := query.WrapQuery(w.Q)
	observers := wrapped.GetModelToObserver()

	if len(observers) != 2 {
		t.Errorf("expected 2 observers, got %d", len(observers))
	}
}

func TestWithoutEventsDisablesEvents(t *testing.T) {
	w := openSQLiteQuery(t)

	newQ := w.Q.WithoutEvents()
	wrapped := query.WrapQuery(newQ.(*query.Query))

	if !wrapped.GetWithoutEvents() {
		t.Error("WithoutEvents should set withoutEvents flag")
	}
}

func TestWithoutEventsReturnsNewQuery(t *testing.T) {
	w := openSQLiteQuery(t)

	newQ := w.Q.WithoutEvents()
	if newQ == w.Q {
		t.Error("WithoutEvents should return a new Query instance")
	}

	wrapped := query.WrapQuery(w.Q)
	if wrapped.GetWithoutEvents() {
		t.Error("original query should not have withoutEvents flag set")
	}
}

func TestObserverDispatchDuringCreate(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE observer_test (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)")

	observer := &TestObserver{}
	w.Q.Observe(&TestModel{}, observer)

	model := &TestModel{Name: "test"}
	err := w.Q.Model(model).Create(model)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if !observer.CreatingCalled {
		t.Error("Creating observer was not called during Create")
	}
	if !observer.CreatedCalled {
		t.Error("Created observer was not called during Create")
	}
}

func TestObserverDispatchDuringUpdate(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE observer_test (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)")
	execSQL(t, w, "INSERT INTO observer_test (name) VALUES ('original')")

	observer := &TestObserver{}
	w.Q.Observe(&TestModel{}, observer)

	model := &TestModel{Name: "updated"}
	_, err := w.Q.Model(model).Where("id", 1).Update(model)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if !observer.UpdatingCalled {
		t.Error("Updating observer was not called during Update")
	}
	if !observer.UpdatedCalled {
		t.Error("Updated observer was not called during Update")
	}
}

func TestObserverDispatchDuringDelete(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE observer_test (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)")
	execSQL(t, w, "INSERT INTO observer_test (name) VALUES ('test')")

	observer := &TestObserver{}
	w.Q.Observe(&TestModel{}, observer)

	model := &TestModel{}
	_, err := w.Q.Model(model).Where("id", 1).Delete()
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if !observer.DeletingCalled {
		t.Error("Deleting observer was not called during Delete")
	}
	if !observer.DeletedCalled {
		t.Error("Deleted observer was not called during Delete")
	}
}

func TestObserverDispatchWithoutEvents(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE observer_test (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)")

	observer := &TestObserver{}
	w.Q.Observe(&TestModel{}, observer)

	w.Q = w.Q.WithoutEvents().(*query.Query)

	model := &TestModel{Name: "test"}
	err := w.Q.Model(model).Create(model)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if observer.CreatingCalled {
		t.Error("Creating observer should not be called when WithoutEvents is set")
	}
	if observer.CreatedCalled {
		t.Error("Created observer should not be called when WithoutEvents is set")
	}
}

func TestMultipleObserversDispatchDuringCreate(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE observer_test (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)")

	observer1 := &TestObserver{}
	observer2 := &TestObserver{}
	w.Q.Observe(&TestModel{}, observer1)
	w.Q.Observe(&TestModel{}, observer2)

	model := &TestModel{Name: "test"}
	err := w.Q.Model(model).Create(model)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if !observer1.CreatedCalled {
		t.Error("Created observer was not called on observer1")
	}
	if !observer2.CreatedCalled {
		t.Error("Created observer was not called on observer2")
	}
}
