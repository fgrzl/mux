package router

import (
	"net/http"
	"testing"

	"github.com/fgrzl/mux/pkg/routing"
	"github.com/fgrzl/mux/test/testhelpers"
	"github.com/stretchr/testify/assert"
)

func TestShouldCreateNewRouterWithDefaultOptions(t *testing.T) {
	// Arrange & Act
	rtr := NewRouter()

	// Assert
	assert.NotNil(t, rtr)
	assert.NotNil(t, rtr.RouteGroup)
	assert.NotNil(t, rtr.routeRegistry)
	assert.NotNil(t, rtr.options)
	assert.Equal(t, "", rtr.prefix)
}

func TestShouldCreateNewRouterWithOptions(t *testing.T) {
	// Arrange
	title := "Test API"
	version := "1.0.0"

	// Act
	rtr := NewRouter(WithTitle(title), WithVersion(version))

	// Assert
	assert.NotNil(t, rtr)
	assert.NotNil(t, rtr.options.openapi)
	assert.Equal(t, title, rtr.options.openapi.Title)
	assert.Equal(t, version, rtr.options.openapi.Version)
}

func TestShouldCreateNewRouteGroupWithPrefix(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	prefix := "/api/v1"

	// Act
	group := rtr.NewRouteGroup(prefix)

	// Assert
	assert.NotNil(t, group)
	assert.Equal(t, "/api/v1", group.prefix) // Based on normalizeRoute behavior
	assert.Equal(t, rtr.routeRegistry, group.routeRegistry)
}

func TestShouldServeHTTPAndReturnNotFoundForUnknownRoute(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	req, rec := testhelpers.NewRequestRecorder(http.MethodGet, "/unknown", nil)

	// Act
	rtr.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestShouldServeHTTPAndCallRegisteredHandler(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	called := false
	rtr.GET("/test", func(c routing.RouteContext) {
		called = true
		c.OK("success")
	})

	req, rec := testhelpers.NewRequestRecorder(http.MethodGet, "/test", nil)

	// Act
	rtr.ServeHTTP(rec, req)

	// Assert
	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestShouldServeHTTPWithMiddleware(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	middlewareExecuted := false

	// Add middleware
	rtr.middleware = append(rtr.middleware, &testMiddleware{
		invoke: func(c routing.RouteContext, next HandlerFunc) {
			middlewareExecuted = true
			next(c)
		},
	})

	handlerExecuted := false
	rtr.GET("/test", func(c routing.RouteContext) {
		handlerExecuted = true
		c.OK("success")
	})

	req, rec := testhelpers.NewRequestRecorder(http.MethodGet, "/test", nil)

	// Act
	rtr.ServeHTTP(rec, req)

	// Assert
	assert.True(t, middlewareExecuted)
	assert.True(t, handlerExecuted)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestShouldSetRouteParamsInContext(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	var receivedID string
	rtr.GET("/users/{id}", func(c routing.RouteContext) {
		receivedID, _ = c.Param("id")
		c.OK("success")
	})

	req, rec := testhelpers.NewRequestRecorder(http.MethodGet, "/users/123", nil)

	// Act
	rtr.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "123", receivedID)
}

func TestShouldExecuteMiddlewareInCorrectOrder(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	executionOrder := []int{}

	// Add middleware in order
	rtr.middleware = append(rtr.middleware, &testMiddleware{
		invoke: func(c routing.RouteContext, next HandlerFunc) {
			executionOrder = append(executionOrder, 1)
			next(c)
		},
	})
	rtr.middleware = append(rtr.middleware, &testMiddleware{
		invoke: func(c routing.RouteContext, next HandlerFunc) {
			executionOrder = append(executionOrder, 2)
			next(c)
		},
	})

	rtr.GET("/test", func(c routing.RouteContext) {
		executionOrder = append(executionOrder, 3)
		c.OK("success")
	})

	req, rec := testhelpers.NewRequestRecorder(http.MethodGet, "/test", nil)

	// Act
	rtr.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, []int{1, 2, 3}, executionOrder)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestShouldStopMiddlewareChainWhenNotContinuing(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	middlewareExecuted := false
	handlerExecuted := false

	// Add middleware that doesn't call next
	rtr.middleware = append(rtr.middleware, &testMiddleware{
		invoke: func(c routing.RouteContext, next HandlerFunc) {
			middlewareExecuted = true
			c.Unauthorized()
			// Don't call next(c)
		},
	})

	rtr.GET("/test", func(c routing.RouteContext) {
		handlerExecuted = true
		c.OK("success")
	})

	req, rec := testhelpers.NewRequestRecorder(http.MethodGet, "/test", nil)

	// Act
	rtr.ServeHTTP(rec, req)

	// Assert
	assert.True(t, middlewareExecuted)
	assert.False(t, handlerExecuted) // Handler should not be executed
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// Test middleware implementation for testing
type testMiddleware struct {
	invoke func(c routing.RouteContext, next HandlerFunc)
}

func (tm *testMiddleware) Invoke(c routing.RouteContext, next HandlerFunc) {
	tm.invoke(c, next)
}
