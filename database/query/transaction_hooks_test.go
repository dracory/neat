package query

import (
	"errors"
	"testing"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
)

// openSQLiteForTx returns a query backed by an in-memory SQLite DB with a simple table.
func openSQLiteForTx(t *testing.T) *Query {
	t.Helper()
	q := openSQLiteQuery(t)
	execSQL(t, q, "CREATE TABLE tx_hooks (id INTEGER, val TEXT)")
	return q
}

func TestBeforeCommitCalledOnCommit(t *testing.T) {
	q := openSQLiteForTx(t)

	called := false
	err := q.Transaction(func(tx contractsorm.Query) error {
		tx.(*Query).BeforeCommit(func() error {
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
	q := openSQLiteForTx(t)

	called := false
	err := q.Transaction(func(tx contractsorm.Query) error {
		tx.(*Query).AfterCommit(func() error {
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
	q := openSQLiteForTx(t)

	called := false
	_ = q.Transaction(func(tx contractsorm.Query) error {
		tx.(*Query).BeforeRollback(func() error {
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
	q := openSQLiteForTx(t)

	called := false
	_ = q.Transaction(func(tx contractsorm.Query) error {
		tx.(*Query).AfterRollback(func() error {
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
	q := openSQLiteForTx(t)

	hookErr := errors.New("hook abort")
	err := q.Transaction(func(tx contractsorm.Query) error {
		tx.(*Query).BeforeCommit(func() error {
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
	q := openSQLiteForTx(t)

	err := q.Transaction(func(tx contractsorm.Query) error {
		return nil
	})
	if err != nil {
		t.Fatalf("clean transaction should not fail: %v", err)
	}
}
