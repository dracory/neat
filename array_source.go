package neat

import (
	"github.com/dracory/neat/support/arraysource"
)

type ArraySourceModel = arraysource.Model

// NewArraySourceFrom creates an array-backed data source from a slice of
// structs or map[string]any. This is the primary entry point for array-backed
// queries — the third constructor in the NewArraySource family.
func NewArraySourceFrom[T any](items []T) *arraysource.Model {
	return arraysource.NewArraySourceFrom(items)
}

func NewArraySource(rows []map[string]any) *arraysource.Model {
	return arraysource.New(rows)
}

func NewArraySourceWithSchema(rows []map[string]any, schema map[string]string) *arraysource.Model {
	return arraysource.NewWithSchema(rows, schema)
}
