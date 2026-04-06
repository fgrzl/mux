package routing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fgrzl/claims"
	"github.com/fgrzl/mux/internal/common"
	"github.com/fgrzl/mux/internal/cookiekit"
	"github.com/fgrzl/mux/internal/tokenizer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testUserSubject = "test-user"
	testCookieName  = "test-cookie"
	testCookieValue = "test-value"
)

// mockProvider implements the TokenProvider interface for tests.
type mockProvider struct{ ttl time.Duration }

func (m *mockProvider) CreateToken(ctx context.Context, principal claims.Principal) (string, error) {
	return "test-token-" + principal.Subject(), nil
}

func (m *mockProvider) ValidateToken(ctx context.Context, token string) (claims.Principal, error) {
	return newMockPrincipal(testUserSubject), nil
}

func (m *mockProvider) GetTTL() time.Duration { return m.ttl }

func (m *mockProvider) CanCreateTokens() bool { return true }

func TestShouldPanicGivenNoTokenProviderWhenAuthenticating(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	res := httptest.NewRecorder()
	ctx := &DefaultRouteContext{
		request:  req,
		response: res,
		services: make(map[ServiceKey]any),
	}

	mockUser := newMockPrincipal(testUserSubject)

	// Act
	ctx.Authenticate(testCookieName, mockUser)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Contains(t, res.Body.String(), "Authentication Misconfigured")
}

func TestShouldReturnServerErrorGivenNoTokenProviderWhenSigningIn(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	res := httptest.NewRecorder()
	ctx := &DefaultRouteContext{
		request:  req,
		response: res,
		services: make(map[ServiceKey]any),
	}

	mockUser := newMockPrincipal(testUserSubject)

	// Act
	ctx.SignIn(mockUser, "/")

	// Assert
	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Contains(t, res.Body.String(), "Authentication Misconfigured")
}

// testPrincipal is a minimal claims.Principal implementation used by routing tests.
type testPrincipal struct{ subject string }

func (p *testPrincipal) Subject() string                      { return p.subject }
func (p *testPrincipal) Issuer() string                       { return "" }
func (p *testPrincipal) Audience() []string                   { return nil }
func (p *testPrincipal) ExpirationTime() int64                { return 0 }
func (p *testPrincipal) NotBefore() int64                     { return 0 }
func (p *testPrincipal) IssuedAt() int64                      { return 0 }
func (p *testPrincipal) JWTI() string                         { return "" }
func (p *testPrincipal) Scopes() []string                     { return nil }
func (p *testPrincipal) Roles() []string                      { return nil }
func (p *testPrincipal) Email() string                        { return "" }
func (p *testPrincipal) Username() string                     { return "" }
func (p *testPrincipal) CustomClaim(name string) claims.Claim { return nil }
func (p *testPrincipal) CustomClaimValue(name string) string  { return "" }
func (p *testPrincipal) Claims() *claims.ClaimSet             { return nil }

func newMockPrincipal(subject string) *testPrincipal {
	return &testPrincipal{subject: subject}
}

func TestShouldCreateCookieWithTTLGivenTokenProviderWhenAuthenticating(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewRouteContext(res, req)

	// Set up the service in context (simulating middleware)
	provider := &mockProvider{ttl: 30 * time.Minute}
	ctx.SetService(tokenizer.ServiceKeyTokenProvider, provider)

	mockUser := newMockPrincipal(testUserSubject)

	// Act
	ctx.Authenticate(testCookieName, mockUser)

	// Assert
	assert.Contains(t, res.Header().Get(common.HeaderSetCookie), "test-cookie=test-token-test-user")
	assert.Contains(t, res.Header().Get(common.HeaderSetCookie), "Max-Age=1800") // 30 minutes = 1800 seconds
}

func TestShouldSetCookieWithAllAttributesGivenSetCookie(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewRouteContext(res, req)

	// Act
	ctx.SetCookie(testCookieName, testCookieValue, 3600, "/path", "example.com", true, true)

	// Assert
	setCookieHeader := res.Header().Get(common.HeaderSetCookie)
	assert.Contains(t, setCookieHeader, "test-cookie=test-value")
	assert.Contains(t, setCookieHeader, "Max-Age=3600")
	assert.Contains(t, setCookieHeader, "Path=/path")
	assert.Contains(t, setCookieHeader, "Domain=example.com")
	assert.Contains(t, setCookieHeader, "Secure")
	assert.Contains(t, setCookieHeader, "HttpOnly")
	assert.Contains(t, setCookieHeader, "SameSite=Lax")
}

