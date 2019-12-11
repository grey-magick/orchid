package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/isutton/orchid/test/mocks"
)

func TestSQL_New(t *testing.T) {
	openAPIV3Schema := mocks.OpenAPIV3SchemaMock()
	schema := NewSchema("cr")
	err := schema.Generate(&openAPIV3Schema)
	assert.NoError(t, err)

	sqlLib := NewSQL(schema)

	t.Run("CreateTables", func(t *testing.T) {
		tables := sqlLib.CreateTables()
		assert.Len(t, tables, 9)

		for _, statement := range tables {
			t.Logf("%s;", statement)
			assert.Contains(t, statement, "create table")
		}
	})
}
