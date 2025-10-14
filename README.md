# Mux

[![CI](https://github.com/fgrzl/mux/actions/workflows/ci.yaml/badge.svg)](https://github.com/fgrzl/mux/actions/workflows/ci.yaml)
[![Dependabot](https://github.com/fgrzl/mux/actions/workflows/dependabot/dependabot-updates/badge.svg)](https://github.com/fgrzl/mux/actions/workflows/dependabot/dependabot-updates)

**Blazing‑fast HTTP router for Go with built-in OpenAPI 3.1 generation.**

Build production-ready REST APIs in minutes. Single import. Two lines. Ship.

## ⚡ Quick Start

```bash
go get github.com/fgrzl/mux
```

```go
package main

import (
    "net/http"
    "github.com/fgrzl/mux"
)

func main() {
    router := mux.NewRouter()
    router.GET("/hello", func(c mux.RouteContext) {
        c.OK("Hello, World!")
    })
    http.ListenAndServe(":8080", router)
}
```

**That's it.** 🎉 Production-ready defaults with elite performance.

## 🌟 Why Mux?

| Feature | What You Get |
||--|
| 🚀 **Fast** | 54ns static routes, 205ns param routes, 0 allocations on hot paths |
| ✨ **OpenAPI Built-in** | Auto-generate OpenAPI 3.1 specs — no codegen, no extra tools |
| 🧩 **Composable** | First-class middleware, route groups, and request binding |
| 💡 **Ergonomic** | Clean DSL with smart helpers — `c.QueryUUID()`, `c.Bind()`, `c.Created()` |
| 🏥 **Production Ready** | Health probes, graceful shutdown, TLS, compression, auth out-of-the-box |
| 🧪 **Type Safe** | Interface-based design for better testing and mocking |

## 📖 Complete Example

<details>
<summary><b>Click to see a full REST API with middleware, OpenAPI docs, and graceful shutdown (100 lines)</b></summary>

```go
package main

import (
    "context"
    "os/signal"
    "syscall"

    "github.com/fgrzl/mux"
    "github.com/google/uuid"
)

type User struct {
    ID    uuid.UUID `json:"id"`
    Name  string    `json:"name"`
    Email string    `json:"email"`
}

func main() {
    // Graceful shutdown context
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    router := mux.NewRouter()

    // Add production middleware
    mux.UseLogging(router)
    mux.UseCompression(router)
    mux.UseEnforceHTTPS(router)

    // Create API v1 routes
    api := router.NewRouteGroup("/api/v1")
    api.WithTags("API v1")

    // User endpoints with OpenAPI documentation
    users := api.NewRouteGroup("/users")
    users.WithTags("Users")

    users.GET("/", listUsers).
        WithOperationID("listUsers").
        WithSummary("List all users").
        WithOKResponse([]User{})

    users.POST("/", createUser).
        WithOperationID("createUser").
        WithSummary("Create a new user").
        WithJsonBody(User{}).
        WithCreatedResponse(User{})

    users.GET("/{id}", getUser).
        WithOperationID("getUser").
        WithSummary("Get a user by ID").
        WithPathParam("id", uuid.Nil).
        WithOKResponse(User{})

    // Start server with graceful shutdown
    server := mux.NewServer(router, ":8080")
    if err := server.Start(ctx); err != nil {
        panic(err)
    }
}

func listUsers(c mux.RouteContext) {
    users := []User{
        {ID: uuid.New(), Name: "John Doe", Email: "john@example.com"},
        {ID: uuid.New(), Name: "Jane Smith", Email: "jane@example.com"},
    }
    c.OK(users)
}

func createUser(c mux.RouteContext) {
    var user User
    if err := c.Bind(&user); err != nil {
        c.BadRequest("Invalid request", err.Error())
        return
    }

    user.ID = uuid.New()
    c.Created(user)
}

func getUser(c mux.RouteContext) {
    userID, ok := c.ParamUUID("id")
    if !ok {
        c.BadRequest("Invalid user ID", "")
        return
    }

    user := User{
        ID:    userID,
        Name:  "John Doe",
        Email: "john@example.com",
    }
    c.OK(user)
}
```

</details>

## 🎓 Learning Resources

| I Want To... | Resource | Time |
|--|-||
| **Get something running now** | ⬆️ Quick Start above | 2 min |
| **Learn by building** | [Interactive Tutorial](docs/interactive-tutorial.md) | 30 min |
| **Understand everything** | [Learning Path](docs/learning-path.md) | 2 hours |
| **Look up syntax** | [Cheat Sheet](docs/cheat-sheet.md) | As needed |

**📂 Working Examples:**

- [Hello World](examples/hello-world/) - Minimal starter
- [Todo API](examples/todo-api/) - Full CRUD with validation
- [WebServer](examples/webserver/) - Production deployment

**Total learning time: ~2-3 hours from zero to production-ready** 🚀

## ✨ Key Features

### Core Routing

```go
router := mux.NewRouter()

// HTTP methods
router.GET("/users", listUsers)
router.POST("/users", createUser)
router.PUT("/users/{id}", updateUser)
router.DELETE("/users/{id}", deleteUser)

// Route groups
api := router.NewRouteGroup("/api/v1")
api.GET("/health", healthCheck)
```

### Smart Request Handling

```go
func createUser(c mux.RouteContext) {
    // Auto-bind JSON/form/query params to struct
    var user User
    if err := c.Bind(&user); err != nil {
        c.BadRequest("Invalid data", err.Error())
        return
    }

    // Type-safe parameter extraction
    orgID, _ := c.QueryUUID("org_id")
    includeDetails, _ := c.QueryBool("include_details")

    // Structured responses
    c.Created(user)  // Returns 201 with JSON body
}
```

### Middleware System

```go
// Built-in production middleware
mux.UseLogging(router)          // Structured logging
mux.UseCompression(router)      // Gzip/Deflate
mux.UseEnforceHTTPS(router)     // Redirect to HTTPS

// Authentication with JWT
mux.UseAuthentication(router,
    mux.WithValidator(validateToken),
    mux.WithTokenCreator(createToken),
)

// Custom middleware
router.Use(&CustomMiddleware{})
```

### OpenAPI Generation

```go
// Routes automatically become OpenAPI operations
router.POST("/users", createUser).
    WithOperationID("createUser").
    WithSummary("Create a user").
    WithJsonBody(User{}).
    WithCreatedResponse(User{}).
    WithTags("Users")

// Types are inferred from examples
router.GET("/users/{id}", getUser).
    WithPathParam("id", uuid.UUID{}).      // → type: string, format: uuid
    WithQueryParam("include", []string{}). // → type: array, items: string
    WithOKResponse(User{})                 // → $ref: #/components/schemas/User

// Generate OpenAPI 3.1 spec
generator := mux.NewGenerator()
spec := generator.GenerateSpec(router)
spec.MarshalToFile("openapi.yaml")
```

**No manual schema definitions needed!** Mux inspects your Go types and generates accurate OpenAPI schemas automatically.

### Production Features

```go
// Kubernetes-style health probes
router.Healthz()  // GET /healthz - basic health
router.Livez()    // GET /livez  - liveness probe
router.Readyz()   // GET /readyz - readiness probe

// WebServer with graceful shutdown
server := mux.NewServer(router, ":8080")
server.Start(ctx)  // Blocks until SIGINT/SIGTERM
```

## 🏎️ Performance

Mux delivers competitive performance while providing production features:

**Core Router Benchmarks** (i9-12900HK, Windows):

- Static routes: **54 ns/op**, 0 allocs
- Wildcard routes: **82 ns/op**, 0 allocs
- Single param: **205 ns/op**, 2 allocs
- Multi param: **279 ns/op**, 2 allocs
- Deep path (5+ segments): **352 ns/op**, 2 allocs

**Middleware Overhead** (per request, pooled):

- Logging: ~1.0 µs, ~12 allocs
- OpenTelemetry: ~2.5 µs, ~37 allocs
- Compression: ~10-74 µs (depends on body size)

**What This Means:**

- ✅ Zero allocations for static and wildcard routes
- ✅ Minimal allocations for param routes (context pooling)
- ✅ Sub-100ns for common routing patterns
- ✅ Competitive with Gin/Echo while providing OpenAPI generation

<details>
<summary><b>Click to see detailed benchmark comparison</b></summary>

| Router      | Static Route | Single Param | Wildcard |
| ----------- | ------------ | ------------ | -------- |
| **Mux**     | **54ns**     | **205ns**    | **82ns** |
| HttpRouter  | 23ns ⚡      | 45ns ⚡      | 46ns ⚡  |
| Echo        | 32ns ⚡      | 43ns ⚡      | 37ns ⚡  |
| Gin         | 36ns ⚡      | 37ns ⚡      | 39ns ⚡  |
| Chi         | 186ns        | 339ns        | 350ns    |
| Gorilla Mux | 480ns        | 748ns        | 657ns    |

**Mux is 1.5-3x slower than ultra-minimalist routers** (HttpRouter/Echo/Gin) **but provides:**

- ✅ Automatic OpenAPI 3.1 generation
- ✅ Rich middleware ecosystem
- ✅ Interface-based handlers (testable, mockable)
- ✅ Type-safe request binding
- ✅ Better developer experience

**The ~150ns "feature cost" is negligible** in real applications where business logic takes microseconds to milliseconds.

</details>

**Run benchmarks yourself:**

```bash
go test ./pkg/router -bench "Comparison" -benchmem -benchtime=5s
```

## 📚 Documentation

### 🚀 Getting Started

- [**Getting Started Guide**](docs/getting-started.md) - Your first 5 minutes with Mux
- [**Installation Guide**](docs/installation.md) - Setup and requirements
- [**Quick Start Tutorial**](docs/quick-start.md) - Build your first API in 10 steps
- [**Examples Directory**](examples/) - Working example applications

### 📖 Core Documentation

- [**Overview**](docs/overview.md) - Architecture and core concepts
- [**Router**](docs/router.md) - Route definition and configuration
- [**WebServer**](docs/webserver.md) - Production server with graceful shutdown
- [**Middleware**](docs/middleware.md) - Complete middleware reference
- [**Authentication**](docs/authentication-middleware.md) - JWT auth setup
- [**Custom Middleware**](docs/custom-middleware.md) - Building custom middleware
- [**Health Probes**](docs/health-probes.md) - Kubernetes-style health checks

### 📚 Advanced Topics

- [**Best Practices**](docs/best-practices.md) - Production patterns and conventions
- [**FAQ**](docs/faq.md) - Common questions and troubleshooting

### 🎯 Quick Links

- [API Reference](https://pkg.go.dev/github.com/fgrzl/mux) - Complete API documentation
- [Examples](examples/) - Working code examples
- [GitHub Issues](https://github.com/fgrzl/mux/issues) - Bug reports and feature requests

## 🧪 Testing

Run the full test suite:

```bash
go test ./... -v
go test ./... -coverprofile=coverage.out
```

Run performance benchmarks:

```bash
go test ./pkg/router -bench . -benchmem
go test ./pkg/middleware/logging -bench . -benchmem
go test ./pkg/middleware/opentelemetry -bench . -benchmem
go test ./pkg/middleware/compression -bench . -benchmem
```

## 🤝 Contributing

We welcome contributions!

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Write tests for your changes
4. Ensure all tests pass (`go test ./...`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request
