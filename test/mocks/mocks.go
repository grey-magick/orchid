package mocks

import (
	"encoding/json"
	"math/rand"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	jsc "github.com/isutton/orchid/pkg/orchid/jsonschema"
)

func JSONSchemaProps(
	jsonSchemaType string,
	format string,
	required []string,
	properties map[string]extv1.JSONSchemaProps,
) extv1.JSONSchemaProps {
	return extv1.JSONSchemaProps{
		Type:       jsonSchemaType,
		Format:     format,
		Properties: properties,
		Required:   required,
	}
}

// OpenAPIV3SchemaMock creates a realistic version of openAPIV3Schema entry in Kubernetes CRDs.
func OpenAPIV3SchemaMock() extv1.JSONSchemaProps {
	maxItems := int64(10)
	specProps := map[string]extv1.JSONSchemaProps{
		"simple": jsc.JSONSchemaProps(jsc.String, "", nil, nil, nil),
		"array": jsc.JSONSchemaProps(jsc.Array, "", nil, jsc.JSONSchemaPropsOrArray(
			extv1.JSONSchemaProps{
				Type:     jsc.String,
				Format:   "",
				MaxItems: &maxItems,
			},
		), nil),
		"complex": jsc.JSONSchemaProps("object", "", nil, nil, map[string]extv1.JSONSchemaProps{
			"simple_nested": jsc.JSONSchemaProps("string", "", nil, nil, nil),
			"complex_nested": jsc.JSONSchemaProps(
				"object", "", []string{"attribute"}, nil, map[string]extv1.JSONSchemaProps{
					"attribute": jsc.StringProp,
				}),
		}),
	}
	spec := jsc.JSONSchemaProps(jsc.Object, "", []string{"simple"}, nil, specProps)
	return extv1.JSONSchemaProps{
		Properties: map[string]extv1.JSONSchemaProps{
			"apiVersion": jsc.StringProp,
			"kind":       jsc.StringProp,
			"metadata":   jsc.JSONSchemaProps(jsc.Object, "", nil, nil, nil),
			"spec":       spec,
		},
	}
}

func toUnstructured(obj interface{}) (*unstructured.Unstructured, error) {
	if marshaler, ok := obj.(json.Marshaler); ok {
		b, err := marshaler.MarshalJSON()
		if err != nil {
			return nil, err
		}
		data := map[string]interface{}{}
		err = json.Unmarshal(b, &data)
		return &unstructured.Unstructured{Object: data}, err
	} else {
		data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			return nil, err
		}
		return &unstructured.Unstructured{Object: data}, nil
	}
}

func UnstructuredCRMock(ns, name string) (*unstructured.Unstructured, error) {
	now := metav1.NewTime(time.Now())
	truePtr := true
	u := &unstructured.Unstructured{}

	u.SetUnstructuredContent(map[string]interface{}{
		"spec": map[string]interface{}{
			"simple": "11",
			"array":  []interface{}{"a", "r", "r", "a", "y"},
			"complex": map[string]interface{}{
				"simple_nested": "11",
				"complex_nested": map[string]interface{}{
					"attribute": "string attribute",
				},
			},
		},
	})

	u.SetNamespace(ns)
	u.SetName(name)

	u.SetGroupVersionKind(schema.GroupVersionKind{Group: "mock", Version: "v1", Kind: "Custom"})
	u.SetKind("Custom")
	u.SetAPIVersion("mock/v1")
	u.SetAnnotations(map[string]string{"annotation": "annotation"})
	u.SetClusterName("cluster-name")
	u.SetGenerateName("generated-name")
	u.SetGeneration(1)
	u.SetLabels(map[string]string{"label": "label"})
	u.SetManagedFields([]metav1.ManagedFieldsEntry{
		{
			Manager:    "manager1",
			APIVersion: "manager1/v1",
			Time:       &now,
			Operation:  metav1.ManagedFieldsOperationApply,
		},
		{
			Manager:    "manager2",
			APIVersion: "manager2/v1",
			Time:       &now,
			Operation:  metav1.ManagedFieldsOperationUpdate,
		},
	})
	u.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion:         "owner/v1",
			BlockOwnerDeletion: &truePtr,
			Controller:         &truePtr,
			Kind:               "Owner1",
			Name:               "ownername1",
			UID:                "owner1-uid",
		},
		{
			APIVersion:         "owner/v2",
			BlockOwnerDeletion: &truePtr,
			Controller:         &truePtr,
			Kind:               "Owner2",
			Name:               "ownername2",
			UID:                "owner2-uid",
		},
	})
	u.SetResourceVersion("v1")
	u.SetSelfLink("self-link")
	u.SetUID("uid")
	u.SetCreationTimestamp(now)
	u.SetDeletionTimestamp(&now)
	u.SetFinalizers([]string{"finalizer"})

	return u, nil
}

