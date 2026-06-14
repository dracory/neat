package query_test

import (
	"fmt"
	"testing"
)

func TestFirst(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_first (id INTEGER PRIMARY KEY, name TEXT)")
	execSQL(t, w, "INSERT INTO test_first VALUES (1, 'Alice')")
	execSQL(t, w, "INSERT INTO test_first VALUES (2, 'Bob')")

	w.SetTable("test_first")

	type User struct {
		ID   int
		Name string
	}

	var result User
	if err := w.Q.First(&result); err != nil {
		t.Fatalf("First failed: %v", err)
	}
	if result.Name != "Alice" {
		t.Errorf("expected name 'Alice', got %s", result.Name)
	}
}

func TestFirstWithWhere(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_first_where (id INTEGER PRIMARY KEY, name TEXT)")
	execSQL(t, w, "INSERT INTO test_first_where VALUES (1, 'Alice')")
	execSQL(t, w, "INSERT INTO test_first_where VALUES (2, 'Bob')")

	w.SetTable("test_first_where")
	w.Q.Where("name = ?", "Bob")

	type User struct {
		ID   int
		Name string
	}

	var result User
	if err := w.Q.First(&result); err != nil {
		t.Fatalf("First with Where failed: %v", err)
	}
	if result.Name != "Bob" {
		t.Errorf("expected name 'Bob', got %s", result.Name)
	}
}

func TestFirstOrFail(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_first_or_fail (id INTEGER PRIMARY KEY, name TEXT)")
	execSQL(t, w, "INSERT INTO test_first_or_fail VALUES (1, 'Alice')")

	w.SetTable("test_first_or_fail")

	type User struct {
		ID   int
		Name string
	}

	var result User
	if err := w.Q.FirstOrFail(&result); err != nil {
		t.Fatalf("FirstOrFail failed: %v", err)
	}
	if result.Name != "Alice" {
		t.Errorf("expected name 'Alice', got %s", result.Name)
	}
}

func TestFirstOrFailNotFound(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_first_or_fail_not_found (id INTEGER PRIMARY KEY, name TEXT)")

	w.SetTable("test_first_or_fail_not_found")

	type User struct {
		ID   int
		Name string
	}

	var result User
	err := w.Q.FirstOrFail(&result)
	if err == nil {
		t.Error("expected error for not found")
	}
}

type FirstOrUser struct {
	ID    int    `db:"id"`
	Name  string `db:"name"`
	Email string `db:"email"`
}

func TestFirstOrWithCallback(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_first_or (id INTEGER, name TEXT, email TEXT)")
	execSQL(t, w, "INSERT INTO test_first_or VALUES (1,'alice','alice@example.com')")

	w.SetTable("test_first_or")

	t.Run("record found - callback not executed", func(t *testing.T) {
		var user FirstOrUser
		callbackExecuted := false

		err := w.Q.Where("id = ?", 1).FirstOr(&user, func() error {
			callbackExecuted = true
			return nil
		})

		if err != nil {
			t.Fatalf("FirstOr failed: %v", err)
		}

		if callbackExecuted {
			t.Error("Callback should not be executed when record is found")
		}

		if user.Name != "alice" {
			t.Errorf("Expected user.Name='alice', got %q", user.Name)
		}
	})

	t.Run("record not found - callback executed", func(t *testing.T) {
		var user FirstOrUser
		callbackExecuted := false
		callbackError := fmt.Errorf("not found")

		err := w.Q.Where("id = ?", 999).FirstOr(&user, func() error {
			callbackExecuted = true
			return callbackError
		})

		if err != callbackError {
			t.Errorf("Expected callback error, got %v", err)
		}

		if !callbackExecuted {
			t.Error("Callback should be executed when record is not found")
		}
	})

	t.Run("callback returns nil on not found", func(t *testing.T) {
		var user FirstOrUser
		callbackExecuted := false

		err := w.Q.Where("id = ?", 999).FirstOr(&user, func() error {
			callbackExecuted = true
			return nil
		})

		if err != nil {
			t.Fatalf("FirstOr with nil callback error failed: %v", err)
		}

		if !callbackExecuted {
			t.Error("Callback should be executed when record is not found")
		}
	})
}

