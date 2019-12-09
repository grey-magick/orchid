package orm

import "fmt"

type Table struct {
	Name    string
	Columns []*Column
}

type Column struct {
	Name string
	Type string
}

const (
	PgTypeText      = "text"
	PgTypeTextArray = "text[]"
	PgTypeBigInt    = "bigint"
	PgTypeBoolean   = "boolean"
)

func (t *Table) append(column *Column) {
	t.Columns = append(t.Columns, column)
}

func (t *Table) formatToType(format string) string {
	switch format {
	case "int32":
		return "integer"
	case "int64":
		return "bigint"
	case "float":
		return "real"
	case "double":
		return "double precision"
	case "byte":
		return "text"
	case "binary":
		return "text"
	}
	return ""
}

func (t *Table) AddColumn(name, format string) {
	column := &Column{Name: name, Type: t.formatToType(format)}
	t.append(column)
}

func (t *Table) AddColumnRaw(name, rawType string) {
	column := &Column{Name: name, Type: rawType}
	t.append(column)
}

func (t *Table) AddArrayColumn(name, schemaType, format string, max *int64) {
	column := &Column{Name: name, Type: t.formatToType(format)}
	if max != nil {
		column.Type = fmt.Sprintf("%s[%d]", column.Type, max)
	} else {
		column.Type = fmt.Sprintf("%s[]", column.Type)
	}
	t.append(column)
}

func NewTable(name string) *Table {
	return &Table{Name: name}
}
