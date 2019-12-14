package orm

// CRD represents the group of tables needed to store CRDs.
type CRD struct {
	schema *Schema // schema instance
}

const CRDRawDataColumn = "data"

// crdTable create a special table to store CRDs.
func (c *CRD) crdTable() {
	table := c.schema.TableFactory(c.schema.TableName("crd"), []string{})
	table.AddSerialPK()

	table.AddColumn(&Column{
		Name:         "apiVersion",
		Type:         PgTypeText,
		OriginalType: JSTypeString,
		NotNull:      true,
	})
	table.AddColumn(&Column{
		Name:         "kind",
		Type:         PgTypeText,
		OriginalType: JSTypeString,
		NotNull:      true,
	})
	table.AddColumn(&Column{
		Name:         CRDRawDataColumn,
		Type:         PgTypeJSONB,
		OriginalType: JSTypeString,
		NotNull:      true,
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
