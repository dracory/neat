package schema

import (
	"fmt"
	"slices"

	"github.com/dracory/neat/contracts/database/orm"
	contractsschema "github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/database/schema/grammars"
	"github.com/dracory/neat/database/schema/processors"
)

type PostgresSchema struct {
	contractsschema.CommonSchema

	grammar   *grammars.Postgres
	orm       orm.Orm
	prefix    string
	processor processors.Postgres
	schema    string
	tx        orm.Query
}

func NewPostgresSchema(grammar *grammars.Postgres, orm orm.Orm, schema, prefix string) *PostgresSchema {
	return &PostgresSchema{
		CommonSchema: NewCommonSchema(grammar, orm),
		grammar:      grammar,
		orm:          orm,
		prefix:       prefix,
		processor:    processors.NewPostgres(),
		schema:       schema,
	}
}

func (r *PostgresSchema) WithTransaction(tx orm.Query) *PostgresSchema {
	return &PostgresSchema{
		CommonSchema: NewCommonSchema(r.grammar, r.orm).WithTransaction(tx),
		grammar:      r.grammar,
		orm:          r.orm,
		prefix:       r.prefix,
		processor:    r.processor,
		schema:       r.schema,
		tx:           tx,
	}
}

func (r *PostgresSchema) getQuery() orm.Query {
	if r.tx != nil {
		return r.tx
	}
	return r.orm.Query()
}

func (r *PostgresSchema) DropAllTables() error {
	excludedTables, err := r.grammar.EscapeNames([]string{"spatial_ref_sys"})
	if err != nil {
		return err
	}
	schemaEscaped, err := r.grammar.EscapeNames([]string{r.schema})
	if err != nil {
		return err
	}
	schema := schemaEscaped[0]

	tables, err := r.GetTables()
	if err != nil {
		return err
	}

	var dropTables []string
	for _, table := range tables {
		qualifiedName := fmt.Sprintf("%s.%s", table.Schema, table.Name)

		isExcludedTable := slices.Contains(excludedTables, qualifiedName) || slices.Contains(excludedTables, table.Name)
		tableSchemaEscaped, err := r.grammar.EscapeNames([]string{table.Schema})
		if err != nil {
			return err
		}
		isInCurrentSchema := schema == tableSchemaEscaped[0]

		if !isExcludedTable && isInCurrentSchema {
			dropTables = append(dropTables, qualifiedName)
		}
	}

	if len(dropTables) == 0 {
		return nil
	}

	query := r.getQuery()
	if query == nil {
		return fmt.Errorf("query not initialized")
	}

	// PostgreSQL DDL statements are transactional, so we can use the transaction if available
	_, err = query.Exec(r.grammar.CompileDropAllTables(dropTables))

	return err
}

func (r *PostgresSchema) DropAllTypes() error {
	types, err := r.GetTypes()
	if err != nil {
		return err
	}

	var dropTypes, dropDomains []string

	for _, t := range types {
		if !t.Implicit && r.schema == t.Schema {
			if t.Type == "domain" {
				dropDomains = append(dropDomains, fmt.Sprintf("%s.%s", t.Schema, t.Name))
			} else {
				dropTypes = append(dropTypes, fmt.Sprintf("%s.%s", t.Schema, t.Name))
			}
		}
	}

	query := r.getQuery()
	if query.InTransaction() {
		// Already in transaction, use it directly
		if len(dropTypes) > 0 {
			if _, err := query.Exec(r.grammar.CompileDropAllTypes(dropTypes)); err != nil {
				return err
			}
		}

		if len(dropDomains) > 0 {
			if _, err := query.Exec(r.grammar.CompileDropAllDomains(dropDomains)); err != nil {
				return err
			}
		}

		return nil
	}

	// Not in transaction, wrap in one
	return r.orm.Transaction(func(tx orm.Query) error {
		if len(dropTypes) > 0 {
			if _, err := tx.Exec(r.grammar.CompileDropAllTypes(dropTypes)); err != nil {
				return err
			}
		}

		if len(dropDomains) > 0 {
			if _, err := tx.Exec(r.grammar.CompileDropAllDomains(dropDomains)); err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *PostgresSchema) DropAllViews() error {
	views, err := r.GetViews()
	if err != nil {
		return err
	}

	var dropViews []string
	for _, view := range views {
		if r.schema == view.Schema {
			dropViews = append(dropViews, fmt.Sprintf("%s.%s", view.Schema, view.Name))
		}
	}
	if len(dropViews) == 0 {
		return nil
	}

	query := r.getQuery()
	if query == nil {
		return fmt.Errorf("query not initialized")
	}
	_, err = query.Exec(r.grammar.CompileDropAllViews(dropViews))

	return err
}

func (r *PostgresSchema) GetColumns(table string) ([]contractsschema.Column, error) {
	schema, table, err := parseSchemaAndTable(table, r.schema)
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

func (r *PostgresSchema) GetIndexes(table string) ([]contractsschema.Index, error) {
	schema, table, err := parseSchemaAndTable(table, r.schema)
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

func (r *PostgresSchema) GetTypes() ([]contractsschema.Type, error) {
	var types []contractsschema.Type
	query := r.getQuery()
	if query == nil {
		return nil, fmt.Errorf("query not initialized")
	}
	if err := query.Raw(r.grammar.CompileTypes()).Scan(&types); err != nil {
		return nil, err
	}

	return r.processor.ProcessTypes(types), nil
}
