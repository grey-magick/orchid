package mocks

import (
	extv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

func JSONSchemaProps(
	jsonSchemaType string,
	format string,
	required []string,
	properties map[string]extv1beta1.JSONSchemaProps,
) extv1beta1.JSONSchemaProps {
	return extv1beta1.JSONSchemaProps{
		Type:       jsonSchemaType,
		Format:     format,
		Properties: properties,
		Required:   required,
	}
}

// OpenAPIV3SchemaMock creates a realistic version of openAPIV3Schema entry in Kubernetes CRDs.
func OpenAPIV3SchemaMock() extv1beta1.JSONSchemaProps {
	return extv1beta1.JSONSchemaProps{
		Properties: map[string]extv1beta1.JSONSchemaProps{
			"apiVersion": JSONSchemaProps("string", "", nil, nil),
			"kind":       JSONSchemaProps("string", "", nil, nil),
			"metadata":   JSONSchemaProps("object", "", nil, nil),
			"spec": JSONSchemaProps("object", "", []string{"complex"}, map[string]extv1beta1.JSONSchemaProps{
				"simple": JSONSchemaProps("integer", "int32", nil, nil),
				"complex": JSONSchemaProps("object", "", nil, map[string]extv1beta1.JSONSchemaProps{
					"simple_nested": JSONSchemaProps("integer", "int32", nil, nil),
					"complex_nested": JSONSchemaProps("object", "", []string{"attribute"}, map[string]extv1beta1.JSONSchemaProps{
						"attribute": JSONSchemaProps("string", "byte", nil, nil),
					}),
				}),
			}),
			"status": JSONSchemaProps("string", "", nil, nil),
		},
	}
}
