package orm

import (
	"strings"

	jsc "github.com/isutton/orchid/pkg/orchid/jsonschema"
)

// CRD represents the group of tables needed to store CRDs.
type CRD struct {
	schema *Schema // schema instance
}

const CRDRawColumn = "data"

// crdTable create a special table to store CRDs.
func (c *CRD) crdTable() {
	tableName := strings.ToLower(c.schema.Name)
	table := c.schema.TableFactory(tableName, false)
	table.AddSerialPK()

	table.AddColumn(
		&Column{Name: "apiVersion", Type: PgTypeText, JSType: jsc.String, NotNull: true})
	table.AddColumn(
		&Column{Name: "kind", Type: PgTypeText, JSType: jsc.String, NotNull: true})
	table.AddColumn(
		&Column{Name: CRDRawColumn, Type: PgTypeJSONB, JSType: jsc.String, NotNull: true})

	metadataTableName := c.schema.TableName("metadata")
	metadataTable := c.schema.TableFactory(metadataTableName, true)
	metadataTable.Path = []string{"metadata"}
	metadataTable.AddSerialPK()
	metadataTable.AddBigIntFK(tableName, tableName, PKColumnName, false)
	metadataTable.AddColumn(&Column{
		Name:         "name",
		Type:         PgTypeText,
		OriginalType: JSTypeString,
		NotNull:      true,
	})
	metadataTable.AddColumn(&Column{
		Name:         "namespace",
		Type:         PgTypeText,
		OriginalType: JSTypeString,
		NotNull:      false,
	})
}

// Add tables belonging to CRD schema.
func (c *CRD) Add() {
	c.crdTable()
}

// NewCRD instantiate CRD.
func NewCRD(schema *Schema) *CRD {
	return &CRD{schema: schema}
}
