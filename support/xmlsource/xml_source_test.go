package xmlsource

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"testing/fstest"
	"time"
)

// writeTempXML writes content to a temporary file and returns its path.
func writeTempXML(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp XML: %v", err)
	}
	return path
}

// --- NewXmlSource (string) tests ---

func TestNewXmlSource_String_Basic(t *testing.T) {
	xmlString := `<users>
		<user id="1">
			<name>Alice</name>
			<email>alice@example.com</email>
			<active>true</active>
		</user>
		<user id="2">
			<name>Bob</name>
			<email>bob@example.com</email>
			<active>false</active>
		</user>
		<user id="3">
			<name>Charlie</name>
			<email>charlie@example.com</email>
			<active>true</active>
		</user>
	</users>`

	model := NewXmlSource(xmlString, "users")

	if model.TableName() != "users" {
		t.Errorf("expected table name 'users', got '%s'", model.TableName())
	}

	rows, err := model.Rows()
	if err != nil {
		t.Fatalf("Rows() error: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}

	// Attribute → column
	if rows[0]["id"] != int64(1) {
		t.Errorf("expected id=1 (int64 from attribute), got %v (%T)", rows[0]["id"], rows[0]["id"])
	}
	// Leaf element → column
	if rows[0]["name"] != "Alice" {
		t.Errorf("expected name='Alice', got %v", rows[0]["name"])
	}
	if rows[0]["email"] != "alice@example.com" {
		t.Errorf("expected email='alice@example.com', got %v", rows[0]["email"])
	}
	// bool string → int64
	if rows[0]["active"] != int64(1) {
		t.Errorf("expected active=1 (bool→int64), got %v (%T)", rows[0]["active"], rows[0]["active"])
	}
	if rows[1]["active"] != int64(0) {
		t.Errorf("expected active=0 (bool→int64), got %v (%T)", rows[1]["active"], rows[1]["active"])
	}
}

func TestNewXmlSource_String_TypeInference(t *testing.T) {
	xmlString := `<products>
		<product id="1">
			<price>19.99</price>
			<name>Widget</name>
			<created>2024-01-15T10:30:00Z</created>
		</product>
	</products>`

	model := NewXmlSource(xmlString, "products")

	rows, _ := model.Rows()
	row := rows[0]

	// id attribute → int64
	if _, ok := row["id"].(int64); !ok {
		t.Errorf("expected id to be int64, got %T (%v)", row["id"], row["id"])
	}
	// price → float64
	if f, ok := row["price"].(float64); !ok || f != 19.99 {
		t.Errorf("expected price=19.99 (float64), got %T (%v)", row["price"], row["price"])
	}
	// name → string
	if s, ok := row["name"].(string); !ok || s != "Widget" {
		t.Errorf("expected name='Widget' (string), got %T (%v)", row["name"], row["name"])
	}
	// created → time.Time
	if _, ok := row["created"].(time.Time); !ok {
		t.Errorf("expected created to be time.Time, got %T (%v)", row["created"], row["created"])
	}
}

func TestNewXmlSource_String_NestedElementAsJSON(t *testing.T) {
	xmlString := `<users>
		<user id="1">
			<name>Alice</name>
			<address>
				<city>NYC</city>
				<zip>10001</zip>
			</address>
		</user>
	</users>`

	model := NewXmlSource(xmlString, "users")

	rows, _ := model.Rows()
	row := rows[0]

	// Nested element should be stored as JSON string
	addrStr, ok := row["address"].(string)
	if !ok {
		t.Fatalf("expected address to be string (JSON), got %T (%v)", row["address"], row["address"])
	}

	var addr map[string]any
	if err := json.Unmarshal([]byte(addrStr), &addr); err != nil {
		t.Fatalf("expected address to be valid JSON, got parse error: %v", err)
	}
	if addr["city"] != "NYC" {
		t.Errorf("expected city='NYC', got %v", addr["city"])
	}
}

func TestNewXmlSource_String_AttributesAndElements(t *testing.T) {
	// Both attributes and leaf elements should become columns
	xmlString := `<items>
		<item id="1" category="tools">
			<name>Hammer</name>
			<price>29.99</price>
		</item>
	</items>`

	model := NewXmlSource(xmlString, "items")

	rows, _ := model.Rows()
	row := rows[0]

	if row["id"] != int64(1) {
		t.Errorf("expected id=1, got %v", row["id"])
	}
	if row["category"] != "tools" {
		t.Errorf("expected category='tools', got %v", row["category"])
	}
	if row["name"] != "Hammer" {
		t.Errorf("expected name='Hammer', got %v", row["name"])
	}
	if row["price"] != 29.99 {
		t.Errorf("expected price=29.99, got %v", row["price"])
	}
}

func TestNewXmlSource_String_DateFormats(t *testing.T) {
	xmlString := `<dates>
		<entry>
			<date_only>2024-01-15</date_only>
			<datetime>2024-01-15 10:30:00</datetime>
			<us_date>01/15/2024</us_date>
		</entry>
	</dates>`

	model := NewXmlSource(xmlString, "dates")

	rows, _ := model.Rows()
	row := rows[0]

	if _, ok := row["date_only"].(time.Time); !ok {
		t.Errorf("expected date_only to be time.Time, got %T (%v)", row["date_only"], row["date_only"])
	}
	if _, ok := row["datetime"].(time.Time); !ok {
		t.Errorf("expected datetime to be time.Time, got %T (%v)", row["datetime"], row["datetime"])
	}
	if _, ok := row["us_date"].(time.Time); !ok {
		t.Errorf("expected us_date to be time.Time, got %T (%v)", row["us_date"], row["us_date"])
	}
}

func TestNewXmlSource_String_PanicsOnInvalidXML(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic on invalid XML")
		}
	}()
	NewXmlSource(`<not valid xml`, "bad")
}

