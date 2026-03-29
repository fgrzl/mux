package logging

import (
	"bytes"
	"log/slog"
	"net/http"
	"testing"

	"github.com/fgrzl/mux/pkg/common"
	"github.com/fgrzl/mux/pkg/router"
	"github.com/fgrzl/mux/pkg/routing"
	"github.com/fgrzl/mux/test/testhelpers"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
)

func newTestLogger(minLevel slog.Level) (*bytes.Buffer, *slog.Logger) {
	var logBuffer bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuffer, &slog.HandlerOptions{Level: minLevel}))
	return &logBuffer, logger
}

func TestLoggingMiddlewareShouldLogRequestDetails(t *testing.T) {
	// Arrange
	logBuffer, logger := newTestLogger(slog.LevelDebug)
	slog.SetDefault(logger)

	middleware := &loggingMiddleware{options: &LoggingOptions{}}
	ctx, _ := testhelpers.NewRouteContext(http.MethodGet, "/test?param=value", nil)
	ctx.Request().RemoteAddr = "192.168.1.1:8080"
	const testAgent = "test-agent/1.0"
	ctx.Request().Header.Set(common.HeaderUserAgent, testAgent)

	nextCalled := false
	next := func(c routing.RouteContext) {
		nextCalled = true
		c.Response().WriteHeader(http.StatusOK)
		_, _ = c.Response().Write([]byte("test response"))
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.True(t, nextCalled, "next handler should be called")

	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "level=DEBUG")
	assert.Contains(t, logOutput, "GET /test -> 200")
	assert.Contains(t, logOutput, "method=GET")
	// Slog text handler may wrap values in quotes; assert on stable substrings instead
	assert.Contains(t, logOutput, "/test")
	assert.Contains(t, logOutput, "status=200")
	assert.Contains(t, logOutput, "remote=192.168.1.1:8080")
	assert.Contains(t, logOutput, testAgent)
	assert.Contains(t, logOutput, "duration=")
}

func TestLoggingMiddlewareShouldCaptureStatusCode(t *testing.T) {
	// Arrange
	logBuffer, logger := newTestLogger(slog.LevelDebug)
	slog.SetDefault(logger)

	middleware := &loggingMiddleware{options: &LoggingOptions{}}
	ctx, _ := testhelpers.NewRouteContext(http.MethodPost, "/test", nil)

	next := func(c routing.RouteContext) {
		c.Response().WriteHeader(http.StatusCreated)
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "level=DEBUG")
	assert.Contains(t, logOutput, "status=201")
}

func TestLoggingMiddlewareShouldUseWarningLevelForClientErrors(t *testing.T) {
	// Arrange
	logBuffer, logger := newTestLogger(slog.LevelInfo)
	slog.SetDefault(logger)

	middleware := &loggingMiddleware{options: &LoggingOptions{}}
	ctx, _ := testhelpers.NewRouteContext(http.MethodGet, "/bad-request", nil)

	next := func(c routing.RouteContext) {
		c.Response().WriteHeader(http.StatusBadRequest)
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "level=WARN")
	assert.Contains(t, logOutput, "status=400")
	assert.Contains(t, logOutput, "GET /bad-request -> 400")
}

func TestLoggingMiddlewareShouldLogConflictAtWarningLevel(t *testing.T) {
	// Arrange
	logBuffer, logger := newTestLogger(slog.LevelWarn)
	slog.SetDefault(logger)

	middleware := &loggingMiddleware{options: &LoggingOptions{}}
	ctx, _ := testhelpers.NewRouteContext(http.MethodPut, "/conflict", nil)

	next := func(c routing.RouteContext) {
		c.Response().WriteHeader(http.StatusConflict)
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "level=WARN")
	assert.Contains(t, logOutput, "status=409")
	assert.Contains(t, logOutput, "PUT /conflict -> 409")
}

func TestLoggingMiddlewareShouldHandleErrorStatus(t *testing.T) {
	// Arrange
	logBuffer, logger := newTestLogger(slog.LevelInfo)
	slog.SetDefault(logger)

	middleware := &loggingMiddleware{options: &LoggingOptions{}}
	ctx, _ := testhelpers.NewRouteContext(http.MethodGet, "/error", nil)

	next := func(c routing.RouteContext) {
		c.Response().WriteHeader(http.StatusInternalServerError)
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "level=ERROR")
	assert.Contains(t, logOutput, "status=500")
}

