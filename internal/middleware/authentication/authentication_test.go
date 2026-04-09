package authentication

import (
	"crypto/rand"
	"crypto/tls"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fgrzl/claims"
	"github.com/fgrzl/mux/internal/common"
	"github.com/fgrzl/mux/internal/cookiekit"
	"github.com/fgrzl/mux/internal/router"
	"github.com/fgrzl/mux/internal/routing"
	"github.com/fgrzl/mux/internal/tokenizer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockPrincipal implements claims.Principal for testing
type mockPrincipal struct {
	subject   string
	issuer    string
	audience  []string
	expTime   int64
	notBefore int64
	issuedAt  int64
	jwti      string
	scopes    []string
	roles     []string
	email     string
	username  string
	// customClaims intentionally omitted; not required for current tests
}

func (m *mockPrincipal) Subject() string                      { return m.subject }
func (m *mockPrincipal) Issuer() string                       { return m.issuer }
func (m *mockPrincipal) Audience() []string                   { return m.audience }
func (m *mockPrincipal) ExpirationTime() int64                { return m.expTime }
func (m *mockPrincipal) NotBefore() int64                     { return m.notBefore }
func (m *mockPrincipal) IssuedAt() int64                      { return m.issuedAt }
func (m *mockPrincipal) JWTI() string                         { return m.jwti }
func (m *mockPrincipal) Scopes() []string                     { return m.scopes }
func (m *mockPrincipal) Roles() []string                      { return m.roles }
func (m *mockPrincipal) Email() string                        { return m.email }
func (m *mockPrincipal) Username() string                     { return m.username }
func (m *mockPrincipal) CustomClaim(name string) claims.Claim { return nil }
func (m *mockPrincipal) CustomClaimValue(name string) string  { return "" }
func (m *mockPrincipal) Claims() *claims.ClaimSet             { return nil }

func newMockPrincipal(subject string) *mockPrincipal {
	return &mockPrincipal{
		subject:  subject,
		audience: []string{"test"},
		scopes:   []string{"read"},
		roles:    []string{"user"},
	}
}

func TestDefaultTokenProviderShouldImplementTokenProviderInterface(t *testing.T) {
	// Arrange
	provider := &defaultTokenProvider{}

	// Act & Assert
	var _ tokenizer.TokenProvider = provider // This will fail to compile if interface is not implemented
}

func TestDefaultTokenProviderShouldCreateTokenWhenSignFnIsSet(t *testing.T) {
	// Arrange
	expectedToken := "test-token"
	mockUser := newMockPrincipal("user123")
	ttl := 30 * time.Minute
	provider := &defaultTokenProvider{ttl: ttl, signFn: func(user claims.Principal, duration time.Duration) (string, error) {
		assert.Equal(t, mockUser, user)
		assert.Equal(t, ttl, duration)
		return expectedToken, nil
	}}
	ctx, _ := newRouteContext(nil)
	// Act
	token, err := provider.CreateToken(ctx, mockUser)
	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedToken, token)
}

func TestDefaultTokenProviderShouldReturnErrorWhenSignFnIsNotSet(t *testing.T) {
	// Arrange
	provider := &defaultTokenProvider{}
	mockUser := newMockPrincipal("user123")

	ctx, _ := newRouteContext(nil)

	// Act
	token, err := provider.CreateToken(ctx, mockUser)

	// Assert
	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "signing function is not set")
}

func TestDefaultTokenProviderShouldReturnErrorWhenSignFnFails(t *testing.T) {
	// Arrange
	expectedError := errors.New("signing failed")
	provider := &defaultTokenProvider{ttl: 30 * time.Minute, signFn: func(claims.Principal, time.Duration) (string, error) { return "", expectedError }}
	mockUser := newMockPrincipal("user123")
	ctx, _ := newRouteContext(nil)
	// Act
	token, err := provider.CreateToken(ctx, mockUser)
	// Assert
	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Equal(t, expectedError, err)
}

func TestDefaultTokenProviderShouldValidateTokenWhenValidateFnIsSet(t *testing.T) {
	// Arrange
	testToken := "valid-token"
	mockUser := newMockPrincipal("user123")
	provider := &defaultTokenProvider{validateFn: func(token string) (claims.Principal, error) { assert.Equal(t, testToken, token); return mockUser, nil }}
	ctx, _ := newRouteContext(nil)
	// Act
	principal, err := provider.ValidateToken(ctx, testToken)
	// Assert
	require.NoError(t, err)
	assert.Equal(t, mockUser, principal)
	assert.Equal(t, "user123", principal.Subject())
}

