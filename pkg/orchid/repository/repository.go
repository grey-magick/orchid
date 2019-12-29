package repository

import (
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"github.com/isutton/orchid/pkg/orchid/config"
	jsc "github.com/isutton/orchid/pkg/orchid/jsonschema"
	"github.com/isutton/orchid/pkg/orchid/orm"
)

// ResourceRepository is the repository interface
type ResourceRepository interface {
	Create(u *unstructured.Unstructured) error
	Read(gvk schema.GroupVersionKind, namespacedName types.NamespacedName) (*unstructured.Unstructured, error)
	List(ns string, gvk schema.GroupVersionKind, options metav1.ListOptions) (*unstructured.UnstructuredList, error)
}

// Repository on which data is handled regarding ORM Schemas and data extrated from Unstructured,
// being ready to store CRD data in a sightly different way than regular CRs.
type Repository struct {
	logger          logr.Logger                    // logger instance
	config          *config.Config                 // configuration instance
	schemas         map[string]*orm.Schema         // schema name and instance
	orms            map[string]map[string]*orm.ORM // namespace and instances by name
	gvkPerNamespace map[string][]string            // cached list of where GVK has been applied
}

// DefaultNamespace namespace name or orchid's metadata
const DefaultNamespace = "orchid"

// CRDGVK CustomServiceDefinition GVK
var CRDGVK = schema.GroupVersionKind{
	Group:   "apiextensions.k8s.io",
	Version: "v1",
	Kind:    "CustomResourceDefinition",
}

// ormFactory creates a single ORM bootstrapped instance per namespace GVK.Group combination.
func (r *Repository) ormFactory(ns string, group string) *orm.ORM {
	if _, exists := r.orms[ns]; !exists {
		r.orms[ns] = map[string]*orm.ORM{}
	}
	if _, exists := r.orms[ns][group]; !exists {
		r.logger.WithValues("database", ns, "group", group).Info("Instantiating ORM...")
		r.orms[ns][group] = orm.NewORM(r.logger, ns, group, r.config)
	}
	return r.orms[ns][group]
}

// schemaFactory creates a single schema instance per name.
func (r *Repository) schemaFactory(schemaName string) *orm.Schema {
	_, exists := r.schemas[schemaName]
	if !exists {
		r.logger.WithValues("schema", schemaName).Info("Instantiating Schema...")
		r.schemas[schemaName] = orm.NewSchema(r.logger, schemaName)
	}
	return r.schemas[schemaName]
}

// schemaNameforGVK returns a orm.Schema name for a given GVK.
func (r *Repository) schemaNameforGVK(gvk schema.GroupVersionKind) string {
	return fmt.Sprintf("%s_%s", gvk.Version, gvk.Kind)
}

// factory instantiate the schema and ORM instances, making sure a single instance is in use for
// the combination of namespace and GVK.
func (r *Repository) factory(ns string, gvk schema.GroupVersionKind) (*orm.ORM, *orm.Schema, error) {
	if gvk.Group == "" || gvk.Version == "" || gvk.Kind == "" {
		return nil, nil, fmt.Errorf("incomplete GVK '%#v'", gvk)
	}

	group := strings.ReplaceAll(gvk.Group, ".", "_")
	o := r.ormFactory(ns, group)
	s := r.schemaFactory(r.schemaNameforGVK(gvk))

	// checking when database connection is not yet instantiated to check if schemas has tables
	// defined, and therefore create them, after this steps the database connection will be
	// instantiated and thus won't be subject to connect or create-tables again
	logger := r.logger.WithValues("namespace", ns, "GVK", gvk)
	if o.DB == nil {
		logger.Info("Bootstrapping database connection...")
		if err := o.Bootstrap(); err != nil {
			return nil, nil, err
		}
		if len(s.Tables) > 0 {
			logger.Info("Creating schema tables")
			if err := o.CreateTables(s); err != nil {
				return nil, nil, err
			}
		}
	}
	return o, s, nil
}

