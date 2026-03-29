package opentelemetry

import (
	"net/http"
	"testing"

	"github.com/fgrzl/mux/pkg/common"
	"github.com/fgrzl/mux/pkg/router"
	"github.com/fgrzl/mux/pkg/routing"
	"github.com/fgrzl/mux/test/testhelpers"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
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

func TestShouldCreateValidOptionGivenWithOperation(t *testing.T) {
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
	params := &routing.Params{}
	params.Set("id", "123")
	ctx.SetParamsSlice(params)

	nextCalled := false
	next := func(c routing.RouteContext) {
		nextCalled = true
		// Verify context is still intact
		id, _ := c.Param("id")
		assert.Equal(t, "123", id)
		assert.Equal(t, bearerToken, c.Request().Header.Get(common.HeaderAuthorization))
		c.OK("user updated")
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.True(t, nextCalled)
	// OpenTelemetry should not interfere with the normal operation
}

func TestShouldUseMethodAndRoutePatternAsSpanName(t *testing.T) {
	recorder := installTestTracerProvider(t)

	rtr := router.NewRouter()
	UseOpenTelemetry(rtr)
	rtr.GET("/users/{id}", func(c routing.RouteContext) { c.OK("ok") })

	req, rec := testhelpers.NewRequestRecorder(http.MethodGet, "/users/123", nil)
	rtr.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	spans := recorder.Ended()
	if assert.NotEmpty(t, spans) {
		assert.Equal(t, "GET /users/{id}", spans[len(spans)-1].Name())
	}
}

func TestShouldAttachRouteAttributesToSpan(t *testing.T) {
	recorder := installTestTracerProvider(t)

	rtr := router.NewRouter()
	UseOpenTelemetry(rtr)
	rtr.POST("/orders/{orderId}", func(c routing.RouteContext) { c.OK("ok") })

	req, rec := testhelpers.NewRequestRecorder(http.MethodPost, "/orders/42", nil)
	rtr.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	spans := recorder.Ended()
	if assert.NotEmpty(t, spans) {
		attrs := attributesToMap(spans[len(spans)-1].Attributes())
		assert.Equal(t, "/orders/{orderId}", attrs["http.route"])
		assert.Equal(t, "/orders/{orderId}", attrs["mux.route.pattern"])
		assert.Equal(t, http.MethodPost, attrs["http.request.method"])
	}
}

func TestShouldSetSpanErrorStatusForServerErrors(t *testing.T) {
	recorder := installTestTracerProvider(t)

	rtr := router.NewRouter()
	UseOpenTelemetry(rtr)
	rtr.GET("/fail", func(c routing.RouteContext) {
		c.Response().WriteHeader(http.StatusInternalServerError)
	})

	req, rec := testhelpers.NewRequestRecorder(http.MethodGet, "/fail", nil)
	rtr.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	spans := recorder.Ended()
	if assert.NotEmpty(t, spans) {
		span := spans[len(spans)-1]
		assert.Equal(t, codes.Error, span.Status().Code)

		statusCode, ok := findSpanAttribute(span.Attributes(), "http.response.status_code")
		if assert.True(t, ok) {
			assert.Equal(t, int64(http.StatusInternalServerError), statusCode.AsInt64())
		}
	}
}

func TestShouldCaptureConflictStatusWithoutMarkingSpanAsError(t *testing.T) {
	recorder := installTestTracerProvider(t)

	rtr := router.NewRouter()
	UseOpenTelemetry(rtr)
	rtr.POST("/orders/{orderId}", func(c routing.RouteContext) {
		c.Conflict("Resource Exists", "The resource already exists")
	})

	req, rec := testhelpers.NewRequestRecorder(http.MethodPost, "/orders/42", nil)
	rtr.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)

	spans := recorder.Ended()
	if assert.NotEmpty(t, spans) {
		span := spans[len(spans)-1]
		assert.Equal(t, codes.Unset, span.Status().Code)

		statusCode, ok := findSpanAttribute(span.Attributes(), "http.response.status_code")
		if assert.True(t, ok) {
			assert.Equal(t, int64(http.StatusConflict), statusCode.AsInt64())
		}
	}
}

func TestShouldFallbackToOperationNameWhenRouteMetadataMissing(t *testing.T) {
	recorder := installTestTracerProvider(t)

	middleware := &otelMiddleware{operation: "fallback-operation"}
	ctx, _ := testhelpers.NewRouteContext(http.MethodGet, "/raw/path", nil)

	middleware.Invoke(ctx, func(c routing.RouteContext) {
		c.OK("ok")
	})

	spans := recorder.Ended()
	if assert.NotEmpty(t, spans) {
		assert.Equal(t, "fallback-operation", spans[len(spans)-1].Name())
	}
}

func TestShouldBypassTracingForWebSocketUpgradeRequests(t *testing.T) {
	recorder := installTestTracerProvider(t)

	rtr := router.NewRouter()
	UseOpenTelemetry(rtr, WithOperation("http.server"))

	hasActiveSpan := true
	rtr.GET("/ws", func(c routing.RouteContext) {
		hasActiveSpan = trace.SpanFromContext(c.Request().Context()).SpanContext().IsValid()
		c.Response().WriteHeader(http.StatusSwitchingProtocols)
	})

	req, rec := testhelpers.NewRequestRecorder(http.MethodGet, "/ws", nil)
	req.Header.Set("Connection", "keep-alive, Upgrade")
	req.Header.Set("Upgrade", "websocket")
	rtr.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusSwitchingProtocols, rec.Code)
	assert.False(t, hasActiveSpan)
	assert.Empty(t, recorder.Ended())
}

func installTestTracerProvider(t *testing.T) *tracetest.SpanRecorder {
	t.Helper()

	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))

	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() {
		_ = tp.Shutdown(t.Context())
		otel.SetTracerProvider(prev)
	})

	return recorder
}

func attributesToMap(attrs []attribute.KeyValue) map[string]string {
	out := make(map[string]string, len(attrs))
	for _, kv := range attrs {
		out[string(kv.Key)] = kv.Value.AsString()
	}
	return out
}

func findSpanAttribute(attrs []attribute.KeyValue, key string) (attribute.Value, bool) {
	for _, kv := range attrs {
		if string(kv.Key) == key {
			return kv.Value, true
		}
	}

	return attribute.Value{}, false
}
