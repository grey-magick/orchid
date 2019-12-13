package apiserver

import (
	"github.com/ghodss/yaml"
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"

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
)

// CRDService manages CRDs.
type CRDService interface {
	// Create registers a CRD.
	Create(object *v1beta1.CustomResourceDefinition) error
}

type APIResourceHandler struct {
	// CRDService is responsible for managing CRDs.
	CRDService CRDService
	Logger     logr.Logger
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
		},
	}
}

// APIGroupLister lists API groups.
func (h *APIResourceHandler) APIGroupLister(vars Vars, body []byte) k8sruntime.Object {
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

func (h *APIResourceHandler) OpenAPIHandler(vars Vars, body []byte) k8sruntime.Object {
	panic("implement me")
}

// ResourcePostHandler handles the create resource action.
func (h *APIResourceHandler) ResourcePostHandler(vars Vars, body []byte) k8sruntime.Object {
	// deserialize the body to an Orchid object
	obj := runtime.NewObject()
	err := yaml.Unmarshal(body, obj)
	if err != nil {
		panic(err)
	}

	if isCustomResourceDefinition(obj) {
		// create all storage resources for this new object.
		var crd *v1beta1.CustomResourceDefinition
		err := yaml.Unmarshal(body, crd)
		if err != nil {
			panic(err)
		}
		err = h.CRDService.Create(crd)
		if err != nil {
			panic(err)
		}
	}

	return obj
}

// isCustomResourceDefinition returns whether obj is a CustomResourceDefinition
func isCustomResourceDefinition(obj runtime.Object) bool {
	gvk := obj.GetObjectKind().GroupVersionKind()

	return gvk.Group == "apiextensions.k8s.io" &&
		gvk.Version == "v1" &&
		gvk.Kind == "CustomResourceDefinition"
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

// NewAPIResourceHandler create a new handler capable of handling APIResources.
func NewAPIResourceHandler(logger logr.Logger, crdService CRDService) *APIResourceHandler {
	return &APIResourceHandler{
		Logger:     logger,
		CRDService: crdService,
	}
}
