package orm

import (
	"fmt"
	"strings"
)

// Table represents database table with columns and constraints.
type Table struct {
	Name        string        // table name
	Columns     []*Column     // table columns
	Constraints []*Constraint // constraints
}

const (
	PgTypeText         = "text"
	PgTypeTextArray    = "text[]"
	PgTypeBoolean      = "boolean"
	PgTypeInt          = "integer"
	PgTypeBigInt       = "bigint"
	PgTypeReal         = "real"
	PgTypeDouble       = "double precision"
	PgTypeSerial8      = "serial8"
	PgConstraintPK     = "primary key"
	PgConstraintFK     = "foreign key"
	PgConstraintUnique = "unique"
)

func (t *Table) appendColumn(column *Column) {
	t.Columns = append(t.Columns, column)
}

// String print out table creation statement.
func (t *Table) String() string {
	columns := []string{}
	for _, column := range t.Columns {
		columns = append(columns, column.String())
	}
	constrains := []string{}
	for _, constraint := range t.Constraints {
		constrains = append(constrains, constraint.String())
	}
	return fmt.Sprintf("create table if not exists %s (%s, %s)",
		t.Name, strings.Join(columns, ", "), strings.Join(constrains, ", "))
}

// AddColumnRaw add a new column using informed type directly.
func (t *Table) AddColumnRaw(name, rawType string) {
	column := &Column{Name: name, Type: rawType}
	t.appendColumn(column)
}

// AddConstraintRaw add a new constraint based on type and expression.
func (t *Table) AddConstraintRaw(constraintType, expr string) {
	constraint := &Constraint{Type: constraintType, Expr: expr}
	t.Constraints = append(t.Constraints, constraint)
}

// AddSerialPK add a new column as primary-key, using serial8 type.
func (t *Table) AddSerialPK() {
	t.AddColumnRaw("id", PgTypeSerial8)
	t.AddConstraintRaw(PgConstraintPK, "(id)")
}

// AddBigIntPK add a new column as a primary-key, using BigInt type.
func (t *Table) AddBigIntPK() {
	t.AddColumnRaw("id", PgTypeBigInt)
	t.AddConstraintRaw(PgConstraintPK, "(id)")
}

// jsonSchemaFormatToPg based on json-schema format, return database type.
func (t *Table) jsonSchemaFormatToPg(format string) string {
	switch format {
	case "int32":
		return PgTypeInt
	case "int64":
		return PgTypeBigInt
	case "float":
		return PgTypeReal
	case "double":
		return PgTypeDouble
	case "byte":
		return PgTypeText
	case "binary":
		return PgTypeText
	}
	return ""
}

// jsonSchemaTypeToPg based on json-schema type, return default database type for it.
func (t *Table) jsonSchemaTypeToPg(jsonSchemaType string) string {
	switch jsonSchemaType {
	case "integer":
		return PgTypeInt
	case "number":
		return PgTypeReal
	case "string":
		return PgTypeText
	case "boolean":
		return PgTypeBoolean
	}
	return ""
}

// pgColumnType based in json-schema type and format, return database column type.
func (t *Table) pgColumnType(jsonSchemaType string, format string) (string, error) {
	if jsonSchemaType == "" && format == "" {
		return "", fmt.Errorf("both type and format are not informed")
	}

	var pgType string
	if format != "" {
		pgType = t.jsonSchemaFormatToPg(format)
	} else {
		pgType = t.jsonSchemaTypeToPg(jsonSchemaType)
	}

	if pgType == "" {
		return "", fmt.Errorf(
			"can't determine column based on type='%s' format='%s'", jsonSchemaType, format)
	}
	return pgType, nil
}

// AddColumn adds a regular column using json-schema type and format to decide database type.
func (t *Table) AddColumn(name, jsonSchemaType, format string) error {
	column := &Column{Name: name}
	var err error
	if column.Type, err = t.pgColumnType(jsonSchemaType, format); err != nil {
		return err
	}
	t.appendColumn(column)
	return nil
}

// AddArrayColumn adds an array column, using schema-type and format to decide database column type,
// and in case of having max items, the array will be capped as well.
func (t *Table) AddArrayColumn(name, jsonSchemaType, format string, max *int64) error {
	column := &Column{Name: name}
	var err error
	if column.Type, err = t.pgColumnType(jsonSchemaType, format); err != nil {
		return err
	}
	if max != nil {
		column.Type = fmt.Sprintf("%s[%d]", column.Type, max)
	} else {
		column.Type = fmt.Sprintf("%s[]", column.Type)
	}
	t.appendColumn(column)
	return nil
}

// NewTable instantiate a new Table.
func NewTable(name string) *Table {
	return &Table{Name: name}
}
