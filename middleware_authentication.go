package mux

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/fgrzl/claims"
)

func (r *Router) UseAuthentication(options *AuthenticationOptions) {
	r.middleware = append(r.middleware, &authenticationMiddleware{options: options})
}

type AuthenticationOptions struct {
	Validate               func(token string) (claims.Principal, error)
	GetSessionDetails      func(sessionID string) (*SessionDetails, error)
	GetSessionCreationTime func(sessionID string) time.Time
}

type authenticationMiddleware struct {
	options *AuthenticationOptions
}

type SessionDetails struct {
	IP        string
	UserAgent string
	CreatedAt time.Time
}

func (m *authenticationMiddleware) Invoke(c *RouteContext, next HandlerFunc) {
	r := c.Request

	if c.Options.AllowAnonymous {
		next(c)
		return
	}

	if cookie, err := r.Cookie(GetUserCookieName()); err == nil {
		if principal, err := m.validateSession(cookie.Value, r); err == nil {
			c.User = principal
			next(c)
			m.extendSessionExpiration(c)
			return
		}
	}

	if token := extractBearerToken(r); token != "" {
		if principal, err := m.options.Validate(token); err == nil {
			c.User = principal
			next(c)
			return
		}
	}

	c.Unauthorized()
}

func (m *authenticationMiddleware) validateSession(sessionID string, r *http.Request) (claims.Principal, error) {
	details, err := m.options.GetSessionDetails(sessionID)
	if err != nil {
		return nil, err
	}

	if details.IP != r.RemoteAddr {
		return nil, fmt.Errorf("invalid remote addr")
	}

	if details.UserAgent != r.UserAgent() {
		return nil, fmt.Errorf("invalid user agent")
	}

	// Use sessionID as token input for validation (can be HMAC'd JWT)
	return m.options.Validate(sessionID)
}

func (m *authenticationMiddleware) extendSessionExpiration(c *RouteContext) {
	cookie, err := c.Request.Cookie(GetUserCookieName())
	if err != nil {
		return
	}

	duration := time.Hour
	cookie.Expires = time.Now().Add(duration)
	cookie.MaxAge = int(duration.Seconds())
	http.SetCookie(c.Response, cookie)
}

func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}
