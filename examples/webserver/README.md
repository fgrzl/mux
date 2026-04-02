# WebServer Example

An example focused on `mux.NewServer(...).Listen(ctx)`, health probes, and clean shutdown.

## Features

- graceful shutdown on `Ctrl+C` or `SIGTERM`
- default timeouts (10s read/write, 120s idle)
- built-in health probes
- structured logging with `slog`

## Run It

```bash
go run .
```

The server listens on `http://localhost:8080`.

## Try It

```bash
curl http://localhost:8080/
curl http://localhost:8080/api/v1/status
curl http://localhost:8080/healthz
curl http://localhost:8080/livez
curl http://localhost:8080/readyz
```

## Graceful Shutdown

1. Start the server with `go run .`
2. Send a request from another terminal
3. Press `Ctrl+C`

The process waits for in-flight requests to finish before it exits.

## Why This Example Exists

The standard-library path is flexible, but you have to wire together timeouts, signal handling, and shutdown behavior yourself. This example shows the shorter default path:

```go
router := mux.NewRouter()

if err := router.Configure(func(router *mux.Router) {
    router.GET("/", helloHandler)
}); err != nil {
    panic(err)
}

server := mux.NewServer(":8080", router)

ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer cancel()

if err := server.Listen(ctx); err != nil {
    panic(err)
}
```

## HTTPS

```go
server := mux.NewServer(":8443", router,
    mux.WithTLS("server.crt", "server.key"),
)
```

Or let Mux search parent directories for a certificate folder:

```go
server := mux.NewServer(":8443", router,
    mux.WithTLSDiscovery("certs", "server.crt", "server.key"),
)
```

## Learn More

- [WebServer Guide](../../docs/webserver.md)
- [Middleware Guide](../../docs/middleware.md)
- [Best Practices](../../docs/best-practices.md)
