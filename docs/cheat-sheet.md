# Mux Cheat Sheet

Quick reference for common tasks in Mux.

---

## Installation

```bash
go get github.com/fgrzl/mux
```

---

## Basic Setup

### Quick Start
```go
import (
    "context"
    "os"
    "os/signal"
    "syscall"

    "github.com/fgrzl/mux"
)

router := mux.NewRouter()
if err := router.Configure(func(router *mux.Router) {
    router.GET("/health", healthHandler)
}); err != nil { panic(err) }

server := mux.NewServer(":8080", router)

ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer cancel()

if err := server.Listen(ctx); err != nil { panic(err) }
```

### Background Start
```go
router := mux.NewRouter()
server := mux.NewServer(":9090", router)

ctx, cancel := context.WithCancel(context.Background())
defer cancel()

if err := server.Start(ctx); err != nil { panic(err) }
```

### With HTTPS
```go
server := mux.NewServer(":8443", router,
    mux.WithTLS("server.crt", "server.key"),
)
if err := server.Listen(ctx); err != nil { panic(err) }
```

---

## Routing

### Simple Routes
```go
router.GET("/path", handler)
router.POST("/path", handler)
router.PUT("/path", handler)
router.DELETE("/path", handler)
router.PATCH("/path", handler)
```

### Path Parameters
```go
router.GET("/users/{id}", func(c mux.RouteContext) {
    id, _ := c.Params().String("id")
    c.OK(map[string]string{"userId": id})
})
```

### Query Parameters
```go
router.GET("/search", func(c mux.RouteContext) {
    query := c.Query()
    term, ok := query.String("q")              // Single value
    values, _ := query.Strings("tag")          // Repeated values: ?tag=a&tag=b
    if !ok {
        c.BadRequest("Missing query", "q parameter is required")
        return
    }
    c.OK(map[string]any{"q": term, "tags": values})
})
```

### Wildcards
```go
router.GET("/files/*", handler)        // Single-segment wildcard
router.GET("/static/**", handler)      // Multi-segment catch-all
```

### Route Groups
```go
api := router.Group("/api/v1")
users := api.Group("/users")
users.GET("/", listUsers)
users.POST("/", createUsers)
```

---

## Request Handling

### Read JSON Body
```go
var data MyStruct
if err := c.Bind(&data); err != nil {
    c.BadRequest("Invalid request", err.Error())
    return
}
```

### Read Headers
```go
token, ok := c.Headers().String("Authorization")
contentType, ok := c.Headers().String("Content-Type")
```

### Read Cookies
```go
sessionID, err := c.Cookies().Get("session_id")
if err != nil {
    c.BadRequest("Missing cookie", err.Error())
    return
}
```

### Get Request Info
```go
method := c.Request().Method
path := c.Request().URL.Path
```

---

## Response Handling

### JSON Responses
```go
c.OK(data)                                // 200 OK
c.Created(data)                           // 201 Created
c.NoContent()                             // 204 No Content
```

### Error Responses
```go
c.BadRequest("Invalid input", "describe the validation error") // 400
c.Unauthorized()                          // 401
c.Forbidden("Access denied")              // 403
c.NotFound()                              // 404
c.Conflict("Resource already exists", "describe the conflict") // 409
c.ServerError("Error occurred", detail)   // 500
```

### Custom Status
```go
c.JSON(http.StatusCreated, data)
c.Plain(http.StatusTeapot, []byte("short and stout"))
```

### Set Headers
```go
c.Response().Header().Set("X-Custom", "value")
```

### Set Cookies
```go
c.Cookies().Set("session", "abc123", 3600, "/", "", true, true)
```

---

## Middleware

### Global Middleware
```go
mux.UseLogging(router)
mux.UseCompression(router)
mux.UseCORS(router, mux.WithCORSAllowedOrigins("*"))
mux.UseRateLimiter(router)
```

### Route Group Defaults
```go
// Middleware is installed on the router.
// Use route-group defaults to mark public or protected areas.
public := router.Group("/public")
public.AllowAnonymous()

admin := router.Group("/admin")
admin.RequireRoles("admin")
```

