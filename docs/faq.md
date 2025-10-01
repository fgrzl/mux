# Frequently Asked Questions (FAQ)

This document answers common questions about using Mux for building HTTP APIs in Go.

## General Questions

### Q: What is Mux and how is it different from other Go routers?

**A:** Mux is a lightweight, modular HTTP router for Go designed specifically for building modern APIs. Key differences:

- **API-first approach**: Built-in OpenAPI 3.1 generation without code generation
- **Type-safe parameter binding**: Automatic conversion and validation of parameters
- **Structured responses**: Built-in helpers for consistent error responses (RFC 7807)
- **Modular middleware**: Comprehensive set of built-in middleware with functional options
- **Request binding**: Automatic data collection from multiple sources (path, query, body)

### Q: Is Mux production-ready?

**A:** Yes, Mux is designed for production use with:
- Comprehensive test coverage
- Performance optimizations
- Panic recovery
- Memory-efficient design
- Concurrent-safe operations

### Q: What's the minimum Go version requirement?

**A:** Mux requires **Go 1.24.4** or later to take advantage of the latest language features and security updates.

## Installation and Setup

### Q: How do I install Mux?

**A:** Use Go modules to install Mux:
```bash
go get github.com/fgrzl/mux
```

### Q: Can I use Mux with existing Go HTTP code?

**A:** Yes! Mux implements the standard `http.Handler` interface, so it works seamlessly with existing HTTP servers, middleware, and tools:

```go
// Works with standard library
http.ListenAndServe(":8080", router)

// Works with custom servers
server := &http.Server{
    Addr:    ":8080",
    Handler: router,
}
```

### Q: How do I migrate from other routers?

**A:** Migration is typically straightforward:

**From Gorilla Mux:**
```go
// Gorilla Mux
r := mux.NewRouter()
r.HandleFunc("/users/{id}", getUserHandler).Methods(http.MethodGet)

// Mux
router := mux.NewRouter()
router.GET("/users/{id}", getUserHandler)
```

**From Chi:**
```go
// Chi
r := chi.NewRouter()
r.Get("/users/{id}", getUserHandler)

// Mux
router := mux.NewRouter()
router.GET("/users/{id}", getUserHandler)
```

## Routing and Parameters

### Q: How do I handle different HTTP methods for the same path?

**A:** Define separate routes for each method:
```go
router.GET("/users/{id}", getUser)
router.PUT("/users/{id}", updateUser)
router.DELETE("/users/{id}", deleteUser)
```

### Q: Can I use wildcards or regex in routes?

**A:** Mux uses simple path parameters (`{param}`) for maintainability and performance. For complex patterns, handle them in your handler:

```go
router.GET("/files/{path}", func(c mux.RouteContext) {
    path, _ := c.ParamValue("path")
    // Handle complex path logic here
    if matched, _ := regexp.MatchString(`\.jpg$`, path); matched {
        // Handle image files
    }
})
```

### Q: How do I access query parameters?

**A:** Use the type-safe query parameter helpers:
```go
func handler(c mux.RouteContext) {
    // String values
    search, _ := c.QueryValue("q")
    
    // Numeric values
    page, _ := c.QueryInt("page")
    limit, _ := c.QueryInt("limit")
    
    // Boolean values
    includeDeleted, _ := c.QueryBool("include_deleted")
    
    // UUID values
    userID, ok := c.QueryUUID("user_id")
    
    // Multiple values
    tags, _ := c.QueryValues("tags")
}
```

### Q: How do I handle form data?

**A:** Use form value helpers or automatic binding:
```go
func handler(c mux.RouteContext) {
    // Individual form fields
    name, _ := c.FormValue("name")
    age, _ := c.FormInt("age")
    
    // Automatic binding to struct
    var user User
    if err := c.Bind(&user); err != nil {
        c.BadRequest("Invalid data", err.Error())
        return
    }
}
```

## Middleware

### Q: What's the difference between router-level and route-specific middleware?

**A:** Router-level middleware applies to all routes, while route-specific features like rate limiting apply only to specific routes:

```go
// Router-level (applies to all routes)
mux.UseLogging(router)
mux.UseAuthentication(...)

// Route-specific (applies only to this route)
router.GET("/api/search", handler).
    WithRateLimit(100, time.Minute)
```

### Q: In what order should I add middleware?

