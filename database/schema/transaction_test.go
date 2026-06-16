package schema_test

import (
	"fmt"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/dracory/neat"
	contractsorm "github.com/dracory/neat/contracts/database/orm"
	contractsschema "github.com/dracory/neat/contracts/database/schema"
)

func TestSchemaWithTransaction(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	t.Run("CreateAndRollback", func(t *testing.T) {
		err := db.Transaction(func(tx contractsorm.Query) error {
			schema := db.Schema().WithTransaction(tx)

			// Create table within transaction
			createErr := schema.Create("users", func(blueprint contractsschema.Blueprint) {
				blueprint.ID()
				blueprint.String("name")
			})
			if createErr != nil {
				t.Fatalf("failed to create table: %v", createErr)
			}

			// Table should be visible within transaction
			if !schema.HasTable("users") {
				t.Error("table should exist within transaction after creation")
			}

			// Return error to trigger rollback
			return fmt.Errorf("intentional rollback")
		})

		if err == nil {
			t.Fatal("expected transaction to roll back")
		}

		// Table should NOT exist after rollback
		schema := db.Schema()
		if schema.HasTable("users") {
			t.Error("table should not exist after transaction rollback")
		}
	})

	t.Run("CreateAndCommit", func(t *testing.T) {
		err := db.Transaction(func(tx contractsorm.Query) error {
			schema := db.Schema().WithTransaction(tx)

			createErr := schema.Create("products", func(blueprint contractsschema.Blueprint) {
				blueprint.ID()
				blueprint.String("name")
			})
			if createErr != nil {
				t.Fatalf("failed to create table: %v", createErr)
			}

			return nil
		})

		if err != nil {
			t.Fatalf("transaction failed: %v", err)
		}

		schema := db.Schema()
		if !schema.HasTable("products") {
			t.Error("table should exist after transaction commit")
		}
	})

	t.Run("DropAndRollback", func(t *testing.T) {
		// First create a table
		schema := db.Schema()
		if err := schema.Create("categories", func(blueprint contractsschema.Blueprint) {
			blueprint.ID()
			blueprint.String("name")
		}); err != nil {
			t.Fatalf("failed to create table: %v", err)
		}

		// Now drop it in a transaction and roll back
		err := db.Transaction(func(tx contractsorm.Query) error {
			txSchema := db.Schema().WithTransaction(tx)

			dropErr := txSchema.Drop("categories")
			if dropErr != nil {
				t.Fatalf("failed to drop table: %v", dropErr)
			}

			if txSchema.HasTable("categories") {
				t.Error("table should not exist within transaction after drop")
			}

			return fmt.Errorf("intentional rollback")
		})

		if err == nil {
			t.Fatal("expected transaction to roll back")
		}

		// Table should still exist after rollback
		if !schema.HasTable("categories") {
			t.Error("table should still exist after rollback")
		}
	})

	t.Run("TableModifyAndRollback", func(t *testing.T) {
		// Create a table
		schema := db.Schema()
		if err := schema.Create("orders", func(blueprint contractsschema.Blueprint) {
			blueprint.ID()
			blueprint.String("status")
		}); err != nil {
			t.Fatalf("failed to create table: %v", err)
		}

		// Modify it in a transaction and roll back
		err := db.Transaction(func(tx contractsorm.Query) error {
			txSchema := db.Schema().WithTransaction(tx)

			modifyErr := txSchema.Table("orders", func(blueprint contractsschema.Blueprint) {
				blueprint.String("description")
			})
			if modifyErr != nil {
				t.Fatalf("failed to modify table: %v", modifyErr)
			}

			if !txSchema.HasColumn("orders", "description") {
				t.Error("new column should exist within transaction")
			}

			return fmt.Errorf("intentional rollback")
		})

		if err == nil {
			t.Fatal("expected transaction to roll back")
		}

		// Column should NOT exist after rollback
		if schema.HasColumn("orders", "description") {
			t.Error("column should not exist after rollback")
		}
	})

	t.Run("WithoutTransactionStillWorks", func(t *testing.T) {
		schema := db.Schema()

		if err := schema.Create("standalone", func(blueprint contractsschema.Blueprint) {
			blueprint.ID()
		}); err != nil {
			t.Fatalf("failed to create table: %v", err)
		}

		if !schema.HasTable("standalone") {
			t.Error("table should exist")
		}
	})
}
