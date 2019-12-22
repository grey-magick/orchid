package repository

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/isutton/orchid/pkg/orchid/orm"
)

// Assembler is the component to build back unstructured objects from an orm.ResultSet.
type Assembler struct {
	schema    *orm.Schema    // ORM schema
	resultSet *orm.ResultSet // data result-set
}

// kvObject create object based on result-set, but assume different names for columns and structure
// of data as key-value only.
func (a *Assembler) kvObject(
	relatedTableName string,
	tableName string,
	pk interface{},
) (map[string]interface{}, error) {
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

// sliceObject return result-set data as slice of interface.
func (a *Assembler) sliceObject(
	relatedTableName string,
	tableName string,
	pk interface{},
	columns []string,
) ([]interface{}, error) {
	relatedEntries, err := a.resultSet.Get(relatedTableName, tableName, pk)
	if err != nil {
		return nil, err
	}

	strippedEntries := []interface{}{}
	for _, relatedEntry := range relatedEntries {
		strippedEntry := a.resultSet.Strip(relatedEntry, columns)
		strippedEntries = append(strippedEntries, strippedEntry)
	}

	return orm.InterfaceSliceReversed(strippedEntries), nil
}

// relatedObject goes after one-to-many relationships.
func (a *Assembler) relatedObject(
	relatedTableName string,
	tableName string,
	pk interface{},
) (map[string]interface{}, error) {
	table, err := a.schema.GetTable(relatedTableName)
	if err != nil {
		return nil, err
	}

	// using path to decide column's name
	tablePathLen := len(table.Path)
	if tablePathLen == 0 {
		return nil, fmt.Errorf("unabel to determine column name based on table's path")
	}
	columnName := table.Path[tablePathLen-1]

	entry := map[string]interface{}{}
	if table.KV {
		entry[columnName], err = a.kvObject(relatedTableName, tableName, pk)
		if err != nil {
			return nil, err
		}
	} else {
		columns := table.ColumnNamesStripped()
		entry[columnName], err = a.sliceObject(relatedTableName, tableName, pk, columns)
		if err != nil {
			return nil, err
		}
	}
	return entry, nil
}

// createObject based on table name and primary-key value, recursively create new objects and
// enrich them with one-to-many related entries.
func (a *Assembler) createObject(tableName string, pk interface{}) (map[string]interface{}, error) {
	table, err := a.schema.GetTable(tableName)
	if err != nil {
		return nil, err
	}

	entry, err := a.resultSet.GetPK(tableName, pk)
	if err != nil {
		return nil, err
	}

	// one-to-ene: when this table is refering another via foreign-keys
	for columnName, columnValue := range entry {
		if table.IsForeignKey(columnName) {
			entry[columnName], err = a.createObject(table.ForeignKeyTable(columnName), columnValue)
			if err != nil {
				return nil, err
			}
		}
	}
	// making sure additional columns are stripped out
	entry = a.resultSet.Strip(entry, table.ColumNames())

	// one-to-many: where other tables are having constraints pointing back to this table
	for _, relatedTableName := range a.schema.OneToManyTables(tableName) {
		relatedEntries, err := a.relatedObject(relatedTableName, tableName, pk)
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
	// getting the primary-keys for schema named table
	pks, err := a.resultSet.GetColumn(a.schema.Name, "id")
	if err != nil {
		return nil, err
	}

	objects := []*unstructured.Unstructured{}
	for _, pk := range pks {
		object, err := a.createObject(a.schema.Name, pk)
		if err != nil {
			return nil, err
		}
		objects = append(objects, &unstructured.Unstructured{Object: object})
	}
	return objects, nil
}

// NewAssembler instantiate Assembler.
func NewAssembler(schema *orm.Schema, resultSet *orm.ResultSet) *Assembler {
	return &Assembler{schema: schema, resultSet: resultSet}
}
