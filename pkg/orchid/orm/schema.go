package orm

import (
	"fmt"

	extv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

// Schema around a given CR (Custom Resource), as in the group of tables required to store CR's
// payload. It also handles JSON-Schema properties to generate additional tables and columns.
type Schema struct {
	Name   string
	Tables map[string]*Table
}

// tableFactory return existing or create new table instance.
func (s *Schema) tableFactory(tableName string) *Table {
	_, exists := s.Tables[tableName]
	if !exists {
		s.Tables[tableName] = NewTable(tableName)
	}
	return s.Tables[tableName]
}

// addCRTable creates the primary table to store CR records. This table can be reused to include
// more columns later on.
func (s *Schema) addCRTable() {
	table := s.tableFactory(s.Name)

	table.AddColumnRaw("kind", PgTypeText)
	table.AddColumnRaw("api_version", PgTypeText)
}

// addObjectMetaTable create the table refering to ObjectMeta CR entry.
func (s *Schema) addObjectMetaTable() {
	table := s.tableFactory(fmt.Sprintf("%s_object_meta", s.Name))

	table.AddColumnRaw("name", PgTypeText)
	table.AddColumnRaw("generate_name", PgTypeText)
	table.AddColumnRaw("namespace", PgTypeText)
	table.AddColumnRaw("self_link", PgTypeText)
	table.AddColumnRaw("uid", PgTypeText)
	table.AddColumnRaw("resource_version", PgTypeText)
	table.AddColumnRaw("generation", PgTypeBigInt)
	table.AddColumnRaw("creation_timestamp", PgTypeText)
	table.AddColumnRaw("deletion_timestamp", PgTypeText)
	table.AddColumnRaw("deletion_grace_period_seconds", PgTypeBigInt)
	table.AddColumnRaw("finalizers", PgTypeTextArray)
	table.AddColumnRaw("cluster_name", PgTypeText)
}

// addObjectMetaLabelsTable part of ObjectMeta, stores labels.
func (s *Schema) addObjectMetaLabelsTable() {
	table := s.tableFactory(fmt.Sprintf("%s_object_meta_labels", s.Name))

	table.AddColumnRaw("name", PgTypeText)
	table.AddColumnRaw("value", PgTypeText)
}

// addObjectMetaAnnotationsTable part of ObjectMeta, stores annotations.
func (s *Schema) addObjectMetaAnnotationsTable() {
	table := s.tableFactory(fmt.Sprintf("%s_object_meta_labels", s.Name))

	table.AddColumnRaw("name", PgTypeText)
	table.AddColumnRaw("value", PgTypeText)
}

// addObjectMetaReferencesTable part of ObjectMeta, stores references.
func (s *Schema) addObjectMetaReferencesTable() {
	table := s.tableFactory(fmt.Sprintf("%s_object_meta_owner_references", s.Name))

	table.AddColumnRaw("api_version", PgTypeText)
	table.AddColumnRaw("kind", PgTypeText)
	table.AddColumnRaw("name", PgTypeText)
	table.AddColumnRaw("controller", PgTypeBoolean)
	table.AddColumnRaw("block_owner_deletion", PgTypeBoolean)
}

// addObjectMetaManagedFieldsTable part of ObjectMeta, stores managed fields.
func (s *Schema) addObjectMetaManagedFieldsTable() {
	table := s.tableFactory(fmt.Sprintf("%s_object_meta_managed_fields", s.Name))

	table.AddColumnRaw("manager", PgTypeText)
	table.AddColumnRaw("operation", PgTypeText)
	table.AddColumnRaw("api_version", PgTypeText)
	table.AddColumnRaw("time", PgTypeText)
	table.AddColumnRaw("fields_type", PgTypeText)
	table.AddColumnRaw("fields_v1", PgTypeText)
}

// Generate trigger generation of metadata and CR tables, plus parsing of JSON-Schema properties to
// create extra columns and tables. Can return error on JSON-Schema parsing.
func (s *Schema) Generate(properties map[string]extv1beta1.JSONSchemaProps) error {
	s.addCRTable()
	s.addObjectMetaTable()
	s.addObjectMetaLabelsTable()
	s.addObjectMetaAnnotationsTable()
	s.addObjectMetaReferencesTable()
	s.addObjectMetaManagedFieldsTable()

	return s.jsonSchemaParser(s.Name, properties)
}

// jsonSchemaParser parse map of properties into more columns or tables, depending on the type of
// entry. It can return errors on not being able to deal with a given JSON-Schema type.
func (s *Schema) jsonSchemaParser(
	tableName string,
	properties map[string]extv1beta1.JSONSchemaProps,
) error {
	table := s.tableFactory(tableName)

	for name, jsonSchema := range properties {
		switch jsonSchema.Type {
		case "object":
			tableName = fmt.Sprintf("%s_%s", tableName, name)
			err := s.jsonSchemaParser(tableName, jsonSchema.Properties)
			if err != nil {
				return err
			}
		case "array":
			itemsSchema := jsonSchema.Items.Schema
			table.AddArrayColumn(name, itemsSchema.Type, itemsSchema.Format, jsonSchema.MaxItems)
		case "string":
			table.AddColumn(name, jsonSchema.Format)
		case "integer":
			table.AddColumn(name, jsonSchema.Format)
		case "long":
			table.AddColumn(name, jsonSchema.Format)
		default:
			return fmt.Errorf("unknown json-schema type '%s'", jsonSchema.Type)
		}
	}
	return nil
}

// NewSchema instantiate new Schema.
func NewSchema(name string) *Schema {
	return &Schema{Name: name, Tables: map[string]*Table{}}
}
