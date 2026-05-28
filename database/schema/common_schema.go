package schema

import (
	"fmt"

	"github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/contracts/database/schema"
)

type CommonSchema struct {
	grammar schema.Grammar
	orm     orm.Orm
}

func NewCommonSchema(grammar schema.Grammar, orm orm.Orm) *CommonSchema {
	return &CommonSchema{
		grammar: grammar,
		orm:     orm,
	}
}

func (r *CommonSchema) GetTables() ([]schema.Table, error) {
	var tables []schema.Table
	query := r.orm.Query()
	if query == nil {
		return nil, fmt.Errorf("query not initialized")
	}
	if err := query.Raw(r.grammar.CompileTables(r.orm.DatabaseName())).Scan(&tables); err != nil {
		return nil, err
	}

	return tables, nil
}

func (r *CommonSchema) GetViews() ([]schema.View, error) {
	var views []schema.View
	query := r.orm.Query()
	if query == nil {
		return nil, fmt.Errorf("query not initialized")
	}
	if err := query.Raw(r.grammar.CompileViews(r.orm.DatabaseName())).Scan(&views); err != nil {
		return nil, err
	}

	return views, nil
}
