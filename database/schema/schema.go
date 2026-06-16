package schema

import (
	"fmt"
	"slices"
	"strings"

	"github.com/dracory/neat/contracts/config"
	contractsdatabase "github.com/dracory/neat/contracts/database"
	contractsorm "github.com/dracory/neat/contracts/database/orm"
	contractsschema "github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/contracts/log"
	"github.com/dracory/neat/database/schema/grammars"
	"github.com/dracory/neat/database/schema/processors"
	"github.com/dracory/neat/errors"
)

var _ contractsschema.Schema = (*Schema)(nil)
var _ contractsschema.MigrationInterface = (*BaseMigration)(nil)

type Schema struct {
	contractsschema.CommonSchema
	contractsschema.DriverSchema

	config     config.Config
	grammar    contractsschema.Grammar
	log        log.Log
	migrations []contractsschema.MigrationInterface
	orm        contractsorm.Orm
	prefix     string
	processor  contractsschema.Processor
	schema     string
	tx         contractsorm.Query
}

func NewSchema(config config.Config, log log.Log, orm contractsorm.Orm) (*Schema, error) {
	driver := contractsdatabase.Driver(config.GetString(fmt.Sprintf("database.connections.%s.driver", orm.Name())))
	prefix := config.GetString(fmt.Sprintf("database.connections.%s.prefix", orm.Name()))
	var (
		driverSchema contractsschema.DriverSchema
		grammar      contractsschema.Grammar
		processor    contractsschema.Processor
		schema       string
	)

	switch driver {
	case contractsdatabase.DriverPostgres:
		schema = config.GetString(fmt.Sprintf("database.connections.%s.schema", orm.Name()), "public")

		postgresGrammar := grammars.NewPostgres(prefix)
		driverSchema = NewPostgresSchema(postgresGrammar, orm, schema, prefix)
		grammar = postgresGrammar
		processor = processors.NewPostgres()
	case contractsdatabase.DriverMysql:
		schema = config.GetString(fmt.Sprintf("database.connections.%s.database", orm.Name()))

		mysqlGrammar := grammars.NewMysql(prefix)
		driverSchema = NewMysqlSchema(mysqlGrammar, orm, prefix)
		grammar = mysqlGrammar
		processor = processors.NewMysql()
	case contractsdatabase.DriverSqlserver:
		sqlserverGrammar := grammars.NewSqlserver(prefix)
		driverSchema = NewSqlserverSchema(sqlserverGrammar, orm, prefix)
		grammar = sqlserverGrammar
		processor = processors.NewSqlserver()
	case contractsdatabase.DriverSqlite:
		sqliteGrammar := grammars.NewSqlite(log, prefix)
		driverSchema = NewSqliteSchema(sqliteGrammar, orm, prefix)
		grammar = sqliteGrammar
		processor = processors.NewSqlite()
	case contractsdatabase.DriverTurso:
		sqliteGrammar := grammars.NewSqlite(log, prefix)
		driverSchema = NewSqliteSchema(sqliteGrammar, orm, prefix)
		grammar = sqliteGrammar
		processor = processors.NewSqlite()
	case contractsdatabase.DriverOracle:
		oracleGrammar := grammars.NewOracle(prefix)
		driverSchema = NewOracleSchema(oracleGrammar, orm, prefix)
		grammar = oracleGrammar
		processor = processors.NewOracle()
	default:
		return nil, errors.SchemaDriverNotSupported.Args(driver)
	}

	return &Schema{
		DriverSchema: driverSchema,
		CommonSchema: NewCommonSchema(grammar, orm),

		config:     config,
		grammar:    grammar,
		log:    log,
		orm:    orm,
		prefix:     prefix,
		processor:  processor,
		schema:     schema,
	}, nil
}

func (r *Schema) GetTables() ([]contractsschema.Table, error) {
	var tables []contractsschema.Table
	query := r.getQuery()
	if query == nil {
		return nil, fmt.Errorf("query not initialized")
	}
	if err := query.Raw(r.grammar.CompileTables(r.orm.DatabaseName())).Scan(&tables); err != nil {
		return nil, err
	}

	if r.processor != nil {
		tables = r.processor.ProcessTables(tables)
	}

	return tables, nil
}

