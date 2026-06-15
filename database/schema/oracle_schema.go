package schema

import (
	"fmt"

	"github.com/dracory/neat/contracts/database/orm"
	contractsschema "github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/database/schema/grammars"
	"github.com/dracory/neat/database/schema/processors"
)

type OracleSchema struct {
	contractsschema.CommonSchema

	grammar   *grammars.Oracle
	orm       orm.Orm
	prefix    string
	processor processors.Oracle
	tx        orm.Query
}

func NewOracleSchema(grammar *grammars.Oracle, orm orm.Orm, prefix string) *OracleSchema {
	return &OracleSchema{
		CommonSchema: NewCommonSchema(grammar, orm),
		grammar:      grammar,
		orm:          orm,
		prefix:       prefix,
		processor:    processors.NewOracle(),
	}
}

func (r *OracleSchema) WithTransaction(tx orm.Query) *OracleSchema {
	return &OracleSchema{
		CommonSchema: NewCommonSchema(r.grammar, r.orm).WithTransaction(tx),
		grammar:      r.grammar,
		orm:          r.orm,
		prefix:       r.prefix,
		processor:    r.processor,
		tx:           tx,
	}
}

func (r *OracleSchema) getQuery() orm.Query {
	if r.tx != nil {
		return r.tx
	}
	return r.orm.Query()
}

func (r *OracleSchema) DropAllTables() error {
	tables, err := r.GetTables()
	if err != nil {
		return err
	}

	if len(tables) == 0 {
		return nil
	}

	query := r.getQuery()
	if query == nil {
		return fmt.Errorf("query not initialized")
	}
	if query.InTransaction() {
		// Already in transaction, use it directly
		var dropTables []string
		for _, table := range tables {
			dropTables = append(dropTables, table.Name)
		}
		if _, execErr := query.Exec(r.grammar.CompileDropAllTables(dropTables)); execErr != nil {
			return execErr
		}

		return nil
	}

	// Not in transaction, wrap in one
	return r.orm.Transaction(func(tx orm.Query) error {
		// Oracle doesn't support foreign key constraint toggling like MySQL
		// Skip the disable/enable foreign key constraints steps

		var dropTables []string
		for _, table := range tables {
			dropTables = append(dropTables, table.Name)
		}
		if _, execErr := tx.Exec(r.grammar.CompileDropAllTables(dropTables)); execErr != nil {
			return execErr
		}

		return nil
	})
}

func (r *OracleSchema) DropAllTypes() error {
	return nil
}

func (r *OracleSchema) DropAllViews() error {
	views, err := r.GetViews()
	if err != nil {
		return err
	}
	if len(views) == 0 {
		return nil
	}

	var dropViews []string
	for _, view := range views {
		dropViews = append(dropViews, view.Name)
	}

	query := r.getQuery()
	if query == nil {
		return fmt.Errorf("query not initialized")
	}
	_, err = query.Exec(r.grammar.CompileDropAllViews(dropViews))

	return err
}

func (r *OracleSchema) GetColumns(table string) ([]contractsschema.Column, error) {
	table = r.prefix + table

	var dbColumns []contractsschema.DBColumn
	query := r.getQuery()
	if query == nil {
		return nil, fmt.Errorf("query not initialized")
	}
	// Use user_tab_columns which automatically uses current user's schema
	if err := query.Raw(r.grammar.CompileColumns("", table)).Scan(&dbColumns); err != nil {
		return nil, err
	}

	return r.processor.ProcessColumns(dbColumns), nil
}

func (r *OracleSchema) GetIndexes(table string) ([]contractsschema.Index, error) {
	table = r.prefix + table

	var dbIndexes []contractsschema.DBIndex
	query := r.getQuery()
	if query == nil {
		return nil, fmt.Errorf("query not initialized")
	}
	if err := query.Raw(r.grammar.CompileIndexes(r.orm.DatabaseName(), table)).Scan(&dbIndexes); err != nil {
		return nil, err
	}

	return r.processor.ProcessIndexes(dbIndexes), nil
}

func (r *OracleSchema) GetTypes() ([]contractsschema.Type, error) {
	return nil, nil
}
