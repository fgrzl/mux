package routing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fgrzl/claims"
	"github.com/fgrzl/mux/pkg/common"
	"github.com/fgrzl/mux/pkg/cookiejar"
	"github.com/fgrzl/mux/pkg/tokenizer"
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

	// Act & Assert
	defer func() {
		r := recover()
		require.NotNil(t, r, "Expected panic but none occurred")

		panicMsg := r.(string)
		assert.Contains(t, panicMsg, "DEVELOPMENT ERROR")
		assert.Contains(t, panicMsg, "UseAuthentication()")
		assert.Contains(t, panicMsg, "c.Authenticate()")
	}()

	ctx.Authenticate(testCookieName, mockUser)
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

	for _, header := range setCookieHeaders {
		if strings.Contains(header, cookiejar.GetUserCookieName()) {
			userCookieCleared = true
		}
		if strings.Contains(header, cookiejar.GetTwoFactorCookieName()) {
			twoFactorCookieCleared = true
		}
		if strings.Contains(header, cookiejar.GetIdpSessionCookieName()) {
			idpSessionCookieCleared = true
		}
	}

	assert.True(t, userCookieCleared, "User cookie should be cleared")
	assert.True(t, twoFactorCookieCleared, "Two-factor cookie should be cleared")
	assert.True(t, idpSessionCookieCleared, "IDP session cookie should be cleared")

	// Should redirect to logout page
	assert.Equal(t, http.StatusTemporaryRedirect, res.Code)
	assert.Equal(t, "http://example.com/logout", res.Header().Get(common.HeaderLocation))
}