### Custom Middleware
```go
func MyMiddleware(c mux.MutableRouteContext, next mux.HandlerFunc) {
    // Before handler
    next(c)
    // After handler
}

router.Use(mux.MiddlewareFunc(MyMiddleware))
```

---

## Authentication

### Bearer Token
```go
mux.UseAuthentication(router,
    mux.WithAuthValidator(func(token string) (claims.Principal, error) {
        // Validate token and return a principal.
        claimSet := claims.NewClaimsSet("user-123")
        return claims.NewPrincipal(claimSet), nil
    }),
)
```

### Get Authenticated User
```go
user := c.User()
if user != nil {
    subject := user.Subject()
    _ = subject
}
```

### Cookie Auth
```go
mux.UseAuthentication(router,
    mux.WithAuthValidator(validateToken),
    mux.WithAuthTokenCreator(createToken),
    mux.WithAuthCSRFProtection(),
)

// Use c.Cookies().SignIn(...) to issue the framework-managed session cookie.
```

---

## OpenAPI Documentation

### Document Endpoint
```go
router.GET("/users/{id}", getUser).
    WithOperationID("getUser").
    WithSummary("Get user by ID").
    WithDescription("Returns a single user").
    WithPathParam("id", "The unique user identifier", "user-123").
    WithOKResponse(User{}).
    WithResponse(404, mux.ProblemDetails{})

// With query parameters
router.GET("/search", searchUsers).
    WithOperationID("searchUsers").
    WithQueryParam("q", "Search query", "john").      // Optional query param
    WithRequiredQueryParam("limit", "Maximum number of results", 10). // Required query param
    WithOKResponse([]User{})

// With header parameter
router.GET("/data", getData).
    WithHeaderParam("X-API-Version", "The API version", "v1").
    WithOKResponse(map[string]any{})

// Low-level (if needed)
router.GET("/custom", handler).
    WithPathParam("id", "Unique identifier", "123").
    WithQueryParam("filter", "Filter criteria", "active")
```

Validate generated input before you call the public builders; the root `mux` API intentionally omits the internal `Err`-returning variants.

### Automatic Type Inference
The framework automatically infers OpenAPI schemas from example values:

```go
// String parameter -> OpenAPI type: "string"
.WithPathParam("name", "Name of the entity", "john")

// Integer parameter -> OpenAPI type: "integer"
.WithQueryParam("age", "Age in years", 25)

// Boolean parameter -> OpenAPI type: "boolean"
.WithQueryParam("active", "Filter by active status", true)

// UUID -> OpenAPI type: "string", format: "uuid"
.WithPathParam("id", "Unique identifier", uuid.UUID{})

// Time -> OpenAPI type: "string", format: "date-time"
.WithQueryParam("createdAt", "Creation timestamp", time.Time{})

// Arrays -> OpenAPI type: "array"
.WithQueryParam("tags", "List of tags", []string{})

// Maps -> OpenAPI type: "object" with additionalProperties
.WithQueryParam("metadata", "Additional metadata", map[string]string{})
```

**Supported Types:**
- Primitives: `string`, `int`, `int64`, `float64`, `bool`
- Standard library: `uuid.UUID`, `time.Time`, `net.IP`, `url.URL`
- Collections: `[]T` (arrays), `map[string]T` (objects)
- Structs: Referenced as `#/components/schemas/TypeName`

### Generate Spec
```go
router.GET("/openapi.json", func(c mux.RouteContext) {
    spec, err := mux.GenerateSpecWithGenerator(mux.NewGenerator(), router)
    if err != nil {
        c.ServerError("Failed to generate OpenAPI spec", err.Error())
        return
    }
    c.OK(spec)
})
```

### Tag Routes
```go
api := router.Group("/api")
api.WithTags("API v1")
```

---

## Router Options

```go
router := mux.NewRouter(
    mux.WithContextPooling(),         // Enable context pooling
    mux.WithHeadFallbackToGet(),      // AUTO handle HEAD requests
    mux.WithMaxBodyBytes(10<<20),     // Set max body size (10MB)
)
```

---

## Health Check Pattern

