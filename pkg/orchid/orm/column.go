package orm

import (
	"database/sql"
	"fmt"
)

// Column represents columns.
type Column struct {
	Name         string // column name
	Type         string // column type
	OriginalType string // hint with original column type
	NotNull      bool   // not null flag
}

// String print out column and type.
func (c *Column) String() string {
	var notNull string
	if c.NotNull {
		notNull = "not null"
	}
	return fmt.Sprintf("%s %s %s", c.Name, c.Type, notNull)
}

func (c *Column) Null() (interface{}, error) {
	switch c.Type {
	case PgTypeBigInt:
		return sql.NullInt64{}, nil
	case PgTypeBoolean:
		return sql.NullBool{}, nil
	case PgTypeDouble:
		return sql.NullFloat64{}, nil
	case PgTypeInt:
		return sql.NullInt32{}, nil
	case PgTypeJSONB:
		return sql.NullString{}, nil
	case PgTypeReal:
		return sql.NullFloat64{}, nil
	}
	return nil, fmt.Errorf("unable to create a null presentation for type '%s'", c.Type)
}

// NewColumn instantiate a new column using type and format.
func NewColumn(name, jsonSchemaType, format string, notNull bool) (*Column, error) {
	columnType, err := ColumnTypeParser(jsonSchemaType, format)
	if err != nil {
		return nil, err
	}
	return &Column{Name: name, Type: columnType, OriginalType: jsonSchemaType, NotNull: notNull}, nil
}

// NewColumnArray instantiate a new array column using type, format and max items.
func NewColumnArray(name, jsonSchemaType, format string, max *int64, notNull bool) (*Column, error) {
	column, err := NewColumn(name, jsonSchemaType, format, notNull)
	if err != nil {
		return nil, err
	}
	if max != nil {
		column.Type = fmt.Sprintf("%s[%d]", column.Type, *max)
	} else {
		column.Type = fmt.Sprintf("%s[]", column.Type)
	}
	return column, nil
}
