package binder

import (
	"testing"

	"github.com/fgrzl/mux/pkg/openapi"
	"github.com/stretchr/testify/assert"
)

func TestParseSliceValues_ExampleStringSlice(t *testing.T) {
	param := &openapi.ParameterObject{Example: []string{"x"}}
	v, ok := ParseSliceValues([]string{"a", "b"}, param)
	assert.True(t, ok)
	arr, ok := v.([]string)
	assert.True(t, ok)
	assert.Equal(t, []string{"a", "b"}, arr)
}

func TestProcessParamAndSet_ExampleSliceCSVSplit(t *testing.T) {
	staging := make(map[string]any)
	param := &openapi.ParameterObject{Example: []string{"x"}}
	handled, err := ProcessParamAndSet(staging, "k", []string{"a,b,c"}, "query", param)
	assert.NoError(t, err)
	assert.True(t, handled)
	v, ok := staging["k"]
	assert.True(t, ok)
	arr, ok := v.([]string)
	assert.True(t, ok)
	assert.Equal(t, []string{"a", "b", "c"}, arr)
}
