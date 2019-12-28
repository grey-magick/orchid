package validation

import (
	"errors"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var GVKNotFoundErr = errors.New("gvk not found")
var MultipleVersionsFoundErr = errors.New("multiple versions found where one is allowed")
var SchemaNotFoundErr = errors.New("openAPIV3Schema not found")
var InvalidObjectErr = errors.New("invalid object")

// Validator provides validation for unstructured objects.
type Validator interface {
	Validate(obj *unstructured.Unstructured) error
}

// FIXME: review those changes in the light of recent Repository changes;
/*
// repositoryValidator validates unstructured objects using a repository.
type repositoryValidator struct {
	Repository repository.ObjectRepository
}

// discoverOpenAPIV3Schema returns the JSON Schema properties associated with the given gvk.
func (v *repositoryValidator) discoverOpenAPIV3Schema(gvk schema.GroupVersionKind) (*extv1.JSONSchemaProps, error) {
	crds, err := v.Repository.List(repository.CRDGVK, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// iterate all returned CRDs, returning the JSON Schema properties associated with the first
	// resource definition matching the given gvk
	for _, curCRD := range crds.Items {
		crGVK, err := repository.ExtractCRGVKFromCRD(curCRD.Object)
		if err != nil {
			return nil, err
		}
		if crGVK != gvk {
			continue
		}

		s, err := ExtractOpenAPIV3Schema(curCRD.Object)
		if err == SchemaNotFoundErr {
			continue
		}
		if err != nil {
			return nil, err
		}
		return s, nil
	}
	return nil, GVKNotFoundErr
}

// Validate validates the given obj according to information available in the repository by finding
// the first resource definition matching the object's gvk.
func (v *repositoryValidator) Validate(obj *unstructured.Unstructured) error {
	if obj == nil {
		return errors.New("input is required")
	}
	openAPIV3Schema, err := v.discoverOpenAPIV3Schema(obj.GroupVersionKind())
	if err != nil {
		return err
	}
	in := &extv1.CustomResourceValidation{
		OpenAPIV3Schema: openAPIV3Schema,
	}
	out := &apiextensions.CustomResourceValidation{}
	err = extv1.Convert_v1_CustomResourceValidation_To_apiextensions_CustomResourceValidation(in, out, nil)
	if err != nil {
		return err
	}
	validator, _, err := validation.NewSchemaValidator(out)
	if err != nil {
		return err
	}
	// perform the actual validation returning the first error if any
	r := validator.Validate(obj)
	if len(r.Errors) > 0 {
		return InvalidObjectErr
	}

	return nil
}

// ExtractOpenAPIV3Schema returns the JSON Schema properties contained in obj, assuming u contains
// the required fields determined by CustomResourceDefinition.
//
// It assumes '.spec.versions' to exist, and to contain exactly one entry; its name is currently
// being ignored.
func ExtractOpenAPIV3Schema(obj map[string]interface{}) (*extv1.JSONSchemaProps, error) {
	versions, exists, err := unstructured.NestedSlice(obj, "spec", "versions")
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, SchemaNotFoundErr
	}
	if len(versions) != 1 {
		return nil, MultipleVersionsFoundErr
	}
	version, ok := versions[0].(map[string]interface{})
	if !ok {
		return nil, SchemaNotFoundErr
	}
	schemaMap, exists, err := unstructured.NestedMap(version, "schema", "openAPIV3Schema")
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, SchemaNotFoundErr
	}
	schemaProps := &extv1.JSONSchemaProps{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(schemaMap, schemaProps)
	if err != nil {
		return nil, err
	}
	return schemaProps, nil
}

// NewRepositoryValidator creates a new validator that knows how to obtain JSON Schema properties for
// validation through the given repository.
func NewRepositoryValidator(repository repository.ObjectRepository) Validator {
	return &repositoryValidator{Repository: repository}
}

*/
