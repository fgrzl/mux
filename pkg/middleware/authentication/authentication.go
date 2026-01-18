package authentication

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/fgrzl/claims"
	"github.com/fgrzl/mux/pkg/common"
	"github.com/fgrzl/mux/pkg/cookiekit"
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

	rtr.Use(newAuthMiddlewareWithOptions(provider, options))
}

// UseAuthenticationWithProvider adds authentication middleware using a custom token provider.
// This allows for more advanced token provider implementations.
func UseAuthenticationWithProvider(rtr *router.Router, provider tokenizer.TokenProvider, opts ...AuthOption) {
	options := &AuthenticationOptions{}
	for _, opt := range opts {
		opt(options)
	}
	rtr.Use(newAuthMiddlewareWithOptions(provider, options))
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

// WithCSRFProtection enables CSRF protection for cookie-based authentication.
// When enabled, state-changing requests (POST, PUT, DELETE, PATCH) must include
// a valid CSRF token in the X-CSRF-Token header that matches the csrf_token cookie.
func WithCSRFProtection() AuthOption {
	return func(o *AuthenticationOptions) {
		o.EnableCSRF = true
	}
}

// WithRateLimiter sets a custom rate limiter for authentication failures.
// The function should return true if the request should be allowed, false if rate limited.
// The string parameter is the client identifier (typically IP address).
func WithRateLimiter(fn func(clientID string) bool) AuthOption {
	return func(o *AuthenticationOptions) {
		o.RateLimiter = fn
	}
}

// WithTokenRevocationChecker sets a function to check if a token has been revoked.
// The function should return true if the token is revoked/blocklisted.
func WithTokenRevocationChecker(fn func(token string) bool) AuthOption {
	return func(o *AuthenticationOptions) {
		o.IsTokenRevoked = fn
	}
}

// WithIssuerValidator sets the expected token issuer for validation.
// If set, tokens with a different issuer will be rejected.
func WithIssuerValidator(issuer string) AuthOption {
	return func(o *AuthenticationOptions) {
		o.ExpectedIssuer = issuer
	}
}

// WithAudienceValidator sets the expected token audience for validation.
// If set, tokens that don't include this audience will be rejected.
func WithAudienceValidator(audience string) AuthOption {
	return func(o *AuthenticationOptions) {
		o.ExpectedAudience = audience
	}
}

// WithContextEnricher sets a function to enrich the request context after successful authentication.
// The enricher receives the full RouteContext (with the authenticated user already set) and can
// add values to the context using c.SetContextValue(key, value). This is useful for extracting
// claims from the principal and setting domain-specific context values (tenant info, permissions, etc.).
//
// Example:
//
//	WithContextEnricher(func(c routing.RouteContext) {
//	    principal := c.User()
//	    c.SetContextValue(tenantKey, principal.CustomClaimValue("tenant_id"))
//	    c.SetContextValue(userKey, principal.Subject())
//	})
//
// Values set via SetContextValue are accessible via both c.Value(key) and c.Request().Context().Value(key).
func WithContextEnricher(fn func(c routing.RouteContext)) AuthOption {
	return func(o *AuthenticationOptions) {
		o.ContextEnricher = fn
	}
}

// ---- Internal Types ----

// AuthenticationOptions contains configuration options for authentication middleware.
type AuthenticationOptions struct {
	TokenTTL         time.Duration
	Validate         func(string) (claims.Principal, error)
	CreateToken      func(claims.Principal, time.Duration) (string, error)
	EnableCSRF       bool
	RateLimiter      func(clientID string) bool
	IsTokenRevoked   func(token string) bool
	ExpectedIssuer   string
	ExpectedAudience string
	ContextEnricher  func(c routing.RouteContext)
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
	options  *AuthenticationOptions
}

// newAuthMiddlewareWithOptions creates a new authentication middleware with options.
func newAuthMiddlewareWithOptions(provider tokenizer.TokenProvider, options *AuthenticationOptions) *authenticationMiddleware {
	return &authenticationMiddleware{provider: provider, options: options}
}

// ErrInvalidToken is returned when a provided token is invalid.
var ErrInvalidToken = errors.New("invalid token")

// ErrTokenRevoked is returned when a token has been revoked/blocklisted.
var ErrTokenRevoked = errors.New("token has been revoked")

// ErrCSRFValidationFailed is returned when CSRF token validation fails.
var ErrCSRFValidationFailed = errors.New("CSRF validation failed")

// ErrRateLimited is returned when the client has been rate limited.
var ErrRateLimited = errors.New("too many authentication failures")

// CSRF constants
const (
	csrfTokenCookieName = "csrf_token"
	csrfTokenHeaderName = "X-CSRF-Token"
	csrfTokenLength     = 32
)

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

	// Check rate limiting before processing authentication
	if m.options != nil && m.options.RateLimiter != nil {
		clientID := getClientID(c.Request())
		if !m.options.RateLimiter(clientID) {
			slog.WarnContext(c, "authentication rate limited", "client", clientID)
			c.JSON(http.StatusTooManyRequests, map[string]string{
				"error": "too many authentication failures, please try again later",
			})
			return
		}
	}

	// Try cookie authentication first
	if m.authenticateViaCookie(c) {
		// CSRF check for cookie-based auth on state-changing requests
		if m.options != nil && m.options.EnableCSRF && isStateChangingMethod(c.Request().Method) {
			if !m.validateCSRFToken(c) {
				slog.WarnContext(c, "CSRF validation failed")
				c.JSON(http.StatusForbidden, map[string]string{
					"error": "CSRF token validation failed",
				})
				return
			}
		}
		m.enrichContext(c)
		next(c)
		return
	}

	// Try bearer token authentication
	if m.authenticateViaBearer(c) {
		m.enrichContext(c)
		next(c)
		return
	}

	// No valid authentication found
	slog.InfoContext(c, "authentication failed: no valid token found")
	c.Unauthorized()
}

