package mux

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloneSliceShouldPreserveNilAndEmptySlices(t *testing.T) {
	require.Nil(t, cloneSlice[string](nil))

	cloned := cloneSlice([]string{})
	require.NotNil(t, cloned)
	assert.Empty(t, cloned)
}

func TestToInternalSecurityRequirementShouldPreserveEmptyScopes(t *testing.T) {
	converted := toInternalSecurityRequirement(SecurityRequirement{"oauth2": {}})
	require.NotNil(t, converted)

	rawScopes, ok := (*converted)["oauth2"]
	require.True(t, ok)

	scopes, ok := rawScopes.([]string)
	require.True(t, ok, "expected []string, got %T", rawScopes)
	require.NotNil(t, scopes)
	assert.Empty(t, scopes)
}
