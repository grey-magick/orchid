package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/isutton/orchid/test/mocks"
)

func TestSQL_New(t *testing.T) {
	properties := mocks.JSONSchemaPropsComplex()
	schema := NewSchema("cr")
	err := schema.Generate(properties)
	assert.NoError(t, err)

	sqlLib := NewSQL(schema)

	t.Run("CreateTables", func(t *testing.T) {
		tables := sqlLib.CreateTables()
		assert.Len(t, tables, 8)

		for _, statement := range tables {
			t.Logf("%s;", statement)
			assert.Contains(t, statement, "create table")
		}
	})
}
