package server

import (
	"github.com/gorilla/mux"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
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
)

type APIResourceHandler struct {
}

// ObjectLister returns a list of objects.
func (h *APIResourceHandler) ObjectLister(vars Vars) runtime.Object {
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
func (h *APIResourceHandler) APIResourceLister(vars Vars) runtime.Object {
	return &metav1.APIResourceList{
		// TODO: add GroupVersion argument
		GroupVersion: examplesGroupVersion,
		APIResources: []metav1.APIResource{
			exampleAPIResource,
		},
	}
}

// APIGroupLister lists API groups.
func (h *APIResourceHandler) APIGroupLister(vars Vars) runtime.Object {
	return &metav1.APIGroupList{
		Groups: []metav1.APIGroup{
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
		},
	}
}

func (h *APIResourceHandler) OpenAPIHandler(vars Vars) runtime.Object {
	panic("implement me")
}

// Register adds the handler routes in the router.
func (h *APIResourceHandler) Register(router *mux.Router) {
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

// NewAPIResourceHandler create a new handler capable of handling APIResources.
func NewAPIResourceHandler() *APIResourceHandler {
	return &APIResourceHandler{}
}
