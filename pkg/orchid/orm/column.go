package orm

import (
	"fmt"
)

// Column represents columns.
type Column struct {
	Name    string // column name
	Type    string // column type
	NotNull bool   // not null flag
	Path    string // path within the table
}

// String print out column and type.
func (c *Column) String() string {
	var notNull string
	if c.NotNull {
		notNull = "not null"
	}
	return fmt.Sprintf("%s %s %s", c.Name, c.Type, notNull)
}

// NewColumn instantiate a new column using type and format.
func NewColumn(name, jsonSchemaType, format string, notNull bool) (*Column, error) {
	columnType, err := ColumnTypeParser(jsonSchemaType, format)
	if err != nil {
		return nil, err
	}
	return &Column{Name: name, Type: columnType, NotNull: notNull}, nil
}

// NewColumnArray instantiate a new array column using type, format and max items.
func NewColumnArray(name, jsonSchemaType, format string, max *int64, notNull bool) (*Column, error) {
	columnType, err := ColumnTypeParser(jsonSchemaType, format)
	if err != nil {
		return nil, err
	}
	if max != nil {
		columnType = fmt.Sprintf("%s[%d]", columnType, *max)
	} else {
		columnType = fmt.Sprintf("%s[]", columnType)
	}
	return &Column{Name: name, Type: columnType, NotNull: notNull}, nil
}
