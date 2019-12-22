package orm

import (
	"fmt"
	"strings"

	extv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

// Schema around a given CR (Custom Resource), as in the group of tables required to store CR's
// payload. It also handles JSON-Schema properties to generate additional tables and columns.
type Schema struct {
	Name   string   // primary table and schema name
	Tables []*Table // schema tables
}

// Relationship describes what a new table needs to have in order to establish a relationship
type Relationship struct {
	Path        []string      // data path in respective to original object
	Constraints []*Constraint // list of constraints
	OneToMany   bool          // one-to-many flag
	KV          bool          // key-value flag
}

// HasOneToMany checks if a table matching fieldPath is one-to-many
func (s *Schema) HasOneToMany(fieldPath []string) bool {
	table := s.GetTableByPath(fieldPath)
	if table == nil {
		return false
	}
	return table.OneToMany
}

// IsKV checks if a table matching fieldPath is key-value.
func (s *Schema) IsKV(fieldPath []string) bool {
	table := s.GetTableByPath(fieldPath)
	if table == nil {
		return false
	}
	return table.KV
}

// GetHint based on table's name return its hint.
func (s *Schema) GetHint(tableName string) string {
	table, err := s.GetTable(tableName)
	if err != nil {
		return ""
	}
	return table.Hint
}

// handleArray creates an array column.
// TableName append suffix on schema name.
func (s *Schema) TableName(suffix string) string {
	name := strings.ReplaceAll(s.Name, ".", "_")
	return fmt.Sprintf("%s_%s", name, suffix)
}

// prepend new tables on the beggining of the slice. Reverse order helps to deal with the creation
// of tables that are dependant on each other.
func (s *Schema) prepend(table *Table) {
	s.Tables = append(s.Tables, table)
	copy(s.Tables[1:], s.Tables)
	s.Tables[0] = table
}

// TableFactory return existing or create new table instance, where when one-to-many flag true it
// appends the table instead of prepending. The sequence of tables matters during table creation
// and insertion of data.
func (s *Schema) TableFactory(tableName string, appendTable bool) *Table {
	for _, table := range s.Tables {
		if tableName == table.Name {
			return table
		}
	}
	table := NewTable(tableName)
	if appendTable {
		s.Tables = append(s.Tables, table)
	} else {
		s.prepend(table)
	}
	return table
}

// TODO: return table and error, so other components don't need to write down the same err;
// GetTable returns a table instance, if exists.
func (s *Schema) GetTable(tableName string) (*Table, error) {
	tableName = strings.ToLower(tableName)
	for _, table := range s.Tables {
		if tableName == table.Name {
			return table, nil
		}
	}
	return nil, fmt.Errorf("unable to find table '%s' in schema", tableName)
}

// GetTableByPath returns the table instance having informed path, or nil.
func (s *Schema) GetTableByPath(fieldPath []string) *Table {
	for _, table := range s.Tables {
		if len(fieldPath) != len(table.Path) {
			continue
		}
		equals := true
		for i, field := range fieldPath {
			if table.Path[i] != field {
				equals = false
				break
			}
		}
		if !equals {
			continue
		}
		return table
	}
	return nil
}

// FIXME: rename it;
func (s *Schema) OneToManyTables(tableName string) []string {
	tables := []string{}
	for _, table := range s.Tables {
		if tableName == table.Name {
			continue
		}

		for _, constraint := range table.ForeignKeys() {
			if constraint.ColumnName != tableName {
				continue
			}
			if constraint.RelatedTableName != tableName {
				continue
			}
			tables = append(tables, table.Name)
		}
	}
	return tables
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

// isRequiredProp checks for required properties by being part of required string slice.
func (s *Schema) isRequiredProp(name string, required []string) bool {
	for _, requiredProp := range required {
		if name == requiredProp {
			return true
		}
	}
	return false
}

// expandAdditionalProperties will create a set of properties to represent a key-value object.
func (s *Schema) expandAdditionalProperties(
	additionalProperties *extv1beta1.JSONSchemaPropsOrBool,
	columnName string,
) extv1beta1.JSONSchemaProps {
	additionalSchema := additionalProperties.Schema
	required := []string{"key", "value"}
	properties := map[string]extv1beta1.JSONSchemaProps{
		"key":   jsonSchemaProps(additionalSchema.Type, additionalSchema.Format, nil, nil, nil),
		"value": jsonSchemaProps(additionalSchema.Type, additionalSchema.Format, nil, nil, nil),
	}
	return jsonSchemaProps(JSTypeObject, "", required, nil, properties)
}

// handleObject creates extra column and recursively new tables.
func (s *Schema) handleObject(
	table *Table,
	columnName string,
	notNull bool,
	jsSchema extv1beta1.JSONSchemaProps,
) error {
	relationship := Relationship{Path: append(table.Path, columnName)}
	additionalProperties := jsSchema.AdditionalProperties
	relatedTableName := fmt.Sprintf("%s_%s", table.Name, columnName)

	// making sure either AdditionalProperties or Properties are set
	if (additionalProperties == nil && len(jsSchema.Properties) == 0) ||
		(additionalProperties != nil && len(jsSchema.Properties) > 0) {
		return fmt.Errorf("unable to generate table/column with json-schema '%+v'", jsSchema)
	}

	if additionalProperties == nil {
		// managing an one-to-one relationship, this table will keep a foreign-key pointing to the
		// next table to be created by primary-key
		if table.GetColumn(relatedTableName) == nil {
			table.AddBigIntFK(columnName, relatedTableName, PKColumnName, notNull)
			table.AddConstraint(&Constraint{Type: PgConstraintUnique, ColumnName: columnName})
		}
	} else {
		if additionalProperties.Schema == nil || additionalProperties.Schema.Type == "" {
			return fmt.Errorf("unable to define schema from additionalProperties: '%+v'",
				additionalProperties)
		}
		jsSchema = s.expandAdditionalProperties(additionalProperties, table.Name)

		// triggering a one-to-many relationship, the next table to be created will get a foreign-key
		// containing this current table's primary-key.
		relationship.KV = true
		relationship.OneToMany = true
		relationship.Constraints = append(relationship.Constraints, &Constraint{
			Type:              PgConstraintFK,
			ColumnName:        table.Name,
			RelatedTableName:  table.Name,
			RelatedColumnName: PKColumnName,
		})
	}

	return s.jsonSchemaParser(relatedTableName, relationship, &jsSchema)
}

func (s *Schema) handleArray(
	table *Table,
	columnName string,
	notNull bool,
	jsSchema extv1beta1.JSONSchemaProps,
) error {
	if jsSchema.Items == nil || jsSchema.Items.Schema == nil {
		return fmt.Errorf("items is not found under json-schema: '%+v'", jsSchema)
	}
	itemsSchema := jsSchema.Items.Schema

	// in case of being an array of objects, it needs to spin off a new table
	if itemsSchema.Type == JSTypeObject {
		constraint := &Constraint{
			Type:              PgConstraintFK,
			ColumnName:        table.Name,
			RelatedTableName:  table.Name,
			RelatedColumnName: PKColumnName,
		}
		relationship := Relationship{
			Path:        append(table.Path, columnName),
			Constraints: []*Constraint{constraint},
			OneToMany:   true,
		}
		relatedTableName := fmt.Sprintf("%s_%s", table.Name, columnName)
		return s.jsonSchemaParser(relatedTableName, relationship, itemsSchema)
	}

	// adding a new column to existing table, a single dimension array
	jsType := itemsSchema.Type
	jsFormat := itemsSchema.Format
	maxItems := jsSchema.MaxItems
	column, err := NewColumnArray(columnName, jsType, jsFormat, maxItems, notNull)
	if err != nil {
		return err
	}
	table.AddColumn(column)
	return nil
}

// handleColumn entries that can be translated to a simple column.
func (s *Schema) handleColumn(
	table *Table,
	columnName string,
	notNull bool,
	jsonSchema extv1beta1.JSONSchemaProps,
) error {
	column, err := NewColumn(columnName, jsonSchema.Type, jsonSchema.Format, notNull)
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
	relationship Relationship,
	jsSchema *extv1beta1.JSONSchemaProps,
) error {
	properties := jsSchema.Properties
	required := jsSchema.Required

	table := s.TableFactory(tableName, relationship.OneToMany)
	// default serial primary key
	table.AddSerialPK()
	// changing table attributes according to the relationship
	table.Path = relationship.Path
	table.OneToMany = relationship.OneToMany
	table.KV = relationship.KV
	for _, constraint := range relationship.Constraints {
		table.AddColumn(&Column{Type: PgTypeBigInt, Name: constraint.ColumnName, NotNull: true})
		table.AddConstraint(constraint)
	}

	var err error
	for name, jsSchema := range properties {
		// checking if property name required, therefore not null column
		notNull := s.isRequiredProp(name, required)

		switch jsSchema.Type {
		case JSTypeObject:
			err = s.handleObject(table, name, notNull, jsSchema)
		case JSTypeArray:
			err = s.handleArray(table, name, notNull, jsSchema)
		case JSTypeBoolean:
			err = s.handleColumn(table, name, notNull, jsSchema)
		case JSTypeString:
			err = s.handleColumn(table, name, notNull, jsSchema)
		case JSTypeInteger:
			err = s.handleColumn(table, name, notNull, jsSchema)
		case JSTypeNumber:
			err = s.handleColumn(table, name, notNull, jsSchema)
		default:
			return fmt.Errorf("unknown json-schema type '%s'", jsSchema.Type)
		}
	}
	return err
}

// GenerateCR trigger generation of metadata and CR tables, plus parsing of OpenAPIV3 Schema to
// create tables and columns. Can return error on JSON-Schema parsing.
func (s *Schema) GenerateCR(openAPIV3Schema *extv1beta1.JSONSchemaProps) error {
	// intercepting "metadata" attribute, making sure only on the first level
	if _, found := openAPIV3Schema.Properties["metadata"]; found {
		metadata := openAPIV3Schema.Properties["metadata"]
		metadata.Properties = metaV1ObjectMetaOpenAPIV3Schema()
		openAPIV3Schema.Properties["metadata"] = metadata
	}

	return s.jsonSchemaParser(s.Name, Relationship{}, openAPIV3Schema)
}

// GenerateCRD creates the tables to store the actual CRDs.
func (s *Schema) GenerateCRD() {
	crd := NewCRD(s)
	crd.Add()
}

// NewSchema instantiate new Schema.
func NewSchema(name string) *Schema {
	return &Schema{Name: name}
}