func TestLoggingMiddlewareShouldDefaultTo200Status(t *testing.T) {
	// Arrange
	logBuffer, logger := newTestLogger(slog.LevelDebug)
	slog.SetDefault(logger)

	middleware := &loggingMiddleware{options: &LoggingOptions{}}
	ctx, _ := testhelpers.NewRouteContext(http.MethodGet, "/test", nil)

	next := func(c routing.RouteContext) {
		// Don't explicitly set status - should default to 200
		_, _ = c.Response().Write([]byte("response"))
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	logOutput := logBuffer.String()
	// When WriteHeader is not called explicitly, our recorder should default to 200 on first Write
	assert.Contains(t, logOutput, "level=DEBUG")
	assert.Contains(t, logOutput, "status=200")
}

func TestStatusRecorderShouldCaptureStatusCode(t *testing.T) {
	// Arrange
	_, recorder := testhelpers.NewRequestRecorder(http.MethodGet, "/test", nil)
	statusRec := &statusRecorder{ResponseWriter: recorder}

	// Act
	statusRec.WriteHeader(http.StatusAccepted)

	// Assert
	assert.Equal(t, http.StatusAccepted, statusRec.Status)
	assert.Equal(t, http.StatusAccepted, recorder.Code)
}

func TestStatusRecorderShouldForwardHeaders(t *testing.T) {
	// Arrange
	_, recorder := testhelpers.NewRequestRecorder(http.MethodGet, "/test", nil)
	statusRec := &statusRecorder{ResponseWriter: recorder}

	// Act
	statusRec.Header().Set("Test-Header", "test-value")

	// Assert
	assert.Equal(t, "test-value", recorder.Header().Get("Test-Header"))
}

func TestStatusRecorderShouldImplementResponseWriter(t *testing.T) {
	// Arrange
	_, recorder := testhelpers.NewRequestRecorder(http.MethodGet, "/test", nil)
	statusRec := &statusRecorder{ResponseWriter: recorder}

	// Assert
	assert.Implements(t, (*http.ResponseWriter)(nil), statusRec)
}

func TestShouldAddLoggingMiddlewareToRouter(t *testing.T) {
	// Arrange
	rtr := router.NewRouter()

	UseLogging(rtr)

	logBuffer, logger := newTestLogger(slog.LevelDebug)
	slog.SetDefault(logger)

	rtr.GET("/test", func(c routing.RouteContext) { _, _ = c.Response().Write([]byte("ok")) })
	req, rec := testhelpers.NewRequestRecorder(http.MethodGet, "/test?param=value", nil)
	req.RemoteAddr = "192.168.1.1:8080"
	req.Header.Set(common.HeaderUserAgent, "test-agent/1.0")

	// Act
	rtr.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, logBuffer.String(), "level=DEBUG")
	assert.Contains(t, logBuffer.String(), "GET /test -> 200")
}

func TestLoggingMiddlewareShouldUseRoutePatternInMessageAndAttrs(t *testing.T) {
	// Arrange
	logBuffer, logger := newTestLogger(slog.LevelDebug)
	slog.SetDefault(logger)

	rtr := router.NewRouter()
	UseLogging(rtr)
	rtr.GET("/users/{id}", func(c routing.RouteContext) {
		c.Response().WriteHeader(http.StatusAccepted)
	})

	req, rec := testhelpers.NewRequestRecorder(http.MethodGet, "/users/42", nil)

	// Act
	rtr.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusAccepted, rec.Code)
	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "level=DEBUG")
	assert.Contains(t, logOutput, "GET /users/{id} -> 202")
	assert.Contains(t, logOutput, "route=/users/{id}")
	assert.Contains(t, logOutput, "path=/users/42")
}

func TestLoggingMiddlewareShouldIncludeTraceAndSpanIDsWhenPresent(t *testing.T) {
	// Arrange
	logBuffer, logger := newTestLogger(slog.LevelDebug)
	slog.SetDefault(logger)

	middleware := &loggingMiddleware{options: &LoggingOptions{}}
	ctx, _ := testhelpers.NewRouteContext(http.MethodGet, "/trace", nil)
	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    trace.TraceID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
		SpanID:     trace.SpanID{0, 1, 2, 3, 4, 5, 6, 7},
		TraceFlags: trace.FlagsSampled,
	})
	ctx.SetRequest(ctx.Request().WithContext(trace.ContextWithSpanContext(ctx.Request().Context(), spanContext)))

	next := func(c routing.RouteContext) {
		c.Response().WriteHeader(http.StatusOK)
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "trace_id=000102030405060708090a0b0c0d0e0f")
	assert.Contains(t, logOutput, "span_id=0001020304050607")
}
