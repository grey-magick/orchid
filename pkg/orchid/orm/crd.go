package orm

type CRD struct {
	schema *Schema // schema instance
}

// crdTable create a special table to store CRDs.
func (c *CRD) crdTable() {
	table := c.schema.TableFactory(c.schema.TableName("crd"))
	table.AddSerialPK()

	table.AddColumn(&Column{Name: "api_version", Type: PgTypeText})
	table.AddColumn(&Column{Name: "kind", Type: PgTypeText})
	table.AddColumn(&Column{Name: "data", Type: PgTypeJSONB})
}

// Add tables belonging to CRD schema.
func (c *CRD) Add() {
	c.crdTable()
}

func NewCRD(schema *Schema) *CRD {
	return &CRD{schema: schema}
}