func TestDefaultTokenProviderShouldReturnErrorWhenValidateFnIsNotSet(t *testing.T) {
	// Arrange
	provider := &defaultTokenProvider{}

	ctx, _ := newRouteContext(nil)

	// Act
	principal, err := provider.ValidateToken(ctx, "any-token")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, principal)
	assert.Contains(t, err.Error(), "validation function is not set")
}

func TestDefaultTokenProviderShouldReturnErrorWhenValidateFnFails(t *testing.T) {
	// Arrange
	expectedError := errors.New("token invalid")
	provider := &defaultTokenProvider{validateFn: func(string) (claims.Principal, error) { return nil, expectedError }}
	ctx, _ := newRouteContext(nil)
	// Act
	principal, err := provider.ValidateToken(ctx, "invalid-token")
	// Assert
	assert.Error(t, err)
	assert.Nil(t, principal)
	assert.Equal(t, expectedError, err)
}

func TestAuthenticationMiddlewareShouldSetTokenProviderAsService(t *testing.T) {
	// Arrange
	rtr := router.NewRouter()
	UseAuthentication(rtr,
		WithTokenTTL(30*time.Minute),
		WithValidator(func(string) (claims.Principal, error) {
			return newMockPrincipal("test"), nil
		}),
	)

	ctx, res := newRouteContext(nil)
	var capturedContext routing.RouteContext
	handler := func(c routing.RouteContext) {
		capturedContext = c
		c.OK("success")
	}

	rtr.GET("/test", handler).AllowAnonymous()

	// Act
	rtr.ServeHTTP(res, ctx.Request())

	// Assert
	require.NotNil(t, capturedContext)
	service, ok := capturedContext.GetService(tokenizer.ServiceKeyTokenProvider)
	assert.True(t, ok)
	assert.NotNil(t, service)

	// Verify it implements TokenProvider interface
	provider, ok := service.(tokenizer.TokenProvider)
	assert.True(t, ok)
	assert.NotNil(t, provider)
}

func TestAuthenticationMiddlewareShouldAllowAnonymousAccessWhenConfigured(t *testing.T) {
	// Arrange
	rtr := router.NewRouter()
	UseAuthentication(rtr,
		WithValidator(func(string) (claims.Principal, error) {
			return nil, errors.New("should not be called")
		}),
	)

	ctx, res := newRouteContext(nil)
	handler := func(c routing.RouteContext) {
		c.OK("success")
	}

	rtr.GET("/test", handler).AllowAnonymous()

	// Act
	rtr.ServeHTTP(res, ctx.Request())

	// Assert
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), "success")
}

