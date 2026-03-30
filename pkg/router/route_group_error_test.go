package router

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithParamErrShouldReturnErrorForInvalidLocationOnRouteGroup(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	api := rtr.NewRouteGroup("/api")

	// Act
	result, err := api.WithParamErr("id", "matrix", "user identifier", "123", false)

	// Assert
	require.Error(t, err)
	assert.Equal(t, api, result)
	assert.Empty(t, api.defaultParams)
	assert.Contains(t, err.Error(), "invalid parameter 'in'")
}

func TestWithParamErrShouldReturnErrorForEmptyNameOrLocationOnRouteGroup(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	api := rtr.NewRouteGroup("/api")

	// Act
	result, err := api.WithParamErr("", "query", "user identifier", "123", false)

	// Assert
	require.Error(t, err)
	assert.Equal(t, api, result)
	assert.Empty(t, api.defaultParams)
	assert.Contains(t, err.Error(), "cannot be empty")
}

func TestWithParamErrShouldReturnErrorWhenExampleTypeIsNilOnRouteGroup(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	api := rtr.NewRouteGroup("/api")

	// Act
	result, err := api.WithParamErr("id", "query", "user identifier", nil, false)

	// Assert
	require.Error(t, err)
	assert.Equal(t, api, result)
	assert.Empty(t, api.defaultParams)
	assert.Contains(t, err.Error(), "nil type")
}

func TestWithPathParamErrShouldForceRequiredOnRouteGroup(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	api := rtr.NewRouteGroup("/api")

	// Act
	result, err := api.WithPathParamErr("id", "user identifier", "123")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, api, result)
	require.Len(t, api.defaultParams, 1)
	assert.True(t, api.defaultParams[0].Required)
	assert.Equal(t, "path", api.defaultParams[0].In)
	assert.Equal(t, "123", api.defaultParams[0].Example)
	assert.NotNil(t, api.defaultParams[0].Converter)
}

func TestShouldAccumulateValidationErrorsWithoutPanickingWhenRouteGroupSafe(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	api := rtr.NewRouteGroup("/api").Safe()

	// Act / Assert
	assert.NotPanics(t, func() {
		api.WithParam("id", "matrix", "user identifier", "123", false)
		api.WithParam("", "query", "user identifier", "123", false)
	})

	// Assert
	assert.Empty(t, api.defaultParams)
	require.Len(t, api.Errors(), 2)
	assert.ErrorContains(t, api.Err(), "invalid parameter 'in'")
	assert.ErrorContains(t, api.Err(), "cannot be empty")
	require.Len(t, rtr.Errors(), 2)
}
