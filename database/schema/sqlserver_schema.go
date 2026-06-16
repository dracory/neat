package schema

import (
	"fmt"

	"github.com/dracory/neat/contracts/database/orm"
	contractsschema "github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/database/schema/grammars"
	"github.com/dracory/neat/database/schema/processors"
)

type SqlserverSchema struct {
	contractsschema.CommonSchema

	grammar   *grammars.Sqlserver
	orm       orm.Orm
	prefix    string
	processor processors.Sqlserver
	tx        orm.Query
}

func NewSqlserverSchema(grammar *grammars.Sqlserver, orm orm.Orm, prefix string) *SqlserverSchema {
	return &SqlserverSchema{
		CommonSchema: NewCommonSchema(grammar, orm),
		grammar:      grammar,
		orm:          orm,
		prefix:       prefix,
		processor:    processors.NewSqlserver(),
	}
}

func (r *SqlserverSchema) WithTransaction(tx orm.Query) *SqlserverSchema {
	return &SqlserverSchema{
		CommonSchema: NewCommonSchema(r.grammar, r.orm).WithTransaction(tx),
		grammar:      r.grammar,
		orm:          r.orm,
		prefix:       r.prefix,
		processor:    r.processor,
		tx:           tx,
	}
}

func (r *SqlserverSchema) getQuery() orm.Query {
	if r.tx != nil {
		return r.tx
	}
	return r.orm.Query()
}

func (r *SqlserverSchema) DropAllTables() error {
	query := r.getQuery()
	if query == nil {
		return fmt.Errorf("query not initialized")
	}
	// Uses the transaction-aware query if available
	if _, err := query.Exec(r.grammar.CompileDropAllForeignKeys()); err != nil {
		return err
	}

	if _, err := query.Exec(r.grammar.CompileDropAllTables(nil)); err != nil {
		return err
	}

	return nil
}

func (r *SqlserverSchema) DropAllTypes() error {
	return nil
}

func (r *SqlserverSchema) DropAllViews() error {
	query := r.getQuery()
	if query == nil {
		return fmt.Errorf("query not initialized")
	}
	_, err := query.Exec(r.grammar.CompileDropAllViews(nil))

	return err
}

func (r *SqlserverSchema) GetColumns(table string) ([]contractsschema.Column, error) {
	schema, table, err := parseSchemaAndTable(table, "")
	if err != nil {
		return nil, err
	}

	table = r.prefix + table

	var dbColumns []contractsschema.DBColumn
	query := r.getQuery()
	if query == nil {
		return nil, fmt.Errorf("query not initialized")
	}
	if err := query.Raw(r.grammar.CompileColumns(schema, table)).Scan(&dbColumns); err != nil {
		return nil, err
	}

	return r.processor.ProcessColumns(dbColumns), nil
}

func (r *SqlserverSchema) GetIndexes(table string) ([]contractsschema.Index, error) {
	schema, table, err := parseSchemaAndTable(table, "")
	if err != nil {
		return nil, err
	}

	table = r.prefix + table

	var dbIndexes []contractsschema.DBIndex
	query := r.getQuery()
	if query == nil {
		return nil, fmt.Errorf("query not initialized")
	}
	if err := query.Raw(r.grammar.CompileIndexes(schema, table)).Scan(&dbIndexes); err != nil {
		return nil, err
	}

	return r.processor.ProcessIndexes(dbIndexes), nil
}

func (r *SqlserverSchema) GetTypes() ([]contractsschema.Type, error) {
	return nil, nil
}
