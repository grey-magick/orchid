package jsonschema

import (
	extv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

var (
	StringProp   = JSONSchemaProps(String, "", nil, nil, nil)
	Int64Prop    = JSONSchemaProps(Integer, "int64", nil, nil, nil)
	DateTimeProp = JSONSchemaProps(String, "date-time", nil, nil, nil)
	BooleanProp  = JSONSchemaProps(Boolean, "", nil, nil, nil)
)

// JSONSchemaPropsOrArray creates a JSONSchemaPropsOrArray skeleton based on properties.
func JSONSchemaPropsOrArray(props extv1beta1.JSONSchemaProps) *extv1beta1.JSONSchemaPropsOrArray {
	return &extv1beta1.JSONSchemaPropsOrArray{Schema: &props}
}

// JSONSchemaProps creates a json-schema object skeleton.
func JSONSchemaProps(
	jsType string,
	format string,
	required []string,
	items *extv1beta1.JSONSchemaPropsOrArray,
	properties map[string]extv1beta1.JSONSchemaProps,
) extv1beta1.JSONSchemaProps {
	return extv1beta1.JSONSchemaProps{
		Type:       jsType,
		Format:     format,
		Required:   required,
		Items:      items,
		Properties: properties,
	}
}
