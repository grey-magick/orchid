package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	jsc "github.com/isutton/orchid/pkg/orchid/jsonschema"
	"github.com/isutton/orchid/test/mocks"
)

func TestExtract_extractPath(t *testing.T) {
	cr, err := mocks.UnstructuredCRMock()
	require.NoError(t, err)

	fieldPath := []string{"spec", "simple"}
	data, err := extractPath(cr.Object, jsc.String, fieldPath)
	require.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, "11", data)
}

func TestExtract_extractCRDOpenAPIV3Schema(t *testing.T) {
	crd, err := mocks.UnstructuredCRDMock()
	require.NoError(t, err)

	openAPIV3Schema, err := ExtractCRDOpenAPIV3Schema(crd.Object)
	require.NoError(t, err)
	assert.NotNil(t, openAPIV3Schema)
}

func TestExtract_extractCRGVKFromCRD(t *testing.T) {
	cr, err := mocks.UnstructuredCRDMock()
	require.NoError(t, err)

	gvk, err := ExtractCRGVKFromCRD(cr.Object)
	require.NoError(t, err)
	assert.NotNil(t, gvk)

	assert.Equal(t, "apiextensions.k8s.io", gvk.Group)
	assert.Equal(t, "v1", gvk.Version)
	assert.Equal(t, "CustomResourceDefinition", gvk.Kind)
}
