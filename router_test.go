package mux

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldCreateNewRouterWithDefaultOptions(t *testing.T) {
	// Arrange & Act
	router := NewRouter()

	// Assert
	assert.NotNil(t, router)
	assert.NotNil(t, router.RouteGroup)
	assert.NotNil(t, router.registry)
	assert.NotNil(t, router.options)
	assert.Equal(t, "", router.prefix)
}

func TestShouldCreateNewRouterWithOptions(t *testing.T) {
	// Arrange
	title := "Test API"
	version := "1.0.0"

	// Act
	router := NewRouter(WithTitle(title), WithVersion(version))

	// Assert
	assert.NotNil(t, router)
	assert.NotNil(t, router.options.openapi)
	assert.Equal(t, title, router.options.openapi.Title)
	assert.Equal(t, version, router.options.openapi.Version)
}

func TestShouldCreateNewRouteGroupWithPrefix(t *testing.T) {
	// Arrange
	router := NewRouter()
	prefix := "/api/v1"

	// Act
	group := router.NewRouteGroup(prefix)

	// Assert
	assert.NotNil(t, group)
	assert.Equal(t, "/api/v1", group.prefix) // Based on normalizeRoute behavior
	assert.Equal(t, router.registry, group.registry)
}

func TestShouldServeHTTPAndReturnNotFoundForUnknownRoute(t *testing.T) {
	// Arrange
	router := NewRouter()
	req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	rec := httptest.NewRecorder()

	// Act
	router.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestShouldServeHTTPAndCallRegisteredHandler(t *testing.T) {
	// Arrange
	router := NewRouter()
	called := false
	router.GET("/test", func(c RouteContext) {
		called = true
		c.OK("success")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Act
	router.ServeHTTP(rec, req)

	// Assert
	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestShouldServeHTTPWithMiddleware(t *testing.T) {
	// Arrange
	router := NewRouter()
	middlewareExecuted := false

	// Add middleware
	router.middleware = append(router.middleware, &testMiddleware{
		invoke: func(c RouteContext, next HandlerFunc) {
			middlewareExecuted = true
			next(c)
		},
	})

	handlerExecuted := false
	router.GET("/test", func(c RouteContext) {
		handlerExecuted = true
		c.OK("success")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Act
	router.ServeHTTP(rec, req)

	// Assert
	assert.True(t, middlewareExecuted)
	assert.True(t, handlerExecuted)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestShouldSetRouteParamsInContext(t *testing.T) {
	// Arrange
	router := NewRouter()
	var receivedParams RouteParams
	router.GET("/users/{id}", func(c RouteContext) {
		receivedParams = c.Params()
		c.OK("success")
	})

	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	rec := httptest.NewRecorder()

	// Act
	router.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotNil(t, receivedParams)
	assert.Equal(t, "123", receivedParams["id"])
}

func TestShouldExecuteMiddlewareInCorrectOrder(t *testing.T) {
	// Arrange
	router := NewRouter()
	executionOrder := []int{}

	// Add middleware in order
	router.middleware = append(router.middleware, &testMiddleware{
		invoke: func(c RouteContext, next HandlerFunc) {
			executionOrder = append(executionOrder, 1)
			next(c)
		},
	})
	router.middleware = append(router.middleware, &testMiddleware{
		invoke: func(c RouteContext, next HandlerFunc) {
			executionOrder = append(executionOrder, 2)
			next(c)
		},
	})

	router.GET("/test", func(c RouteContext) {
		executionOrder = append(executionOrder, 3)
		c.OK("success")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Act
	router.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, []int{1, 2, 3}, executionOrder)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestShouldStopMiddlewareChainWhenNotContinuing(t *testing.T) {
	// Arrange
	router := NewRouter()
	middlewareExecuted := false
	handlerExecuted := false

	// Add middleware that doesn't call next
	router.middleware = append(router.middleware, &testMiddleware{
		invoke: func(c RouteContext, next HandlerFunc) {
			middlewareExecuted = true
			c.Unauthorized()
			// Don't call next(c)
		},
	})

	router.GET("/test", func(c RouteContext) {
		handlerExecuted = true
		c.OK("success")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Act
	router.ServeHTTP(rec, req)

	// Assert
	assert.True(t, middlewareExecuted)
	assert.False(t, handlerExecuted) // Handler should not be executed
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// Test middleware implementation for testing
type testMiddleware struct {
	invoke func(c RouteContext, next HandlerFunc)
}

func (tm *testMiddleware) Invoke(c RouteContext, next HandlerFunc) {
	tm.invoke(c, next)
}
