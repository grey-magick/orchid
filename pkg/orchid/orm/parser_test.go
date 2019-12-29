package orm

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/klog/klogr"

	jsc "github.com/isutton/orchid/pkg/orchid/jsonschema"
	"github.com/isutton/orchid/test/mocks"
)

func TestParser_Parse(t *testing.T) {
	logger := klogr.New().WithName("test")
	schemaName := "parser"

	t.Run("mocked", func(t *testing.T) {
		schema := NewSchema(logger, schemaName)
		parser := NewParser(logger, schema)

		openAPIV3Schema := mocks.OpenAPIV3SchemaMock()
		err := parser.Parse(schema.Name, Relationship{}, &openAPIV3Schema)
		assert.NoError(t, err)
	})

	t.Run("x-kubernetes-embedded-resource", func(t *testing.T) {
		schema := NewSchema(logger, schemaName)
		parser := NewParser(logger, schema)

		openAPIV3Schema := jsc.ExtV1CRDOpenAPIV3Schema()
		err := parser.Parse(schema.Name, Relationship{}, &openAPIV3Schema)
		assert.NoError(t, err)

		table, err := parser.schema.GetTable(schemaName)
		assert.NoError(t, err)

		column := table.GetColumn("data")
		assert.NotNil(t, column)
		assert.Equal(t, PgTypeJSONB, column.Type)
	})

	t.Run("x-list-map-keys", func(t *testing.T) {
		schema := NewSchema(logger, schemaName)
		parser := NewParser(logger, schema)

		openAPIV3Schema := jsc.MetaV1ObjectMetaOpenAPIV3Schema()
		require.Len(t, openAPIV3Schema.XListMapKeys, 2)

		err := parser.Parse(schema.Name, Relationship{}, &openAPIV3Schema)
		assert.NoError(t, err)

		table, err := parser.schema.GetTable(schemaName)
		assert.NoError(t, err)

		uniqueColumns := strings.Join(openAPIV3Schema.XListMapKeys, ",")
		found := false
		for _, constraint := range table.Constraints {
			if constraint.Type != PgConstraintUnique {
				continue
			}
			if uniqueColumns == constraint.ColumnName {
				t.Logf("unique constraint: '%s'", constraint.String())
				found = true
			}
		}
		require.True(t, found)
	})
}
