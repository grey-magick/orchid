package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/isutton/orchid/test/mocks"
)

func TestORM_New(t *testing.T) {
	orm := NewORM("user=postgres password=1 dbname=postgres sslmode=disable")

	t.Run("Connect", func(t *testing.T) {
		err := orm.Connect()
		assert.NoError(t, err)
	})

	t.Run("CreateSchemaTables", func(t *testing.T) {
		openAPIV3Schema := mocks.OpenAPIV3SchemaMock()
		schema := NewSchema("cr")
		err := schema.GenerateCR(&openAPIV3Schema)
		assert.NoError(t, err)

		err = orm.CreateSchemaTables(schema)
		assert.NoError(t, err)
	})
}
