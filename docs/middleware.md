# Built-in Middleware Guide

Mux includes a comprehensive set of built-in middleware to handle common cross-cutting concerns in web applications. This guide covers all available middleware, their configuration options, and usage patterns.

## Overview

Middleware in Mux follows a functional options pattern and implements the `Middleware` interface:

```go
type Middleware interface {
    Invoke(c routing.RouteContext, next HandlerFunc)
}
```

Built-in middleware is installed on the router and executes in the order it is added.

## Execution Model

- Order: Middleware runs in the order you register it with `router.Use(...)`. The last registered middleware wraps the handler last.
- Short-circuiting: A middleware may choose not to call `next(c)`. In that case, it terminates the pipeline early (e.g., to reject unauthorized requests).
- Scope: Middleware is registered on the router. Use `AllowAnonymous()` and route-group defaults such as `RequireRoles(...)`, `RequireScopes(...)`, and `RequirePermission(...)` to shape behavior for subsets of routes.
- Performance: Mux composes the middleware pipeline and caches it. The pipeline is rebuilt only when middleware are added, avoiding per-request allocations.

## Authentication Middleware

Provides JWT token validation and creation capabilities with support for multiple token sources.

### Basic Setup
```go
mux.UseAuthentication(router,
    mux.WithValidator(validateToken),
    mux.WithTokenCreator(createToken),
    mux.WithTokenTTL(30 * time.Minute),
)
```

### Configuration Options
- `WithValidator(func(string) (claims.Principal, error))` - Token validation function
- `WithTokenCreator(func(claims.Principal, time.Duration) (string, error))` - Token creation function
- `WithTokenTTL(time.Duration)` - Token time-to-live duration
- `WithCSRFProtection()` - Double-submit CSRF protection for cookie-authenticated state-changing requests
- `WithAuthRateLimiter(func(string) bool)` - Rate limiting for failed authentication attempts

Mark routes or route groups public with `AllowAnonymous()` rather than configuring anonymous access on the middleware itself.

### Token Sources
The middleware checks tokens in these locations:
1. **Authorization header**: `Authorization: Bearer <token>`
2. **App session cookie**: The framework-managed session cookie used by `c.SignIn(...)`

### Example Implementation
```go
func validateToken(tokenString string) (claims.Principal, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        return []byte("your-secret-key"), nil
    })
    
    if err != nil || !token.Valid {
        return nil, errors.New("invalid token")
    }

    mapClaims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        return nil, errors.New("invalid token claims")
    }

    claimSet := claims.NewClaimsSet("")
    if sub, ok := mapClaims["sub"].(string); ok {
        claimSet.SetSubject(sub)
    }

    return claims.NewPrincipal(claimSet), nil
}

func createToken(principal claims.Principal, ttl time.Duration) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "sub":   principal.Subject(),
        "roles": principal.Roles(),
        "exp":   time.Now().Add(ttl).Unix(),
    })
    
    secret := os.Getenv("JWT_SECRET")
    return token.SignedString([]byte(secret))
}
```

## Authorization Middleware

Provides role-based and permission-based access control that works with the authentication middleware.

### Setup
```go
mux.UseAuthorization(router,
    mux.WithRoles("admin", "user"),
    mux.WithPermissions("read", "write", "delete"),
)
```

### Configuration Options
- `WithRoles(roles ...string)` - Require roles at middleware level
- `WithPermissions(permissions ...string)` - Require permissions at middleware level

### Route-Level Authorization
```go
// Require specific roles for a route
router.GET("/admin", adminHandler).RequireRoles("admin")

// Require specific permissions
router.POST("/users", createUser).RequirePermission("write")

// Combine roles and permissions
router.DELETE("/users/{id}", deleteUser).
    RequireRoles("admin", "moderator").
    RequirePermission("delete")
```

## Compression Middleware

Provides automatic response compression using gzip or deflate encoding based on client Accept-Encoding headers.

### Setup
```go
mux.UseCompression(router)
```

### Features
- **Automatic encoding detection**: Prefers gzip over deflate
- **Content-type aware**: Only compresses appropriate content types
- **Minimal overhead**: Efficient compression with built-in buffering
- **Client negotiation**: Respects client Accept-Encoding preferences

### Usage Example
```go
router.UseCompression()

router.GET("/api/data", func(c mux.RouteContext) {
    // Large JSON response will be automatically compressed
    data := generateLargeDataSet()
    c.OK(data)
})
```

## Logging Middleware

Provides structured HTTP request/response logging using Go's structured logging (slog).
Successful requests are emitted at DEBUG, 4xx responses at WARN, and 5xx responses at ERROR.

### Setup
```go
mux.UseLogging(router)
```

### Log Output
Each request generates a structured log entry with:
- **method**: HTTP method (GET, POST, etc.)
- **route**: Matched route pattern when available
- **path**: Request path
- **status**: HTTP response status code
- **level**: `DEBUG` for responses below 400, `WARN` for 4xx, `ERROR` for 5xx
- **remote**: Client IP address
- **user_agent**: Client user agent string
- **duration**: Request processing time
- **trace_id** / **span_id**: Included automatically when OpenTelemetry tracing is active

