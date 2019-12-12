package orm

import (
	"fmt"
)

const (
	PgTypeText         = "text"
	PgTypeTextArray    = "text[]"
	PgTypeBoolean      = "boolean"
	PgTypeInt          = "integer"
	PgTypeBigInt       = "bigint"
	PgTypeReal         = "real"
	PgTypeDouble       = "double precision"
	PgTypeSerial8      = "serial8"
	PgTypeJSONB        = "jsonb"
	PgConstraintPK     = "primary key"
	PgConstraintFK     = "foreign key"
	PgConstraintUnique = "unique"
)

// jsonSchemaFormatToPg based on json-schema format, return database type.
func jsonSchemaFormatToPg(format string) string {
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
func jsonSchemaTypeToPg(jsonSchemaType string) string {
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

// ColumnTypeParser based in json-schema type and format, return database column type.
func ColumnTypeParser(jsonSchemaType string, format string) (string, error) {
	if jsonSchemaType == "" && format == "" {
		return "", fmt.Errorf("both type and format are not informed")
	}
	var pgType string
	if format != "" {
		pgType = jsonSchemaFormatToPg(format)
	} else {
		pgType = jsonSchemaTypeToPg(jsonSchemaType)
	}
	if pgType == "" {
		return "", fmt.Errorf(
			"can't determine column based on type='%s' format='%s'", jsonSchemaType, format)
	}
	return pgType, nil
}
