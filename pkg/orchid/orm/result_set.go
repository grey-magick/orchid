package orm

import (
	"fmt"
	"strings"
)

type ResultSet struct {
	schema *Schema
	Data   MappedEntries
}

func (r *ResultSet) getTableData(tableName string) (*Table, []Entry, error) {
	table, err := r.schema.GetTable(tableName)
	if err != nil {
		return nil, nil, err
	}
	data, found := r.Data[tableName]
	if !found {
		return nil, nil, fmt.Errorf("no data found for for tale named '%s'", tableName)
	}
	return table, data, nil
}

func (r *ResultSet) match(entry Entry, columnName string, columnValue interface{}) bool {
	value, found := entry[columnName]
	if !found {
		return false
	}
	return columnValue == value
}

func (r *ResultSet) Strip(entry Entry, columns []string) map[string]interface{} {
	stripped := Entry{}
	for _, column := range columns {
		stripped[column] = entry[column]
	}
	return stripped
}

func (r *ResultSet) Get(
	tableName,
	columnName string,
	columnValue interface{},
) ([]map[string]interface{}, error) {
	// FIXME: why having to lower it here?
	tableName = strings.ToLower(tableName)

	_, data, err := r.getTableData(tableName)
	if err != nil {
		return nil, err
	}

	entries := []map[string]interface{}{}
	for _, entry := range data {
		if !r.match(entry, columnName, columnValue) {
			continue
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (r *ResultSet) GetPK(tableName string, pk interface{}) (Entry, error) {
	// FIXME: why having to lower it here?
	tableName = strings.ToLower(tableName)

	entries, err := r.Get(tableName, PKColumnName, pk)
	if err != nil {
		return nil, err
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("unable to find data for '%s.%s'='%v'", tableName, PKColumnName, pk)
	}
	if len(entries) > 1 {
		return nil, fmt.Errorf("too many results for for '%s.%s'='%v'", tableName, PKColumnName, pk)
	}
	return entries[0], nil
}

func (r *ResultSet) GetColumn(tableName, columnName string) (List, error) {
	// FIXME: why having to lower it here?
	tableName = strings.ToLower(tableName)

	_, data, err := r.getTableData(tableName)
	if err != nil {
		return nil, err
	}

	list := List{}
	for _, entry := range data {
		value, found := entry[columnName]
		if !found {
			continue
		}
		list = append(list, value)
	}
	return list, nil
}

// processRow receives a single row (from results matrix) and split results into a new map based
// on table hints, present in each column name.
func (r *ResultSet) processRow(columnIDs map[string]int, row List) (EntryMap, error) {
	hintedEntries := EntryMap{}

	for hintedColumn, id := range columnIDs {
		hintColumnSlice := strings.Split(hintedColumn, ".")
		if len(hintColumnSlice) < 2 {
			return nil, fmt.Errorf("unable to obtain hint and column name from '%s'", hintedColumn)
		}

		hint := hintColumnSlice[0]
		column := strings.Join(hintColumnSlice[1:], ".")

		if _, found := hintedEntries[hint]; !found {
			hintedEntries[hint] = Entry{}
		}
		hintedEntries[hint][column] = row[id]
	}
	return hintedEntries, nil
}

// buildMappedMatrix creates the MappedMatrix data structure.
func (r *ResultSet) buildMappedMatrix(columnIDs map[string]int, matrix []List) error {
	columnIDsLen := len(columnIDs)
	pkCache := make(map[string][]interface{}, len(r.schema.Tables))

	for _, row := range matrix {
		if columnIDsLen != len(row) {
			return fmt.Errorf("wrong amount of columns vs. amount of row columns")
		}

		// extracting a map splitting the different tables in different maps organized by hint
		hintedEntries, err := r.processRow(columnIDs, row)
		if err != nil {
			return err
		}

		for _, table := range r.schema.Tables {
			entry, found := hintedEntries[table.Hint]
			if !found {
				continue
			}
			pk, found := entry[PKColumnName]
			if !found {
				continue
			}

			// in all cases, the primary-key should not repeat, therefore ingnoring the one
			// repeating is a way to avoid duplicated results due one-to-many relationships
			if InterfaceSliceContains(pkCache[table.Hint], pk) {
				continue
			}
			pkCache[table.Hint] = append(pkCache[table.Hint], pk)

			r.Data[table.Name] = append(r.Data[table.Name], entry)
		}
	}
	return nil
}

// NewResultSet instantiate an ResultSet.
func NewResultSet(schema *Schema, columnIDs map[string]int, matrix []List) (*ResultSet, error) {
	r := &ResultSet{schema: schema, Data: MappedEntries{}}
	if err := r.buildMappedMatrix(columnIDs, matrix); err != nil {
		return nil, err
	}
	return r, nil
}