func TestAuthenticationMiddlewareShouldRejectRequestWithoutValidAuthentication(t *testing.T) {
	// Arrange
	rtr := router.NewRouter()
	UseAuthentication(rtr,
		WithValidator(func(string) (claims.Principal, error) {
			return nil, ErrInvalidToken
		}),
	)

	ctx, res := newRouteContext(nil)
	rtr.GET("/test", func(c routing.RouteContext) {
		c.OK("success")
	})

	// Act
	rtr.ServeHTTP(res, ctx.Request())

	// Assert
	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestAuthenticationMiddlewareShouldAuthenticateViaCookie(t *testing.T) {
	// Arrange
	mockUser := newMockPrincipal("user123")
	rtr := router.NewRouter()
	UseAuthentication(rtr,
		WithValidator(func(token string) (claims.Principal, error) {
			if token == "valid-cookie-token" {
				return mockUser, nil
			}
			return nil, ErrInvalidToken
		}),
	)

	ctx, res := newRouteContext(func(r *http.Request) {
		r.AddCookie(&http.Cookie{
			Name:  cookiekit.GetUserCookieName(),
			Value: "valid-cookie-token",
		})
	})

	var authenticatedUser claims.Principal
	rtr.GET("/test", func(c routing.RouteContext) {
		authenticatedUser = c.User()
		c.OK("success")
	})

	// Act
	rtr.ServeHTTP(res, ctx.Request())

	// Assert
	assert.Equal(t, http.StatusOK, res.Code)
	assert.NotNil(t, authenticatedUser)
	assert.Equal(t, "user123", authenticatedUser.Subject())
}

func TestAuthenticationMiddlewareShouldApplyConfiguredCookieOptionsWhenSigningIn(t *testing.T) {
	mockUser := newMockPrincipal("user123")
	rtr := router.NewRouter()
	UseAuthentication(rtr,
		WithTokenTTL(30*time.Minute),
		WithTokenCreator(func(principal claims.Principal, ttl time.Duration) (string, error) {
			assert.Equal(t, mockUser, principal)
			assert.Equal(t, 30*time.Minute, ttl)
			return "signed-user123", nil
		}),
		WithCookieOptions(
			cookiekit.WithDomain(".example.com"),
			cookiekit.WithPath("/portal"),
			cookiekit.WithSameSite(http.SameSiteNoneMode),
			cookiekit.WithSecure(true),
		),
	)

	ctx, res := newRouteContext(nil)
	rtr.GET("/test", func(c routing.RouteContext) {
		c.SignIn(mockUser, "/dashboard")
	}).AllowAnonymous()

	rtr.ServeHTTP(res, ctx.Request())

	assert.Equal(t, http.StatusTemporaryRedirect, res.Code)
	setCookieHeader := res.Header().Get(common.HeaderSetCookie)
	assert.Contains(t, setCookieHeader, "app_token=signed-user123")
	assert.Contains(t, setCookieHeader, "Domain=example.com")
	assert.Contains(t, setCookieHeader, "Path=/portal")
	assert.Contains(t, setCookieHeader, "SameSite=None")
	assert.Contains(t, setCookieHeader, "Max-Age=1800")
	assert.Contains(t, setCookieHeader, "Secure")
	assert.Contains(t, setCookieHeader, "HttpOnly")
}

func TestAuthenticationMiddlewareShouldRenewCookieWithConfiguredCookieOptions(t *testing.T) {
	mockUser := newMockPrincipal("user123")
	rtr := router.NewRouter()
	UseAuthentication(rtr,
		WithTokenTTL(30*time.Minute),
		WithValidator(func(token string) (claims.Principal, error) {
			if token == "valid-cookie-token" {
				return mockUser, nil
			}
			return nil, ErrInvalidToken
		}),
		WithTokenCreator(func(principal claims.Principal, ttl time.Duration) (string, error) {
			assert.Equal(t, mockUser, principal)
			assert.Equal(t, 30*time.Minute, ttl)
			return "renewed-user123", nil
		}),
		WithCookieOptions(
			cookiekit.WithDomain(".example.com"),
			cookiekit.WithPath("/"),
			cookiekit.WithSameSite(http.SameSiteLaxMode),
			cookiekit.WithSecure(true),
		),
	)

	ctx, res := newRouteContext(func(r *http.Request) {
		r.AddCookie(&http.Cookie{
			Name:  cookiekit.GetUserCookieName(),
			Value: "valid-cookie-token",
		})
	})

	rtr.GET("/test", func(c routing.RouteContext) {
		c.OK("success")
	})

	rtr.ServeHTTP(res, ctx.Request())

	assert.Equal(t, http.StatusOK, res.Code)
	setCookieHeader := res.Header().Get(common.HeaderSetCookie)
	assert.Contains(t, setCookieHeader, "app_token=renewed-user123")
	assert.Contains(t, setCookieHeader, "Domain=example.com")
	assert.Contains(t, setCookieHeader, "Path=/")
	assert.Contains(t, setCookieHeader, "SameSite=Lax")
	assert.Contains(t, setCookieHeader, "Max-Age=1800")
	assert.Contains(t, setCookieHeader, "Secure")
	assert.Contains(t, setCookieHeader, "HttpOnly")
}

func TestAuthenticationMiddlewareShouldApplyConfiguredCookieOptionsWhenSigningOut(t *testing.T) {
	rtr := router.NewRouter()
	UseAuthentication(rtr,
		WithCookieOptions(
			cookiekit.WithDomain(".example.com"),
			cookiekit.WithPath("/portal"),
		),
	)

	ctx, res := newRouteContext(nil)
	rtr.GET("/test", func(c routing.RouteContext) {
		c.SignOut("/signed-out")
	}).AllowAnonymous()

	rtr.ServeHTTP(res, ctx.Request())

	assert.Equal(t, http.StatusTemporaryRedirect, res.Code)
	setCookieHeaders := res.Header().Values(common.HeaderSetCookie)
	require.Len(t, setCookieHeaders, 4)
	for _, header := range setCookieHeaders {
		assert.Contains(t, header, "Domain=example.com")
		assert.Contains(t, header, "Path=/portal")
		assert.Contains(t, header, "Expires=Thu, 01 Jan 1970 00:00:01 GMT")
	}
}

func TestAuthenticationMiddlewareShouldAuthenticateViaBearerToken(t *testing.T) {
	// Arrange
	mockUser := newMockPrincipal("user456")
	rtr := router.NewRouter()
	UseAuthentication(rtr,
		WithValidator(func(token string) (claims.Principal, error) {
			if token == "valid-bearer-token" {
				return mockUser, nil
			}
			return nil, ErrInvalidToken
		}),
	)

	ctx, res := newRouteContext(func(r *http.Request) {
		r.Header.Set(common.HeaderAuthorization, "Bearer valid-bearer-token")
	})

	var authenticatedUser claims.Principal
	rtr.GET("/test", func(c routing.RouteContext) {
		authenticatedUser = c.User()
		c.OK("success")
	})

	// Act
	rtr.ServeHTTP(res, ctx.Request())

	// Assert
	assert.Equal(t, http.StatusOK, res.Code)
	assert.NotNil(t, authenticatedUser)
	assert.Equal(t, "user456", authenticatedUser.Subject())
}

// newBenchAuthMiddleware returns an authenticationMiddleware with a validator that accepts
// the token "valid-token" and returns a mock principal. Extracted to avoid duplication
// across benchmarks and tests.
func newBenchAuthMiddleware() *authenticationMiddleware {
	return &authenticationMiddleware{provider: &defaultTokenProvider{validateFn: func(token string) (claims.Principal, error) {
		if token == "valid-token" {
			return newMockPrincipal("bench-user"), nil
		}
		return nil, errors.New("invalid")
	}}}
}

// newBenchRouteContext creates a routing.RouteContext for benchmarks and tests.
// The optional setup function can modify the request (e.g., set headers/cookies).
func newBenchRouteContext(setup func(r *http.Request)) routing.RouteContext {
	req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
	rec := httptest.NewRecorder()
	if setup != nil {
		setup(req)
	}
	return routing.NewRouteContext(rec, req)
}

// newRouteContext creates a request, recorder and routing.RouteContext for tests.
// Returns the RouteContext and the underlying ResponseRecorder so tests can inspect the response.
func newRouteContext(setup func(r *http.Request)) (routing.RouteContext, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	if setup != nil {
		setup(req)
	}
	return routing.NewRouteContext(rec, req), rec
}

// ---- Tests for Bearer Token Case Insensitivity (RFC 7235) ----

func TestExtractBearerTokenShouldBeCaseInsensitive(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{"standard Bearer", "Bearer valid-token", "valid-token"},
		{"lowercase bearer", "bearer valid-token", "valid-token"},
		{"uppercase BEARER", "BEARER valid-token", "valid-token"},
		{"mixed case BeArEr", "BeArEr valid-token", "valid-token"},
		{"empty header", "", ""},
		{"no token", "Bearer ", ""},
		{"wrong scheme", "Basic dXNlcjpwYXNz", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractBearerToken(tt.header)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ---- Tests for X-Forwarded-Proto Support ----

func TestIsSecureRequestShouldDetectTLS(t *testing.T) {
	// Test direct TLS
	req := httptest.NewRequest(http.MethodGet, "https://example.com/test", nil)
	req.TLS = &tls.ConnectionState{} // Simulate TLS
	assert.True(t, isSecureRequest(req))
}

func TestIsSecureRequestShouldIgnoreXForwardedProtoWithoutTrustedMiddleware(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	assert.False(t, isSecureRequest(req))
}

func TestIsSecureRequestShouldRespectRequestScheme(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
	req.URL.Scheme = "https"
	assert.True(t, isSecureRequest(req))
}

func TestIsSecureRequestShouldReturnFalseForHTTP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
	assert.False(t, isSecureRequest(req))
}

// ---- Tests for CSRF Protection ----

func TestCSRFValidationShouldPassWithMatchingTokens(t *testing.T) {
	rtr := router.NewRouter()
	mockUser := newMockPrincipal("user123")

	UseAuthentication(rtr,
		WithCSRFProtection(),
		WithValidator(func(token string) (claims.Principal, error) {
			if token == "valid-token" {
				return mockUser, nil
			}
			return nil, ErrInvalidToken
		}),
	)

	csrfToken := "test-csrf-token-12345"
	ctx, res := newRouteContext(func(r *http.Request) {
		r.Method = http.MethodPost
		r.AddCookie(&http.Cookie{Name: cookiekit.GetUserCookieName(), Value: "valid-token"})
		r.AddCookie(&http.Cookie{Name: csrfTokenCookieName, Value: csrfToken})
		r.Header.Set(csrfTokenHeaderName, csrfToken)
	})

	rtr.POST("/test", func(c routing.RouteContext) {
		c.OK("success")
	})

	rtr.ServeHTTP(res, ctx.Request())
	assert.Equal(t, http.StatusOK, res.Code)
}

func TestCSRFValidationShouldFailWithMismatchedTokens(t *testing.T) {
	rtr := router.NewRouter()
	mockUser := newMockPrincipal("user123")

	UseAuthentication(rtr,
		WithCSRFProtection(),
		WithValidator(func(token string) (claims.Principal, error) {
			if token == "valid-token" {
				return mockUser, nil
			}
			return nil, ErrInvalidToken
		}),
	)

	ctx, res := newRouteContext(func(r *http.Request) {
		r.Method = http.MethodPost
		r.AddCookie(&http.Cookie{Name: cookiekit.GetUserCookieName(), Value: "valid-token"})
		r.AddCookie(&http.Cookie{Name: csrfTokenCookieName, Value: "csrf-token-1"})
		r.Header.Set(csrfTokenHeaderName, "csrf-token-2") // Mismatched!
	})

	rtr.POST("/test", func(c routing.RouteContext) {
		c.OK("success")
	})

	rtr.ServeHTTP(res, ctx.Request())
	assert.Equal(t, http.StatusForbidden, res.Code)
}

func TestCSRFValidationShouldNotApplyToGETRequests(t *testing.T) {
	rtr := router.NewRouter()
	mockUser := newMockPrincipal("user123")

	UseAuthentication(rtr,
		WithCSRFProtection(),
		WithValidator(func(token string) (claims.Principal, error) {
			if token == "valid-token" {
				return mockUser, nil
			}
			return nil, ErrInvalidToken
		}),
	)

	// GET request without CSRF token should succeed
	ctx, res := newRouteContext(func(r *http.Request) {
		r.Method = http.MethodGet
		r.AddCookie(&http.Cookie{Name: cookiekit.GetUserCookieName(), Value: "valid-token"})
	})

	rtr.GET("/test", func(c routing.RouteContext) {
		c.OK("success")
	})

	rtr.ServeHTTP(res, ctx.Request())
	assert.Equal(t, http.StatusOK, res.Code)
}

func TestCSRFValidationShouldNotApplyToBearerAuth(t *testing.T) {
	rtr := router.NewRouter()
	mockUser := newMockPrincipal("user123")

	UseAuthentication(rtr,
		WithCSRFProtection(),
		WithValidator(func(token string) (claims.Principal, error) {
			if token == "valid-token" {
				return mockUser, nil
			}
			return nil, ErrInvalidToken
		}),
	)

	// POST with Bearer token should succeed without CSRF
	ctx, res := newRouteContext(func(r *http.Request) {
		r.Method = http.MethodPost
		r.Header.Set(common.HeaderAuthorization, "Bearer valid-token")
	})

	rtr.POST("/test", func(c routing.RouteContext) {
		c.OK("success")
	})

	rtr.ServeHTTP(res, ctx.Request())
	assert.Equal(t, http.StatusOK, res.Code)
}

// ---- Tests for Rate Limiting ----

func TestRateLimiterShouldBlockExcessiveFailures(t *testing.T) {
	limiter := NewInMemoryRateLimiter(3, time.Minute)

	// First 3 attempts should be allowed
	assert.True(t, limiter("client1"))
	assert.True(t, limiter("client1"))
	assert.True(t, limiter("client1"))

	// 4th attempt should be blocked
	assert.False(t, limiter("client1"))

	// Different client should still be allowed
	assert.True(t, limiter("client2"))
}

func TestAuthenticationWithRateLimitingShouldRejectRateLimitedRequests(t *testing.T) {
	blockedClients := make(map[string]bool)

	rtr := router.NewRouter()
	UseAuthentication(rtr,
		WithValidator(func(token string) (claims.Principal, error) {
			return nil, ErrInvalidToken
		}),
		WithRateLimiter(func(clientID string) bool {
			return !blockedClients[clientID]
		}),
	)

	// Block the client (httptest uses 192.0.2.1 as default RemoteAddr)
	blockedClients["192.0.2.1"] = true

	ctx, res := newRouteContext(nil)
	rtr.GET("/test", func(c routing.RouteContext) {
		c.OK("success")
	})

	rtr.ServeHTTP(res, ctx.Request())
	assert.Equal(t, http.StatusTooManyRequests, res.Code)
}

func TestAuthenticationWithRateLimitingShouldNotBlockValidAuthenticatedRequests(t *testing.T) {
	limiterCalled := false
	mockUser := newMockPrincipal("user123")

	rtr := router.NewRouter()
	UseAuthentication(rtr,
		WithValidator(func(token string) (claims.Principal, error) {
			if token == "valid-token" {
				return mockUser, nil
			}
			return nil, ErrInvalidToken
		}),
		WithRateLimiter(func(clientID string) bool {
			limiterCalled = true
			return false
		}),
	)

	ctx, res := newRouteContext(func(r *http.Request) {
		r.Header.Set(common.HeaderAuthorization, "Bearer valid-token")
	})
	rtr.GET("/test", func(c routing.RouteContext) {
		c.OK("success")
	})

	rtr.ServeHTTP(res, ctx.Request())

	assert.Equal(t, http.StatusOK, res.Code)
	assert.False(t, limiterCalled)
}

// ---- Tests for Token Revocation ----

func TestTokenRevocationShouldBlockRevokedTokens(t *testing.T) {
	rtr := router.NewRouter()
	mockUser := newMockPrincipal("user123")
	revokedTokens := map[string]bool{"revoked-token": true}

	UseAuthentication(rtr,
		WithValidator(func(token string) (claims.Principal, error) {
			return mockUser, nil
		}),
		WithTokenRevocationChecker(func(token string) bool {
			return revokedTokens[token]
		}),
	)

	// Valid token should work
	ctx, res := newRouteContext(func(r *http.Request) {
		r.Header.Set(common.HeaderAuthorization, "Bearer valid-token")
	})
	rtr.GET("/test", func(c routing.RouteContext) {
		c.OK("success")
	})
	rtr.ServeHTTP(res, ctx.Request())
	assert.Equal(t, http.StatusOK, res.Code)

	// Revoked token should fail
	ctx2, res2 := newRouteContext(func(r *http.Request) {
		r.Header.Set(common.HeaderAuthorization, "Bearer revoked-token")
	})
	rtr.ServeHTTP(res2, ctx2.Request())
	assert.Equal(t, http.StatusUnauthorized, res2.Code)
}

// ---- Tests for Issuer/Audience Validation ----

func TestIssuerValidationShouldRejectInvalidIssuer(t *testing.T) {
	rtr := router.NewRouter()
	mockUser := &mockPrincipal{
		subject: "user123",
		issuer:  "wrong-issuer",
	}

	UseAuthentication(rtr,
		WithValidator(func(token string) (claims.Principal, error) {
			return mockUser, nil
		}),
		WithIssuerValidator("expected-issuer"),
	)

	ctx, res := newRouteContext(func(r *http.Request) {
		r.Header.Set(common.HeaderAuthorization, "Bearer valid-token")
	})
	rtr.GET("/test", func(c routing.RouteContext) {
		c.OK("success")
	})

	rtr.ServeHTTP(res, ctx.Request())
	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestIssuerValidationShouldAcceptValidIssuer(t *testing.T) {
	rtr := router.NewRouter()
	mockUser := &mockPrincipal{
		subject: "user123",
		issuer:  "expected-issuer",
	}

	UseAuthentication(rtr,
		WithValidator(func(token string) (claims.Principal, error) {
			return mockUser, nil
		}),
		WithIssuerValidator("expected-issuer"),
	)

	ctx, res := newRouteContext(func(r *http.Request) {
		r.Header.Set(common.HeaderAuthorization, "Bearer valid-token")
	})
	rtr.GET("/test", func(c routing.RouteContext) {
		c.OK("success")
	})

	rtr.ServeHTTP(res, ctx.Request())
	assert.Equal(t, http.StatusOK, res.Code)
}

func TestAudienceValidationShouldRejectInvalidAudience(t *testing.T) {
	rtr := router.NewRouter()
	mockUser := &mockPrincipal{
		subject:  "user123",
		audience: []string{"other-audience"},
	}

	UseAuthentication(rtr,
		WithValidator(func(token string) (claims.Principal, error) {
			return mockUser, nil
		}),
		WithAudienceValidator("expected-audience"),
	)

	ctx, res := newRouteContext(func(r *http.Request) {
		r.Header.Set(common.HeaderAuthorization, "Bearer valid-token")
	})
	rtr.GET("/test", func(c routing.RouteContext) {
		c.OK("success")
	})

	rtr.ServeHTTP(res, ctx.Request())
	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestAudienceValidationShouldAcceptValidAudience(t *testing.T) {
	rtr := router.NewRouter()
	mockUser := &mockPrincipal{
		subject:  "user123",
		audience: []string{"other", "expected-audience"},
	}

	UseAuthentication(rtr,
		WithValidator(func(token string) (claims.Principal, error) {
			return mockUser, nil
		}),
		WithAudienceValidator("expected-audience"),
	)

	ctx, res := newRouteContext(func(r *http.Request) {
		r.Header.Set(common.HeaderAuthorization, "Bearer valid-token")
	})
	rtr.GET("/test", func(c routing.RouteContext) {
		c.OK("success")
	})

	rtr.ServeHTTP(res, ctx.Request())
	assert.Equal(t, http.StatusOK, res.Code)
}

// ---- Tests for Client ID Extraction ----

func TestGetClientIDShouldIgnoreXForwardedForWithoutTrustedMiddleware(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.195, 70.41.3.18, 150.172.238.178")
	assert.Equal(t, "192.168.1.1", getClientID(req))
}

func TestGetClientIDShouldIgnoreXRealIPWithoutTrustedMiddleware(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	req.Header.Set("X-Real-IP", "203.0.113.195")
	assert.Equal(t, "192.168.1.1", getClientID(req))
}

func TestGetClientIDShouldFallbackToRemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	assert.Equal(t, "192.168.1.1", getClientID(req))
}

func TestGetClientIDShouldParseIPv6RemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "[2001:db8::1]:12345"
	assert.Equal(t, "2001:db8::1", getClientID(req))
}

// ---- Tests for CSRF Token Generation ----

func TestGenerateCSRFTokenShouldSetCookie(t *testing.T) {
	ctx, res := newRouteContext(nil)
	token := GenerateCSRFToken(ctx)

	assert.NotEmpty(t, token)
	assert.Len(t, token, csrfTokenLength)

	// Check that cookie was set
	cookies := res.Result().Cookies()
	var csrfCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == csrfTokenCookieName {
			csrfCookie = c
			break
		}
	}
	require.NotNil(t, csrfCookie)
	assert.Equal(t, token, csrfCookie.Value)
	assert.False(t, csrfCookie.HttpOnly) // Must be readable by JS
	assert.Equal(t, http.SameSiteStrictMode, csrfCookie.SameSite)
}

func TestGenerateCSRFTokenShouldIgnoreSpoofedForwardedProto(t *testing.T) {
	ctx, res := newRouteContext(func(r *http.Request) {
		r.Header.Set("X-Forwarded-Proto", "https")
	})
	token, err := GenerateCSRFTokenErr(ctx)

	require.NoError(t, err)
	assert.NotEmpty(t, token)

	cookies := res.Result().Cookies()
	var csrfCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == csrfTokenCookieName {
			csrfCookie = c
			break
		}
	}
	require.NotNil(t, csrfCookie)
	assert.False(t, csrfCookie.Secure)
}

