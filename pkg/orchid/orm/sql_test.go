package orm

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/klog/klogr"

	"github.com/isutton/orchid/test/mocks"
)

func TestSQL(t *testing.T) {
	logger := klogr.New().WithName("test")

	openAPIV3Schema := mocks.OpenAPIV3SchemaMock()
	schema := NewSchema(logger, "cr")

	err := schema.GenerateCR(&openAPIV3Schema)
	assert.NoError(t, err)

	expectedAmountOfTables := len(schema.Tables)

	t.Run("CreateTables", func(t *testing.T) {
		createTableStmts := CreateTablesStatement(schema)
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
		insertStmts := InsertStatement(schema)
		assert.Len(t, insertStmts, expectedAmountOfTables)

		for _, statement := range insertStmts {
			t.Logf("insert='%#v'", statement)
			assert.Contains(t, statement, "insert into")
		}
	})

	t.Run("Select", func(t *testing.T) {
		selectStmt := SelectStatement(schema, nil)
		t.Logf("select='%s'", selectStmt)
	})
}
