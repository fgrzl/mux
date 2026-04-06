# Mux

[![CI](https://github.com/fgrzl/mux/actions/workflows/ci.yaml/badge.svg)](https://github.com/fgrzl/mux/actions/workflows/ci.yaml)
[![Dependabot](https://github.com/fgrzl/mux/actions/workflows/dependabot/dependabot-updates/badge.svg)](https://github.com/fgrzl/mux/actions/workflows/dependabot/dependabot-updates)

**Mux** is an OpenAPI-native HTTP framework for Go. Routes, request binding, structured errors, and generated OpenAPI stay in one model.

## Get Started in 2 Minutes

Mux requires Go 1.25.6 or later.

### 1. Create a project

```bash
mkdir my-api
cd my-api
go mod init my-api
go get github.com/fgrzl/mux
```

### 2. Paste this into `main.go`

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
    router := mux.NewRouter()

    if err := router.Configure(func(router *mux.Router) {
        router.GET("/hello", func(c mux.RouteContext) {
            c.OK("Hello, World!")
        })
    }); err != nil {
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

### 3. Run it

```bash
go run .
curl http://localhost:8080/hello
```

Expected output:

```json
"Hello, World!"
```

`Configure` is the recommended startup path because it returns route-registration errors directly before the server starts serving traffic.

## What To Open Next

- Want a real API quickly: [examples/todo-api/](examples/todo-api/)
- Want the smallest example: [examples/hello-world/](examples/hello-world/)
- Want the guided version: [docs/quick-start.md](docs/quick-start.md)
- Want a broader walkthrough: [docs/getting-started.md](docs/getting-started.md)

## Why Mux

Use **Mux** if you want:

- **Schema-driven APIs** where routes, validation, and OpenAPI stay in sync
- **Explicit behavior** for middleware order, errors, and lifecycle
- **Production defaults** like graceful shutdown and structured error helpers
- **Scoped services** on routers, groups, and routes without extra plumbing
- **Stdlib-friendly adoption** when you need to keep existing `net/http` handlers

## Next: JSON + OpenAPI

Once the hello-world route works, the next step is usually `Bind`, route groups, and generated docs:

```go
type CreateUserRequest struct {
    Name string `json:"name"`
}

if err := router.Configure(func(router *mux.Router) {
    api := router.Group("/api")
    api.Tags("API")

    api.POST("/users", func(c mux.RouteContext) {
        var req CreateUserRequest
        if err := c.Bind(&req); err != nil {
            c.BadRequest("Invalid request", err.Error())
            return
        }

        c.Created(map[string]any{"id": "user-123", "name": req.Name})
    }).
        OperationID("createUser").
        Summary("Create a user").
        AcceptJSON(CreateUserRequest{}).
        Created(map[string]any{"id": "user-123", "name": "Ada"})
}); err != nil {
    panic(err)
}

spec, err := mux.GenerateSpecWithGenerator(mux.NewGenerator(), router)
if err != nil {
    panic(err)
}
if err := spec.MarshalToFile("openapi.yaml"); err != nil {
    panic(err)
}
```

If you want the full version of that flow, start with [examples/todo-api/](examples/todo-api/) or the [interactive tutorial](docs/interactive-tutorial.md).

## Already Have `net/http` Handlers?

Use `Handle` or `HandleFunc` to keep standard-library handlers and only opt into Mux features when you need them.

```go
import "net/http"

router.HandleFunc(http.MethodGet, "/healthz", func(w http.ResponseWriter, r *http.Request) {
    if routeCtx, ok := mux.RouteContextFromRequest(r); ok {
        _, _ = routeCtx.Services().Get(mux.ServiceKey("db"))
    }
    w.WriteHeader(http.StatusNoContent)
})
```

## Add Middleware

```go
router := mux.NewRouter()

mux.UseLogging(router)
mux.UseCompression(router)
mux.UseEnforceHTTPS(router)
mux.UseOpenTelemetry(router)

mux.UseAuthentication(router,
    mux.WithAuthValidator(validateToken),
    mux.WithAuthTokenCreator(createToken),
)

router.Use(&CustomMiddleware{})
```

## Register Services

```go
router := mux.NewRouter()

router.Services().
    Register(mux.ServiceKey("auditWriter"), auditWriter).
    Register(mux.ServiceKey("clock"), systemClock)

api := router.Group("/api")
api.Services().Register(mux.ServiceKey("mailer"), mailer)
```

Handlers and middleware read scoped values with `c.Services().Get(...)`. Child groups and route builders can override root registrations when they need different behavior.

## Docs and Examples

- [docs/quick-start.md](docs/quick-start.md)
- [docs/getting-started.md](docs/getting-started.md)
- [docs/router.md](docs/router.md)
- [docs/middleware.md](docs/middleware.md)
- [docs/overview.md](docs/overview.md)
- [examples/hello-world/](examples/hello-world/)
- [examples/todo-api/](examples/todo-api/)
- [examples/webserver/](examples/webserver/)
- [examples/](examples/)



