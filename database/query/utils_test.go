package query

import (
	"context"
	"database/sql"
	"reflect"
	"testing"
)

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"snake_case", "SnakeCase"},
		{"user_name", "UserName"},
		{"id", "Id"},
		{"created_at", "CreatedAt"},
		{"multiple_words_here", "MultipleWordsHere"},
		{"", ""},
		{"single", "Single"},
		{"alreadyCamel", "Alreadycamel"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toCamelCase(tt.input)
			if result != tt.expected {
				t.Errorf("toCamelCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCamelToSnake(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"CamelCase", "camel_case"},
		{"UserName", "user_name"},
		{"ID", "id"},
		{"CreatedAt", "created_at"},
		{"MultipleWordsHere", "multiple_words_here"},
		{"", ""},
		{"Single", "single"},
		{"already_snake", "already_snake"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := camelToSnake(tt.input)
			if result != tt.expected {
				t.Errorf("camelToSnake(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetPrimaryKeyValue(t *testing.T) {
	type User struct {
		ID uint
	}
	type UserWithIntID struct {
		ID int
	}
	type UserWithLowerId struct {
		Id uint
	}
	type UserNoID struct {
		Name string
	}

	tests := []struct {
		name     string
		value    any
		expected int64
	}{
		{"uint ID", &User{ID: 42}, 42},
		{"int ID", &UserWithIntID{ID: 42}, 42},
		{"lowercase Id", &UserWithLowerId{Id: 42}, 42},
		{"no ID field", &UserNoID{Name: "test"}, 0},
		{"nil pointer", (*User)(nil), 0},
		{"non-struct", "string", 0},
		{"zero ID", &User{ID: 0}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPrimaryKeyValue(tt.value)
			if result != tt.expected {
				t.Errorf("getPrimaryKeyValue(%v) = %d, want %d", tt.value, result, tt.expected)
			}
		})
	}
}

func TestSetModelPrimaryKey(t *testing.T) {
	type User struct {
		ID uint
	}
	type UserWithIntID struct {
		ID int
	}
	type UserWithLowerId struct {
		Id uint
	}
	type UserNoID struct {
		Name string
	}

	tests := []struct {
		name  string
		value any
		id    int64
	}{
		{"uint ID", &User{}, 42},
		{"int ID", &UserWithIntID{}, 42},
		{"lowercase Id", &UserWithLowerId{}, 42},
		{"no ID field", &UserNoID{}, 42},
		{"nil pointer", (*User)(nil), 42},
		{"non-struct", "string", 42},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setModelPrimaryKey(tt.value, tt.id)
			// Verify the ID was set if the struct has an ID field
			if u, ok := tt.value.(*User); ok && u != nil {
				if u.ID != uint(tt.id) {
					t.Errorf("Expected ID to be set to %d, got %d", tt.id, u.ID)
				}
			}
			if u, ok := tt.value.(*UserWithIntID); ok && u != nil {
				if u.ID != int(tt.id) {
					t.Errorf("Expected ID to be set to %d, got %d", tt.id, u.ID)
				}
			}
			if u, ok := tt.value.(*UserWithLowerId); ok && u != nil {
				if u.Id != uint(tt.id) {
					t.Errorf("Expected Id to be set to %d, got %d", tt.id, u.Id)
				}
			}
		})
	}
}

func TestStructFieldColumnName(t *testing.T) {
	type User struct {
		ID       int    `db:"id"`
		Name     string `db:"name"`
		Email    string `neat:"email"`
		Password string `gorm:"column:password"`
		NoTag    string
		Ignore   string `db:"-"`
	}

	tests := []struct {
		name     string
		field    reflect.StructField
		expected string
	}{
		{"db tag", reflect.TypeOf(User{}).Field(0), "id"},
		{"db tag name", reflect.TypeOf(User{}).Field(1), "name"},
		{"neat tag", reflect.TypeOf(User{}).Field(2), "email"},
		{"gorm column tag", reflect.TypeOf(User{}).Field(3), "password"},
		{"no tag", reflect.TypeOf(User{}).Field(4), "no_tag"},
		{"ignore tag", reflect.TypeOf(User{}).Field(5), "ignore"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := structFieldColumnName(tt.field)
			if result != tt.expected {
				t.Errorf("structFieldColumnName(%v) = %q, want %q", tt.field.Name, result, tt.expected)
			}
		})
	}
}

func TestGetColumnToIndexPath(t *testing.T) {
	type User struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	type Embedded struct {
		Age int `db:"age"`
	}

	type UserWithEmbedded struct {
		User
		Embedded
	}

	t.Run("simple struct", func(t *testing.T) {
		result := getColumnToIndexPath(reflect.TypeOf(User{}))
		if len(result) != 2 {
			t.Errorf("Expected 2 columns, got %d", len(result))
		}
		if path, ok := result["id"]; !ok || len(path) != 1 || path[0] != 0 {
			t.Errorf("Expected id path to be [0], got %v", path)
		}
		if path, ok := result["name"]; !ok || len(path) != 1 || path[0] != 1 {
			t.Errorf("Expected name path to be [1], got %v", path)
		}
	})

	t.Run("embedded struct", func(t *testing.T) {
		result := getColumnToIndexPath(reflect.TypeOf(UserWithEmbedded{}))
		if len(result) != 3 {
			t.Errorf("Expected 3 columns, got %d", len(result))
		}
		// Check that embedded fields are included
		if _, ok := result["age"]; !ok {
			t.Error("Expected age field from embedded struct")
		}
	})
}

func TestStructScanDests(t *testing.T) {
	type User struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	t.Run("simple struct", func(t *testing.T) {
		user := User{}
		columns := []string{"id", "name"}
		dests := structScanDests(reflect.ValueOf(&user).Elem(), columns)

		if len(dests) != 2 {
			t.Errorf("Expected 2 destinations, got %d", len(dests))
		}

		// Verify that destinations are nullable wrappers for non-pointer value types
		if dests[0] != nil {
			if _, ok := dests[0].(*sql.NullInt64); !ok {
				t.Errorf("Expected dests[0] to be *sql.NullInt64, got %T", dests[0])
			}
		}
		if dests[1] != nil {
			if _, ok := dests[1].(*sql.NullString); !ok {
				t.Errorf("Expected dests[1] to be *sql.NullString, got %T", dests[1])
			}
		}
	})

	t.Run("unknown column", func(t *testing.T) {
		user := User{}
		columns := []string{"id", "unknown_column"}
		dests := structScanDests(reflect.ValueOf(&user).Elem(), columns)

		if len(dests) != 2 {
			t.Errorf("Expected 2 destinations, got %d", len(dests))
		}

		// Unknown column should have a placeholder
		if dests[1] == nil {
			t.Error("Expected placeholder for unknown column")
		}
	})
}

func TestCopyScanResults(t *testing.T) {
	type User struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	t.Run("copy results", func(t *testing.T) {
		user := User{}
		columns := []string{"id", "name"}
		dests := structScanDests(reflect.ValueOf(&user).Elem(), columns)

		// Simulate scan results using nullable wrappers
		if ni, ok := dests[0].(*sql.NullInt64); ok {
			ni.Int64 = 42
			ni.Valid = true
		}
		if ns, ok := dests[1].(*sql.NullString); ok {
			ns.String = "test"
			ns.Valid = true
		}

		copyScanResults(reflect.ValueOf(&user).Elem(), columns, dests)

		if user.ID != 42 {
			t.Errorf("Expected ID to be 42, got %d", user.ID)
		}
		if user.Name != "test" {
			t.Errorf("Expected Name to be 'test', got %s", user.Name)
		}
	})
}

func TestApplyWhereConditions(t *testing.T) {
	t.Run("map attributes", func(t *testing.T) {
		q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
		attrs := map[string]any{"name": "Alice", "age": 30}

		err := applyWhereConditions(q, attrs)
		if err != nil {
			t.Fatalf("applyWhereConditions failed: %v", err)
		}

		// Verify conditions were added
		if len(q.wheres) != 2 {
			t.Errorf("expected 2 where conditions, got %d", len(q.wheres))
		}
	})

	t.Run("struct attributes", func(t *testing.T) {
		q := NewQuery(context.TODO(), nil, nil, "", nil, nil)

		type User struct {
			Name string
			Age  int
			Zero int
		}

		user := User{Name: "Bob", Age: 25, Zero: 0}
		err := applyWhereConditions(q, user)
		if err != nil {
			t.Fatalf("applyWhereConditions failed: %v", err)
		}

		// Verify conditions were added (zero values should be skipped)
		if len(q.wheres) != 2 {
			t.Errorf("expected 2 where conditions (zero values skipped), got %d", len(q.wheres))
		}
	})

	t.Run("struct pointer", func(t *testing.T) {
		q := NewQuery(context.TODO(), nil, nil, "", nil, nil)

		type User struct {
			Name string
		}

		user := &User{Name: "Charlie"}
		err := applyWhereConditions(q, user)
		if err != nil {
			t.Fatalf("applyWhereConditions failed: %v", err)
		}

		if len(q.wheres) != 1 {
			t.Errorf("expected 1 where condition, got %d", len(q.wheres))
		}
	})
}

func TestApplyAttributes(t *testing.T) {
	t.Run("map to struct with db tag", func(t *testing.T) {
		type User struct {
			Name string `db:"name"`
			Age  int    `db:"age"`
		}

		user := User{}
		attrs := map[string]any{"name": "Alice", "age": 30}

		err := applyAttributes(&user, attrs)
		if err != nil {
			t.Fatalf("applyAttributes failed: %v", err)
		}

		if user.Name != "Alice" {
			t.Errorf("expected Name 'Alice', got %s", user.Name)
		}
		if user.Age != 30 {
			t.Errorf("expected Age 30, got %d", user.Age)
		}
	})

	t.Run("map to struct with case-insensitive matching", func(t *testing.T) {
		type User struct {
			Name string
			Age  int
		}

		user := User{}
		attrs := map[string]any{"NAME": "Bob", "AGE": 25}

		err := applyAttributes(&user, attrs)
		if err != nil {
			t.Fatalf("applyAttributes failed: %v", err)
		}

		if user.Name != "Bob" {
			t.Errorf("expected Name 'Bob', got %s", user.Name)
		}
		if user.Age != 25 {
			t.Errorf("expected Age 25, got %d", user.Age)
		}
	})

	t.Run("struct to struct", func(t *testing.T) {
		type User struct {
			Name string
			Age  int
		}

		dest := User{}
		src := User{Name: "Charlie", Age: 35}

		err := applyAttributes(&dest, src)
		if err != nil {
			t.Fatalf("applyAttributes failed: %v", err)
		}

		if dest.Name != "Charlie" {
			t.Errorf("expected Name 'Charlie', got %s", dest.Name)
		}
		if dest.Age != 35 {
			t.Errorf("expected Age 35, got %d", dest.Age)
		}
	})

	t.Run("struct pointer to struct", func(t *testing.T) {
		type User struct {
			Name string
		}

		dest := User{}
		src := &User{Name: "David"}

		err := applyAttributes(&dest, src)
		if err != nil {
			t.Fatalf("applyAttributes failed: %v", err)
		}

		if dest.Name != "David" {
			t.Errorf("expected Name 'David', got %s", dest.Name)
		}
	})
}

func TestIsSimpleIdentifier(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Valid simple identifiers
		{"simple lowercase", "name", true},
		{"simple uppercase", "NAME", true},
		{"mixed case", "UserName", true},
		{"with underscore", "user_name", true},
		{"single letter", "a", true},
		{"single underscore", "_", true},
		{"multiple underscores", "user__name", true},
		{"with numbers", "user123", true},
		{"numbers after letters", "user123name", true},

		// Invalid identifiers
		{"empty string", "", false},
		{"starts with number", "123name", false},
		{"contains dot", "table.column", false},
		{"contains parentheses", "function()", false},
		{"contains opening paren", "func(", false},
		{"contains closing paren", "func)", false},
		{"contains space", "user name", false},
		{"contains hyphen", "user-name", false},
		{"contains special char", "user@name", false},
		{"starts with digit", "1name", false},
		{"only digits", "123", false},
		{"contains multiple dots", "table.schema.column", false},
		{"complex expression", "COUNT(*)", false},
		{"with comma", "user,name", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSimpleIdentifier(tt.input)
			if result != tt.expected {
				t.Errorf("isSimpleIdentifier(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}
