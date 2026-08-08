package structref

import (
	"reflect"
	"testing"
)

func TestFieldColumnName(t *testing.T) {
	type User struct {
		ID       int    `db:"id"`
		Name     string `db:"name"`
		Email    string `neat:"email"`
		Password string `gorm:"column:password"`
		NoTag    string
		Ignore   string `db:"-"`
		Priority string `db:"db_prio" neat:"neat_prio" gorm:"column:gorm_prio"`
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

	uType := reflect.TypeOf(User{})
	prioField := uType.Field(6)
	if name := FieldColumnName(prioField); name != "db_prio" {
		t.Errorf("Priority test failed: expected db_prio, got %s", name)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FieldColumnName(tt.field)
			if result != tt.expected {
				t.Errorf("FieldColumnName(%v) = %q, want %q", tt.field.Name, result, tt.expected)
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
		{"HTTPServer", "httpserver"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := CamelToSnake(tt.input)
			if result != tt.expected {
				t.Errorf("CamelToSnake(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
