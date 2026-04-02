# Learning Path for Mux

**Making Mux the easiest-to-learn Go router** 🚀

This document outlines a progressive learning path from "Hello World" to production-ready applications.

---

## 📚 Learning Levels

### Level 1: Your First API (5 minutes)
**Goal**: Get something running immediately

```go
package main

import (
    "net/http"
    "github.com/fgrzl/mux"
)

func main() {
    router := mux.NewRouter()
    
    router.GET("/", func(c mux.RouteContext) {
        c.OK("Welcome to Mux!")
    })
    
    http.ListenAndServe(":8080", router)
}
```

**You just learned**:
- ✅ How to create a router
- ✅ How to add a GET endpoint
- ✅ How to send a response

**Next**: [Level 2 - Working with Data](#level-2-working-with-data-10-minutes)

---

### Level 2: Working with Data (10 minutes)
**Goal**: Handle JSON requests and responses

```go
type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

func main() {
    router := mux.NewRouter()
    
    // GET - Return JSON
    router.GET("/user", func(c mux.RouteContext) {
        user := User{Name: "John", Email: "john@example.com"}
        c.OK(user) // Automatically converts to JSON
    })
    
    // POST - Receive JSON
    router.POST("/user", func(c mux.RouteContext) {
        var user User
        if err := c.Bind(&user); err != nil {
            c.BadRequest(err.Error())
            return
        }
        c.Created(user)
    })
    
    http.ListenAndServe(":8080", router)
}
```

**Test it**:
```bash
# GET
curl http://localhost:8080/user

# POST
curl -X POST http://localhost:8080/user \
  -H "Content-Type: application/json" \
  -d '{"name":"Jane","email":"jane@example.com"}'
```

**You just learned**:
- ✅ How to return JSON with `c.OK(data)`
- ✅ How to parse JSON with `c.Bind(&model)`
- ✅ How to return different status codes (`Created`, `BadRequest`)

**Next**: [Level 3 - URL Parameters](#level-3-url-parameters-10-minutes)

---

### Level 3: URL Parameters (10 minutes)
**Goal**: Work with dynamic routes

```go
func main() {
    router := mux.NewRouter()
    
    // Path parameters: /users/123
    router.GET("/users/{id}", func(c mux.RouteContext) {
        id := c.Param("id")
        c.OK(map[string]string{"userId": id})
    })
    
    // Query parameters: /search?q=golang
    router.GET("/search", func(c mux.RouteContext) {
        query, ok := c.Query("q")
        if !ok {
            c.BadRequest("Missing 'q' parameter")
            return
        }
        c.OK(map[string]string{"searching": query})
    })
    
    http.ListenAndServe(":8080", router)
}
```

**Test it**:
```bash
curl http://localhost:8080/users/123
curl http://localhost:8080/search?q=golang
```

**You just learned**:
- ✅ Path parameters with `{name}` syntax
- ✅ Reading parameters with `c.Param("name")`
- ✅ Query parameters with `c.Query("name")`

**Next**: [Level 4 - Organizing Routes](#level-4-organizing-routes-15-minutes)

---

### Level 4: Organizing Routes (15 minutes)
**Goal**: Structure your API with route groups

```go
func main() {
    router := mux.NewRouter()
    
    // API version 1
    v1 := router.NewRouteGroup("/api/v1")
    
    // Users endpoints
    users := v1.NewRouteGroup("/users")
    users.GET("/", listUsers)
    users.POST("/", createUser)
    users.GET("/{id}", getUser)
    users.PUT("/{id}", updateUser)
    users.DELETE("/{id}", deleteUser)
    
    // Posts endpoints
    posts := v1.NewRouteGroup("/posts")
    posts.GET("/", listPosts)
    posts.POST("/", createPost)
    
    http.ListenAndServe(":8080", router)
}

func listUsers(c mux.RouteContext) {
    c.OK([]string{"user1", "user2"})
}

func createUser(c mux.RouteContext) {
    var user User
    if err := c.Bind(&user); err != nil {
        c.BadRequest(err.Error())
        return
    }
    c.Created(user)
}

// ... other handlers
```

**You just learned**:
- ✅ Route groups with `NewRouteGroup("/prefix")`
- ✅ Nested route groups for organization
- ✅ Extracting handlers to separate functions

**Next**: [Level 5 - Middleware](#level-5-middleware-20-minutes)

---

### Level 5: Middleware (20 minutes)
**Goal**: Add cross-cutting concerns

```go
func main() {
    router := mux.NewRouter()
    
    // Global middleware (applies to all routes)
    mux.UseLogging(router)      // Log all requests
    mux.UseCompression(router)  // Compress responses
    
    // Public routes (no auth needed)
    router.GET("/", func(c mux.RouteContext) {
        c.OK("Public homepage")
    })
    
    // Protected routes (require authentication)
    api := router.NewRouteGroup("/api")
    
    // Add auth middleware to this group only
    mux.UseAuthentication(api, &mux.AuthenticationOptions{
        Scheme: "Bearer",
        ValidateToken: func(token string) (any, error) {
            if token == "secret-token" {
                return "user123", nil
            }
            return nil, errors.New("invalid token")
        },
    })
    
    api.GET("/profile", func(c mux.RouteContext) {
        // Get the authenticated user
        user := c.Principal()
        c.OK(map[string]any{"user": user})
    })
    
    http.ListenAndServe(":8080", router)
}
```

**Test it**:
```bash
# Public route (works)
curl http://localhost:8080/

# Protected route (fails)
curl http://localhost:8080/api/profile

# Protected route (works with auth)
curl http://localhost:8080/api/profile \
  -H "Authorization: Bearer secret-token"
```

**You just learned**:
- ✅ Global middleware with `UseMiddleware(router, ...)`
- ✅ Group-specific middleware
- ✅ Built-in authentication middleware
- ✅ Accessing authenticated user with `c.Principal()`

**Next**: [Level 6 - OpenAPI Documentation](#level-6-openapi-documentation-20-minutes)

---

### Level 6: OpenAPI Documentation (20 minutes)
**Goal**: Auto-generate API documentation

```go
type User struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func main() {
    router := mux.NewRouter()
    
    api := router.NewRouteGroup("/api/v1")
    api.WithTags("API v1")
    
    users := api.NewRouteGroup("/users")
    users.WithTags("Users")
    
    // Document each endpoint
    users.GET("/", listUsers).
        WithOperationID("listUsers").
        WithSummary("List all users").
        WithDescription("Returns a list of all registered users").
        WithOKResponse([]User{})
    
    users.GET("/{id}", getUser).
        WithOperationID("getUser").
        WithSummary("Get user by ID").
        WithPathParam("id", "The unique identifier of the user", "user-123").
        WithOKResponse(User{})
    
    users.POST("/", createUser).
        WithOperationID("createUser").
        WithSummary("Create a new user").
        WithJsonBody(User{}).
        WithCreatedResponse(User{})
    
    // Serve OpenAPI spec
    router.GET("/openapi.json", func(c mux.RouteContext) {
        spec := router.OpenAPI(&mux.OpenAPIOptions{
            Title:       "My API",
            Version:     "1.0.0",
            Description: "A simple user management API",
        })
        c.OK(spec)
    })
    
    http.ListenAndServe(":8080", router)
}
```

**View your docs**:
```bash
# Get OpenAPI spec
curl http://localhost:8080/openapi.json

# Or use Swagger UI
# Copy the JSON and paste into https://editor.swagger.io/
```

**You just learned**:
- ✅ Adding OpenAPI metadata with `.With...()` methods
- ✅ Documenting request/response bodies
- ✅ Documenting parameters
- ✅ Generating OpenAPI 3.1 spec

**Next**: [Level 7 - Error Handling](#level-7-error-handling-15-minutes)

---

### Level 7: Error Handling (15 minutes)
**Goal**: Handle errors consistently

```go
func main() {
    router := mux.NewRouter()
    
    router.GET("/users/{id}", func(c mux.RouteContext) {
        id := c.Param("id")
        
        // Validate input
        if id == "" {
            c.BadRequest("User ID is required")
            return
        }
        
        // Simulate database lookup
        user, err := getUserFromDB(id)
        if err != nil {
            if err == ErrNotFound {
                c.NotFound()
                return
            }
            c.ServerError("Failed to fetch user", err.Error())
            return
        }
        
        c.OK(user)
    })
    
    http.ListenAndServe(":8080", router)
}

var ErrNotFound = errors.New("user not found")

func getUserFromDB(id string) (*User, error) {
    if id == "999" {
        return nil, ErrNotFound
    }
    return &User{ID: id, Name: "John"}, nil
}
```

**Available error responses**:
- `c.BadRequest(message)` - 400
- `c.Unauthorized()` - 401
- `c.Forbidden()` - 403
- `c.NotFound()` - 404
- `c.Conflict(message)` - 409
- `c.ServerError(title, detail)` - 500

**You just learned**:
- ✅ Built-in error response methods
- ✅ Validation patterns
- ✅ Error propagation

**Next**: [Level 8 - Production Ready](#level-8-production-ready-30-minutes)

---

### Level 8: Production Ready (30 minutes)
**Goal**: Add all the production essentials

```go
package main

import (
    "context"
    "net/http"
    "os"
    "os/signal"
    "time"
    "github.com/fgrzl/mux"
)

func main() {
    // Create router with production settings
    router := mux.NewRouter(
        mux.WithContextPooling(),     // Reduce allocations
        mux.WithHeadFallbackToGet(),  // Auto-handle HEAD requests
        mux.WithMaxBodyBytes(10<<20), // 10MB limit
    )
    
    // Production middleware stack
    mux.UseLogging(router)
    mux.UseCompression(router)
    mux.UseCORS(router, &mux.CORSOptions{
        AllowedOrigins: []string{"https://yourdomain.com"},
        AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
    })
    mux.UseRateLimit(router, &mux.RateLimitOptions{
        RequestsPerMinute: 100,
    })
    
    // Built-in Kubernetes-style health probes
    router.Healthz()  // GET /healthz
    router.Livez()    // GET /livez  
    router.Readyz()   // GET /readyz
    
    // Or with custom health checks
    router.ReadyzWithCheck(func(c mux.RouteContext) bool {
        // Check database, cache, migrations, etc.
        return db.Ping() == nil
    })
    
    // Your API routes
    setupRoutes(router)
    
    // WebServer provides graceful shutdown out of the box
    server := mux.NewServer(":8080", router)
    
    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
    defer cancel()
    
    // Listen blocks until shutdown
    if err := server.Listen(ctx); err != nil {
        panic(err)
    }
}

func setupRoutes(router *mux.Router) {
    api := router.NewRouteGroup("/api/v1")
    // ... your routes
}
```

**You just learned**:
- ✅ Production router configuration
- ✅ Full middleware stack
- ✅ WebServer with graceful shutdown
- ✅ Health probes for Kubernetes
- ✅ Context-based lifecycle management

**Alternative (Manual http.Server)**:
```go
func main() {
    router := mux.NewRouter()
    setupRoutes(router)
    
    // Manual configuration
    server := &http.Server{
        Addr:         ":8080",
        Handler:      router,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }
    
    go func() {
        if err := server.ListenAndServe(); err != http.ErrServerClosed {
            panic(err)
        }
    }()
    
    // Wait for interrupt
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt)
    <-quit
    
    // Graceful shutdown
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := server.Shutdown(ctx); err != nil {
        panic(err)
    }
}

func setupRoutes(router *mux.Router) {
    api := router.NewRouteGroup("/api/v1")
    // ... your routes
}
```

**You just learned**:
- ✅ Production router configuration
- ✅ Full middleware stack
- ✅ Graceful shutdown
- ✅ Timeouts and limits
- ✅ Health checks

---

## 🎯 Common Patterns

### Pattern 1: CRUD Resource

```go
type UserResource struct {
    // Could contain DB connection, services, etc.
}

func (ur *UserResource) Register(group *mux.RouteGroup) {
    group.GET("/", ur.List)
    group.POST("/", ur.Create)
    group.GET("/{id}", ur.Get)
    group.PUT("/{id}", ur.Update)
    group.DELETE("/{id}", ur.Delete)
}

func (ur *UserResource) List(c mux.RouteContext) {
    // Implementation
}

// Use it:
users := &UserResource{}
users.Register(router.NewRouteGroup("/users"))
```

### Pattern 2: Validation Helper

```go
func validateUser(user *User) error {
    if user.Email == "" {
        return errors.New("email is required")
    }
    if !strings.Contains(user.Email, "@") {
        return errors.New("invalid email format")
    }
    return nil
}

func createUser(c mux.RouteContext) {
    var user User
    if err := c.Bind(&user); err != nil {
        c.BadRequest(err.Error())
        return
    }
    
    if err := validateUser(&user); err != nil {
        c.BadRequest(err.Error())
        return
    }
    
    c.Created(user)
}
```

### Pattern 3: Database Integration

```go
type Handler struct {
    db *sql.DB
}

func (h *Handler) GetUser(c mux.RouteContext) {
    id := c.Param("id")
    
    var user User
    err := h.db.QueryRow("SELECT id, name, email FROM users WHERE id = ?", id).
        Scan(&user.ID, &user.Name, &user.Email)
    
    if err == sql.ErrNoRows {
        c.NotFound()
        return
    }
    if err != nil {
        c.ServerError("Database error", err.Error())
        return
    }
    
    c.OK(user)
}
```

---

## 💡 Tips for Learning

### Start Simple
Don't try to learn everything at once. Follow the levels in order.

### Run Examples
Type out the code yourself - don't just copy/paste. You'll learn faster.

### Check the Docs
Each level links to relevant documentation for deeper understanding.

### Use Auto-Complete
Modern IDEs will show you all available methods on `mux.RouteContext`.

### Common Mistakes to Avoid

❌ **Don't** do this:
```go
router.GET("/users", func(c mux.RouteContext) {
    c.OK("users")
    c.OK("more users") // Second call is ignored!
})
```

✅ **Do** this:
```go
router.GET("/users", func(c mux.RouteContext) {
    c.OK("users") // Only one response per request
})
```

---

## 📖 Further Reading

After completing all levels, explore:

- [Middleware Guide](middleware.md) - Deep dive into middleware
- [Authentication](authentication-middleware.md) - Advanced auth patterns
- [Best Practices](best-practices.md) - Production tips
- [Router Comparison](router-comparison-quick.md) - How Mux compares
- [FAQ](faq.md) - Common questions

---

## 🆘 Getting Help

**Stuck? Have questions?**

1. Check the [FAQ](faq.md)
2. Review the [examples](../examples/)
3. Open a [GitHub Discussion](https://github.com/fgrzl/mux/discussions)
4. Read the [API documentation](https://pkg.go.dev/github.com/fgrzl/mux)

**Remember**: Everyone starts as a beginner. Take it one level at a time! 🚀
