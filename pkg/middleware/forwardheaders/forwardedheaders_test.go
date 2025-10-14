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

const (
	testURL             = "http://example.com/test"
	fhProtoHttps        = "https"
	fhAddr1             = "192.168.1.100"
	fhAddr2             = "10.0.0.1"
	fhAddr3             = "1.2.3.4"
	assertNextCalledMsg = "next handler should be called"
)

func TestShouldProcessXForwardedProtoHeader(t *testing.T) {
	// Arrange
	middleware := &forwardedHeadersMiddleware{}

	req := httptest.NewRequest(http.MethodGet, testURL, nil)
	req.Header.Set(common.HeaderXForwardedProto, fhProtoHttps)
	recorder := httptest.NewRecorder()
	ctx := routing.NewRouteContext(recorder, req)

	nextCalled := false
	next := func(c routing.RouteContext) {
		nextCalled = true
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.True(t, nextCalled, assertNextCalledMsg)
	assert.Equal(t, "https", ctx.Request().URL.Scheme)
}

func TestShouldProcessXForwardedForHeader(t *testing.T) {
	// Arrange
	middleware := &forwardedHeadersMiddleware{}
	req := httptest.NewRequest(http.MethodGet, testURL, nil)
	req.Header.Set(common.HeaderXForwardedFor, fhAddr1)
	recorder := httptest.NewRecorder()
	ctx := routing.NewRouteContext(recorder, req)

	nextCalled := false
	next := func(c routing.RouteContext) {
		nextCalled = true
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.True(t, nextCalled, assertNextCalledMsg)
	assert.Equal(t, "192.168.1.100", ctx.Request().RemoteAddr)
}

func TestShouldProcessBothForwardedHeaders(t *testing.T) {
	// Arrange
	middleware := &forwardedHeadersMiddleware{}
	req := httptest.NewRequest(http.MethodGet, testURL, nil)
	req.Header.Set(common.HeaderXForwardedProto, fhProtoHttps)
	req.Header.Set(common.HeaderXForwardedFor, fhAddr2)
	recorder := httptest.NewRecorder()
	ctx := routing.NewRouteContext(recorder, req)

	nextCalled := false
	next := func(c routing.RouteContext) {
		nextCalled = true
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.True(t, nextCalled, assertNextCalledMsg)
	assert.Equal(t, "https", ctx.Request().URL.Scheme)
	assert.Equal(t, "10.0.0.1", ctx.Request().RemoteAddr)
}

func TestShouldIgnoreEmptyForwardedHeaders(t *testing.T) {
	// Arrange
	middleware := &forwardedHeadersMiddleware{}
	req := httptest.NewRequest(http.MethodGet, testURL, nil)
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
	assert.True(t, nextCalled, assertNextCalledMsg)
	assert.Equal(t, originalScheme, ctx.Request().URL.Scheme)
	assert.Equal(t, originalRemoteAddr, ctx.Request().RemoteAddr)
}

func TestShouldAddForwardedHeadersMiddlewareToRouter(t *testing.T) {
	rtr := router.NewRouter()

	// Act - register middleware
	UseForwardedHeaders(rtr)

	// Register a route whose handler will echo the request scheme and remote addr
	rtr.GET("/test", func(c routing.RouteContext) {
		_, _ = c.Response().Write([]byte(c.Request().URL.Scheme + "|" + c.Request().RemoteAddr))
	})

	// Make a request with forwarded headers so middleware will modify the request
	req := httptest.NewRequest(http.MethodGet, testURL, nil)
	req.Header.Set(common.HeaderXForwardedProto, fhProtoHttps)
	req.Header.Set(common.HeaderXForwardedFor, fhAddr3)
	rec := httptest.NewRecorder()
	rtr.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "https|1.2.3.4", rec.Body.String())
}
