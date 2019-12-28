package jsonschema

import (
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func OrchidOpenAPIV3Schema() extv1.JSONSchemaProps {
	properties := map[string]extv1.JSONSchemaProps{
		"apiVersion": StringProp,
		"kind":       StringProp,
	}
	return extv1.JSONSchemaProps{
		Type:              Object,
		Properties:        properties,
		Required:          []string{"apiVersion", "kind"},
		XEmbeddedResource: true,
	}
}
