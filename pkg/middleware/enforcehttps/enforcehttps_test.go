package enforcehttps

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fgrzl/mux/pkg/common"
	"github.com/fgrzl/mux/pkg/router"
	"github.com/fgrzl/mux/pkg/routing"
	"github.com/stretchr/testify/assert"
)

const (
	testBaseURL  = "https://example.com/test"
	testHTTPURL  = "http://example.com/test"
	testHTTPQURL = "http://example.com/test?param=value&other=123"
	testHTTPSURL = "https://example.com/test"
)

// newCtx creates a routing context and recorder for tests to reduce duplication.
func newCtx(method, urlStr string) (routing.RouteContext, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, urlStr, nil)
	rec := httptest.NewRecorder()
	return routing.NewRouteContext(rec, req), rec
}

// newCtxWithTLS creates a routing context simulating a TLS connection.
func newCtxWithTLS(method, urlStr string) (routing.RouteContext, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, urlStr, nil)
	req.TLS = &tls.ConnectionState{} // Simulate TLS connection
	rec := httptest.NewRecorder()
	return routing.NewRouteContext(rec, req), rec
}

// newCtxWithHeader creates a routing context with a specific header.
func newCtxWithHeader(method, urlStr, headerKey, headerVal string) (routing.RouteContext, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, urlStr, nil)
	req.Header.Set(headerKey, headerVal)
	rec := httptest.NewRecorder()
	return routing.NewRouteContext(rec, req), rec
}

func TestShouldRedirectHTTPToHTTPS(t *testing.T) {
	// Arrange
	middleware := &enforceHTTPSMiddleware{}
	ctx, rec := newCtx(http.MethodGet, testHTTPURL)

	nextCalled := false
	next := func(c routing.RouteContext) { nextCalled = true }

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.False(t, nextCalled, "next handler should not be called for HTTP requests")
	assert.Equal(t, http.StatusMovedPermanently, rec.Code)
	assert.Equal(t, testBaseURL, rec.Header().Get(common.HeaderLocation))
}

func TestShouldAllowHTTPSRequestsViaTLS(t *testing.T) {
	// Arrange
	middleware := &enforceHTTPSMiddleware{}
	ctx, rec := newCtxWithTLS(http.MethodGet, testHTTPURL) // URL says http but TLS is set

	nextCalled := false
	next := func(c routing.RouteContext) { nextCalled = true }

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.True(t, nextCalled, "next handler should be called for TLS connections")
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestShouldAllowHTTPSViaXForwardedProto(t *testing.T) {
	// Arrange
	middleware := &enforceHTTPSMiddleware{}
	ctx, rec := newCtxWithHeader(http.MethodGet, testHTTPURL, "X-Forwarded-Proto", "https")

	nextCalled := false
	next := func(c routing.RouteContext) { nextCalled = true }

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.True(t, nextCalled, "next handler should be called when X-Forwarded-Proto is https")
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestShouldAllowHTTPSViaForwardedHeader(t *testing.T) {
	// Arrange
	middleware := &enforceHTTPSMiddleware{}
	ctx, rec := newCtxWithHeader(http.MethodGet, testHTTPURL, "Forwarded", "for=192.0.2.60;proto=https;by=203.0.113.43")

	nextCalled := false
	next := func(c routing.RouteContext) { nextCalled = true }

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.True(t, nextCalled, "next handler should be called when Forwarded header has proto=https")
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestShouldRedirectWhenXForwardedProtoIsHTTP(t *testing.T) {
	// Arrange
	middleware := &enforceHTTPSMiddleware{}
	ctx, rec := newCtxWithHeader(http.MethodGet, testHTTPURL, "X-Forwarded-Proto", "http")

	nextCalled := false
	next := func(c routing.RouteContext) { nextCalled = true }

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.False(t, nextCalled, "next handler should not be called for HTTP")
	assert.Equal(t, http.StatusMovedPermanently, rec.Code)
}

func TestShouldAddEnforceHTTPSMiddlewareToRouter(t *testing.T) {
	// Arrange
	rtr := router.NewRouter()

	UseEnforceHTTPS(rtr)
	rtr.GET("/test", func(c routing.RouteContext) {
		c.Response().WriteHeader(http.StatusOK)
		_, _ = c.Response().Write([]byte("ok"))
	})

	req := httptest.NewRequest(http.MethodGet, testHTTPURL, nil)
	rec := httptest.NewRecorder()

	// Act
	rtr.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusMovedPermanently, rec.Code)
	assert.Equal(t, testBaseURL, rec.Header().Get(common.HeaderLocation))
}

func TestShouldPreserveQueryParametersInRedirect(t *testing.T) {
	// Arrange
	middleware := &enforceHTTPSMiddleware{}
	ctx, rec := newCtx(http.MethodGet, testHTTPQURL)

	next := func(c routing.RouteContext) {
		// no-op: used to exercise middleware without invoking downstream behavior
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.Equal(t, http.StatusMovedPermanently, rec.Code)
	assert.Equal(t, testBaseURL+"?param=value&other=123", rec.Header().Get(common.HeaderLocation))
}
