package apiserver

import (
	"fmt"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	"github.com/isutton/orchid/pkg/orchid/orm"
	"github.com/isutton/orchid/pkg/orchid/runtime"
)

type SchemaMapping map[string]*orm.Schema

// crdService is a concrete implementation of CRDService.
type crdService struct {
	schemas SchemaMapping
	orm     *orm.ORM
}

func (c *crdService) schemaFactory(schemaName string) *orm.Schema {
	if c.schemas == nil {
		c.schemas = make(SchemaMapping)
	}
	_, exists := c.schemas[schemaName]
	if !exists {
		c.schemas[schemaName] = orm.NewSchema(schemaName)
	}
	return c.schemas[schemaName]
}

func mapCRDToObject(crd *v1beta1.CustomResourceDefinition) runtime.Object {
	panic("")
}

// Create builds the underlying storage for object.
func (c *crdService) Create(crd *v1beta1.CustomResourceDefinition) error {
	s := c.schemaFactory(buildSchemaName(crd))
	return c.orm.Create(s, mapCRDToObject(crd))
}

func buildSchemaName(crd *v1beta1.CustomResourceDefinition) string {
	return fmt.Sprintf("%s_%s_%s", crd.GroupVersionKind().Group, crd.GroupVersionKind().Version, crd.Spec.Names.Plural)
}

var crdMetaDefinition = &v1beta1.CustomResourceDefinition{}

// NewCRDService returns a concrete implementation of CRDService.
func NewCRDService(orm *orm.ORM) (CRDService, error) {
	crdService := &crdService{orm: orm}
	crdSchema := crdService.schemaFactory(buildSchemaName(crdMetaDefinition))
	crdSchema.GenerateCRD()

	for _, e := range crdMetaDefinition.Spec.Versions {
		err := crdSchema.GenerateCR(e.Schema.OpenAPIV3Schema)
		if err != nil {
			return nil, err
		}
	}

	err := orm.CreateSchemaTables(crdSchema)
	if err != nil {
		return nil, err
	}

	return crdService, nil
}