func TestGenerateCSRFTokenErrShouldReturnErrorWhenEntropyFails(t *testing.T) {
	// Arrange
	ctx, res := newRouteContext(nil)
	originalReader := csrfTokenEntropySource
	csrfTokenEntropySource = failingCSRFEntropyReader{}
	t.Cleanup(func() {
		csrfTokenEntropySource = originalReader
	})

	// Act
	token, err := GenerateCSRFTokenErr(ctx)

	// Assert
	require.Error(t, err)
	assert.Empty(t, token)
	assert.Empty(t, res.Result().Cookies())
}

func TestGenerateCSRFTokenShouldReturnEmptyStringWhenEntropyFails(t *testing.T) {
	// Arrange
	ctx, res := newRouteContext(nil)
	originalReader := csrfTokenEntropySource
	csrfTokenEntropySource = failingCSRFEntropyReader{}
	t.Cleanup(func() {
		csrfTokenEntropySource = originalReader
	})

	// Act
	token := GenerateCSRFToken(ctx)

	// Assert
	assert.Empty(t, token)
	assert.Empty(t, res.Result().Cookies())
}

// ---- Tests for Context Enricher ----

type testContextKey string

const (
	tenantIDKey testContextKey = "tenant_id"
	userInfoKey testContextKey = "user_info"
)

