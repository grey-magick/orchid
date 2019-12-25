package runtime

import (
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/require"
	"gotest.tools/assert/cmp"

	"github.com/isutton/orchid/test/util"
)

// TestObject contains Object interface tests.
func TestObject(t *testing.T) {
	t.Run("get-DefaultObject-spec", func(t *testing.T) {
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

// Test_object_UnmarshalJSON verifies whether sample manifests can be unmarshalled to an Orchid
// DefaultObject.
func Test_object_UnmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	var tests = []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"cr",
			args{data: util.ReadAsset("../../../test/crds/cr.yaml")},
			false,
		},
		{
			"crd",
			args{data: util.ReadAsset("../../../test/crds/crd.yaml")},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &DefaultObject{}
			err := yaml.Unmarshal(tt.args.data, o)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
