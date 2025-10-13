package authentication

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fgrzl/claims"
	"github.com/fgrzl/mux/pkg/common"
	"github.com/fgrzl/mux/pkg/cookiejar"
	"github.com/fgrzl/mux/pkg/router"
	"github.com/fgrzl/mux/pkg/routing"
	"github.com/fgrzl/mux/pkg/tokenizer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockPrincipal implements claims.Principal for testing
type mockPrincipal struct {
	subject      string
	issuer       string
	audience     []string
	expTime      int64
	notBefore    int64
	issuedAt     int64
	jwti         string
	scopes       []string
	roles        []string
	email        string
	username     string
	customClaims map[string]interface{}
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
			return nil, ErrorInvalidToken
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
			return nil, ErrorInvalidToken
		}),
	)

	ctx, res := newRouteContext(func(r *http.Request) {
		r.AddCookie(&http.Cookie{
			Name:  cookiejar.GetUserCookieName(),
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

func TestAuthenticationMiddlewareShouldAuthenticateViaBearerToken(t *testing.T) {
	// Arrange
	mockUser := newMockPrincipal("user456")
	rtr := router.NewRouter()
	UseAuthentication(rtr,
		WithValidator(func(token string) (claims.Principal, error) {
			if token == "valid-bearer-token" {
				return mockUser, nil
			}
			return nil, ErrorInvalidToken
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
