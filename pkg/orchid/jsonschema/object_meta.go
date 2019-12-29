package jsonschema

import (
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// objectMetaStringKV creates a key-value entry.
func objectMetaStringKV() extv1.JSONSchemaProps {
	return extv1.JSONSchemaProps{
		Type:                 Object,
		AdditionalProperties: &extv1.JSONSchemaPropsOrBool{Schema: &StringProp},
	}
}

// objectMetaFinalizers
func objectMetaFinalizers() extv1.JSONSchemaProps {
	return JSONSchemaProps(Array, "", nil, JSONSchemaPropsOrArray(StringProp), nil)
}

// objectMetaManagedFields defines ObjectMeta.managedFields entry.
func objectMetaManagedFields() extv1.JSONSchemaProps {
	items := JSONSchemaPropsOrArray(
		JSONSchemaProps(Object, "", nil, nil, map[string]extv1.JSONSchemaProps{
			"apiVersion": StringProp,
			"manager":    StringProp,
			"operation":  StringProp,
			"time":       DateTimeProp,
		}),
	)
	return JSONSchemaProps(Array, "", nil, items, nil)
}

// objectMetaOwnerReferences defines ObjectMeta.ownerReferences entry.
func objectMetaOwnerReferences() extv1.JSONSchemaProps {
	items := JSONSchemaPropsOrArray(
		JSONSchemaProps(Object, "", nil, nil, map[string]extv1.JSONSchemaProps{
			"apiVersion":         StringProp,
			"blockOwnerDeletion": BooleanProp,
			"controller":         BooleanProp,
			"kind":               StringProp,
			"name":               StringProp,
			"uid":                StringProp,
		}),
	)
	return JSONSchemaProps(Array, "", nil, items, nil)
}

// MetaV1ObjectMetaOpenAPIV3Schema creates an ObjectMeta object based on metav1.
func MetaV1ObjectMetaOpenAPIV3Schema() extv1.JSONSchemaProps {
	properties := map[string]extv1.JSONSchemaProps{
		"name":                       StringProp,
		"namespace":                  StringProp,
		"annotations":                objectMetaStringKV(),
		"clusterName":                StringProp,
		"creationTimestamp":          StringProp,
		"deletionGracePeriodSeconds": Int64Prop,
		"deletionTimestamp":          DateTimeProp,
		"finalizers":                 objectMetaFinalizers(),
		"generateName":               StringProp,
		"generation":                 Int64Prop,
		"labels":                     objectMetaStringKV(),
		"managedFields":              objectMetaManagedFields(),
		"ownerReferences":            objectMetaOwnerReferences(),
		"resourceVersion":            StringProp,
		"selfLink":                   StringProp,
		"uid":                        StringProp,
	}
	keys := []string{"namespace", "name"}
	return extv1.JSONSchemaProps{
		Type:       Object,
		Properties: properties,
		Required:   keys,
		// trigger using keys as unique columns in table
		XListMapKeys: keys,
	}
}
