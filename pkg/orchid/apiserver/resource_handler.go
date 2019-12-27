package apiserver

import (
	"errors"

	"github.com/ghodss/yaml"
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"github.com/isutton/orchid/pkg/orchid/repository"
	orchid "github.com/isutton/orchid/pkg/orchid/runtime"
	"github.com/isutton/orchid/pkg/orchid/validation"
)

var (
	exampleAPIResource = metav1.APIResource{
		Group:        "examples.orchid.io",
		Kind:         "Example",
		Name:         "examples",
		ShortNames:   []string{"ex"},
		SingularName: "example",
		Verbs:        []string{"get", "create"},
		Version:      "v1alpha1",
	}

	examplesGroup        = exampleAPIResource.Group
	examplesVersion      = exampleAPIResource.Version
	examplesGroupVersion = examplesGroup + "/" + examplesVersion

	// TODO: this definition should be transformed into a JsonSchema at some point, and automatically
	//       added to the engine at startup time
	crdAPIResource = metav1.APIResource{
		Group:        "apiextensions.k8s.io",
		Kind:         "CustomResourceDefinition",
		Name:         "customresourcedefinitions",
		ShortNames:   []string{"crd"},
		SingularName: "customresourcedefinition",
		Verbs:        []string{"get", "create"},
		Version:      "v1",
	}

	crdGroup        = crdAPIResource.Group
	crdVersion      = crdAPIResource.Version
	crdGroupVersion = crdGroup + "/" + crdVersion
)

// APIResourceHandler is responsible for responding API resource requests.
type APIResourceHandler struct {
	Validator  validation.Validator
	Logger     logr.Logger
	Repository repository.ObjectRepository
}

// ObjectLister returns a list of objects.
func (h *APIResourceHandler) ObjectLister(vars Vars, body []byte) (k8sruntime.Object, error) {
	apiVersion, err := vars.GetAPIVersion()
	if err != nil {
		return nil, err
	}
	m := make(map[string]interface{})
	u := &unstructured.Unstructured{Object: m}
	u.SetAPIVersion(apiVersion)
	u.SetKind(exampleAPIResource.Kind + "List")
	items := make([]interface{}, 0)
	_ = unstructured.SetNestedSlice(m, items, "items")
	return u, nil
}

// APIResourceLister lists API resources.
func (h *APIResourceHandler) APIResourceLister(vars Vars, body []byte) (k8sruntime.Object, error) {
	return &metav1.APIResourceList{
		// TODO: add GroupVersion argument
		GroupVersion: examplesGroupVersion,
		APIResources: []metav1.APIResource{
			exampleAPIResource,
			crdAPIResource,
		},
	}, nil
}

// APIGroupLister lists API groups.
func (h *APIResourceHandler) APIGroupLister(vars Vars, body []byte) (k8sruntime.Object, error) {
	crdAPIGroups := h.CRDAPIGroups()
	groups := []metav1.APIGroup{
		{
			Name: examplesGroup,
			PreferredVersion: metav1.GroupVersionForDiscovery{
				GroupVersion: examplesGroupVersion,
				Version:      examplesVersion,
			},
			Versions: []metav1.GroupVersionForDiscovery{
				{
					GroupVersion: examplesGroupVersion,
					Version:      examplesVersion,
				},
			},
		},
		{
			Name: crdGroup,
			PreferredVersion: metav1.GroupVersionForDiscovery{
				GroupVersion: crdGroupVersion,
				Version:      crdVersion,
			},
			Versions: []metav1.GroupVersionForDiscovery{
				{
					GroupVersion: crdGroupVersion,
					Version:      crdVersion,
				},
			},
		},
	}
	return &metav1.APIGroupList{
		Groups: append(groups, crdAPIGroups...),
	}, nil
}

func (h *APIResourceHandler) OpenAPIHandler(vars Vars, body []byte) (k8sruntime.Object, error) {
	return nil, nil
}

var BodyEmptyErr = errors.New("body is empty")

// ResourcePostHandler handles the create resource action.
func (h *APIResourceHandler) ResourcePostHandler(vars Vars, body []byte) (k8sruntime.Object, error) {
	// do not proceed if body is empty
	if len(body) == 0 {
		return nil, BodyEmptyErr
	}

	// deserialize the body to an Orchid object to validate the object
	obj := orchid.NewObject()
	err := yaml.Unmarshal(body, obj)
	if err != nil {
		return nil, err
	}

	// deserialize the body to a map[string]interface{} as well, since we'll be using it to feed the
	// Repository to create the resource
	uObj := &map[string]interface{}{}
	err = yaml.Unmarshal(body, uObj)
	if err != nil {
		return nil, err
	}
	u := &unstructured.Unstructured{Object: *uObj}

	// validate body against its schema
	err = h.Validator.Validate(u)
	if err != nil {
		return nil, err
	}

	err = h.Repository.Create(u)
	if err != nil {
		return nil, err
	}
	name := types.NamespacedName{
		Namespace: u.GetNamespace(),
		Name:      u.GetName(),
	}
	createdObj, err := h.Repository.Read(u.GroupVersionKind(), name)
	if err != nil {
		return nil, err
	}
	return createdObj, err
}

// Register adds the handler routes in the router.
func (h *APIResourceHandler) Register(router *mux.Router) {
	// create a resource
	// used by kubectl to create a resource
	// should support same serializations kubectl does
	router.HandleFunc("/apis/{group}/{version}/{resource}", Adapt(h.ResourcePostHandler)).
		Methods("POST")

	// used by kubectl to list objects of a particular resource
	// TODO: investigate create a Handler specialized in resource entities
	router.HandleFunc("/apis/{group}/{version}/{resource}", Adapt(h.ObjectLister)).
		Methods("GET")
	// used by kubectl to discover all the resources for an API Group
	router.HandleFunc("/apis/{group}/{version}", Adapt(h.APIResourceLister)).
		Methods("GET")
	// used by kubectl to discover available API Groups
	router.HandleFunc("/apis", Adapt(h.APIGroupLister))

	// used by kubectl to gather the OpenAPI specification of resources managed by this server.
	// TODO: implement OpenAPI v2 generator from registered CRDs
	router.HandleFunc("/openapi/v2", Adapt(h.OpenAPIHandler))
}

// CRDAPIGroups returns all supported APIGroups in the server.
func (h *APIResourceHandler) CRDAPIGroups() []metav1.APIGroup {
	return nil
}

// NewAPIResourceHandler create a new handler capable of handling APIResources.
func NewAPIResourceHandler(logger logr.Logger, repository *repository.Repository) *APIResourceHandler {
	return &APIResourceHandler{
		Repository: repository,
		Logger:     logger,
		Validator:  validation.NewRepositoryValidator(repository),
	}
}
