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
	err := schema.GenerateCR(&openAPIV3Schema)
	assert.NoError(t, err)

	sqlLib := NewSQL(schema)

	t.Run("CreateTables", func(t *testing.T) {
		createTableStmts := sqlLib.CreateTables()
		assert.Len(t, createTableStmts, expectedAmountOfTables)

		for _, statement := range createTableStmts {
			t.Logf("%s;", statement)
			assert.Contains(t, statement, "create table")
		}
	})

	t.Run("Insert", func(t *testing.T) {
		insertStmts := sqlLib.Insert()
		assert.Len(t, insertStmts, expectedAmountOfTables)

		for name, statement := range insertStmts {
			t.Logf("table='%s', insert='%s'", name, statement)
			assert.Contains(t, statement, "insert into")
		}
	})

	t.Run("Select", func(t *testing.T) {
		selectStmt := sqlLib.Select()
		t.Logf("select='%s'", selectStmt)
	})
}
