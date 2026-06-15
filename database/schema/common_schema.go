package schema

import (
	"fmt"

	"github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/contracts/database/schema"
)

type CommonSchema struct {
	grammar   schema.Grammar
	orm       orm.Orm
	processor schema.Processor
	tx        orm.Query
}

func NewCommonSchema(grammar schema.Grammar, orm orm.Orm) *CommonSchema {
	return &CommonSchema{
		grammar: grammar,
		orm:     orm,
	}
}

func (r *CommonSchema) WithTransaction(tx orm.Query) *CommonSchema {
	return &CommonSchema{
		grammar:   r.grammar,
		orm:       r.orm,
		processor: r.processor,
		tx:        tx,
	}
}

func (r *CommonSchema) getQuery() orm.Query {
	if r.tx != nil {
		return r.tx
	}
	return r.orm.Query()
}

func (r *CommonSchema) GetTables() ([]schema.Table, error) {
	var tables []schema.Table
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

func (r *CommonSchema) GetViews() ([]schema.View, error) {
	var views []schema.View
	query := r.getQuery()
	if query == nil {
		return nil, fmt.Errorf("query not initialized")
	}
	if err := query.Raw(r.grammar.CompileViews(r.orm.DatabaseName())).Scan(&views); err != nil {
		return nil, err
	}

	return views, nil
}
