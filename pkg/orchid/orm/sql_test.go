package orm

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/isutton/orchid/test/mocks"
)

func TestSQL_New(t *testing.T) {
	openAPIV3Schema := mocks.OpenAPIV3SchemaMock()
	schema := NewSchema("cr")

	err := schema.GenerateCR(&openAPIV3Schema)
	assert.NoError(t, err)

	expectedAmountOfTables := len(schema.Tables)

	sqlLib := NewSQL(schema)

	t.Run("CreateTables", func(t *testing.T) {
		createTableStmts := sqlLib.CreateTables()
		assert.True(t, len(createTableStmts) >= expectedAmountOfTables)

		for _, statement := range createTableStmts {
			t.Logf("%s;", statement)
			assert.True(
				t,
				strings.Contains(statement, "create table") ||
					strings.Contains(statement, "alter table"),
			)
		}
	})

	t.Run("Insert", func(t *testing.T) {
		insertStmts := sqlLib.Insert()
		assert.Len(t, insertStmts, expectedAmountOfTables)

		for _, statement := range insertStmts {
			t.Logf("insert='%#v'", statement)
			assert.Contains(t, statement, "insert into")
		}
	})

	t.Run("Select", func(t *testing.T) {
		selectStmt := sqlLib.Select()
		t.Logf("select='%s'", selectStmt)
	})
}
