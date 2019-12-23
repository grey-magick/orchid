package repository

import (
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/isutton/orchid/pkg/orchid/orm"
)

// Assembler is the component to build back unstructured objects from an orm.ResultSet.
type Assembler struct {
	logger    logr.Logger    // logger instance
	schema    *orm.Schema    // ORM schema
	resultSet *orm.ResultSet // data result-set
}

// keyValue create object based on result-set, but assume different names for columns and structure
// of data as key-value only.
func (a *Assembler) keyValue(
	relatedTableName string,
	tableName string,
	pk interface{},
) (map[string]interface{}, error) {
	a.logger.WithValues("related", relatedTableName, "table", tableName).
		Info("Retrieving data to assemble key-value")
	nestedEntries, err := a.resultSet.Get(relatedTableName, tableName, pk)
	if err != nil {
		return nil, err
	}

	item := map[string]interface{}{}
	for _, nestedEntry := range nestedEntries {
		key, found := nestedEntry["key"]
		if !found {
			continue
		}
		value, found := nestedEntry["value"]
		if !found {
			continue
		}
		item[key.(string)] = value
	}
	return item, nil
}

// slice return result-set data as slice of interface.
func (a *Assembler) slice(
	relatedTableName string,
	tableName string,
	pk interface{},
	columns []string,
) ([]interface{}, error) {
	a.logger.WithValues("related", relatedTableName, "table", tableName, "columns", columns).
		Info("Retrieving data to assemble slice")
	relatedEntries, err := a.resultSet.Get(relatedTableName, tableName, pk)
	if err != nil {
		return nil, err
	}

	strippedEntries := []interface{}{}
	for _, relatedEntry := range relatedEntries {
		strippedEntry := a.resultSet.Strip(relatedEntry, columns)
		strippedEntries = append(strippedEntries, strippedEntry)
	}
	return strippedEntries, nil
}

// related goes after one-to-many relationships, by reaching for the results of other tables that
// are having foregin-keys pointing back to another.
func (a *Assembler) related(
	relatedTableName string,
	tableName string,
	pk interface{},
) (map[string]interface{}, error) {
	table, err := a.schema.GetTable(relatedTableName)
	if err != nil {
		return nil, err
	}

	a.logger.WithValues("related", relatedTableName, "table", tableName).
		Info("Retrieving data to assemble related object")

	// using path to decide column's name
	tablePathLen := len(table.Path)
	if tablePathLen == 0 {
		return nil, fmt.Errorf("unabel to determine column name based on table's path")
	}
	columnName := table.Path[tablePathLen-1]

	entry := map[string]interface{}{}
	if table.KV {
		entry[columnName], err = a.keyValue(relatedTableName, tableName, pk)
		if err != nil {
			return nil, err
		}
	} else {
		columns := table.ColumnNamesStripped()
		entry[columnName], err = a.slice(relatedTableName, tableName, pk, columns)
		if err != nil {
			return nil, err
		}
	}
	return entry, nil
}

// object based on table name and primary-key value, recursively create new objects when finding
// one-to-one relationships, and enriching the current object with one-to-many related entries.
func (a *Assembler) object(tableName string, pk interface{}) (map[string]interface{}, error) {
	table, err := a.schema.GetTable(tableName)
	if err != nil {
		return nil, err
	}

	a.logger.WithValues("table", tableName).Info("Retrieving data to assemble new object")
	entry, err := a.resultSet.GetPK(tableName, pk)
	if err != nil {
		return nil, err
	}

	// one-to-ene: when this table is refering another via foreign-keys
	for columnName, columnValue := range entry {
		if table.IsForeignKey(columnName) {
			entry[columnName], err = a.object(table.ForeignKeyTable(columnName), columnValue)
			if err != nil {
				return nil, err
			}
		}
	}
	// making sure additional columns are stripped out
	entry = a.resultSet.Strip(entry, table.ColumNames())

	// one-to-many: where other tables are having constraints pointing back to this table
	for _, relatedTableName := range a.schema.OneToManyTables(tableName) {
		relatedEntries, err := a.related(relatedTableName, tableName, pk)
		if err != nil {
			return nil, err
		}
		for k, v := range relatedEntries {
			entry[k] = v
		}
	}
	return entry, nil
}

// Build create unstructured objects out of result-set.
func (a *Assembler) Build() ([]*unstructured.Unstructured, error) {
	// to find schema named table, it must be lowered string
	schemaName := strings.ToLower(a.schema.Name)

	// getting the primary-keys for schema named table
	pks, err := a.resultSet.GetColumn(schemaName, orm.PKColumnName)
	if err != nil {
		return nil, err
	}

	a.logger.WithValues("entries", len(pks)).Info("Building objects based on ResultSet.")

	objects := []*unstructured.Unstructured{}
	for _, pk := range pks {
		object, err := a.object(schemaName, pk)
		if err != nil {
			return nil, err
		}
		objects = append(objects, &unstructured.Unstructured{Object: object})
	}
	return objects, nil
}

// NewAssembler instantiate Assembler.
func NewAssembler(logger logr.Logger, schema *orm.Schema, resultSet *orm.ResultSet) *Assembler {
	return &Assembler{
		logger:    logger.WithName("assembler").WithName(schema.Name),
		schema:    schema,
		resultSet: resultSet,
	}
}
