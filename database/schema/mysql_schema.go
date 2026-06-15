package schema

import (
	"fmt"

	"github.com/dracory/neat/contracts/database/orm"
	contractsschema "github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/database/schema/grammars"
	"github.com/dracory/neat/database/schema/processors"
)

type MysqlSchema struct {
	contractsschema.CommonSchema

	grammar   *grammars.Mysql
	orm       orm.Orm
	prefix    string
	processor processors.Mysql
	tx        orm.Query
}

func NewMysqlSchema(grammar *grammars.Mysql, orm orm.Orm, prefix string) *MysqlSchema {
	return &MysqlSchema{
		CommonSchema: NewCommonSchema(grammar, orm),
		grammar:      grammar,
		orm:          orm,
		prefix:       prefix,
		processor:    processors.NewMysql(),
	}
}

func (r *MysqlSchema) WithTransaction(tx orm.Query) *MysqlSchema {
	return &MysqlSchema{
		CommonSchema: NewCommonSchema(r.grammar, r.orm).WithTransaction(tx),
		grammar:      r.grammar,
		orm:          r.orm,
		prefix:       r.prefix,
		processor:    r.processor,
		tx:           tx,
	}
}

func (r *MysqlSchema) getQuery() orm.Query {
	if r.tx != nil {
		return r.tx
	}
	return r.orm.Query()
}

func (r *MysqlSchema) DropAllTables() error {
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
		if _, err = query.Exec(r.grammar.CompileDisableForeignKeyConstraints()); err != nil {
			return err
		}

		var dropTables []string
		for _, table := range tables {
			dropTables = append(dropTables, table.Name)
		}
		if _, err = query.Exec(r.grammar.CompileDropAllTables(dropTables)); err != nil {
			return err
		}

		if _, err = query.Exec(r.grammar.CompileEnableForeignKeyConstraints()); err != nil {
			return err
		}

		return err
	}

	// Not in transaction, wrap in one
	return r.orm.Transaction(func(tx orm.Query) error {
		if _, err = tx.Exec(r.grammar.CompileDisableForeignKeyConstraints()); err != nil {
			return err
		}

		var dropTables []string
		for _, table := range tables {
			dropTables = append(dropTables, table.Name)
		}
		if _, err = tx.Exec(r.grammar.CompileDropAllTables(dropTables)); err != nil {
			return err
		}

		if _, err = tx.Exec(r.grammar.CompileEnableForeignKeyConstraints()); err != nil {
			return err
		}

		return err
	})
}

func (r *MysqlSchema) DropAllTypes() error {
	return nil
}

func (r *MysqlSchema) DropAllViews() error {
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

func (r *MysqlSchema) GetColumns(table string) ([]contractsschema.Column, error) {
	table = r.prefix + table

	var dbColumns []contractsschema.DBColumn
	query := r.getQuery()
	if query == nil {
		return nil, fmt.Errorf("query not initialized")
	}
	if err := query.Raw(r.grammar.CompileColumns(r.orm.DatabaseName(), table)).Scan(&dbColumns); err != nil {
		return nil, err
	}

	return r.processor.ProcessColumns(dbColumns), nil
}

func (r *MysqlSchema) GetIndexes(table string) ([]contractsschema.Index, error) {
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

func (r *MysqlSchema) GetTypes() ([]contractsschema.Type, error) {
	return nil, nil
}
