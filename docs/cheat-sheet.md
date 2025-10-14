# Mux Cheat Sheet

Quick reference for common tasks in Mux.

---

## 📦 Installation

```bash
go get github.com/fgrzl/mux
```

---

## 🚀 Basic Setup

### Quick Start (Manual)
```go
router := mux.NewRouter()
http.ListenAndServe(":8080", router)
```

### Production Ready (WebServer)
```go
import (
    "context"
    "os/signal"
    "github.com/fgrzl/mux"
)

router := mux.NewRouter()
server := mux.NewServer(":8080", router)

ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
defer cancel()

server.Listen(ctx)  // Graceful shutdown included!
```

### With HTTPS
```go
server := mux.NewServer(":8443", router,
    mux.WithTLS("server.crt", "server.key"),
)
server.Listen(ctx)
```

---

## 🛣️ Routing

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
    id := c.Param("id")
    c.OK(map[string]string{"userId": id})
})
```

### Query Parameters
```go
router.GET("/search", func(c mux.RouteContext) {
    query, ok := c.Query("q")          // Single value
    values := c.QueryValues("tags")    // Multiple values
})
```

### Wildcards
```go
router.GET("/files/*", handler)        // Single-segment wildcard
router.GET("/static/**", handler)      // Multi-segment catch-all
```

### Route Groups
```go
api := router.NewRouteGroup("/api/v1")
users := api.NewRouteGroup("/users")
users.GET("/", listUsers)
users.POST("/", createUsers)
```

---

## 📥 Request Handling

### Read JSON Body
```go
var data MyStruct
if err := c.Bind(&data); err != nil {
    c.BadRequest(err.Error())
    return
}
```

### Read Headers
```go
token, ok := c.Header("Authorization")
contentType, ok := c.Header("Content-Type")
```

### Read Cookies
```go
sessionID, ok := c.Cookie("session_id")
```

### Get Request Info
```go
method := c.Request().Method
path := c.Request().URL.Path
```

---

## 📤 Response Handling

### JSON Responses
```go
c.OK(data)                                 // 200 OK
c.Created(data)                            // 201 Created
c.NoContent()                              // 204 No Content
```

### Error Responses
```go
c.BadRequest("Invalid input")              // 400
c.Unauthorized()                           // 401
c.Forbidden()                              // 403
c.NotFound()                               // 404
c.Conflict("Resource already exists")      // 409
c.ServerError("Error occurred", detail)    // 500
```

### Custom Status
```go
c.Status(http.StatusTeapot)
c.JSON(http.StatusCreated, data)
```

### Set Headers
```go
c.Response().Header().Set("X-Custom", "value")
```

### Set Cookies
```go
c.SetCookie(&http.Cookie{
    Name:  "session",
    Value: "abc123",
    Path:  "/",
})
```

---

## 🔧 Middleware

### Global Middleware
```go
mux.UseLogging(router)
mux.UseCompression(router)
mux.UseCORS(router, &mux.CORSOptions{...})
mux.UseRateLimit(router, &mux.RateLimitOptions{...})
```

### Group Middleware
```go
api := router.NewRouteGroup("/api")
mux.UseAuthentication(api, &mux.AuthenticationOptions{...})
```

### Custom Middleware
```go
func MyMiddleware(c mux.RouteContext, next mux.HandlerFunc) {
    // Before handler
    next(c)
    // After handler
}

router.Use(mux.MiddlewareFunc(MyMiddleware))
```

---

## 🔐 Authentication

### Bearer Token
```go
mux.UseAuthentication(router, &mux.AuthenticationOptions{
    Scheme: "Bearer",
    ValidateToken: func(token string) (any, error) {
        // Validate token, return user info
        return userID, nil
    },
})
```

### Get Authenticated User
```go
user := c.Principal()  // Returns the value from ValidateToken
```

### Cookie Auth
```go
mux.UseAuthentication(router, &mux.AuthenticationOptions{
    Scheme: "Cookie",
    CookieName: "session_id",
    ValidateToken: func(token string) (any, error) {
        // Validate session
        return user, nil
    },
})
```

---

## 📝 OpenAPI Documentation

### Document Endpoint
```go
router.GET("/users/{id}", getUser).
    WithOperationID("getUser").
    WithSummary("Get user by ID").
    WithDescription("Returns a single user").
    WithPathParam("id", "user-123").
    WithOKResponse(User{}).
    WithNotFoundResponse()

