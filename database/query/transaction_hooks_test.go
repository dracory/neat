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

func TestNestedTransactionCommit(t *testing.T) {
	w := openSQLiteForTx(t)

	err := w.Q.Transaction(func(tx contractsorm.Query) error {
		return tx.Transaction(func(innerTx contractsorm.Query) error {
			return nil
		})
	})
	if err != nil {
		t.Fatalf("nested transaction commit should not fail: %v", err)
	}
}

func TestNestedTransactionRollback(t *testing.T) {
	w := openSQLiteForTx(t)

	err := w.Q.Transaction(func(tx contractsorm.Query) error {
		_ = tx.Transaction(func(innerTx contractsorm.Query) error {
			return errors.New("force inner rollback")
		})
		return nil
	})
	if err != nil {
		t.Fatalf("nested transaction with inner rollback should not fail: %v", err)
	}
}

func TestNestedTransactionWithOperations(t *testing.T) {
	w := openSQLiteForTx(t)

	err := w.Q.Transaction(func(tx contractsorm.Query) error {
		_, err := tx.Exec("INSERT INTO tx_hooks (id, val) VALUES (1, 'outer')")
		if err != nil {
			return err
		}

		return tx.Transaction(func(innerTx contractsorm.Query) error {
			_, err := innerTx.Exec("INSERT INTO tx_hooks (id, val) VALUES (2, 'inner')")
			return err
		})
	})
	if err != nil {
		t.Fatalf("nested transaction with operations should not fail: %v", err)
	}

	// Verify both records were committed
	var count int
	db, err := w.Q.DB()
	if err != nil {
		t.Fatalf("failed to get DB: %v", err)
	}
	err = db.QueryRow("SELECT COUNT(*) FROM tx_hooks").Scan(&count)
	if err != nil {
		t.Fatalf("failed to count records: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 records, got %d", count)
	}
}