// enrichContext calls the context enricher if configured, updating
// the underlying request with the enriched context.
func (m *authenticationMiddleware) enrichContext(c routing.RouteContext) {
	if m.options == nil || m.options.ContextEnricher == nil {
		return
	}

	// The enricher calls c.SetContextValue() to add values.
	// SetContextValue handles updating both the embedded context and the request properly.
	m.options.ContextEnricher(c)
}

// authenticateViaCookie attempts to authenticate using session cookie.
func (m *authenticationMiddleware) authenticateViaCookie(c routing.RouteContext) bool {
	req := c.Request()
	cookie, err := req.Cookie(cookiekit.GetUserCookieName())
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
	// Check if token is revoked/blocklisted
	if m.options != nil && m.options.IsTokenRevoked != nil {
		if m.options.IsTokenRevoked(token) {
			slog.WarnContext(c, "token revoked", "method", method)
			return false
		}
	}

	p := m.provider
	principal, err := p.ValidateToken(c, token)
	if err != nil {
		slog.WarnContext(c, "invalid token", "method", method, "error", err)
		return false
	}

	// Validate issuer if configured
	if m.options != nil && m.options.ExpectedIssuer != "" {
		if principal.Issuer() != m.options.ExpectedIssuer {
			slog.WarnContext(c, "invalid issuer", "method", method, "expected", m.options.ExpectedIssuer, "got", principal.Issuer())
			return false
		}
	}

	// Validate audience if configured
	if m.options != nil && m.options.ExpectedAudience != "" {
		if !containsAudience(principal.Audience(), m.options.ExpectedAudience) {
			slog.WarnContext(c, "invalid audience", "method", method, "expected", m.options.ExpectedAudience, "got", principal.Audience())
			return false
		}
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
	isSecure := isSecureRequest(c.Request())
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

// isSecureRequest determines if the request was made over HTTPS.
// It checks both direct TLS connections and proxy headers (X-Forwarded-Proto).
func isSecureRequest(r *http.Request) bool {
	// Direct TLS connection
	if r.TLS != nil {
		return true
	}
	// Check X-Forwarded-Proto header (set by reverse proxies)
	proto := r.Header.Get("X-Forwarded-Proto")
	return strings.EqualFold(proto, "https")
}

// extractBearerToken extracts the bearer token from the Authorization header value.
// Per RFC 7235, the scheme comparison is case-insensitive.
func extractBearerToken(authHeader string) string {
	if authHeader == "" {
		return ""
	}
	const prefix = "Bearer "
	if len(authHeader) <= len(prefix) {
		return ""
	}
	// Case-insensitive comparison per RFC 7235 section 2.1
	if strings.EqualFold(authHeader[:len(prefix)-1], prefix[:len(prefix)-1]) && authHeader[len(prefix)-1] == ' ' {
		return authHeader[len(prefix):]
	}
	return ""
}

// ---- CSRF Protection ----

// validateCSRFToken validates the CSRF token from the request header against the cookie.
func (m *authenticationMiddleware) validateCSRFToken(c routing.RouteContext) bool {
	req := c.Request()

	// Get CSRF token from cookie
	csrfCookie, err := req.Cookie(csrfTokenCookieName)
	if err != nil || csrfCookie.Value == "" {
		return false
	}

	// Get CSRF token from header
	csrfHeader := req.Header.Get(csrfTokenHeaderName)
	if csrfHeader == "" {
		return false
	}

	// Constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare([]byte(csrfCookie.Value), []byte(csrfHeader)) == 1
}

// GenerateCSRFToken generates a new CSRF token and sets it as a cookie.
// Call this function when establishing a session to provide the client with a CSRF token.
func GenerateCSRFToken(c routing.RouteContext) string {
	token := generateSecureToken(csrfTokenLength)

	isSecure := isSecureRequest(c.Request())
	cookie := &http.Cookie{
		Name:     csrfTokenCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: false, // Must be readable by JavaScript
		Secure:   isSecure,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(c.Response(), cookie)

	return token
}

// generateSecureToken generates a cryptographically secure random token.
func generateSecureToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		// SECURITY WARNING: Falling back to weak token generation.
		// This should never happen in practice, but if it does, the generated
		// token will be predictable and should not be relied upon for security.
		slog.Error("crypto/rand failed, falling back to weak CSRF token generation",
			"error", err,
			"warning", "CSRF tokens may be predictable - investigate system entropy source")
		return base64.URLEncoding.EncodeToString([]byte(time.Now().String()))
	}
	return base64.URLEncoding.EncodeToString(b)[:length]
}

// isStateChangingMethod returns true for HTTP methods that can change state.
func isStateChangingMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch:
		return true
	default:
		return false
	}
}