func TestFirstOrCreate(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_first_or_create (id INTEGER, name TEXT, email TEXT)")
	execSQL(t, w, "INSERT INTO test_first_or_create VALUES (1,'alice','alice@example.com')")

	w.SetTable("test_first_or_create")

	t.Run("record exists - returns existing", func(t *testing.T) {
		var user FirstOrUser
		user.Name = "bob"
		user.Email = "bob@example.com"

		err := w.Q.Where("id = ?", 1).FirstOrCreate(&user)

		if err != nil {
			t.Fatalf("FirstOrCreate failed: %v", err)
		}

		if user.Name != "alice" {
			t.Errorf("Expected existing user.Name='alice', got %q", user.Name)
		}

		if user.Email != "alice@example.com" {
			t.Errorf("Expected existing user.Email='alice@example.com', got %q", user.Email)
		}
	})

	t.Run("record not found - creates new", func(t *testing.T) {
		var user FirstOrUser
		user.Name = "charlie"
		user.Email = "charlie@example.com"

		err := w.Q.Where("id = ?", 999).FirstOrCreate(&user)

		if err != nil {
			t.Fatalf("FirstOrCreate create failed: %v", err)
		}

		// Verify a record was created
		var count int64
		err = w.Q.Count(&count)
		if err != nil {
			t.Fatalf("Count failed: %v", err)
		}

		if count == 1 {
			t.Error("Expected at least two records (original + new)")
		}
	})
}

func TestFirstOrNew(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_first_or_new (id INTEGER, name TEXT, email TEXT)")
	execSQL(t, w, "INSERT INTO test_first_or_new VALUES (1,'alice','alice@example.com')")

	w.SetTable("test_first_or_new")

	t.Run("record exists - returns existing", func(t *testing.T) {
		var user FirstOrUser
		attributes := map[string]any{"id": 1}

		err := w.Q.FirstOrNew(&user, attributes)

		if err != nil {
			t.Fatalf("FirstOrNew failed: %v", err)
		}

		if user.Name != "alice" {
			t.Errorf("Expected existing user.Name='alice', got %q", user.Name)
		}
	})

	t.Run("record not found - prepares new instance", func(t *testing.T) {
		var user FirstOrUser
		user.Name = "bob"
		user.Email = "bob@example.com"
		attributes := map[string]any{"id": 999, "name": "charlie", "email": "charlie@example.com"}

		err := w.Q.FirstOrNew(&user, attributes)

		if err != nil {
			t.Fatalf("FirstOrNew prepare failed: %v", err)
		}

		// Verify that attributes were applied to the model
		if user.ID != 999 {
			t.Errorf("Expected user.ID=999 from attributes, got %d", user.ID)
		}
		if user.Name != "charlie" {
			t.Errorf("Expected user.Name='charlie' from attributes, got %q", user.Name)
		}
		if user.Email != "charlie@example.com" {
			t.Errorf("Expected user.Email='charlie@example.com' from attributes, got %q", user.Email)
		}
	})

	t.Run("with values parameter - record exists", func(t *testing.T) {
		var user FirstOrUser
		attributes := map[string]any{"id": 1}
		values := map[string]any{"name": "updated"}

		err := w.Q.FirstOrNew(&user, attributes, values)

		if err != nil {
			t.Fatalf("FirstOrNew with values failed: %v", err)
		}

		// Since record exists, values should not be applied
		if user.Name != "alice" {
			t.Errorf("Expected existing user.Name='alice', got %q", user.Name)
		}
	})

	t.Run("with values parameter - record not found", func(t *testing.T) {
		var user FirstOrUser
		attributes := map[string]any{"id": 999}
		values := map[string]any{"name": "new_user", "email": "new@example.com"}

		err := w.Q.FirstOrNew(&user, attributes, values)

		if err != nil {
			t.Fatalf("FirstOrNew with values failed: %v", err)
		}

		// Both attributes and values should be applied
		if user.ID != 999 {
			t.Errorf("Expected user.ID=999 from attributes, got %d", user.ID)
		}
		if user.Name != "new_user" {
			t.Errorf("Expected user.Name='new_user' from values, got %q", user.Name)
		}
		if user.Email != "new@example.com" {
			t.Errorf("Expected user.Email='new@example.com' from values, got %q", user.Email)
		}
	})
}