**A:** Follow this recommended order:
1. **Infrastructure**: `UseForwardedHeaders()`, `UseLogging()`
2. **Security**: `UseEnforceHTTPS()`, `UseExportControl()`  
3. **Application**: `UseCompression()`, `UseOpenTelemetry()`
4. **Authentication**: `UseAuthentication()`, `UseAuthorization()`
5. **Services**: `UseServices()`

### Q: Can I create custom middleware?

**A:** Yes! Implement the `Middleware` interface:
```go
type CustomMiddleware struct{}

func (m *CustomMiddleware) Invoke(c mux.RouteContext, next mux.HandlerFunc) {
    // Pre-processing
    start := time.Now()
    
    next(c) // Always call next
    
    // Post-processing
    duration := time.Since(start)
    log.Printf("Request took %v", duration)
}

// Add to router
router.Use(&CustomMiddleware{})
```

## Request/Response Handling

### Q: How do I handle JSON requests and responses?

**A:** Mux provides automatic JSON handling:
```go
func createUser(c mux.RouteContext) {
    // Automatic JSON binding
    var user User
    if err := c.Bind(&user); err != nil {
        c.BadRequest("Invalid JSON", err.Error())
        return
    }
    
    // Automatic JSON response
    c.Created(user)
}
```

### Q: How do I handle file uploads?

**A:** Access multipart form data through the standard request:
```go
func uploadFile(c mux.RouteContext) {
    file, header, err := c.Request().FormFile("file")
    if err != nil {
        c.BadRequest("No file provided", err.Error())
        return
    }
    defer file.Close()
    
    // Process file...
    c.OK(map[string]string{
        "filename": header.Filename,
        "size":     fmt.Sprintf("%d", header.Size),
    })
}
```

### Q: How do I return different response formats?

**A:** Use appropriate response helpers or set headers manually:
```go
func handler(c mux.RouteContext) {
    accept := c.Request().Header.Get("Accept")
    
    switch {
    case strings.Contains(accept, "application/xml"):
    c.Response().Header().Set("Content-Type", "application/xml")
    c.Response().Write([]byte("<message>Hello</message>"))
    case strings.Contains(accept, "text/plain"):
    c.Response().Header().Set("Content-Type", "text/plain")
    c.Response().Write([]byte("Hello"))
    default:
        c.OK("Hello")
    }
}
```

## Authentication and Security

### Q: How do I implement JWT authentication?

**A:** Use the built-in authentication middleware:
```go
router.UseAuthentication(
    mux.WithValidator(validateToken),
    mux.WithTokenCreator(createToken),
    mux.WithTokenTTL(30 * time.Minute),
)

func validateToken(tokenString string) (claims.Principal, error) {
    // Parse and validate JWT
    // Return principal with user info
}
```

### Q: How do I protect specific routes?

**A:** Use route-level authorization:
```go
router.GET("/admin", adminHandler).RequireRoles("admin")
router.POST("/users", createUser).RequirePermissions("write")
```

### Q: How do I handle CORS?

**A:** Implement custom CORS middleware:
```go
type CORSMiddleware struct{}

func (m *CORSMiddleware) Invoke(c mux.RouteContext, next mux.HandlerFunc) {
    c.Response().Header().Set("Access-Control-Allow-Origin", "*")
    c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
    c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
    
    if c.Request().Method == "OPTIONS" {
    c.Response().WriteHeader(http.StatusOK)
        return
    }
    
    next(c)
}
```

## OpenAPI and Documentation

### Q: How do I generate OpenAPI specifications?

**A:** Document your routes and generate the spec:
```go
// Document routes
router.POST("/users", createUser).
    WithSummary("Create a user").
    WithJsonBody(User{}).
    WithCreatedResponse(User{})

// Generate spec
generator := mux.NewGenerator()
spec := generator.GenerateSpec(router)
spec.MarshalToFile("openapi.yaml")
```

### Q: Can I customize the OpenAPI output?

**A:** Yes, configure the router with API information:
```go
router := mux.NewRouter(
    mux.WithTitle("My API"),
    mux.WithVersion("1.0.0"),
    mux.WithDescription("API description"),
    mux.WithContact("Support", "https://example.com", "support@example.com"),
    mux.WithLicense("MIT", "https://opensource.org/licenses/MIT"),
)
```

### Q: How do I document complex request/response types?

**A:** Use struct tags and provide examples:
```go
type User struct {
    ID    uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
    Name  string    `json:"name" example:"John Doe"`
    Email string    `json:"email" example:"john@example.com"`
}
```

