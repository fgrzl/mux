# Router

The Router is the core component of Mux that handles HTTP request routing and middleware execution.

## Creating a Router

```go
router := mux.NewRouter(
    mux.WithTitle("My API"),
    mux.WithVersion("1.0.0"),
)

if err := router.Configure(func(router *mux.Router) {
    router.GET("/health", healthCheck)
}); err != nil {
    panic(err)
}
```

`Configure` is the recommended startup path because it returns configuration errors directly.

## Router Options

### Basic Information
```go
mux.WithTitle("My API")
mux.WithVersion("1.0.0")
mux.WithTermsOfService("https://example.com/terms")
mux.WithContact("API Support", "https://example.com/support", "support@example.com")
mux.WithLicense("MIT", "https://opensource.org/licenses/MIT")
```

### Request Handling Options
```go
// Serve HEAD via GET when no explicit HEAD handler exists for a path
mux.WithHeadFallbackToGet()

// Limit request body size used by Bind (JSON/form). Default is 1MB when <= 0
mux.WithMaxBodyBytes(2 << 20) // 2MB
```

> **Note**: Built-in middleware helpers (like `mux.UseLogging(router)`, `mux.UseCompression(router)`, etc.) are called with your router instance. See the [Middleware](middleware.md) guide for details.

## Adding Routes

The Router exposes the same day-to-day route registration methods you use on groups, so you can add routes directly:

```go
router.GET("/health", healthCheck)
router.POST("/users", createUser)
router.PUT("/users/{id}", updateUser)
router.DELETE("/users/{id}", deleteUser)
router.Use(&LoggingMiddleware{})
```

## Route Groups

Create organized route groups with shared configuration:

```go
// API v1 group
api := router.Group("/api/v1")
api.Tags("API v1")

// Users sub-group
users := api.Group("/users")
users.Tags("Users")
users.RequireRoles("user")

users.GET("/", listUsers)
users.POST("/", createUser)
```

## Middleware

Add middleware to apply cross-cutting concerns:

### Built-in Middleware
```go
// Request logging
mux.UseLogging(router)

// Response compression  
mux.UseCompression(router)

// HTTPS enforcement
mux.UseEnforceHTTPS(router)

// Parse forwarded headers
mux.UseForwardedHeaders(router)

// Authentication
mux.UseAuthentication(router,
    mux.WithAuthValidator(validateToken),
    mux.WithAuthTokenCreator(createToken),
    mux.WithAuthTokenTTL(30 * time.Minute),
)

// Authorization
mux.UseAuthorization(router,
    mux.WithAuthorizationRoles("admin", "user"),
    mux.WithAuthorizationPermissions("read", "write"),
)

// Scoped services
router.Services().
    Register(mux.ServiceKey("db"), database).
    Register(mux.ServiceKey("cache"), redisClient)

// Rate limiting (applied per route)
router.GET("/api/data", handler).
    RateLimit(100, time.Minute)

// OpenTelemetry tracing
mux.UseOpenTelemetry(router)

// Export control with GeoIP
mux.UseExportControl(router, mux.WithExportControlGeoIPDatabase(geoipDB))
```

### Custom Middleware
```go
type LoggingMiddleware struct{}

func (m *LoggingMiddleware) Invoke(c mux.MutableRouteContext, next mux.HandlerFunc) {
    start := time.Now()
    log.Printf("Starting request: %s %s", c.Request().Method, c.Request().URL.Path)
    
    next(c)
    
    duration := time.Since(start)
    log.Printf("Completed request in %v", duration)
}

// Add to router
router.Use(&LoggingMiddleware{})
```

## Route Patterns

Mux supports flexible route patterns:

### Static Routes
```go
router.GET("/health", healthCheck)
router.GET("/api/status", statusCheck)
```

### Path Parameters
```go
router.GET("/users/{id}", getUser)
router.GET("/users/{userID}/posts/{postID}", getUserPost)
```

