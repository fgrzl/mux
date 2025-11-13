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
// with defaults filled in. Note: MaxAge is returned as-is (may be 0)
// to allow callers to decide fallback behavior (e.g. provider TTL).
func ResolveCookieOptions(opts ...CookieOption) (maxAge int, path, domain string, secure, httpOnly bool, sameSite http.SameSite) {
	o := &cookieOptions{}
	for _, opt := range opts {
		opt(o)
	}

	maxAge = o.maxAge
	path = o.path
	if path == "" {
		path = "/"
	}
	domain = o.domain

	secure = true
	if o.secure != nil {
		secure = *o.secure
	}

	httpOnly = true
	if o.httpOnly != nil {
		httpOnly = *o.httpOnly
	}

	sameSite = http.SameSiteLaxMode
	if o.sameSite != nil {
		sameSite = *o.sameSite
	}

	return
}
