package query_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/database/query"
	_ "modernc.org/sqlite"
)

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
