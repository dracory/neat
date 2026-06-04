package query

import (
	"context"
	"database/sql"
	"testing"

	"github.com/dracory/neat/contracts/log"
	"github.com/dracory/neat/database/driver"
)

func TestNewQueryInitializesFields(t *testing.T) {
	ctx := context.Background()
	db := &sql.DB{}
	drv := driver.NewSQLite()
	connection := "test"
	dbConfig := MakeDBConfig()
	logger := log.NewNoopLogger()

	q := NewQuery(ctx, db, drv, connection, dbConfig, logger)

	if q.ctx != ctx {
		t.Error("Context not set correctly")
	}
	if q.db != db {
		t.Error("DB not set correctly")
	}
	if q.driver != drv {
		t.Error("Driver not set correctly")
	}
	if q.connection != connection {
		t.Error("Connection not set correctly")
	}
	if q.dbConfig != dbConfig {
		t.Error("DBConfig not set correctly")
	}
	if q.log != logger {
		t.Error("Log not set correctly")
	}
	if q.enableLog {
		t.Error("enableLog should be false by default")
	}
	if q.queryLog == nil {
		t.Error("queryLog should be initialized")
	}
	if q.modelToObserver == nil {
		t.Error("modelToObserver should be initialized")
	}
	if q.withoutEvents {
		t.Error("withoutEvents should be false by default")
	}
	if q.dispatcher == nil {
		t.Error("dispatcher should be initialized")
	}
}

func TestNewQueryWithReplicasInitializesFields(t *testing.T) {
	ctx := context.Background()
	writeConn := &sql.DB{}
	readConn := &sql.DB{}
	drv := driver.NewSQLite()
	connection := "test"
	dbConfig := MakeDBConfig()
	logger := log.NewNoopLogger()

	q := NewQueryWithReplicas(ctx, writeConn, readConn, drv, connection, dbConfig, logger)

	if q.db != writeConn {
		t.Error("Primary DB not set to writeConn")
	}
	if q.readDB != readConn {
		t.Error("readDB not set correctly")
	}
	if q.writeDB != writeConn {
		t.Error("writeDB not set correctly")
	}
}

// TestNewQueryWithReplicasSetsFields is in query_accessors_test.go
// This file contains basic constructor initialization tests
