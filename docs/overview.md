# Mux Library Overview

Mux is a lightweight, modular HTTP router for Go designed for building modern APIs with built-in support for middleware, request binding, OpenAPI 3.1 generation, and flexible authentication.

## Architecture

### Core Components

- **Router**: The main entry point that handles HTTP routing and middleware execution
- **RouteGroup**: Groups routes with shared configuration and route defaults
- **RouteContext**: Provides context for handling HTTP requests with type-safe parameter access
- **Middleware**: Modular components for cross-cutting concerns

### Request Flow

1. HTTP request arrives at the Router
2. Router looks up the matching route and parameters
3. Middleware pipeline executes in order
4. Handler function processes the request
5. Response is sent back through the middleware pipeline

## Key Features

### Type-Safe Parameter Binding

Mux provides type-safe helpers for accessing request data:

```go
// Query parameters
userID, ok := c.Query().UUID("user_id")
page, _ := c.Query().Int("page")
tags, _ := c.Query().Strings("tags")

// Path parameters  
resourceID, ok := c.ParamUUID("id")

// Form data
name, ok := c.Form().String("name")
age, _ := c.Form().Int("age")
```

### Automatic Request Binding

The `Bind()` method automatically collects data according to the HTTP method:

- `POST`, `PUT`, and `PATCH`: query params, path params, headers, and supported request bodies
- `GET`, `HEAD`, `DELETE`, and other methods without body binding: query params, path params, and headers only

Top-level JSON arrays bind directly into slice targets. That bind is body-only: if the same request also carries query params, path params, or declared header params that you need, read them from the context separately instead of mixing them into the same `Bind()` call.

Declared object query parameters also support dot notation and bracket notation for nested fields, such as `user.address.city=Paris` or `user[address][city]=Paris`. When a JSON body and query binding populate different nested fields on the same object, `Bind()` merges those subtrees; when both sources set the same leaf, the later body value wins.

```go
type User struct {
    ID    uuid.UUID `json:"id"`
    Name  string    `json:"name"`
    Email string    `json:"email"`
}

func updateUser(c mux.RouteContext) {
    var user User
    if err := c.Bind(&user); err != nil {
        c.BadRequest("Invalid request", err.Error())
        return
    }
    // user struct is populated from path params, query params, headers,
    // and the request body when the HTTP method allows one.
    c.OK(user)
}
```

### Structured Responses

Built-in response helpers for common HTTP status codes:

```go
c.OK(data)                    // 200 OK
c.Created(resource)           // 201 Created
c.BadRequest(title, detail)   // 400 Bad Request
c.Unauthorized()              // 401 Unauthorized
c.NotFound()                  // 404 Not Found
c.ServerError(title, detail)  // 500 Internal Server Error
```

### OpenAPI 3.1 Generation

Define your API documentation alongside your routes:

```go
router.POST("/users", createUser).
    OperationID("createUser").
    Summary("Create a new user").
    AcceptJSON(User{}).
    Created(User{}).
    Responds(http.StatusBadRequest, mux.ProblemDetails{}).
    Tags("Users")
```

## Middleware System

Mux includes built-in middleware for common needs:

- **Authentication**: JWT token validation and creation
- **Authorization**: Role-based access control with permissions  
- **Compression**: Gzip/deflate response compression
- **Logging**: Structured request/response logging
- **Rate Limiting**: Token bucket rate limiting per route
- **HTTPS Enforcement**: Automatic HTTP to HTTPS redirects
- **Export Control**: Geographic access restrictions
- **OpenTelemetry**: Distributed tracing and metrics
- **Method-aware routing**: Returns 405 Method Not Allowed with an "Allow" header when a path exists but the method is not permitted
- **Optional HEAD fallback**: Enable serving HEAD via GET handler (headers/status only) when no HEAD route is defined
- **Configurable body size**: Limit request body size used by Bind with `WithMaxBodyBytes`

Custom middleware is easy to implement:

```go
type CustomMiddleware struct{}

func (m *CustomMiddleware) Invoke(c mux.MutableRouteContext, next mux.HandlerFunc) {
    // Pre-processing
    start := time.Now()
    
    next(c) // Call next middleware/handler
    
    // Post-processing  
    duration := time.Since(start)
    log.Printf("Request took %v", duration)
}
```

## Route Groups

Organize routes with shared configuration:

```go
api := router.Group("/api/v1")
api.Tags("API v1")
api.RequireRoles("user")

users := api.Group("/users")
users.Tags("Users")

users.GET("/", listUsers)
users.POST("/", createUser)
users.GET("/{id}", getUser)
```

## Error Handling

Mux uses structured error responses following RFC 7807 (Problem Details):

```go
type ProblemDetails struct {
    Type     string `json:"type,omitempty"`
    Title    string `json:"title"`
    Status   int    `json:"status"`
    Detail   string `json:"detail,omitempty"`
    Instance string `json:"instance,omitempty"`
}
```

Built-in response helpers automatically create properly structured error responses.

## Scoped Services

Register services on the router, a route group, or an individual route when middleware and handlers need shared collaborators:

```go
router.Services().
    Register(mux.ServiceKey("db"), database).
    Register(mux.ServiceKey("logger"), logger)

admin := router.Group("/admin")
admin.Services().Register(mux.ServiceKey("auditSink"), auditSink)

func handler(c mux.RouteContext) {
    db, _ := c.Services().Get(mux.ServiceKey("db"))
    logger, _ := c.Services().Get(mux.ServiceKey("logger"))
    // Use services...
}
```

Child groups and route builders can override root registrations when a narrower scope needs different behavior.

## Performance Considerations

- Minimal overhead with efficient route matching
- Middleware pipeline optimized for speed
- Built-in response compression
- Concurrent-safe design
- Memory-efficient parameter parsing

## Best Practices

1. **Group related routes** using RouteGroup for better organization
2. **Use type-safe parameter helpers** instead of manual string parsing
3. **Implement proper error handling** with structured responses
4. **Add OpenAPI documentation** as you build routes
5. **Use middleware sparingly** - only add what you need
6. **Test your middleware** and handlers thoroughly
7. **Follow Go conventions** for naming and structure

## See Also

- [Getting Started](getting-started.md) - Comprehensive introduction
- [Quick Start](quick-start.md) - Get running in 5 minutes
- [Router](router.md) - Routing fundamentals
- [Middleware](middleware.md) - Built-in middleware guide
- [Best Practices](best-practices.md) - Detailed patterns and conventions


