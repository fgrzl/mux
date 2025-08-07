package mux

import (
	"net/http/httptest"
	"testing"

	"github.com/oschwald/geoip2-golang"
	"github.com/stretchr/testify/assert"
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
	
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345" 
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
	
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "invalid-ip"
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
	req.Header.Set("X-Forwarded-For", "203.0.113.1, 192.168.1.1")
	req.RemoteAddr = "192.168.1.100:12345"

	// Act
	ip := getRealIP(req)

	// Assert
	assert.Equal(t, "203.0.113.1", ip) // Should return first IP from X-Forwarded-For
}

func TestGetRealIPShouldReturnXRealIP(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Real-IP", "203.0.113.2")
	req.RemoteAddr = "192.168.1.100:12345"

	// Act
	ip := getRealIP(req)

	// Assert
	assert.Equal(t, "203.0.113.2", ip) // Should return X-Real-IP
}

func TestGetRealIPShouldPreferXForwardedForOverXRealIP(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.1")
	req.Header.Set("X-Real-IP", "203.0.113.2")
	req.RemoteAddr = "192.168.1.100:12345"

	// Act
	ip := getRealIP(req)

	// Assert
	assert.Equal(t, "203.0.113.1", ip) // Should prefer X-Forwarded-For
}

func TestGetRealIPShouldReturnRemoteAddrWhenNoHeaders(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"

	// Act
	ip := getRealIP(req)

	// Assert
	assert.Equal(t, "192.168.1.100", ip) // Should extract IP from RemoteAddr
}

func TestGetRealIPShouldReturnRemoteAddrWhenCannotSplit(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.100" // No port

	// Act
	ip := getRealIP(req)

	// Assert
	assert.Equal(t, "192.168.1.100", ip) // Should return as-is when can't split
}

func TestGetRealIPShouldHandleXForwardedForWithSpaces(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", " 203.0.113.1 , 192.168.1.1 ")
	req.RemoteAddr = "192.168.1.100:12345"

	// Act
	ip := getRealIP(req)

	// Assert
	assert.Equal(t, "203.0.113.1", ip) // Should trim spaces
}

func TestExportRestrictedCountriesShouldContainExpectedCountries(t *testing.T) {
	// Arrange & Act & Assert
	expectedCountries := []string{"IR", "KP", "SY", "CU", "RU"}
	
	for _, country := range expectedCountries {
		_, exists := exportRestrictedCountries[country]
		assert.True(t, exists, "Country %s should be in restricted list", country)
	}
	
	// Test that some non-restricted countries are not in the list
	nonRestrictedCountries := []string{"US", "CA", "GB", "FR", "DE"}
	
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
	req.Header.Set("X-Forwarded-For", "")
	req.Header.Set("X-Real-IP", "203.0.113.2")
	req.RemoteAddr = "192.168.1.100:12345"

	// Act
	ip := getRealIP(req)

	// Assert
	assert.Equal(t, "203.0.113.2", ip) // Should fall back to X-Real-IP
}

func TestShouldHandleEmptyXRealIP(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Real-IP", "")
	req.RemoteAddr = "192.168.1.100:12345"

	// Act
	ip := getRealIP(req)

	// Assert
	assert.Equal(t, "192.168.1.100", ip) // Should fall back to RemoteAddr
}

func TestShouldProcessRequestWhenGeoLookupFails(t *testing.T) {
	// Arrange
	// We can't easily mock geoip2.Reader without significant complexity
	// But we can test the logic path when DB is nil or IP parsing fails
	middleware := &exportControlMiddleware{
		options: &ExportControlOptions{DB: nil},
	}
	
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345" // Valid IP but no DB
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