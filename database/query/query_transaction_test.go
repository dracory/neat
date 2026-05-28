package query

import (
	"errors"
	"testing"

	_ "modernc.org/sqlite"
)

func TestBeforeCommitHooks(t *testing.T) {
	q := NewQuery(nil, nil, nil, "", nil, nil)

	called := false
	q.beforeCommit = []func() error{
		func() error {
			called = true
			return nil
		},
	}

	if len(q.beforeCommit) != 1 {
		t.Error("Expected 1 beforeCommit hook")
	}

	err := q.beforeCommit[0]()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !called {
		t.Error("Expected beforeCommit hook to be called")
	}
}

func TestAfterCommitHooks(t *testing.T) {
	q := NewQuery(nil, nil, nil, "", nil, nil)

	called := false
	q.afterCommit = []func() error{
		func() error {
			called = true
			return nil
		},
	}

	if len(q.afterCommit) != 1 {
		t.Error("Expected 1 afterCommit hook")
	}

	err := q.afterCommit[0]()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !called {
		t.Error("Expected afterCommit hook to be called")
	}
}

func TestBeforeCommitHookError(t *testing.T) {
	q := NewQuery(nil, nil, nil, "", nil, nil)

	testErr := errors.New("hook error")
	q.beforeCommit = []func() error{
		func() error {
			return testErr
		},
	}

	err := q.beforeCommit[0]()
	if err != testErr {
		t.Errorf("Expected hook error, got %v", err)
	}
}

func TestAfterCommitHookError(t *testing.T) {
	q := NewQuery(nil, nil, nil, "", nil, nil)

	testErr := errors.New("hook error")
	q.afterCommit = []func() error{
		func() error {
			return testErr
		},
	}

	err := q.afterCommit[0]()
	if err != testErr {
		t.Errorf("Expected hook error, got %v", err)
	}
}
