package orm

import (
	jsc "github.com/isutton/orchid/pkg/orchid/jsonschema"
)

// CRD represents the group of tables needed to store CRDs.
type CRD struct {
	schema *Schema // schema instance
}

const CRDRawColumn = "data"

// crdTable create a special table to store CRDs.
func (c *CRD) crdTable() {
	table := c.schema.TableFactory(c.schema.TableName("crd"), false)
	table.AddSerialPK()

	table.AddColumn(
		&Column{Name: "apiVersion", Type: PgTypeText, JSType: jsc.String, NotNull: true})
	table.AddColumn(
		&Column{Name: "kind", Type: PgTypeText, JSType: jsc.String, NotNull: true})
	table.AddColumn(
		&Column{Name: CRDRawColumn, Type: PgTypeJSONB, JSType: jsc.String, NotNull: true})
}

// Add tables belonging to CRD schema.
func (c *CRD) Add() {
	c.crdTable()
}

// NewCRD instantiate CRD.
func NewCRD(schema *Schema) *CRD {
	return &CRD{schema: schema}
}
