# Mux

[![CI](https://github.com/fgrzl/mux/actions/workflows/ci.yaml/badge.svg)](https://github.com/fgrzl/mux/actions/workflows/ci.yaml)
[![Dependabot](https://github.com/fgrzl/mux/actions/workflows/dependabot/dependabot-updates/badge.svg)](https://github.com/fgrzl/mux/actions/workflows/dependabot/dependabot-updates)

HTTP router for Go with an OpenAPI 3.1 documentation DSL.

## Install

```bash
go get github.com/fgrzl/mux
```

## 1) Start simple

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

## 2) Build a small API with docs

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
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    router := mux.NewRouter()

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
        WithPathParam("id", uuid.Nil).
        WithOKResponse(User{})

    // Generate OpenAPI after routes are declared.
    generator := mux.NewGenerator()
    spec := generator.GenerateSpec(router)
    _ = spec.MarshalToFile("openapi.yaml")

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
    // Use Bind when your handler accepts a request body or mixed inputs
    // (JSON, form values, query params, and route params).
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

## 3) Add middleware the usual way

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

## Docs and examples

- [Getting started](docs/getting-started.md)
- [Middleware](docs/middleware.md)
- [Custom middleware](docs/custom-middleware.md)
- [Router](docs/router.md)
- [OpenAPI guide](docs/overview.md)
- [Examples](examples/)