func TestShouldReturnCookieValueGivenExistingCookie(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: testCookieName, Value: testCookieValue})
	res := httptest.NewRecorder()
	ctx := NewRouteContext(res, req)

	// Act
	value, err := ctx.GetCookie(testCookieName)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, testCookieValue, value)
}

func TestShouldReturnErrorGivenMissingCookie(t *testing.T) {
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

func TestShouldSetExpiredCookieGivenClearCookie(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewRouteContext(res, req)

	// Act
	ctx.ClearCookie(testCookieName)

	// Assert
	setCookieHeader := res.Header().Get(common.HeaderSetCookie)
	assert.Contains(t, setCookieHeader, "test-cookie=")
	assert.Contains(t, setCookieHeader, "Max-Age=0") // Go changes -1 to 0 in the output
	assert.Contains(t, setCookieHeader, "Expires=Thu, 01 Jan 1970 00:00:01 GMT")
	assert.Contains(t, setCookieHeader, "Path=/")
	assert.Contains(t, setCookieHeader, "HttpOnly")
	assert.Contains(t, setCookieHeader, "Secure")
	assert.Contains(t, setCookieHeader, "SameSite=Lax")
}

func TestShouldClearAllCookiesAndRedirectGivenSignOut(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodPost, "/signout", nil)
	res := httptest.NewRecorder()
	ctx := NewRouteContext(res, req)

	// Act
	ctx.SignOut("http://example.com/logout")

	// Assert
	setCookieHeaders := res.Header().Values(common.HeaderSetCookie)

	// Should clear all three cookies
	userCookieCleared := false
	twoFactorCookieCleared := false
	idpSessionCookieCleared := false
	csrfCookieCleared := false

	for _, header := range setCookieHeaders {
		if strings.Contains(header, cookiekit.GetUserCookieName()) {
			userCookieCleared = true
		}
		if strings.Contains(header, cookiekit.GetTwoFactorCookieName()) {
			twoFactorCookieCleared = true
		}
		if strings.Contains(header, cookiekit.GetIdpSessionCookieName()) {
			idpSessionCookieCleared = true
		}
		if strings.Contains(header, "csrf_token") {
			csrfCookieCleared = true
		}
	}

	assert.True(t, userCookieCleared, "User cookie should be cleared")
	assert.True(t, twoFactorCookieCleared, "Two-factor cookie should be cleared")
	assert.True(t, idpSessionCookieCleared, "IDP session cookie should be cleared")
	assert.True(t, csrfCookieCleared, "CSRF cookie should be cleared")

	// Should redirect to logout page
	assert.Equal(t, http.StatusTemporaryRedirect, res.Code)
	assert.Equal(t, "http://example.com/logout", res.Header().Get(common.HeaderLocation))
}

func TestShouldClearCustomCookieAttributesGivenSignOutWithOptions(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodPost, "/signout", nil)
	res := httptest.NewRecorder()
	ctx := NewRouteContext(res, req)

	// Act
	SignOutWithOptions(ctx, "http://example.com/logout",
		cookiekit.WithDomain(".example.com"),
		cookiekit.WithPath("/app"),
		cookiekit.WithSameSite(http.SameSiteStrictMode),
	)

	// Assert
	setCookieHeaders := res.Header().Values(common.HeaderSetCookie)
	require.Len(t, setCookieHeaders, 4)

	for _, header := range setCookieHeaders {
		assert.Contains(t, header, "Domain=example.com")
		assert.Contains(t, header, "Path=/app")
		assert.Contains(t, header, "Expires=Thu, 01 Jan 1970 00:00:01 GMT")
	}

	assert.Equal(t, http.StatusTemporaryRedirect, res.Code)
	assert.Equal(t, "http://example.com/logout", res.Header().Get(common.HeaderLocation))
}

func TestShouldApplyCookieOptionsWhenSigningIn(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodPost, "/signin", nil)
	res := httptest.NewRecorder()
	ctx := NewRouteContext(res, req)

	// Set up the service in context (simulating middleware)
	provider := &mockProvider{ttl: 30 * time.Minute}
	ctx.SetService(tokenizer.ServiceKeyTokenProvider, provider)

	mockUser := newMockPrincipal(testUserSubject)

	// Act - Sign in with custom cookie options
	ctx.SignIn(mockUser, "/dashboard",
		cookiekit.WithDomain(".example.com"),
		cookiekit.WithPath("/app"),
		cookiekit.WithSameSite(http.SameSiteStrictMode),
		cookiekit.WithMaxAge(7200),
	)

	// Assert
	setCookieHeader := res.Header().Get(common.HeaderSetCookie)
	assert.Contains(t, setCookieHeader, cookiekit.GetUserCookieName())
	assert.Contains(t, setCookieHeader, "test-token-test-user")
	assert.Contains(t, setCookieHeader, "Domain=example.com", "Should use custom domain")
	assert.Contains(t, setCookieHeader, "Path=/app", "Should use custom path")
	assert.Contains(t, setCookieHeader, "SameSite=Strict", "Should use custom SameSite")
	assert.Contains(t, setCookieHeader, "Max-Age=7200", "Should use custom MaxAge")
	assert.Contains(t, setCookieHeader, "Secure", "Should still be secure by default")
	assert.Contains(t, setCookieHeader, "HttpOnly", "Should still be HttpOnly by default")

	// Should redirect to dashboard
	assert.Equal(t, http.StatusTemporaryRedirect, res.Code)
	assert.Contains(t, res.Header().Get(common.HeaderLocation), "/dashboard", "Should redirect to dashboard")
}

