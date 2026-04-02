# Getting Started with Mux

Welcome! This guide will get you from zero to a working API in minutes.

## 📋 Prerequisites

- **Go 1.24.4 or later** - [Download here](https://go.lang.org/dl/)
- **Basic Go knowledge** - Understand functions, structs, and packages
- **A code editor** - VS Code, GoLand, or your favorite editor

## ⚡ Quick Start (5 Minutes)

### 1. Install Mux

```bash
# Create a new project
mkdir my-api
cd my-api
go mod init my-api

# Install Mux
go get github.com/fgrzl/mux
```

### 2. Create Your First API

Create `main.go`:

```go
package main

import (
    "net/http"
    "github.com/fgrzl/mux"
)

func main() {
    router := mux.NewRouter()
    
    router.GET("/", func(c mux.RouteContext) {
        c.OK(map[string]string{
            "message": "Welcome to my API!",
            "status":  "running",
        })
    })
    
    http.ListenAndServe(":8080", router)
}
```

### 3. Run It!

```bash
go run main.go
```

### 4. Test It!

```bash
curl http://localhost:8080/
```

**Expected output:**
```json
{
  "message": "Welcome to my API!",
  "status": "running"
}
```

✅ **Congratulations!** You have a working API!

---

## 🎯 What's Next?

Choose your own adventure:

### Option 1: Learn by Doing (Recommended)
**→ [Interactive Tutorial](interactive-tutorial.md)** - Build a complete Todo API in 30 minutes

This hands-on tutorial will teach you:
- CRUD operations
- JSON handling
- Validation
- Error handling
- OpenAPI documentation

### Option 2: Structured Learning
**→ [Learning Path](learning-path.md)** - Progressive 8-level course

Start at your level:
- **Beginner**: Levels 1-3 (basic routing and parameters)
- **Intermediate**: Levels 4-6 (groups, middleware, OpenAPI)
- **Advanced**: Levels 7-8 (error handling, production)

### Option 3: Copy & Paste
**→ [Cheat Sheet](cheat-sheet.md)** - Quick reference for common patterns

Perfect for experienced developers who just need syntax examples.

### Option 4: See Complete Examples
**→ [Examples Directory](../examples/)** - Working applications

- **hello-world**: Minimal example
- **todo-api**: Full CRUD API with OpenAPI docs

---

## 🧭 Roadmap

Here's a typical learning progression:

```
Day 1: Hello World + Basic Routes (30 min)
  ↓
Day 2: JSON APIs + Path Parameters (1 hour)
  ↓
Day 3: Middleware + Authentication (1 hour)
  ↓
Day 4: OpenAPI Documentation (30 min)
  ↓
Day 5: Production Deployment (1 hour)
```

**Total investment: ~4-5 hours to full proficiency** 🚀

---

## 📖 Core Concepts

Before diving deeper, understand these key concepts:

### 1. Router
The router matches HTTP requests to handlers:

```go
router := mux.NewRouter()
router.GET("/users", listUsers)    // Match GET /users
router.POST("/users", createUser)  // Match POST /users
```

### 2. RouteContext
Every handler receives a `RouteContext` with request data and response helpers:

```go
func myHandler(c mux.RouteContext) {
    // Read request
    name, _ := c.Param("name")
    
    // Send response
    c.OK(map[string]string{"hello": name})
}
```

### 3. Middleware
Middleware runs before handlers to add cross-cutting functionality:

```go
// Add logging to all routes
mux.UseLogging(router)

// Add authentication to specific routes
api := router.NewRouteGroup("/api")
api.Use(authMiddleware)
```

### 4. Route Groups
Organize related routes with shared configuration:

```go
api := router.NewRouteGroup("/api/v1")
api.WithTags("API v1")

users := api.NewRouteGroup("/users")
users.GET("/", listUsers)
users.POST("/", createUser)
// Results in: /api/v1/users
```

### 5. WebServer (Production)
Production-ready server with graceful shutdown and TLS:

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

server.Listen(ctx)  // Graceful shutdown on Ctrl+C
```

**Features**:
- Automatic graceful shutdown
- Production-ready timeouts (10s read/write, 120s idle)
- TLS/HTTPS support
- Context-based lifecycle

---

## 🔑 Common Patterns

### Pattern 1: JSON API

```go
type User struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

router.POST("/users", func(c mux.RouteContext) {
    var user User
    if err := c.Bind(&user); err != nil {
        c.BadRequest("Invalid JSON", err.Error())
        return
    }
    
    c.Created(user)
})
```

### Pattern 2: Path Parameters

```go
router.GET("/users/{id}", func(c mux.RouteContext) {
    id, ok := c.Param("id")
    if !ok {
        c.BadRequest("Missing parameter", "id is required")
        return
    }
    
    user := fetchUser(id)
    c.OK(user)
})
```

### Pattern 3: Query Parameters

```go
router.GET("/search", func(c mux.RouteContext) {
    query := c.Params()["q"]
    limit, _ := c.ParamInt("limit")
    
    results := search(query, limit)
    c.OK(results)
})
```

### Pattern 4: Error Handling

```go
router.GET("/users/{id}", func(c mux.RouteContext) {
    id, _ := c.Param("id")
    
    user, err := fetchUser(id)
    if err == ErrNotFound {
        c.NotFound()
        return
    }
    if err != nil {
        c.InternalServerError("Database error", err.Error())
        return
    }
    
    c.OK(user)
})
```

### Pattern 5: Health Checks (Kubernetes-style)

```go
// Built-in probe endpoints (automatically allow anonymous access)
router.Healthz()  // GET /healthz - simple health check
router.Livez()    // GET /livez - liveness probe
router.Readyz()   // GET /readyz - readiness probe

// With custom checks
router.ReadyzWithCheck(func(c mux.RouteContext) bool {
    // Returns 200 OK if ready, 503 Service Unavailable if not
    return db.Ping() == nil && cache.Ready()
})
```

---

## 🛠️ Development Workflow

### 1. Local Development

```bash
# Run with auto-reload (using air)
go install github.com/cosmtrek/air@latest
air

# Or run manually
go run main.go
```

### 2. Testing

```bash
# Run tests
go test ./...

# With coverage
go test ./... -cover

# Verbose output
go test ./... -v
```

### 3. Building

```bash
# Build binary
go build -o myapi

# Run binary
./myapi

# Build for production (smaller binary)
CGO_ENABLED=0 go build -ldflags="-s -w" -o myapi
```

### 4. Deployment

```bash
# Docker
docker build -t myapi .
docker run -p 8080:8080 myapi

# Or deploy to your favorite platform
# - Heroku, Railway, Fly.io
# - AWS Lambda, Google Cloud Run
# - Kubernetes
```

---

## 🆘 Troubleshooting

### "cannot find package"

**Problem:** Import errors when running code

**Solution:**
```bash
go mod tidy
go get github.com/fgrzl/mux
```

### "404 Not Found" for valid routes

**Problem:** Routes not matching as expected

**Solutions:**
- Check HTTP method matches (GET vs POST)
- Verify path exactly matches (case-sensitive)
- Ensure router is passed to `http.ListenAndServe`

### "Invalid JSON" errors

**Problem:** `c.Bind()` failing

**Solutions:**
- Verify `Content-Type: application/json` header is set
- Check JSON syntax with a validator
- Ensure struct fields are exported (capitalized)

### JSON fields not appearing in response

**Problem:** Struct fields not serializing

**Solutions:**
- Capitalize field names (exported fields only)
- Add JSON tags: `json:"fieldName"`
- Check for `json:"-"` tags that hide fields

---

## 📚 Next Steps

You're ready to build! Here are your best next steps:

### For Hands-On Learners
1. Complete the [Interactive Tutorial](interactive-tutorial.md)
2. Customize the [Todo API example](../examples/todo-api/)
3. Build your own API

### For Systematic Learners
1. Follow the [Learning Path](learning-path.md) from Level 1
2. Read each level's documentation
3. Complete the exercises at each level

### For Quick Reference
1. Bookmark the [Cheat Sheet](cheat-sheet.md)
2. Keep the [API Reference](https://pkg.go.dev/github.com/fgrzl/mux) handy
3. Browse the [examples directory](../examples/)

---

## 🎓 Learning Resources Summary

| Resource | Best For | Time |
|----------|----------|------|
| [Interactive Tutorial](interactive-tutorial.md) | Hands-on learners | 30 min |
| [Learning Path](learning-path.md) | Structured progression | 2 hours |
| [Cheat Sheet](cheat-sheet.md) | Quick reference | 5 min |
| [Hello World Example](../examples/hello-world/) | Verify setup | 5 min |
| [Todo API Example](../examples/todo-api/) | Complete reference | 15 min |

---

## 💬 Get Help

- **📖 Documentation**: Check the [docs](.) directory
- **🐛 Issues**: [GitHub Issues](https://github.com/fgrzl/mux/issues)
- **💡 Examples**: Browse [examples](../examples/)
- **📦 API Reference**: [pkg.go.dev](https://pkg.go.dev/github.com/fgrzl/mux)

---

## ✨ Pro Tips

1. **Use route groups** - Organize routes logically and avoid repetition
2. **Enable logging early** - `mux.UseLogging(router)` helps debugging
3. **Document as you go** - Add OpenAPI metadata while writing handlers
4. **Test incrementally** - Test each endpoint before moving to the next
5. **Read the examples** - They demonstrate best practices

---

**Ready to build something amazing? Let's go! 🚀**

Start with the [Interactive Tutorial →](interactive-tutorial.md)

## See Also

- [Quick Start](quick-start.md) - Get running in 5 minutes
- [Interactive Tutorial](interactive-tutorial.md) - Build a Todo API
- [Learning Path](learning-path.md) - Structured learning progression
- [Cheat Sheet](cheat-sheet.md) - Quick reference guide
- [Router](router.md) - Routing fundamentals
- [Middleware](middleware.md) - Built-in middleware guide
