package authentication

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/fgrzl/claims"
	"github.com/fgrzl/mux/pkg/common"
	"github.com/fgrzl/mux/pkg/cookiejar"
	"github.com/fgrzl/mux/pkg/router"
	"github.com/fgrzl/mux/pkg/routing"
	"github.com/fgrzl/mux/pkg/tokenizer"
)

// UseAuthentication adds authentication middleware to the router with the given options.
// This middleware enforces authentication and provides token creation capabilities.
func UseAuthentication(rtr *router.Router, opts ...AuthOption) {
	options := &AuthenticationOptions{}

	for _, opt := range opts {
		opt(options)
	}

	provider := &defaultTokenProvider{
		ttl:        options.TokenTTL,
		signFn:     options.CreateToken,
		validateFn: options.Validate,
	}

	rtr.Use(newAuthMiddleware(provider))
}

// UseAuthenticationWithProvider adds authentication middleware using a custom token provider.
// This allows for more advanced token provider implementations.
func UseAuthenticationWithProvider(rtr *router.Router, provider tokenizer.TokenProvider) {
	rtr.Use(newAuthMiddleware(provider))
}

// newAuthMiddleware creates a new authentication middleware instance with the
// provided token provider. Extracted to reduce duplication between constructors.
func newAuthMiddleware(provider tokenizer.TokenProvider) *authenticationMiddleware {
	return &authenticationMiddleware{provider: provider}
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
func (p *defaultTokenProvider) CreateToken(ctx context.Context, user claims.Principal) (string, error) {
	if p.signFn == nil {
		return "", errors.New("signing function is not set")
	}
	return p.signFn(user, p.ttl)
}

// ValidateToken validates the given token and returns the principal.
func (p *defaultTokenProvider) ValidateToken(ctx context.Context, token string) (claims.Principal, error) {
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
	provider tokenizer.TokenProvider
}

// ErrInvalidToken is returned when a provided token is invalid.
var ErrInvalidToken = errors.New("invalid token")

// ---- Middleware Logic ----

// Invoke implements the middleware interface for authentication.
func (m *authenticationMiddleware) Invoke(c routing.RouteContext, next router.HandlerFunc) {
	c.SetService(tokenizer.ServiceKeyTokenProvider, m.provider)

	opts := c.Options()
	// Options may be nil for some contexts (pooled or partially-initialized).
	// Treat nil as zero-value options (no AllowAnonymous).
	if opts != nil && opts.AllowAnonymous {
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
func (m *authenticationMiddleware) authenticateViaCookie(c routing.RouteContext) bool {
	req := c.Request()
	cookie, err := req.Cookie(cookiejar.GetUserCookieName())
	if err != nil {
		return false
	}

	if !m.validateAndSetUser(c, cookie.Value, "cookie") {
		return false
	}

	m.extendSessionExpiration(c, cookie)
	return true
}

// authenticateViaBearer attempts to authenticate using bearer token.
func (m *authenticationMiddleware) authenticateViaBearer(c routing.RouteContext) bool {
	req := c.Request()
	token := extractBearerToken(req.Header.Get(common.HeaderAuthorization))
	if token == "" {
		return false
	}
	return m.validateAndSetUser(c, token, "bearer")
}

// validateAndSetUser validates the token using the provider and, if valid, sets
// the authenticated user on the context and logs the success. Returns true on
// success, false otherwise.
func (m *authenticationMiddleware) validateAndSetUser(c routing.RouteContext, token string, method string) bool {
	p := m.provider
	principal, err := p.ValidateToken(c, token)
	if err != nil {
		slog.WarnContext(c, "invalid token", "method", method, "error", err)
		return false
	}

	m.setAuthenticatedUser(c, principal, method)
	return true
}

// setAuthenticatedUser sets the authenticated user and logs the success.
func (m *authenticationMiddleware) setAuthenticatedUser(c routing.RouteContext, principal claims.Principal, method string) {
	userID := principal.Subject()
	if userID == "" {
		userID = "unknown"
	}

	c.SetUser(principal)
	slog.DebugContext(c, "authentication success", "method", method, "user", userID)
}

// extendSessionExpiration extends the session expiration time and renews the token if possible.
func (m *authenticationMiddleware) extendSessionExpiration(c routing.RouteContext, cookie *http.Cookie) {
	p := m.provider
	ttl := p.GetTTL()
	if ttl <= 0 {
		slog.DebugContext(c, "session extension skipped: TTL not set")
		return
	}
	// Renew token if possible
	m.renewTokenIfPossible(c, cookie)

	// Update cookie properties
	isSecure := c.Request().TLS != nil
	m.updateCookieProperties(cookie, ttl, isSecure)

	http.SetCookie(c.Response(), cookie)
	slog.DebugContext(c, "session cookie extended", "expires", cookie.Expires)
}

// renewTokenIfPossible attempts to create a new token for the current user and
// updates the provided cookie.Value when successful. It logs success and
// failures appropriately. No error is returned since renewal is best-effort.
func (m *authenticationMiddleware) renewTokenIfPossible(c routing.RouteContext, cookie *http.Cookie) {
	p := m.provider
	if !p.CanCreateTokens() {
		return
	}

	user := c.User()
	token, err := p.CreateToken(c, user)
	if err != nil {
		slog.WarnContext(c, "failed to renew token", "error", err)
		return
	}

	cookie.Value = token
	if user != nil {
		slog.DebugContext(c, "session token renewed", "user", user.Subject())
	} else {
		slog.DebugContext(c, "session token renewed")
	}
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

// extractBearerToken extracts the bearer token from the Authorization header value.
func extractBearerToken(authHeader string) string {
	if authHeader == "" {
		return ""
	}
	const prefix = "Bearer "
	if len(authHeader) <= len(prefix) {
		return ""
	}
	// Direct slice comparison is slightly faster than HasPrefix.
	if authHeader[:len(prefix)] == prefix {
		return authHeader[len(prefix):]
	}
	return ""
}