// ---- Rate Limiting Helpers ----

// getClientID extracts a client identifier from the request for rate limiting.
// It checks X-Forwarded-For first (for proxied requests), then falls back to RemoteAddr.
func getClientID(r *http.Request) string {
	// Check X-Forwarded-For header first (common in proxied setups)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the list (original client)
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header (nginx convention)
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr, stripping port if present
	addr := r.RemoteAddr
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return addr[:idx]
	}
	return addr
}

// NewInMemoryRateLimiter creates a simple in-memory rate limiter.
// maxAttempts is the maximum number of failed attempts allowed within the window.
// window is the time window for counting attempts.
// This is suitable for single-instance deployments; use Redis-based limiting for distributed systems.
func NewInMemoryRateLimiter(maxAttempts int, window time.Duration) func(clientID string) bool {
	var mu sync.Mutex
	attempts := make(map[string]*rateLimitEntry)

	// Start a cleanup goroutine
	go func() {
		ticker := time.NewTicker(window)
		defer ticker.Stop()
		for range ticker.C {
			mu.Lock()
			now := time.Now()
			for k, v := range attempts {
				if now.Sub(v.firstAttempt) > window {
					delete(attempts, k)
				}
			}
			mu.Unlock()
		}
	}()

	return func(clientID string) bool {
		mu.Lock()
		defer mu.Unlock()

		now := time.Now()
		entry, exists := attempts[clientID]

		if !exists || now.Sub(entry.firstAttempt) > window {
			// Start fresh window
			attempts[clientID] = &rateLimitEntry{
				count:        1,
				firstAttempt: now,
			}
			return true
		}

		entry.count++
		return entry.count <= maxAttempts
	}
}

type rateLimitEntry struct {
	count        int
	firstAttempt time.Time
}

// ---- Audience Validation ----

// containsAudience checks if the expected audience is present in the token's audience list.
func containsAudience(audiences []string, expected string) bool {
	for _, aud := range audiences {
		if aud == expected {
			return true
		}
	}
	return false
}
