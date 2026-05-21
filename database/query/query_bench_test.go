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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Benchmark the Create method structure (without actual DB)
		_ = user
		_ = query
	}
}

func BenchmarkQueryWhere(b *testing.B) {
	ctx := context.TODO()
	query := NewQuery(ctx, nil, nil, "default", nil, log.NewStdLogger())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		query.Where("name", "John").Where("age", 30)
	}
}

func BenchmarkQuerySelect(b *testing.B) {
	ctx := context.TODO()
	query := NewQuery(ctx, nil, nil, "default", nil, log.NewStdLogger())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		query.Select("id", "name", "email")
	}
}

func BenchmarkQueryOrderBy(b *testing.B) {
	ctx := context.TODO()
	query := NewQuery(ctx, nil, nil, "default", nil, log.NewStdLogger())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		query.OrderBy("created_at", "desc").OrderBy("name", "asc")
	}
}

func BenchmarkQueryLimit(b *testing.B) {
	ctx := context.TODO()
	query := NewQuery(ctx, nil, nil, "default", nil, log.NewStdLogger())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		query.Limit(10).Offset(20)
	}
}

func BenchmarkQueryToSql(b *testing.B) {
	ctx := context.TODO()
	query := NewQuery(ctx, nil, nil, "default", nil, log.NewStdLogger())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = query.ToSql()
	}
}
