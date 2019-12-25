package util

import (
	"io/ioutil"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/stretchr/testify/require"
	yaml2 "gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ReadAsset reads an asset from the filesystem, panicking in case of error
func ReadAsset(path string) []byte {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return b
}

func LoadObject(path string, obj interface{}) error {
	b := ReadAsset(path)
	err := yaml.Unmarshal(b, &obj)
	if err != nil {
		return err
	}
	return nil
}

func LoadUnstructured(path string) *unstructured.Unstructured {
	b := ReadAsset(path)

	obj := map[string]interface{}{}
	err := yaml.Unmarshal(b, &obj)
	if err != nil {
		panic(err)
	}
	return &unstructured.Unstructured{Object: obj}
}

func RequireYamlEqual(t *testing.T, a interface{}, b interface{}) {
	yamlActual, _ := yaml2.Marshal(a)
	yamlExpected, _ := yaml2.Marshal(b)

	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(string(yamlExpected)),
		B:        difflib.SplitLines(string(yamlActual)),
		FromFile: "Expected",
		ToFile:   "Actual",
		Context:  4,
	}

	m, err := difflib.GetUnifiedDiffString(diff)
	require.NoError(t, err)

	if len(m) > 0 {
		t.Errorf("The following differences have been found:\n%s", m)
	}

}
