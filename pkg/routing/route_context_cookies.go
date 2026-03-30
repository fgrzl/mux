package routing

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/fgrzl/claims"
	"github.com/fgrzl/mux/pkg/cookiekit"
	"github.com/fgrzl/mux/pkg/tokenizer"
)

const csrfCookieName = "csrf_token"

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
	ClearCookieWithOptions(c, name, cookiekit.WithSameSite(http.SameSiteLaxMode))
}

// ClearCookieWithOptions deletes the specified cookie using the provided path,
// domain, and attribute options. Use this when the cookie was originally set
// with non-default attributes so the browser matches the deletion cookie.
func ClearCookieWithOptions(c RouteContext, name string, opts ...cookiekit.CookieOption) {
	if c == nil || c.Response() == nil {
		return
	}

	_, path, domain, secure, httpOnly, sameSite := cookiekit.ResolveCookieOptions(opts...)
	http.SetCookie(c.Response(), &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     path,
		Domain:   domain,
		MaxAge:   -1,
		Expires:  time.Unix(1, 0).UTC(), // old enough to be invalid
		HttpOnly: httpOnly,
		Secure:   secure,
		SameSite: sameSite,
	})
}

// Authenticate creates a JWT token for the user and stores it in a secure cookie.
// This method requires that authentication middleware has been added to the router using UseAuthentication().
func (c *DefaultRouteContext) Authenticate(cookieName string, user claims.Principal, opts ...cookiekit.CookieOption) {
	if err := c.authenticate(cookieName, user, opts...); err != nil {
		c.ServerError("Authentication Misconfigured", err.Error())
	}
}

func (c *DefaultRouteContext) authenticate(cookieName string, user claims.Principal, opts ...cookiekit.CookieOption) error {
	service, ok := c.GetService(tokenizer.ServiceKeyTokenProvider)
	if !ok {
		return fmt.Errorf("no token provider available; register authentication middleware before calling Authenticate")
	}

	provider, ok := service.(tokenizer.TokenProvider)
	if !ok {
		return fmt.Errorf("invalid token provider service in route context")
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
		return fmt.Errorf("could not create session token")
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
	return nil
}

// SignIn authenticates the user and redirects to the given URL (or "/" by default).
func (c *DefaultRouteContext) SignIn(user claims.Principal, redirectUrl string, opts ...cookiekit.CookieOption) {
	c.Authenticate(cookiekit.GetUserCookieName(), user, opts...)
	if c.responseCommitted {
		return
	}
	if redirectUrl == "" {
		redirectUrl = "/"
	}
	c.TemporaryRedirect(redirectUrl)
}

// SignOut clears user-related cookies and redirects to the logout page.
func (c *DefaultRouteContext) SignOut(redirectUrl string) {
	SignOutWithOptions(c, redirectUrl, cookiekit.WithSameSite(http.SameSiteLaxMode))
}

// SignOutWithOptions clears user-related cookies using the provided cookie
// attributes and redirects to the logout page. Use this when the cookies were
// originally issued with non-default path or domain options.
func SignOutWithOptions(c RouteContext, redirectUrl string, opts ...cookiekit.CookieOption) {
	if c == nil {
		return
	}

	ClearCookieWithOptions(c, cookiekit.GetUserCookieName(), opts...)
	ClearCookieWithOptions(c, cookiekit.GetTwoFactorCookieName(), opts...)
	ClearCookieWithOptions(c, cookiekit.GetIdpSessionCookieName(), opts...)
	ClearCookieWithOptions(c, csrfCookieName, opts...)
	c.TemporaryRedirect(redirectUrl)
}
