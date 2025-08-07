package mux

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fgrzl/claims"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthenticateShouldPanicWhenNoTokenProviderIsAvailable(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	res := httptest.NewRecorder()
	ctx := &RouteContext{
		Request:  req,
		Response: res,
		services: make(map[ServiceKey]any),
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
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewRouteContext(res, req)

	// Set up the service in context (simulating middleware)
	provider := &defaultTokenProvider{
		ttl: 30 * time.Minute,
	}
	provider.signFn = func(user claims.Principal, ttl time.Duration) (string, error) {
		return "test-token-" + user.Subject(), nil
	}
	ctx.SetService(ServiceKeyTokenProvider, provider)

	mockUser := newMockPrincipal("test-user")

	// Act
	ctx.Authenticate("test-cookie", mockUser)

	// Assert
	assert.Contains(t, res.Header().Get("Set-Cookie"), "test-cookie=test-token-test-user")
	assert.Contains(t, res.Header().Get("Set-Cookie"), "Max-Age=1800") // 30 minutes = 1800 seconds
}

func TestSetCookieShouldSetCookieWithAllAttributes(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewRouteContext(res, req)

	// Act
	ctx.SetCookie("test-cookie", "test-value", 3600, "/path", "example.com", true, true)

	// Assert
	setCookieHeader := res.Header().Get("Set-Cookie")
	assert.Contains(t, setCookieHeader, "test-cookie=test-value")
	assert.Contains(t, setCookieHeader, "Max-Age=3600")
	assert.Contains(t, setCookieHeader, "Path=/path")
	assert.Contains(t, setCookieHeader, "Domain=example.com")
	assert.Contains(t, setCookieHeader, "Secure")
	assert.Contains(t, setCookieHeader, "HttpOnly")
	assert.Contains(t, setCookieHeader, "SameSite=Lax")
}

func TestGetCookieShouldReturnCookieValueWhenExists(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: "test-cookie", Value: "test-value"})
	res := httptest.NewRecorder()
	ctx := NewRouteContext(res, req)

	// Act
	value, err := ctx.GetCookie("test-cookie")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "test-value", value)
}

func TestGetCookieShouldReturnErrorWhenCookieDoesNotExist(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewRouteContext(res, req)

	// Act
	value, err := ctx.GetCookie("nonexistent-cookie")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "", value)
	assert.Equal(t, http.ErrNoCookie, err)
}

func TestClearCookieShouldSetExpiredCookie(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewRouteContext(res, req)

	// Act
	ctx.ClearCookie("test-cookie")

	// Assert
	setCookieHeader := res.Header().Get("Set-Cookie")
	assert.Contains(t, setCookieHeader, "test-cookie=")
	assert.Contains(t, setCookieHeader, "Max-Age=0") // Go changes -1 to 0 in the output
	assert.Contains(t, setCookieHeader, "Expires=Thu, 01 Jan 1970 00:00:01 GMT")
	assert.Contains(t, setCookieHeader, "Path=/")
	assert.Contains(t, setCookieHeader, "HttpOnly")
	assert.Contains(t, setCookieHeader, "Secure")
	assert.Contains(t, setCookieHeader, "SameSite=Lax")
}

func TestSignOutShouldClearAllCookiesAndRedirect(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodPost, "/signout", nil)
	res := httptest.NewRecorder()
	ctx := NewRouteContext(res, req)

	// Act
	ctx.SignOut()

	// Assert
	setCookieHeaders := res.Header().Values("Set-Cookie")

	// Should clear all three cookies
	userCookieCleared := false
	twoFactorCookieCleared := false
	idpSessionCookieCleared := false

	for _, header := range setCookieHeaders {
		if strings.Contains(header, GetUserCookieName()) {
			userCookieCleared = true
		}
		if strings.Contains(header, GetTwoFactorCookieName()) {
			twoFactorCookieCleared = true
		}
		if strings.Contains(header, GetIdpSessionCookieName()) {
			idpSessionCookieCleared = true
		}
	}

	assert.True(t, userCookieCleared, "User cookie should be cleared")
	assert.True(t, twoFactorCookieCleared, "Two-factor cookie should be cleared")
	assert.True(t, idpSessionCookieCleared, "IDP session cookie should be cleared")

	// Should redirect to logout page
	assert.Equal(t, http.StatusTemporaryRedirect, res.Code)
	assert.Equal(t, "http://example.com/logout", res.Header().Get("Location"))
}
