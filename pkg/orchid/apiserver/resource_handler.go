package apiserver

import (
	"github.com/ghodss/yaml"
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"

	"github.com/isutton/orchid/pkg/orchid/repository"
	"github.com/isutton/orchid/pkg/orchid/runtime"
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

type APIResourceHandler struct {
	Logger     logr.Logger
	Repository *repository.Repository
}

// ObjectLister returns a list of objects.
func (h *APIResourceHandler) ObjectLister(vars Vars, body []byte) k8sruntime.Object {
	apiVersion, err := vars.GetAPIVersion()
	if err != nil {
		// TODO: handle error properly
	}
	m := make(map[string]interface{})
	u := &unstructured.Unstructured{Object: m}
	u.SetAPIVersion(apiVersion)
	u.SetKind(exampleAPIResource.Kind + "List")
	items := make([]interface{}, 0)
	_ = unstructured.SetNestedSlice(m, items, "items")
	return u
}

// APIResourceLister lists API resources.
func (h *APIResourceHandler) APIResourceLister(vars Vars, body []byte) k8sruntime.Object {
	return &metav1.APIResourceList{
		// TODO: add GroupVersion argument
		GroupVersion: examplesGroupVersion,
		APIResources: []metav1.APIResource{
			exampleAPIResource,
			crdAPIResource,
		},
	}
}

// APIGroupLister lists API groups.
func (h *APIResourceHandler) APIGroupLister(vars Vars, body []byte) k8sruntime.Object {
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
	}
}

func (h *APIResourceHandler) OpenAPIHandler(vars Vars, body []byte) k8sruntime.Object {
	return nil
}

// ResourcePostHandler handles the create resource action.
func (h *APIResourceHandler) ResourcePostHandler(vars Vars, body []byte) k8sruntime.Object {
	// deserialize the body to an Orchid object
	obj := runtime.NewObject()
	err := yaml.Unmarshal(body, obj)
	if err != nil {
		panic(err)
	}

	uObj := &map[string]interface{}{}
	err = yaml.Unmarshal(body, uObj)
	if err != nil {
		panic(err)
	}
	u := &unstructured.Unstructured{Object: *uObj}
	// TODO: this handler should return an error
	err = h.Repository.Create(u)
	if err != nil {
		panic(err)
	}

	return obj
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
	router.HandleFunc("/apis/{group}/{version}/{resource}", Adapt(h.ObjectLister))
	// used by kubectl to discover all the resources for an API Group
	router.HandleFunc("/apis/{group}/{version}", Adapt(h.APIResourceLister))
	// used by kubectl to discover available API Groups
	router.HandleFunc("/apis", Adapt(h.APIGroupLister))

	// used by kubectl to gather the OpenAPI specification of resources managed by this server.
	// TODO: implement OpenAPI v2 generator from registered CRDs
	router.HandleFunc("/openapi/v2", Adapt(h.OpenAPIHandler))
}

func (h *APIResourceHandler) CRDAPIGroups() []metav1.APIGroup {
	return nil
}

// NewAPIResourceHandler create a new handler capable of handling APIResources.
func NewAPIResourceHandler(logger logr.Logger, repository *repository.Repository) *APIResourceHandler {
	return &APIResourceHandler{
		Repository: repository,
		Logger:     logger,
	}
}
