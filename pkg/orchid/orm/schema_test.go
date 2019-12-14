package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	extv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	"github.com/isutton/orchid/test/mocks"
)

func assertJsonSchemaVsORMSchema(
	t *testing.T,
	schema *Schema,
	table *Table,
	properties map[string]extv1beta1.JSONSchemaProps,
) {
	t.Logf("Inspecting table '%s' with '%d' properties", table.Name, len(properties))
	for name, jsonSchema := range properties {
		if jsonSchema.Type == JSTypeObject {
			tableName := schema.TableName(name)
			t.Logf("table-name='%s'", tableName)
			require.NotEmpty(t, tableName)

			objectTable := schema.GetTable(tableName)
			if objectTable == nil {
				continue
			}

			assert.True(t, len(objectTable.ColumNames()) >= len(jsonSchema.Properties))
			assert.NotNil(t, jsonSchema.Properties)

			assertJsonSchemaVsORMSchema(t, schema, objectTable, jsonSchema.Properties)
		} else {
			column := table.GetColumn(name)
			assert.NotNil(t, column)
		}
	}
}

func TestSchema_CR(t *testing.T) {
	openAPIV3Schema := mocks.OpenAPIV3SchemaMock()
	schemaName := "cr"
	schema := NewSchema(schemaName)

	t.Run("GenerateCR", func(t *testing.T) {
		err := schema.GenerateCR(&openAPIV3Schema)

		assert.NoError(t, err)
		assert.True(t, len(schema.Tables) > 1)

		table := schema.GetTable(schemaName)
		assert.NotNil(t, table)
		assertJsonSchemaVsORMSchema(t, schema, table, openAPIV3Schema.Properties)
	})

	t.Run("TablesReversed", func(t *testing.T) {
		reversed := schema.TablesReversed()
		reversedLen := len(reversed)

		assert.Equal(t, reversedLen, len(schema.Tables))
		assert.Equal(t, schema.Tables[0].Name, reversed[reversedLen-1].Name)
	})
}

func TestSchema_CRD(t *testing.T) {
	const expectedAmountOfTables = 1

	schema := NewSchema("crd")

	t.Run("GenerateCRD", func(t *testing.T) {
		schema.GenerateCRD()

		assert.Len(t, schema.Tables, expectedAmountOfTables)
	})
}