func TestNewXmlSource_String_PanicsOnNoChildElements(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic on no child elements")
		}
	}()
	NewXmlSource(`<empty></empty>`, "empty")
}

func TestNewXmlSource_String_PanicsOnEmptyString(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic on empty string")
		}
	}()
	NewXmlSource("", "empty")
}

// --- NewXmlFileSource (file path) tests ---

func TestNewXmlFileSource_Basic(t *testing.T) {
	content := `<users>
		<user id="1">
			<name>Alice</name>
			<active>true</active>
		</user>
		<user id="2">
			<name>Bob</name>
			<active>false</active>
		</user>
	</users>`

	path := writeTempXML(t, "users.xml", content)
	model := NewXmlFileSource(path)

	if model.TableName() != "users" {
		t.Errorf("expected table name 'users', got '%s'", model.TableName())
	}

	rows, err := model.Rows()
	if err != nil {
		t.Fatalf("Rows() error: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[0]["name"] != "Alice" {
		t.Errorf("expected name='Alice', got %v", rows[0]["name"])
	}
}

func TestNewXmlFileSource_TableNameFromPath(t *testing.T) {
	content := `<items><item id="1"><name>Widget</name></item></items>`
	path := writeTempXML(t, "my_table.xml", content)
	model := NewXmlFileSource(path)

	if model.TableName() != "my_table" {
		t.Errorf("expected table name 'my_table', got '%s'", model.TableName())
	}
}

func TestNewXmlFileSource_PanicsOnNonExistentFile(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic on non-existent file")
		}
	}()
	NewXmlFileSource("/nonexistent/path/to/file.xml")
}

func TestNewXmlFileSource_PanicsOnEmptyFile(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic on empty file")
		}
	}()
	path := writeTempXML(t, "empty.xml", "")
	NewXmlFileSource(path)
}

// --- NewXmlFSSource (fstest.MapFS) tests ---

