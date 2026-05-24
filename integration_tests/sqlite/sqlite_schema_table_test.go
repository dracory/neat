package sqlite

import (
	"testing"

	"github.com/dracory/neat"
	"github.com/dracory/neat/contracts/database/schema"
)

func TestSQLiteSchemaTableCreateHasDrop(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)

	tableName := "test_table"

	_ = db.Schema().DropIfExists(tableName)
	if db.Schema().HasTable(tableName) {
		t.Error("Table should not exist after DropIfExists")
	}

	err := db.Schema().Create(tableName, func(table schema.Blueprint) {
		table.ID()
		table.String("name")
	})
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}
	if !db.Schema().HasTable(tableName) {
		t.Error("Table should exist after Create")
	}

	err = db.Schema().Drop(tableName)
	if err != nil {
		t.Fatalf("Failed to drop table: %v", err)
	}
	if db.Schema().HasTable(tableName) {
		t.Error("Table should not exist after Drop")
	}

	err = db.Schema().DropIfExists(tableName)
	if err != nil {
		t.Fatalf("Failed to drop if exists: %v", err)
	}
}

func TestSQLiteSchemaTableRename(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)

	oldName := "old_table"
	newName := "new_table"

	_ = db.Schema().DropIfExists(oldName)
	_ = db.Schema().DropIfExists(newName)

	err := db.Schema().Create(oldName, func(table schema.Blueprint) {
		table.ID()
	})
	if err != nil {
		t.Fatalf("Failed to create old table: %v", err)
	}

	err = db.Schema().Rename(oldName, newName)
	if err != nil {
		t.Fatalf("Failed to rename table: %v", err)
	}

	if db.Schema().HasTable(oldName) {
		t.Error("Old table should not exist after rename")
	}
	if !db.Schema().HasTable(newName) {
		t.Error("New table should exist after rename")
	}

	_ = db.Schema().Drop(newName)
}

func TestSQLiteSchemaTableGetTables(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)

	table1 := "table1"
	table2 := "table2"

	_ = db.Schema().DropIfExists(table1)
	_ = db.Schema().DropIfExists(table2)

	if err := db.Schema().Create(table1, func(table schema.Blueprint) { table.ID() }); err != nil {
		t.Fatalf("Failed to create table1: %v", err)
	}
	if err := db.Schema().Create(table2, func(table schema.Blueprint) { table.ID() }); err != nil {
		t.Fatalf("Failed to create table2: %v", err)
	}

	tables := db.Schema().GetTableListing()
	table1Found := false
	table2Found := false
	for _, t := range tables {
		if t == table1 {
			table1Found = true
		}
		if t == table2 {
			table2Found = true
		}
	}
	if !table1Found {
		t.Error("table1 should be in table listing")
	}
	if !table2Found {
		t.Error("table2 should be in table listing")
	}

	tableInfos, err := db.Schema().GetTables()
	if err != nil {
		t.Fatalf("Failed to get tables: %v", err)
	}

	found1 := false
	found2 := false
	for _, ti := range tableInfos {
		if ti.Name == table1 {
			found1 = true
		}
		if ti.Name == table2 {
			found2 = true
		}
	}
	if !found1 {
		t.Error("table1 should be in GetTables result")
	}
	if !found2 {
		t.Error("table2 should be in GetTables result")
	}

	_ = db.Schema().Drop(table1)
	_ = db.Schema().Drop(table2)
}

func TestSQLiteSchemaTableModify(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)

	tableName := "modify_table"
	_ = db.Schema().DropIfExists(tableName)

	if err := db.Schema().Create(tableName, func(table schema.Blueprint) {
		table.ID()
	}); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	err := db.Schema().Table(tableName, func(table schema.Blueprint) {
		table.String("new_column").Nullable()
	})
	if err != nil {
		t.Fatalf("Failed to modify table: %v", err)
	}
	if !db.Schema().HasColumn(tableName, "new_column") {
		t.Error("new_column should exist after Table modify")
	}

	_ = db.Schema().Drop(tableName)
}

func TestSQLiteSchemaTableDropAll(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	dsn := "sqlite://:memory:?multi_stmts=true"
	db2, err := neat.NewFromDSN(dsn)
	if err != nil {
		t.Fatalf("Failed to create DB: %v", err)
	}
	defer db2.Close()

	if err := db2.Schema().Create("table_a", func(table schema.Blueprint) { table.ID() }); err != nil {
		t.Fatalf("Failed to create table_a: %v", err)
	}
	if err := db2.Schema().Create("table_b", func(table schema.Blueprint) { table.ID() }); err != nil {
		t.Fatalf("Failed to create table_b: %v", err)
	}

	if !db2.Schema().HasTable("table_a") {
		t.Error("table_a should exist")
	}
	if !db2.Schema().HasTable("table_b") {
		t.Error("table_b should exist")
	}

	err = db2.Schema().DropAllTables()
	if err != nil {
		t.Fatalf("Failed to drop all tables: %v", err)
	}

	if db2.Schema().HasTable("table_a") {
		t.Error("table_a should not exist after DropAllTables")
	}
	if db2.Schema().HasTable("table_b") {
		t.Error("table_b should not exist after DropAllTables")
	}
}

func TestSQLiteSchemaTablePrefix(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	prefix := "pre_"
	config := neat.DBConfig{
		Default: "sqlite",
		Connections: map[string]neat.ConnectionConfig{
			"sqlite": {
				Driver:   "sqlite",
				Database: ":memory:",
				Prefix:   prefix,
			},
		},
	}
	db2, err := neat.New(config)
	if err != nil {
		t.Fatalf("Failed to create DB with prefix: %v", err)
	}
	defer db2.Close()

	tableName := "test"
	err = db2.Schema().Create(tableName, func(table schema.Blueprint) {
		table.ID()
	})
	if err != nil {
		t.Fatalf("Failed to create table with prefix: %v", err)
	}

	if !db2.Schema().HasTable(tableName) {
		t.Error("Table should exist with logical name")
	}

	var actualName string
	sqlDB, err := db2.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}
	err = sqlDB.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", prefix+tableName).Scan(&actualName)
	if err != nil {
		t.Fatalf("Failed to query actual table name: %v", err)
	}
	if actualName != prefix+tableName {
		t.Errorf("Expected %s, got %s", prefix+tableName, actualName)
	}

	tables := db2.Schema().GetTableListing()
	found := false
	for _, t := range tables {
		if t == prefix+tableName {
			found = true
			break
		}
	}
	if !found {
		t.Error("Prefixed table should be in GetTableListing")
	}

	newTableName := "new_test"
	err = db2.Schema().Rename(tableName, newTableName)
	if err != nil {
		t.Fatalf("Failed to rename table with prefix: %v", err)
	}

	if db2.Schema().HasTable(tableName) {
		t.Error("Old table name should not exist after rename")
	}
	if !db2.Schema().HasTable(newTableName) {
		t.Error("New table name should exist after rename")
	}

	sqlDB, err = db2.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}
	err = sqlDB.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", prefix+newTableName).Scan(&actualName)
	if err != nil {
		t.Fatalf("Failed to query actual table name after rename: %v", err)
	}
	if actualName != prefix+newTableName {
		t.Errorf("Expected %s after rename, got %s", prefix+newTableName, actualName)
	}
}
