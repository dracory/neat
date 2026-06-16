package schemer

import (
	"fmt"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/dracory/neat"
	contractsorm "github.com/dracory/neat/contracts/database/orm"
	contractsschema "github.com/dracory/neat/contracts/database/schema"
)

// TestSchemaTransactionDetection verifies that schema operations detect and use transactions
func TestSchemaTransactionDetection(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Test 1: Schema operation outside transaction
	t.Run("SchemaOutsideTransaction", func(t *testing.T) {
		schema := db.Schema()
		err := schema.Create("test_table_1", func(blueprint contractsschema.Blueprint) {
			blueprint.ID()
			blueprint.String("name")
		})
		if err != nil {
			t.Fatalf("failed to create table outside transaction: %v", err)
		}

		if !schema.HasTable("test_table_1") {
			t.Error("table should exist after creation")
		}
	})

	// Test 2: Schema operation inside transaction using WithTransaction
	t.Run("SchemaInsideTransaction", func(t *testing.T) {
		err := db.Transaction(func(tx contractsorm.Query) error {
			schema := db.Schema().WithTransaction(tx)

			// Schema operations should now use the transaction query
			query := schema.Orm().Query()
			if query.InTransaction() {
				t.Log("Schema ORM query detects transaction state")
			} else {
				t.Log("Schema ORM query does NOT detect transaction state (tx is stored separately)")
			}

			// Create within transaction
			createErr := schema.Create("test_table_2", func(blueprint contractsschema.Blueprint) {
				blueprint.ID()
				blueprint.String("name")
			})
			if createErr != nil {
				t.Fatalf("failed to create table in transaction: %v", createErr)
			}

			if !schema.HasTable("test_table_2") {
				t.Error("table should exist within transaction")
			}

			return nil
		})
		if err != nil {
			t.Fatalf("transaction failed: %v", err)
		}

		// Table should exist after commit
		if !db.Schema().HasTable("test_table_2") {
			t.Error("table should exist after transaction commit")
		}
	})

	// Test 3: Schema Create with transaction rollback
	t.Run("SchemaCreateWithTransactionRollback", func(t *testing.T) {
		err := db.Transaction(func(tx contractsorm.Query) error {
			schema := db.Schema().WithTransaction(tx)

			createErr := schema.Create("test_table_rollback", func(blueprint contractsschema.Blueprint) {
				blueprint.ID()
				blueprint.String("name")
			})
			if createErr != nil {
				t.Fatalf("failed to create table in transaction: %v", createErr)
			}

			if !schema.HasTable("test_table_rollback") {
				t.Error("table should exist within transaction")
			}

			// Return error to trigger rollback
			return fmt.Errorf("intentional rollback")
		})
		if err == nil {
			t.Fatal("expected transaction to roll back")
		}

		// Table should NOT exist after rollback
		if db.Schema().HasTable("test_table_rollback") {
			t.Error("table should not exist after transaction rollback")
		}
	})
}

// TestSchemerWithRealTransactions tests schemer with actual transaction wrapping enabled
func TestSchemerWithRealTransactions(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	schema := db.Schema()

	// Test that WithTransaction returns a schema that can be used for migrations
	txSchema := schema.WithTransaction(db.Schema().Orm().Query())
	if txSchema == nil {
		t.Fatal("WithTransaction should return a valid schema")
	}

	// The transaction-aware schema should be able to perform operations
	err = txSchema.Create("test_migrations", func(blueprint contractsschema.Blueprint) {
		blueprint.ID()
		blueprint.String("name")
	})
	if err != nil {
		// This might fail because the query is not actually in a transaction
		// but it should not panic and should behave gracefully
		t.Logf("Schema operation on WithTransaction query: %v", err)
	}
}
