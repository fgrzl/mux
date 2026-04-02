# WebServer

`WebServer` is the production-oriented HTTP server wrapper exposed by Mux. It wraps `http.Server` with sensible defaults, graceful shutdown, and TLS helpers.

## Why Use It?

- Production defaults out of the box: 10 second read timeout, 10 second write timeout, 120 second idle timeout
- Graceful shutdown when the provided context is canceled
- Built-in TLS helpers for explicit cert paths or discovery-based lookup
- A simple blocking `Listen` path and a background `Start` path

## Quick Start

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
        router.GET("/", func(c mux.RouteContext) {
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

`Listen` blocks until the server exits or the context is canceled. On cancellation, `WebServer` performs a graceful shutdown using a 10 second shutdown timeout.

## Core API

### `mux.NewServer`

```go
server := mux.NewServer(":8080", router)
```

Constructs a `WebServer` with production defaults and optional configuration.

### `server.Listen(ctx)`

- Binds the listener
- Starts serving immediately
- Blocks until shutdown or an unexpected server error
- Gracefully shuts down when `ctx` is canceled

Use `Listen` for your main application server.

### `server.Start(ctx)`

Starts serving in the background and returns immediately.

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

server := mux.NewServer(":8080", router)
if err := server.Start(ctx); err != nil {
    panic(err)
}

// Do other work here.
```

Use `Start` when your process needs to keep doing work after the HTTP server begins accepting traffic.

### `server.Stop(ctx)`

Triggers graceful shutdown explicitly.

```go
shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

if err := server.Stop(shutdownCtx); err != nil {
    panic(err)
}
```

## WebServer Options

### Timeout Options

- `mux.WithReadTimeout(d time.Duration)`
- `mux.WithWriteTimeout(d time.Duration)`
- `mux.WithIdleTimeout(d time.Duration)`

Example:

```go
server := mux.NewServer(":8080", router,
    mux.WithReadTimeout(30*time.Second),
    mux.WithWriteTimeout(30*time.Second),
    mux.WithIdleTimeout(2*time.Minute),
)
```

### TLS Options

- `mux.WithTLS(certFile, keyFile)`
- `mux.WithTLSDiscovery(certsDir, certFile, keyFile)`

Use explicit TLS paths when you know where the certs live:

```go
server := mux.NewServer(":8443", router,
    mux.WithTLS("certs/server.crt", "certs/server.key"),
)
```

Use discovery when the executable may start from different working directories:

```go
server := mux.NewServer(":8443", router,
    mux.WithTLSDiscovery("certs", "server.crt", "server.key"),
)
```

`WithTLSDiscovery` searches upward for a `certs` directory, up to 10 parent directories.

## Health Probe Integration

`WebServer` pairs naturally with the router's built-in health probe helpers:

```go
router := mux.NewRouter()

if err := router.Configure(func(router *mux.Router) {
    router.Healthz()
    router.Livez()
    router.Readyz()
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

For a full walkthrough, see [health-probes.md](health-probes.md).

## When to Use `http.Server` Directly

Use `WebServer` by default. Drop down to raw `http.Server` only when you need server features that Mux does not surface directly, such as advanced HTTP/2 or transport-level customization beyond the provided timeout and TLS helpers.

Even in that case, the router still implements `http.Handler`, so it can be used directly:

```go
srv := &http.Server{
    Addr:    ":8080",
    Handler: router,
}
```

## Recommended Pattern

The common production pattern in this repository is:

1. Create the router with `mux.NewRouter(...)`
2. Register routes inside `router.Configure(...)`
3. Add middleware during startup
4. Run the service with `mux.NewServer(...).Listen(ctx)`

That keeps route validation explicit and server lifecycle management predictable.

## See Also

- [Quick Start](quick-start.md)
- [Getting Started](getting-started.md)
- [Health Probes](health-probes.md)
- [Best Practices](best-practices.md)
- [WebServer example](../examples/webserver/)


