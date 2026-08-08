package arraysource

import (
	"strings"
	"testing"
	"time"
)

type Status struct {
	ID    int    `db:"id"`
	Name  string `db:"name"`
	Color string `db:"color"`
}

type EmbeddedInfo struct {
	Age int `db:"age"`
}

type UserWithEmbedded struct {
	ID   int `db:"id"`
	Name string
	EmbeddedInfo
	Address *string   `db:"address"`
	Created time.Time `db:"created"`
}

type UserWithAssociation struct {
	ID       int `db:"id"`
	Statuses []Status
	MainType *Status
	Ignored  *int `db:"-"`
}

func TestNewArraySourceFrom_Structs(t *testing.T) {
	items := []Status{
		{ID: 1, Name: "Pending", Color: "yellow"},
		{ID: 2, Name: "Active", Color: "green"},
	}

	model := NewArraySourceFrom(items)
	if model == nil {
		t.Fatal("expected model to not be nil")
	}

	if !strings.HasPrefix(model.TableName(), "array_status_") {
		t.Errorf("expected table name starting with array_status_, got %s", model.TableName())
	}

	rows, err := model.Rows()
	if err != nil {
		t.Fatalf("Rows() err: %v", err)
	}

	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	if rows[0]["id"] != 1 || rows[0]["name"] != "Pending" || rows[0]["color"] != "yellow" {
		t.Errorf("unexpected first row: %v", rows[0])
	}
}

func TestNewArraySourceFrom_EmbeddedStructs(t *testing.T) {
	address := "123 Main St"
	now := time.Now()
	items := []UserWithEmbedded{
		{
			ID:           1,
			Name:         "Alice",
			EmbeddedInfo: EmbeddedInfo{Age: 30},
			Address:      &address,
			Created:      now,
		},
	}

	model := NewArraySourceFrom(items)
	rows, _ := model.Rows()

	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	row := rows[0]
	if row["id"] != 1 || row["name"] != "Alice" || row["age"] != 30 {
		t.Errorf("unexpected row content: %v", row)
	}
	if row["address"].(string) != "123 Main St" {
		t.Errorf("expected address to be dereferenced, got %v", row["address"])
	}
	if !row["created"].(time.Time).Equal(now) {
		t.Errorf("expected created to be now, got %v", row["created"])
	}
}

func TestNewArraySourceFrom_SkipsAssociations(t *testing.T) {
	items := []UserWithAssociation{
		{
			ID:       1,
			Statuses: []Status{{ID: 1}},
			MainType: &Status{ID: 2},
		},
	}

	model := NewArraySourceFrom(items)
	rows, _ := model.Rows()

	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	row := rows[0]
	if len(row) != 1 || row["id"] != 1 {
		t.Errorf("expected only 'id' to be present, got %v", row)
	}
}

func TestNewArraySourceFrom_NullablePointer(t *testing.T) {
	type Sample struct {
		ID   int
		Val  *string
		Time *time.Time
	}

	strVal := "hello"
	now := time.Now()
	items := []Sample{
		{ID: 1, Val: &strVal, Time: &now},
		{ID: 2, Val: nil, Time: nil},
	}

	model := NewArraySourceFrom(items)
	rows, _ := model.Rows()

	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	if rows[0]["val"] != "hello" {
		t.Errorf("expected 'hello', got %v", rows[0]["val"])
	}
	if !rows[0]["time"].(time.Time).Equal(now) {
		t.Errorf("expected time, got %v", rows[0]["time"])
	}
	if rows[1]["val"] != nil {
		t.Errorf("expected nil for Val, got %v", rows[1]["val"])
	}
	if rows[1]["time"] != nil {
		t.Errorf("expected nil for Time, got %v", rows[1]["time"])
	}
}

func TestNewArraySourceFrom_MapSlice(t *testing.T) {
	items := []map[string]any{
		{"id": 1, "name": "Alice"},
	}

	model := NewArraySourceFrom(items)
	rows, _ := model.Rows()

	if len(rows) != 1 || rows[0]["name"] != "Alice" {
		t.Errorf("expected Alice, got %v", rows)
	}
}

func TestNewArraySourceFrom_MapSlice_SnapshotSemantics(t *testing.T) {
	items := []map[string]any{
		{"id": 1, "name": "Alice"},
	}

	model := NewArraySourceFrom(items)
	items[0]["name"] = "Bob" // mutate original map

	rows, _ := model.Rows()
	if rows[0]["name"] != "Alice" {
		t.Errorf("expected snapshot semantics, got mutated value: %v", rows[0]["name"])
	}
}

func TestNewArraySourceFrom_StructsWithTags(t *testing.T) {
	type TaggedStruct struct {
		ID         int    `db:"custom_id"`
		NeatVal    string `neat:"custom_neat_val"`
		GormVal    string `gorm:"column:custom_gorm_val"`
		DefaultVal string
	}

	items := []TaggedStruct{
		{ID: 123, NeatVal: "neat", GormVal: "gorm", DefaultVal: "default"},
	}

	model := NewArraySourceFrom(items)
	rows, _ := model.Rows()

	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	row := rows[0]
	if row["custom_id"] != 123 {
		t.Errorf("expected 'custom_id' field, got: %v", row)
	}
	if row["custom_neat_val"] != "neat" {
		t.Errorf("expected 'custom_neat_val' field, got: %v", row)
	}
	if row["custom_gorm_val"] != "gorm" {
		t.Errorf("expected 'custom_gorm_val' field, got: %v", row)
	}
	if row["default_val"] != "default" {
		t.Errorf("expected 'default_val' field, got: %v", row)
	}
}

func TestNewArraySourceFrom_TimeTime(t *testing.T) {
	type TimeStruct struct {
		ID       int
		ValTime  time.Time
		PtrTime  *time.Time
		NilTime  *time.Time
	}

	now := time.Now()
	items := []TimeStruct{
		{ID: 1, ValTime: now, PtrTime: &now, NilTime: nil},
	}

	model := NewArraySourceFrom(items)
	rows, _ := model.Rows()

	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	row := rows[0]
	if !row["val_time"].(time.Time).Equal(now) {
		t.Errorf("expected 'val_time' to be %v, got %v", now, row["val_time"])
	}
	if !row["ptr_time"].(time.Time).Equal(now) {
		t.Errorf("expected 'ptr_time' to be %v, got %v", now, row["ptr_time"])
	}
	if row["nil_time"] != nil {
		t.Errorf("expected 'nil_time' to be nil, got %v", row["nil_time"])
	}
}

func TestNewArraySourceFrom_PointerSlice(t *testing.T) {
	items := []*Status{
		{ID: 1, Name: "Pending"},
	}

	model := NewArraySourceFrom(items)
	if !strings.HasPrefix(model.TableName(), "array_status_") {
		t.Errorf("expected array_status_ table name, got %s", model.TableName())
	}
}

func TestNewArraySourceFrom_PanicsOnEmpty(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic")
		}
	}()
	var empty []Status
	NewArraySourceFrom(empty)
}

func TestNewArraySourceFrom_PanicsOnUnsupportedType(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic")
		}
	}()
	NewArraySourceFrom([]int{1, 2, 3})
}
