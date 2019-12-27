package repository

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"github.com/isutton/orchid/pkg/orchid/orm"
)

// Repository on which data is handled regarding ORM Schemas and data extrated from Unstructured,
// being ready to store CRD data in a sightly different way than regular CRs.
type Repository struct {
	logger  logr.Logger            // logger instance
	schemas map[string]*orm.Schema // schema name and instance
	orm     *orm.ORM               // orm instance
}

/*
[
  {
    "id": 1,
    "apiversion": "apiextensions.k8s.io/v1",
    "kind": "CustomResourceDefinition",
    "data": "{\"kind\": \"CustomResourceDefinition\", \"spec\": {\"group\": \"stable.example.com\", \"names\": {\"kind\": \"CronTab\", \"plural\": \"crontabs\", \"singular\": \"crontab\", \"shortNames\": [\"ct\"]}, \"scope\": \"Namespaced\", \"schema\": {\"openAPIV3Schema\": {\"type\": \"object\", \"properties\": {\"spec\": {\"type\": \"object\", \"properties\": {\"image\": {\"type\": \"string\"}, \"cronSpec\": {\"type\": \"string\"}, \"replicas\": {\"type\": \"integer\"}}}}}}, \"version\": \"v1\"}, \"metadata\": {\"name\": \"crontabs.stable.example.com\", \"namespace\": \"\", \"annotations\": {\"kubectl.kubernetes.io/last-applied-configuration\": \"{\\\"apiVersion\\\":\\\"apiextensions.k8s.io/v1\\\",\\\"kind\\\":\\\"CustomResourceDefinition\\\",\\\"metadata\\\":{\\\"annotations\\\":{},\\\"name\\\":\\\"crontabs.stable.example.com\\\",\\\"namespace\\\":\\\"\\\"},\\\"spec\\\":{\\\"group\\\":\\\"stable.example.com\\\",\\\"names\\\":{\\\"kind\\\":\\\"CronTab\\\",\\\"plural\\\":\\\"crontabs\\\",\\\"shortNames\\\":[\\\"ct\\\"],\\\"singular\\\":\\\"crontab\\\"},\\\"schema\\\":{\\\"openAPIV3Schema\\\":{\\\"properties\\\":{\\\"spec\\\":{\\\"properties\\\":{\\\"cronSpec\\\":{\\\"type\\\":\\\"string\\\"},\\\"image\\\":{\\\"type\\\":\\\"string\\\"},\\\"replicas\\\":{\\\"type\\\":\\\"integer\\\"}},\\\"type\\\":\\\"object\\\"}},\\\"type\\\":\\\"object\\\"}},\\\"scope\\\":\\\"Namespaced\\\",\\\"version\\\":\\\"v1\\\"}}\\n\"}}, \"apiVersion\": \"apiextensions.k8s.io/v1\"}"
  }
]
*/

var CRDNotFoundErr = errors.New("CRD not found")

func buildResourceGVK(crd *unstructured.Unstructured) schema.GroupVersionKind {
	group, _, _ := unstructured.NestedString(crd.Object, "spec", "group")
	version, _, _ := unstructured.NestedString(crd.Object, "spec", "version")
	kind, _, _ := unstructured.NestedString(crd.Object, "spec", "names", "kind")
	return schema.GroupVersionKind{
		Group:   group,
		Version: version,
		Kind:    kind,
	}
}

