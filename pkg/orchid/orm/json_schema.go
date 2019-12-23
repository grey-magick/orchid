package orm

import (
	"fmt"

	"github.com/go-logr/logr"
	extv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

// JSONSchemaParser recursively create a set of tables, having one-to-one and one-to-many
// relationships based on informed JSON-Schema.
type JSONSchemaParser struct {
	logger logr.Logger // logger instance
	schema *Schema     // schema instance
}

const (
	JSTypeArray   = "array"
	JSTypeBoolean = "boolean"
	JSTypeInteger = "integer"
	JSTypeNumber  = "number"
	JSTypeObject  = "object"
	JSTypeString  = "string"
)

// expandAdditionalProperties will create a set of properties to represent a key-value object.
func (j *JSONSchemaParser) expandAdditionalProperties(
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

// object creates extra column and recursively new tables.
func (j *JSONSchemaParser) object(
	table *Table,
	columnName string,
	notNull bool,
	jsSchema extv1beta1.JSONSchemaProps,
) error {
	logger := j.logger.WithValues("table", table.Name, "column", columnName, "notNull", notNull)
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
			logger.Info("Adding local foreign-key for one-to-one relationship.")
			logger.Info("Adding new column on table.")
			table.AddBigIntFK(columnName, relatedTableName, PKColumnName, notNull)
			table.AddConstraint(&Constraint{Type: PgConstraintUnique, ColumnName: columnName})
		}
	} else {
		if additionalProperties.Schema == nil || additionalProperties.Schema.Type == "" {
			return fmt.Errorf("unable to define schema from additionalProperties: '%+v'",
				additionalProperties)
		}
		logger.Info("Expanding additional properties to a key-value table, one-to-many.")
		jsSchema = j.expandAdditionalProperties(additionalProperties, table.Name)

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

	return j.Parse(relatedTableName, relationship, &jsSchema)
}

// array by either adding a column type array, but when it's an array of objects, an
// one-to-many relationship is created.
func (j *JSONSchemaParser) array(
	table *Table,
	columnName string,
	notNull bool,
	jsSchema extv1beta1.JSONSchemaProps,
) error {
	logger := j.logger.WithValues("table", table.Name, "column", columnName, "notNull", notNull)

	if jsSchema.Items == nil || jsSchema.Items.Schema == nil {
		return fmt.Errorf("items is not found under json-schema: '%+v'", jsSchema)
	}
	itemsSchema := jsSchema.Items.Schema

	// in case of being an array of objects, it needs to spin off a new table
	if itemsSchema.Type == JSTypeObject {
		logger.Info("Creating new object to handle array column.")

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
		return j.Parse(relatedTableName, relationship, itemsSchema)
	}

	// adding a new column to existing table, a single dimension array
	jsType := itemsSchema.Type
	jsFormat := itemsSchema.Format
	maxItems := jsSchema.MaxItems

	logger.WithValues("type", jsType, "format", jsFormat, "maxItems", maxItems).
		Info("Adding a new array column.")

	column, err := NewColumnArray(columnName, jsType, jsFormat, maxItems, notNull)
	if err != nil {
		return err
	}
	table.AddColumn(column)
	return nil
}

// column entries that can be translated to a simple column.
func (j *JSONSchemaParser) column(
	table *Table,
	columnName string,
	notNull bool,
	jsonSchema extv1beta1.JSONSchemaProps,
) error {
	j.logger.WithValues(
		"table", table.Name,
		"column", columnName,
		"notNull", notNull,
		"type", jsonSchema.Type,
		"format", jsonSchema.Format,
	).Info("Adding new column to table.")
	column, err := NewColumn(columnName, jsonSchema.Type, jsonSchema.Format, notNull)
	if err != nil {
		return err
	}
	table.AddColumn(column)
	return nil
}

// jsonSchemaParser parse map of properties into more columns or tables, depending on the type of
// entry. It can return errors on not being able to deal with a given JSON-Schema type.
func (j *JSONSchemaParser) Parse(
	tableName string,
	relationship Relationship,
	jsSchema *extv1beta1.JSONSchemaProps,
) error {
	j.logger.WithValues("table", tableName, "relationship", relationship).
		Info("Adding new table on schema.")

	properties := jsSchema.Properties
	required := jsSchema.Required

	table := j.schema.TableFactory(tableName, relationship.OneToMany)
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
		notNull := StringSliceContains(required, name)

		switch jsSchema.Type {
		case JSTypeObject:
			err = j.object(table, name, notNull, jsSchema)
		case JSTypeArray:
			err = j.array(table, name, notNull, jsSchema)
		case JSTypeBoolean:
			err = j.column(table, name, notNull, jsSchema)
		case JSTypeString:
			err = j.column(table, name, notNull, jsSchema)
		case JSTypeInteger:
			err = j.column(table, name, notNull, jsSchema)
		case JSTypeNumber:
			err = j.column(table, name, notNull, jsSchema)
		default:
			return fmt.Errorf("unknown json-schema type '%s'", jsSchema.Type)
		}
	}
	return err
}

// NewJSONSchemaParser instantiate a new JSONSchemaParser.
func NewJSONSchemaParser(logger logr.Logger, schema *Schema) *JSONSchemaParser {
	return &JSONSchemaParser{
		logger: logger.WithName("json-schema-parser"),
		schema: schema,
	}
}