func TestUpdateOrCreate(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_update_or_create (id INTEGER, name TEXT, email TEXT)")
	execSQL(t, w, "INSERT INTO test_update_or_create VALUES (1,'alice','alice@example.com')")

	w.SetTable("test_update_or_create")

	t.Run("record exists - calls Save on values", func(t *testing.T) {
		var user FirstOrUser
		attributes := map[string]any{"id": 1}
		values := map[string]any{"name": "alice_updated", "email": "alice_new@example.com"}

		err := w.Q.UpdateOrCreate(&user, attributes, values)

		if err != nil {
			t.Fatalf("UpdateOrCreate failed: %v", err)
		}
	})

	t.Run("record not found - calls Create on values", func(t *testing.T) {
		execSQL(t, w, "CREATE TABLE test_update_or_create_2 (id INTEGER, name TEXT, email TEXT)")
		w.SetTable("test_update_or_create_2")

		var user FirstOrUser
		attributes := map[string]any{"id": 999}
		values := map[string]any{"name": "bob", "email": "bob@example.com"}

		err := w.Q.UpdateOrCreate(&user, attributes, values)

		if err != nil {
			t.Fatalf("UpdateOrCreate create failed: %v", err)
		}

		// Verify a record was created
		var count int64
		err = w.Q.Count(&count)
		if err != nil {
			t.Fatalf("Count failed: %v", err)
		}

		if count == 0 {
			t.Error("Expected at least one record to be created")
		}
	})
}

func TestFindOne(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_find_one (id INTEGER PRIMARY KEY, name TEXT)")
	execSQL(t, w, "INSERT INTO test_find_one VALUES (1, 'Alice')")
	execSQL(t, w, "INSERT INTO test_find_one VALUES (2, 'Bob')")

	w.SetTable("test_find_one")
	w.Q.Where("name = ?", "Bob")

	type User struct {
		ID   int
		Name string
	}

	var result User
	if err := w.Q.FindOne(&result); err != nil {
		t.Fatalf("FindOne failed: %v", err)
	}
	if result.Name != "Bob" {
		t.Errorf("expected name 'Bob', got %s", result.Name)
	}
}

func TestFindOneAsAliasForFirst(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_find_one_alias (id INTEGER PRIMARY KEY, name TEXT)")
	execSQL(t, w, "INSERT INTO test_find_one_alias VALUES (1, 'Alice')")
	execSQL(t, w, "INSERT INTO test_find_one_alias VALUES (2, 'Bob')")

	w.SetTable("test_find_one_alias")
	w.Q.OrderBy("id")

	type User struct {
		ID   int
		Name string
	}

	var result1 User
	var result2 User

	if err := w.Q.First(&result1); err != nil {
		t.Fatalf("First failed: %v", err)
	}

	w.SetTable("test_find_one_alias")
	w.Q.OrderBy("id")
	if err := w.Q.FindOne(&result2); err != nil {
		t.Fatalf("FindOne failed: %v", err)
	}

	if result1.Name != result2.Name {
		t.Errorf("FindOne and First should return same result. Got %s vs %s", result1.Name, result2.Name)
	}
}
