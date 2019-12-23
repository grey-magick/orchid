package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/klog/klogr"

	"github.com/isutton/orchid/test/mocks"
)

func TestORM_New(t *testing.T) {
	logger := klogr.New().WithName("test")
	orm := NewORM(logger, "user=postgres password=1 dbname=postgres sslmode=disable")

	t.Run("Connect", func(t *testing.T) {
		err := orm.Connect()
		assert.NoError(t, err)
	})

	openAPIV3Schema := mocks.OpenAPIV3SchemaMock()
	schema := NewSchema(logger, "cr")

	t.Run("CreateSchemaTables", func(t *testing.T) {
		err := schema.GenerateCR(&openAPIV3Schema)
		assert.NoError(t, err)

		err = orm.CreateSchemaTables(schema)
		assert.NoError(t, err)
	})

	// t.Run("Create", func(t *testing.T) {
	// 	arguments := mocks.RepositoryArgumentsMock()
	// 	err := orm.Create(schema, arguments)
	// 	require.NoError(t, err)
	// })

	// t.Run("Read", func(t *testing.T) {
	// 	namespacedName := types.NamespacedName{Namespace: "namespace", Name: "testing"}
	// 	data, err := orm.Read(schema, namespacedName)
	// 	assert.NoError(t, err)
	// 	t.Logf("data='%+v'", data)
	// })
}
