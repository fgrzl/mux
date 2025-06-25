package mux

import (
	"net/http"
	"strings"
	"time"

	"log/slog"

	"github.com/fgrzl/claims"
)

func (r *Router) UseAuthentication(opts ...AuthOption) {
	options := &AuthenticationOptions{}
	for _, opt := range opts {
		opt(options)
	}
	r.middleware = append(r.middleware, &authenticationMiddleware{options: options})
}

// ---- Functional Options ----

type AuthOption func(*AuthenticationOptions)

func WithTokenTTL(ttl time.Duration) AuthOption {
	return func(o *AuthenticationOptions) {
		o.TokenTTL = ttl
	}
}

func WithValidator(fn func(string) (claims.Principal, error)) AuthOption {
	return func(o *AuthenticationOptions) {
		o.Validate = fn
	}
}

func WithTokenCreator(fn func(claims.Principal, time.Duration) (string, error)) AuthOption {
	return func(o *AuthenticationOptions) {
		o.CreateToken = fn
	}
}

// ---- Internal Types ----

type AuthenticationOptions struct {
	TokenTTL    time.Duration
	Validate    func(string) (claims.Principal, error)
	CreateToken func(claims.Principal, time.Duration) (string, error)
}

type authenticationMiddleware struct {
	options *AuthenticationOptions
}

type SessionDetails struct {
	IP        string
	UserAgent string
	CreatedAt time.Time
}

// ---- Middleware Logic ----

func (m *authenticationMiddleware) Invoke(c *RouteContext, next HandlerFunc) {
	if c.Options.AllowAnonymous {
		slog.DebugContext(c, "authentication skipped: anonymous access allowed")
		next(c)
		return
	}

	// 1. Check for cookie
	if cookie, err := c.Request.Cookie(GetUserCookieName()); err == nil {
		if principal, err := m.options.Validate(cookie.Value); err == nil {
			userID := principal.Subject()
			if userID == "" {
				userID = "unknown"
			}
			slog.DebugContext(c, "authentication success via cookie", "user", userID)
			c.User = principal
			next(c)
			m.extendSessionExpiration(c, cookie)
			return
		} else {
			slog.WarnContext(c, "invalid session cookie", "error", err)
		}
	}

	// 2. Check bearer token
	if token := extractBearerToken(c.Request); token != "" {
		if principal, err := m.options.Validate(token); err == nil {
			userID := principal.Subject()
			if userID == "" {
				userID = "unknown"
			}
			slog.DebugContext(c, "authentication success via bearer", "user", userID)
			c.User = principal
			next(c)
			return
		} else {
			slog.WarnContext(c, "invalid bearer token", "error", err)
		}
	}

	// 3. Unauthorized
	slog.InfoContext(c, "authentication failed: no valid token found")
	c.Unauthorized()
}

func (m *authenticationMiddleware) extendSessionExpiration(c *RouteContext, cookie *http.Cookie) {
	ttl := m.options.TokenTTL
	if ttl <= 0 {
		slog.DebugContext(c, "session extension skipped: TTL not set")
		return
	}

	if m.options.CreateToken != nil {
		if token, err := m.options.CreateToken(c.User, ttl); err == nil {
			cookie.Value = token
			slog.DebugContext(c, "session token renewed", "user", c.User.Subject())
		} else {
			slog.WarnContext(c, "failed to renew token", "error", err)
		}
	}

	cookie.Expires = time.Now().Add(ttl)
	cookie.MaxAge = int(ttl.Seconds())
	cookie.HttpOnly = true
	cookie.SameSite = http.SameSiteLaxMode
	cookie.Secure = c.Request.TL
