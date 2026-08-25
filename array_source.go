package neat

import (
	"io/fs"

	"github.com/dracory/neat/support/arraysource"
	"github.com/dracory/neat/support/csvsource"
	"github.com/dracory/neat/support/jsonsource"
	"github.com/dracory/neat/support/xmlsource"
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

// NewCsvSource parses a CSV string and returns an array-backed data source.
// The first line must be a header defining column names. Column types are
// inferred from the data (int, float, bool, time, string). The table name
// must be provided explicitly since there is no filename to derive it from.
//
//	database.Query().
//	    Model(neat.NewCsvSource(csvString, "users")).
//	    Where("active = ?", true).
//	    Get(&users)
//
// Panics if the CSV string is empty or has no header row.
func NewCsvSource(csvString string, tableName string) *arraysource.Model {
	return csvsource.NewCsvSource(csvString, tableName)
}

// NewCsvFileSource reads a CSV file and returns an array-backed data source.
// The first row must be a header defining column names. Column types are
// inferred from the data (int, float, bool, time, string). The table name
// is derived from the filename (e.g., "data/users.csv" → "users").
//
//	database.Query().
//	    Model(neat.NewCsvFileSource("data/users.csv")).
//	    Where("active = ?", true).
//	    Get(&users)
//
// Panics if the file cannot be opened or is empty.
func NewCsvFileSource(filePath string) *arraysource.Model {
	return csvsource.NewCsvFileSource(filePath)
}

// NewCsvFileSourceWithDelimiter is like NewCsvFileSource but allows specifying
// a custom field delimiter (e.g., '\t' for TSV files).
func NewCsvFileSourceWithDelimiter(filePath string, delimiter rune) *arraysource.Model {
	return csvsource.NewCsvFileSourceWithDelimiter(filePath, delimiter)
}

// NewCsvFSSource reads a CSV file from an embedded filesystem (embed.FS / fs.FS)
// and returns an array-backed data source ready for querying.
func NewCsvFSSource(sys fs.FS, filePath string) *arraysource.Model {
	return csvsource.NewCsvFSSource(sys, filePath)
}

// NewCsvFSSourceWithDelimiter reads a CSV file from an embedded filesystem (embed.FS / fs.FS)
// with a custom field delimiter and returns an array-backed data source.
func NewCsvFSSourceWithDelimiter(sys fs.FS, filePath string, delimiter rune) *arraysource.Model {
	return csvsource.NewCsvFSSourceWithDelimiter(sys, filePath, delimiter)
}

// NewJsonSource parses a JSON or JSONL string and returns an array-backed
// data source. Pass isJSONL=true for JSONL content (one object per line),
// false for a JSON array. JSON native types are preserved. RFC3339 strings
// are converted to time.Time. Nested objects/arrays are stored as JSON
// strings. The table name must be provided explicitly.
//
//	database.Query().
//	    Model(neat.NewJsonSource(jsonString, "users", false)).
//	    Where("active = ?", true).
//	    Get(&users)
//
// Panics if the content cannot be parsed.
func NewJsonSource(jsonString string, tableName string, isJSONL bool) *arraysource.Model {
	return jsonsource.NewJsonSource(jsonString, tableName, isJSONL)
}

// NewJsonFileSource reads a JSON or JSONL file and returns an array-backed
// data source. The file must contain a JSON array of objects (for .json) or
// one JSON object per line (for .jsonl/.ndjson). JSON native types are
// preserved. RFC3339 strings are converted to time.Time. Nested
// objects/arrays are stored as JSON strings. The table name is derived from
// the filename (e.g., "data/users.json" → "users").
//
//	database.Query().
//	    Model(neat.NewJsonFileSource("data/users.json")).
//	    Where("active = ?", true).
//	    Get(&users)
//
// Panics if the file cannot be opened or parsed.
func NewJsonFileSource(filePath string) *arraysource.Model {
	return jsonsource.NewJsonFileSource(filePath)
}

// NewJsonFSSource reads a JSON or JSONL file from an embedded filesystem
// (embed.FS / fs.FS) and returns an array-backed data source.
func NewJsonFSSource(sys fs.FS, filePath string) *arraysource.Model {
	return jsonsource.NewJsonFSSource(sys, filePath)
}

// NewXmlSource parses an XML string and returns an array-backed data source.
// The XML must have a root element containing repeated child elements. Each
// child becomes a row. Attributes and leaf sub-elements become columns.
// Nested sub-elements are stored as JSON strings. Column types are inferred
// (int, float, bool, time, string). The table name must be provided explicitly.
//
//	database.Query().
//	    Model(neat.NewXmlSource(xmlString, "users")).
//	    Where("active = ?", true).
//	    Get(&users)
//
// Panics if the XML cannot be parsed or has no child elements.
func NewXmlSource(xmlString string, tableName string) *arraysource.Model {
	return xmlsource.NewXmlSource(xmlString, tableName)
}

// NewXmlFileSource reads an XML file and returns an array-backed data source.
// The XML must have a root element containing repeated child elements. Each
// child becomes a row. Attributes and leaf sub-elements become columns.
// The table name is derived from the filename (e.g., "data/users.xml" → "users").
//
//	database.Query().
//	    Model(neat.NewXmlFileSource("data/users.xml")).
//	    Where("active = ?", true).
//	    Get(&users)
//
// Panics if the file cannot be opened, parsed, or has no child elements.
func NewXmlFileSource(filePath string) *arraysource.Model {
	return xmlsource.NewXmlFileSource(filePath)
}

// NewXmlFSSource reads an XML file from an embedded filesystem (embed.FS / fs.FS)
// and returns an array-backed data source.
func NewXmlFSSource(sys fs.FS, filePath string) *arraysource.Model {
	return xmlsource.NewXmlFSSource(sys, filePath)
}
