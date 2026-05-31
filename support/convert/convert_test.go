package convert

import (
	"strconv"
	"testing"
	"time"
)

type foo struct {
	Name string
	Age  int
}

// TestTap tests the Tap function.
func TestTap(t *testing.T) {
	// pointer
	f := &foo{Name: "foo"}

	if f.Name != "foo" {
		t.Errorf("expected foo, got %s", f.Name)
	}
	if f.Age != 0 {
		t.Errorf("expected 0, got %d", f.Age)
	}

	got1 := Tap(f, func(f *foo) {
		f.Name = "bar" //nolint:goconst
		f.Age = 18
	})
	if got1.Name != "bar" {
		t.Errorf("expected bar, got %s", got1.Name)
	}
	if got1.Age != 18 {
		t.Errorf("expected 18, got %d", got1.Age)
	}

	// int
	got2 := Tap(10, func(i int) {
		if i != 10 {
			t.Errorf("expected 10, got %d", i)
		}
		i = 20
		if i != 20 {
			t.Errorf("expected 20, got %d", i)
		}
	})
	if got2 != 10 {
		t.Errorf("expected 10, got %d", got2)
	}

	// string
	got3 := Tap("foo", func(s string) {
		if s != "foo" {
			t.Errorf("expected foo, got %s", s)
		}
		s = "bar"
		if s != "bar" {
			t.Errorf("expected bar, got %s", s)
		}
	})
	if got3 != "foo" {
		t.Errorf("expected foo, got %s", got3)
	}
}

// TestWith tests the With function.
func TestWith(t *testing.T) {
	// pointer
	f := &foo{Name: "foo"}

	if f.Name != "foo" {
		t.Errorf("expected foo, got %s", f.Name)
	}
	if f.Age != 0 {
		t.Errorf("expected 0, got %d", f.Age)
	}

	got1 := With(f, func(f *foo) *foo {
		f.Name = "bar" //nolint:goconst
		f.Age = 18
		return f
	})
	if got1.Name != "bar" {
		t.Errorf("expected bar, got %s", got1.Name)
	}
	if got1.Age != 18 {
		t.Errorf("expected 18, got %d", got1.Age)
	}

	// int
	got2 := With(10, func(i int) int {
		return i + 10
	})
	if got2 != 20 {
		t.Errorf("expected 20, got %d", got2)
	}

	// string
	got3 := With("foo", func(s string) string {
		return s + "bar"
	})
	if got3 != "foobar" {
		t.Errorf("expected foobar, got %s", got3)
	}
}

// TestTransform tests the Transform function.
func TestTransform(t *testing.T) {
	if got := Transform(1, strconv.Itoa); got != "1" {
		t.Errorf("expected 1, got %s", got)
	}
	expected := &foo{Name: "foo"}
	if got := Transform("foo", func(s string) *foo {
		return &foo{Name: s}
	}); got.Name != expected.Name {
		t.Errorf("expected %v, got %v", expected, got)
	}
}

// TestDefault tests the Default function.
func TestDefault(t *testing.T) {
	// string
	if got := Default("", "foo"); got != "foo" {
		t.Errorf("expected foo, got %s", got)
	}
	if got := Default("bar", "foo"); got != "bar" {
		t.Errorf("expected bar, got %s", got)
	}
	if got := Default("", "", "foo"); got != "foo" {
		t.Errorf("expected foo, got %s", got)
	}

	// int
	if got := Default(0, 1); got != 1 {
		t.Errorf("expected 1, got %d", got)
	}
	if got := Default(2, 1); got != 2 {
		t.Errorf("expected 2, got %d", got)
	}
	if got := Default(0, 0, 1); got != 1 {
		t.Errorf("expected 1, got %d", got)
	}

	// pointer
	expected1 := &foo{Name: "foo"}
	if got := Default(nil, &foo{Name: "foo"}); got.Name != expected1.Name {
		t.Errorf("expected %v, got %v", expected1, got)
	}
	expected2 := &foo{Name: "bar"}
	if got := Default(&foo{Name: "bar"}, &foo{Name: "foo"}); got.Name != expected2.Name {
		t.Errorf("expected %v, got %v", expected2, got)
	}

	// struct
	if got := Default(foo{}, foo{Name: "foo"}); got.Name != "foo" {
		t.Errorf("expected foo, got %s", got.Name)
	}
	if got := Default(foo{Name: "bar"}, foo{Name: "foo"}); got.Name != "bar" {
		t.Errorf("expected bar, got %s", got.Name)
	}

	// zero
	if got := Default(0, 0); got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}

// TestPointer tests the Pointer function.
func TestPointer(t *testing.T) {
	if got := *Pointer("foo"); got != "foo" {
		t.Errorf("expected foo, got %s", got)
	}
	if got := *Pointer(1); got != 1 {
		t.Errorf("expected 1, got %d", got)
	}
	expected := &foo{Name: "foo"}
	if got := *Pointer(&foo{Name: "foo"}); got.Name != expected.Name {
		t.Errorf("expected %v, got %v", expected, got)
	}
	if got := *Pointer(time.Time{}); !got.IsZero() {
		t.Errorf("expected zero time, got %v", got)
	}
}