func TestShouldUseProviderTTLWhenMaxAgeNotSpecified(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodPost, "/signin", nil)
	res := httptest.NewRecorder()
	ctx := NewRouteContext(res, req)

	// Set up the service with 1 hour TTL
	provider := &mockProvider{ttl: 1 * time.Hour}
	ctx.SetService(tokenizer.ServiceKeyTokenProvider, provider)

	mockUser := newMockPrincipal(testUserSubject)

	// Act - Sign in without specifying MaxAge
	ctx.SignIn(mockUser, "/",
		cookiekit.WithDomain(".example.com"),
	)

	// Assert - Should use provider TTL (3600 seconds)
	setCookieHeader := res.Header().Get(common.HeaderSetCookie)
	assert.Contains(t, setCookieHeader, "Max-Age=3600", "Should use provider TTL when MaxAge not specified")
	assert.Contains(t, setCookieHeader, "Domain=example.com")
}

func TestShouldAllowOverridingSecureAndHttpOnly(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodPost, "/signin", nil)
	res := httptest.NewRecorder()
	ctx := NewRouteContext(res, req)

	provider := &mockProvider{ttl: 30 * time.Minute}
	ctx.SetService(tokenizer.ServiceKeyTokenProvider, provider)

	mockUser := newMockPrincipal(testUserSubject)

	// Act - Sign in with custom Secure and HttpOnly flags (for testing purposes)
	ctx.SignIn(mockUser, "/",
		cookiekit.WithSecure(false),
		cookiekit.WithHttpOnly(false),
	)

	// Assert
	setCookieHeader := res.Header().Get(common.HeaderSetCookie)
	assert.NotContains(t, setCookieHeader, "; Secure", "Should allow disabling Secure flag")
	assert.NotContains(t, setCookieHeader, "; HttpOnly", "Should allow disabling HttpOnly flag")
}

func TestShouldApplySameSiteNoneWithSecure(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodPost, "/signin", nil)
	res := httptest.NewRecorder()
	ctx := NewRouteContext(res, req)

	provider := &mockProvider{ttl: 30 * time.Minute}
	ctx.SetService(tokenizer.ServiceKeyTokenProvider, provider)

	mockUser := newMockPrincipal(testUserSubject)

	// Act - Sign in with SameSite=None (requires Secure=true)
	ctx.SignIn(mockUser, "/",
		cookiekit.WithSameSite(http.SameSiteNoneMode),
		cookiekit.WithSecure(true),
	)

	// Assert
	setCookieHeader := res.Header().Get(common.HeaderSetCookie)
	assert.Contains(t, setCookieHeader, "SameSite=None")
	assert.Contains(t, setCookieHeader, "Secure")
}

func TestShouldAuthenticateWithCustomCookieOptions(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewRouteContext(res, req)

	provider := &mockProvider{ttl: 15 * time.Minute}
	ctx.SetService(tokenizer.ServiceKeyTokenProvider, provider)

	mockUser := newMockPrincipal(testUserSubject)

	// Act - Authenticate with custom options
	ctx.Authenticate("custom-cookie", mockUser,
		cookiekit.WithDomain(".api.example.com"),
		cookiekit.WithPath("/api"),
		cookiekit.WithSameSite(http.SameSiteLaxMode),
		cookiekit.WithMaxAge(3600),
	)

	// Assert
	setCookieHeader := res.Header().Get(common.HeaderSetCookie)
	assert.Contains(t, setCookieHeader, "custom-cookie=test-token-test-user")
	assert.Contains(t, setCookieHeader, "Domain=api.example.com")
	assert.Contains(t, setCookieHeader, "Path=/api")
	assert.Contains(t, setCookieHeader, "Max-Age=3600")
	assert.Contains(t, setCookieHeader, "SameSite=Lax")
}
