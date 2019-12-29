package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/klog/klogr"

	jsc "github.com/isutton/orchid/pkg/orchid/jsonschema"
	"github.com/isutton/orchid/test/mocks"
)

func assertJsonSchemaVsORMSchema(
	t *testing.T,
	schema *Schema,
	table *Table,
	properties map[string]extv1.JSONSchemaProps,
) {
	t.Logf("Inspecting table '%s' with '%d' properties", table.Name, len(properties))

	// expects to find more than one column, aside of PK
	assert.True(t, len(table.ColumNames()) > 0)

	for name, jsonSchema := range properties {
		if jsonSchema.Type != jsc.Object || jsonSchema.Properties == nil {
			continue
		}

		tableName := schema.TableName(name)
		t.Logf("table-name='%s'", tableName)
		require.NotEmpty(t, tableName)

		objectTable, err := schema.GetTable(tableName)
		if err != nil {
			continue
		}

		assert.True(t, len(objectTable.ColumNames()) >= 2)
		assertJsonSchemaVsORMSchema(t, schema, objectTable, jsonSchema.Properties)
	}
}

func TestSchema_CR(t *testing.T) {
	apiSchema := mocks.OpenAPIV3SchemaMock()
	schemaName := "cr"

	logger := klogr.New().WithName("test")
	schema := NewSchema(logger, schemaName)

	t.Run("Generate", func(t *testing.T) {
		err := schema.Generate(&apiSchema)

		assert.NoError(t, err)
		assert.True(t, len(schema.Tables) > 1)

		table, err := schema.GetTable(schemaName)
		assert.NoError(t, err)
		assert.NotNil(t, table)
		assertJsonSchemaVsORMSchema(t, schema, table, apiSchema.Properties)
	})

	t.Run("TablesReversed", func(t *testing.T) {
		reversed := schema.TablesReversed()
		reversedLen := len(reversed)

		assert.Equal(t, reversedLen, len(schema.Tables))
		assert.Equal(t, schema.Tables[0].Name, reversed[reversedLen-1].Name)
	})
}

func TestSchema_ObjectMeta(t *testing.T) {
	logger := klogr.New().WithName("test")

	schemaName := "metadata"
	schema := NewSchema(logger, schemaName)

	jsonSchema := jsc.MetaV1ObjectMetaOpenAPIV3Schema()
	err := schema.Generate(&jsonSchema)
	require.NoError(t, err)
	assert.True(t, len(schema.Tables) > 1)

	table, err := schema.GetTable(schemaName)
	assert.NoError(t, err)
	assert.NotNil(t, table)

	assertJsonSchemaVsORMSchema(t, schema, table, jsonSchema.Properties)
}

// TestSchema_CRD will generate the schema that Orchid will be using to store its own data.
func TestSchema_CRD(t *testing.T) {
	logger := klogr.New().WithName("test")

	schemaName := "orchid"
	schema := NewSchema(logger, schemaName)

	jsonSchema := jsc.ExtV1CRDOpenAPIV3Schema()
	err := schema.Generate(&jsonSchema)
	require.NoError(t, err)
	assert.True(t, len(schema.Tables) > 1)
}
