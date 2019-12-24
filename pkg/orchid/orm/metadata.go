package orm

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

var (
	jsPropString   = jsonSchemaProps(JSTypeString, "", nil, nil, nil)
	jsPropInt64    = jsonSchemaProps(JSTypeInteger, "int64", nil, nil, nil)
	jsPropDateTime = jsonSchemaProps(JSTypeString, "date-time", nil, nil, nil)
	jsPropBoolean  = jsonSchemaProps(JSTypeBoolean, "", nil, nil, nil)
)

// jsonSchemaPropsOrArray creates a JSONSchemaPropsOrArray skeleton based on properties.
func jsonSchemaPropsOrArray(props apiextensionsv1.JSONSchemaProps) *apiextensionsv1.JSONSchemaPropsOrArray {
	return &apiextensionsv1.JSONSchemaPropsOrArray{Schema: &props}
}

// jsonSchemaProps creates a json-schema object skeleton.
func jsonSchemaProps(
	jsType string,
	format string,
	required []string,
	items *apiextensionsv1.JSONSchemaPropsOrArray,
	properties map[string]apiextensionsv1.JSONSchemaProps,
) apiextensionsv1.JSONSchemaProps {
	return apiextensionsv1.JSONSchemaProps{
		Type:       jsType,
		Format:     format,
		Required:   required,
		Items:      items,
		Properties: properties,
	}
}

// objectMetaStringKV creates a key-value entry.
func objectMetaStringKV() apiextensionsv1.JSONSchemaProps {
	return apiextensionsv1.JSONSchemaProps{
		Type:                 JSTypeObject,
		AdditionalProperties: &apiextensionsv1.JSONSchemaPropsOrBool{Schema: &jsPropString},
	}
}

// objectMetaFinalizers
func objectMetaFinalizers() apiextensionsv1.JSONSchemaProps {
	return jsonSchemaProps(JSTypeArray, "", nil, jsonSchemaPropsOrArray(jsPropString), nil)
}

// objectMetaManagedFields defines ObjectMeta.managedFields entry.
func objectMetaManagedFields() apiextensionsv1.JSONSchemaProps {
	items := jsonSchemaPropsOrArray(
		jsonSchemaProps(JSTypeObject, "", nil, nil, map[string]apiextensionsv1.JSONSchemaProps{
			"apiVersion": jsPropString,
			"manager":    jsPropString,
			"operation":  jsPropString,
			"time":       jsPropDateTime,
		}),
	)
	return jsonSchemaProps(JSTypeArray, "", nil, items, nil)
}

// objectMetaOwnerReferences defines ObjectMeta.ownerReferences entry.
func objectMetaOwnerReferences() apiextensionsv1.JSONSchemaProps {
	items := jsonSchemaPropsOrArray(
		jsonSchemaProps(JSTypeObject, "", nil, nil, map[string]apiextensionsv1.JSONSchemaProps{
			"apiVersion":         jsPropString,
			"blockOwnerDeletion": jsPropBoolean,
			"controller":         jsPropBoolean,
			"kind":               jsPropString,
			"name":               jsPropString,
			"uid":                jsPropString,
		}),
	)
	return jsonSchemaProps(JSTypeArray, "", nil, items, nil)
}

// metaV1ObjectMetaOpenAPIV3Schema creates an ObjectMeta object based on metav1.
func metaV1ObjectMetaOpenAPIV3Schema() map[string]apiextensionsv1.JSONSchemaProps {
	return map[string]apiextensionsv1.JSONSchemaProps{
		"annotations":                objectMetaStringKV(),
		"clusterName":                jsPropString,
		"creationTimestamp":          jsPropString,
		"deletionGracePeriodSeconds": jsPropInt64,
		"deletionTimestamp":          jsPropDateTime,
		"finalizers":                 objectMetaFinalizers(),
		"generateName":               jsPropString,
		"generation":                 jsPropInt64,
		"labels":                     objectMetaStringKV(),
		"managedFields":              objectMetaManagedFields(),
		"name":                       jsPropString,
		"namespace":                  jsPropString,
		"ownerReferences":            objectMetaOwnerReferences(),
		"resourceVersion":            jsPropString,
		"selfLink":                   jsPropString,
		"uid":                        jsPropString,
	}
}
