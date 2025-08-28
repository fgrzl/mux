package enforcehttps

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fgrzl/mux/internal/router"
	"github.com/fgrzl/mux/internal/routing"
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
	assert.Equal(t, "https://example.com/test", recorder.Header().Get("Location"))
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
	// Arrange
	rtr := router.NewRouter()
	initialMiddlewareCount := len(rtr.middleware)

	// Act
	UseEnforceHTTPS(router)

	// Assert
	assert.Equal(t, initialMiddlewareCount+1, len(rtr.middleware))
	assert.IsType(t, &enforceHTTPSMiddleware{}, rtr.middleware[len(rtr.middleware)-1])
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
	assert.Equal(t, "https://example.com/test?param=value&other=123", recorder.Header().Get("Location"))
}
