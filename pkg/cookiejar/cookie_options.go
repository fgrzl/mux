package cookiejar

import "net/http"

// CookieOption allows callers to customize cookie attributes when creating
// authentication cookies via `Authenticate`. All fields are optional; nil
// pointer fields indicate "unspecified" so defaults are applied.
type CookieOption struct {
	// MaxAge in seconds. 0 means "unspecified" (caller should fall back
	// to provider TTL when appropriate); negative values indicate delete/expired cookie semantics.
	MaxAge int
	Path   string
	Domain string
	// Use pointers for booleans so callers can explicitly set false.
	Secure   *bool
	HttpOnly *bool
	SameSite *http.SameSite
}

// WithMaxAge sets the MaxAge (seconds).
func (o CookieOption) WithMaxAge(max int) CookieOption { o.MaxAge = max; return o }

// WithPath sets the cookie Path.
func (o CookieOption) WithPath(p string) CookieOption { o.Path = p; return o }

// WithDomain sets the cookie Domain.
func (o CookieOption) WithDomain(d string) CookieOption { o.Domain = d; return o }

// WithSecure sets the Secure flag.
func (o CookieOption) WithSecure(b bool) CookieOption { o.Secure = &b; return o }

// WithHttpOnly sets the HttpOnly flag.
func (o CookieOption) WithHttpOnly(b bool) CookieOption { o.HttpOnly = &b; return o }

// WithSameSite sets the SameSite attribute.
func (o CookieOption) WithSameSite(s http.SameSite) CookieOption { o.SameSite = &s; return o }

// Resolve fills in defaults for unspecified fields and returns concrete values
// suitable for passing to SetCookie. Note: MaxAge is returned as-is (may be 0)
// to allow callers to decide fallback behavior (e.g. provider TTL).
func (o CookieOption) Resolve() (maxAge int, path, domain string, secure, httpOnly bool, sameSite http.SameSite) {
	maxAge = o.MaxAge
	path = o.Path
	if path == "" {
		path = "/"
	}
	domain = o.Domain

	secure = true
	if o.Secure != nil {
		secure = *o.Secure
	}

	httpOnly = true
	if o.HttpOnly != nil {
		httpOnly = *o.HttpOnly
	}

	sameSite = http.SameSiteLaxMode
	if o.SameSite != nil {
		sameSite = *o.SameSite
	}

	return
}
