# Authentication Middleware

The authentication middleware provides JWT token validation and creation capabilities.

> **Note**: This document provides detailed authentication examples. For a complete overview of all built-in middleware, see the [Middleware Guide](middleware.md).

## Setup

```go
mux.UseAuthentication(router,
    mux.WithValidator(validateToken),
    mux.WithTokenCreator(createToken),
    mux.WithTokenTTL(30 * time.Minute),
)
```

## Configuration Options

### Token Validator

Provide a function to validate JWT tokens:

```go
func validateToken(tokenString string) (claims.Principal, error) {
    // Parse and validate JWT token
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        return []byte("secret-key"), nil
    })
    
    if err != nil || !token.Valid {
        return nil, errors.New("invalid token")
    }
    
    // Extract claims and create Principal
    claims := token.Claims.(jwt.MapClaims)
    principal := claims.NewPrincipal()
    principal.SetUserID(claims["sub"].(string))
    principal.SetEmail(claims["email"].(string))
    
    return principal, nil
}

mux.UseAuthentication(router,
    mux.WithValidator(validateToken),
)
```

### Token Creator

Provide a function to create JWT tokens:

```go
func createToken(principal claims.Principal, ttl time.Duration) (string, error) {
    now := time.Now()
    claims := jwt.MapClaims{
        "sub":   principal.UserID(),
        "email": principal.Email(),
        "iat":   now.Unix(),
        "exp":   now.Add(ttl).Unix(),
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte("secret-key"))
}

mux.UseAuthentication(router,
    mux.WithTokenCreator(createToken),
)
```

### Token TTL

Set the time-to-live for tokens:

```go
router.UseAuthentication(
    mux.WithTokenTTL(24 * time.Hour), // 24 hours
)
```

## Token Sources

The middleware automatically looks for tokens in multiple locations:

1. **Authorization Header**: `Authorization: Bearer <token>`
2. **Cookie**: Session cookies (configurable name)
3. **Query Parameter**: `?token=<token>` (for special cases)

## Usage in Handlers

Access the authenticated user in your handlers:

```go
func protectedHandler(c mux.RouteContext) {
    // Check if user is authenticated
    if !c.User.IsAuthenticated() {
        c.Unauthorized()
        return
    }
    
    // Access user information
    userID := c.User.UserID()
    email := c.User.Email()
    roles := c.User.Roles()
    
    c.OK(map[string]interface{}{
        "user_id": userID,
        "email":   email,
        "roles":   roles,
    })
}
```

## Token Refresh

Tokens are automatically refreshed when they're close to expiring:

```go
// The middleware checks if token expires within 10% of its TTL
// If so, it creates a new token and sets it as a cookie
```

## Anonymous Access

Allow some routes to be accessed without authentication:

```go
// Route-level anonymous access
router.GET("/public", publicHandler).AllowAnonymous()

// Group-level anonymous access  
public := router.NewRouteGroup("/public")
public.AllowAnonymous()
```

## Error Handling

The middleware handles various authentication errors:

- **Missing Token**: Returns 401 Unauthorized
- **Invalid Token**: Returns 401 Unauthorized  
- **Expired Token**: Attempts refresh, returns 401 if refresh fails
- **Validation Error**: Returns 401 Unauthorized

## Cookie Configuration

Configure authentication cookies:

```go
// Set custom cookie names
mux.SetAppSessionCookieName("my_app_token")
mux.SetTwoFactorCookieName("my_2fa_token")
mux.SetIdpSessionCookieName("my_idp_token")

// Get current cookie names
appCookie := mux.GetUserCookieName()
```

## Integration with Authorization

Authentication middleware works seamlessly with authorization:

```go
// Authentication first, then authorization
mux.UseAuthentication(...)
mux.UseAuthorization(router,
    mux.WithRoles("admin", "user"),
)
```

## Complete Example

```go
package main

import (
    "time"
    "github.com/fgrzl/mux"
    "github.com/fgrzl/claims"
    "github.com/golang-jwt/jwt/v5"
)

func main() {
    router := mux.NewRouter()
    
    // Configure authentication
    mux.UseAuthentication(router,
        mux.WithValidator(validateJWT),
        mux.WithTokenCreator(createJWT),
        mux.WithTokenTTL(1 * time.Hour),
    )
    
    // Public routes
    router.POST("/login", loginHandler).AllowAnonymous()
    
    // Protected routes
    api := router.NewRouteGroup("/api")
    api.GET("/profile", getProfile)
    api.PUT("/profile", updateProfile)
    
    http.ListenAndServe(":8080", router)
}

func validateJWT(tokenString string) (claims.Principal, error) {
    // Implementation...
}

func createJWT(principal claims.Principal, ttl time.Duration) (string, error) {
    // Implementation...
}

func loginHandler(c mux.RouteContext) {
    // Validate credentials and create user session
    var creds LoginRequest
    if err := c.Bind(&creds); err != nil {
        c.BadRequest("Invalid credentials", err.Error())
        return
    }
    
    // Authenticate user...
    user := authenticateUser(creds)
    if user == nil {
        c.Unauthorized()
        return
    }
    
    // Sign in user (creates session cookie)
    c.SignIn(user, "/dashboard")
}

func getProfile(c mux.RouteContext) {
    if !c.User.IsAuthenticated() {
        c.Unauthorized()
        return
    }
    
    profile := getUserProfile(c.User.UserID())
    c.OK(profile)
}
```

## Best Practices

1. **Use strong secrets** for JWT signing
2. **Set appropriate TTLs** based on security requirements
3. **Implement token refresh** for long-lived sessions
4. **Log authentication failures** for security monitoring
5. **Use HTTPS** to protect tokens in transit
6. **Validate all token claims** thoroughly
7. **Handle token expiration** gracefully
8. **Consider rate limiting** login attempts
9. **Implement proper logout** functionality
10. **Use secure cookie flags** in production

## See Also

- [Middleware](middleware.md) - Built-in middleware guide (includes authentication)
- [Custom Middleware](custom-middleware.md) - Build your own middleware
- [Best Practices](best-practices.md) - Security and authentication patterns
- [Router](router.md) - Routing fundamentals