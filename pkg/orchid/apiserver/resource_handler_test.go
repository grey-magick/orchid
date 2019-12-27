package apiserver

import (
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/klogr"

	"github.com/isutton/orchid/pkg/orchid/validation"
	"github.com/isutton/orchid/test/util"
)

type TestResourcePostHandlerRepository struct {
	CRDs                 []unstructured.Unstructured
	Created              runtime.Object
	CreatedError         error
	ReadObject           runtime.Object
	ReadError            error
	OpenAPIV3Schema      *extv1.JSONSchemaProps
	OpenAPIV3SchemaError error
}

func (m *TestResourcePostHandlerRepository) List(gvk schema.GroupVersionKind, options metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	return &unstructured.UnstructuredList{Items: m.CRDs}, nil
}

func (m *TestResourcePostHandlerRepository) Create(u *unstructured.Unstructured) error {
	m.Created = u
	return m.ReadError
}

func (m *TestResourcePostHandlerRepository) Read(gvk schema.GroupVersionKind, namespacedName types.NamespacedName) (runtime.Object, error) {
	return m.ReadObject, m.ReadError
}

var (
	CustomResourceDefintionAsset = "../../../test/crds/customresourcedefinition.yaml"
	InvalidCRAsset               = "../../../test/crds/cr-invalid.yaml"
	InvalidCRDAsset              = "../../../test/crds/crd-invalid.yaml"
	ValidCRAsset                 = "../../../test/crds/cr.yaml"
	ValidCRDAsset                = "../../../test/crds/crd.yaml"
)

func TestAPIResourceHandler_ResourcePostHandler(t *testing.T) {
	logger := klogr.New()
	crd := &extv1.CustomResourceDefinition{}
	err := util.LoadObject(CustomResourceDefintionAsset, crd)
	require.NoError(t, err)
	require.Len(t, crd.Spec.Versions, 1)
	openAPIV3Schema := crd.Spec.Versions[0].Schema.OpenAPIV3Schema
	require.NotNil(t, openAPIV3Schema)

	type args struct {
		body       []byte
		logger     logr.Logger
		repository *TestResourcePostHandlerRepository
		vars       Vars
		want       runtime.Object
		wantErr    bool
	}

	assertPost := func(args args) func(*testing.T) {
		return func(t *testing.T) {
			h := &APIResourceHandler{
				Logger:     args.logger,
				Repository: args.repository,
				Validator:  validation.NewValidator(args.repository),
			}
			got, err := h.ResourcePostHandler(args.vars, args.body)
			if args.wantErr {
				require.Error(t, err)
				if args.repository.ReadError != nil {
					require.Equal(t, args.repository.ReadError, err)
				}
				return
			}
			require.NoError(t, err)
			util.RequireYamlEqual(t, got, args.want)
			util.RequireYamlEqual(t, args.repository.Created, args.want)
		}
	}

	t.Run("body empty", assertPost(
		args{
			body:       []byte{},
			logger:     logger,
			repository: &TestResourcePostHandlerRepository{ReadError: BodyEmptyErr},
			wantErr:    true,
		},
	))

	t.Run("resource definition manifest is invalid", assertPost(
		args{
			body:   util.ReadAsset(InvalidCRDAsset),
			logger: logger,
			repository: &TestResourcePostHandlerRepository{
				CRDs: []unstructured.Unstructured{
					*(util.LoadUnstructured(CustomResourceDefintionAsset)),
				},
				ReadObject:      util.LoadUnstructured(InvalidCRAsset),
				ReadError:       validation.InvalidObjectErr,
				OpenAPIV3Schema: openAPIV3Schema,
			},
			wantErr: true,
		},
	))

	t.Run("resource manifest is invalid", assertPost(
		args{
			body:   util.ReadAsset(InvalidCRAsset),
			logger: logger,
			repository: &TestResourcePostHandlerRepository{
				CRDs: []unstructured.Unstructured{
					*(util.LoadUnstructured(ValidCRDAsset)),
				},
				ReadObject:      util.LoadUnstructured(InvalidCRAsset),
				ReadError:       validation.InvalidObjectErr,
				OpenAPIV3Schema: openAPIV3Schema,
			},
			wantErr: true,
		},
	))

	t.Run("resource has unknown gvk", assertPost(
		args{
			body:   util.ReadAsset(ValidCRAsset),
			logger: logger,
			repository: &TestResourcePostHandlerRepository{
				ReadError:       validation.GVKNotFoundErr,
				OpenAPIV3Schema: openAPIV3Schema,
			},
			wantErr: true,
		},
	))

	t.Run("crontab resource definition can be created", assertPost(
		args{
			body:   util.ReadAsset(ValidCRDAsset),
			logger: logger,
			repository: &TestResourcePostHandlerRepository{
				CRDs: []unstructured.Unstructured{
					*(util.LoadUnstructured(CustomResourceDefintionAsset)),
				},
				ReadObject:      util.LoadUnstructured(ValidCRDAsset),
				OpenAPIV3Schema: openAPIV3Schema,
			},
			wantErr: false,
			want:    util.LoadUnstructured(ValidCRDAsset),
		},
	))

	t.Run("crontab can be created", assertPost(
		args{
			body:   util.ReadAsset(ValidCRAsset),
			logger: logger,
			repository: &TestResourcePostHandlerRepository{
				CRDs: []unstructured.Unstructured{
					*(util.LoadUnstructured(ValidCRDAsset)),
				},
				ReadObject:      util.LoadUnstructured(ValidCRAsset),
				OpenAPIV3Schema: openAPIV3Schema,
			},
			wantErr: false,
			want:    util.LoadUnstructured(ValidCRAsset),
		},
	))

	t.Run("crontab resource definition does not exist", assertPost(
		args{
			body:   util.ReadAsset(ValidCRAsset),
			logger: logger,
			repository: &TestResourcePostHandlerRepository{
				ReadObject:      util.LoadUnstructured(ValidCRAsset),
				ReadError:       validation.GVKNotFoundErr,
				OpenAPIV3Schema: openAPIV3Schema,
			},
			wantErr: true,
		},
	))
}

