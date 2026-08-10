package driver

import (
	"strings"
	"testing"
	"time"
)

type godbTestBlog struct {
	ID         int64  `db:"id"`
	Title      string `db:"title"`
	CategoryID int64  `db:"category_id"`
}

type godbTestCategory struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

func TestGODB_Dialect(t *testing.T) {
	g := NewGODB()
	if g.Dialect() != "sqlite" {
		t.Errorf("expected dialect 'sqlite', got '%s'", g.Dialect())
	}
}

func TestGODB_Open_Empty(t *testing.T) {
	g := NewGODB()
	g.SetTables(nil)

	db, err := g.Open("")
	if err != nil {
		t.Fatalf("failed to open empty GODB: %v", err)
	}
	defer func() { _ = db.Close() }()

	if err := db.Ping(); err != nil {
		t.Errorf("ping failed: %v", err)
	}
}

func TestGODB_Open_StyleAMap(t *testing.T) {
	g := NewGODB()

	blogs := []godbTestBlog{
		{ID: 1, Title: "Hello World", CategoryID: 10},
		{ID: 2, Title: "Go Tips", CategoryID: 20},
	}
	categories := []godbTestCategory{
		{ID: 10, Name: "General"},
		{ID: 20, Name: "Programming"},
	}

	g.SetTables(Tables{
		"blogs":      blogs,
		"categories": categories,
	})

	db, err := g.Open("")
	if err != nil {
		t.Fatalf("failed to open GODB: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Verify tables are created and populated
	var blogCount int
	err = db.QueryRow("SELECT COUNT(*) FROM blogs").Scan(&blogCount)
	if err != nil {
		t.Fatalf("failed to query blogs: %v", err)
	}
	if blogCount != 2 {
		t.Errorf("expected 2 blogs, got %d", blogCount)
	}

	var catCount int
	err = db.QueryRow("SELECT COUNT(*) FROM categories").Scan(&catCount)
	if err != nil {
		t.Fatalf("failed to query categories: %v", err)
	}
	if catCount != 2 {
		t.Errorf("expected 2 categories, got %d", catCount)
	}
}

func TestGODB_Open_StyleBSlice(t *testing.T) {
	g := NewGODB()

	blogs := []godbTestBlog{
		{ID: 1, Title: "Hello World", CategoryID: 10},
	}

	g.SetTables([]Table{
		{Name: "blogs", Data: blogs},
	})

	db, err := g.Open("")
	if err != nil {
		t.Fatalf("failed to open GODB: %v", err)
	}
	defer func() { _ = db.Close() }()

	var blogCount int
	err = db.QueryRow("SELECT COUNT(*) FROM blogs").Scan(&blogCount)
	if err != nil {
		t.Fatalf("failed to query blogs: %v", err)
	}
	if blogCount != 1 {
		t.Errorf("expected 1 blog, got %d", blogCount)
	}
}

func TestGODB_Open_TypeMapping(t *testing.T) {
	g := NewGODB()

	type AllTypesStruct struct {
		ID       int64     `db:"id"`
		IntVal   int       `db:"int_val"`
		FloatVal float32   `db:"float_val"`
		BoolVal  bool      `db:"bool_val"`
		StrVal   string    `db:"str_val"`
		TimeVal  time.Time `db:"time_val"`
		BlobVal  []byte    `db:"blob_val"`
	}

	now := time.Now().UTC().Truncate(time.Second)
	data := []AllTypesStruct{
		{
			ID:       1,
			IntVal:   42,
			FloatVal: 3.14,
			BoolVal:  true,
			StrVal:   "Neat",
			TimeVal:  now,
			BlobVal:  []byte("hello binary"),
		},
	}

	g.SetTables(Tables{
		"all_types": data,
	})

	db, err := g.Open("")
	if err != nil {
		t.Fatalf("failed to open GODB: %v", err)
	}
	defer func() { _ = db.Close() }()

	var (
		id       int64
		intVal   int
		floatVal float32
		boolVal  int // SQLite stores bool as 0/1 integer
		strVal   string
		timeStr  string
		blobVal  []byte
	)

	row := db.QueryRow("SELECT id, int_val, float_val, bool_val, str_val, time_val, blob_val FROM all_types WHERE id = 1")
	err = row.Scan(&id, &intVal, &floatVal, &boolVal, &strVal, &timeStr, &blobVal)
	if err != nil {
		t.Fatalf("failed to scan row: %v", err)
	}

	if id != 1 || intVal != 42 || floatVal != 3.14 || boolVal != 1 || strVal != "Neat" {
		t.Errorf("unexpected field values scanned: %d, %d, %f, %d, %s", id, intVal, floatVal, boolVal, strVal)
	}

	parsedTime, err := time.Parse("2006-01-02T15:04:05Z", timeStr)
	if err != nil {
		// Try fallback format
		parsedTime, err = time.Parse("2006-01-02 15:04:05", timeStr)
	}
	if err != nil {
		t.Logf("Raw scanned time string: %q", timeStr)
		// Try RFC3339
		parsedTime, err = time.Parse(time.RFC3339, timeStr)
	}
	if err != nil {
		t.Errorf("failed to parse scanned time: %v", err)
	} else if !parsedTime.Equal(now) {
		t.Errorf("expected time %v, got %v", now, parsedTime)
	}

	if string(blobVal) != "hello binary" {
		t.Errorf("expected blob 'hello binary', got '%s'", string(blobVal))
	}
}

func TestGODB_Open_PointerStructsWithNil(t *testing.T) {
	g := NewGODB()

	blogs := []*godbTestBlog{
		{ID: 1, Title: "Hello", CategoryID: 5},
		nil,
	}

	g.SetTables(Tables{
		"blogs": blogs,
	})

	db, err := g.Open("")
	if err != nil {
		t.Fatalf("failed to open GODB: %v", err)
	}
	defer func() { _ = db.Close() }()

	var blogCount int
	err = db.QueryRow("SELECT COUNT(*) FROM blogs").Scan(&blogCount)
	if err != nil {
		t.Fatalf("failed to query: %v", err)
	}
	if blogCount != 1 {
		t.Errorf("expected 1 blog (nil elements should be skipped), got %d", blogCount)
	}
}

func TestGODB_Open_TableCollision(t *testing.T) {
	g := NewGODB()

	// "users" and "Users" collide case-insensitively
	g.SetTables(Tables{
		"users": []godbTestBlog{{ID: 1}},
		"Users": []godbTestBlog{{ID: 2}},
	})

	db, err := g.Open("")
	if err == nil {
		_ = db.Close()
		t.Fatal("expected case-insensitive table name collision error")
	}
	if !strings.Contains(err.Error(), "table name collision") {
		t.Errorf("expected table name collision error, got: %v", err)
	}
}

func TestGODB_Open_InvalidIdentifiers(t *testing.T) {
	// 1. Invalid Table Name
	g1 := NewGODB()
	g1.SetTables(Tables{
		"invalid-table-name": []godbTestBlog{{ID: 1}},
	})
	db1, err1 := g1.Open("")
	if err1 == nil {
		_ = db1.Close()
		t.Fatal("expected invalid table name error")
	}

	// 2. Invalid Column Name
	g2 := NewGODB()
	type BadStruct struct {
		BadField int `db:"bad-col-name"`
	}
	g2.SetTables(Tables{
		"good_table": []BadStruct{{BadField: 1}},
	})
	db2, err2 := g2.Open("")
	if err2 == nil {
		_ = db2.Close()
		t.Fatal("expected invalid column name error")
	}
}

func TestGODB_Open_RowLimitEnforced(t *testing.T) {
	g := NewGODB()

	// Construct a slice with more than MaxGODBRows rows
	largeData := make([]godbTestBlog, MaxGODBRows+1)
	for i := range largeData {
		largeData[i] = godbTestBlog{ID: int64(i)}
	}

	g.SetTables(Tables{
		"large_table": largeData,
	})

	db, err := g.Open("")
	if err == nil {
		_ = db.Close()
		t.Fatal("expected MaxGODBRows limit error")
	}
	if !strings.Contains(err.Error(), "exceeding the limit") {
		t.Errorf("expected row limit error, got: %v", err)
	}
}

func TestGODB_Open_EmptyDatasetSilentSkip(t *testing.T) {
	g := NewGODB()

	var emptyBlogs []godbTestBlog
	g.SetTables(Tables{
		"blogs": emptyBlogs,
	})

	db, err := g.Open("")
	if err != nil {
		t.Fatalf("failed to open GODB: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Verify table was not created
	var count int
	err = db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='blogs'").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query sqlite_master: %v", err)
	}
	if count != 0 {
		t.Error("expected empty dataset table 'blogs' to be skipped (not created)")
	}
}
