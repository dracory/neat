package xmlsource

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dracory/neat/support/arraysource"
)

// NewXmlSource parses an XML string and returns an array-backed data source
// ready for querying with the array driver.
//
// The XML must have a root element containing repeated child elements. Each
// child element becomes a row. Column values come from:
//   - Attributes on the child element: <user id="1"> → column "id"
//   - Leaf sub-elements (text-only): <name>Alice</name> → column "name"
//   - Nested sub-elements: stored as JSON strings (queryable via SQLite JSON)
//
// For example:
//
//	<users>
//	  <user id="1">
//	    <name>Alice</name>
//	    <active>true</active>
//	  </user>
//	  <user id="2">
//	    <name>Bob</name>
//	    <active>false</active>
//	  </user>
//	</users>
//
// Column types are inferred from string values (int, float, bool, time,
// string), same as the CSV source.
//
// The table name must be provided explicitly since there is no filename
// to derive it from. Override with the .Table() method if needed.
//
//	database.Query().
//	    Model(neat.NewXmlSource(xmlString, "users")).
//	    Where("active = ?", true).
//	    Get(&users)
//
// Panics if the XML cannot be parsed or has no child elements.
func NewXmlSource(xmlString string, tableName string) *arraysource.Model {
	rows, err := parseXMLString(xmlString)
	if err != nil {
		panic(fmt.Sprintf("xmlsource: failed to parse XML string: %v", err))
	}

	if len(rows) == 0 {
		panic("xmlsource: no child elements found — cannot infer schema without data")
	}

	return arraysource.New(rows).Table(tableName)
}

// NewXmlFileSource reads an XML file and returns an array-backed data source
// ready for querying with the array driver.
//
// The XML must have a root element containing repeated child elements. Each
// child element becomes a row. See NewXmlSource for details on column
// extraction and type inference.
//
// The table name is derived from the filename (without the extension).
// For example, "data/users.xml" → table name "users".
//
//	database.Query().
//	    Model(neat.NewXmlFileSource("data/users.xml")).
//	    Where("active = ?", true).
//	    Get(&users)
//
// Panics if the file cannot be opened, parsed, or has no child elements.
func NewXmlFileSource(filePath string) *arraysource.Model {
	rows, err := parseXMLFile(filePath)
	if err != nil {
		panic(fmt.Sprintf("xmlsource: failed to parse %s: %v", filePath, err))
	}

	if len(rows) == 0 {
		panic(fmt.Sprintf("xmlsource: %s contains no child elements — cannot infer schema without data", filePath))
	}

	tableName := deriveTableName(filePath)
	return arraysource.New(rows).Table(tableName)
}

// deriveTableName extracts the table name from the file path by taking
// the base filename and removing the extension.
// "data/users.xml" → "users"
func deriveTableName(filePath string) string {
	base := filepath.Base(filePath)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}

// parseXMLString parses an XML string and returns rows.
func parseXMLString(content string) ([]map[string]any, error) {
	return parseXMLReader(strings.NewReader(content))
}

// parseXMLFile reads an XML file and returns rows.
func parseXMLFile(filePath string) ([]map[string]any, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}
	defer func() { _ = f.Close() }()
	return parseXMLReader(f)
}

// parseXMLReader parses XML from any io.Reader. It finds the root element,
// then treats each direct child element as a row. Attributes and leaf
// sub-elements become columns.
func parseXMLReader(r io.Reader) ([]map[string]any, error) {
	dec := xml.NewDecoder(r)

	// Find and skip the root element (advances decoder past root's opening tag)
	_, err := findRootElement(dec)
	if err != nil {
		return nil, err
	}

	// Parse each direct child of the root as a row.
	// parseElement fully consumes each child (including its EndElement),
	// so the next token is either the next child's StartElement or the
	// root's EndElement.
	var rows []map[string]any

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("XML parse error: %w", err)
		}

		switch t := tok.(type) {
		case xml.StartElement:
			// Direct child of root — parseElement consumes the entire subtree
			row, err := parseElement(dec, t)
			if err != nil {
				return nil, err
			}
			rows = append(rows, row)
		case xml.EndElement:
			// End of root element — done
			return rows, nil
		}
	}

	return rows, nil
}