func (r *Repository) OpenAPIV3SchemaForGVK(gvk schema.GroupVersionKind) (*extv1.JSONSchemaProps, error) {
	crds, err := r.List(CRDGVK, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, crd := range crds.Items {
		if buildResourceGVK(&crd).String() == gvk.String() {
			v, exists, err := unstructured.NestedFieldNoCopy(crd.Object, "spec", "validation", "openAPIV3Schema")
			if err != nil {
				return nil, err
			}
			if !exists {
				return nil, errors.New("field does not exist")
			}
			o, ok := v.(*extv1.JSONSchemaProps)
			if !ok {
				return nil, errors.New("field is not JSONSchemaProps")
			}
			return o, nil
		}
	}

	return nil, CRDNotFoundErr
}

// CRDGVK CustomServiceDefinition GVK
var CRDGVK = schema.GroupVersionKind{
	Group:   "apiextensions.k8s.io",
	Version: "v1",
	Kind:    "CustomResourceDefinition",
}

// schemaFactory make sure a single instance per schema name is used.
func (r *Repository) schemaFactory(schemaName string) *orm.Schema {
	_, exists := r.schemas[schemaName]
	if !exists {
		r.schemas[schemaName] = orm.NewSchema(r.logger, schemaName)
	}
	return r.schemas[schemaName]
}

// schemaName returns the desired schema name based on GVK.
func (r *Repository) schemaName(gvk schema.GroupVersionKind) string {
	group := strings.ReplaceAll(gvk.Group, ".", "_")
	return fmt.Sprintf("%s_%s_%s", group, gvk.Version, gvk.Kind)
}

// createCRTables execute the create tables for the CR schema. It loads the CR schema based on
// parsing the unstructured object informed. It can return error on extracting data from the
// object.
func (r *Repository) createCRTables(u *unstructured.Unstructured) error {
	openAPIV3Schema, err := extractCRDOpenAPIV3Schema(u.Object)
	if err != nil {
		return err
	}
	crGVK, err := ExtractCRGVKFromCRD(u.Object)
	if err != nil {
		return err
	}

	crSchema := r.schemaFactory(r.schemaName(crGVK))
	if err = crSchema.GenerateCR(openAPIV3Schema); err != nil {
		return err
	}

	return r.orm.CreateSchemaTables(crSchema)
}

// prepareCRD prepare the data matrix from a given CRD resource, informed as unstructured. It can
// return error in case of trying to extract data.
func (r *Repository) prepareCRD(
	s *orm.Schema,
	u *unstructured.Unstructured,
) (orm.MappedMatrix, error) {
	obj := u.Object
	crd := make(orm.MappedMatrix)

	for _, table := range s.Tables {
		dataColumns := []interface{}{}
		for _, column := range table.Columns {
			if table.IsPrimaryKey(column.Name) || table.IsForeignKey(column.Name) {
				continue
			}

			// CRD data is saved as regular json, in a JSONB column, therefore it's extracted as a
			// single entry.
			var data interface{}
			if column.Name == orm.CRDRawColumn {
				json, err := u.MarshalJSON()
				if err != nil {
					return nil, err
				}
				data = string(json)
			} else {
				var err error
				columnFieldPath := append(table.Path, column.Name)
				data, err = extractPath(obj, column.JSType, columnFieldPath)
				if err != nil {
					if column.NotNull {
						return nil, err
					}
					if data, err = column.Null(); err != nil {
						return nil, err
					}
				}
			}
			dataColumns = append(dataColumns, data)
		}
		crd[table.Name] = append(crd[table.Name], dataColumns)
	}

	return crd, nil
}

// prepareCR prepare the data matrix from any CR resource, informed as unstructured. It can return
// error on trying to find expected data entries.
func (r *Repository) prepareCR(
	s *orm.Schema,
	u *unstructured.Unstructured,
) (orm.MappedMatrix, error) {
	obj := u.Object
	cr := orm.MappedMatrix{}
	nested := NewNested(s, obj)

	for _, table := range s.Tables {
		fieldPath := table.Path
		dataTable := []orm.List{}
		extracted := []orm.Entry{}

		if len(fieldPath) == 0 {
			extracted = append(extracted, obj)
		} else {
			var err error
			if extracted, err = nested.Extract(fieldPath); err != nil {
				return nil, err
			}
		}

		for _, entry := range extracted {
			if table.KV {
				dataTable = append(dataTable, extractKV(entry)...)
			} else {
				dataColumns, err := extractColumns(entry, []string{}, table)
				if err != nil {
					return nil, err
				}
				dataTable = append(dataTable, dataColumns)
			}
		}

		cr[table.Name] = append(cr[table.Name], dataTable...)
	}
	return cr, nil
}

// Create will persist a given resource, informed as unstructured, using the ORM instance. It gives
// special treatment for CRD objects, besides of being stored, they also trigger parsing of
// OpenAPI Schema and creation of respective tables. When not an CRD object, it will only take care
// of storing the data. It can return error on extracting object data and on storing.
func (r *Repository) Create(u *unstructured.Unstructured) error {
	gvk := u.GetObjectKind().GroupVersionKind()
	s := r.schemaFactory(r.schemaName(gvk))

	// slice of slices to capture insert data per table
	var arguments orm.MappedMatrix
	var err error
	if gvk.String() == CRDGVK.String() {
		if arguments, err = r.prepareCRD(s, u); err != nil {
			return err
		}
		if err = r.createCRTables(u); err != nil {
			return err
		}
	} else {
		if arguments, err = r.prepareCR(s, u); err != nil {
			return err
		}
	}

	if len(arguments) == 0 {
		return fmt.Errorf("unable to parse arguments from object")
	}
	return r.orm.Create(s, arguments)
}

// Read a single object from ORM, searching for a namespaced-name. It can return errors from
// querying the database, preparing the result-set, and assembling an unstructured object.
func (r *Repository) Read(gvk schema.GroupVersionKind, namespacedName types.NamespacedName) (runtime.Object, error) {
	s := r.schemaFactory(r.schemaName(gvk))

	rs, err := r.orm.Read(s, namespacedName)
	if err != nil {
		return nil, err
	}

	assembler := NewAssembler(r.logger, s, rs)
	objects, err := assembler.Build()
	if err != nil {
		return nil, err
	}
	if len(objects) > 1 {
		r.logger.WithValues("objects", len(objects)).Info("WARNING: unexpected number of objects!")
	}

	u := objects[0]
	u.SetGroupVersionKind(gvk)
	return u, nil
}

// List objects from schema based on metav1.ListOptions.
func (r *Repository) List(
	gvk schema.GroupVersionKind,
	options metav1.ListOptions,
) (*unstructured.UnstructuredList, error) {
	s := r.schemaFactory(r.schemaName(gvk))

	labelsSet, err := labels.ConvertSelectorToLabelsMap(options.LabelSelector)
	if err != nil {
		return nil, err
	}

	rs, err := r.orm.List(s, labelsSet)
	if err != nil {
		return nil, err
	}

	assembler := NewAssembler(r.logger, s, rs)
	objects, err := assembler.Build()
	if err != nil {
		return nil, err
	}

	list := &unstructured.UnstructuredList{Items: []unstructured.Unstructured{}}
	for _, u := range objects {
		u.SetGroupVersionKind(gvk)
		list.Items = append(list.Items, *u)
	}
	return list, nil
}

// CRDDefinition is the CustomResourceDefinition of a CustomResourceDefinition.
var CRDDefinition = &apiextensionsv1.CustomResourceDefinition{
	ObjectMeta: metav1.ObjectMeta{
		Name: "customresourcedefinitions.apiextensions.k8s.io",
	},
	TypeMeta: metav1.TypeMeta{
		Kind:       "CustomResourceDefinition",
		APIVersion: "apiextensions.k8s.io/v1",
	},
	Spec: apiextensionsv1.CustomResourceDefinitionSpec{
		Group: "apiextensions.k8s.io",
		Names: apiextensionsv1.CustomResourceDefinitionNames{
			Plural:     "customresourcedefinitions",
			Singular:   "customresourcedefinition",
			ShortNames: []string{"crd", "crds"},
			Kind:       "CustomResourceDefinition",
			ListKind:   "CustomResourceDefinitionList",
			Categories: nil,
		},
		Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
			{
				Name: "v1",
				Schema: &apiextensionsv1.CustomResourceValidation{
					OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
						Properties: map[string]apiextensionsv1.JSONSchemaProps{
							"spec": {
								Type: "object",
								Properties: map[string]apiextensionsv1.JSONSchemaProps{
									"group": {Type: "string"},
									"names": {
										Type: "object",
										Properties: map[string]apiextensionsv1.JSONSchemaProps{
											"plural":   {Type: "string"},
											"singular": {Type: "string"},
											"kind":     {Type: "string"},
											"listKind": {Type: "string"},
										},
									},
									"versions": {
										Type: "array",
										AdditionalItems: &apiextensionsv1.JSONSchemaPropsOrBool{
											Schema: &apiextensionsv1.JSONSchemaProps{
												Type: "object",
												Properties: map[string]apiextensionsv1.JSONSchemaProps{
													"name":   {Type: "string"},
													"schema": {Type: "object"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	},
}

// Bootstrap the repository instance by instantiating CRD schema, and making sure the CRD storage
// has tables created. It can return error on creating CRD tables.
func (r *Repository) Bootstrap() error {
	s := r.schemaFactory(r.schemaName(CRDGVK))
	s.GenerateCRD()
	err := r.orm.CreateSchemaTables(s)
	if err != nil {
		return err
	}
	uObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(CRDDefinition)
	if err != nil {
		return err
	}
	u := &unstructured.Unstructured{Object: uObj}
	arguments, err := r.prepareCRD(s, u)
	if err != nil {
		return err
	}
	return r.orm.Create(s, arguments)
}

// NewRepository instantiate repository.
func NewRepository(logger logr.Logger, pgORM *orm.ORM) *Repository {
	return &Repository{
		logger:  logger.WithName("repository"),
		orm:     pgORM,
		schemas: map[string]*orm.Schema{},
	}
}
