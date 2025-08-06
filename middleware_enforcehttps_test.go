package mux

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldRedirectHTTPToHTTPS(t *testing.T) {
	// Arrange
	middleware := &enforceHTTPSMiddleware{}
	req := httptest.NewRequest("GET", "http://example.com/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)
	
	nextCalled := false
	next := func(c *RouteContext) {
		nextCalled = true
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.False(t, nextCalled, "next handler should not be called for HTTP requests")
	assert.Equal(t, http.StatusMovedPermanently, recorder.Code)
	assert.Equal(t, "https://example.com/test", recorder.Header().Get("Location"))
}

func TestShouldAllowHTTPSRequests(t *testing.T) {
	// Arrange
	middleware := &enforceHTTPSMiddleware{}
	req := httptest.NewRequest("GET", "https://example.com/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)
	
	nextCalled := false
	next := func(c *RouteContext) {
		nextCalled = true
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.True(t, nextCalled, "next handler should be called for HTTPS requests")
	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestShouldAddEnforceHTTPSMiddlewareToRouter(t *testing.T) {
	// Arrange
	router := NewRouter()
	initialMiddlewareCount := len(router.middleware)

	// Act
	router.UseEnforceHTTPS()

	// Assert
	assert.Equal(t, initialMiddlewareCount+1, len(router.middleware))
	assert.IsType(t, &enforceHTTPSMiddleware{}, router.middleware[len(router.middleware)-1])
}

func TestShouldPreserveQueryParametersInRedirect(t *testing.T) {
	// Arrange
	middleware := &enforceHTTPSMiddleware{}
	req := httptest.NewRequest("GET", "http://example.com/test?param=value&other=123", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)
	
	next := func(c *RouteContext) {}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.Equal(t, http.StatusMovedPermanently, recorder.Code)
	assert.Equal(t, "https://example.com/test?param=value&other=123", recorder.Header().Get("Location"))
}