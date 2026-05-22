package query_test

import (
	"testing"

	"github.com/dracory/neat/database/db"
)

func twoConnectionDBConfig() *db.DBConfig {
	return &db.DBConfig{
		Default: "primary",
		Connections: map[string]db.ConnectionConfig{
			"primary":   {Driver: "sqlite", Database: ":memory:"},
			"secondary": {Driver: "sqlite", Database: ":memory:"},
		},
	}
}

// TestConnectionSwitchUnknownNameReturnsSelf verifies that Connection() with an
// unknown name returns the original query (current behaviour — no error signal).
func TestConnectionSwitchUnknownNameReturnsSelf(t *testing.T) {
	w := openSQLiteQuery(t)
	w.SetDBConfig(twoConnectionDBConfig())

	returned := w.Q.Connection("nonexistent")
	if returned != w.Q {
		t.Error("expected Connection('nonexistent') to return the original query")
	}
}

// TestConnectionSwitchReturnsNewQuery verifies that Connection() returns a
// different Query instance from the original for a valid connection name.
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

// TestConnectionSwitchUsesCorrectDriver verifies that the returned query uses
// the driver configured for the named connection.
func TestConnectionSwitchUsesCorrectDriver(t *testing.T) {
	w := openSQLiteQuery(t)
	w.SetDBConfig(twoConnectionDBConfig())

	newQ := w.Q.Connection("secondary")
	got := string(newQ.Driver())
	if got != "sqlite" {
		t.Errorf("expected driver 'sqlite' for secondary connection, got %q", got)
	}
}

// TestConnectionSwitchEmptyNameReturnsSelf verifies that Connection("") returns
// the original query.
func TestConnectionSwitchEmptyNameReturnsSelf(t *testing.T) {
	w := openSQLiteQuery(t)
	w.SetDBConfig(twoConnectionDBConfig())

	returned := w.Q.Connection("")
	if returned != w.Q {
		t.Error("expected Connection('') to return the original query")
	}
}
