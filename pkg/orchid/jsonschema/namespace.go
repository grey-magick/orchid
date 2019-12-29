package jsonschema

import (
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// namespaceSpec spec attribute of namespace.
func namespaceSpec() extv1.JSONSchemaProps {
	properties := map[string]extv1.JSONSchemaProps{
		"finalizers": objectMetaFinalizers(),
	}
	return extv1.JSONSchemaProps{Type: Object, Properties: properties}
}

// CoreV1NamespaceOpenAPIV3Schema JSON-Schema specification of core/v1 Namespaces.
func CoreV1NamespaceOpenAPIV3Schema() extv1.JSONSchemaProps {
	properties := map[string]extv1.JSONSchemaProps{
		"apiVersion": StringProp,
		"kind":       StringProp,
		"spec":       namespaceSpec(),
	}
	return extv1.JSONSchemaProps{Type: Object, Properties: properties}
}
