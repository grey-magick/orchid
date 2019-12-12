package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/isutton/orchid/test/mocks"
)

func TestSchema_CR(t *testing.T) {
	const expectedAmountOfTables = 9

	openAPIV3Schema := mocks.OpenAPIV3SchemaMock()
	schema := NewSchema("cr")

	t.Run("GenerateCR", func(t *testing.T) {
		err := schema.GenerateCR(&openAPIV3Schema)

		assert.NoError(t, err)
		assert.Len(t, schema.Tables, expectedAmountOfTables)

		table := schema.TableFactory("cr_spec_complex_complex_nested")
		column := table.GetColumn("attribute")

		assert.NotNil(t, column)
		assert.True(t, column.NotNull)
	})

	t.Run("TablesReversed", func(t *testing.T) {
		reversed := schema.TablesReversed()
		reversedLen := len(reversed)

		assert.Equal(t, reversedLen, expectedAmountOfTables)
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
