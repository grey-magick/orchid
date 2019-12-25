package jsonschema

import (
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

var (
	StringProp   = JSONSchemaProps(String, "", nil, nil, nil)
	Int64Prop    = JSONSchemaProps(Integer, "int64", nil, nil, nil)
	DateTimeProp = JSONSchemaProps(String, "date-time", nil, nil, nil)
	BooleanProp  = JSONSchemaProps(Boolean, "", nil, nil, nil)
)

// JSONSchemaPropsOrArray creates a JSONSchemaPropsOrArray skeleton based on properties.
func JSONSchemaPropsOrArray(props extv1.JSONSchemaProps) *extv1.JSONSchemaPropsOrArray {
	return &extv1.JSONSchemaPropsOrArray{Schema: &props}
}

// JSONSchemaProps creates a json-schema object skeleton.
func JSONSchemaProps(
	jsType string,
	format string,
	required []string,
	items *extv1.JSONSchemaPropsOrArray,
	properties map[string]extv1.JSONSchemaProps,
) extv1.JSONSchemaProps {
	return extv1.JSONSchemaProps{
		Type:       jsType,
		Format:     format,
		Required:   required,
		Items:      items,
		Properties: properties,
	}
}
