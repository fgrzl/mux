package authentication

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/fgrzl/claims"
	"github.com/fgrzl/mux/internal/common"
	"github.com/fgrzl/mux/internal/cookiekit"
	"github.com/fgrzl/mux/internal/router"
	"github.com/fgrzl/mux/internal/routing"
	"github.com/fgrzl/mux/internal/tokenizer"
)

// UseAuthentication adds authentication middleware to the router with the given options.
// This middleware enforces authentication and provides token creation capabilities.
func UseAuthentication(rtr *router.Router, opts ...AuthOption) {
	options := &AuthenticationOptions{}

	for _, opt := range opts {
		opt(options)
	}
	options.CookieNames = cookiekit.NormalizeCookieNames(options.CookieNames)

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
	options.CookieNames = cookiekit.NormalizeCookieNames(options.CookieNames)
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

func WithAppSessionCookieName(name string) AuthOption {
	return func(o *AuthenticationOptions) {
		o.CookieNames.AppSession = name
	}
}

func WithTwoFactorCookieName(name string) AuthOption {
	return func(o *AuthenticationOptions) {
		o.CookieNames.TwoFactor = name
	}
}

func WithIDPSessionCookieName(name string) AuthOption {
	return func(o *AuthenticationOptions) {
		o.CookieNames.IDPSession = name
	}
}

// WithCookieOptions sets default cookie attributes for auth cookies issued,
// renewed, and cleared through the authentication middleware.
func WithCookieOptions(opts ...cookiekit.CookieOption) AuthOption {
	return func(o *AuthenticationOptions) {
		o.CookieOptions = append(o.CookieOptions, cookiekit.CloneCookieOptions(opts)...)
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

// WithRateLimiter sets a custom rate limiter for failed authentication attempts.
// The function should return true if another failed attempt should be allowed,
// false if the client should be rate limited. The string parameter is the client
// identifier (typically IP address).
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
	CookieNames      cookiekit.CookieNames
	CookieOptions    []cookiekit.CookieOption
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
	if options != nil {
		options.CookieNames = cookiekit.NormalizeCookieNames(options.CookieNames)
		options.CookieOptions = cookiekit.CloneCookieOptions(options.CookieOptions)
	}
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

var csrfTokenEntropySource io.Reader = rand.Reader

// ---- Middleware Logic ----

// Invoke implements the middleware interface for authentication.
func (m *authenticationMiddleware) Invoke(c routing.RouteContext, next router.HandlerFunc) {
	c.SetService(tokenizer.ServiceKeyTokenProvider, m.provider)
	c.SetService(cookiekit.ServiceKeyCookieNames, m.cookieNames())
	if cookieOptions := m.authCookieOptions(); len(cookieOptions) > 0 {
		c.SetService(cookiekit.ServiceKeyAuthCookieOptions, cookieOptions)
	}

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
	if m.rateLimitFailure(c) {
		return
	}
	slog.DebugContext(c, "authentication failed: no valid token found")
	c.Unauthorized()
}

func (m *authenticationMiddleware) rateLimitFailure(c routing.RouteContext) bool {
	if m.options == nil || m.options.RateLimiter == nil {
		return false
	}

	clientID := getClientID(c.Request())
	if m.options.RateLimiter(clientID) {
		return false
	}

	slog.WarnContext(c, "authentication rate limited", "client", clientID)
	c.JSON(http.StatusTooManyRequests, map[string]string{
		"error": "too many authentication failures, please try again later",
	})
	return true
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
	cookieNames := m.cookieNames()
	cookie, err := req.Cookie(cookieNames.AppSession)
	if err != nil {
		return false
	}

	if !m.validateAndSetUser(c, cookie.Value, "cookie") {
		return false
	}

	m.extendSessionExpiration(c, cookie)
	return true
}

func (m *authenticationMiddleware) cookieNames() cookiekit.CookieNames {
	if m == nil || m.options == nil {
		return cookiekit.DefaultCookieNames()
	}
	return cookiekit.NormalizeCookieNames(m.options.CookieNames)
}

func (m *authenticationMiddleware) authCookieOptions() []cookiekit.CookieOption {
	if m == nil || m.options == nil {
		return nil
	}
	return cookiekit.CloneCookieOptions(m.options.CookieOptions)
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

	maxAge, path, domain, secure, httpOnly, sameSite := m.resolveSessionCookieSettings(ttl)
	c.SetCookie(cookie.Name, cookie.Value, maxAge, path, domain, secure, httpOnly, sameSite)

	expires := time.Time{}
	if maxAge > 0 {
		expires = time.Now().Add(time.Duration(maxAge) * time.Second)
	}
	slog.DebugContext(c, "session cookie extended", "expires", expires)
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

// resolveSessionCookieSettings returns the configured cookie attributes for auth
// cookies, defaulting Max-Age to the provider TTL when not explicitly set.
func (m *authenticationMiddleware) resolveSessionCookieSettings(ttl time.Duration) (maxAge int, path, domain string, secure, httpOnly bool, sameSite http.SameSite) {
	maxAge, path, domain, secure, httpOnly, sameSite = cookiekit.ResolveCookieOptions(m.authCookieOptions()...)
	if maxAge == 0 {
		maxAge = int(ttl.Seconds())
	}
	return
}

// isSecureRequest determines if the request was made over HTTPS.
// Trusted proxy middleware should normalize request state before authentication
// runs so raw client headers are not treated as authoritative here.
func isSecureRequest(r *http.Request) bool {
	if r == nil {
		return false
	}
	// Direct TLS connection
	if r.TLS != nil {
		return true
	}
	return strings.EqualFold(r.URL.Scheme, "https")
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
	token, err := GenerateCSRFTokenErr(c)
	if err != nil {
		logCtx := context.Background()
		if c != nil && c.Request() != nil && c.Request().Context() != nil {
			logCtx = c.Request().Context()
		}
		slog.ErrorContext(logCtx, "failed to generate CSRF token", "error", err)
		return ""
	}
	return token
}

// GenerateCSRFTokenErr generates a new CSRF token and sets it as a cookie.
// It returns an error instead of falling back to weak entropy if the token
// cannot be generated securely.
func GenerateCSRFTokenErr(c routing.RouteContext) (string, error) {
	token, err := generateSecureToken(csrfTokenLength)
	if err != nil {
		return "", err
	}
	if c == nil || c.Request() == nil || c.Response() == nil {
		return "", errors.New("route context is not initialized")
	}

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

	return token, nil
}

// generateSecureToken generates a cryptographically secure random token.
func generateSecureToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := io.ReadFull(csrfTokenEntropySource, b); err != nil {
		return "", fmt.Errorf("generate CSRF token: %w", err)
	}
	return strings.TrimRight(base64.RawURLEncoding.EncodeToString(b), "=")[:length], nil
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
// It relies on the normalized RemoteAddr so only trusted middleware can affect
// the effective client identity.
func getClientID(r *http.Request) string {
	if r == nil {
		return ""
	}
	return normalizeClientID(r.RemoteAddr)
}

func normalizeClientID(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	if host, _, err := net.SplitHostPort(value); err == nil {
		value = host
	}

	return strings.Trim(value, "[]")
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
