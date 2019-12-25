package orm

import (
	"fmt"

	"github.com/go-logr/logr"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	jsc "github.com/isutton/orchid/pkg/orchid/jsonschema"
)

// Parser recursively create a set of tables, having one-to-one and one-to-many
// relationships based on informed JSON-Schema.
type Parser struct {
	logger logr.Logger // logger instance
	schema *Schema     // schema instance
}

// expandAdditionalProperties will create a set of properties to represent a key-value object.
func (j *Parser) expandAdditionalProperties(
	additionalProperties *extv1.JSONSchemaPropsOrBool,
	columnName string,
) extv1.JSONSchemaProps {
	additionalSchema := additionalProperties.Schema
	required := []string{"key", "value"}
	properties := map[string]extv1.JSONSchemaProps{
		"key":   jsc.JSONSchemaProps(additionalSchema.Type, additionalSchema.Format, nil, nil, nil),
		"value": jsc.JSONSchemaProps(additionalSchema.Type, additionalSchema.Format, nil, nil, nil),
	}
	return jsc.JSONSchemaProps(jsc.Object, "", required, nil, properties)
}

// object creates extra column and recursively new tables.
func (j *Parser) object(
	table *Table,
	columnName string,
	notNull bool,
	jsSchema extv1.JSONSchemaProps,
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
func (j *Parser) array(
	table *Table,
	columnName string,
	notNull bool,
	jsSchema extv1.JSONSchemaProps,
) error {
	logger := j.logger.WithValues("table", table.Name, "column", columnName, "notNull", notNull)

	if jsSchema.Items == nil || jsSchema.Items.Schema == nil {
		return fmt.Errorf("items is not found under json-schema: '%+v'", jsSchema)
	}
	itemsSchema := jsSchema.Items.Schema

	// in case of being an array of objects, it needs to spin off a new table
	if itemsSchema.Type == jsc.Object {
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
func (j *Parser) column(
	table *Table,
	columnName string,
	notNull bool,
	jsSchema extv1.JSONSchemaProps,
) error {
	j.logger.WithValues(
		"table", table.Name,
		"column", columnName,
		"notNull", notNull,
		"type", jsSchema.Type,
		"format", jsSchema.Format,
	).Info("Adding new column to table.")
	column, err := NewColumn(columnName, jsSchema.Type, jsSchema.Format, notNull)
	if err != nil {
		return err
	}
	table.AddColumn(column)
	return nil
}

// Parse map of properties into more columns or tables, depending on the type of entry. It can
// return errors on not being able to deal with a given JSON-Schema type.
func (j *Parser) Parse(
	tableName string,
	relationship Relationship,
	jsSchema *extv1.JSONSchemaProps,
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
		case jsc.Object:
			err = j.object(table, name, notNull, jsSchema)
		case jsc.Array:
			err = j.array(table, name, notNull, jsSchema)
		case jsc.Boolean:
			err = j.column(table, name, notNull, jsSchema)
		case jsc.String:
			err = j.column(table, name, notNull, jsSchema)
		case jsc.Integer:
			err = j.column(table, name, notNull, jsSchema)
		case jsc.Number:
			err = j.column(table, name, notNull, jsSchema)
		default:
			return fmt.Errorf("unknown json-schema type '%s'", jsSchema.Type)
		}
	}
	return err
}

// NewParser instantiate a new JSON-Schema parser.
func NewParser(logger logr.Logger, schema *Schema) *Parser {
	return &Parser{
		logger: logger.WithName("json-schema-parser"),
		schema: schema,
	}
}
