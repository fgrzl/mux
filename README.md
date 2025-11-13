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
|---------|--------------|
| 🚀 **Fast** | 45ns static routes, 124ns param routes, 0 allocations with pooling |
| ✨ **OpenAPI Built-in** | Auto-generate OpenAPI 3.1 specs — no codegen, no extra tools |
| 🧩 **Composable** | First-class middleware, route groups, and request binding |
| 💡 **Ergonomic** | Clean DSL with smart helpers — `c.QueryUUID()`, `c.Bind()`, `c.Created()` |
| 🏥 **Production Ready** | Health probes, graceful shutdown, TLS, compression, auth out-of-the-box |
| 🧪 **Type Safe** | Interface-based design for better testing and mocking |

## 📖 Complete Example

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

    // Add basic middleware
    mux.UseLogging(router)
    mux.UseCompression(router)
    mux.UseEnforceHTTPS(router)

    // Create API routes
    api := router.NewRouteGroup("/api/v1")
    api.WithTags("API v1")

    // User routes with OpenAPI documentation
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

    // Use WebServer for production-ready startup
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

## 🎓 Learning Resources

| I Want To... | Resource | Time |
|--------------|----------|------|
| **Get something running now** | ⬆️ Quick Start above | 2 min |
| **Learn by building** | [Interactive Tutorial](docs/interactive-tutorial.md) | 30 min |
| **Understand everything** | [Learning Path](docs/learning-path.md) | 2 hours |
| **Look up syntax** | [Cheat Sheet](docs/cheat-sheet.md) | As needed |

### 🎯 Start Here

- [**Getting Started Guide**](docs/getting-started.md) - Zero to first API in 5 minutes, then choose your learning path

### 🎓 Tutorials

- [**Interactive Tutorial**](docs/interactive-tutorial.md) - Build a complete Todo API in 30 minutes with step-by-step guidance
- [**Learning Path**](docs/learning-path.md) - Progressive 8-level path from "Hello World" to production-ready apps
- [**Cheat Sheet**](docs/cheat-sheet.md) - Quick reference for common patterns and code snippets

### 💡 Examples

- [**Hello World**](examples/hello-world/) - Minimal example to verify your setup
- [**Todo API**](examples/todo-api/) - Complete REST API demonstrating CRUD operations, validation, and OpenAPI docs

## ✨ Key Features

| Goal                  | Resource                                             | Time      |
| --------------------- | ---------------------------------------------------- | --------- |
| Get something running | [Getting Started](docs/getting-started.md)           | 5 min     |
| Build a complete API  | [Interactive Tutorial](docs/interactive-tutorial.md) | 30 min    |
| Full proficiency      | [Learning Path](docs/learning-path.md)               | 2 hours   |
| Quick syntax lookup   | [Cheat Sheet](docs/cheat-sheet.md)                   | As needed |

**Total time from zero to production-ready: ~2-3 hours** 🚀

## Key Features

- **🛣️ Route Management**: Flexible patterns, parameter binding, route groups
- **🔧 Middleware System**: Built-in and custom middleware support
- **📝 Request Binding**: Automatic data collection from multiple sources
- **📤 Response Helpers**: Structured responses for common HTTP status codes
- **🖥️ WebServer**: Production-ready server with graceful shutdown and TLS
- **🔐 Authentication**: JWT-based auth with role-based access control
- **⚡ Rate Limiting**: Per-route token bucket rate limiting
- **📖 OpenAPI 3.1**: Automatic spec generation with inline documentation
- **🏥 Health Probes**: Built-in Kubernetes-style `/healthz`, `/livez`, `/readyz` endpoints
- **🌍 Geographic Control**: Export control with GeoIP support
- **📊 Observability**: OpenTelemetry integration and structured logging

## Performance

Mux is engineered for hot‑path efficiency with **zero allocations on optimized paths**. Representative benchmark results from this repo:

**Core Router (with context pooling enabled):**
- Static routes: **45 ns/op**, 0 allocs/op
- Single param routes: **124 ns/op**, 0 allocs/op
- Multi param routes: **157 ns/op**, 0 allocs/op
- Wildcard routes: **76 ns/op**, 0 allocs/op
- Deep paths (5+ segments): **217 ns/op**, 0 allocs/op

