package query_test

import (
	"testing"

	"github.com/dracory/neat/contracts/database/orm"
)

type txLeakRow struct {
	ID   string `db:"id"`
	Name string `db:"name"`
}

// TestTransactionStatementsDoNotAccumulateWheres proves that statements
// executed on the tx query inside Transaction() must not leak WHERE clauses
// into each other. The tx object is a single mutable *Query, so chaining
// Table().Where() twice currently produces
//
//	SELECT ... WHERE id = 'A' AND id = 'B'
//
// for the second statement — matching zero rows. This test FAILS until the
// tx builder state is isolated per statement (or callers clone per statement).
func TestTransactionStatementsDoNotAccumulateWheres(t *testing.T) {
	w := openSQLiteQuery(t)

	execSQL(t, w, `CREATE TABLE items (id TEXT PRIMARY KEY, name TEXT)`)
	execSQL(t, w, `INSERT INTO items (id, name) VALUES ('A', 'alpha'), ('B', 'beta')`)

	var rowsA, rowsB []txLeakRow

	err := w.Q.Transaction(func(tx orm.Query) error {
		if err := tx.Table("items").Where("id = ?", "A").Get(&rowsA); err != nil {
			return err
		}
		return tx.Table("items").Where("id = ?", "B").Get(&rowsB)
	})
	if err != nil {
		t.Fatalf("transaction failed: %v", err)
	}

	if len(rowsA) != 1 || rowsA[0].Name != "alpha" {
		t.Fatalf("statement 1: expected row A/alpha, got %+v", rowsA)
	}
	// The bug: second statement runs WHERE id='A' AND id='B' → 0 rows.
	if len(rowsB) != 1 || rowsB[0].Name != "beta" {
		t.Fatalf("statement 2: expected row B/beta, got %+v — tx WHERE clauses leaked across statements", rowsB)
	}
}

// TestTransactionClonePerStatementIsolates documents the workaround: cloning
// the tx query per statement isolates builder state. This must keep passing.
func TestTransactionClonePerStatementIsolates(t *testing.T) {
	w := openSQLiteQuery(t)

	execSQL(t, w, `CREATE TABLE items (id TEXT PRIMARY KEY, name TEXT)`)
	execSQL(t, w, `INSERT INTO items (id, name) VALUES ('A', 'alpha'), ('B', 'beta')`)

	var rowsA, rowsB []txLeakRow

	err := w.Q.Transaction(func(tx orm.Query) error {
		cloneable, ok := tx.(interface{ Clone() orm.Query })
		if !ok {
			t.Fatal("tx query does not support Clone()")
		}
		if err := cloneable.Clone().Table("items").Where("id = ?", "A").Get(&rowsA); err != nil {
			return err
		}
		return cloneable.Clone().Table("items").Where("id = ?", "B").Get(&rowsB)
	})
	if err != nil {
		t.Fatalf("transaction failed: %v", err)
	}

	if len(rowsA) != 1 || rowsA[0].Name != "alpha" {
		t.Fatalf("statement 1: expected row A/alpha, got %+v", rowsA)
	}
	if len(rowsB) != 1 || rowsB[0].Name != "beta" {
		t.Fatalf("statement 2: expected row B/beta, got %+v", rowsB)
	}
}
