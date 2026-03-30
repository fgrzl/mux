package binder

import (
	"testing"

	"github.com/fgrzl/mux/pkg/openapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldParseSliceValuesGivenExampleStringSlice(t *testing.T) {
	// Arrange
	param := &openapi.ParameterObject{Example: []string{"x"}}
	// Act
	v, ok := ParseSliceValues([]string{"a", "b"}, param)
	// Assert
	assert.True(t, ok)
	arr, ok := v.([]string)
	require.True(t, ok)
	assert.Equal(t, []string{"a", "b"}, arr)
}

func TestShouldSplitCSVWhenProcessingExampleStringSliceParam(t *testing.T) {
	// Arrange
	staging := make(map[string]any)
	param := &openapi.ParameterObject{Example: []string{"x"}}
	// Act
	handled, err := ProcessParamAndSet(staging, "k", []string{"a,b,c"}, "query", param)
	// Assert
	require.NoError(t, err)
	assert.True(t, handled)
	v, ok := staging["k"]
	assert.True(t, ok)
	arr, ok := v.([]string)
	require.True(t, ok)
	assert.Equal(t, []string{"a", "b", "c"}, arr)
}

func TestShouldSplitQuotedCSVWhenProcessingExampleStringSliceParam(t *testing.T) {
	// Arrange
	staging := make(map[string]any)
	param := &openapi.ParameterObject{Example: []string{"x"}}
	// Act
	handled, err := ProcessParamAndSet(staging, "k", []string{`"a,b",c`}, "query", param)
	// Assert
	require.NoError(t, err)
	assert.True(t, handled)
	v, ok := staging["k"]
	assert.True(t, ok)
	arr, ok := v.([]string)
	require.True(t, ok)
	assert.Equal(t, []string{"a,b", "c"}, arr)
}
