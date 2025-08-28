package forwardheaders

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fgrzl/mux/pkg/common"
	"github.com/fgrzl/mux/pkg/router"
	"github.com/fgrzl/mux/pkg/routing"
	"github.com/stretchr/testify/assert"
)

func TestShouldProcessXForwardedProtoHeader(t *testing.T) {
	// Arrange
	middleware := &forwardedHeadersMiddleware{}
	req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
	req.Header.Set(common.HeaderXForwardedProto, "https")
	recorder := httptest.NewRecorder()
	ctx := routing.NewRouteContext(recorder, req)

	nextCalled := false
	next := func(c routing.RouteContext) {
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
	req.Header.Set(common.HeaderXForwardedFor, "192.168.1.100")
	recorder := httptest.NewRecorder()
	ctx := routing.NewRouteContext(recorder, req)

	nextCalled := false
	next := func(c routing.RouteContext) {
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
	req.Header.Set(common.HeaderXForwardedProto, "https")
	req.Header.Set(common.HeaderXForwardedFor, "10.0.0.1")
	recorder := httptest.NewRecorder()
	ctx := routing.NewRouteContext(recorder, req)

	nextCalled := false
	next := func(c routing.RouteContext) {
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
	ctx := routing.NewRouteContext(recorder, req)

	nextCalled := false
	next := func(c routing.RouteContext) {
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
	rtr := router.NewRouter()

	// Act - register middleware
	UseForwardedHeaders(rtr)

	// Register a route whose handler will echo the request scheme and remote addr
	rtr.GET("/test", func(c routing.RouteContext) {
		c.Response().Write([]byte(c.Request().URL.Scheme + "|" + c.Request().RemoteAddr))
	})

	// Make a request with forwarded headers so middleware will modify the request
	req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
	req.Header.Set(common.HeaderXForwardedProto, "https")
	req.Header.Set(common.HeaderXForwardedFor, "1.2.3.4")
	rec := httptest.NewRecorder()
	rtr.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "https|1.2.3.4", rec.Body.String())
}
