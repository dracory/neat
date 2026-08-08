package arraysource

import (
	"strconv"
	"sync/atomic"
)

type Model struct {
	table  string
	data   []map[string]any
	schema map[string]string
}

func (m *Model) TableName() string              { return m.table }
func (m *Model) Rows() ([]map[string]any, error) { return m.data, nil }
func (m *Model) Schema() map[string]string       { return m.schema }

// Table sets a custom table name. Must be called before passing to Model().
// Not safe to call concurrently with a query using this source.
func (m *Model) Table(name string) *Model {
	m.table = name
	return m
}

// New creates an ArraySource from []map[string]any rows. Table name is
// auto-generated. Rows are shallow-copied (snapshot at call time).
// Panics if rows is nil or empty — use NewWithSchema for empty datasets.
func New(rows []map[string]any) *Model {
	if len(rows) == 0 {
		panic("arraysource: New() requires non-empty rows; use NewWithSchema() for empty datasets")
	}
	return &Model{
		table:  "array_map_" + strconv.FormatUint(atomic.AddUint64(&arrayCounter, 1), 10),
		data:   copyRows(rows),
	}
}

// NewWithSchema creates an ArraySource with an explicit column schema.
// Rows are shallow-copied (snapshot at call time).
func NewWithSchema(rows []map[string]any, schema map[string]string) *Model {
	return &Model{
		table:  "array_map_" + strconv.FormatUint(atomic.AddUint64(&arrayCounter, 1), 10),
		data:   copyRows(rows),
		schema: schema,
	}
}