type failingCSRFEntropyReader struct{}

func (failingCSRFEntropyReader) Read(p []byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

func TestGenerateSecureTokenShouldUseCryptoRandReaderByDefault(t *testing.T) {
	assert.Same(t, rand.Reader, csrfTokenEntropySource)
}

func TestContextEnricherShouldEnrichContextAfterAuthentication(t *testing.T) {
	rtr := router.NewRouter()
	mockUser := &mockPrincipal{
		subject: "user123",
	}

	UseAuthentication(rtr,
		WithValidator(func(token string) (claims.Principal, error) {
			if token == "valid-token" {
				return mockUser, nil
			}
			return nil, ErrInvalidToken
		}),
		WithContextEnricher(func(c routing.RouteContext) {
			principal := c.User()
			c.SetContextValue(tenantIDKey, "tenant-abc")
			c.SetContextValue(userInfoKey, map[string]string{
				"id":   principal.Subject(),
				"role": "admin",
			})
		}),
	)

	var capturedTenantID string
	var capturedUserInfo map[string]string

	ctx, res := newRouteContext(func(r *http.Request) {
		r.Header.Set(common.HeaderAuthorization, "Bearer valid-token")
	})

	rtr.GET("/test", func(c routing.RouteContext) {
		// Verify context values are accessible from request context
		if tid, ok := c.Request().Context().Value(tenantIDKey).(string); ok {
			capturedTenantID = tid
		}
		if ui, ok := c.Request().Context().Value(userInfoKey).(map[string]string); ok {
			capturedUserInfo = ui
		}
		c.OK("success")
	})

	rtr.ServeHTTP(res, ctx.Request())

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "tenant-abc", capturedTenantID)
	assert.NotNil(t, capturedUserInfo)
	assert.Equal(t, "user123", capturedUserInfo["id"])
	assert.Equal(t, "admin", capturedUserInfo["role"])
}

