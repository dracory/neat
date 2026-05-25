package query

import (
	"reflect"
	"testing"
)

func TestInitializeRelations(t *testing.T) {
	q := NewQuery(nil, nil, nil, "", nil, nil)
	q.withRelations = []string{"User", "Posts"}

	type TestModel struct {
		User  *string
		Posts []int
	}

	model := &TestModel{}
	v := reflect.ValueOf(model)

	q.initializeRelations(v)

	// Check that relations are initialized (should be non-nil after initialization)
	userField := v.Elem().FieldByName("User")
	if !userField.IsValid() {
		t.Error("Expected User field to be valid")
	}
	if userField.Kind() != reflect.Ptr {
		t.Error("Expected User field to be a pointer")
	}
	if userField.IsNil() {
		t.Error("Expected User field to be non-nil after initialization")
	}

	postsField := v.Elem().FieldByName("Posts")
	if !postsField.IsValid() {
		t.Error("Expected Posts field to be valid")
	}
	if postsField.Kind() != reflect.Slice {
		t.Error("Expected Posts field to be a slice")
	}
	if postsField.IsNil() {
		t.Error("Expected Posts field to be non-nil after initialization")
	}
}

func TestInitializeRelationsWithNilValue(t *testing.T) {
	q := NewQuery(nil, nil, nil, "", nil, nil)
	q.withRelations = []string{"User"}

	var model *struct{}
	v := reflect.ValueOf(model)

	// Should not panic
	q.initializeRelations(v)
}

func TestInitializeRelationsWithNonStruct(t *testing.T) {
	q := NewQuery(nil, nil, nil, "", nil, nil)
	q.withRelations = []string{"User"}

	model := "not a struct"
	v := reflect.ValueOf(model)

	// Should not panic
	q.initializeRelations(v)
}
