package cors

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
	testOriginExample = "https://example.com"
	testAPITestURL    = "https://api/test"
)

func TestShouldReturnNoContentWithoutCallingNextGivenPreflightRequest(t *testing.T) {
	// Arrange
	m := newCORSMiddleware(CORSOptions{AllowedOrigins: []string{testOriginExample}})

	req := httptest.NewRequest(http.MethodOptions, testAPITestURL, nil)
	req.Header.Set(common.HeaderOrigin, testOriginExample)
	req.Header.Set(common.HeaderAccessControlRequestMethod, "POST")
	req.Header.Set(common.HeaderAccessControlRequestHeaders, "X-Custom, Content-Type")

	rec := httptest.NewRecorder()
	ctx := routing.NewRouteContext(rec, req)

	called := false
	next := func(c routing.RouteContext) { called = true }

	// Act
	m.Invoke(ctx, next)

	// Assert
	assert.False(t, called, "preflight should not call next handler")
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, testOriginExample, rec.Header().Get(common.HeaderAccessControlAllowOrigin))
	assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", rec.Header().Get(common.HeaderAccessControlAllowMethods))
	assert.Equal(t, "X-Custom, Content-Type", rec.Header().Get(common.HeaderAccessControlAllowHeaders))
}

func TestShouldSetWildcardAllowOriginGivenSimpleRequestWithWildcardConfig(t *testing.T) {
	// Arrange
	m := newCORSMiddleware(CORSOptions{AllowedOrigins: []string{"*"}})

	req := httptest.NewRequest(http.MethodGet, testAPITestURL, nil)
	req.Header.Set(common.HeaderOrigin, "https://evil.com")
	rec := httptest.NewRecorder()
	ctx := routing.NewRouteContext(rec, req)

	next := func(c routing.RouteContext) {
		c.Response().WriteHeader(http.StatusOK)
	}

	// Act
	m.Invoke(ctx, next)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
	// when wildcard allowed and credentials not allowed, header should be '*'
	assert.Equal(t, "*", rec.Header().Get(common.HeaderAccessControlAllowOrigin))
}

func TestShouldReflectOriginAndSetCredentialsGivenAllowCredentialsEnabled(t *testing.T) {
	// Arrange
	m := newCORSMiddleware(CORSOptions{AllowedOrigins: []string{testOriginExample}, AllowCredentials: true})

	req := httptest.NewRequest(http.MethodGet, testAPITestURL, nil)
	req.Header.Set(common.HeaderOrigin, testOriginExample)
	rec := httptest.NewRecorder()
	ctx := routing.NewRouteContext(rec, req)

	next := func(c routing.RouteContext) { c.Response().WriteHeader(http.StatusOK) }

	// Act
	m.Invoke(ctx, next)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
	// when credentials allowed, origin must be echoed back (not '*')
	assert.Equal(t, testOriginExample, rec.Header().Get(common.HeaderAccessControlAllowOrigin))
	assert.Equal(t, "true", rec.Header().Get(common.HeaderAccessControlAllowCredentials))
}

func TestShouldApplyCORSHeadersGivenUseCORSMiddleware(t *testing.T) {
	// Arrange
	rtr := router.NewRouter()
	UseCORS(rtr, WithAllowedOrigins("*"))

	rtr.GET("/test", func(c routing.RouteContext) {
		c.Response().WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, testAPITestURL, nil)
	req.Header.Set(common.HeaderOrigin, testOriginExample)
	rec := httptest.NewRecorder()

	// Act
	rtr.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "*", rec.Header().Get(common.HeaderAccessControlAllowOrigin))
}
