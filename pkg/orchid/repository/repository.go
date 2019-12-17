package repository

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/isutton/orchid/pkg/orchid/orm"
)

// Repository on which data is handled regarding ORM Schemas and data extrated from Unstructured,
// being ready to store CRD data in a sightly different way than regular CRs.
type Repository struct {
	schemas map[string]*orm.Schema
	orm     *orm.ORM
}

// schemaTablesLoopFn function called on each column per table in schema.
type schemaTablesLoopFn func(table *orm.Table, column *orm.Column) (interface{}, error)

// crdGVK CustomServiceDefinition GVK
var crdGVK = schema.GroupVersionKind{
	Group:   "apiextensions.k8s.io",
	Version: "v1",
	Kind:    "CustomResourceDefinition",
}

// schemaFactory make sure a single instance per schema name is used.
func (r *Repository) schemaFactory(schemaName string) *orm.Schema {
	_, exists := r.schemas[schemaName]
	if !exists {
		r.schemas[schemaName] = orm.NewSchema(schemaName)
	}
	return r.schemas[schemaName]
}

// schemaName returns the desired schema name based on GVK.
func (r *Repository) schemaName(gvk schema.GroupVersionKind) string {
	return fmt.Sprintf("%s_%s_%s", gvk.Group, gvk.Version, gvk.Kind)
}

// createCRTables execute the create tables for the CR schema. It loads the CR schema based on
// parsing the unstructured object informed. It can return error on extracting data from the
// object.
func (r *Repository) createCRTables(u *unstructured.Unstructured) error {
	openAPIV3Schema, err := extractCRDOpenAPIV3Schema(u.Object)
	if err != nil {
		return err
	}
	crGVK, err := extractCRGVKFromCRD(u.Object)
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
) (map[string][][]interface{}, error) {
	dataCRD := map[string][][]interface{}{}

	for _, table := range s.Tables {
		dataColumns := []interface{}{}
		for _, column := range table.Columns {
			if table.IsPrimaryKey(column.Name) || table.IsForeignKey(column.Name) {
				continue
			}

			// CRD data is saved as regular json, in a JSONB column, therefore it's extracted as a
			// single entry.
			var data interface{}
			if column.Name == orm.CRDRawDataColumn {
				json, err := u.MarshalJSON()
				if err != nil {
					return nil, err
				}
				data = string(json)
			} else {
				var err error
				columnFieldPath := append(table.Path, column.Name)
				data, err = extractPath(u.Object, column.OriginalType, columnFieldPath)
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
		dataCRD[table.Name] = append(dataCRD[table.Name], dataColumns)
	}

	return dataCRD, nil
}

// prepareCR prepare the data matrix from any CR resource, informed as unstructured. It can return
// error on trying to find expected data entries.
func (r *Repository) prepareCR(
	s *orm.Schema,
	u *unstructured.Unstructured,
) (map[string][][]interface{}, error) {
	obj := u.Object
	dataCR := map[string][][]interface{}{}

	for _, table := range s.Tables {
		fieldPath := table.Path
		dataTable := [][]interface{}{}

		if table.OneToMany && table.KV {
			itemMap, err := nestedMap(obj, fieldPath)
			if err != nil {
				return nil, err
			}
			dataTable = append(dataTable, extractKV(itemMap)...)
		} else if table.OneToMany {
			// if len(fieldPath) > 1 && s.HasOneToMany(fieldPath[0:len(fieldPath)-1]) {
			// 	fieldPath = fieldPath[0 : len(fieldPath)-1]
			// }
			slice, err := nestedSlice(obj, fieldPath)
			if err != nil {
				return nil, err
			}

			for _, item := range slice {
				// expect to always find KV format
				itemMap := item.(map[string]interface{})

				if table.KV {
					dataTable = append(dataTable, extractKV(itemMap)...)
				} else {
					dataColumns, err := extractColumns(itemMap, []string{}, table)
					if err != nil {
						return nil, err
					}
					dataTable = append(dataTable, dataColumns)
				}
			}
		} else {
			if table.KV {
				dataTable = append(dataTable, extractKV(obj)...)
			} else {
				dataColumns, err := extractColumns(obj, table.Path, table)
				if err != nil {
					return nil, err
				}
				dataTable = append(dataTable, dataColumns)
			}
		}
		dataCR[table.Name] = append(dataCR[table.Name], dataTable...)
	}
	return dataCR, nil
}

// Create will persist a given resource, informed as unstructured, using the ORM instance. It gives
// special treatment for CRD objects, besides of being stored, they also trigger parsing of
// OpenAPI Schema and creation of respective tables. When not an CRD object, it will only take care
// of storing the data. It can return error on extracting object data and on storing.
func (r *Repository) Create(u *unstructured.Unstructured) error {
	gvk := u.GetObjectKind().GroupVersionKind()
	s := r.schemaFactory(r.schemaName(gvk))

	// slice of slices to capture insert data per table
	var arguments map[string][][]interface{}
	var err error
	if gvk.String() == crdGVK.String() {
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

// Bootstrap the repository instance by instantiating CRD schema, and making sure the CRD storage
// has tables created. It can return error on creating CRD tables.
func (r *Repository) Bootstrap() error {
	s := r.schemaFactory(r.schemaName(crdGVK))
	s.GenerateCRD()
	return r.orm.CreateSchemaTables(s)
}

// NewRepository instantiate repository.
func NewRepository(pgORM *orm.ORM) *Repository {
	return &Repository{orm: pgORM, schemas: map[string]*orm.Schema{}}
}
