package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fgrzl/mux/pkg/builder"
	"github.com/fgrzl/mux/pkg/routing"
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

func TestConfigureShouldReturnValidationErrorsOnRouteGroupWithoutChangingDefaultPanicBehavior(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	api := rtr.NewRouteGroup("/api")

	// Act
	err := api.Configure(func(group *RouteGroup) {
		group.WithParam("id", "matrix", "user identifier", "123", false)
	})

	// Assert
	require.Error(t, err)
	assert.ErrorContains(t, err, "invalid parameter 'in'")
	require.Len(t, api.Errors(), 1)
	require.Len(t, rtr.Errors(), 1)
	assert.Panics(t, func() {
		api.WithParam("id", "matrix", "user identifier", "123", false)
	})
}

func TestShouldRejectDuplicateRouteRegistrationWithoutOverwritingExistingHandler(t *testing.T) {
	// Arrange
	rtr := NewRouter().Safe()
	api := rtr.NewRouteGroup("/api")
	api.GET("/users", func(c routing.RouteContext) {
		c.Response().Header().Set("X-Handler", "first")
		c.OK("first")
	})

	// Act
	duplicate := api.GET("/users", func(c routing.RouteContext) {
		c.Response().Header().Set("X-Handler", "second")
		c.OK("second")
	})
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rr := httptest.NewRecorder()
	rtr.ServeHTTP(rr, req)

	// Assert
	require.Error(t, duplicate.Err())
	assert.ErrorContains(t, duplicate.Err(), "already registered")
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "first", rr.Header().Get("X-Handler"))
	require.Len(t, rtr.Errors(), 1)
}

func TestShouldNotServeRouteAfterBuilderValidationFailsInSafeMode(t *testing.T) {
	// Arrange
	rtr := NewRouter().Safe()
	api := rtr.NewRouteGroup("/api")

	// Act
	route := api.GET("/users", func(c routing.RouteContext) {
		c.OK("users")
	}).WithOperationID("invalid-id")
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rr := httptest.NewRecorder()
	rtr.ServeHTTP(rr, req)

	// Assert
	require.Error(t, route.Err())
	assert.ErrorContains(t, route.Err(), "invalid OperationID")
	assert.Equal(t, http.StatusNotFound, rr.Code)
	require.Len(t, rtr.Errors(), 1)
}

func TestShouldNotAttachDetachedRouteBuilderWithExistingValidationErrors(t *testing.T) {
	// Arrange
	rtr := NewRouter().Safe()
	api := rtr.NewRouteGroup("/api")
	detached := builder.DetachedRoute(http.MethodGet, "/users").Safe()
	detached.WithOperationID("invalid-id")

	// Act
	attached := api.HandleRoute(detached, func(c routing.RouteContext) {
		c.OK("users")
	})
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rr := httptest.NewRecorder()
	rtr.ServeHTTP(rr, req)

	// Assert
	require.Error(t, attached.Err())
	assert.ErrorContains(t, attached.Err(), "invalid OperationID")
	assert.Equal(t, http.StatusNotFound, rr.Code)
	require.Len(t, rtr.Errors(), 1)
}
