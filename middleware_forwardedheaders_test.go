package mux

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldProcessXForwardedProtoHeader(t *testing.T) {
	// Arrange
	middleware := &forwardedHeadersMiddleware{}
	req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	nextCalled := false
	next := func(c RouteContext) {
		nextCalled = true
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.True(t, nextCalled, "next handler should be called")
	assert.Equal(t, "https", ctx.Request().URL.Scheme)
}

func TestShouldProcessXForwardedForHeader(t *testing.T) {
	// Arrange
	middleware := &forwardedHeadersMiddleware{}
	req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.100")
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	nextCalled := false
	next := func(c RouteContext) {
		nextCalled = true
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.True(t, nextCalled, "next handler should be called")
	assert.Equal(t, "192.168.1.100", ctx.Request().RemoteAddr)
}

func TestShouldProcessBothForwardedHeaders(t *testing.T) {
	// Arrange
	middleware := &forwardedHeadersMiddleware{}
	req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-For", "10.0.0.1")
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	nextCalled := false
	next := func(c RouteContext) {
		nextCalled = true
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.True(t, nextCalled, "next handler should be called")
	assert.Equal(t, "https", ctx.Request().URL.Scheme)
	assert.Equal(t, "10.0.0.1", ctx.Request().RemoteAddr)
}

func TestShouldIgnoreEmptyForwardedHeaders(t *testing.T) {
	// Arrange
	middleware := &forwardedHeadersMiddleware{}
	req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
	originalScheme := req.URL.Scheme
	originalRemoteAddr := req.RemoteAddr
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	nextCalled := false
	next := func(c RouteContext) {
		nextCalled = true
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.True(t, nextCalled, "next handler should be called")
	assert.Equal(t, originalScheme, ctx.Request().URL.Scheme)
	assert.Equal(t, originalRemoteAddr, ctx.Request().RemoteAddr)
}

func TestShouldAddForwardedHeadersMiddlewareToRouter(t *testing.T) {
	// Arrange
	router := NewRouter()
	initialMiddlewareCount := len(router.middleware)

	// Act
	UseForwardedHeaders(router)

	// Assert
	assert.Equal(t, initialMiddlewareCount+1, len(router.middleware))
	assert.IsType(t, &forwardedHeadersMiddleware{}, router.middleware[len(router.middleware)-1])
}
