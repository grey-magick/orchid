package runtime

import (
	"io/ioutil"
	"testing"

	"github.com/ghodss/yaml"
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

// readAsset reads an asset from the filesystem, panicking in case of error
func readAsset(path string) []byte {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return b
}

// Test_object_UnmarshalJSON verifies whether sample manifests can be unmarshalled to an Orchid
// object.
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
			args{data: readAsset("cr.yaml")},
			false,
		},
		{
			"crd",
			args{data: readAsset("crd.yaml")},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &object{}
			err := yaml.Unmarshal(tt.args.data, o)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
