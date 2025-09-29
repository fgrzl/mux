package routing

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/fgrzl/claims"
	"github.com/fgrzl/mux/pkg/cookiejar"
	"github.com/fgrzl/mux/pkg/tokenizer"
)

// ...existing code...
// SetCookie writes a cookie with the given attributes.
func (c *DefaultRouteContext) SetCookie(
	name, value string,
	maxAge int,
	path, domain string,
	secure, httpOnly bool,
	sameSite ...http.SameSite, // optional SameSite (defaults to Lax)
) {
	var ss http.SameSite = http.SameSiteLaxMode
	if len(sameSite) > 0 {
		ss = sameSite[0]
	}
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		Domain:   domain,
		MaxAge:   maxAge,
		Secure:   secure,
		HttpOnly: httpOnly,
		SameSite: ss,
	}

	if maxAge > 0 {
		cookie.Expires = time.Now().Add(time.Duration(maxAge) * time.Second)
	} else if maxAge < 0 {
		// delete immediately
		cookie.Expires = time.Unix(1, 0).UTC()
	}

	// enforce Secure if SameSite=None
	if cookie.SameSite == http.SameSiteNoneMode && !cookie.Secure {
		cookie.Secure = true
	}

	http.SetCookie(c.Response(), cookie)
}

// GetCookie returns the value of a named cookie, or an error if not found.
func (c *DefaultRouteContext) GetCookie(name string) (string, error) {
	cookie, err := c.Request().Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// ClearCookie deletes the specified cookie.
func (c *DefaultRouteContext) ClearCookie(name string) {
	http.SetCookie(c.Response(), &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(1, 0), // old enough to be invalid
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

// Authenticate creates a JWT token for the user and stores it in a secure cookie.
// This method requires that authentication middleware has been added to the router using UseAuthentication().
func (c *DefaultRouteContext) Authenticate(cookieName string, user claims.Principal) {
	service, ok := c.GetService(tokenizer.ServiceKeyTokenProvider)
	if !ok {
		panic("DEVELOPMENT ERROR: No token provider available. Did you forget to call router.UseAuthentication() before using c.Authenticate()?")
	}

	provider, ok := service.(tokenizer.TokenProvider)
	if !ok {
		panic("DEVELOPMENT ERROR: Invalid token provider service. This indicates a bug in the authentication middleware.")
	}

	token, err := provider.CreateToken(c, user)
	if err != nil {
		slog.ErrorContext(c, "failed to create token", "error", err)
		return
	}

	// Use the provider's TTL for the cookie if available
	var maxAge int
	if ttl := provider.GetTTL(); ttl > 0 {
		maxAge = int(ttl.Seconds())
	}

	c.SetCookie(cookieName, token, maxAge, "/", "", true, true)
}

// SignIn authenticates the user and redirects to the given URL (or "/" by default).
func (c *DefaultRouteContext) SignIn(user claims.Principal, redirectUrl string) {
	c.Authenticate(cookiejar.GetUserCookieName(), user)
	if redirectUrl == "" {
		redirectUrl = "/"
	}
	c.TemporaryRedirect(redirectUrl)
}

// SignOut clears user-related cookies and redirects to the logout page.
func (c *DefaultRouteContext) SignOut() {
	c.ClearCookie(cookiejar.GetUserCookieName())
	c.ClearCookie(cookiejar.GetTwoFactorCookieName())
	c.ClearCookie(cookiejar.GetIdpSessionCookieName())
	c.TemporaryRedirect("/logout")
}
