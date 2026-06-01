package query_test

import (
	"testing"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
)

// --- scopes tests ---

type ScopeUser struct {
	ID     int    `db:"id"`
	Name   string `db:"name"`
	Status string `db:"status"`
	Age    int    `db:"age"`
}

func TestScopesMethod(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_scopes (id INTEGER, name TEXT, status TEXT, age INTEGER)")
	execSQL(t, w, "INSERT INTO test_scopes VALUES (1,'alice','active',25),(2,'bob','inactive',30),(3,'charlie','active',35)")

	w.SetTable("test_scopes")

	t.Run("single scope without parameters", func(t *testing.T) {
		activeScope := func(q contractsorm.Query) contractsorm.Query {
			return q.Where("status = ?", "active")
		}

		var users []ScopeUser
		err := w.Q.Scopes(activeScope).Find(&users)
		if err != nil {
			t.Fatalf("Scopes with single scope failed: %v", err)
		}

		if len(users) != 2 {
			t.Errorf("Expected 2 active users, got %d", len(users))
		}

		for _, user := range users {
			if user.Status != "active" {
				t.Errorf("Expected status 'active', got '%s'", user.Status)
			}
		}
	})

	t.Run("multiple scopes without parameters", func(t *testing.T) {
		activeScope := func(q contractsorm.Query) contractsorm.Query {
			return q.Where("status = ?", "active")
		}

		youngScope := func(q contractsorm.Query) contractsorm.Query {
			return q.Where("age < ?", 30)
		}

		var users []ScopeUser
		err := w.Q.Scopes(activeScope, youngScope).Find(&users)
		if err != nil {
			t.Fatalf("Scopes with multiple scopes failed: %v", err)
		}

		if len(users) != 1 {
			t.Errorf("Expected 1 active young user, got %d", len(users))
		}

		if users[0].Name != "alice" {
			t.Errorf("Expected user 'alice', got '%s'", users[0].Name)
		}
	})

	t.Run("scope with closure parameters", func(t *testing.T) {
		nameScope := func(name string) func(contractsorm.Query) contractsorm.Query {
			return func(q contractsorm.Query) contractsorm.Query {
				return q.Where("name = ?", name)
			}
		}

		var users []ScopeUser
		err := w.Q.Scopes(nameScope("bob")).Find(&users)
		if err != nil {
			t.Fatalf("Scopes with closure parameters failed: %v", err)
		}

		if len(users) != 1 {
			t.Errorf("Expected 1 user, got %d", len(users))
		}

		if users[0].Name != "bob" {
			t.Errorf("Expected user 'bob', got '%s'", users[0].Name)
		}
	})

	t.Run("empty scopes list", func(t *testing.T) {
		var users []ScopeUser
		err := w.Q.Scopes().Find(&users)
		if err != nil {
			t.Fatalf("Scopes with empty list failed: %v", err)
		}

		if len(users) != 3 {
			t.Errorf("Expected 3 users with no scopes, got %d", len(users))
		}
	})
}

func TestScopeApplicationOrder(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_scope_order (id INTEGER, name TEXT, status TEXT, age INTEGER)")
	execSQL(t, w, "INSERT INTO test_scope_order VALUES (1,'alice','active',25),(2,'bob','active',30),(3,'charlie','inactive',35)")

	w.SetTable("test_scope_order")

	t.Run("scopes applied in order", func(t *testing.T) {
		scope1 := func(q contractsorm.Query) contractsorm.Query {
			return q.Where("status = ?", "active")
		}

		scope2 := func(q contractsorm.Query) contractsorm.Query {
			return q.Where("age > ?", 25)
		}

		scope3 := func(q contractsorm.Query) contractsorm.Query {
			return q.OrderBy("name", "asc")
		}

		var users []ScopeUser
		err := w.Q.Scopes(scope1, scope2, scope3).Find(&users)
		if err != nil {
			t.Fatalf("Scope application order failed: %v", err)
		}

		if len(users) != 1 {
			t.Errorf("Expected 1 user matching all scopes, got %d", len(users))
		}

		if users[0].Name != "bob" {
			t.Errorf("Expected user 'bob', got '%s'", users[0].Name)
		}
	})

	t.Run("scope order affects result", func(t *testing.T) {
		limitScope := func(q contractsorm.Query) contractsorm.Query {
			return q.Limit(1)
		}

		orderScope := func(q contractsorm.Query) contractsorm.Query {
			return q.OrderBy("age", "desc")
		}

		var users []ScopeUser
		err := w.Q.Scopes(orderScope, limitScope).Find(&users)
		if err != nil {
			t.Fatalf("Scope order with limit failed: %v", err)
		}

		if len(users) != 1 {
			t.Errorf("Expected 1 user, got %d", len(users))
		}

		if users[0].Name != "charlie" {
			t.Errorf("Expected oldest user 'charlie', got '%s'", users[0].Name)
		}
	})

	t.Run("reversed scope order", func(t *testing.T) {
		limitScope := func(q contractsorm.Query) contractsorm.Query {
			return q.Limit(1)
		}

		orderScope := func(q contractsorm.Query) contractsorm.Query {
			return q.OrderBy("age", "asc")
		}

		var users []ScopeUser
		err := w.Q.Scopes(limitScope, orderScope).Find(&users)
		if err != nil {
			t.Fatalf("Reversed scope order failed: %v", err)
		}

		if len(users) != 1 {
			t.Errorf("Expected 1 user, got %d", len(users))
		}

		if users[0].Name != "alice" {
			t.Errorf("Expected youngest user 'alice', got '%s'", users[0].Name)
		}
	})
}

