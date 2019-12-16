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

// TableFactory return existing or create new table instance.
func (s *Schema) TableFactory(tableName string, tablePath []string) *Table {
	for _, table := range s.Tables {
		if tableName == table.Name {
			return table
		}
	}
	table := NewTable(tableName, tablePath)
	s.prepend(table)
	return table
}

// GetTable returns a table instance, if exists.
func (s *Schema) GetTable(tableName string) *Table {
	for _, table := range s.Tables {
		if tableName == table.Name {
			return table
		}
	}
	return nil
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
	tablePath []string,
	columnName string,
	notNull bool,
	jsSchema extv1beta1.JSONSchemaProps,
) error {
	additionalProperties := jsSchema.AdditionalProperties
	relatedTableName := fmt.Sprintf("%s_%s", table.Name, columnName)

	// making sure either AdditionalProperties or Properties are set
	if (additionalProperties == nil && len(jsSchema.Properties) == 0) ||
		(additionalProperties != nil && len(jsSchema.Properties) > 0) {
		return fmt.Errorf("unable to generate table/column with json-schema '%+v'", jsSchema)
	}

	constraints := []*Constraint{}
	if additionalProperties != nil {
		if additionalProperties.Schema == nil || additionalProperties.Schema.Type == "" {
			return fmt.Errorf("unable to define schema from additionalProperties: '%+v'",
				additionalProperties)
		}
		jsSchema = s.expandAdditionalProperties(additionalProperties, table.Name)
		// triggering a one-to-many relationship, the next table to be created will get a foreign-key
		// containing this current table's primary-key.
		constraints = append(constraints, &Constraint{
			Type:              PgConstraintFK,
			ColumnName:        table.Name,
			RelatedTableName:  table.Name,
			RelatedColumnName: "id",
		})
	} else {
		// managing an one-to-one relationship, this table will keep a foreign-key pointing to the
		// next table to be created by primary-key
		if table.GetColumn(relatedTableName) == nil {
			table.AddBigIntFK(columnName, relatedTableName, "id", notNull)
			table.AddConstraint(&Constraint{Type: PgConstraintUnique, ColumnName: columnName})
		}
	}

	tablePath = append(tablePath, columnName)
	return s.jsonSchemaParser(relatedTableName, tablePath, constraints, &jsSchema)
}

// handleArray creates an array column.
func (s *Schema) handleArray(
	table *Table,
	name string,
	notNull bool,
	jsonSchema extv1beta1.JSONSchemaProps,
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
	jsonSchema extv1beta1.JSONSchemaProps,
) error {
	column, err := NewColumn(name, jsonSchema.Type, jsonSchema.Format, notNull)
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
	tablePath []string,
	constraints []*Constraint,
	jsSchema *extv1beta1.JSONSchemaProps,
) error {
	properties := jsSchema.Properties
	required := jsSchema.Required

	table := s.TableFactory(tableName, tablePath)
	// default serial primary key
	table.AddSerialPK()
	// adding expected table constraints
	if len(constraints) > 0 {
		for _, constraint := range constraints {
			table.AddColumn(&Column{Type: PgTypeBigInt, Name: constraint.ColumnName, NotNull: true})
			table.AddConstraint(constraint)
		}
	}

	for name, jsSchema := range properties {
		// checking if property name required, therefore not null column
		notNull := s.isRequiredProp(name, required)

		switch jsSchema.Type {
		case JSTypeObject:
			if err := s.handleObject(table, tablePath, name, notNull, jsSchema); err != nil {
				return err
			}
		case JSTypeArray:
			if err := s.handleArray(table, name, notNull, jsSchema); err != nil {
				return err
			}
		case JSTypeBoolean:
			if err := s.handleColumn(table, name, notNull, jsSchema); err != nil {
				return err
			}
		case JSTypeString:
			if err := s.handleColumn(table, name, notNull, jsSchema); err != nil {
				return err
			}
		case JSTypeInteger:
			if err := s.handleColumn(table, name, notNull, jsSchema); err != nil {
				return err
			}
		case JSTypeNumber:
			if err := s.handleColumn(table, name, notNull, jsSchema); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown json-schema type '%s'", jsSchema.Type)
		}
	}
	return nil
}

// GenerateCR trigger generation of metadata and CR tables, plus parsing of OpenAPIV3 Schema to create
// tables and columns. Can return error on JSON-Schema parsing.
func (s *Schema) GenerateCR(openAPIV3Schema *extv1beta1.JSONSchemaProps) error {
	// intercepting "metadata" attribute, making sure only on the first level
	if _, found := openAPIV3Schema.Properties["metadata"]; found {
		metadata := openAPIV3Schema.Properties["metadata"]
		metadata.Properties = metaV1ObjectMetaOpenAPIV3Schema()
		openAPIV3Schema.Properties["metadata"] = metadata
	}

	return s.jsonSchemaParser(s.Name, []string{}, nil, openAPIV3Schema)
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
