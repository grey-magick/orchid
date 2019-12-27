package repository

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	jsc "github.com/isutton/orchid/pkg/orchid/jsonschema"
	"github.com/isutton/orchid/pkg/orchid/orm"
)

// Assembler is the component to build back unstructured objects from an orm.ResultSet.
type Assembler struct {
	logger logr.Logger    // logger instance
	schema *orm.Schema    // ORM schema
	rs     *orm.ResultSet // orm's result-set
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
	nestedEntries, err := a.rs.Get(relatedTableName, tableName, pk)
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
	relatedEntries, err := a.rs.Get(relatedTableName, tableName, pk)
	if err != nil {
		return nil, err
	}

	strippedEntries := []interface{}{}
	for _, relatedEntry := range relatedEntries {
		strippedEntry := a.rs.Strip(relatedEntry, columns)
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

// amend informed object checking each table's column. Array columns will have a special treatment
// to convert them back into an slice of interface{}. It can return error on casting informed item
// values.
func (a *Assembler) amend(
	table *orm.Table,
	entry map[string]interface{},
) (map[string]interface{}, error) {
	amended := map[string]interface{}{}
	for _, column := range table.Columns {
		if table.IsPrimaryKey(column.Name) {
			continue
		}

		// TODO: check for empty values, for instance instead of nil, it should replace with an
		// empty string. The same for the other types.
		value, found := entry[column.Name]
		if !found {
			return nil, fmt.Errorf("column '%s' not found on entry '%#v'", column.Name, entry)
		}

		if column.JSType == jsc.Array {
			byteSlice, ok := value.([]byte)
			if !ok {
				return nil, fmt.Errorf("unable to scan '%#v' into a byte slice", entry[column.Name])
			}

			trimmed := strings.TrimPrefix(string(byteSlice), "{")
			trimmed = strings.TrimSuffix(trimmed, "}")

			array := []interface{}{}
			for _, item := range strings.Split(trimmed, ",") {
				array = append(array, item)
			}
			amended[column.Name] = array
		} else {
			amended[column.Name] = value
		}
	}
	return amended, nil
}

// object based on table name and primary-key value, recursively create new objects when finding
// one-to-one relationships, and enriching the current object with one-to-many related entries.
func (a *Assembler) object(tableName string, pk interface{}) (map[string]interface{}, error) {
	table, err := a.schema.GetTable(tableName)
	if err != nil {
		return nil, err
	}

	a.logger.WithValues("table", tableName).Info("Retrieving data to assemble new object")
	entry, err := a.rs.GetPK(tableName, pk)
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
	// making sure additional columns are stripped out, and columns are handled properly
	if entry, err = a.amend(table, entry); err != nil {
		return nil, err
	}

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
	pks, err := a.rs.GetColumn(schemaName, orm.PKColumnName)
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
		u := &unstructured.Unstructured{Object: object}
		if u.GroupVersionKind().String() == CRDGVK.String() {
			data, exists, err := unstructured.NestedFieldNoCopy(object, "data")
			if err != nil {
				return nil, err
			}
			if !exists {
				return nil, errors.New("expected field 'data' not present in object")
			}
			uObj, ok := data.([]byte)
			if !ok {
				return nil, errors.New("value in field 'data' can not be converted to map[string]interface{}")
			}
			err = u.UnmarshalJSON(uObj)
			if err != nil {
				return nil, err
			}
		}
		objects = append(objects, u)
	}
	return objects, nil
}

// NewAssembler instantiate Assembler.
func NewAssembler(logger logr.Logger, schema *orm.Schema, rs *orm.ResultSet) *Assembler {
	return &Assembler{
		logger: logger.WithName("assembler").WithName(schema.Name),
		schema: schema,
		rs:     rs,
	}
}