func TestNewXmlFSSource_Basic(t *testing.T) {
	sys := fstest.MapFS{
		"data/users.xml": &fstest.MapFile{
			Data: []byte(`<users><user id="1"><name>Alice</name></user><user id="2"><name>Bob</name></user></users>`),
		},
	}

	model := NewXmlFSSource(sys, "data/users.xml")
	if model.TableName() != "users" {
		t.Errorf("expected table name 'users', got '%s'", model.TableName())
	}

	rows, err := model.Rows()
	if err != nil {
		t.Fatalf("Rows() error: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[0]["name"] != "Alice" {
		t.Errorf("expected name='Alice', got %v", rows[0]["name"])
	}
}

// --- Internal function tests ---

func TestInferAndConvert(t *testing.T) {
	// int
	if v := InferAndConvert("42"); v != int64(42) {
		t.Errorf("InferAndConvert(\"42\") = %v, want int64(42)", v)
	}
	// negative int
	if v := InferAndConvert("-456"); v != int64(-456) {
		t.Errorf("InferAndConvert(\"-456\") = %v, want int64(-456)", v)
	}
	// float
	if v := InferAndConvert("19.99"); v != 19.99 {
		t.Errorf("InferAndConvert(\"19.99\") = %v, want 19.99", v)
	}
	// bool true → int64(1)
	if v := InferAndConvert("true"); v != int64(1) {
		t.Errorf("InferAndConvert(\"true\") = %v, want int64(1)", v)
	}
	// bool false → int64(0)
	if v := InferAndConvert("false"); v != int64(0) {
		t.Errorf("InferAndConvert(\"false\") = %v, want int64(0)", v)
	}
	// RFC3339 → time.Time
	if v, ok := InferAndConvert("2024-01-15T10:30:00Z").(time.Time); !ok {
		t.Errorf("InferAndConvert(\"2024-01-15T10:30:00Z\") expected time.Time, got %T", v)
	}
	// date only → time.Time
	if v, ok := InferAndConvert("2024-01-15").(time.Time); !ok {
		t.Errorf("InferAndConvert(\"2024-01-15\") expected time.Time, got %T", v)
	}
	// string
	if v := InferAndConvert("hello"); v != "hello" {
		t.Errorf("InferAndConvert(\"hello\") = %v, want 'hello'", v)
	}
	// empty → nil
	if v := InferAndConvert(""); v != nil {
		t.Errorf("InferAndConvert(\"\") = %v, want nil", v)
	}
}

func TestNewXmlSource_String_RepeatedLeafElementsAsArray(t *testing.T) {
	// Repeated leaf child elements with the same name should be collected
	// into a JSON array within the container's JSON object, not silently
	// collapsed (last-wins).
	// <tags><tag>red</tag><tag>blue</tag></tags>
	// → "tags" column holds {"tag":["red","blue"]} as a JSON string.
	xmlString := `<items>
		<item id="1">
			<tags>
				<tag>red</tag>
				<tag>blue</tag>
			</tags>
		</item>
	</items>`

	model := NewXmlSource(xmlString, "items")
	rows, _ := model.Rows()
	row := rows[0]

	tagsStr, ok := row["tags"].(string)
	if !ok {
		t.Fatalf("expected tags to be string (JSON), got %T (%v)", row["tags"], row["tags"])
	}

	var tagsObj map[string]any
	if err := json.Unmarshal([]byte(tagsStr), &tagsObj); err != nil {
		t.Fatalf("expected tags to be valid JSON object, got parse error: %v (value: %s)", err, tagsStr)
	}
	tags, ok := tagsObj["tag"].([]any)
	if !ok {
		t.Fatalf("expected tags.tag to be a JSON array, got %T (%v)", tagsObj["tag"], tagsObj["tag"])
	}
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %d: %v", len(tags), tags)
	}
	if tags[0] != "red" || tags[1] != "blue" {
		t.Errorf("expected [red, blue], got %v", tags)
	}
}

func TestNewXmlSource_String_RepeatedLeafElementsSingleElement(t *testing.T) {
	// A single child element inside a container should NOT be wrapped in an
	// array — it should remain a scalar value (backward compatibility).
	xmlString := `<items>
		<item id="1">
			<tags>
				<tag>red</tag>
			</tags>
		</item>
	</items>`

	model := NewXmlSource(xmlString, "items")
	rows, _ := model.Rows()
	row := rows[0]

	tagsStr, ok := row["tags"].(string)
	if !ok {
		t.Fatalf("expected tags to be string (JSON), got %T (%v)", row["tags"], row["tags"])
	}

	// Single element should be a JSON object {"tag":"red"}, not an array
	var tags map[string]any
	if err := json.Unmarshal([]byte(tagsStr), &tags); err != nil {
		t.Fatalf("expected tags to be valid JSON object, got parse error: %v (value: %s)", err, tagsStr)
	}
	if tags["tag"] != "red" {
		t.Errorf("expected tag='red', got %v", tags["tag"])
	}
}

