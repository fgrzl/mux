# Mux

[![CI](https://github.com/fgrzl/mux/actions/workflows/ci.yaml/badge.svg)](https://github.com/fgrzl/mux/actions/workflows/ci.yaml)
[![Dependabot](https://github.com/fgrzl/mux/actions/workflows/dependabot/dependabot-updates/badge.svg)](https://github.com/fgrzl/mux/actions/workflows/dependabot/dependabot-updates)

**Mux** is an OpenAPI‑native HTTP framework for Go.

## Is Mux a fit for you?

Use **Mux** if you want:

- **Schema-driven APIs** where routes, validation, and OpenAPI stay in sync
- **Explicit behavior** (middleware order, errors, lifecycle) you can reason about and test
- **Production defaults** like structured errors and graceful shutdown without extra glue

## Install

```bash
go get github.com/fgrzl/mux
```

## Quick start

```go
package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"

    "github.com/fgrzl/mux"
)

func main() {
    router := mux.NewRouter().Safe()
    router.GET("/hello", func(c mux.RouteContext) {
        c.OK("Hello, World!")
    })

    if err := router.Err(); err != nil {
        panic(err)
    }

    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    server := mux.NewServer(":8080", router)
    if err := server.Listen(ctx); err != nil {
        panic(err)
    }
}
```

If you want to keep standard-library handlers while adopting Mux incrementally, use `Handle` or `HandleFunc`:

```go
router.HandleFunc(http.MethodGet, "/healthz", func(w http.ResponseWriter, r *http.Request) {
    routeCtx, ok := mux.RouteContextFromRequest(r)
    if ok {
        if traceID := r.Context().Value("traceID"); traceID != nil {
            _ = traceID
        }
        _, _ = routeCtx.GetService("db")
    }
    w.WriteHeader(http.StatusNoContent)
})
```

## What you get

- **One model** for routing, binding, validation, and OpenAPI generation
- **Deterministic middleware** order and structured error responses
- **Scoped services** via `Services()` registries on routers, groups, and routes
- **Server lifecycle** helpers (including graceful shutdown)
- **OpenAPI artifacts** generated from declared operations

## Build a full API with docs

```go
package main

import (
    "context"
    "os"
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
    router := mux.NewRouter().Safe()

    // Route groups help keep versioned paths and tags consistent.
    api := router.NewRouteGroup("/api/v1")
    api.WithTags("API v1")

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
        WithPathParam("id", "The unique user identifier", uuid.Nil).
        WithOKResponse(User{})

    if err := router.Err(); err != nil {
        panic(err)
    }

    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    // Generate OpenAPI after routes are declared.
    generator := mux.NewGenerator()
    spec, err := mux.GenerateSpecWithGenerator(generator, router)
    if err != nil {
        panic(err)
    }
    if err := spec.MarshalToFile("openapi.yaml"); err != nil {
        panic(err)
    }

    server := mux.NewServer(":8080", router)
    if err := server.Listen(ctx); err != nil {
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
    // Use Bind when your handler accepts JSON, form-urlencoded,
    // or multipart form data on POST, PUT, or PATCH handlers.
    // Query params, route params, and headers are bound for all methods.
    // Top-level JSON arrays bind into slice targets, but that bind is body-only;
    // read query, path, or declared header values separately in array handlers.
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

## Add middleware

```go
router := mux.NewRouter()

// Built-in middleware
mux.UseLogging(router)
mux.UseCompression(router)
mux.UseEnforceHTTPS(router)
mux.UseOpenTelemetry(router)

// Auth middleware
mux.UseAuthentication(router,
    mux.WithValidator(validateToken),
    mux.WithTokenCreator(createToken),
)

// Custom middleware
router.Use(&CustomMiddleware{})
```

## Register services

```go
router := mux.NewRouter()

router.Services().
    Register("auditWriter", auditWriter).
    Register("clock", systemClock)

api := router.NewRouteGroup("/api")
api.Services().Register("mailer", mailer)
```

Handlers and middleware can read scoped values with `c.GetService(...)`. Child groups and route builders can override root registrations when they need different behavior.

## Docs and examples

- [Getting started](docs/getting-started.md)
- [Middleware](docs/middleware.md)
- [Custom middleware](docs/custom-middleware.md)
- [Router](docs/router.md)
- [OpenAPI guide](docs/overview.md)
- [Examples](examples/)