// findRootElement skips leading tokens and returns the first StartElement.
func findRootElement(dec *xml.Decoder) (xml.StartElement, error) {
	for {
		tok, err := dec.Token()
		if err != nil {
			return xml.StartElement{}, fmt.Errorf("XML parse error: %w", err)
		}
		if start, ok := tok.(xml.StartElement); ok {
			return start, nil
		}
	}
}

// parseElement parses a single XML element into a map[string]any row.
// Attributes become columns. Sub-elements that are leaf nodes (text-only)
// become columns. Sub-elements with nested children are stored as JSON strings.
func parseElement(dec *xml.Decoder, start xml.StartElement) (map[string]any, error) {
	row := make(map[string]any)

	// Add attributes as columns
	for _, attr := range start.Attr {
		row[attr.Name.Local] = inferAndConvert(attr.Value)
	}

	// Parse sub-elements and text content
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, fmt.Errorf("XML parse error in element %s: %w", start.Name.Local, err)
		}

		switch t := tok.(type) {
		case xml.StartElement:
			// Sub-element: parse it recursively
			childRow, isLeaf, text, err := parseChildElement(dec, t)
			if err != nil {
				return nil, err
			}
			if isLeaf {
				// Leaf element: text content becomes the column value
				row[t.Name.Local] = inferAndConvert(text)
			} else {
				// Nested element: store as JSON string
				b, err := json.Marshal(childRow)
				if err != nil {
					row[t.Name.Local] = fmt.Sprintf("%v", childRow)
				} else {
					row[t.Name.Local] = string(b)
				}
			}
		case xml.EndElement:
			// End of this element
			return row, nil
		case xml.CharData:
			// Text directly inside this element (not in a sub-element).
			// For row elements, this is usually whitespace between child
			// elements — ignore it.
			continue
		}
	}
}

// parseChildElement parses a sub-element. Returns:
//   - childRow: the parsed content (for nested elements)
//   - isLeaf: true if the element contains only text (no child elements)
//   - text: the text content (valid when isLeaf is true)
func parseChildElement(dec *xml.Decoder, start xml.StartElement) (map[string]any, bool, string, error) {
	row := make(map[string]any)
	var textContent strings.Builder
	hasChildElements := false

	// Add attributes
	for _, attr := range start.Attr {
		row[attr.Name.Local] = inferAndConvert(attr.Value)
	}

	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, false, "", fmt.Errorf("XML parse error in element %s: %w", start.Name.Local, err)
		}

		switch t := tok.(type) {
		case xml.StartElement:
			hasChildElements = true
			childRow, isLeaf, text, err := parseChildElement(dec, t)
			if err != nil {
				return nil, false, "", err
			}
			if isLeaf {
				row[t.Name.Local] = inferAndConvert(text)
			} else {
				b, err := json.Marshal(childRow)
				if err != nil {
					row[t.Name.Local] = fmt.Sprintf("%v", childRow)
				} else {
					row[t.Name.Local] = string(b)
				}
			}
		case xml.EndElement:
			if !hasChildElements {
				return nil, true, strings.TrimSpace(textContent.String()), nil
			}
			return row, false, "", nil
		case xml.CharData:
			textContent.Write(t)
		}
	}
}

// inferAndConvert tries to convert a string value to its native Go type
// (int64, float64, bool, time.Time), falling back to string.
// This mirrors the CSV source's type inference logic.
func inferAndConvert(val string) any {
	if val == "" {
		return nil
	}

	// Try int
	if n, err := strconv.ParseInt(val, 10, 64); err == nil {
		return n
	}
	// Try float
	if f, err := strconv.ParseFloat(val, 64); err == nil {
		return f
	}
	// Try bool (stored as int64 to match CSV/JSON sources)
	switch val {
	case "true", "True":
		return int64(1)
	case "false", "False":
		return int64(0)
	}
	// Try RFC3339 timestamp
	if t, err := time.Parse(time.RFC3339, val); err == nil {
		return t
	}
	// Try common date formats
	for _, layout := range []string{"2006-01-02", "2006-01-02 15:04:05", "01/02/2006"} {
		if t, err := time.Parse(layout, val); err == nil {
			return t
		}
	}
	return val
}
