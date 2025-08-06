# Router

The Router is the core component of Mux that handles HTTP request routing and middleware execution.

## Creating a Router

```go
// Basic router
router := mux.NewRouter()

// Router with options
router := mux.NewRouter(
    mux.WithTitle("My API"),
    mux.WithVersion("1.0.0"),
    mux.WithDescription("A sample API built with Mux"),
)
```

## Router Options

Configure your router with these options:

### Basic Information
```go
mux.WithTitle("My API")
mux.WithVersion("1.0.0") 
mux.WithDescription("API description")
mux.WithTermsOfService("https://example.com/terms")
```

### Contact Information
```go
mux.WithContact("API Support", "https://example.com/support", "support@example.com")
```

### License Information
```go
mux.WithLicense("MIT", "https://opensource.org/licenses/MIT")
```

### Authentication
```go
signer, _ := jwtkit.NewSigner("secret-key")
ttl := time.Hour * 24
mux.WithAuth(signer, &ttl)
```

## Adding Routes

The Router embeds RouteGroup, so you can add routes directly:

```go
router.GET("/health", healthCheck)
router.POST("/users", createUser)
router.PUT("/users/{id}", updateUser)
router.DELETE("/users/{id}", deleteUser)
```

## Route Groups

Create organized route groups with shared configuration:

```go
// API v1 group
api := router.NewRouteGroup("/api/v1")
api.WithTags("API v1")

// Users sub-group
users := api.NewRouteGroup("/users")
users.WithTags("Users")
users.RequireRoles("user")

users.GET("/", listUsers)
users.POST("/", createUser)
```

## Middleware

Add middleware to apply cross-cutting concerns:

### Built-in Middleware
```go
// Request logging
router.UseLogging()

// Response compression  
router.UseCompression()

// HTTPS enforcement
router.UseEnforceHTTPS()

// Parse forwarded headers
router.UseForwardedHeaders()

// Authentication
router.UseAuthentication(
    mux.WithValidator(validateToken),
    mux.WithTokenCreator(createToken),
    mux.WithTokenTTL(30 * time.Minute),
)

// Authorization
router.UseAuthorization(
    mux.WithRoles("admin", "user"),
    mux.WithPermissions("read", "write"),
)

// Service injection
router.UseServices(
    mux.WithService("db", database),
    mux.WithService("cache", redisClient),
)

// Rate limiting (applied per route)
router.GET("/api/data", handler).
    WithRateLimit(100, time.Minute)

// OpenTelemetry tracing
router.UseOpenTelemetry()

// Export control with GeoIP
router.UseExportControl(mux.WithGeoIPDatabase(geoipDB))
```

### Custom Middleware
```go
type LoggingMiddleware struct{}

func (m *LoggingMiddleware) Invoke(c *mux.RouteContext, next mux.HandlerFunc) {
    start := time.Now()
    log.Printf("Starting request: %s %s", c.Request.Method, c.Request.URL.Path)
    
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
func searchUsers(c *mux.RouteContext) {
    query, _ := c.QueryValue("q")
    page, _ := c.QueryInt("page")
    limit, _ := c.QueryInt("limit")
    
    // Use parameters...
}

router.GET("/users/search", searchUsers)
```

## HTTP Methods

All standard HTTP methods are supported:

```go
router.GET("/resource", getResource)
router.POST("/resource", createResource)  
router.PUT("/resource/{id}", updateResource)
router.DELETE("/resource/{id}", deleteResource)
router.HEAD("/resource/{id}", headResource)
```

## Handler Functions

Handler functions receive a RouteContext with request/response access:

```go
func createUser(c *mux.RouteContext) {
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

## Error Handling

The router automatically handles panics and returns structured error responses:

```go
func riskyHandler(c *mux.RouteContext) {
    panic("Something went wrong") // Automatically recovered
}

// Results in 500 Internal Server Error with proper JSON response
```

## Serving HTTP

The Router implements `http.Handler`:

```go
// Standard HTTP server
http.ListenAndServe(":8080", router)

// With TLS
http.ListenAndServeTLS(":8443", "cert.pem", "key.pem", router)

// Using Mux's built-in server
server := mux.NewServer(":8080", router)
server.Start()
```

## OpenAPI Specification

Generate OpenAPI specs from your routes:

```go
generator := mux.NewGenerator()
spec := generator.GenerateSpec(router)
spec.MarshalToFile("openapi.yaml")
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