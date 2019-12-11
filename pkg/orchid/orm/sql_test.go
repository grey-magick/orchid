package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/isutton/orchid/test/mocks"
)

func TestSQL_New(t *testing.T) {
	const expectedAmountOfTables = 9

	openAPIV3Schema := mocks.OpenAPIV3SchemaMock()
	schema := NewSchema("cr")
	err := schema.Generate(&openAPIV3Schema)
	assert.NoError(t, err)

	sqlLib := NewSQL(schema)

	t.Run("CreateTables", func(t *testing.T) {
		tables := sqlLib.CreateTables()
		assert.Len(t, tables, expectedAmountOfTables)

		for _, statement := range tables {
			t.Logf("%s;", statement)
			assert.Contains(t, statement, "create table")
		}
	})

	t.Run("Insert", func(t *testing.T) {
		inserts := sqlLib.Insert()
		assert.Len(t, inserts, expectedAmountOfTables)

		for name, insert := range inserts {
			t.Logf("table='%s', insert='%s'", name, insert)
			assert.Contains(t, insert, "insert into")
		}
	})
}
