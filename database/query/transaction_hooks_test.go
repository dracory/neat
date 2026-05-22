package query_test

import (
	"errors"
	"testing"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/database/query"
)

// openSQLiteForTx returns a TestQuery wrapper with a simple table already created.
func openSQLiteForTx(t *testing.T) *query.TestQuery {
	t.Helper()
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE tx_hooks (id INTEGER, val TEXT)")
	return w
}

func TestBeforeCommitCalledOnCommit(t *testing.T) {
	w := openSQLiteForTx(t)

	called := false
	err := w.Q.Transaction(func(tx contractsorm.Query) error {
		tx.(*query.Query).BeforeCommit(func() error {
			called = true
			return nil
		})
		return nil
	})
	if err != nil {
		t.Fatalf("Transaction error: %v", err)
	}
	if !called {
		t.Error("expected BeforeCommit callback to be called")
	}
}

func TestAfterCommitCalledOnCommit(t *testing.T) {
	w := openSQLiteForTx(t)

	called := false
	err := w.Q.Transaction(func(tx contractsorm.Query) error {
		tx.(*query.Query).AfterCommit(func() error {
			called = true
			return nil
		})
		return nil
	})
	if err != nil {
		t.Fatalf("Transaction error: %v", err)
	}
	if !called {
		t.Error("expected AfterCommit callback to be called")
	}
}

func TestBeforeRollbackCalledOnRollback(t *testing.T) {
	w := openSQLiteForTx(t)

	called := false
	_ = w.Q.Transaction(func(tx contractsorm.Query) error {
		tx.(*query.Query).BeforeRollback(func() error {
			called = true
			return nil
		})
		return errors.New("force rollback")
	})
	if !called {
		t.Error("expected BeforeRollback callback to be called on rollback")
	}
}

func TestAfterRollbackCalledOnRollback(t *testing.T) {
	w := openSQLiteForTx(t)

	called := false
	_ = w.Q.Transaction(func(tx contractsorm.Query) error {
		tx.(*query.Query).AfterRollback(func() error {
			called = true
			return nil
		})
		return errors.New("force rollback")
	})
	if !called {
		t.Error("expected AfterRollback callback to be called on rollback")
	}
}

func TestBeforeCommitErrorAbortsCommit(t *testing.T) {
	w := openSQLiteForTx(t)

	hookErr := errors.New("hook abort")
	err := w.Q.Transaction(func(tx contractsorm.Query) error {
		tx.(*query.Query).BeforeCommit(func() error {
			return hookErr
		})
		return nil
	})
	if err == nil {
		t.Fatal("expected transaction to fail when BeforeCommit returns error")
	}
	if !errors.Is(err, hookErr) {
		t.Errorf("expected error to wrap hookErr, got: %v", err)
	}
}

func TestTransactionCommitSucceeds(t *testing.T) {
	w := openSQLiteForTx(t)

	err := w.Q.Transaction(func(tx contractsorm.Query) error {
		return nil
	})
	if err != nil {
		t.Fatalf("clean transaction should not fail: %v", err)
	}
}
