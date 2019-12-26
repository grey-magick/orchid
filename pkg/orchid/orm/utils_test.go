package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUtils_InterfaceSliceContains(t *testing.T) {
	slice := []interface{}{"a", "b", "c"}
	assert.True(t, InterfaceSliceContains(slice, "b"))
	assert.False(t, InterfaceSliceContains(slice, "d"))
}

func TestUtils_StringSliceContains(t *testing.T) {
	slice := []string{"a", "b", "c"}
	assert.True(t, StringSliceContains(slice, "b"))
	assert.False(t, StringSliceContains(slice, "d"))
}

func TestUtils_InterfaceSliceReversed(t *testing.T) {
	slice := []interface{}{"a", "b", "c"}
	assert.Equal(t, []interface{}{"c", "b", "a"}, InterfaceSliceReversed(slice))
}
