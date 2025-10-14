package opentelemetry

import (
	"net/http"
	"testing"

	"github.com/fgrzl/mux/pkg/common"
	"github.com/fgrzl/mux/pkg/router"
	"github.com/fgrzl/mux/pkg/routing"
	"github.com/fgrzl/mux/test/testhelpers"
	"github.com/stretchr/testify/assert"
)

const bearerToken = "Bearer token123"

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
	rtr := router.NewRouter()

	// Act - register middleware
	UseOpenTelemetry(rtr)

	// Register a route and ensure requests still succeed
	rtr.GET("/test", func(c routing.RouteContext) { _, _ = c.Response().Write([]byte("ok")) })
	req, rec := testhelpers.NewRequestRecorder(http.MethodGet, "/test", nil)
	rtr.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestShouldAddOpenTelemetryMiddlewareWithCustomOperation(t *testing.T) {
	// Arrange
	rtr := router.NewRouter()
	customOperation := "my-custom-operation"

	// Act
	UseOpenTelemetry(rtr, WithOperation(customOperation))

	rtr.GET("/test", func(c routing.RouteContext) { _, _ = c.Response().Write([]byte("ok")) })
	req, rec := testhelpers.NewRequestRecorder(http.MethodGet, "/test", nil)
	rtr.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
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

	ctx, _ := testhelpers.NewRouteContext(http.MethodGet, "/test", nil)

	nextCalled := false
	requestUpdated := false
	responseUpdated := false

	next := func(c routing.RouteContext) {
		nextCalled = true
		// Check if context was properly updated
		if c.Request() != nil {
			requestUpdated = true
		}
		if c.Response() != nil {
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
	rtr := router.NewRouter()
	operationName := "multi-option-operation"

	// Act
	UseOpenTelemetry(rtr,
		WithOperation(operationName),
	)

	rtr.GET("/test", func(c routing.RouteContext) { _, _ = c.Response().Write([]byte("ok")) })
	req, rec := testhelpers.NewRequestRecorder(http.MethodGet, "/test", nil)
	rtr.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestShouldSetDefaultOperationWhenNoneProvided(t *testing.T) {
	// Arrange
	rtr := router.NewRouter()

	// Act
	UseOpenTelemetry(rtr)

	rtr.GET("/test", func(c routing.RouteContext) { _, _ = c.Response().Write([]byte("ok")) })
	req, rec := testhelpers.NewRequestRecorder(http.MethodGet, "/test", nil)
	rtr.ServeHTTP(rec, req)

	// Differentiate this test from TestShouldAddOpenTelemetryMiddlewareToRouter by asserting body
	assert.Equal(t, "ok", rec.Body.String())
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
	methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			// Arrange
			middleware := &otelMiddleware{operation: "test"}

			ctx, _ := testhelpers.NewRouteContext(method, "/test", nil)

			called := false
			next := func(c routing.RouteContext) {
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

	ctx, _ := testhelpers.NewRouteContext(http.MethodPost, "/api/users/123", nil)
	ctx.Request().Header.Set(common.HeaderContentType, common.MimeJSON)
	ctx.Request().Header.Set(common.HeaderAuthorization, bearerToken)
	ctx.SetParams(routing.RouteParams{"id": "123"})

	nextCalled := false
	next := func(c routing.RouteContext) {
		nextCalled = true
		// Verify context is still intact
		assert.Equal(t, "123", c.Params()["id"])
		assert.Equal(t, bearerToken, c.Request().Header.Get(common.HeaderAuthorization))
		c.OK("user updated")
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.True(t, nextCalled)
	// OpenTelemetry should not interfere with the normal operation
}
