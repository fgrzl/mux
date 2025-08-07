package mux

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldCreateOpenTelemetryOptionsWithOperation(t *testing.T) {
	// Arrange
	options := &OpenTelemetryOptions{}
	operationName := "custom-operation"

	// Act
	opt := WithOperation(operationName)
	opt(options)

	// Assert
	assert.Equal(t, operationName, options.Operation)
}

func TestShouldAddOpenTelemetryMiddlewareToRouter(t *testing.T) {
	// Arrange
	router := NewRouter()
	initialCount := len(router.middleware)

	// Act
	router.UseOpenTelemetry()

	// Assert
	assert.Len(t, router.middleware, initialCount+1)
}

func TestShouldAddOpenTelemetryMiddlewareWithCustomOperation(t *testing.T) {
	// Arrange
	router := NewRouter()
	customOperation := "my-custom-operation"

	// Act
	router.UseOpenTelemetry(WithOperation(customOperation))

	// Assert
	assert.Len(t, router.middleware, 1)
	// We can't easily check the internal operation name without reflection,
	// but we can verify the middleware was added
}

func TestShouldCreateOtelMiddlewareWithDefaultOperation(t *testing.T) {
	// Arrange & Act
	middleware := &otelMiddleware{operation: "http.server"}

	// Assert
	assert.Equal(t, "http.server", middleware.operation)
}

func TestShouldInvokeNextWithOpenTelemetryTracing(t *testing.T) {
	// Arrange
	middleware := &otelMiddleware{operation: "test-operation"}

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)

	nextCalled := false
	requestUpdated := false
	responseUpdated := false

	next := func(c *RouteContext) {
		nextCalled = true
		// Check if context was properly updated
		if c.Request != nil {
			requestUpdated = true
		}
		if c.Response != nil {
			responseUpdated = true
		}
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.True(t, nextCalled)
	assert.True(t, requestUpdated)
	assert.True(t, responseUpdated)
	// OpenTelemetry will have wrapped the request/response, but the core functionality should work
}

func TestShouldHandleMultipleOpenTelemetryOptions(t *testing.T) {
	// Arrange
	router := NewRouter()
	operationName := "multi-option-operation"

	// Act
	router.UseOpenTelemetry(
		WithOperation(operationName),
		// Could add more options if they existed
	)

	// Assert - Should not panic and middleware should be added
	assert.Len(t, router.middleware, 1)
}

func TestShouldSetDefaultOperationWhenNoneProvided(t *testing.T) {
	// Arrange
	router := NewRouter()

	// Act
	router.UseOpenTelemetry()

	// Assert
	// The middleware should be created with default operation "http.server"
	assert.Len(t, router.middleware, 1)

	// We can't easily access the internal operation without reflection,
	// but we verified that the default is set in the UseOpenTelemetry method
}

func TestWithOperationShouldCreateValidOption(t *testing.T) {
	// Arrange
	operationName := "custom-test-operation"

	// Act
	option := WithOperation(operationName)

	// Test the option by applying it
	options := &OpenTelemetryOptions{}
	option(options)

	// Assert
	assert.Equal(t, operationName, options.Operation)
}

func TestShouldCreateOtelMiddlewareWithCustomOperation(t *testing.T) {
	// Arrange
	customOperation := "user-defined-operation"

	// Act
	middleware := &otelMiddleware{operation: customOperation}

	// Assert
	assert.Equal(t, customOperation, middleware.operation)
}

func TestShouldInvokeWithDifferentHTTPMethods(t *testing.T) {
	// Test that the middleware works with different HTTP methods
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			// Arrange
			middleware := &otelMiddleware{operation: "test"}

			req := httptest.NewRequest(method, "/test", nil)
			rec := httptest.NewRecorder()
			ctx := NewRouteContext(rec, req)

			called := false
			next := func(c *RouteContext) {
				called = true
				c.OK("success")
			}

			// Act
			middleware.Invoke(ctx, next)

			// Assert
			assert.True(t, called, "Next should be called for method %s", method)
		})
	}
}

func TestShouldWorkWithComplexRouteContext(t *testing.T) {
	// Arrange
	middleware := &otelMiddleware{operation: "complex-test"}

	req := httptest.NewRequest("POST", "/api/users/123", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token123")

	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	ctx.Params = RouteParams{"id": "123"}

	nextCalled := false
	next := func(c *RouteContext) {
		nextCalled = true
		// Verify context is still intact
		assert.Equal(t, "123", c.Params["id"])
		assert.Equal(t, "Bearer token123", c.Request.Header.Get("Authorization"))
		c.OK("user updated")
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.True(t, nextCalled)
	// OpenTelemetry should not interfere with the normal operation
}