// With query parameters
router.GET("/search", searchUsers).
    WithOperationID("searchUsers").
    WithQueryParam("q", "john").           // Optional query param
    WithRequiredQueryParam("limit", 10).   // Required query param
    WithOKResponse([]User{})

// With header parameter
router.GET("/data", getData).
    WithHeaderParam("X-API-Version", "v1", false).  // Optional header
    WithOKResponse(map[string]any{})

// Low-level (if needed)
router.GET("/custom", handler).
    WithParam("id", "path", "123", true).  // name, in, example, required
    WithParam("filter", "query", "active", false)
```

### Automatic Type Inference
The framework automatically infers OpenAPI schemas from example values:

```go
// String parameter → OpenAPI type: "string"
.WithPathParam("name", "john")

// Integer parameter → OpenAPI type: "integer"
.WithQueryParam("age", 25)

// Boolean parameter → OpenAPI type: "boolean"
.WithQueryParam("active", true)

// UUID → OpenAPI type: "string", format: "uuid"
.WithPathParam("id", uuid.UUID{})

// Time → OpenAPI type: "string", format: "date-time"
.WithQueryParam("createdAt", time.Time{})

// Arrays → OpenAPI type: "array"
.WithQueryParam("tags", []string{})

// Maps → OpenAPI type: "object" with additionalProperties
.WithQueryParam("metadata", map[string]string{})
```

**Supported Types:**
- Primitives: `string`, `int`, `int64`, `float64`, `bool`
- Standard library: `uuid.UUID`, `time.Time`, `net.IP`, `url.URL`
- Collections: `[]T` (arrays), `map[string]T` (objects)
- Structs: Referenced as `#/components/schemas/TypeName`

### Generate Spec
```go
router.GET("/openapi.json", func(c mux.RouteContext) {
    spec := router.OpenAPI(&mux.OpenAPIOptions{
        Title:       "My API",
        Version:     "1.0.0",
        Description: "API description",
    })
    c.OK(spec)
})
```

### Tag Routes
```go
api := router.NewRouteGroup("/api")
api.WithTags("API v1")
```

---

## ⚙️ Router Options

```go
router := mux.NewRouter(
    mux.WithContextPooling(),         // Enable context pooling
    mux.WithHeadFallbackToGet(),      // AUTO handle HEAD requests
    mux.WithMaxBodyBytes(10<<20),     // Set max body size (10MB)
)
```

---

## 🏥 Health Check Pattern

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

## 📦 CRUD Pattern

```go
users := router.NewRouteGroup("/users")
users.GET("/", listUsers)           // List all
users.POST("/", createUser)         // Create
users.GET("/{id}", getUser)         // Get one
users.PUT("/{id}", updateUser)      // Update
users.DELETE("/{id}", deleteUser)   // Delete
```

---

## 🔄 Common Patterns

### Validation
```go
func createUser(c mux.RouteContext) {
    var user User
    if err := c.Bind(&user); err != nil {
        c.BadRequest(err.Error())
        return
    }
    
    if err := validate(user); err != nil {
        c.BadRequest(err.Error())
        return
    }
    
    c.Created(user)
}
```

### Pagination
```go
func listUsers(c mux.RouteContext) {
    page := c.QueryDefault("page", "1")
    limit := c.QueryDefault("limit", "10")
    
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
        c.BadRequest("No file uploaded")
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

## 🐛 Debugging Tips

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

## 🚀 Performance Tips

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

## 📚 Quick Links

- [Full Documentation](../README.md)
- [Learning Path](learning-path.md)
- [Examples](../examples/)
- [API Reference](https://pkg.go.dev/github.com/fgrzl/mux)

---

## 💡 Remember

- **One response per request** - Don't call `c.OK()` multiple times
- **Always validate input** - Use `c.Bind()` for type safety
- **Handle errors gracefully** - Use appropriate status codes
- **Document your API** - Use OpenAPI annotations
- **Test your endpoints** - Write tests for critical paths

---

**Print this page for quick reference while coding!** 📄

## See Also

- [Quick Start](quick-start.md) - Get running in 5 minutes
- [Getting Started](getting-started.md) - Comprehensive introduction
- [Interactive Tutorial](interactive-tutorial.md) - Build a Todo API
- [Learning Path](learning-path.md) - Structured learning progression
- [Router](router.md) - Routing fundamentals
- [Middleware](middleware.md) - Built-in middleware guide
