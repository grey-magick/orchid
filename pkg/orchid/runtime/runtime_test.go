package runtime

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/assert/cmp"
)

// TestObject contains Object interface tests.
func TestObject(t *testing.T) {
	t.Run("get-object-spec", func(t *testing.T) {
		// arrange
		field := "spec"
		o := buildObjectWithFields(field, buildExampleSpec())

		// act
		got := o.GetJSONSchemaField(field)

		// assert
		require.NotNil(t, got)
		require.True(t, DeepEqual(o.GetJSONSchemaField(field), got))
	})
}

// DeepEqual verifies whether the given values are deeply equal.
func DeepEqual(x interface{}, y interface{}) bool {
	return cmp.DeepEqual(x, y)().Success()
}

// buildObjectWithFields returns a new Object filled with the given fields and values.
func buildObjectWithFields(fieldsAndValues ...interface{}) Object {
	fields := NewJSONSchemaFields(fieldsAndValues...)
	obj := NewObject()
	obj.SetJSONSchemaFields(fields)
	return obj
}

// buildExampleSpec creates an example spec node.
func buildExampleSpec() JSONSchemaNode {
	specField := make(JSONSchemaNode)
	specField["field1"] = "value1"
	return specField
}