**End-to-End ServeHTTP (router + handler):**
- Catch-all with pooling: **67 ns/op**, 0 allocs/op
- Single param with pooling: **111 ns/op**, 0 allocs/op
- Full middleware pipeline: **1.4 µs**, 8 allocs/op

**Middleware Overhead (per request, pooled):**
- Logging: ~1.0 µs, ~12 allocs/op
- OpenTelemetry: ~2.5 µs, ~37 allocs/op
- Compression (gzip/deflate): ~10-74 µs depending on body size

Environment: Windows, 12th Gen Intel Core i9‑12900HK. See "Testing" for how to reproduce on your machine — numbers will vary by hardware and OS.

## Basic Usage

### Route Definition

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

    // Access individual parameters
    orgID, _ := c.QueryUUID("org_id")
    includeDetails, _ := c.QueryBool("include_details")

    // Create user and respond
    createdUser := service.CreateUser(user)
    c.Created(createdUser)
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

---

## � OpenAPI Generation

Generate OpenAPI specifications from your routes with automatic type inference:

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

## 🏎️ Performance

Mux delivers **elite-tier performance** with zero allocations on optimized paths while providing production-ready features:

**Core Router Benchmarks** (i9-12900HK, Windows, pooled contexts):

- Static routes: **45 ns/op**, 0 allocs
- Single param: **124 ns/op**, 0 allocs
- Multi param: **157 ns/op**, 0 allocs
- Wildcard routes: **76 ns/op**, 0 allocs
- Deep path (5+ segments): **217 ns/op**, 0 allocs

**End-to-End ServeHTTP** (router + handler):

- Catch-all: **67 ns/op**, 0 allocs
- Single param: **111 ns/op**, 0 allocs
- Full middleware pipeline: **1.4 µs**, 8 allocs

**Middleware Overhead** (per request, pooled):

- Logging: ~1.0 µs, ~12 allocs
- OpenTelemetry: ~2.5 µs, ~37 allocs
- Compression: ~10-74 µs (depends on body size)

**What This Means:**

- ✅ **Zero allocations** for all routing patterns with pooling enabled
- ✅ **Sub-150ns** routing for common REST/SPA patterns
- ✅ **3-5x slower** than ultra-minimalist routers (Gin/Echo) but provides rich features
- ✅ **2-3x faster** than Chi and **10x faster** than Gorilla Mux
- ✅ Production-ready with OpenAPI generation, type-safe binding, and middleware ecosystem

<details>
<summary><b>Click to see detailed benchmark comparison</b></summary>

| Router          | Static Route | Single Param | Multi Param | Deep Path | Wildcard |
| --------------- | ------------ | ------------ | ----------- | --------- | -------- |
| **Mux**         | **45ns** ✅   | **124ns** ✅  | **157ns** ✅ | **217ns** | **76ns** |
| HttpRouter ⚡   | 22ns         | 42ns         | 57ns        | 62ns      | 44ns     |
| Echo ⚡         | 33ns         | 43ns         | 64ns        | 70ns      | 41ns     |
| Gin ⚡          | 34ns         | 37ns         | 49ns        | 54ns      | 38ns     |
| Chi             | 171ns        | 302ns        | 355ns       | 360ns     | 313ns    |
| Gorilla Mux     | 452ns        | 723ns        | 1196ns      | 1237ns    | 705ns    |

**Performance Position:**

- 🥇 **Mux is 2-3x faster than Chi** and **10x faster than Gorilla Mux**
- 🥈 **Mux is 3-5x slower than ultra-minimalist routers** (HttpRouter/Echo/Gin)
- ✅ **Zero allocations** with context pooling enabled (vs 1-7 allocs for competitors)
- 🎯 **The ~100ns "feature cost"** buys you:
  - Automatic OpenAPI 3.1 generation
  - Rich middleware ecosystem (auth, logging, compression, telemetry)
  - Interface-based handlers (testable, mockable)
  - Type-safe request binding (`c.QueryUUID()`, `c.ParamInt()`)
  - Production-ready defaults

**The 100ns difference is negligible** in real applications where business logic (database queries, external APIs, processing) takes microseconds to milliseconds.

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
