// Package benchmarks contains cross-library performance comparisons between
// neat and GORM. It lives in a separate package so that the SQLite driver
// registrations used by each library do not conflict at init time.
//
// GORM is configured with the MySQL dialector in DryRun mode. DryRun mode
// only generates SQL without executing it, so no real database connection
// is required. The MySQL driver is used (rather than SQLite) to avoid
// driver-name conflicts with modernc.org/sqlite used by neat.
package benchmarks

import (
	"context"
	"database/sql"
	"testing"

	"github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/contracts/log"
	neatquery "github.com/dracory/neat/database/query"
	gosql "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// gormUser mirrors the neat benchmark model for fair comparisons.
type gormUser struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"column:name"`
	Email string `gorm:"column:email"`
	Age   int    `gorm:"column:age"`
}

func (gormUser) TableName() string {
	return "users"
}

// newGormDryRunDB returns a GORM DB configured in DryRun mode so that
// query building is exercised without touching a real database.
// A pre-opened *sql.DB is supplied as the ConnPool and version probing
// is skipped, so no network connection is ever attempted.
func newGormDryRunDB(b *testing.B) *gorm.DB {
	b.Helper()
	cfg := gosql.NewConfig()
	cfg.Net = "tcp"
	cfg.Addr = "127.0.0.1:3306"
	cfg.DBName = "test"
	cfg.User = "root"
	cfg.Passwd = "root"
	dsn := cfg.FormatDSN()

	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		b.Fatalf("failed to sql.Open: %v", err)
	}

	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{
		Logger:               gormlogger.Default.LogMode(gormlogger.Silent),
		DryRun:               true,
		DisableAutomaticPing: true,
	})
	if err != nil {
		b.Fatalf("failed to open gorm: %v", err)
	}
	return db
}

// newNeatQuery returns a neat Query with no DB connection, suitable for
// query-building-only benchmarks.
func newNeatQuery(b *testing.B) orm.Query {
	b.Helper()
	return neatquery.NewQuery(context.Background(), nil, nil, "default", nil, log.NewNoopLogger())
}

// --- Simple WHERE chain ---

func BenchmarkNeat_WhereChain(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = newNeatQuery(b).Table("users").
			Where("name", "John").
			Where("age", 30).
			Where("active", true)
	}
}

func BenchmarkGorm_WhereChain(b *testing.B) {
	db := newGormDryRunDB(b)

	b.ReportAllocs()
	for b.Loop() {
		tx := db.Session(&gorm.Session{DryRun: true}).
			Model(&gormUser{}).
			Where("name = ?", "John").
			Where("age = ?", 30).
			Where("active = ?", true)
		_ = tx.Find(&gormUser{}).Statement.SQL.String()
	}
}

// --- Select with columns ---

func BenchmarkNeat_Select(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = newNeatQuery(b).Table("users").Select("id", "name", "email")
	}
}

func BenchmarkGorm_Select(b *testing.B) {
	db := newGormDryRunDB(b)

	b.ReportAllocs()
	for b.Loop() {
		tx := db.Session(&gorm.Session{DryRun: true}).
			Model(&gormUser{}).
			Select("id", "name", "email")
		_ = tx.Find(&gormUser{}).Statement.SQL.String()
	}
}

// --- OrderBy + Limit + Offset ---

func BenchmarkNeat_OrderByLimitOffset(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = newNeatQuery(b).Table("users").
			OrderBy("created_at", "desc").
			OrderBy("name", "asc").
			Limit(10).
			Offset(20)
	}
}

func BenchmarkGorm_OrderByLimitOffset(b *testing.B) {
	db := newGormDryRunDB(b)

	b.ReportAllocs()
	for b.Loop() {
		tx := db.Session(&gorm.Session{DryRun: true}).
			Model(&gormUser{}).
			Order("created_at desc").
			Order("name asc").
			Limit(10).
			Offset(20)
		_ = tx.Find(&gormUser{}).Statement.SQL.String()
	}
}

// --- Complex query: multiple WHERE + ORDER BY + LIMIT ---

func BenchmarkNeat_ComplexQuery(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		q := newNeatQuery(b).Table("users").
			Where("age", ">=", 25).
			Where("active", true).
			Where("score", ">", 50).
			OrderBy("created_at", "desc").
			Limit(100)
		_ = q.ToSql().Get(nil)
	}
}

func BenchmarkGorm_ComplexQuery(b *testing.B) {
	db := newGormDryRunDB(b)

	b.ReportAllocs()
	for b.Loop() {
		tx := db.Session(&gorm.Session{DryRun: true}).
			Model(&gormUser{}).
			Where("age >= ?", 25).
			Where("active = ?", true).
			Where("score > ?", 50).
			Order("created_at desc").
			Limit(100)
		_ = tx.Find(&gormUser{}).Statement.SQL.String()
	}
}

// --- ToSql / SQL generation ---

func BenchmarkNeat_ToSql(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		q := newNeatQuery(b).Table("users").
			Where("name", "John").
			Where("age", 30).
			OrderBy("id", "desc").
			Limit(50)
		_ = q.ToSql().Get(nil)
	}
}

func BenchmarkGorm_ToSql(b *testing.B) {
	db := newGormDryRunDB(b)

	b.ReportAllocs()
	for b.Loop() {
		tx := db.Session(&gorm.Session{DryRun: true}).
			Model(&gormUser{}).
			Where("name = ?", "John").
			Where("age = ?", 30).
			Order("id desc").
			Limit(50)
		_ = tx.Find(&gormUser{}).Statement.SQL.String()
	}
}

// --- Fresh build per iteration (realistic per-request usage) ---

func BenchmarkNeat_FreshBuildPerIteration(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		q := newNeatQuery(b).Table("users").
			Where("name", "John").
			Where("age", 30).
			Where("active", true).
			OrderBy("id", "desc").
			Limit(50)
		_ = q.ToSql().Get(nil)
	}
}

func BenchmarkGorm_FreshBuildPerIteration(b *testing.B) {
	db := newGormDryRunDB(b)

	b.ReportAllocs()
	for b.Loop() {
		tx := db.Session(&gorm.Session{DryRun: true}).
			Model(&gormUser{}).
			Where("name = ?", "John").
			Where("age = ?", 30).
			Where("active = ?", true).
			Order("id desc").
			Limit(50)
		_ = tx.Find(&gormUser{}).Statement.SQL.String()
	}
}