### Example Log Entry
```json
{
"time": "2024-01-15T10:30:00Z",
"level": "DEBUG",
"msg": "GET /api/users/{id} -> 200",
"method": "GET",
"route": "/api/users/{id}",
"path": "/api/users/42",
"status": 200,
"remote": "192.168.1.100",
"user_agent": "Mozilla/5.0...",
"duration": "25ms"
}
```

## Rate Limiting Middleware

Provides per-route token bucket rate limiting with automatic cleanup of expired entries.

### Setup
Rate limiting is configured per route, not globally:

```go
// Allow 100 requests per minute for this endpoint
router.GET("/api/search", searchHandler).
    WithRateLimit(100, time.Minute)

// Different limits for different endpoints
router.POST("/api/upload", uploadHandler).
    WithRateLimit(10, time.Minute)
```

### Configuration Options
- **limit**: Number of requests allowed
- **interval**: Time window for the limit
- **CleanupInterval**: How often to clean expired visitor records (default: 10 minutes)

### Features
- **Per-IP tracking**: Separate limits for each client IP
- **Per-route limits**: Different limits for different endpoints
- **Token bucket algorithm**: Allows bursts up to the limit
- **Automatic cleanup**: Expired visitor records are automatically removed
- **Memory efficient**: Uses minimal memory for tracking

### Advanced Configuration
```go
// Create rate limiter with custom cleanup interval
rateLimiter := mux.NewSelectiveRateLimiter(
    mux.WithCleanupInterval(5 * time.Minute),
)

// Use the custom rate limiter
router.Use(rateLimiter)
```

## HTTPS Enforcement Middleware

Automatically redirects HTTP requests to HTTPS and sets security headers.

### Setup
```go
mux.UseEnforceHTTPS(router)
```

### Features
- **Automatic redirection**: HTTP requests are redirected to HTTPS
- **Security headers**: Sets appropriate security headers
- **Development mode**: Can be disabled for local development

### Behavior
- Responds with `301 Moved Permanently` for HTTP requests
- Redirects to the same URL with HTTPS scheme
- Preserves query parameters and path

### Example
```go
router.UseEnforceHTTPS()

// All routes now require HTTPS
router.GET("/api/secure", secureHandler)
```

## Forwarded Headers Middleware

Parses and validates forwarded headers from proxies and load balancers.

### Setup
```go
mux.UseForwardedHeaders(router)
```

### Supported Headers
- `X-Forwarded-For` - Original client IP
- `X-Forwarded-Proto` - Original protocol (http/https)  
- `X-Forwarded-Host` - Original host header
- `X-Real-IP` - Real client IP (alternative to X-Forwarded-For)

### Usage
After adding the middleware, forwarded headers are automatically parsed and made available:

```go
func handler(c mux.RouteContext) {
    // Get real client IP (considering forwarded headers)
    realIP := getRealIP(c.Request())
    
    // Original protocol
    proto := c.Request().Header.Get("X-Forwarded-Proto")
    
    c.OK(map[string]string{
        "client_ip": realIP,
        "protocol":  proto,
    })
}
```

## Export Control Middleware

Provides geographic access restrictions using GeoIP databases for compliance with export control regulations.

### Setup
```go
// Load GeoIP database
geoipDB, err := geoip2.Open("GeoLite2-Country.mmdb")
if err != nil {
    log.Fatal(err)
}
defer geoipDB.Close()

// Add export control middleware
mux.UseExportControl(router,
    mux.WithGeoIPDatabase(geoipDB),
)
```

### Features
- **Country-based blocking**: Blocks requests from specific countries
- **GeoIP integration**: Uses MaxMind GeoLite2/GeoIP2 databases
- **Configurable restrictions**: Built-in export control country list
- **Real IP detection**: Works with forwarded headers

### Export Restricted Countries
The middleware includes a built-in list of export-restricted countries based on common compliance requirements. Requests from these countries receive a `403 Forbidden` response.

### Database Setup
1. Download GeoLite2 Country database from MaxMind
2. Extract the `.mmdb` file
3. Load it in your application:

```go
db, err := geoip2.Open("path/to/GeoLite2-Country.mmdb")
if err != nil {
    log.Fatal(err)
}
```

## OpenTelemetry Middleware

Provides distributed tracing and metrics collection using OpenTelemetry.

### Setup
```go
mux.UseOpenTelemetry(router,
    mux.WithOperation("my-api"),
)
```

### Configuration Options
- `WithOperation(name string)` - Sets the operation name for traces (default: "http.server")

### Default Route Tracing Behavior
- Span name uses `METHOD + route pattern` when route metadata is available (example: `GET /users/{id}`)
- Adds `http.route` with the resolved route pattern
- Adds `http.request.method` with the HTTP method
- Adds `mux.route.pattern` as a mux-specific route label
- Falls back to `WithOperation(...)` (or `http.server`) when route metadata is unavailable

