package orm

import (
	"fmt"

	extv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

// Schema around a given CR (Custom Resource), as in the group of tables required to store CR's
// payload. It also handles JSON-Schema properties to generate additional tables and columns.
type Schema struct {
	Name   string            // primary table and schema name
	Tables map[string]*Table // schema tables
}

const (
	// omSuffix ObjectMeta
	omSuffix = "object_meta"
	// omLabelsSuffix ObjectMeta.Labels
	omLabelsSuffix = "object_meta_labels"
	// omAnnotationsSuffix ObjectMeta.Annotations
	omAnnotationsSuffix = "object_meta_annotations"
	// omOwnerReferencesSuffix ObjectMeta.OwnerReferences
	omOwnerReferencesSuffix = "object_meta_owner_references"
	// omManagedFieldsSuffix ObjectMeta.ManagedFields
	omManagedFieldsSuffix = "object_meta_managed_fields"
)

// tableName append suffix on schema name.
func (s *Schema) tableName(suffix string) string {
	return fmt.Sprintf("%s_%s", s.Name, suffix)
}

// tableFactory return existing or create new table instance.
func (s *Schema) tableFactory(tableName string) *Table {
	table, exists := s.Tables[tableName]
	if !exists {
		table = NewTable(tableName)
	}
	s.Tables[tableName] = table
	return s.Tables[tableName]
}

// addCRTable creates the primary table to store CR records. This table can be reused to include
// more columns later on.
func (s *Schema) addCRTable() {
	table := s.tableFactory(s.Name)
	table.AddSerialPK()

	table.AddColumnRaw("kind", PgTypeText)
	table.AddColumnRaw("api_version", PgTypeText)

	table.AddColumnRaw("object_meta_id", PgTypeBigInt)
	table.AddConstraintRaw(
		PgConstraintFK,
		fmt.Sprintf("(object_meta_id) references %s (id)", s.tableName(omSuffix)),
	)
}

// addObjectMetaTable create the table refering to ObjectMeta CR entry. The ObjectMeta type is
// described at https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#ObjectMeta.
func (s *Schema) addObjectMetaTable() {
	table := s.tableFactory(s.tableName(omSuffix))
	table.AddSerialPK()

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

	table.AddColumnRaw("labels_id", PgTypeBigInt)
	table.AddConstraintRaw(
		PgConstraintFK,
		fmt.Sprintf("(labels_id) references %s (id)", s.tableName(omLabelsSuffix)),
	)

	table.AddColumnRaw("annotations_id", PgTypeBigInt)
	table.AddConstraintRaw(
		PgConstraintFK,
		fmt.Sprintf("(annotations_id) references %s (id)", s.tableName(omAnnotationsSuffix)),
	)

	table.AddColumnRaw("owner_references_id", PgTypeBigInt)
	table.AddConstraintRaw(
		PgConstraintFK,
		fmt.Sprintf("(owner_references_id) references %s (id)",
			s.tableName(omOwnerReferencesSuffix)),
	)

	table.AddColumnRaw("finalizers", PgTypeTextArray)
	table.AddColumnRaw("cluster_name", PgTypeText)

	table.AddColumnRaw("managed_fields_id", PgTypeBigInt)
	table.AddConstraintRaw(
		PgConstraintFK,
		fmt.Sprintf("(owner_references_id) references %s (id)",
			s.tableName(omManagedFieldsSuffix)),
	)
}

// addObjectMetaLabelsTable part of ObjectMeta, stores labels.
func (s *Schema) addObjectMetaLabelsTable() {
	table := s.tableFactory(s.tableName(omLabelsSuffix))
	table.AddSerialPK()

	table.AddColumnRaw("id", PgTypeBigInt)
	table.AddConstraintRaw(PgConstraintPK, "(id)")

	table.AddColumnRaw("name", PgTypeText)
	table.AddConstraintRaw(PgConstraintUnique, "(name)")

	table.AddColumnRaw("value", PgTypeText)
}

// addObjectMetaAnnotationsTable part of ObjectMeta, stores annotations.
func (s *Schema) addObjectMetaAnnotationsTable() {
	table := s.tableFactory(s.tableName(omAnnotationsSuffix))

	table.AddColumnRaw("id", PgTypeBigInt)
	table.AddConstraintRaw(PgConstraintPK, "(id)")

	table.AddColumnRaw("name", PgTypeText)
	table.AddConstraintRaw(PgConstraintUnique, "(name)")

	table.AddColumnRaw("value", PgTypeText)
}

// addObjectMetaReferencesTable part of ObjectMeta, stores references.
func (s *Schema) addObjectMetaReferencesTable() {
	table := s.tableFactory(s.tableName(omOwnerReferencesSuffix))

	table.AddColumnRaw("id", PgTypeBigInt)
	table.AddConstraintRaw(PgConstraintPK, "(id)")

	table.AddColumnRaw("api_version", PgTypeText)
	table.AddColumnRaw("kind", PgTypeText)
	table.AddColumnRaw("name", PgTypeText)
	table.AddColumnRaw("controller", PgTypeBoolean)
	table.AddColumnRaw("block_owner_deletion", PgTypeBoolean)
}

// addObjectMetaManagedFieldsTable part of ObjectMeta, stores managed fields.
func (s *Schema) addObjectMetaManagedFieldsTable() {
	table := s.tableFactory(s.tableName(omManagedFieldsSuffix))

	table.AddColumnRaw("id", PgTypeBigInt)
	table.AddConstraintRaw(PgConstraintPK, "(id)")

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
	s.addObjectMetaReferencesTable()
	s.addObjectMetaManagedFieldsTable()
	s.addObjectMetaLabelsTable()
	s.addObjectMetaAnnotationsTable()
	s.addObjectMetaTable()
	s.addCRTable()

	return s.jsonSchemaParser(s.Name, properties)
}

// jsonSchemaParser parse map of properties into more columns or tables, depending on the type of
// entry. It can return errors on not being able to deal with a given JSON-Schema type.
func (s *Schema) jsonSchemaParser(
	tableName string,
	properties map[string]extv1beta1.JSONSchemaProps,
) error {
	table := s.tableFactory(tableName)

	// on creating new tables, making sure they have a primary-key
	if s.Name != tableName {
		table.AddSerialPK()
	}

	for name, jsonSchema := range properties {
		switch jsonSchema.Type {
		case "object":
			childTableName := fmt.Sprintf("%s_%s", tableName, name)
			table.AddConstraintRaw(
				PgConstraintFK,
				fmt.Sprintf("(id) references %s", childTableName),
			)
			if err := s.jsonSchemaParser(childTableName, jsonSchema.Properties); err != nil {
				return err
			}
		case "array":
			itemsSchema := jsonSchema.Items.Schema
			itemType := itemsSchema.Type
			itemformat := itemsSchema.Format
			max := jsonSchema.MaxItems
			if err := table.AddArrayColumn(name, itemType, itemformat, max); err != nil {
				return err
			}
		case "string":
			if err := table.AddColumn(name, jsonSchema.Type, jsonSchema.Format); err != nil {
				return err
			}
		case "integer":
			if err := table.AddColumn(name, jsonSchema.Type, jsonSchema.Format); err != nil {
				return err
			}
		case "long":
			if err := table.AddColumn(name, jsonSchema.Type, jsonSchema.Format); err != nil {
				return err
			}
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
