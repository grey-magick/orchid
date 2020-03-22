package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/klogr"

	"github.com/isutton/orchid/pkg/orchid/config"
	"github.com/isutton/orchid/test/mocks"
)

func TestORM_New(t *testing.T) {
	logger := klogr.New().WithName("test")
	config := &config.Config{Username: "postgres", Password: "1", Options: "sslmode=disable"}
	pgORM := NewORM(logger, "postgres", "public", config)

	t.Run("Bootstrap", func(t *testing.T) {
		err := pgORM.Bootstrap()
		assert.NoError(t, err)
	})

	openAPIV3Schema := mocks.OpenAPIV3SchemaMock()
	schema := NewSchema(logger, "cr")

	t.Run("CreateSchemaTables", func(t *testing.T) {
		err := schema.Generate(&openAPIV3Schema)
		assert.NoError(t, err)

		err = pgORM.CreateTables(schema)
		assert.NoError(t, err)
	})

	// t.Run("Create", func(t *testing.T) {
	// 	arguments := mocks.RepositoryArgumentsMock()
	// 	err := orm.Create(schema, arguments)
	// 	require.NoError(t, err)
	// })

	t.Run("Read", func(t *testing.T) {
		namespacedName := types.NamespacedName{Namespace: "namespace", Name: "testing"}
		data, err := pgORM.Read(schema, namespacedName)
		assert.NoError(t, err)
		t.Logf("data='%+v'", data)
	})
}