// UnstructuredReplicaSetMock returns an replica-set which contains nested one-to-many
// relationships.
func UnstructuredReplicaSetMock() (*unstructured.Unstructured, error) {
	rs := &appsv1.ReplicaSet{
		Spec: appsv1.ReplicaSetSpec{
			Selector: &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      "first-key",
						Values:   []string{"value1", "value2"},
						Operator: metav1.LabelSelectorOpDoesNotExist,
					},
					{
						Key:      "second-key",
						Values:   []string{"value1", "value2"},
						Operator: metav1.LabelSelectorOpExists,
					},
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "replicaSet",
					Namespace: "namespace",
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "first-container",
							Image: "image",
							Ports: []v1.ContainerPort{
								{
									Name:          "first-port",
									HostPort:      123,
									ContainerPort: 456,
								},
								{
									Name:          "second-port",
									HostPort:      678,
									ContainerPort: 456,
								},
							},
						},
						{
							Name:  "second-container",
							Image: "image",
							Ports: []v1.ContainerPort{
								{
									Name:          "first-port",
									HostPort:      123,
									ContainerPort: 456,
								},
								{
									Name:          "second-port",
									HostPort:      678,
									ContainerPort: 456,
								},
							},
						},
					},
				},
			},
		},
	}
	return toUnstructured(rs)
}

func UnstructuredCRDMock(ns, name string) (*unstructured.Unstructured, error) {
	crd := CRDMock(ns, name)
	u, err := toUnstructured(crd)
	if err != nil {
		return nil, err
	}

	u.SetNamespace(ns)
	u.SetName(name)

	now := metav1.NewTime(time.Now())
	u.SetManagedFields([]metav1.ManagedFieldsEntry{
		{
			Manager:    "manager1",
			APIVersion: "manager1/v1",
			Time:       &now,
			Operation:  metav1.ManagedFieldsOperationApply,
		},
	})

	u.SetAnnotations(map[string]string{})
	u.SetLabels(map[string]string{})

	truePtr := true
	u.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion:         "owner/v1",
			BlockOwnerDeletion: &truePtr,
			Controller:         &truePtr,
			Kind:               "Owner1",
			Name:               "ownername1",
			UID:                "owner1-uid",
		},
	})

	return u, nil
}

func CRDMock(ns, name string) *extv1.CustomResourceDefinition {
	openAPIV3Schema := OpenAPIV3SchemaMock()
	return &extv1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apiextensions.k8s.io/v1",
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
		Spec: extv1.CustomResourceDefinitionSpec{
			Group: "mock",
			Versions: []extv1.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  false,
					Storage: true,
					Schema: &extv1.CustomResourceValidation{
						OpenAPIV3Schema: &openAPIV3Schema,
					},
					Subresources:             &extv1.CustomResourceSubresources{},
					AdditionalPrinterColumns: nil,
				},
			},
			Names: extv1.CustomResourceDefinitionNames{
				Kind:       "Custom",
				ListKind:   "CustomList",
				Singular:   "custom",
				Plural:     "customs",
				ShortNames: []string{"cst", "csts"},
			},
			Scope: "Namespaced",
		},
		Status: extv1.CustomResourceDefinitionStatus{},
	}
}

// RandomString creates a random string on informed size
func RandomString(length int) string {
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	letter := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]rune, length)

	for i := range b {
		b[i] = letter[seededRand.Intn(len(letter))]
	}
	return string(b)
}
