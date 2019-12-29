package jsonschema

import (
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func OrchidOpenAPIV3Schema() extv1.JSONSchemaProps {
	properties := map[string]extv1.JSONSchemaProps{
		"apiVersion": StringProp,
		"kind":       StringProp,
		"metadata": {
			Type: Object,
			Properties: map[string]extv1.JSONSchemaProps{
				"labels": {
					Type: Object,
					AdditionalProperties: &extv1.JSONSchemaPropsOrBool{
						Schema: &StringProp,
					},
				},
				"annotations": {
					Type: Object,
					AdditionalProperties: &extv1.JSONSchemaPropsOrBool{
						Schema: &StringProp,
					},
				},
			},
		},
	}
	return extv1.JSONSchemaProps{
		Type:              Object,
		Properties:        properties,
		Required:          []string{"apiVersion", "kind"},
		XEmbeddedResource: true,
	}
}