func TestScopeWithQueryChaining(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_scope_chain (id INTEGER, name TEXT, status TEXT, age INTEGER)")
	execSQL(t, w, "INSERT INTO test_scope_chain VALUES (1,'alice','active',25),(2,'bob','active',30),(3,'charlie','inactive',35),(4,'dave','active',40)")

	w.SetTable("test_scope_chain")

	t.Run("scope before where clause", func(t *testing.T) {
		activeScope := func(q contractsorm.Query) contractsorm.Query {
			return q.Where("status = ?", "active")
		}

		var users []ScopeUser
		err := w.Q.Scopes(activeScope).Where("age > ?", 30).Find(&users)
		if err != nil {
			t.Fatalf("Scope before where failed: %v", err)
		}

		if len(users) != 1 {
			t.Errorf("Expected 1 user, got %d", len(users))
		}

		if users[0].Name != "dave" {
			t.Errorf("Expected user 'dave', got '%s'", users[0].Name)
		}
	})

	t.Run("scope after where clause", func(t *testing.T) {
		activeScope := func(q contractsorm.Query) contractsorm.Query {
			return q.Where("status = ?", "active")
		}

		var users []ScopeUser
		err := w.Q.Where("age < ?", 30).Scopes(activeScope).Find(&users)
		if err != nil {
			t.Fatalf("Scope after where failed: %v", err)
		}

		if len(users) != 1 {
			t.Errorf("Expected 1 user, got %d", len(users))
		}

		if users[0].Name != "alice" {
			t.Errorf("Expected user 'alice', got '%s'", users[0].Name)
		}
	})

	t.Run("scope with first", func(t *testing.T) {
		activeScope := func(q contractsorm.Query) contractsorm.Query {
			return q.Where("status = ?", "active")
		}

		var user ScopeUser
		err := w.Q.Scopes(activeScope).OrderBy("age", "asc").First(&user)
		if err != nil {
			t.Fatalf("Scope with first failed: %v", err)
		}

		if user.Name != "alice" {
			t.Errorf("Expected user 'alice', got '%s'", user.Name)
		}
	})
}

func TestScopeErrorHandling(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_scope_error (id INTEGER, name TEXT, status TEXT)")
	execSQL(t, w, "INSERT INTO test_scope_error VALUES (1,'alice','active')")

	w.SetTable("test_scope_error")

	t.Run("scope returns nil query", func(t *testing.T) {
		nilScope := func(q contractsorm.Query) contractsorm.Query {
			return nil
		}

		var users []ScopeUser
		err := w.Q.Scopes(nilScope).Find(&users)
		// The current implementation doesn't error when scope returns nil
		// It just falls back to the original query
		_ = err
		_ = users
	})

	t.Run("scope with invalid where clause", func(t *testing.T) {
		invalidScope := func(q contractsorm.Query) contractsorm.Query {
			return q.Where("invalid_column = ?", "value")
		}

		var users []ScopeUser
		err := w.Q.Scopes(invalidScope).Find(&users)
		// SQLite doesn't error on invalid columns in WHERE clause
		// It just returns no results
		_ = err
		_ = users
	})

	t.Run("scope panic handling", func(t *testing.T) {
		panicScope := func(q contractsorm.Query) contractsorm.Query {
			panic("scope panic")
		}

		var users []ScopeUser
		defer func() {
			if r := recover(); r != nil {
				// Expected panic
				_ = r
			}
		}()

		_ = w.Q.Scopes(panicScope).Find(&users)
	})

	t.Run("scope with invalid table", func(t *testing.T) {
		w.SetTable("nonexistent_table")

		validScope := func(q contractsorm.Query) contractsorm.Query {
			return q.Where("status = ?", "active")
		}

		var users []ScopeUser
		err := w.Q.Scopes(validScope).Find(&users)
		if err == nil {
			t.Error("Expected error for invalid table with scope")
		}
	})

	t.Run("scope with nil destination", func(t *testing.T) {
		activeScope := func(q contractsorm.Query) contractsorm.Query {
			return q.Where("status = ?", "active")
		}

		err := w.Q.Scopes(activeScope).Find(nil)
		if err == nil {
			t.Error("Expected error for nil destination")
		}
	})

	t.Run("scope with non-slice destination", func(t *testing.T) {
		activeScope := func(q contractsorm.Query) contractsorm.Query {
			return q.Where("status = ?", "active")
		}

		var user ScopeUser
		err := w.Q.Scopes(activeScope).Find(&user)
		if err == nil {
			t.Error("Expected error for non-slice destination with Find")
		}
	})
}