func TestAPIResourceHandler_Validate(t *testing.T) {
	crd := &extv1.CustomResourceDefinition{}
	require.NoError(t, util.LoadObject(ValidCRDAsset, crd))
	require.Len(t, crd.Spec.Versions, 1)
	props := crd.Spec.Versions[0].Schema.OpenAPIV3Schema
	require.NotNil(t, props)

	type args struct {
		obj        *unstructured.Unstructured
		repository *TestResourcePostHandlerRepository
		vars       Vars
		wantErr    bool
		err        error
	}

	assertValidation := func(args args) func(*testing.T) {
		return func(t *testing.T) {
			v := validation.NewValidator(args.repository)
			err := v.Validate(args.obj)
			if args.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		}
	}

	t.Run("schema not found", assertValidation(
		args{
			obj:        util.LoadUnstructured(ValidCRAsset),
			repository: &TestResourcePostHandlerRepository{},
			wantErr:    true,
		},
	))

	t.Run("valid cr", assertValidation(
		args{
			obj: util.LoadUnstructured(ValidCRAsset),
			repository: &TestResourcePostHandlerRepository{
				CRDs: []unstructured.Unstructured{
					*(util.LoadUnstructured(ValidCRDAsset)),
				},
				OpenAPIV3Schema: props,
			},
			wantErr: false,
		},
	))

	t.Run("invalid cr", assertValidation(
		args{
			obj:        util.LoadUnstructured(InvalidCRAsset),
			repository: &TestResourcePostHandlerRepository{OpenAPIV3Schema: props},
			wantErr:    true,
		},
	))
}
