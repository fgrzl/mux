package mux

import (
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fgrzl/claims"
	"github.com/stretchr/testify/assert"
)

// mockTokenProvider implements TokenProvider for testing
type mockTokenProvider struct {
	createTokenFunc   func(ctx *RouteContext, principal claims.Principal) (string, error)
	validateTokenFunc func(ctx *RouteContext, token string) (claims.Principal, error)
	ttl               time.Duration
	canCreateTokens   bool
}

func (m *mockTokenProvider) CreateToken(ctx *RouteContext, principal claims.Principal) (string, error) {
	if m.createTokenFunc != nil {
		return m.createTokenFunc(ctx, principal)
	}
	return "", errors.New("not implemented")
}

func (m *mockTokenProvider) ValidateToken(ctx *RouteContext, token string) (claims.Principal, error) {
	if m.validateTokenFunc != nil {
		return m.validateTokenFunc(ctx, token)
	}
	return nil, errors.New("not implemented")
}

func (m *mockTokenProvider) GetTTL() time.Duration {
	return m.ttl
}

func (m *mockTokenProvider) CanCreateTokens() bool {
	return m.canCreateTokens
}

// mockPrincipalForToken implements claims.Principal for testing
type mockPrincipalForToken struct {
	subject string
}

func (m *mockPrincipalForToken) Subject() string                      { return m.subject }
func (m *mockPrincipalForToken) Issuer() string                       { return "test-issuer" }
func (m *mockPrincipalForToken) Audience() []string                   { return []string{"test"} }
func (m *mockPrincipalForToken) ExpirationTime() int64                { return 0 }
func (m *mockPrincipalForToken) NotBefore() int64                     { return 0 }
func (m *mockPrincipalForToken) IssuedAt() int64                      { return 0 }
func (m *mockPrincipalForToken) JWTI() string                         { return "test-jwt" }
func (m *mockPrincipalForToken) Scopes() []string                     { return []string{} }
func (m *mockPrincipalForToken) Roles() []string                      { return []string{} }
func (m *mockPrincipalForToken) Email() string                        { return "test@example.com" }
func (m *mockPrincipalForToken) Username() string                     { return "testuser" }
func (m *mockPrincipalForToken) CustomClaim(name string) claims.Claim { return nil }
func (m *mockPrincipalForToken) CustomClaimValue(name string) string  { return "" }
func (m *mockPrincipalForToken) Claims() *claims.ClaimSet             { return nil }

func TestTokenProviderInterfaceShouldBeImplementable(t *testing.T) {
	// Arrange
	provider := &mockTokenProvider{
		ttl:             time.Hour,
		canCreateTokens: true,
	}

	// Act & Assert
	var tokenProvider TokenProvider = provider
	assert.NotNil(t, tokenProvider)
	assert.Equal(t, time.Hour, tokenProvider.GetTTL())
	assert.True(t, tokenProvider.CanCreateTokens())
}

func TestTokenProviderShouldCreateToken(t *testing.T) {
	// Arrange
	expectedToken := "test-token-123"
	principal := &mockPrincipalForToken{subject: "test-user"}
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)

	provider := &mockTokenProvider{
		createTokenFunc: func(ctx *RouteContext, principal claims.Principal) (string, error) {
			return expectedToken, nil
		},
		canCreateTokens: true,
	}

	// Act
	token, err := provider.CreateToken(ctx, principal)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedToken, token)
}

func TestTokenProviderShouldReturnErrorWhenCreateTokenFails(t *testing.T) {
	// Arrange
	expectedError := errors.New("token creation failed")
	principal := &mockPrincipalForToken{subject: "test-user"}
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)

	provider := &mockTokenProvider{
		createTokenFunc: func(ctx *RouteContext, principal claims.Principal) (string, error) {
			return "", expectedError
		},
		canCreateTokens: true,
	}

	// Act
	token, err := provider.CreateToken(ctx, principal)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, token)
}

func TestTokenProviderShouldValidateToken(t *testing.T) {
	// Arrange
	token := "valid-token-123"
	expectedPrincipal := &mockPrincipalForToken{subject: "test-user"}
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)

	provider := &mockTokenProvider{
		validateTokenFunc: func(ctx *RouteContext, token string) (claims.Principal, error) {
			if token == "valid-token-123" {
				return expectedPrincipal, nil
			}
			return nil, errors.New("invalid token")
		},
	}

	// Act
	principal, err := provider.ValidateToken(ctx, token)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedPrincipal, principal)
}

func TestTokenProviderShouldReturnErrorForInvalidToken(t *testing.T) {
	// Arrange
	token := "invalid-token"
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)

	provider := &mockTokenProvider{
		validateTokenFunc: func(ctx *RouteContext, token string) (claims.Principal, error) {
			return nil, errors.New("invalid token")
		},
	}

	// Act
	principal, err := provider.ValidateToken(ctx, token)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, principal)
	assert.Contains(t, err.Error(), "invalid token")
}

func TestTokenProviderShouldReturnTTL(t *testing.T) {
	// Arrange
	expectedTTL := 2 * time.Hour
	provider := &mockTokenProvider{
		ttl: expectedTTL,
	}

	// Act
	ttl := provider.GetTTL()

	// Assert
	assert.Equal(t, expectedTTL, ttl)
}

func TestTokenProviderShouldIndicateTokenCreationCapability(t *testing.T) {
	tests := []struct {
		name            string
		canCreateTokens bool
		expected        bool
	}{
		{
			name:            "can create tokens",
			canCreateTokens: true,
			expected:        true,
		},
		{
			name:            "cannot create tokens",
			canCreateTokens: false,
			expected:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			provider := &mockTokenProvider{
				canCreateTokens: tt.canCreateTokens,
			}

			// Act
			canCreate := provider.CanCreateTokens()

			// Assert
			assert.Equal(t, tt.expected, canCreate)
		})
	}
}

func TestTokenProviderShouldHandleNilPrincipal(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)

	provider := &mockTokenProvider{
		createTokenFunc: func(ctx *RouteContext, principal claims.Principal) (string, error) {
			if principal == nil {
				return "", errors.New("principal cannot be nil")
			}
			return "token", nil
		},
		canCreateTokens: true,
	}

	// Act
	token, err := provider.CreateToken(ctx, nil)

	// Assert
	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "principal cannot be nil")
}

func TestTokenProviderShouldHandleEmptyToken(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)

	provider := &mockTokenProvider{
		validateTokenFunc: func(ctx *RouteContext, token string) (claims.Principal, error) {
			if token == "" {
				return nil, errors.New("token cannot be empty")
			}
			return &mockPrincipalForToken{}, nil
		},
	}

	// Act
	principal, err := provider.ValidateToken(ctx, "")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, principal)
	assert.Contains(t, err.Error(), "token cannot be empty")
}

func TestTokenProviderShouldBeUsableInRouteContext(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)

	provider := &mockTokenProvider{
		ttl:             time.Minute * 30,
		canCreateTokens: true,
	}

	// Act
	ctx.SetService(ServiceKeyTokenProvider, provider)
	retrieved, exists := ctx.GetService(ServiceKeyTokenProvider)

	// Assert
	assert.True(t, exists)
	assert.Equal(t, provider, retrieved)

	// Verify it's still usable as TokenProvider interface
	if tp, ok := retrieved.(TokenProvider); ok {
		assert.Equal(t, time.Minute*30, tp.GetTTL())
		assert.True(t, tp.CanCreateTokens())
	} else {
		t.Fatal("Retrieved service is not a TokenProvider")
	}
}
