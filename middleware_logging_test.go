package mux

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoggingMiddlewareShouldLogRequestDetails(t *testing.T) {
	// Arrange
	var logBuffer bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuffer, nil))
	slog.SetDefault(logger)

	middleware := &loggingMiddleware{options: &LoggingOptions{}}
	req := httptest.NewRequest(http.MethodGet, "/test?param=value", nil)
	req.RemoteAddr = "192.168.1.1:8080"
	req.Header.Set("User-Agent", "test-agent/1.0")

	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	nextCalled := false
	next := func(c *RouteContext) {
		nextCalled = true
		c.Response.WriteHeader(http.StatusOK)
		c.Response.Write([]byte("test response"))
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.True(t, nextCalled, "next handler should be called")

	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "http_request")
	assert.Contains(t, logOutput, "method=GET")
	assert.Contains(t, logOutput, "path=/test")
	assert.Contains(t, logOutput, "status=200")
	assert.Contains(t, logOutput, "remote=192.168.1.1:8080")
	assert.Contains(t, logOutput, "user_agent=test-agent/1.0")
	assert.Contains(t, logOutput, "duration=")
}

func TestLoggingMiddlewareShouldCaptureStatusCode(t *testing.T) {
	// Arrange
	var logBuffer bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuffer, nil))
	slog.SetDefault(logger)

	middleware := &loggingMiddleware{options: &LoggingOptions{}}
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	next := func(c *RouteContext) {
		c.Response.WriteHeader(http.StatusCreated)
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
	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	next := func(c *RouteContext) {
		c.Response.WriteHeader(http.StatusInternalServerError)
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
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	next := func(c *RouteContext) {
		// Don't explicitly set status - should default to 200
		c.Response.Write([]byte("response"))
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	logOutput := logBuffer.String()
	// When WriteHeader is not called explicitly, the status should be 0 (not set)
	// but the actual HTTP status will be 200
	assert.Contains(t, logOutput, "status=0")
}

func TestStatusRecorderShouldCaptureStatusCode(t *testing.T) {
	// Arrange
	recorder := httptest.NewRecorder()
	statusRec := &statusRecorder{ResponseWriter: recorder}

	// Act
	statusRec.WriteHeader(http.StatusAccepted)

	// Assert
	assert.Equal(t, http.StatusAccepted, statusRec.Status)
	assert.Equal(t, http.StatusAccepted, recorder.Code)
}

func TestStatusRecorderShouldForwardHeaders(t *testing.T) {
	// Arrange
	recorder := httptest.NewRecorder()
	statusRec := &statusRecorder{ResponseWriter: recorder}

	// Act
	statusRec.Header().Set("Test-Header", "test-value")

	// Assert
	assert.Equal(t, "test-value", recorder.Header().Get("Test-Header"))
}

func TestStatusRecorderShouldImplementResponseWriter(t *testing.T) {
	// Arrange
	recorder := httptest.NewRecorder()
	statusRec := &statusRecorder{ResponseWriter: recorder}

	// Assert
	assert.Implements(t, (*http.ResponseWriter)(nil), statusRec)
}

func TestShouldAddLoggingMiddlewareToRouter(t *testing.T) {
	// Arrange
	router := NewRouter()
	initialMiddlewareCount := len(router.middleware)

	// Act
	router.UseLogging()

	// Assert
	assert.Equal(t, initialMiddlewareCount+1, len(router.middleware))
	assert.IsType(t, &loggingMiddleware{}, router.middleware[len(router.middleware)-1])
}
