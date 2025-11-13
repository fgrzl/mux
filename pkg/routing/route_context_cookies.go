package routing

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/fgrzl/claims"
	"github.com/fgrzl/mux/pkg/cookiekit"
	"github.com/fgrzl/mux/pkg/tokenizer"
)

// EnforceSecureForSameSiteNone controls whether SetCookie will force the Secure
// flag when SameSite=None is specified. Default true to match browser
// expectations; flip to false in local development environments that use plain HTTP.
var EnforceSecureForSameSiteNone = true

// ...existing code...
// SetCookie writes a cookie with the given attributes.
// It sets Expires based on maxAge and enforces Secure when SameSite=None
// unless EnforceSecureForSameSiteNone is disabled for development.
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
	// Behavior note: some browsers require Secure when SameSite=None. Making
	// this configurable allows development workflows using plain HTTP to opt-out.
	if cookie.SameSite == http.SameSiteNoneMode && !cookie.Secure {
		if EnforceSecureForSameSiteNone {
			cookie.Secure = true
			slog.Warn("SetCookie: Secure=true enforced because SameSite=None is set; this may break on HTTP in development", "cookieName", cookie.Name)
		} else {
			// Log a warning but do not modify the cookie when enforcement is disabled.
			slog.Warn("SetCookie: SameSite=None set but Secure enforcement is disabled; cookie will be sent without Secure flag", "cookieName", cookie.Name)
		}
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

// ClearCookie deletes the specified cookie by setting an expired cookie value.
func (c *DefaultRouteContext) ClearCookie(name string) {
	http.SetCookie(c.Response(), &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(1, 0).UTC(), // old enough to be invalid
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

// Authenticate creates a JWT token for the user and stores it in a secure cookie.
// This method requires that authentication middleware has been added to the router using UseAuthentication().
func (c *DefaultRouteContext) Authenticate(cookieName string, user claims.Principal, opts ...cookiekit.CookieOption) {
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
		// Use request context for logging to avoid passing the RouteContext
		// directly (its embedded Context may be nil if the instance was
		// pooled and later cleared). Fall back to background if unavailable.
		var logCtx context.Context = context.Background()
		if c != nil && c.Request() != nil && c.Request().Context() != nil {
			logCtx = c.Request().Context()
		}
		slog.ErrorContext(logCtx, "failed to create token", "error", err)
		return
	}

	// Use the provider's TTL for the cookie if available
	var providerMaxAge int
	if ttl := provider.GetTTL(); ttl > 0 {
		providerMaxAge = int(ttl.Seconds())
	}

	// Resolve cookie options provided by caller (if any)
	optMaxAge, path, domain, secure, httpOnly, sameSite := cookiekit.ResolveCookieOptions(opts...)

	// If caller didn't specify MaxAge (0), fall back to provider TTL
	finalMaxAge := optMaxAge
	if finalMaxAge == 0 {
		finalMaxAge = providerMaxAge
	}

	c.SetCookie(cookieName, token, finalMaxAge, path, domain, secure, httpOnly, sameSite)
}

// SignIn authenticates the user and redirects to the given URL (or "/" by default).
func (c *DefaultRouteContext) SignIn(user claims.Principal, redirectUrl string, opts ...cookiekit.CookieOption) {
	c.Authenticate(cookiekit.GetUserCookieName(), user, opts...)
	if redirectUrl == "" {
		redirectUrl = "/"
	}
	c.TemporaryRedirect(redirectUrl)
}

// SignOut clears user-related cookies and redirects to the logout page.
func (c *DefaultRouteContext) SignOut(redirectUrl string) {
	c.ClearCookie(cookiekit.GetUserCookieName())
	c.ClearCookie(cookiekit.GetTwoFactorCookieName())
	c.ClearCookie(cookiekit.GetIdpSessionCookieName())
	c.TemporaryRedirect(redirectUrl)
}
