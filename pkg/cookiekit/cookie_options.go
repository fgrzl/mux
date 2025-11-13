package cookiekit

import "net/http"

// cookieOptions holds the internal cookie configuration.
type cookieOptions struct {
	maxAge   int
	path     string
	domain   string
	secure   *bool
	httpOnly *bool
	sameSite *http.SameSite
}

// CookieOption is a function that configures cookie options.
type CookieOption func(*cookieOptions)

// WithMaxAge sets the MaxAge (seconds).
// 0 means "unspecified" (falls back to provider TTL).
// Negative values indicate delete/expired cookie semantics.
func WithMaxAge(max int) CookieOption {
	return func(o *cookieOptions) {
		o.maxAge = max
	}
}

// WithPath sets the cookie Path.
func WithPath(p string) CookieOption {
	return func(o *cookieOptions) {
		o.path = p
	}
}

// WithDomain sets the cookie Domain.
func WithDomain(d string) CookieOption {
	return func(o *cookieOptions) {
		o.domain = d
	}
}

// WithSecure sets the Secure flag.
func WithSecure(b bool) CookieOption {
	return func(o *cookieOptions) {
		o.secure = &b
	}
}

// WithHttpOnly sets the HttpOnly flag.
func WithHttpOnly(b bool) CookieOption {
	return func(o *cookieOptions) {
		o.httpOnly = &b
	}
}

// WithSameSite sets the SameSite attribute.
func WithSameSite(s http.SameSite) CookieOption {
	return func(o *cookieOptions) {
		o.sameSite = &s
	}
}

// ResolveCookieOptions applies the option functions and returns resolved values
// with secure-by-default settings. Note: MaxAge is returned as-is (may be 0)
// to allow callers to decide fallback behavior (e.g. provider TTL).
//
// Smart Defaults:
//   - Path: "/" (most permissive, override with WithPath for API-specific cookies)
//   - Secure: true (prevents transmission over HTTP)
//   - HttpOnly: true (prevents JavaScript access, mitigates XSS)
//   - SameSite: Strict (strongest CSRF protection, prevents all cross-site cookie sending)
//
// Note: SameSite=Strict may break legitimate cross-site flows (OAuth redirects, external payment returns).
// Use WithSameSite(http.SameSiteLaxMode) for authentication cookies in sites with external integrations.
// Use WithSameSite(http.SameSiteNoneMode) only for cross-site APIs (requires Secure=true).
func ResolveCookieOptions(opts ...CookieOption) (maxAge int, path, domain string, secure, httpOnly bool, sameSite http.SameSite) {
	o := &cookieOptions{}
	for _, opt := range opts {
		opt(o)
	}

	// MaxAge: 0 means "use session/provider TTL"
	maxAge = o.maxAge

	// Path: Default to root, allowing cookie across entire site
	path = o.path
	if path == "" {
		path = "/"
	}

	// Domain: Empty means current domain only (no subdomains)
	domain = o.domain

	// Secure: Always true by default (HTTPS-only transmission)
	secure = true
	if o.secure != nil {
		secure = *o.secure
	}

	// HttpOnly: Always true by default (blocks JavaScript access)
	httpOnly = true
	if o.httpOnly != nil {
		httpOnly = *o.httpOnly
	}

	// SameSite: Strict by default (strongest CSRF protection)
	// Strict = no cross-site cookie sending (even on safe GET navigation)
	// Lax = allows cross-site GET from top-level navigation
	// None = allows all cross-site requests (requires Secure=true)
	sameSite = http.SameSiteStrictMode
	if o.sameSite != nil {
		sameSite = *o.sameSite
	}

	return
}
