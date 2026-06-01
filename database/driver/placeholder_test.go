package driver

import (
	"context"
	"testing"
)

func TestPlaceholderFuncs(t *testing.T) {
	tests := []struct {
		dialect string
		n       int
		want    string
	}{
		{"mysql", 1, "?"},
		{"mysql", 5, "?"},
		{"postgres", 1, "$1"},
		{"postgres", 3, "$3"},
		{"sqlite", 1, "?"},
		{"sqlite", 9, "?"},
		{"sqlserver", 1, "@p1"},
		{"sqlserver", 4, "@p4"},
		{"turso", 1, "?"},
		{"turso", 2, "?"},
	}
	for _, tt := range tests {
		fn, ok := PlaceholderFuncs[tt.dialect]
		if !ok {
			t.Errorf("PlaceholderFuncs missing dialect %q", tt.dialect)
			continue
		}
		if got := fn(tt.n); got != tt.want {
			t.Errorf("PlaceholderFuncs[%q](%d) = %q, want %q", tt.dialect, tt.n, got, tt.want)
		}
	}
}

func TestGetPlaceholderFuncKnownDialects(t *testing.T) {
	cases := []struct {
		dialect string
		n       int
		want    string
	}{
		{"mysql", 1, "?"},
		{"postgres", 2, "$2"},
		{"sqlite", 1, "?"},
		{"sqlserver", 3, "@p3"},
		{"turso", 1, "?"},
	}
	for _, c := range cases {
		fn := GetPlaceholderFunc(c.dialect)
		if fn == nil {
			t.Errorf("GetPlaceholderFunc(%q) returned nil", c.dialect)
			continue
		}
		if got := fn(c.n); got != c.want {
			t.Errorf("GetPlaceholderFunc(%q)(%d) = %q, want %q", c.dialect, c.n, got, c.want)
		}
	}
}

func TestGetPlaceholderFuncUnknownFallback(t *testing.T) {
	fn := GetPlaceholderFunc("unknown_dialect")
	if fn == nil {
		t.Fatal("expected non-nil fallback func")
	}
	if got := fn(1); got != "?" {
		t.Errorf("fallback placeholder = %q, want ?", got)
	}
}

func TestSQLiteDriverDialectAndPlaceholder(t *testing.T) {
	d := NewSQLite()
	if d.Dialect() != "sqlite" {
		t.Errorf("Dialect = %q, want sqlite", d.Dialect())
	}
	for _, n := range []int{1, 5, 100} {
		if got := d.Placeholder(n); got != "?" {
			t.Errorf("Placeholder(%d) = %q, want ?", n, got)
		}
	}
}

func TestMySQLDriverDialectAndPlaceholder(t *testing.T) {
	d := NewMySQL()
	if d.Dialect() != "mysql" {
		t.Errorf("Dialect = %q, want mysql", d.Dialect())
	}
	for _, n := range []int{1, 5, 100} {
		if got := d.Placeholder(n); got != "?" {
			t.Errorf("Placeholder(%d) = %q, want ?", n, got)
		}
	}
}

func TestPostgreSQLDriverDialectAndPlaceholder(t *testing.T) {
	d := NewPostgreSQL()
	if d.Dialect() != "postgres" {
		t.Errorf("Dialect = %q, want postgres", d.Dialect())
	}
	cases := []struct {
		n    int
		want string
	}{{1, "$1"}, {3, "$3"}, {10, "$10"}}
	for _, c := range cases {
		if got := d.Placeholder(c.n); got != c.want {
			t.Errorf("Placeholder(%d) = %q, want %q", c.n, got, c.want)
		}
	}
}

func TestSQLServerDriverDialectAndPlaceholder(t *testing.T) {
	d := NewSQLServer()
	if d.Dialect() != "sqlserver" {
		t.Errorf("Dialect = %q, want sqlserver", d.Dialect())
	}
	cases := []struct {
		n    int
		want string
	}{{1, "@p1"}, {2, "@p2"}, {10, "@p10"}}
	for _, c := range cases {
		if got := d.Placeholder(c.n); got != c.want {
			t.Errorf("Placeholder(%d) = %q, want %q", c.n, got, c.want)
		}
	}
}

func TestSQLiteDriverOpen(t *testing.T) {
	d := NewSQLite()
	db, err := d.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = d.Close(db) }()

	ctx := context.Background()
	if err := d.Ping(ctx, db); err != nil {
		t.Errorf("Ping: %v", err)
	}

	tx, err := d.BeginTx(ctx, db, nil)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}
	_ = tx.Rollback()
}

func TestSQLiteDriverClose(t *testing.T) {
	d := NewSQLite()
	db, err := d.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := d.Close(db); err != nil {
		t.Errorf("Close: %v", err)
	}
}

func TestMySQLDriverOpenReturnsDB(t *testing.T) {
	d := NewMySQL()
	// sql.Open is lazy — it returns a *sql.DB without connecting.
	db, err := d.Open("root:@tcp(127.0.0.1:3306)/nonexistent")
	if err != nil {
		t.Fatalf("Open (lazy): %v", err)
	}
	if db == nil {
		t.Fatal("expected non-nil *sql.DB")
	}
	_ = d.Close(db)
}

func TestPostgreSQLDriverOpenReturnsDB(t *testing.T) {
	d := NewPostgreSQL()
	db, err := d.Open("host=127.0.0.1 port=5432 dbname=nonexistent sslmode=disable")
	if err != nil {
		t.Fatalf("Open (lazy): %v", err)
	}
	if db == nil {
		t.Fatal("expected non-nil *sql.DB")
	}
	_ = d.Close(db)
}

func TestSQLServerDriverOpenReturnsDB(t *testing.T) {
	d := NewSQLServer()
	db, err := d.Open("sqlserver://sa:password@127.0.0.1:1433?database=nonexistent")
	if err != nil {
		t.Fatalf("Open (lazy): %v", err)
	}
	if db == nil {
		t.Fatal("expected non-nil *sql.DB")
	}
	_ = d.Close(db)
}
