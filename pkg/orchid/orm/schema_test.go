package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/isutton/orchid/test/mocks"
)

func TestSchema_New(t *testing.T) {
	openAPIV3Schema := mocks.OpenAPIV3SchemaMock()
	schema := NewSchema("cr")

	t.Run("Generate", func(t *testing.T) {
		err := schema.Generate(&openAPIV3Schema)
		assert.NoError(t, err)
		assert.Len(t, schema.Tables, 9)
	})
}