### Features
- **Automatic span creation**: Creates spans for each HTTP request
- **Per-route grouping**: Uses route templates instead of raw URL paths to avoid high-cardinality span names
- **Request/response metrics**: Collects timing and status metrics  
- **Context propagation**: Properly propagates trace context
- **Integration ready**: Works with standard OpenTelemetry exporters

### Integration Example
```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/jaeger"
    "go.opentelemetry.io/otel/sdk/trace"
)

func main() {
    // Setup OpenTelemetry exporter
    exp, err := jaeger.New(jaeger.WithCollectorEndpoint())
    if err != nil {
        log.Fatal(err)
    }
    
    tp := trace.NewTracerProvider(
        trace.WithBatcher(exp),
        trace.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceNameKey.String("my-api"),
        )),
    )
    otel.SetTracerProvider(tp)
    
    // Add OpenTelemetry middleware
    mux.UseOpenTelemetry(router)
}
```

## Scoped Services

Register services explicitly when middleware and handlers need shared collaborators.

### Setup
```go
router.Services().
    Register("db", databaseConnection).
    Register("cache", redisClient).
    Register("logger", logger)
```

### Using Services in Handlers
```go
func getUserHandler(c mux.RouteContext) {
    // Retrieve services from context
    db, ok := c.GetService("db")
    if !ok {
        c.ServerError("Database unavailable", "")
        return
    }
    
    cache, _ := c.GetService("cache")
    logger, _ := c.GetService("logger")
    
    // Use services
    user := db.(*sql.DB).QueryRow("SELECT * FROM users WHERE id = ?", userID)
    logger.(*slog.Logger).Info("User retrieved", "id", userID)
    
    c.OK(user)
}
```

### Type-Safe Service Access
```go
type Services struct {
    DB     *sql.DB
    Cache  redis.Client
    Logger *slog.Logger
}

func getServicesFromContext(c mux.RouteContext) *Services {
    db, _ := c.GetService("db")
    cache, _ := c.GetService("cache") 
    logger, _ := c.GetService("logger")
    
    return &Services{
        DB:     db.(*sql.DB),
        Cache:  cache.(redis.Client),
        Logger: logger.(*slog.Logger),
    }
}
```

## Middleware Ordering

The order in which middleware is added matters. Here's the recommended order:

```go
// 1. Infrastructure middleware (comes first)
mux.UseForwardedHeaders(router)    // Parse proxy headers
mux.UseLogging(router)             // Log all requests

// 2. Security middleware
mux.UseEnforceHTTPS(router)        // Force HTTPS
mux.UseExportControl(router, ...)  // Geographic restrictions

// 3. Application middleware
mux.UseCompression(router)         // Compress responses
mux.UseOpenTelemetry(router)       // Tracing and metrics

// 4. Authentication & Authorization
mux.UseAuthentication(router, ...) // Authenticate users
mux.UseAuthorization(router, ...)    // Authorize access

// 5. Application services
router.Services().Register(...) // Shared collaborators

// 6. Rate limiting (per route)
// Applied individually to routes as needed
```

## Best Practices

### 1. Minimal Middleware
Only add middleware you actually need. Each middleware adds processing overhead.

### 2. Proper Ordering
Add middleware in logical order - authentication before authorization, logging early in the pipeline.

### 3. Error Handling
Middleware should handle errors gracefully and return appropriate HTTP responses.

### 4. Configuration
Use environment variables for middleware configuration in production:

```go
if os.Getenv("ENABLE_COMPRESSION") == "true" {
    mux.UseCompression(router)
}

if os.Getenv("ENABLE_RATE_LIMITING") == "true" {
    // Add rate limiting to specific routes
}
```

### 5. Testing
Test your middleware independently and as part of the full pipeline:

```go
func TestLoggingMiddleware(t *testing.T) {
    router := mux.NewRouter()
    mux.UseLogging(router)
    
    // Test that requests are logged
    // ...
}
```

### 6. Custom Middleware
When building custom middleware, follow the same patterns as built-in middleware:

```go
type CustomMiddleware struct {
    options *CustomOptions
}

func (m *CustomMiddleware) Invoke(c mux.RouteContext, next mux.HandlerFunc) {
    // Pre-processing
    
    next(c) // Always call next
    
    // Post-processing
}
```

## Performance Considerations

- **Compression**: Reduces bandwidth but increases CPU usage
- **Rate Limiting**: Memory usage grows with unique IP/route combinations
- **Authentication**: JWT validation has cryptographic overhead
- **OpenTelemetry**: Adds telemetry overhead for observability benefits
- **GeoIP**: Database lookups add latency but provide security

Monitor your application's performance and adjust middleware configuration as needed.

## See Also

- [Authentication Middleware](authentication-middleware.md) - Detailed JWT authentication guide
- [Custom Middleware](custom-middleware.md) - Build your own middleware
- [Router](router.md) - Routing fundamentals
- [Best Practices](best-practices.md) - Patterns and conventions
- [WebServer](webserver.md) - Production server setup