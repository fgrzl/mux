# Authentication Middleware

The authentication middleware validates bearer tokens and session cookies,
optionally issues refreshed session tokens, and can enforce CSRF protection for
cookie-based authentication.

> **Note**: This page focuses on the actual behavior shipped by the framework.
> For the full middleware catalog, see [Middleware](middleware.md).

## Setup

```go
mux.UseAuthentication(router,
    mux.WithAuthValidator(validateToken),
    mux.WithAuthTokenCreator(createToken),
    mux.WithAuthTokenTTL(30 * time.Minute),
)
```

## What The Middleware Checks

The middleware looks for credentials in this order:

1. `Authorization: Bearer <token>`
2. The configured app session cookie

If neither source yields a valid principal, the middleware returns `401 Unauthorized`.

## Core Configuration

### Token Validation

Provide a validator that turns a token string into a `claims.Principal`.

```go
func validateToken(token string) (claims.Principal, error) {
    // Parse token, validate signature and claims, and return a principal.
}

mux.UseAuthentication(router,
    mux.WithAuthValidator(validateToken),
)
```

### Token Creation

Provide a token creator if you want the middleware and `c.Cookies().SignIn()` to issue or renew session tokens.

```go
func createToken(principal claims.Principal, ttl time.Duration) (string, error) {
    // Build and sign a token for principal that expires after ttl.
}

mux.UseAuthentication(router,
    mux.WithAuthValidator(validateToken),
    mux.WithAuthTokenCreator(createToken),
    mux.WithAuthTokenTTL(30 * time.Minute),
)
```

### Additional Checks

The middleware also supports:

- `mux.WithAuthTokenRevocationChecker(...)`
- `mux.WithAuthIssuerValidator(...)`
- `mux.WithAuthAudienceValidator(...)`
- `mux.WithAuthContextEnricher(...)`
- `mux.WithAuthRateLimiter(...)`

`WithAuthRateLimiter` is applied to **failed authentication attempts**, not all incoming requests.

## Accessing The User In Handlers

`RouteContext.User()` returns the authenticated principal or `nil` when the request is unauthenticated.

```go
func getProfile(c mux.RouteContext) {
    user := c.User()
    if user == nil {
        c.Unauthorized()
        return
    }

    c.OK(map[string]any{
        "subject": user.Subject(),
        "email":   user.Email(),
        "roles":   user.Roles(),
        "scopes":  user.Scopes(),
    })
}
```

## Anonymous Routes

Mark public routes explicitly:

```go
router.POST("/login", loginHandler).AllowAnonymous()

public := router.Group("/public")
public.AllowAnonymous()
```

## Cookie Sessions And CSRF

If you use cookie-based authentication, enable CSRF protection:

```go
mux.UseAuthentication(router,
    mux.WithAuthValidator(validateToken),
    mux.WithAuthTokenCreator(createToken),
    mux.WithAuthCSRFProtection(),
)
```

When CSRF protection is enabled, state-changing cookie-authenticated requests
(`POST`, `PUT`, `PATCH`, `DELETE`) must send both:

- the `csrf_token` cookie
- the matching `X-CSRF-Token` header

Bearer-token requests are not subject to this CSRF check.

### Issuing A CSRF Token

Prefer the error-returning API when establishing a session:

```go
func loginHandler(c mux.RouteContext) {
    var creds LoginRequest
    if err := c.Bind(&creds); err != nil {
        c.BadRequest("Invalid credentials", err.Error())
        return
    }

    user := authenticateUser(creds)
    if user == nil {
        c.Unauthorized()
        return
    }

    if _, err := c.Cookies().CSRFTokenErr(); err != nil {
        c.ServerError("CSRF Setup Failed", err.Error())
        return
    }

    c.Cookies().SignIn(user, "/dashboard")
}
```

## Logout

`c.Cookies().SignOut(...)` clears the framework-managed authentication cookies, including
the CSRF token cookie, and then redirects.

If you issued cookies with custom `Path` or `Domain` options, use
`mux.SignOutWithOptions(...)` with the same cookie options so the browser
matches the deletion cookies correctly.

```go
func logoutHandler(c mux.RouteContext) {
    c.Cookies().SignOut("/signed-out")
}

func logoutFromScopedCookie(c mux.RouteContext) {
    mux.SignOutWithOptions(c, "/signed-out",
        mux.WithCookiePath("/app"),
        mux.WithCookieDomain(".example.com"),
    )
}
```

## Cookie Names

You can customize the framework-managed cookie names when you install the middleware:

```go
mux.UseAuthentication(router,
    mux.WithAuthValidator(validateToken),
    mux.WithAuthTokenCreator(createToken),
    mux.WithAuthAppSessionCookieName("my_app_token"),
    mux.WithAuthTwoFactorCookieName("my_2fa_token"),
    mux.WithAuthIDPSessionCookieName("my_idp_token"),
)
```

## Rate Limiting Notes

`mux.NewInMemoryRateLimiter(...)` is a simple in-memory limiter for failed
authentication attempts. It is appropriate for single-instance deployments and
tests, not distributed enforcement.

Client identification uses the normalized request `RemoteAddr`. If your app runs
behind a reverse proxy, install the forwarded-header middleware so trusted proxy
metadata is applied to the request before authentication runs.

## Integration With Authorization

Authentication should be installed before authorization:

```go
mux.UseAuthentication(router,
    mux.WithAuthValidator(validateToken),
)

mux.UseAuthorization(router,
    mux.WithAuthorizationRoles("admin"),
)
```

Middleware-level and route-level roles, scopes, and permissions are all enforced.

## Best Practices

1. Use strong signing keys and validate all token claims.
2. Prefer `c.Cookies().CSRFTokenErr()` in session-establishing handlers so failures stay explicit.
3. Use HTTPS and keep cookie `Secure` and `HttpOnly` defaults unless you have a specific reason not to.
4. Use `WithAuthRateLimiter` to slow down failed authentication attempts.
5. Treat forwarded client-IP headers as trusted only when your proxy strips and rewrites them.
6. Keep public routes explicit with `AllowAnonymous()` instead of implicitly bypassing auth.

## See Also

- [Middleware](middleware.md)
- [Custom Middleware](custom-middleware.md)
- [Best Practices](best-practices.md)
- [Router](router.md)