### Query Parameters
Query parameters are accessed through the RouteContext:

```go
func searchUsers(c mux.RouteContext) {
    query := c.Query()
    search, _ := query.String("q")
    page, _ := query.Int("page")
    limit, _ := query.Int("limit")
    
    // Use parameters...
    _ = search
}

router.GET("/users/search", searchUsers)
```

## HTTP Methods

All standard HTTP methods are supported:

```go
router.GET("/resource", getResource)
router.POST("/resource", createResource)  
router.PUT("/resource/{id}", updateResource)
router.PATCH("/resource/{id}", patchResource)
router.DELETE("/resource/{id}", deleteResource)
router.HEAD("/resource/{id}", headResource)
router.OPTIONS("/resource", optionsResource)
router.TRACE("/resource", traceResource)
```

Notes:
- 405 Method Not Allowed: If a path matches but the HTTP method is not allowed, the router responds with 405 and sets the "Allow" header with the permitted methods for that path.
- Optional HEAD fallback: When WithHeadFallbackToGet() is enabled, HEAD requests without an explicit route will be served by the GET handler with the body suppressed.

## Handler Functions

Handler functions receive a RouteContext with request/response access:

```go
func createUser(c mux.RouteContext) {
    var user User
    if err := c.Bind(&user); err != nil {
        c.BadRequest("Invalid request", err.Error())
        return
    }
    
    // Create user...
    createdUser := service.CreateUser(user)
    
    c.Created(createdUser)
}
```

## Request Access Model

Mux keeps request data access source-grouped so handlers learn one rule and reuse it everywhere:

```go
func handler(c mux.RouteContext) {
    params := c.Params()
    query := c.Query()
    form := c.Form()
    headers := c.Headers()

    id, _ := params.String("id")
    page, _ := query.Int("page")
    name, _ := form.String("name")
    traceID, _ := headers.String("X-Trace-ID")

    _ = id
    _ = page
    _ = name
    _ = traceID
}
```

That source-first pattern is the canonical public model: `Params()`, `Query()`, `Form()`, `Headers()`, and `Cookies()`.

## Error Handling

The router automatically handles panics and returns structured error responses:

```go
func riskyHandler(c mux.RouteContext) {
    panic("Something went wrong") // Automatically recovered
}

// Results in 500 Internal Server Error with proper JSON response
```

## Serving HTTP

The Router implements `http.Handler`:

```go
// Recommended default
if err := router.Configure(func(router *mux.Router) {
    // Register routes and groups here.
}); err != nil {
    panic(err)
}

ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer cancel()

server := mux.NewServer(":8080", router)
if err := server.Listen(ctx); err != nil {
    panic(err)
}
```

The router still implements `http.Handler`, so you can pass it to a custom `http.Server` when you need lower-level control.

## OpenAPI Specification

Generate OpenAPI specs from your routes:

```go
generator := mux.NewGenerator()
spec, err := mux.GenerateSpecWithGenerator(generator, router)
if err != nil {
    panic(err)
}
if err := spec.MarshalToFile("openapi.yaml"); err != nil {
    panic(err)
}
```

## Best Practices

1. **Group related routes** using RouteGroup
2. **Add middleware in logical order** (auth before authorization)
3. **Use path parameters** for resource identifiers
4. **Implement proper error handling** in all handlers
5. **Document routes** with OpenAPI DSL methods
6. **Test your routes** thoroughly
7. **Use structured logging** for debugging
8. **Implement health checks** for monitoring
9. **Add rate limiting** for public endpoints
10. **Use HTTPS in production**

## See Also

- [Getting Started](getting-started.md) - Introduction to Mux
- [Middleware](middleware.md) - Built-in middleware guide
- [WebServer](webserver.md) - Production server with graceful shutdown
- [Best Practices](best-practices.md) - Patterns and conventions
- [Health Probes](health-probes.md) - Kubernetes-style health checks


