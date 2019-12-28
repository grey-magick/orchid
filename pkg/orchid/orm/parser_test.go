package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/klog/klogr"

	jsc "github.com/isutton/orchid/pkg/orchid/jsonschema"
	"github.com/isutton/orchid/test/mocks"
)

func TestParser_Parse(t *testing.T) {
	logger := klogr.New().WithName("test")
	schemaName := "parser"
	schema := NewSchema(logger, schemaName)

	parser := NewParser(logger, schema)

	t.Run("mocked", func(t *testing.T) {
		openAPIV3Schema := mocks.OpenAPIV3SchemaMock()
		err := parser.Parse(schema.Name, Relationship{}, &openAPIV3Schema)
		assert.NoError(t, err)
	})

	t.Run("x-kubernetes-embedded-resource", func(t *testing.T) {
		openAPIV3Schema := jsc.OrchidOpenAPIV3Schema()
		err := parser.Parse(schema.Name, Relationship{}, &openAPIV3Schema)
		assert.NoError(t, err)

		table, err := parser.schema.GetTable(schemaName)
		assert.NoError(t, err)

		column := table.GetColumn("data")
		assert.NotNil(t, column)
		assert.Equal(t, PgTypeJSONB, column.Type)
	})
}
