package apiserver

import (
	"errors"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/klogr"

	"github.com/isutton/orchid/test/util"
)

type TestResourcePostHandlerRepository struct {
	Created              runtime.Object
	CreatedError         error
	ReadObject           runtime.Object
	ReadError            error
	OpenAPIV3Schema      *apiextensionsv1.JSONSchemaProps
	OpenAPIV3SchemaError error
}

var SchemaNotFoundErr = errors.New("schema not found")

func (m *TestResourcePostHandlerRepository) OpenAPIV3SchemaForGVK(gvk schema.GroupVersionKind) (*apiextensionsv1.JSONSchemaProps, error) {
	if m.OpenAPIV3SchemaError != nil {
		return nil, m.OpenAPIV3SchemaError
	}
	if m.OpenAPIV3Schema == nil {
		return nil, SchemaNotFoundErr
	}
	return m.OpenAPIV3Schema, nil
}

func (m *TestResourcePostHandlerRepository) Create(u *unstructured.Unstructured) error {
	m.Created = u
	return m.ReadError
}

func (m *TestResourcePostHandlerRepository) Read(gvk schema.GroupVersionKind, namespacedName types.NamespacedName) (runtime.Object, error) {
	return m.ReadObject, m.ReadError
}

var (
	InvalidCRAsset = "../../../test/crds/cr-invalid.yaml"
	ValidCRAsset   = "../../../test/crds/cr.yaml"
	ValidCRDAsset  = "../../../test/crds/crd.yaml"
)

func TestAPIResourceHandler_ResourcePostHandler(t *testing.T) {
	logger := klogr.New()
	crd := &apiextensionsv1.CustomResourceDefinition{}
	err := util.LoadObject(ValidCRDAsset, crd)
	require.NoError(t, err)
	require.Len(t, crd.Spec.Versions, 1)
	openAPIV3Schema := crd.Spec.Versions[0].Schema.OpenAPIV3Schema
	require.NotNil(t, openAPIV3Schema)

	tests := []struct {
		body       []byte
		logger     logr.Logger
		name       string
		repository *TestResourcePostHandlerRepository
		vars       Vars
		want       runtime.Object
		wantErr    bool
	}{
		{
			body:       []byte{},
			logger:     logger,
			name:       "empty body",
			repository: &TestResourcePostHandlerRepository{ReadError: BodyEmptyErr},
			wantErr:    true,
		},
		{
			body:   util.ReadAsset(InvalidCRAsset),
			logger: logger,
			name:   "invalid cr",
			repository: &TestResourcePostHandlerRepository{
				ReadObject:      util.LoadUnstructured(InvalidCRAsset),
				ReadError:       InvalidObjectErr,
				OpenAPIV3Schema: openAPIV3Schema,
			},
			wantErr: true,
		},
		{
			body:   util.ReadAsset(ValidCRDAsset),
			logger: logger,
			name:   "crontabs crd",
			repository: &TestResourcePostHandlerRepository{
				ReadObject:      util.LoadUnstructured(ValidCRDAsset),
				OpenAPIV3Schema: openAPIV3Schema,
			},
			wantErr: false,
			want:    util.LoadUnstructured(ValidCRDAsset),
		},
		{
			body:   util.ReadAsset(ValidCRDAsset),
			logger: logger,
			name:   "error in crontabs crd",
			repository: &TestResourcePostHandlerRepository{
				ReadObject:      util.LoadUnstructured(ValidCRDAsset),
				ReadError:       errors.New(""),
				OpenAPIV3Schema: openAPIV3Schema,
			},
			wantErr: true,
		},
		{
			body:   util.ReadAsset(ValidCRAsset),
			logger: logger,
			name:   "crontabs cr",
			repository: &TestResourcePostHandlerRepository{
				ReadObject:      util.LoadUnstructured(ValidCRAsset),
				OpenAPIV3Schema: openAPIV3Schema,
			},
			wantErr: false,
			want:    util.LoadUnstructured(ValidCRAsset),
		},
		{
			body:   util.ReadAsset(ValidCRAsset),
			logger: logger,
			name:   "error in crontabs cr",
			repository: &TestResourcePostHandlerRepository{
				ReadObject:      util.LoadUnstructured(ValidCRAsset),
				ReadError:       errors.New(""),
				OpenAPIV3Schema: openAPIV3Schema,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &APIResourceHandler{
				Logger:     tt.logger,
				Repository: tt.repository,
			}
			got, err := h.ResourcePostHandler(tt.vars, tt.body)
			if tt.wantErr {
				require.Error(t, err)
				if tt.repository.ReadError != nil {
					require.Equal(t, tt.repository.ReadError, err)
				}
				return
			}
			require.NoError(t, err)
			util.RequireYamlEqual(t, got, tt.want)
			util.RequireYamlEqual(t, tt.repository.Created, tt.want)
		})
	}
}

func TestAPIResourceHandler_Validate(t *testing.T) {
	logger := klogr.New()
	crd := &apiextensionsv1.CustomResourceDefinition{}
	require.NoError(t, util.LoadObject(ValidCRDAsset, crd))
	require.Len(t, crd.Spec.Versions, 1)
	props := crd.Spec.Versions[0].Schema.OpenAPIV3Schema
	require.NotNil(t, props)

	tests := []struct {
		name       string
		obj        runtime.Object
		repository *TestResourcePostHandlerRepository
		vars       Vars
		wantErr    bool
	}{
		{
			name:       "schema not found",
			obj:        util.LoadUnstructured(ValidCRAsset),
			repository: &TestResourcePostHandlerRepository{},
			wantErr:    true,
		},
		{
			name:       "valid cr",
			obj:        util.LoadUnstructured(ValidCRAsset),
			repository: &TestResourcePostHandlerRepository{OpenAPIV3Schema: props},
			wantErr:    false,
		},
		{
			name:       "invalid cr",
			obj:        util.LoadUnstructured(InvalidCRAsset),
			repository: &TestResourcePostHandlerRepository{OpenAPIV3Schema: props},
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &APIResourceHandler{
				Logger:     logger,
				Repository: tt.repository,
			}
			err := h.Validate(tt.obj)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
