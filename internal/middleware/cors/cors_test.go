package cors

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fgrzl/mux/internal/common"
	"github.com/fgrzl/mux/internal/router"
	"github.com/fgrzl/mux/internal/routing"
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

// ---- Wildcard Pattern Tests ----

func TestWildcardPattern_SingleSubdomain(t *testing.T) {
	// Arrange
	m := newCORSMiddleware(CORSOptions{
		AllowedOrigins: []string{"*.example.com"},
	})

	tests := []struct {
		name     string
		origin   string
		expected bool
	}{
		{"single subdomain", "https://api.example.com", true},
		{"different subdomain", "https://www.example.com", true},
		{"multi-level subdomain", "https://api.v2.example.com", true},
		{"no subdomain", "https://example.com", false},
		{"wrong domain", "https://example.org", false},
		{"partial match", "https://notexample.com", false},
		{"with port", "https://api.example.com:8080", true},
		{"uppercase", "https://API.example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.isOriginAllowed(tt.origin)
			assert.Equal(t, tt.expected, result, "origin: %s", tt.origin)
		})
	}
}

func TestWildcardPattern_MultiplePatterns(t *testing.T) {
	// Arrange
	m := newCORSMiddleware(CORSOptions{
		AllowedOrigins: []string{"*.example.com", "*.test.io"},
	})

	tests := []struct {
		name     string
		origin   string
		expected bool
	}{
		{"matches first pattern", "https://api.example.com", true},
		{"matches second pattern", "https://api.test.io", true},
		{"no match", "https://api.other.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.isOriginAllowed(tt.origin)
			assert.Equal(t, tt.expected, result, "origin: %s", tt.origin)
		})
	}
}

func TestWildcardPattern_MixedWithExactOrigins(t *testing.T) {
	// Arrange
	m := newCORSMiddleware(CORSOptions{
		AllowedOrigins: []string{
			"https://exact.com",
			"*.example.com",
			"https://another-exact.org",
		},
	})

	tests := []struct {
		name     string
		origin   string
		expected bool
	}{
		{"exact match first", "https://exact.com", true},
		{"exact match second", "https://another-exact.org", true},
		{"wildcard match", "https://api.example.com", true},
		{"no match", "https://nomatch.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.isOriginAllowed(tt.origin)
			assert.Equal(t, tt.expected, result, "origin: %s", tt.origin)
		})
	}
}

func TestWildcardPattern_ResponseHeaders(t *testing.T) {
	// Arrange
	m := newCORSMiddleware(CORSOptions{
		AllowedOrigins: []string{"*.example.com"},
	})

	req := httptest.NewRequest(http.MethodGet, testAPITestURL, nil)
	req.Header.Set(common.HeaderOrigin, "https://api.example.com")
	rec := httptest.NewRecorder()
	ctx := routing.NewRouteContext(rec, req)

	next := func(c routing.RouteContext) {
		c.Response().WriteHeader(http.StatusOK)
	}

	// Act
	m.Invoke(ctx, next)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
	// Should reflect the exact origin, not the wildcard pattern
	assert.Equal(t, "https://api.example.com", rec.Header().Get(common.HeaderAccessControlAllowOrigin))
}

func TestWildcardPattern_PreflightRequest(t *testing.T) {
	// Arrange
	m := newCORSMiddleware(CORSOptions{
		AllowedOrigins: []string{"*.example.com"},
	})

	req := httptest.NewRequest(http.MethodOptions, testAPITestURL, nil)
	req.Header.Set(common.HeaderOrigin, "https://sub.example.com")
	req.Header.Set(common.HeaderAccessControlRequestMethod, "POST")
	req.Header.Set(common.HeaderAccessControlRequestHeaders, "Content-Type")

	rec := httptest.NewRecorder()
	ctx := routing.NewRouteContext(rec, req)

	called := false
	next := func(c routing.RouteContext) { called = true }

	// Act
	m.Invoke(ctx, next)

	// Assert
	assert.False(t, called, "preflight should not call next handler")
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "https://sub.example.com", rec.Header().Get(common.HeaderAccessControlAllowOrigin))
	assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", rec.Header().Get(common.HeaderAccessControlAllowMethods))
}

