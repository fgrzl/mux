package mux

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/fgrzl/claims"
)

// SetCookie writes a cookie with the given attributes.
func (c *RouteContext) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	http.SetCookie(c.Response, &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		Secure:   secure,
		HttpOnly: httpOnly,
		SameSite: http.SameSiteLaxMode,
	})
}

// GetCookie returns the value of a named cookie, or an error if not found.
func (c *RouteContext) GetCookie(name string) (string, error) {
	cookie, err := c.Request.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// ClearCookie deletes the specified cookie.
func (c *RouteContext) ClearCookie(name string) {
	http.SetCookie(c.Response, &http.Cookie{
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
func (c *RouteContext) Authenticate(cookieName string, user claims.Principal) {
	if c.Options.AuthProvider == nil {
		panic("a signer is required if using authentication")
	}

	token, err := c.Options.AuthProvider.CreateToken(c, user)
	if err != nil {
		slog.ErrorContext(c, "failed to create token")
		return
	}

	c.SetCookie(cookieName, token, 0, "/", "", true, true)
}

// SignIn authenticates the user and redirects to the given URL (or "/" by default).
func (c *RouteContext) SignIn(user claims.Principal, redirectUrl string) {
	c.Authenticate(GetUserCookieName(), user)
	if redirectUrl == "" {
		redirectUrl = "/"
	}
	c.TemporaryRedirect(redirectUrl)
}

// SignOut clears user-related cookies and redirects to the logout page.
func (c *RouteContext) SignOut() {
	c.ClearCookie(GetUserCookieName())
	c.ClearCookie(GetTwoFactorCookieName())
	c.ClearCookie(GetIdpSessionCookieName())
	c.TemporaryRedirect("/logout")
}
