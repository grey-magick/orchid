package mocks

import (
	"time"

	extv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
			"spec": JSONSchemaProps("object", "", []string{"simple"}, map[string]extv1beta1.JSONSchemaProps{
				"simple": JSONSchemaProps("string", "", nil, nil),
				"complex": JSONSchemaProps("object", "", nil, map[string]extv1beta1.JSONSchemaProps{
					"simple_nested": JSONSchemaProps("string", "", nil, nil),
					"complex_nested": JSONSchemaProps("object", "", []string{"attribute"}, map[string]extv1beta1.JSONSchemaProps{
						"attribute": JSONSchemaProps("string", "", nil, nil),
					}),
				}),
			}),
			// "status": JSONSchemaProps("string", "", nil, nil),
		},
	}
}

func UnstructuredCRMock() (*unstructured.Unstructured, error) {
	u := &unstructured.Unstructured{}

	u.SetUnstructuredContent(map[string]interface{}{
		"spec": map[string]interface{}{
			"simple": "11",
			"complex": map[string]interface{}{
				"simple_nested": "11",
				"complex_nested": map[string]interface{}{
					"attribute": "string attribute",
				},
			},
		},
	})
	u.SetGroupVersionKind(schema.GroupVersionKind{Group: "mock", Version: "v1", Kind: "Custom"})
	u.SetKind("Custom")
	u.SetAPIVersion("mock/v1")
	u.SetName("testing")
	u.SetAnnotations(map[string]string{"annotation": "annotation"})
	u.SetClusterName("cluster-name")
	u.SetGenerateName("generated-name")
	u.SetGeneration(1)
	u.SetLabels(map[string]string{"label": "label"})
	u.SetManagedFields([]metav1.ManagedFieldsEntry{})
	u.SetNamespace("namespace")
	u.SetOwnerReferences([]metav1.OwnerReference{
		{APIVersion: "owner/v1"},
		{APIVersion: "owner2/v1"},
	})
	u.SetResourceVersion("v1")
	u.SetSelfLink("self-link")
	u.SetUID("uid")
	now := metav1.NewTime(time.Now())
	u.SetCreationTimestamp(now)
	u.SetDeletionTimestamp(nil)
	u.SetFinalizers([]string{"finalizer"})

	return u, nil
}

func toUnstructured(obj interface{}) (*unstructured.Unstructured, error) {
	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}
	return &unstructured.Unstructured{Object: data}, nil
}

func UnstructuredCRDMock() (*unstructured.Unstructured, error) {
	crd := CRDMock()
	return toUnstructured(&crd)
}

func CRDMock() extv1beta1.CustomResourceDefinition {
	openAPIV3Schema := OpenAPIV3SchemaMock()
	return extv1beta1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apiextensions.k8s.io/v1",
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{},
		Spec: extv1beta1.CustomResourceDefinitionSpec{
			Group:   "mock",
			Version: "v1",
			Names: extv1beta1.CustomResourceDefinitionNames{
				Kind:       "Custom",
				ListKind:   "CustomList",
				Singular:   "custom",
				Plural:     "customs",
				ShortNames: []string{"cst", "csts"},
			},
			Scope:        "Namespaced",
			Subresources: &extv1beta1.CustomResourceSubresources{},
			Validation: &extv1beta1.CustomResourceValidation{
				OpenAPIV3Schema: &openAPIV3Schema,
			},
		},
		Status: extv1beta1.CustomResourceDefinitionStatus{},
	}
}
