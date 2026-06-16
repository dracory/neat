package schema

import (
	"fmt"

	"github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/database/schema/grammars"
	"github.com/dracory/neat/database/schema/processors"
)

type SqliteSchema struct {
	schema.CommonSchema

	grammar   *grammars.Sqlite
	orm       orm.Orm
	prefix    string
	processor processors.Sqlite
	tx        orm.Query
}

func NewSqliteSchema(grammar *grammars.Sqlite, orm orm.Orm, prefix string) *SqliteSchema {
	return &SqliteSchema{
		CommonSchema: NewCommonSchema(grammar, orm),
		grammar:      grammar,
		orm:          orm,
		prefix:       prefix,
		processor:    processors.NewSqlite(),
	}
}

func (r *SqliteSchema) WithTransaction(tx orm.Query) *SqliteSchema {
	return &SqliteSchema{
		CommonSchema: NewCommonSchema(r.grammar, r.orm).WithTransaction(tx),
		grammar:      r.grammar,
		orm:          r.orm,
		prefix:       r.prefix,
		processor:    r.processor,
		tx:           tx,
	}
}

func (r *SqliteSchema) getQuery() orm.Query {
	if r.tx != nil {
		return r.tx
	}
	return r.orm.Query()
}

func (r *SqliteSchema) DropAllTables() error {
	// PRAGMA writable_schema and VACUUM cannot run inside a transaction.
	// Always use the base query, bypassing any WithTransaction wrapper.
	query := r.orm.Query()
	if query == nil {
		return fmt.Errorf("query not initialized")
	}
	if _, err := query.Exec(r.grammar.CompileEnableWriteableSchema()); err != nil {
		return err
	}
	if _, err := query.Exec(r.grammar.CompileDropAllTables(nil)); err != nil {
		return err
	}
	if _, err := query.Exec(r.grammar.CompileDisableWriteableSchema()); err != nil {
		return err
	}
	if _, err := query.Exec(r.grammar.CompileRebuild()); err != nil {
		return err
	}

	return nil
}

func (r *SqliteSchema) DropAllTypes() error {
	return nil
}

func (r *SqliteSchema) DropAllViews() error {
	// Use transaction-aware query for the drop operations if available
	query := r.getQuery()
	if query == nil {
		return fmt.Errorf("query not initialized")
	}
	if _, err := query.Exec(r.grammar.CompileEnableWriteableSchema()); err != nil {
		return err
	}
	if _, err := query.Exec(r.grammar.CompileDropAllViews(nil)); err != nil {
		return err
	}
	if _, err := query.Exec(r.grammar.CompileDisableWriteableSchema()); err != nil {
		return err
	}

	// cannot VACUUM from within a transaction
	nonTxQuery := r.orm.Query()
	if nonTxQuery == nil {
		return fmt.Errorf("query not initialized")
	}
	if _, err := nonTxQuery.Exec(r.grammar.CompileRebuild()); err != nil {
		return err
	}

	return nil
}

func (r *SqliteSchema) GetColumns(table string) ([]schema.Column, error) {
	table = r.prefix + table

	var dbColumns []schema.DBColumn
	query := r.getQuery()
	if query == nil {
		return nil, fmt.Errorf("query not initialized")
	}
	if err := query.Raw(r.grammar.CompileColumns("", table)).Scan(&dbColumns); err != nil {
		return nil, err
	}

	return r.processor.ProcessColumns(dbColumns), nil
}

func (r *SqliteSchema) GetIndexes(table string) ([]schema.Index, error) {
	table = r.prefix + table

	var dbIndexes []schema.DBIndex
	query := r.getQuery()
	if query == nil {
		return nil, fmt.Errorf("query not initialized")
	}
	if err := query.Raw(r.grammar.CompileIndexes("", table)).Scan(&dbIndexes); err != nil {
		return nil, err
	}

	return r.processor.ProcessIndexes(dbIndexes), nil
}

func (r *SqliteSchema) GetTypes() ([]schema.Type, error) {
	return nil, nil
}
