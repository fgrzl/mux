package enforcehttps

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fgrzl/mux/pkg/common"
	"github.com/fgrzl/mux/pkg/router"
	"github.com/fgrzl/mux/pkg/routing"
	"github.com/stretchr/testify/assert"
)

func TestShouldRedirectHTTPToHTTPS(t *testing.T) {
	// Arrange
	middleware := &enforceHTTPSMiddleware{}
	req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
	recorder := httptest.NewRecorder()
	ctx := routing.NewRouteContext(recorder, req)

	nextCalled := false
	next := func(c routing.RouteContext) {
		nextCalled = true
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.False(t, nextCalled, "next handler should not be called for HTTP requests")
	assert.Equal(t, http.StatusMovedPermanently, recorder.Code)
	assert.Equal(t, "https://example.com/test", recorder.Header().Get(common.HeaderLocation))
}

func TestShouldAllowHTTPSRequests(t *testing.T) {
	// Arrange
	middleware := &enforceHTTPSMiddleware{}
	req := httptest.NewRequest(http.MethodGet, "https://example.com/test", nil)
	recorder := httptest.NewRecorder()
	ctx := routing.NewRouteContext(recorder, req)

	nextCalled := false
	next := func(c routing.RouteContext) {
		nextCalled = true
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.True(t, nextCalled, "next handler should be called for HTTPS requests")
	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestShouldAddEnforceHTTPSMiddlewareToRouter(t *testing.T) {
	rtr := router.NewRouter()

	// Act - register the middleware
	UseEnforceHTTPS(rtr)

	// Register a route and make an HTTP (non-HTTPS) request to ensure middleware redirects
	rtr.GET("/test", func(c routing.RouteContext) {
		c.Response().WriteHeader(http.StatusOK)
		_, _ = c.Response().Write([]byte("ok"))
	})

	req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
	rec := httptest.NewRecorder()
	rtr.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusMovedPermanently, rec.Code)
	assert.Equal(t, "https://example.com/test", rec.Header().Get(common.HeaderLocation))
}

func TestShouldPreserveQueryParametersInRedirect(t *testing.T) {
	// Arrange
	middleware := &enforceHTTPSMiddleware{}
	req := httptest.NewRequest(http.MethodGet, "http://example.com/test?param=value&other=123", nil)
	recorder := httptest.NewRecorder()
	ctx := routing.NewRouteContext(recorder, req)

	next := func(c routing.RouteContext) {}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.Equal(t, http.StatusMovedPermanently, recorder.Code)
	assert.Equal(t, "https://example.com/test?param=value&other=123", recorder.Header().Get(common.HeaderLocation))
}
