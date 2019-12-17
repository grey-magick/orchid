package orm

import (
	extv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

var (
	jsPropString   = jsonSchemaProps(JSTypeString, "", nil, nil, nil)
	jsPropInt64    = jsonSchemaProps(JSTypeInteger, "int64", nil, nil, nil)
	jsPropDateTime = jsonSchemaProps(JSTypeString, "date-time", nil, nil, nil)
	jsPropBoolean  = jsonSchemaProps(JSTypeBoolean, "", nil, nil, nil)
)

// jsonSchemaPropsOrArray creates a JSONSchemaPropsOrArray skeleton based on properties.
func jsonSchemaPropsOrArray(props extv1beta1.JSONSchemaProps) *extv1beta1.JSONSchemaPropsOrArray {
	return &extv1beta1.JSONSchemaPropsOrArray{Schema: &props}
}

// jsonSchemaProps creates a json-schema object skeleton.
func jsonSchemaProps(
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

// objectMetaStringKV creates a key-value entry.
func objectMetaStringKV() extv1beta1.JSONSchemaProps {
	return extv1beta1.JSONSchemaProps{
		Type:                 JSTypeObject,
		AdditionalProperties: &extv1beta1.JSONSchemaPropsOrBool{Schema: &jsPropString},
	}
}

// objectMetaFinalizers
func objectMetaFinalizers() extv1beta1.JSONSchemaProps {
	return jsonSchemaProps(JSTypeArray, "", nil, jsonSchemaPropsOrArray(jsPropString), nil)
}

// // objectMetaFields io.k8s.apimachinery.pkg.apis.meta.v1.Fields
// func objectMetaFields() extv1beta1.JSONSchemaProps {
// 	properties := map[string]extv1beta1.JSONSchemaProps{
// 		"groupVersion": jsPropString,
// 		"version":      jsPropString,
// 	}
// 	return jsonSchemaProps(JSTypeObject, "", []string{"groupVersion", "version"}, nil, properties)
// }

// objectMetaManagedFields defines ObjectMeta.managedFields entry.
func objectMetaManagedFields() extv1beta1.JSONSchemaProps {
	items := jsonSchemaPropsOrArray(
		jsonSchemaProps(JSTypeObject, "", nil, nil, map[string]extv1beta1.JSONSchemaProps{
			"apiVersion": jsPropString,
			"manager":    jsPropString,
			"operation":  jsPropString,
			"time":       jsPropDateTime,
			// FIXME: how to support fieldsV1?
			// "fields":     objectMetaFields(),
		}),
	)
	return jsonSchemaProps(JSTypeArray, "", nil, items, nil)
}

// objectMetaOwnerReferences defines ObjectMeta.ownerReferences entry.
func objectMetaOwnerReferences() extv1beta1.JSONSchemaProps {
	items := jsonSchemaPropsOrArray(
		jsonSchemaProps(JSTypeObject, "", nil, nil, map[string]extv1beta1.JSONSchemaProps{
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
func metaV1ObjectMetaOpenAPIV3Schema() map[string]extv1beta1.JSONSchemaProps {
	return map[string]extv1beta1.JSONSchemaProps{
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