func (r *Schema) Connection(name string) contractsschema.Schema {
	s, err := NewSchema(r.config, r.log, r.orm.Connection(name))
	if err != nil {
		r.log.Errorf("failed to create schema for connection %s: %v", name, err)
		return nil
	}
	return s
}

func (r *Schema) Create(table string, callback func(table contractsschema.Blueprint)) error {
	blueprint := r.createBlueprint(table)
	blueprint.Create()
	callback(blueprint)

	if err := r.build(blueprint); err != nil {
		return errors.SchemaFailedToCreateTable.Args(table, err)
	}

	return nil
}

func (r *Schema) Drop(table string) error {
	blueprint := r.createBlueprint(table)
	blueprint.Drop()

	if err := r.build(blueprint); err != nil {
		return errors.SchemaFailedToDropTable.Args(table, err)
	}

	return nil
}

func (r *Schema) DropColumns(table string, columns []string) error {
	blueprint := r.createBlueprint(table)
	blueprint.DropColumn(columns...)

	if err := r.build(blueprint); err != nil {
		return errors.SchemaFailedToDropColumns.Args(table, err)
	}

	return nil
}

func (r *Schema) DropIfExists(table string) error {
	blueprint := r.createBlueprint(table)
	blueprint.DropIfExists()

	if err := r.build(blueprint); err != nil {
		return errors.SchemaFailedToDropTable.Args(table, err)
	}

	return nil
}

func (r *Schema) GetColumnListing(table string) []string {
	columns, err := r.GetColumns(table)
	if err != nil {
		r.log.Errorf("failed to get %s columns: %v", table, err)
		return nil
	}

	var names []string
	for _, column := range columns {
		names = append(names, column.Name)
	}

	return names
}

func (r *Schema) GetConnection() string {
	return r.orm.Name()
}

func (r *Schema) GetIndexes(table string) ([]contractsschema.Index, error) {
	indexes, err := r.DriverSchema.GetIndexes(table)
	if err != nil {
		return nil, err
	}

	for i := range indexes {
		indexes[i].Name = strings.ToLower(indexes[i].Name)
	}

	return indexes, nil
}

func (r *Schema) GetForeignKeys(table string) ([]contractsschema.ForeignKey, error) {
	table = r.prefix + table

	var dbForeignKeys []contractsschema.DBForeignKey
	query := r.getQuery()
	if query == nil {
		return nil, fmt.Errorf("query not initialized")
	}
	if err := query.Raw(r.grammar.CompileForeignKeys(r.schema, table)).Scan(&dbForeignKeys); err != nil {
		return nil, err
	}

	return r.processor.ProcessForeignKeys(dbForeignKeys), nil
}

func (r *Schema) GetIndexListing(table string) []string {
	indexes, err := r.GetIndexes(table)
	if err != nil {
		r.log.Errorf("failed to get %s indexes: %v", table, err)
		return nil
	}

	var names []string
	for _, index := range indexes {
		names = append(names, index.Name)
	}

	return names
}

func (r *Schema) GetTableListing() []string {
	tables, err := r.GetTables()
	if err != nil {
		r.log.Errorf("failed to get tables: %v", err)
		return nil
	}

	var names []string
	for _, table := range tables {
		names = append(names, table.Name)
	}

	return names
}

func (r *Schema) HasColumn(table, column string) bool {
	return slices.Contains(r.GetColumnListing(table), column)
}

func (r *Schema) HasColumns(table string, columns []string) bool {
	columnListing := r.GetColumnListing(table)
	for _, column := range columns {
		if !slices.Contains(columnListing, column) {
			return false
		}
	}

	return true
}

func (r *Schema) HasIndex(table, index string) bool {
	indexListing := r.GetIndexListing(table)

	return slices.Contains(indexListing, index)
}

func (r *Schema) HasTable(name string) bool {
	var schema string
	if strings.Contains(name, ".") {
		lastDotIndex := strings.LastIndex(name, ".")
		schema = name[:lastDotIndex]
		name = name[lastDotIndex+1:]
	}

	tableName := r.prefix + name

	tables, err := r.GetTables()
	if err != nil {
		r.log.Errorf(errors.SchemaFailedToGetTables.Args(r.orm.Name(), err).Error())
		return false
	}

	for _, table := range tables {
		if table.Name == tableName {
			if schema == "" || schema == table.Schema {
				return true
			}
		}
	}

	return false
}

