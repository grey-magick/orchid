package orm

import (
	extv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	jsc "github.com/isutton/orchid/pkg/orchid/jsonschema"
)

// objectMetaStringKV creates a key-value entry.
func objectMetaStringKV() extv1beta1.JSONSchemaProps {
	return extv1beta1.JSONSchemaProps{
		Type:                 jsc.Object,
		AdditionalProperties: &extv1beta1.JSONSchemaPropsOrBool{Schema: &jsc.StringProp},
	}
}

// objectMetaFinalizers
func objectMetaFinalizers() extv1beta1.JSONSchemaProps {
	return jsc.JSONSchemaProps(jsc.Array, "", nil, jsc.JSONSchemaPropsOrArray(jsc.StringProp), nil)
}

// objectMetaManagedFields defines ObjectMeta.managedFields entry.
func objectMetaManagedFields() extv1beta1.JSONSchemaProps {
	items := jsc.JSONSchemaPropsOrArray(
		jsc.JSONSchemaProps(jsc.Object, "", nil, nil, map[string]extv1beta1.JSONSchemaProps{
			"apiVersion": jsc.StringProp,
			"manager":    jsc.StringProp,
			"operation":  jsc.StringProp,
			"time":       jsc.DateTimeProp,
		}),
	)
	return jsc.JSONSchemaProps(jsc.Array, "", nil, items, nil)
}

// objectMetaOwnerReferences defines ObjectMeta.ownerReferences entry.
func objectMetaOwnerReferences() extv1beta1.JSONSchemaProps {
	items := jsc.JSONSchemaPropsOrArray(
		jsc.JSONSchemaProps(jsc.Object, "", nil, nil, map[string]extv1beta1.JSONSchemaProps{
			"apiVersion":         jsc.StringProp,
			"blockOwnerDeletion": jsc.BooleanProp,
			"controller":         jsc.BooleanProp,
			"kind":               jsc.StringProp,
			"name":               jsc.StringProp,
			"uid":                jsc.StringProp,
		}),
	)
	return jsc.JSONSchemaProps(jsc.Array, "", nil, items, nil)
}

// metaV1ObjectMetaOpenAPIV3Schema creates an ObjectMeta object based on metav1.
func metaV1ObjectMetaOpenAPIV3Schema() map[string]extv1beta1.JSONSchemaProps {
	return map[string]extv1beta1.JSONSchemaProps{
		"annotations":                objectMetaStringKV(),
		"clusterName":                jsc.StringProp,
		"creationTimestamp":          jsc.StringProp,
		"deletionGracePeriodSeconds": jsc.Int64Prop,
		"deletionTimestamp":          jsc.DateTimeProp,
		"finalizers":                 objectMetaFinalizers(),
		"generateName":               jsc.StringProp,
		"generation":                 jsc.Int64Prop,
		"labels":                     objectMetaStringKV(),
		"managedFields":              objectMetaManagedFields(),
		"name":                       jsc.StringProp,
		"namespace":                  jsc.StringProp,
		"ownerReferences":            objectMetaOwnerReferences(),
		"resourceVersion":            jsc.StringProp,
		"selfLink":                   jsc.StringProp,
		"uid":                        jsc.StringProp,
	}
}
