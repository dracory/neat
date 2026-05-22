package query

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
	q := openSQLiteQuery(t)
	q.dbConfig = twoConnectionDBConfig()

	returned := q.Connection("nonexistent")
	// Unknown name → returns original query unchanged
	if returned != q {
		t.Error("expected Connection('nonexistent') to return the original query")
	}
}

// TestConnectionSwitchReturnsNewQuery verifies that Connection() returns a
// different Query instance from the original for a valid connection name.
func TestConnectionSwitchReturnsNewQuery(t *testing.T) {
	q := openSQLiteQuery(t)
	q.dbConfig = twoConnectionDBConfig()

	newQ := q.Connection("secondary")
	if newQ == nil {
		t.Fatal("expected non-nil query from Connection()")
	}
	if newQ == q {
		t.Error("expected Connection() to return a new Query instance, not the same")
	}
}

// TestConnectionSwitchUsesCorrectDriver verifies that the returned query uses
// the driver configured for the named connection.
func TestConnectionSwitchUsesCorrectDriver(t *testing.T) {
	q := openSQLiteQuery(t)
	q.dbConfig = twoConnectionDBConfig()

	newQ := q.Connection("secondary")
	got := string(newQ.Driver())
	if got != "sqlite" {
		t.Errorf("expected driver 'sqlite' for secondary connection, got %q", got)
	}
}

// TestConnectionSwitchEmptyNameReturnsSelf verifies that Connection("") returns
// the original query.
func TestConnectionSwitchEmptyNameReturnsSelf(t *testing.T) {
	q := openSQLiteQuery(t)
	q.dbConfig = twoConnectionDBConfig()

	returned := q.Connection("")
	if returned != q {
		t.Error("expected Connection('') to return the original query")
	}
}
