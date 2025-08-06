package mux

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fgrzl/claims"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthenticateShouldPanicWhenNoTokenProviderIsAvailable(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := &RouteContext{
		Request:  req,
		Response: res,
		services: make(map[ServiceKey]any), // No token provider service
	}

	mockUser := newMockPrincipal("test-user")

	// Act & Assert
	defer func() {
		r := recover()
		require.NotNil(t, r, "Expected panic but none occurred")

		panicMsg := r.(string)
		assert.Contains(t, panicMsg, "DEVELOPMENT ERROR")
		assert.Contains(t, panicMsg, "UseAuthentication()")
		assert.Contains(t, panicMsg, "c.Authenticate()")
	}()

	ctx.Authenticate("test-cookie", mockUser)
}

func TestAuthenticateShouldCreateCookieWithTTLWhenProviderIsAvailable(t *testing.T) {
	// Arrange
	router := NewRouter()
	router.UseAuthentication(
		WithTokenTTL(30*time.Minute),
		WithValidator(func(string) (claims.Principal, error) {
			return newMockPrincipal("validator-user"), nil
		}),
		WithTokenCreator(func(user claims.Principal, ttl time.Duration) (string, error) {
			return "test-token-" + user.Subject(), nil
		}),
	)

	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()

	var capturedContext *RouteContext
	router.GET("/test", func(c *RouteContext) {
		capturedContext = c
		// Call Authenticate
		testUser := newMockPrincipal("auth-user")
		c.Authenticate("test-auth-cookie", testUser)
		c.OK("success")
	}).AllowAnonymous()

	// Act
	router.ServeHTTP(res, req)

	// Assert
	require.NotNil(t, capturedContext)
	assert.Equal(t, http.StatusOK, res.Code)

	// Verify cookie was set
	cookies := res.Header().Get("Set-Cookie")
	assert.Contains(t, cookies, "test-auth-cookie=test-token-auth-user")
	assert.Contains(t, cookies, "Max-Age=1800") // 30 minutes = 1800 seconds
	assert.Contains(t, cookies, "HttpOnly")
	assert.Contains(t, cookies, "Secure")
}
