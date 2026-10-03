package query

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/dracory/neat/database/db"
	"github.com/dracory/neat/database/soft_delete"
)

func newLocationTestBuilder(dialect string, loc *time.Location) *Builder {
	cfg := &db.DBConfig{TimeLocation: loc}
	return NewBuilder(NewQuery(context.TODO(), nil, &FakeDriver{DialectName: dialect}, "users", cfg, nil))
}

func TestTimeLocationDefaultsToUTC(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	if q.timeLocation() != time.UTC {
		t.Errorf("Expected UTC without config, got %v", q.timeLocation())
	}
	q = NewQuery(context.TODO(), nil, nil, "", &db.DBConfig{}, nil)
	if q.timeLocation() != time.UTC {
		t.Errorf("Expected UTC with nil TimeLocation, got %v", q.timeLocation())
	}
}

func TestNormalizeTimeArgUsesConfiguredLocation(t *testing.T) {
	loc := time.FixedZone("UTC+2", 2*3600)
	ts := time.Date(2026, 10, 3, 5, 42, 45, 0, time.UTC)

	if got := newLocationTestBuilder("sqlite", loc).normalizeTimeArg(ts); got != "2026-10-03 07:42:45" {
		t.Errorf("Expected SQLite string in configured location, got %v", got)
	}
	if got := newLocationTestBuilder("sqlite", loc).normalizeTimeArg(&ts); got != "2026-10-03 07:42:45" {
		t.Errorf("Expected SQLite string for *time.Time, got %v", got)
	}
	if got := newLocationTestBuilder("sqlite", loc).normalizeTimeArg(sql.NullTime{Time: ts, Valid: true}); got != "2026-10-03 07:42:45" {
		t.Errorf("Expected SQLite string for NullTime, got %v", got)
	}

	got, ok := newLocationTestBuilder("postgres", loc).normalizeTimeArg(ts).(time.Time)
	if !ok || got.Location() != loc || !got.Equal(ts) {
		t.Errorf("Expected same instant in configured location, got %v", got)
	}
}

func TestNormalizeTimeArgIsIdempotentWithLocation(t *testing.T) {
	loc := time.FixedZone("UTC-5", -5*3600)
	ts := time.Date(2026, 10, 3, 5, 42, 45, 0, time.UTC)

	for _, dialect := range []string{"sqlite", "mysql"} {
		b := newLocationTestBuilder(dialect, loc)
		once := b.normalizeTimeArg(ts)
		twice := b.normalizeTimeArg(once)
		if once != twice {
			t.Errorf("%s: expected idempotent normalization, got %v then %v", dialect, once, twice)
		}
	}
}

func TestBuildUpdateUsesConfiguredLocation(t *testing.T) {
	loc := time.FixedZone("UTC+2", 2*3600)
	b := newLocationTestBuilder("sqlite", loc)

	_, args := b.BuildUpdate(map[string]any{"updated_at": time.Date(2026, 10, 3, 5, 0, 0, 0, time.UTC)})

	if len(args) != 1 || args[0] != "2026-10-03 07:00:00" {
		t.Errorf("Expected time in configured location, got %v", args)
	}
}

func TestBuildUpdateStructIncludesNonNilTimePointer(t *testing.T) {
	type row struct {
		Name      string
		DeletedAt *time.Time `db:"deleted_at"`
	}
	ts := time.Date(2026, 10, 3, 5, 42, 45, 0, time.UTC)
	b := newLocationTestBuilder("sqlite", nil)

	sqlStr, args := b.BuildUpdate(&row{Name: "Ann", DeletedAt: &ts})

	if !strings.Contains(sqlStr, "deleted_at") {
		t.Errorf("Expected deleted_at in SET clause, got %s", sqlStr)
	}
	if len(args) != 2 || args[1] != "2026-10-03 05:42:45" {
		t.Errorf("Expected normalized time pointer arg, got %v", args)
	}

	sqlStr, _ = b.BuildUpdate(&row{Name: "Ann"})
	if strings.Contains(sqlStr, "deleted_at") {
		t.Errorf("Expected nil *time.Time to stay out of SET clause, got %s", sqlStr)
	}
}

func TestBuildInsertStructIncludesNonNilTimePointer(t *testing.T) {
	type row struct {
		Name      string
		DeletedAt *time.Time `db:"deleted_at"`
	}
	ts := time.Date(2026, 10, 3, 5, 42, 45, 0, time.UTC)
	b := newLocationTestBuilder("sqlite", nil)

	sqlStr, args := b.BuildInsert(&row{Name: "Ann", DeletedAt: &ts})

	if !strings.Contains(sqlStr, "deleted_at") || len(args) != 2 || args[1] != "2026-10-03 05:42:45" {
		t.Errorf("Expected deleted_at insert with normalized time, got %s %v", sqlStr, args)
	}
}

// strategyTestModel uses the max-date soft delete strategy, which binds one
// parameter in its soft-delete condition.
type strategyTestModel struct {
	ID   int
	Name string
	soft_delete.SoftDeletesMaxDate
}

func TestSelectSoftDeleteStrategyPlaceholdersPostgres(t *testing.T) {
	q := NewQuery(context.TODO(), nil, &FakeDriver{DialectName: "postgres"}, "items", nil, nil)
	q.model = &strategyTestModel{}
	q.Where("name = ?", "Ann")
	b := NewBuilder(q)

	where, args := b.buildWheresWithSoftDelete()

	if strings.Contains(where, "?") {
		t.Errorf("Expected no literal ? placeholders for postgres, got %s", where)
	}
	if !strings.Contains(where, "$1") || !strings.Contains(where, "= $2") {
		t.Errorf("Expected sequential $1 (soft delete) and $2 (name), got %s", where)
	}
	if len(args) != 2 || args[1] != "Ann" {
		t.Errorf("Expected soft-delete arg followed by name arg, got %v", args)
	}
}
