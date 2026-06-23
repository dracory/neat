package orm

// ArraySource is implemented by any "model" that wants array-backed storage.
type ArraySource interface {
	TableName() string
	Rows() ([]map[string]any, error)
}

// ArraySchema is an optional interface for empty-dataset or type-ambiguous cases.
type ArraySchema interface {
	Schema() map[string]string // column -> type ("string", "int", "float", "bool", "time")
}
