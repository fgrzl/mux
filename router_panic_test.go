package mux

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServeHTTP_PanicRecovery_Production(t *testing.T) {
	// Arrange
	router := NewRouter()
	router.GET("/panic", func(c *RouteContext) {
		panic("test panic")
	})

	// Ensure we're in production mode (no dev environment set)
	os.Unsetenv("GO_ENV")
	os.Unsetenv("ENVIRONMENT")

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	recorder := httptest.NewRecorder()

	// Act
	router.ServeHTTP(recorder, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Equal(t, "application/problem+json", recorder.Header().Get("Content-Type"))

	var problemDetails ProblemDetails
	err := json.Unmarshal(recorder.Body.Bytes(), &problemDetails)
	require.NoError(t, err)

	assert.Equal(t, http.StatusInternalServerError, problemDetails.Status)
	assert.Equal(t, "Internal Server Error", problemDetails.Title)
	assert.Equal(t, "An unexpected error occurred", problemDetails.Detail)
	assert.NotContains(t, problemDetails.Detail, "test panic") // Should not leak panic details
}

func TestServeHTTP_PanicRecovery_Development(t *testing.T) {
	// Arrange
	router := NewRouter()
	router.GET("/panic", func(c *RouteContext) {
		panic("test panic")
	})

	// Set development environment
	os.Setenv("GO_ENV", "development")
	defer os.Unsetenv("GO_ENV")

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	recorder := httptest.NewRecorder()

	// Act
	router.ServeHTTP(recorder, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Equal(t, "application/problem+json", recorder.Header().Get("Content-Type"))

	var problemDetails ProblemDetails
	err := json.Unmarshal(recorder.Body.Bytes(), &problemDetails)
	require.NoError(t, err)

	assert.Equal(t, http.StatusInternalServerError, problemDetails.Status)
	assert.Equal(t, "Panic Recovered", problemDetails.Title)
	assert.Contains(t, problemDetails.Detail, "test panic") // Should include panic details
}

func TestServeHTTP_PanicRecovery_DevEnvironmentVariable(t *testing.T) {
	// Arrange
	router := NewRouter()
	router.GET("/panic", func(c *RouteContext) {
		panic("test panic")
	})

	// Set dev environment using "dev" value
	os.Setenv("GO_ENV", "dev")
	defer os.Unsetenv("GO_ENV")

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	recorder := httptest.NewRecorder()

	// Act
	router.ServeHTTP(recorder, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)

	var problemDetails ProblemDetails
	err := json.Unmarshal(recorder.Body.Bytes(), &problemDetails)
	require.NoError(t, err)

	assert.Equal(t, "Panic Recovered", problemDetails.Title)
	assert.Contains(t, problemDetails.Detail, "test panic") // Should include panic details
}

func TestServeHTTP_PanicRecovery_EnvironmentVariable(t *testing.T) {
	// Arrange
	router := NewRouter()
	router.GET("/panic", func(c *RouteContext) {
		panic("test panic")
	})

	// Test using ENVIRONMENT variable instead of GO_ENV
	os.Unsetenv("GO_ENV")
	os.Setenv("ENVIRONMENT", "development")
	defer os.Unsetenv("ENVIRONMENT")

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	recorder := httptest.NewRecorder()

	// Act
	router.ServeHTTP(recorder, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)

	var problemDetails ProblemDetails
	err := json.Unmarshal(recorder.Body.Bytes(), &problemDetails)
	require.NoError(t, err)

	assert.Equal(t, "Panic Recovered", problemDetails.Title)
	assert.Contains(t, problemDetails.Detail, "test panic") // Should include panic details
}

func TestServeHTTP_NormalOperation_NotAffected(t *testing.T) {
	// Arrange
	router := NewRouter()
	router.GET("/normal", func(c *RouteContext) {
		c.OK(map[string]string{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/normal", nil)
	recorder := httptest.NewRecorder()

	// Act
	router.ServeHTTP(recorder, req)

	// Assert
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
	assert.Contains(t, recorder.Body.String(), "success")
}

func TestServeHTTP_PanicInMiddleware(t *testing.T) {
	// Arrange
	router := NewRouter()

	// Add middleware that panics
	panicMiddleware := &testPanicMiddleware{}
	router.middleware = append(router.middleware, panicMiddleware)

	router.GET("/test", func(c *RouteContext) {
		c.OK(map[string]string{"message": "should not reach here"})
	})

	// Ensure production mode
	os.Unsetenv("GO_ENV")
	os.Unsetenv("ENVIRONMENT")

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()

	// Act
	router.ServeHTTP(recorder, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)

	var problemDetails ProblemDetails
	err := json.Unmarshal(recorder.Body.Bytes(), &problemDetails)
	require.NoError(t, err)

	assert.Equal(t, "Internal Server Error", problemDetails.Title)
	assert.Equal(t, "An unexpected error occurred", problemDetails.Detail)
}

// Helper middleware that panics for testing
type testPanicMiddleware struct{}

func (m *testPanicMiddleware) Invoke(ctx *RouteContext, next HandlerFunc) {
	panic("middleware panic")
}
