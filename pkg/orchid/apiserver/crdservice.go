package apiserver

import (
	"fmt"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	"github.com/isutton/orchid/pkg/orchid/orm"
	"github.com/isutton/orchid/pkg/orchid/runtime"
)

// crdService is a concrete implementation of CRDService.
type crdService struct {
	schemas map[string]*orm.Schema
	orm     *orm.ORM
}

func (c *crdService) schemaFactory(schemaName string) *orm.Schema {
	_, exists := c.schemas[schemaName]
	if !exists {
		c.schemas[schemaName] = orm.NewSchema(schemaName)
	}
	return c.schemas[schemaName]
}

func schemaName(crd *v1beta1.CustomResourceDefinition) string {
	return fmt.Sprintf("%s_%s_%s",
		crd.GroupVersionKind().Group, crd.GroupVersionKind().Version, crd.Spec.Names.Plural)
}

func mapCRDToObject(crd *v1beta1.CustomResourceDefinition) runtime.Object {
	panic("")
}

// Create builds the underlying storage for object.
func (c *crdService) Create(crd *v1beta1.CustomResourceDefinition) error {
	s := c.schemaFactory(schemaName(crd))
	return c.orm.Create(s, mapCRDToObject(crd))
}

// NewCRDService returns a concrete implementation of CRDService.
func NewCRDService(orm *orm.ORM) (CRDService, error) {
	crdService := &crdService{orm: orm}
	crdMetaDefinition := &v1beta1.CustomResourceDefinition{}
	crdSchema := crdService.schemaFactory(schemaName(crdMetaDefinition))

	crdSchema.GenerateCRD()

	for _, e := range crdMetaDefinition.Spec.Versions {
		err := crdSchema.GenerateCR(e.Schema.OpenAPIV3Schema)
		if err != nil {
			return nil, err
		}
	}

	if err := orm.CreateSchemaTables(crdSchema); err != nil {
		return nil, err
	}
	return crdService, nil
}
