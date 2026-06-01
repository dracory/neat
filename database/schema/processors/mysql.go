package processors

import (
	"strings"

	"github.com/spf13/cast"

	"github.com/dracory/neat/contracts/database/schema"
)

type Mysql struct {
}

func NewMysql() Mysql {
	return Mysql{}
}

func (r Mysql) ProcessColumns(dbColumns []schema.DBColumn) []schema.Column {
	columns := make([]schema.Column, 0)
	for _, dbColumn := range dbColumns {
		var nullable bool
		if dbColumn.Nullable == "YES" {
			nullable = true
		}
		var autoIncrement bool
		if dbColumn.Extra == "auto_increment" {
			autoIncrement = true
		}

		// Handle NULL collation and comment - MySQL returns NULL for non-text columns
		var collation, comment string
		if dbColumn.Collation != nil {
			collation = *dbColumn.Collation
		}
		if dbColumn.Comment != nil {
			comment = *dbColumn.Comment
		}

		// If Default is a pointer, dereference it safely
		defaultStr := cast.ToString(dbColumn.Default)
		columns = append(columns, schema.Column{
			Autoincrement: autoIncrement,
			Collation:     collation,
			Comment:       comment,
			Default:       defaultStr,
			Name:          dbColumn.Name,
			Nullable:      nullable,
			Type:          dbColumn.Type,
			TypeName:      dbColumn.TypeName,
		})
	}

	return columns
}

func (r Mysql) ProcessForeignKeys(dbForeignKeys []schema.DBForeignKey) []schema.ForeignKey {
	foreignKeys := make([]schema.ForeignKey, 0)
	for _, dbForeignKey := range dbForeignKeys {
		foreignKeys = append(foreignKeys, schema.ForeignKey{
			Name:           dbForeignKey.Name,
			Columns:        strings.Split(dbForeignKey.Columns, ","),
			ForeignSchema:  dbForeignKey.ForeignSchema,
			ForeignTable:   dbForeignKey.ForeignTable,
			ForeignColumns: strings.Split(dbForeignKey.ForeignColumns, ","),
			OnUpdate:       strings.ToLower(dbForeignKey.OnUpdate),
			OnDelete:       strings.ToLower(dbForeignKey.OnDelete),
		})
	}

	return foreignKeys
}

func (r Mysql) ProcessIndexes(dbIndexes []schema.DBIndex) []schema.Index {
	indexes := make([]schema.Index, 0)
	for _, dbIndex := range dbIndexes {
		name := strings.ToLower(dbIndex.Name)
		indexes = append(indexes, schema.Index{
			Columns: strings.Split(dbIndex.Columns, ","),
			Name:    name,
			Type:    strings.ToLower(dbIndex.Type),
			Primary: name == "primary",
			Unique:  dbIndex.Unique,
		})
	}

	return indexes
}

func (r Mysql) ProcessTables(dbTables []schema.Table) []schema.Table {
	return dbTables
}
