package driver

import (
	"context"
	"database/sql"
	"testing"
)

// MockDriver is a mock implementation of the Driver interface for testing.
type MockDriver struct {
	OpenFunc       func(dsn string) (*sql.DB, error)
	CloseFunc      func(db *sql.DB) error
	PingFunc       func(ctx context.Context, db *sql.DB) error
	BeginTxFunc    func(ctx context.Context, db *sql.DB, opts *sql.TxOptions) (*sql.Tx, error)
	PlaceholderFunc func(n int) string
	DialectFunc    func() string
}

func (m *MockDriver) Open(dsn string) (*sql.DB, error) {
	if m.OpenFunc != nil {
		return m.OpenFunc(dsn)
	}
	return nil, nil
}

func (m *MockDriver) Close(db *sql.DB) error {
	if m.CloseFunc != nil {
		return m.CloseFunc(db)
	}
	return nil
}

func (m *MockDriver) Ping(ctx context.Context, db *sql.DB) error {
	if m.PingFunc != nil {
		return m.PingFunc(ctx, db)
	}
	return nil
}

func (m *MockDriver) BeginTx(ctx context.Context, db *sql.DB, opts *sql.TxOptions) (*sql.Tx, error) {
	if m.BeginTxFunc != nil {
		return m.BeginTxFunc(ctx, db, opts)
	}
	return nil, nil
}

func (m *MockDriver) Placeholder(n int) string {
	if m.PlaceholderFunc != nil {
		return m.PlaceholderFunc(n)
	}
	return "?"
}

func (m *MockDriver) Dialect() string {
	if m.DialectFunc != nil {
		return m.DialectFunc()
	}
	return "mock"
}

func TestDriverInterface(t *testing.T) {
	// Test that MockDriver implements the Driver interface
	var _ Driver = (*MockDriver)(nil)
}

func TestMockDriverOpen(t *testing.T) {
	called := false
	mock := &MockDriver{
		OpenFunc: func(dsn string) (*sql.DB, error) {
			called = true
			return nil, nil
		},
	}

	mock.Open("test-dsn")
	if !called {
		t.Error("OpenFunc was not called")
	}
}

func TestMockDriverClose(t *testing.T) {
	called := false
	mock := &MockDriver{
		CloseFunc: func(db *sql.DB) error {
			called = true
			return nil
		},
	}

	mock.Close(nil)
	if !called {
		t.Error("CloseFunc was not called")
	}
}

func TestMockDriverPing(t *testing.T) {
	called := false
	mock := &MockDriver{
		PingFunc: func(ctx context.Context, db *sql.DB) error {
			called = true
			return nil
		},
	}

	mock.Ping(context.Background(), nil)
	if !called {
		t.Error("PingFunc was not called")
	}
}

func TestMockDriverBeginTx(t *testing.T) {
	called := false
	mock := &MockDriver{
		BeginTxFunc: func(ctx context.Context, db *sql.DB, opts *sql.TxOptions) (*sql.Tx, error) {
			called = true
			return nil, nil
		},
	}

	mock.BeginTx(context.Background(), nil, nil)
	if !called {
		t.Error("BeginTxFunc was not called")
	}
}

func TestMockDriverPlaceholder(t *testing.T) {
	mock := &MockDriver{
		PlaceholderFunc: func(n int) string {
			return "$" + string(rune('0'+n))
		},
	}

	result := mock.Placeholder(1)
	if result != "$1" {
		t.Errorf("Expected $1, got %s", result)
	}
}

func TestMockDriverDialect(t *testing.T) {
	mock := &MockDriver{
		DialectFunc: func() string {
			return "test-dialect"
		},
	}

	result := mock.Dialect()
	if result != "test-dialect" {
		t.Errorf("Expected test-dialect, got %s", result)
	}
}

func TestMockDriverDefaultBehavior(t *testing.T) {
	mock := &MockDriver{}

	// Test default behaviors when no custom functions are provided
	db, err := mock.Open("test-dsn")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if db != nil {
		t.Error("Expected nil DB")
	}

	err = mock.Close(nil)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	err = mock.Ping(context.Background(), nil)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	tx, err := mock.BeginTx(context.Background(), nil, nil)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if tx != nil {
		t.Error("Expected nil Tx")
	}

	placeholder := mock.Placeholder(1)
	if placeholder != "?" {
		t.Errorf("Expected ?, got %s", placeholder)
	}

	dialect := mock.Dialect()
	if dialect != "mock" {
		t.Errorf("Expected mock, got %s", dialect)
	}
}
