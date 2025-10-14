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

- 🚀 **Fast** | 54ns static routes, 205ns param routes, 0 allocations for static routes, 2 allocations for param routes |
- ✨ **OpenAPI Built-in** | Auto-generate OpenAPI 3.1 specs — no codegen, no extra tools |
- 🧩 **Composable** | First-class middleware, route groups, and request binding |
- 💡 **Ergonomic** | Clean DSL with smart helpers — `c.QueryUUID()`, `c.Bind()`, `c.Created()` |
- 🏥 **Production Ready** | Health probes, graceful shutdown, TLS, compression, auth out-of-the-box |
- 🧪 **Type Safe** | Interface-based design for better testing and mocking |

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
|--|-||
| **Get something running now** | ⬆️ Quick Start above | 2 min |
| **Learn by building** | [Interactive Tutorial](docs/interactive-tutorial.md) | 30 min |
| **Understand everything** | [Learning Path](docs/learning-path.md) | 2 hours |
| **Look up syntax** | [Cheat Sheet](docs/cheat-sheet.md) | As needed |

### � Start Here

- [**Getting Started Guide**](docs/getting-started.md) - Zero to first API in 5 minutes, then choose your learning path

### �🎓 Tutorials

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
- **� WebServer**: Production-ready server with graceful shutdown and TLS
- **�🔐 Authentication**: JWT-based auth with role-based access control
- **⚡ Rate Limiting**: Per-route token bucket rate limiting
- **📖 OpenAPI 3.1**: Automatic spec generation with inline documentation
- **🏥 Health Probes**: Built-in Kubernetes-style `/healthz`, `/livez`, `/readyz` endpoints
- **� Geographic Control**: Export control with GeoIP support
- **📊 Observability**: OpenTelemetry integration and structured logging

## Performance

Mux is engineered for hot‑path efficiency. Representative benchmark results from this repo:

- Core router (pooled context):
  - Exact/catch‑all routes: ~26–65 ns/op, 0 allocs/op
  - Param routes: ~128 ns/op, 0 allocs/op
  - 10k routes in registry: ~267 ns/op
- Middleware (per request):
  - Logging: ~1.0 µs, ~12 allocs/op (pooled)
  - OpenTelemetry: ~2.5 µs, ~37 allocs/op (pooled)
  - Compression (gzip/deflate): ~10–74 µs depending on body size, ~12–14 allocs/op (pooled)
- End‑to‑end sample pipeline (router + common middleware): ~1.4 µs, 8 allocs/op (pooled)

Environment: Windows, 12th Gen Intel Core i9‑12900HK. See “Testing” for how to reproduce on your machine — numbers will vary by hardware and OS.

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

## Documentation

Comprehensive documentation to help you build production-ready APIs:

### 🚀 Getting Started

- [**Getting Started Guide**](docs/getting-started.md) - Your first 5 minutes with Mux
- [**Installation Guide**](docs/installation.md) - Setup requirements and installation
- [**Quick Start Tutorial**](docs/quick-start.md) - Build your first API in 10 steps
- [**Examples**](examples/) - Working example applications

### 📖 Core Documentation

- [**Overview**](docs/overview.md) - Architecture and core concepts
- [**Router**](docs/router.md) - Route definition and configuration
- [**WebServer**](docs/webserver.md) - Production server with graceful shutdown
- [**Built-in Middleware Guide**](docs/middleware.md) - Complete middleware reference
- [**Authentication Middleware**](docs/authentication-middleware.md) - JWT auth setup
- [**Custom Middleware**](docs/custom-middleware.md) - Building custom middleware
- [**Health Probes**](docs/health-probes.md) - Kubernetes-style health checks

### 📚 Advanced Topics

- [**Best Practices Guide**](docs/best-practices.md) - Production-ready patterns and conventions
- [**FAQ**](docs/faq.md) - Common questions and troubleshooting

### 🎯 Quick Links

- [API Reference](https://pkg.go.dev/github.com/fgrzl/mux) - Complete API documentation on pkg.go.dev
- [Examples Directory](examples/) - Working code examples
- [GitHub Issues](https://github.com/fgrzl/mux/issues) - Bug reports and feature requests

## OpenAPI Generation

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

### Production Features

Mux is designed to be fast and allocation-free on hot paths.
Latest benchmarks (Go 1.22, i9-12900HK):

- Exact match: ~66 ns/op, 0 allocs
- Param match: ~135 ns/op, 0 allocs
- Catch-all: ~64 ns/op, 0 allocs
- Many routes (10k): ~241 ns/op, 3 allocs
- ServeHTTP (end-to-end): ~1.5 µs, 8 allocs

➡️ Mux is a zero-alloc, sub-150 ns router for common paths, competitive with chi and httprouter.
Wildcard (\*) routes remain slower (~1.4 µs, 8 allocs), but typical REST and SPA routing are elite-tier.

## 🏎️ Performance

Mux delivers competitive performance while providing production features:

**Core Router Benchmarks** (i9-12900HK, Windows):

- Static routes: **70 ns/op**, 0 allocs
- Wildcard routes: **73 ns/op**, 0 allocs
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
