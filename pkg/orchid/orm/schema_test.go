package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/isutton/orchid/test/mocks"
)

func TestSchema_New(t *testing.T) {
	properties := mocks.JSONSchemaPropsComplex()
	schema := NewSchema("cr")

	t.Run("Generate", func(t *testing.T) {
		err := schema.Generate(properties)
		assert.NoError(t, err)
		assert.Equal(t, 8, len(schema.Tables))
	})
}