## Performance and Production

### Q: How do I optimize performance?

**A:** Follow these best practices:
- Only add middleware you need
- Use appropriate rate limiting
- Enable compression for large responses
- Use connection pooling for databases
- Profile your application under load

### Q: How do I handle high traffic?

**A:** Consider these strategies:
- Use rate limiting middleware
- Implement caching layers
- Use load balancers
- Monitor with OpenTelemetry
- Scale horizontally

### Q: How do I handle graceful shutdown?

**A:** Implement graceful shutdown with context cancellation:
```go
func main() {
    router := mux.NewRouter()
    // ... setup routes
    
    server := &http.Server{
        Addr:    ":8080",
        Handler: router,
    }
    
    // Handle shutdown gracefully
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    
    go func() {
        <-sigChan
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        server.Shutdown(ctx)
    }()
    
    server.ListenAndServe()
}
```

## Testing

### Q: How do I test my handlers?

**A:** Create test handlers and use HTTP testing tools:
```go
func TestCreateUser(t *testing.T) {
    router := mux.NewRouter()
    router.POST("/users", createUser)
    
    body := `{"name": "John", "email": "john@example.com"}`
    req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    
    rec := httptest.NewRecorder()
    rtr.ServeHTTP(rec, req)
    
    assert.Equal(t, http.StatusCreated, rec.Code)
}
```

### Q: How do I test middleware?

**A:** Test middleware independently:
```go
func TestAuthenticationMiddleware(t *testing.T) {
    router := mux.NewRouter()
    router.UseAuthentication(mux.WithValidator(mockValidator))
    router.GET("/protected", protectedHandler)
    
    // Test with valid token
    req := httptest.NewRequest(http.MethodGet, "/protected", nil)
    req.Header.Set("Authorization", "Bearer valid-token")
    
    rec := httptest.NewRecorder()
    rtr.ServeHTTP(rec, req)
    
    assert.Equal(t, http.StatusOK, rec.Code)
}
```

## Troubleshooting

### Q: My middleware isn't working. What's wrong?

**A:** Check these common issues:
1. Middleware order matters - add them in the right sequence
2. Always call `next(c)` in custom middleware
3. Ensure middleware is added before routes that should use it
4. Check if the middleware is properly implementing the interface

### Q: Parameters aren't being parsed correctly. Why?

**A:** Common issues:
1. Parameter names in routes must match extraction calls
2. Use type-specific helpers (`ParamUUID`, `QueryInt`, etc.)
3. Check for typos in parameter names
4. Ensure the route pattern matches your request URL

### Q: Why am I getting 404 errors?

**A:** Check:
1. Route patterns match request URLs exactly
2. HTTP methods match (GET vs POST, etc.)
3. Route parameters are correctly defined
4. Router is properly configured as the HTTP handler

### Q: How do I debug routing issues?

**A:** Enable request logging and check:
```go
mux.UseLogging(router) // Logs all requests and responses

// Or add custom debug middleware
type DebugMiddleware struct{}

func (m *DebugMiddleware) Invoke(c mux.RouteContext, next mux.HandlerFunc) {
    log.Printf("Request: %s %s", c.Request().Method, c.Request().URL.Path)
    next(c)
}
```

## Getting Help

### Q: Where can I get more help?

**A:** Try these resources:
- **Documentation**: Check the [docs](../docs/) directory
- **Examples**: Review [example applications](../examples/)
- **Issues**: Report bugs on [GitHub Issues](https://github.com/fgrzl/mux/issues)  
- **Source Code**: Read the well-documented source code

### Q: How do I contribute to Mux?

**A:** We welcome contributions! 
1. Fork the repository
2. Create a feature branch
3. Write tests for your changes
4. Ensure all tests pass
5. Submit a pull request

### Q: How do I report a bug?

**A:** When reporting bugs, include:
- Go version (`go version`)
- Mux version
- Minimal code example that reproduces the issue
- Expected vs actual behavior
- Error messages (if any)

## Best Practices Summary

1. **Use type-safe parameter helpers** instead of manual parsing
2. **Add middleware in the correct order** (infrastructure → security → application)
3. **Implement proper error handling** with structured responses
4. **Document your APIs** as you build them with OpenAPI
5. **Test your handlers and middleware** thoroughly
6. **Use route groups** to organize related endpoints
7. **Follow Go conventions** for naming and structure
8. **Monitor your applications** in production