// decompose prepare the data matrix from any CR resource, informed as unstructured. It can return
// error on trying to find expected data entries.
func (r *Repository) decompose(
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

// initializeSchema extracts the GVK and OpenAPI Schema from CRD object, and initialize orm.Schema.
// It can return errors on extracting data.
func (r *Repository) initializeSchema(obj map[string]interface{}) error {
	gvk, err := ExtractCRGVKFromCRD(obj)
	if err != nil {
		return err
	}
	openAPIV3Schema, err := ExtractCRDOpenAPIV3Schema(obj)
	if err != nil {
		return err
	}
	crSchema := r.schemaFactory(r.schemaNameforGVK(gvk))
	return crSchema.Generate(openAPIV3Schema)
}

// Create will persist a given resource, informed as unstructured, using the ORM instance. It gives
// special treatment for CRD objects, besides of being stored, they also trigger parsing of
// OpenAPI Schema and creation of respective tables. When not an CRD object, it will only take care
// of storing the data. It can return error on extracting object data and on storing.
func (r *Repository) Create(u *unstructured.Unstructured) error {
	gvk := u.GetObjectKind().GroupVersionKind()
	isCRD := gvk.String() == CRDGVK.String()

	var ns string
	if isCRD {
		ns = DefaultNamespace
	} else {
		ns = u.GetNamespace()
	}

	o, s, err := r.factory(ns, gvk)
	if err != nil {
		return err
	}
	if o.DB == nil {
		if err = o.CreateTables(s); err != nil {
			return err
		}
	}

	arguments, err := r.decompose(s, u)
	if err != nil {
		return err
	}
	if len(arguments) == 0 {
		return fmt.Errorf("unable to parse arguments from object")
	}
	if err = o.Create(s, arguments); err != nil {
		return err
	}

	if isCRD {
		return r.initializeSchema(u.Object)
	}
	return nil
}

// Read a single object from ORM, searching for a namespaced-name. It can return errors from
// querying the database, preparing the result-set, and assembling an unstructured object.
func (r *Repository) Read(
	gvk schema.GroupVersionKind,
	namespacedName types.NamespacedName,
) (*unstructured.Unstructured, error) {
	o, s, err := r.factory(namespacedName.Namespace, gvk)
	if err != nil {
		return nil, err
	}
	rs, err := o.Read(s, namespacedName)
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
	ns string,
	gvk schema.GroupVersionKind,
	options metav1.ListOptions,
) (*unstructured.UnstructuredList, error) {
	o, s, err := r.factory(ns, gvk)
	if err != nil {
		return nil, err
	}

	labelsSet, err := labels.ConvertSelectorToLabelsMap(options.LabelSelector)
	if err != nil {
		return nil, err
	}

	rs, err := o.List(s, labelsSet)
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

// Bootstrap the repository instance by instantiating CRD schema, and making sure the CRD storage
// has tables created. It can return error on creating CRD tables.
func (r *Repository) Bootstrap() error {
	// TODO: bootstrap existing schemas, check created CRDs and instantiate orm and schemas;

	o, s, err := r.factory(DefaultNamespace, CRDGVK)
	if err != nil {
		return err
	}

	// preparing database structure, before being able to store and retrieve data
	if err = o.Bootstrap(); err != nil {
		return err
	}

	// generating schema for CRD and creating tables
	openAPIV3Schema := jsc.OrchidOpenAPIV3Schema()
	if err = s.Generate(&openAPIV3Schema); err != nil {
		return err
	}
	return o.CreateTables(s)
}

// NewRepository instantiate repository.
func NewRepository(logger logr.Logger, config *config.Config) *Repository {
	return &Repository{
		logger:          logger.WithName("repository"),
		config:          config,
		orms:            map[string]map[string]*orm.ORM{},
		schemas:         map[string]*orm.Schema{},
		gvkPerNamespace: map[string][]string{},
	}
}
