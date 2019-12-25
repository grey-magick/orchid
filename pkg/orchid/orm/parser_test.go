package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/klog/klogr"

	"github.com/isutton/orchid/test/mocks"
)

func TestParser_New(t *testing.T) {
	logger := klogr.New().WithName("test")
	schema := NewSchema(logger, "parser")

	parser := NewParser(logger, schema)

	openAPIV3Schema := mocks.OpenAPIV3SchemaMock()
	err := parser.Parse(schema.Name, Relationship{}, &openAPIV3Schema)
	assert.NoError(t, err)
}