func TestContextEnricherShouldWorkWithCookieAuth(t *testing.T) {
	rtr := router.NewRouter()
	mockUser := newMockPrincipal("user456")

	UseAuthentication(rtr,
		WithValidator(func(token string) (claims.Principal, error) {
			if token == "cookie-token" {
				return mockUser, nil
			}
			return nil, ErrInvalidToken
		}),
		WithContextEnricher(func(c routing.RouteContext) {
			c.SetContextValue(tenantIDKey, "tenant-from-cookie")
		}),
	)

	var capturedTenantID string
	ctx, res := newRouteContext(func(r *http.Request) {
		r.AddCookie(&http.Cookie{Name: cookiekit.GetUserCookieName(), Value: "cookie-token"})
	})

	rtr.GET("/test", func(c routing.RouteContext) {
		if tid, ok := c.Request().Context().Value(tenantIDKey).(string); ok {
			capturedTenantID = tid
		}
		c.OK("success")
	})

	rtr.ServeHTTP(res, ctx.Request())

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "tenant-from-cookie", capturedTenantID)
}

func TestContextEnricherShouldNotBeCalledOnAuthFailure(t *testing.T) {
	rtr := router.NewRouter()
	enricherCalled := false

	UseAuthentication(rtr,
		WithValidator(func(token string) (claims.Principal, error) {
			return nil, ErrInvalidToken
		}),
		WithContextEnricher(func(c routing.RouteContext) {
			enricherCalled = true
		}),
	)

	ctx, res := newRouteContext(func(r *http.Request) {
		r.Header.Set(common.HeaderAuthorization, "Bearer invalid-token")
	})

	rtr.GET("/test", func(c routing.RouteContext) {
		c.OK("success")
	})

	rtr.ServeHTTP(res, ctx.Request())

	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.False(t, enricherCalled, "enricher should not be called on auth failure")
}

func TestContextEnricherShouldHandleNoOpEnricher(t *testing.T) {
	rtr := router.NewRouter()
	mockUser := newMockPrincipal("user789")

	UseAuthentication(rtr,
		WithValidator(func(token string) (claims.Principal, error) {
			return mockUser, nil
		}),
		WithContextEnricher(func(c routing.RouteContext) {
			// No-op enricher - doesn't add any context values
		}),
	)

	ctx, res := newRouteContext(func(r *http.Request) {
		r.Header.Set(common.HeaderAuthorization, "Bearer valid-token")
	})

	rtr.GET("/test", func(c routing.RouteContext) {
		// Should still work, user should be accessible
		assert.NotNil(t, c.User())
		assert.Equal(t, "user789", c.User().Subject())
		c.OK("success")
	})

	rtr.ServeHTTP(res, ctx.Request())
	assert.Equal(t, http.StatusOK, res.Code)
}
