package orm

import (
	"fmt"

	extv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

// Schema around a given CR (Custom Resource), as in the group of tables required to store CR's
// payload. It also handles JSON-Schema properties to generate additional tables and columns.
type Schema struct {
	Name   string   // primary table and schema name
	Tables []*Table // schema tables
}

// TableName append suffix on schema name.
func (s *Schema) TableName(suffix string) string {
	return fmt.Sprintf("%s_%s", s.Name, suffix)
}

// prepend new tables on the beggining of the slice. Reverse order helps to deal with the creation
// of tables that are dependant on each other.
func (s *Schema) prepend(table *Table) {
	s.Tables = append(s.Tables, table)
	copy(s.Tables[1:], s.Tables)
	s.Tables[0] = table
}

// TableFactory return existing or create new table instance.
func (s *Schema) TableFactory(tableName string) *Table {
	for _, table := range s.Tables {
		if tableName == table.Name {
			return table
		}
	}
	table := NewTable(tableName)
	s.prepend(table)
	return table
}

// isRequiredProp checks for required properties by being part of required string slice.
func (s *Schema) isRequiredProp(name string, required []string) bool {
	for _, requiredProp := range required {
		if name == requiredProp {
			return true
		}
	}
	return false
}

// handleObject creates extra column and recursively new tables.
func (s *Schema) handleObject(
	table *Table,
	name string,
	notNull bool,
	jsonSchema *extv1beta1.JSONSchemaProps,
) error {
	if name == "metadata" && s.Name == table.Name {
		metadata := NewMetadata(s)
		metadata.Add(table)
		return nil
	}

	relatedTableName := fmt.Sprintf("%s_%s", table.Name, name)
	columnName := fmt.Sprintf("%s_id", name)
	table.AddBigIntFK(columnName, relatedTableName, notNull)

	return s.jsonSchemaParser(relatedTableName, jsonSchema.Properties, jsonSchema.Required)
}

// handleArray creates an array column.
func (s *Schema) handleArray(
	table *Table,
	name string,
	notNull bool,
	jsonSchema *extv1beta1.JSONSchemaProps,
) error {
	itemsSchema := jsonSchema.Items.Schema
	column, err := NewColumnArray(
		name,
		itemsSchema.Type,
		itemsSchema.Format,
		jsonSchema.MaxItems,
		notNull,
	)
	if err != nil {
		return err
	}
	table.AddColumn(column)
	return nil
}

// handleColumn entries that can be translated to a simple column.
func (s *Schema) handleColumn(
	table *Table,
	name string,
	notNull bool,
	jsonSchema *extv1beta1.JSONSchemaProps,
) error {
	column, err := NewColumn(
		name,
		jsonSchema.Type,
		jsonSchema.Format,
		notNull,
	)
	if err != nil {
		return err
	}
	table.AddColumn(column)
	return nil
}

// jsonSchemaParser parse map of properties into more columns or tables, depending on the type of
// entry. It can return errors on not being able to deal with a given JSON-Schema type.
func (s *Schema) jsonSchemaParser(
	tableName string,
	properties map[string]extv1beta1.JSONSchemaProps,
	required []string,
) error {
	table := s.TableFactory(tableName)
	table.AddSerialPK()

	for name, jsonSchema := range properties {
		notNull := s.isRequiredProp(name, required)
		switch jsonSchema.Type {
		case "object":
			if err := s.handleObject(table, name, notNull, &jsonSchema); err != nil {
				return err
			}
		case "array":
			if err := s.handleArray(table, name, notNull, &jsonSchema); err != nil {
				return err
			}
		case "string":
			if err := s.handleColumn(table, name, notNull, &jsonSchema); err != nil {
				return err
			}
		case "integer":
			if err := s.handleColumn(table, name, notNull, &jsonSchema); err != nil {
				return err
			}
		case "long":
			if err := s.handleColumn(table, name, notNull, &jsonSchema); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown json-schema type '%s'", jsonSchema.Type)
		}
	}
	return nil
}

// GenerateCR trigger generation of metadata and CR tables, plus parsing of OpenAPIV3 Schema to create
// tables and columns. Can return error on JSON-Schema parsing.
func (s *Schema) GenerateCR(openAPIV3Schema *extv1beta1.JSONSchemaProps) error {
	return s.jsonSchemaParser(s.Name, openAPIV3Schema.Properties, openAPIV3Schema.Required)
}

// GenerateCRD creates the tables to store the actual CRDs.
func (s *Schema) GenerateCRD() {
	crd := NewCRD(s)
	crd.Add()
}

// TablesReversed reverse list of tables in Schema.
func (s *Schema) TablesReversed() []*Table {
	reversed := make([]*Table, len(s.Tables))
	copy(reversed, s.Tables)
	for i := len(reversed)/2 - 1; i >= 0; i-- {
		opposite := len(reversed) - 1 - i
		reversed[i], reversed[opposite] = reversed[opposite], reversed[i]
	}
	return reversed
}

// NewSchema instantiate new Schema.
func NewSchema(name string) *Schema {
	return &Schema{Name: name}
}