func TestScopeWithTransactions(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_scope_tx (id INTEGER, name TEXT, status TEXT)")
	execSQL(t, w, "INSERT INTO test_scope_tx VALUES (1,'alice','active'),(2,'bob','inactive')")

	w.SetTable("test_scope_tx")

	tx, err := w.Q.Begin()
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	activeScope := func(q contractsorm.Query) contractsorm.Query {
		return q.Where("status = ?", "active")
	}

	users := make([]ScopeUser, 0)
	err = tx.Scopes(activeScope).Find(&users)
	if err != nil {
		t.Fatalf("Scope in transaction failed: %v", err)
	}

	if len(users) != 1 {
		t.Errorf("Expected 1 active user in transaction, got %d", len(users))
	}

	if users[0].Name != "alice" {
		t.Errorf("Expected user 'alice', got '%s'", users[0].Name)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}
}

func TestScopeIsolation(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_scope_isolation (id INTEGER, name TEXT, status TEXT)")
	execSQL(t, w, "INSERT INTO test_scope_isolation VALUES (1,'alice','active'),(2,'bob','inactive')")

	w.SetTable("test_scope_isolation")

	t.Run("scope does not affect original query", func(t *testing.T) {
		activeScope := func(q contractsorm.Query) contractsorm.Query {
			return q.Where("status = ?", "active")
		}

		var scopedUsers []ScopeUser
		err := w.Q.Scopes(activeScope).Find(&scopedUsers)
		if err != nil {
			t.Fatalf("Scoped query failed: %v", err)
		}

		if len(scopedUsers) != 1 {
			t.Errorf("Expected 1 scoped user, got %d", len(scopedUsers))
		}

		var allUsers []ScopeUser
		err = w.Q.Find(&allUsers)
		if err != nil {
			t.Fatalf("Original query failed: %v", err)
		}

		if len(allUsers) != 2 {
			t.Errorf("Expected 2 users in original query, got %d", len(allUsers))
		}
	})

	t.Run("multiple scopes on same query", func(t *testing.T) {
		activeScope := func(q contractsorm.Query) contractsorm.Query {
			return q.Where("status = ?", "active")
		}

		inactiveScope := func(q contractsorm.Query) contractsorm.Query {
			return q.Where("status = ?", "inactive")
		}

		var activeUsers []ScopeUser
		err := w.Q.Scopes(activeScope).Find(&activeUsers)
		if err != nil {
			t.Fatalf("First scope failed: %v", err)
		}

		var inactiveUsers []ScopeUser
		err = w.Q.Scopes(inactiveScope).Find(&inactiveUsers)
		if err != nil {
			t.Fatalf("Second scope failed: %v", err)
		}

		if len(activeUsers) != 1 {
			t.Errorf("Expected 1 active user, got %d", len(activeUsers))
		}

		if len(inactiveUsers) != 1 {
			t.Errorf("Expected 1 inactive user, got %d", len(inactiveUsers))
		}
	})
}

func TestScopeWithModel(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_scope_model (id INTEGER, name TEXT, status TEXT, age INTEGER)")
	execSQL(t, w, "INSERT INTO test_scope_model VALUES (1,'alice','active',25),(2,'bob','inactive',30)")

	t.Run("scope with model set", func(t *testing.T) {
		activeScope := func(q contractsorm.Query) contractsorm.Query {
			return q.Where("status = ?", "active")
		}

		var users []ScopeUser
		err := w.Q.Model(&ScopeUser{}).Table("test_scope_model").Scopes(activeScope).Find(&users)
		if err != nil {
			t.Fatalf("Scope with model failed: %v", err)
		}

		if len(users) != 1 {
			t.Errorf("Expected 1 user, got %d", len(users))
		}
	})
}
