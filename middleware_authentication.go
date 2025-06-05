package mux

import (
	"net/http"
	"strings"
	"time"

	"log/slog"

	"github.com/fgrzl/claims"
)

// UseAuthentication registers authentication middleware with the router.
func (r *Router) UseAuthentication(options *AuthenticationOptions) {
	r.middleware = append(r.middleware, &authenticationMiddleware{options: options})
}

// AuthenticationOptions define hooks and configuration for token authentication.
type AuthenticationOptions struct {
	// TokenTTL is the default duration to extend sessions by.
	// If zero or negative, session will not be extended.
	TokenTTL time.Duration

	// Validate verifies the token and returns associated claims.
	Validate func(token string) (claims.Principal, error)

	// CreateToken optionally re-issues a token to extend the session.
	CreateToken func(principal claims.Principal, ttl time.Duration) (string, error)
}

type authenticationMiddleware struct {
	options *AuthenticationOptions
}

type SessionDetails struct {
	IP        string
	UserAgent string
	CreatedAt time.Time
}

// Invoke checks for valid authentication via cookie or bearer token.
// If anonymous access is allowed, the request proceeds without validation.
func (m *authenticationMiddleware) Invoke(c *RouteContext, next HandlerFunc) {

	if c.Options.AllowAnonymous {
		slog.DebugContext(c, "authentication skipped: anonymous access allowed")
		next(c)
		return
	}

	// 1. Check for user token in cookie
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

	// 2. Check for bearer token
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

	// 3. Fallback: unauthorized
	slog.InfoContext(c, "authentication failed: no valid token found")
	c.Unauthorized()
}

// extendSessionExpiration optionally extends the cookie expiration and re-issues the token.
func (m *authenticationMiddleware) extendSessionExpiration(c *RouteContext, cookie *http.Cookie) {
	ttl := m.options.TokenTTL
	if ttl <= 0 {
		slog.DebugContext(c, "session extension skipped: TTL not set")
		return
	}

	duration := ttl

	if m.options.CreateToken != nil {
		if token, err := m.options.CreateToken(c.User, duration); err == nil {
			cookie.Value = token
			slog.DebugContext(c, "session token renewed", "user", c.User.Subject())
		} else {
			slog.WarnContext(c, "failed to renew token", "error", err)
		}
	}

	cookie.Expires = time.Now().Add(duration)
	cookie.MaxAge = int(duration.Seconds())
	cookie.HttpOnly = true
	cookie.SameSite = http.SameSiteLaxMode
	cookie.Secure = c.Request.TLS != nil
	cookie.Path = "/" // ensures consistent scope

	http.SetCookie(c.Response, cookie)
	slog.DebugContext(c, "session cookie extended", "expires", cookie.Expires)
}

// extractBearerToken retrieves a bearer token from the Authorization header.
func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}