func (r *Schema) HasType(name string) bool {
	types, err := r.GetTypes()
	if err != nil {
		r.log.Errorf(errors.SchemaFailedToGetTables.Args(r.orm.Name(), err).Error())
		return false
	}

	for _, t := range types {
		if t.Name == name {
			return true
		}
	}

	return false
}

func (r *Schema) HasView(name string) bool {
	views, err := r.GetViews()
	if err != nil {
		r.log.Errorf(errors.SchemaFailedToGetTables.Args(r.orm.Name(), err).Error())
		return false
	}

	for _, view := range views {
		if view.Name == name {
			return true
		}
	}

	return false
}

func (r *Schema) Orm() contractsorm.Orm {
	return r.orm
}

func (r *Schema) Rename(from, to string) error {
	blueprint := r.createBlueprint(from)
	blueprint.Rename(to)

	if err := r.build(blueprint); err != nil {
		return errors.SchemaFailedToRenameTable.Args(from, err)
	}

	return nil
}

func (r *Schema) RenameColumn(table, from, to string) error {
	blueprint := r.createBlueprint(table)
	blueprint.RenameColumn(from, to)

	if err := r.build(blueprint); err != nil {
		return errors.SchemaFailedToChangeTable.Args(table, err)
	}

	return nil
}

func (r *Schema) SetConnection(name string) {
	r.orm = r.orm.Connection(name)
	r.tx = nil
}

func (r *Schema) Sql(sql string) error {
	query := r.getQuery()
	if query == nil {
		return fmt.Errorf("query not initialized")
	}
	_, err := query.Exec(sql)

	return err
}

func (r *Schema) Table(table string, callback func(table contractsschema.Blueprint)) error {
	blueprint := r.createBlueprint(table)
	callback(blueprint)

	if err := r.build(blueprint); err != nil {
		return errors.SchemaFailedToChangeTable.Args(table, err)
	}

	return nil
}

func (r *Schema) WithTransaction(tx contractsorm.Query) contractsschema.Schema {
	// Create new CommonSchema with tx
	newCommon := NewCommonSchema(r.grammar, r.orm).WithTransaction(tx)

	// Create new driver schema with tx
	var newDriver contractsschema.DriverSchema
	switch d := r.DriverSchema.(type) {
	case *PostgresSchema:
		newDriver = d.WithTransaction(tx)
	case *MysqlSchema:
		newDriver = d.WithTransaction(tx)
	case *SqlserverSchema:
		newDriver = d.WithTransaction(tx)
	case *SqliteSchema:
		newDriver = d.WithTransaction(tx)
	case *OracleSchema:
		newDriver = d.WithTransaction(tx)
	default:
		// Unknown driver type - keep original driver schema
		// This shouldn't happen in practice as all supported drivers are covered
		newDriver = r.DriverSchema
	}

	s := &Schema{
		CommonSchema: newCommon,
		DriverSchema: newDriver,
		config:       r.config,
		grammar:      r.grammar,
		log:          r.log,
		orm:          r.orm,
		prefix:       r.prefix,
		processor:    r.processor,
		schema:       r.schema,
		tx:           tx,
	}

	return s
}

func (r *Schema) build(blueprint contractsschema.Blueprint) error {
	query := r.getQuery()
	if query == nil {
		return fmt.Errorf("query not initialized")
	}

	// Skip transaction wrapper for RenameIndex on SQLite/Turso to avoid savepoint timeout
	if bp, ok := blueprint.(*Blueprint); ok && bp.ShouldSkipTransaction() {
		if _, isSqlite := r.grammar.(*grammars.Sqlite); isSqlite {
			return blueprint.Build(query, r.grammar)
		}
	}

	if query.InTransaction() {
		return blueprint.Build(query, r.grammar)
	}

	return r.orm.Transaction(func(tx contractsorm.Query) error {
		return blueprint.Build(tx, r.grammar)
	})
}

func (r *Schema) getQuery() contractsorm.Query {
	if r.tx != nil {
		return r.tx
	}
	return r.orm.Query()
}

func (r *Schema) createBlueprint(table string) contractsschema.Blueprint {
	return NewBlueprint(r, r.prefix, table)
}
