package query

import (
	"context"
	"testing"

	"github.com/dracory/neat/contracts/log"
)

type BenchmarkUser struct {
	ID    uint
	Name  string
	Email string
}

func BenchmarkQueryCreate(b *testing.B) {
	ctx := context.Background()
	query := NewQuery(ctx, nil, nil, "default", nil, log.NewStdLogger())
	user := BenchmarkUser{Name: "Test User", Email: "test@example.com"}

	for b.Loop() {
		// Benchmark the Create method structure (without actual DB)
		_ = user
		_ = query
	}
}

func BenchmarkQueryWhere(b *testing.B) {
	ctx := context.TODO()
	query := NewQuery(ctx, nil, nil, "default", nil, log.NewStdLogger())

	for b.Loop() {
		query.Where("name", "John").Where("age", 30)
	}
}

func BenchmarkQuerySelect(b *testing.B) {
	ctx := context.TODO()
	query := NewQuery(ctx, nil, nil, "default", nil, log.NewStdLogger())

	for b.Loop() {
		query.Select("id", "name", "email")
	}
}

func BenchmarkQueryOrderBy(b *testing.B) {
	ctx := context.TODO()
	query := NewQuery(ctx, nil, nil, "default", nil, log.NewStdLogger())

	for b.Loop() {
		query.OrderBy("created_at", "desc").OrderBy("name", "asc")
	}
}

func BenchmarkQueryLimit(b *testing.B) {
	ctx := context.TODO()
	query := NewQuery(ctx, nil, nil, "default", nil, log.NewStdLogger())

	for b.Loop() {
		query.Limit(10).Offset(20)
	}
}

func BenchmarkQueryToSql(b *testing.B) {
	ctx := context.TODO()
	query := NewQuery(ctx, nil, nil, "default", nil, log.NewStdLogger())

	for b.Loop() {
		_ = query.ToSql()
	}
}
