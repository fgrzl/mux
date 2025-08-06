package mux

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/fgrzl/claims"
)

// UseAuthentication adds authentication middleware to the router with the given options.
// This middleware enforces authentication and provides token creation capabilities.
func (r *Router) UseAuthentication(opts ...AuthOption) {
	options := &AuthenticationOptions{}

	for _, opt := range opts {
		opt(options)
	}

	provider := &defaultTokenProvider{
		ttl:        options.TokenTTL,
		signFn:     options.CreateToken,
		validateFn: options.Validate,
	}

	m := &authenticationMiddleware{
		provider: provider,
	}

	r.middleware = append(r.middleware, m)
}

// UseAuthenticationWithProvider adds authentication middleware using a custom token provider.
// This allows for more advanced token provider implementations.
func (r *Router) UseAuthenticationWithProvider(provider TokenProvider) {
	m := &authenticationMiddleware{
		provider: provider,
	}

	r.middleware = append(r.middleware, m)
}

// ---- Functional Options ----

// AuthOption represents a functional option for configuring authentication middleware.
type AuthOption func(*AuthenticationOptions)

// WithTokenTTL sets the token time-to-live duration.
func WithTokenTTL(ttl time.Duration) AuthOption {
	return func(o *AuthenticationOptions) {
		o.TokenTTL = ttl
	}
}

// WithValidator sets the token validation function.
func WithValidator(fn func(string) (claims.Principal, error)) AuthOption {
	return func(o *AuthenticationOptions) {
		o.Validate = fn
	}
}

// WithTokenCreator sets the token creation function.
func WithTokenCreator(fn func(claims.Principal, time.Duration) (string, error)) AuthOption {
	return func(o *AuthenticationOptions) {
		o.CreateToken = fn
	}
}

// ---- Internal Types ----

// AuthenticationOptions contains configuration options for authentication middleware.
type AuthenticationOptions struct {
	TokenTTL    time.Duration
	Validate    func(string) (claims.Principal, error)
	CreateToken func(claims.Principal, time.Duration) (string, error)
}

// defaultTokenProvider handles token creation and validation.
type defaultTokenProvider struct {
	ttl        time.Duration
	signFn     func(claims.Principal, time.Duration) (string, error)
	validateFn func(string) (claims.Principal, error)
}

// CreateToken creates a new token for the given user.
func (p *defaultTokenProvider) CreateToken(c *RouteContext, user claims.Principal) (string, error) {
	if p.signFn == nil {
		return "", errors.New("signing function is not set")
	}
	return p.signFn(user, p.ttl)
}

// ValidateToken validates the given token and returns the principal.
func (p *defaultTokenProvider) ValidateToken(ctx *RouteContext, token string) (claims.Principal, error) {
	if p.validateFn == nil {
		return nil, errors.New("validation function is not set")
	}
	return p.validateFn(token)
}

// GetTTL returns the token TTL for this provider.
func (p *defaultTokenProvider) GetTTL() time.Duration {
	return p.ttl
}

// CanCreateTokens returns true if this provider can create tokens.
func (p *defaultTokenProvider) CanCreateTokens() bool {
	return p.signFn != nil
}

// authenticationMiddleware implements authentication middleware functionality.
type authenticationMiddleware struct {
	provider TokenProvider
}

// ---- Middleware Logic ----

// Invoke implements the middleware interface for authentication.
func (m *authenticationMiddleware) Invoke(c *RouteContext, next HandlerFunc) {
	c.SetService("token.provider", m.provider)

	if c.Options.AllowAnonymous {
		slog.DebugContext(c, "authentication skipped: anonymous access allowed")
		next(c)
		return
	}

	// Try cookie authentication first
	if m.authenticateViaCookie(c) {
		next(c)
		return
	}

	// Try bearer token authentication
	if m.authenticateViaBearer(c) {
		next(c)
		return
	}

	// No valid authentication found
	slog.InfoContext(c, "authentication failed: no valid token found")
	c.Unauthorized()
}

// authenticateViaCookie attempts to authenticate using session cookie.
func (m *authenticationMiddleware) authenticateViaCookie(c *RouteContext) bool {
	cookie, err := c.Request.Cookie(GetUserCookieName())
	if err != nil {
		return false
	}

	principal, err := m.provider.ValidateToken(c, cookie.Value)
	if err != nil {
		slog.WarnContext(c, "invalid session cookie", "error", err)
		return false
	}

	m.setAuthenticatedUser(c, principal, "cookie")
	m.extendSessionExpiration(c, cookie)
	return true
}

// authenticateViaBearer attempts to authenticate using bearer token.
func (m *authenticationMiddleware) authenticateViaBearer(c *RouteContext) bool {
	token := extractBearerToken(c.Request)
	if token == "" {
		return false
	}

	principal, err := m.provider.ValidateToken(c, token)
	if err != nil {
		slog.WarnContext(c, "invalid bearer token", "error", err)
		return false
	}

	m.setAuthenticatedUser(c, principal, "bearer")
	return true
}

// setAuthenticatedUser sets the authenticated user and logs the success.
func (m *authenticationMiddleware) setAuthenticatedUser(c *RouteContext, principal claims.Principal, method string) {
	userID := principal.Subject()
	if userID == "" {
		userID = "unknown"
	}

	c.User = principal
	slog.DebugContext(c, "authentication success", "method", method, "user", userID)
}

// extendSessionExpiration extends the session expiration time and renews the token if possible.
func (m *authenticationMiddleware) extendSessionExpiration(c *RouteContext, cookie *http.Cookie) {
	ttl := m.provider.GetTTL()
	if ttl <= 0 {
		slog.DebugContext(c, "session extension skipped: TTL not set")
		return
	}

	// Renew token if possible
	if m.provider.CanCreateTokens() {
		if token, err := m.provider.CreateToken(c, c.User); err == nil {
			cookie.Value = token
			slog.DebugContext(c, "session token renewed", "user", c.User.Subject())
		} else {
			slog.WarnContext(c, "failed to renew token", "error", err)
		}
	}

	// Update cookie properties
	m.updateCookieProperties(cookie, ttl, c.Request.TLS != nil)

	http.SetCookie(c.Response, cookie)
	slog.DebugContext(c, "session cookie extended", "expires", cookie.Expires)
}

// updateCookieProperties updates cookie security properties.
func (m *authenticationMiddleware) updateCookieProperties(cookie *http.Cookie, ttl time.Duration, isSecure bool) {
	cookie.Expires = time.Now().Add(ttl)
	cookie.MaxAge = int(ttl.Seconds())
	cookie.HttpOnly = true
	cookie.SameSite = http.SameSiteLaxMode
	cookie.Secure = isSecure
	cookie.Path = "/"
}

// extractBearerToken extracts the bearer token from the Authorization header.
func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}
