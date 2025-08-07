package mux

import (
	"net/http/httptest"
	"testing"

	"github.com/oschwald/geoip2-golang"
	"github.com/stretchr/testify/assert"
)

const (
	loopbackAddr       = "127.0.0.1:12345"
	loopbackIP         = "127.0.0.1"
	forwardedForIP     = "203.0.113.1"
	forwardedForList   = "203.0.113.1, 192.168.1.1"
	forwardedForListW  = " 203.0.113.1 , 192.168.1.1 "
	realIP             = "203.0.113.2"
	realIPHeader       = "X-Real-IP"
	forwardedForHeader = "X-Forwarded-For"
	invalidIP          = "invalid-ip"
)

var (
	expectedCountries      = []string{"IR", "KP", "SY", "CU", "RU"}
	nonRestrictedCountries = []string{"US", "CA", "GB", "FR", "DE"}
)

func TestShouldCreateExportControlOptionsWithDatabase(t *testing.T) {
	// Arrange
	options := &ExportControlOptions{}
	// Note: We can't create a real geoip2.Reader in unit tests without the actual database file
	// So we'll test with nil and focus on the option pattern

	// Act
	opt := WithGeoIPDatabase(nil)
	opt(options)

	// Assert
	assert.Nil(t, options.DB) // Since we passed nil
}

func TestShouldAddExportControlMiddlewareToRouter(t *testing.T) {
	// Arrange
	router := NewRouter()
	initialCount := len(router.middleware)

	// Act
	router.UseExportControl()

	// Assert
	assert.Len(t, router.middleware, initialCount+1)
}

func TestShouldAllowAccessWhenNoDatabaseConfigured(t *testing.T) {
	// Arrange
	middleware := &exportControlMiddleware{
		options: &ExportControlOptions{DB: nil},
	}
	// Safe: loopback address for testing
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = loopbackAddr
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)

	nextCalled := false
	next := func(c *RouteContext) {
		nextCalled = true
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.True(t, nextCalled)
	assert.Equal(t, 200, rec.Code) // Should not be blocked
}

func TestShouldAllowAccessWhenIPCannotBeParsed(t *testing.T) {
	// Arrange
	middleware := &exportControlMiddleware{
		options: &ExportControlOptions{DB: nil}, // Even with DB, invalid IP should pass through
	}
	// Safe: loopback address for testing
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = invalidIP
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)

	nextCalled := false
	next := func(c *RouteContext) {
		nextCalled = true
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.True(t, nextCalled)
}

func TestGetRealIPShouldReturnXForwardedFor(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(forwardedForHeader, forwardedForList)
	// Safe: loopback address for testing
	req.RemoteAddr = loopbackAddr

	// Act
	ip := getRealIP(req)

	// Assert
	assert.Equal(t, forwardedForIP, ip) // Should return first IP from X-Forwarded-For
}

func TestGetRealIPShouldReturnXRealIP(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(realIPHeader, realIP)
	// Safe: loopback address for testing
	req.RemoteAddr = loopbackAddr

	// Act
	ip := getRealIP(req)

	// Assert
	assert.Equal(t, realIP, ip) // Should return X-Real-IP
}

func TestGetRealIPShouldPreferXForwardedForOverXRealIP(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(forwardedForHeader, forwardedForIP)
	req.Header.Set(realIPHeader, realIP)
	// Safe: loopback address for testing
	req.RemoteAddr = loopbackAddr

	// Act
	ip := getRealIP(req)

	// Assert
	assert.Equal(t, forwardedForIP, ip) // Should prefer X-Forwarded-For
}

func TestGetRealIPShouldReturnRemoteAddrWhenNoHeaders(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	// Safe: loopback address for testing
	req.RemoteAddr = loopbackAddr

	// Act
	ip := getRealIP(req)

	// Assert
	assert.Equal(t, loopbackIP, ip) // Should extract IP from RemoteAddr
}

func TestGetRealIPShouldReturnRemoteAddrWhenCannotSplit(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	// Safe: loopback address for testing
	req.RemoteAddr = loopbackIP // No port

	// Act
	ip := getRealIP(req)

	// Assert
	assert.Equal(t, loopbackIP, ip) // Should return as-is when can't split
}

func TestGetRealIPShouldHandleXForwardedForWithSpaces(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(forwardedForHeader, forwardedForListW)
	// Safe: loopback address for testing
	req.RemoteAddr = loopbackAddr

	// Act
	ip := getRealIP(req)

	// Assert
	assert.Equal(t, "203.0.113.1", ip) // Should trim spaces
}

func TestExportRestrictedCountriesShouldContainExpectedCountries(t *testing.T) {
	// Arrange & Act & Assert
	for _, country := range expectedCountries {
		_, exists := exportRestrictedCountries[country]
		assert.True(t, exists, "Country %s should be in restricted list", country)
	}
	// Test that some non-restricted countries are not in the list
	for _, country := range nonRestrictedCountries {
		_, exists := exportRestrictedCountries[country]
		assert.False(t, exists, "Country %s should not be in restricted list", country)
	}
}

func TestShouldCreateExportControlMiddleware(t *testing.T) {
	// Arrange
	options := &ExportControlOptions{}

	// Act
	middleware := &exportControlMiddleware{options: options}

	// Assert
	assert.NotNil(t, middleware)
	assert.Equal(t, options, middleware.options)
}

func TestShouldHandleMultipleExportControlOptions(t *testing.T) {
	// Arrange
	router := NewRouter()
	var db *geoip2.Reader // Would be a real DB in production

	// Act
	router.UseExportControl(
		WithGeoIPDatabase(db),
		// Could add more options here
	)

	// Assert - Should not panic and middleware should be added
	assert.Len(t, router.middleware, 1)
}

func TestShouldHandleEmptyXForwardedFor(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(forwardedForHeader, "")
	req.Header.Set(realIPHeader, realIP)
	// Safe: loopback address for testing
	req.RemoteAddr = loopbackAddr

	// Act
	ip := getRealIP(req)

	// Assert
	assert.Equal(t, realIP, ip) // Should fall back to X-Real-IP
}

func TestShouldHandleEmptyXRealIP(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(realIPHeader, "")
	// Safe: loopback address for testing
	req.RemoteAddr = loopbackAddr

	// Act
	ip := getRealIP(req)

	// Assert
	assert.Equal(t, loopbackIP, ip) // Should fall back to RemoteAddr
}

func TestShouldProcessRequestWhenGeoLookupFails(t *testing.T) {
	// Arrange
	// We can't easily mock geoip2.Reader without significant complexity
	// But we can test the logic path when DB is nil or IP parsing fails
	middleware := &exportControlMiddleware{
		options: &ExportControlOptions{DB: nil},
	}
	// Safe: loopback address for testing
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = loopbackAddr // Valid IP but no DB
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)

	nextCalled := false
	next := func(c *RouteContext) {
		nextCalled = true
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.True(t, nextCalled) // Should continue since no DB is configured
}
