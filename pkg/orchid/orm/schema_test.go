package orm

import (
	"testing"

	extv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

func TestSchema_New(t *testing.T) {
	properties := map[string]extv1beta1.JSONSchemaProps{
		"simple": {
			Type:   "integer",
			Format: "int32",
		},
		"complex": {
			Type: "object",
			Properties: map[string]extv1beta1.JSONSchemaProps{
				"attribute": {
					Type:   "string",
					Format: "byte",
				},
				"complex_attribute": {
					Type: "object",
					Properties: map[string]extv1beta1.JSONSchemaProps{
						"inner_attribute": {
							Type:   "string",
							Format: "byte",
						},
					},
				},
			},
		},
	}

	schema := NewSchema("cr")

	t.Run("Generate", func(t *testing.T) {
		err := schema.Generate(properties)
		t.Logf("err='%#v'", err)
	})
}
