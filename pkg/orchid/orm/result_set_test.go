package orm

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/klog/klogr"

	"github.com/isutton/orchid/test/mocks"
)

func mockedColumnIDs(schema *Schema) map[string]int {
	columnIDs := map[string]int{}
	id := 0
	for _, table := range schema.Tables {
		for _, column := range table.Columns {
			columnIDs[fmt.Sprintf("%s.%s", table.Hint, column.Name)] = id
			id++
		}
	}
	return columnIDs
}

func mockedMatrix(columnIDs map[string]int) []List {
	list := List{}
	for range columnIDs {
		list = append(list, 0)
	}
	return []List{list}
}

func TestResultSet_New(t *testing.T) {
	logger := klogr.New().WithName("test")
	schema := NewSchema(logger, "result_set")

	openAPIV3Schema := mocks.OpenAPIV3SchemaMock()
	err := schema.Generate(&openAPIV3Schema)
	assert.NoError(t, err)

	columnIDs := mockedColumnIDs(schema)
	matrix := mockedMatrix(columnIDs)

	rs, err := NewResultSet(schema, columnIDs, matrix)
	assert.NoError(t, err)
	assert.NotNil(t, rs)

	values, err := rs.Get(schema.Name, "id", 0)
	assert.NoError(t, err)
	assert.Len(t, values, 1)

	pk, err := rs.GetPK(schema.Name, 0)
	assert.NoError(t, err)
	assert.Len(t, pk, 5)

	assert.Equal(t, len(rs.Data), len(schema.Tables))
}