```go
// Custom health check
router.GET("/health", func(c mux.RouteContext) {
    c.OK(map[string]string{
        "status": "healthy",
        "version": "1.0.0",
    })
})

// Built-in Kubernetes-style probes (always returns "ok")
router.Healthz()  // GET /healthz
router.Livez()    // GET /livez
router.Readyz()   // GET /readyz

// With custom health checks
router.HealthzWithReady(func(c mux.RouteContext) bool {
    return db.Ping() == nil && cache.Ready()
})

router.LivezWithCheck(func(c mux.RouteContext) bool {
    // Check if app is alive (not deadlocked)
    return runtime.NumGoroutine() < 10000
})

router.ReadyzWithCheck(func(c mux.RouteContext) bool {
    // Check if ready to serve traffic
    return db.Ready() && cache.Ready() && migration.Complete()
})
```

---

## CRUD Pattern

```go
users := router.Group("/users")
users.GET("/", listUsers)           // List all
users.POST("/", createUser)         // Create
users.GET("/{id}", getUser)         // Get one
users.PUT("/{id}", updateUser)      // Update
users.DELETE("/{id}", deleteUser)   // Delete
```

---

## Common Patterns

### Validation
```go
func createUser(c mux.RouteContext) {
    var user User
    if err := c.Bind(&user); err != nil {
        c.BadRequest("Invalid request", err.Error())
        return
    }
    
    if err := validate(user); err != nil {
        c.BadRequest("Validation failed", err.Error())
        return
    }
    
    c.Created(user)
}
```

### Pagination
```go
func listUsers(c mux.RouteContext) {
    page, ok := c.Query().String("page")
    if !ok {
        page = "1"
    }
    limit, ok := c.Query().String("limit")
    if !ok {
        limit = "10"
    }
    
    // Fetch paginated data
    users, total := getUsers(page, limit)
    
    c.OK(map[string]any{
        "data":  users,
        "total": total,
        "page":  page,
    })
}
```

### File Upload
```go
router.POST("/upload", func(c mux.RouteContext) {
    file, header, err := c.Request().FormFile("file")
    if err != nil {
        c.BadRequest("Missing file", "no file uploaded")
        return
    }
    defer file.Close()
    
    // Process file...
    c.OK(map[string]string{
        "filename": header.Filename,
        "size":     fmt.Sprintf("%d", header.Size),
    })
})
```

---

## Debugging Tips

### Enable Verbose Logging
```go
mux.UseLogging(router) // Logs all requests
```

### Check Route Registration
```go
// The router will log warnings for conflicting routes
```

### Inspect Request
```go
func handler(c mux.RouteContext) {
    fmt.Printf("Method: %s\n", c.Request().Method)
    fmt.Printf("Path: %s\n", c.Request().URL.Path)
    fmt.Printf("Headers: %v\n", c.Request().Header)
}
```

---

## Performance Tips

1. **Enable Context Pooling** for high-traffic APIs
   ```go
   router := mux.NewRouter(mux.WithContextPooling())
   ```

2. **Use Compression** for large responses
   ```go
   mux.UseCompression(router)
   ```

3. **Set Appropriate Limits**
   ```go
   router := mux.NewRouter(mux.WithMaxBodyBytes(5<<20)) // 5MB
   ```

4. **Cache OpenAPI Spec** instead of generating on every request

---

## Quick Links

- [Full Documentation](../README.md)
- [Learning Path](learning-path.md)
- [Examples](../examples/)
- [API Reference](https://pkg.go.dev/github.com/fgrzl/mux)

---

## Remember

- **One response per request** - Don't call `c.OK()` multiple times
- **Always validate input** - Use `c.Bind()` for type safety
- **Handle errors gracefully** - Use appropriate status codes
- **Document your API** - Use OpenAPI annotations
- **Test your endpoints** - Write tests for critical paths

---

**Print this page for quick reference while coding!**

## See Also

- [Quick Start](quick-start.md) - Get running in 5 minutes
- [Getting Started](getting-started.md) - Comprehensive introduction
- [Interactive Tutorial](interactive-tutorial.md) - Build a Todo API
- [Learning Path](learning-path.md) - Structured learning progression
- [Router](router.md) - Routing fundamentals
- [Middleware](middleware.md) - Built-in middleware guide