func TestNewXmlSource_String_RepeatedLeafElementsAtRowLevel(t *testing.T) {
	// Repeated leaf elements directly under the row element should also be
	// collected into a JSON array string.
	xmlString := `<users>
		<user id="1">
			<role>admin</role>
			<role>editor</role>
			<name>Alice</name>
		</user>
	</users>`

	model := NewXmlSource(xmlString, "users")
	rows, _ := model.Rows()
	row := rows[0]

	rolesStr, ok := row["role"].(string)
	if !ok {
		t.Fatalf("expected role to be string (JSON array), got %T (%v)", row["role"], row["role"])
	}

	var roles []any
	if err := json.Unmarshal([]byte(rolesStr), &roles); err != nil {
		t.Fatalf("expected role to be valid JSON array, got parse error: %v (value: %s)", err, rolesStr)
	}
	if len(roles) != 2 {
		t.Fatalf("expected 2 roles, got %d: %v", len(roles), roles)
	}
	if roles[0] != "admin" || roles[1] != "editor" {
		t.Errorf("expected [admin, editor], got %v", roles)
	}

	// Non-repeated element should remain scalar
	if row["name"] != "Alice" {
		t.Errorf("expected name='Alice', got %v", row["name"])
	}
}

func TestNewXmlSource_String_RepeatedNestedElementsAsArray(t *testing.T) {
	// Repeated nested (non-leaf) elements with the same name should also be
	// collected into a JSON array within the container's JSON object.
	// <addresses><address>...</address><address>...</address></addresses>
	// → "addresses" column holds {"address":[{...},{...}]} as a JSON string.
	xmlString := `<data>
		<record id="1">
			<addresses>
				<address>
					<city>NYC</city>
					<zip>10001</zip>
				</address>
				<address>
					<city>SF</city>
					<zip>94102</zip>
				</address>
			</addresses>
		</record>
	</data>`

	model := NewXmlSource(xmlString, "data")
	rows, _ := model.Rows()
	row := rows[0]

	addrsStr, ok := row["addresses"].(string)
	if !ok {
		t.Fatalf("expected addresses to be string (JSON), got %T (%v)", row["addresses"], row["addresses"])
	}

	var addrsObj map[string]any
	if err := json.Unmarshal([]byte(addrsStr), &addrsObj); err != nil {
		t.Fatalf("expected addresses to be valid JSON object, got parse error: %v (value: %s)", err, addrsStr)
	}
	addrs, ok := addrsObj["address"].([]any)
	if !ok {
		t.Fatalf("expected addresses.address to be a JSON array, got %T (%v)", addrsObj["address"], addrsObj["address"])
	}
	if len(addrs) != 2 {
		t.Fatalf("expected 2 addresses, got %d: %v", len(addrs), addrs)
	}
	// Verify elements are JSON objects (not double-encoded strings)
	firstAddr, ok := addrs[0].(map[string]any)
	if !ok {
		t.Fatalf("expected first address to be a JSON object, got %T (%v)", addrs[0], addrs[0])
	}
	if firstAddr["city"] != "NYC" {
		t.Errorf("expected first address city='NYC', got %v", firstAddr["city"])
	}
	secondAddr, ok := addrs[1].(map[string]any)
	if !ok {
		t.Fatalf("expected second address to be a JSON object, got %T (%v)", addrs[1], addrs[1])
	}
	if secondAddr["city"] != "SF" {
		t.Errorf("expected second address city='SF', got %v", secondAddr["city"])
	}
}

func TestDeriveTableName_XML(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"data/users.xml", "users"},
		{"users.xml", "users"},
		{"/home/user/data/products.xml", "products"},
	}
	for _, tt := range tests {
		got := deriveTableName(tt.path)
		if got != tt.want {
			t.Errorf("deriveTableName(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}

	if runtime.GOOS == "windows" {
		got := deriveTableName("C:\\data\\orders.xml")
		if got != "orders" {
			t.Errorf("deriveTableName(\"C:\\\\data\\\\orders.xml\") = %q, want \"orders\"", got)
		}
	}
}
