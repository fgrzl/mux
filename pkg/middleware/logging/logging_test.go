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
)

func TestLoggingMiddlewareShouldLogRequestDetails(t *testing.T) {
	// Arrange
	var logBuffer bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuffer, nil))
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
	assert.Contains(t, logOutput, "http_request")
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
	var logBuffer bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuffer, nil))
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
	assert.Contains(t, logOutput, "status=201")
}

func TestLoggingMiddlewareShouldHandleErrorStatus(t *testing.T) {
	// Arrange
	var logBuffer bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuffer, nil))
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
	assert.Contains(t, logOutput, "status=500")
}

func TestLoggingMiddlewareShouldDefaultTo200Status(t *testing.T) {
	// Arrange
	var logBuffer bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuffer, nil))
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
	rtr := router.NewRouter()

	// Act - register middleware
	UseLogging(rtr)

	// Register a route and make a request to ensure middleware runs and logs
	var logBuffer bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuffer, nil))
	slog.SetDefault(logger)

	rtr.GET("/test", func(c routing.RouteContext) { _, _ = c.Response().Write([]byte("ok")) })
	req, rec := testhelpers.NewRequestRecorder(http.MethodGet, "/test?param=value", nil)
	req.RemoteAddr = "192.168.1.1:8080"
	req.Header.Set(common.HeaderUserAgent, "test-agent/1.0")
	rtr.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	// Ensure something was logged
	assert.Contains(t, logBuffer.String(), "http_request")
}