func TestWildcardPattern_WithCredentials(t *testing.T) {
	// Arrange
	m := newCORSMiddleware(CORSOptions{
		AllowedOrigins:   []string{"*.example.com"},
		AllowCredentials: true,
	})

	req := httptest.NewRequest(http.MethodGet, testAPITestURL, nil)
	req.Header.Set(common.HeaderOrigin, "https://secure.example.com")
	rec := httptest.NewRecorder()
	ctx := routing.NewRouteContext(rec, req)

	next := func(c routing.RouteContext) {
		c.Response().WriteHeader(http.StatusOK)
	}

	// Act
	m.Invoke(ctx, next)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "https://secure.example.com", rec.Header().Get(common.HeaderAccessControlAllowOrigin))
	assert.Equal(t, "true", rec.Header().Get(common.HeaderAccessControlAllowCredentials))
}

func TestWildcardPattern_RejectsNonMatching(t *testing.T) {
	// Arrange
	m := newCORSMiddleware(CORSOptions{
		AllowedOrigins: []string{"*.example.com"},
	})

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
	// Should NOT set CORS headers for non-matching origin
	assert.Empty(t, rec.Header().Get(common.HeaderAccessControlAllowOrigin))
}

func TestWithOriginWildcard_FunctionalOption(t *testing.T) {
	// Arrange
	rtr := router.NewRouter()
	UseCORS(rtr,
		WithOriginWildcard("*.example.com"),
		WithCredentials(true),
	)

	rtr.GET("/test", func(c routing.RouteContext) {
		c.Response().WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, testAPITestURL, nil)
	req.Header.Set(common.HeaderOrigin, "https://api.example.com")
	rec := httptest.NewRecorder()

	// Act
	rtr.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "https://api.example.com", rec.Header().Get(common.HeaderAccessControlAllowOrigin))
	assert.Equal(t, "true", rec.Header().Get(common.HeaderAccessControlAllowCredentials))
}

func TestWithOriginWildcard_MultiplePatterns(t *testing.T) {
	// Arrange
	rtr := router.NewRouter()
	UseCORS(rtr,
		WithAllowedOrigins("https://exact.com"),          // Set exact origins first
		WithOriginWildcard("*.example.com", "*.test.io"), // Then add wildcard patterns
	)

	rtr.GET("/test", func(c routing.RouteContext) {
		c.Response().WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name      string
		origin    string
		hasHeader bool
	}{
		{"wildcard 1", "https://api.example.com", true},
		{"wildcard 2", "https://api.test.io", true},
		{"exact", "https://exact.com", true},
		{"no match", "https://evil.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, testAPITestURL, nil)
			req.Header.Set(common.HeaderOrigin, tt.origin)
			rec := httptest.NewRecorder()

			rtr.ServeHTTP(rec, req)

			if tt.hasHeader {
				assert.Equal(t, tt.origin, rec.Header().Get(common.HeaderAccessControlAllowOrigin))
			} else {
				assert.Empty(t, rec.Header().Get(common.HeaderAccessControlAllowOrigin))
			}
		})
	}
}

func TestWildcardPattern_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		origin      string
		shouldMatch bool
	}{
		{"http vs https", "*.example.com", "http://api.example.com", true},
		{"trailing dot", "*.example.com", "https://api.example.com.", false}, // browsers don't send trailing dots
		{"deeply nested", "*.example.com", "https://a.b.c.d.example.com", true},
		{"port handling", "*.example.com", "https://api.example.com:3000", true},
		{"localhost wildcard", "*.localhost", "http://app.localhost", true},
		{"ip address (no match)", "*.example.com", "http://192.168.1.1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newCORSMiddleware(CORSOptions{
				AllowedOrigins: []string{tt.pattern},
			})
			result := m.isOriginAllowed(tt.origin)
			assert.Equal(t, tt.shouldMatch, result, "pattern: %s, origin: %s", tt.pattern, tt.origin)
		})
	}
}
