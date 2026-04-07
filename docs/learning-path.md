# Learning Path for Mux

This guide gives you a deliberate path from your first route to production-ready services without making you learn every part of the framework at once.
It follows the way Mux is meant to be used: one cohesive stack for routing, request handling, middleware, and API description rather than a router plus a pile of add-ons.

## Stage 1: First Route

Start here if you just want a working endpoint.

Read:

- [Installation](installation.md)
- [Quick Start](quick-start.md)

Goal:

- Start a service with `mux.NewRouter()`, `router.Configure(...)`, and `mux.NewServer(...).Listen(ctx)`

Example:

```go
router := mux.NewRouter()

if err := router.Configure(func(router *mux.Router) {
    router.GET("/hello", func(c mux.RouteContext) {
        c.OK("Hello, World!")
    })
}); err != nil {
    panic(err)
}
```

## Stage 2: Request Data

Once you can serve one route, learn how to read path parameters, query values, and request bodies.

Read:

- [Getting Started](getting-started.md)
- [Cheat Sheet](cheat-sheet.md)

Focus on:

- `c.Params().String("name")`
- `c.Query().String("name")`, `c.Query().Int("limit")`, `c.Query().Bool("completed")`
- `c.Bind(&value)`
- `c.BadRequest(title, detail)`

Example:

```go
router.GET("/users/{id}", func(c mux.RouteContext) {
    id, ok := c.Params().String("id")
    if !ok {
        c.BadRequest("Missing user ID", "id parameter is required")
        return
    }

    limit, _ := c.Query().Int("limit")
    c.OK(map[string]any{"id": id, "limit": limit})
})
```

## Stage 3: Structure Your API

Learn how to group routes, share tags, and apply middleware at startup.

Read:

- [Router](router.md)
- [Middleware](middleware.md)

Focus on:

- `router.Group(...)`
- `mux.UseLogging(router)` and other startup middleware
- `router.Services().Register(...)` for shared collaborators

Example:

```go
router := mux.NewRouter()

mux.UseLogging(router)
router.Services().Register(mux.ServiceKey("clock"), time.Now)

if err := router.Configure(func(router *mux.Router) {
    api := router.Group("/api/v1")
    api.WithTags("API v1")

    users := api.Group("/users")
    users.GET("/", listUsers)
    users.POST("/", createUser)
}); err != nil {
    panic(err)
}
```

## Stage 4: Add OpenAPI Metadata

Once the routes feel stable, add documentation directly to the route builders.

Read:

- [Cheat Sheet](cheat-sheet.md)
- [Overview](overview.md)

Focus on:

- `WithOperationID(...)`
- `WithSummary(...)`
- `WithJsonBody(...)`
- `WithOKResponse(...)`, `WithCreatedResponse(...)`, `WithResponse(404, mux.ProblemDetails{})`
- `mux.GenerateSpecWithGenerator(...)`

Example:

```go
users.POST("/", createUser).
    WithOperationID("createUser").
    WithSummary("Create a new user").
    WithJsonBody(CreateUserRequest{}).
    WithCreatedResponse(User{})

router.GET("/openapi.json", func(c mux.RouteContext) {
    spec, err := mux.GenerateSpecWithGenerator(mux.NewGenerator(), router)
    if err != nil {
        c.ServerError("OpenAPI generation failed", err.Error())
        return
    }
    c.OK(spec)
})
```

## Stage 5: Incremental Adoption and Interop

You do not need to rewrite all existing handlers at once. Mux supports standard-library handlers and lets you recover the active route context when needed.

Read:

- [Overview](overview.md)
- [Router](router.md)

Focus on:

- `router.Handle(...)`
- `router.HandleFunc(...)`
- `mux.RouteContextFromRequest(r)`

Example:

```go
router.HandleFunc(http.MethodGet, "/legacy/users/{id}", func(w http.ResponseWriter, r *http.Request) {
    routeCtx, ok := mux.RouteContextFromRequest(r)
    if !ok {
        http.Error(w, "route context not available", http.StatusInternalServerError)
        return
    }

    id, _ := routeCtx.Params().String("id")
    routeCtx.OK(map[string]string{"id": id})
})
```

## Stage 6: Production Readiness

Finish with server lifecycle, health probes, and router options.

Read:

- [WebServer](webserver.md)
- [Health Probes](health-probes.md)
- [Best Practices](best-practices.md)

Focus on:

- `mux.NewServer(...).Listen(ctx)`
- `router.Healthz()`, `router.Livez()`, `router.Readyz()`
- `mux.WithContextPooling()`
- `mux.WithHeadFallbackToGet()`
- `mux.WithMaxBodyBytes(...)`

Example:

```go
router := mux.NewRouter(
    mux.WithContextPooling(),
    mux.WithHeadFallbackToGet(),
    mux.WithMaxBodyBytes(10<<20),
)

if err := router.Configure(func(router *mux.Router) {
    router.Healthz()
    router.Readyz()
    setupRoutes(router)
}); err != nil {
    panic(err)
}

ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()

server := mux.NewServer(":8080", router)
if err := server.Listen(ctx); err != nil {
    panic(err)
}
```

## Suggested Order

1. [Installation](installation.md)
2. [Quick Start](quick-start.md)
3. [Getting Started](getting-started.md)
4. [Interactive Tutorial](interactive-tutorial.md)
5. [Middleware](middleware.md)
6. [WebServer](webserver.md)
7. [Examples](../examples/)

## When You Are Ready for Production

You are in good shape when you can do all of the following comfortably:

- Start a service with `Configure(...)` and fail fast on startup validation errors
- Read params, query values, and request bodies without touching raw `http.Request` unnecessarily
- Organize routes with groups and shared middleware
- Publish an OpenAPI document from the registered routes
- Run the service through `WebServer` with graceful shutdown and health probes

## Recommended Example Apps

- [Hello World](../examples/hello-world/)
- [Todo API](../examples/todo-api/)
- [WebServer](../examples/webserver/)
- [Redirects](../examples/redirects/)


