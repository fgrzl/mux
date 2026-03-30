# WebServer Example

Production-ready HTTP server with graceful shutdown.

## Features

✅ Automatic graceful shutdown on Ctrl+C or SIGTERM  
✅ Production-ready timeouts (10s read/write, 120s idle)  
✅ Health probe endpoints  
✅ Clean logging with slog  
✅ Context-based lifecycle management  

## Quick Start

```bash
go run main.go
```

Server starts on `http://localhost:8080`

## Testing

### Test the API

```bash
# Root endpoint
curl http://localhost:8080/

# API endpoint
curl http://localhost:8080/api/v1/status

# Health probes
curl http://localhost:8080/healthz
curl http://localhost:8080/livez
curl http://localhost:8080/readyz
```

### Test Graceful Shutdown

1. Start the server:
```bash
go run main.go
```

2. In another terminal, make a request:
```bash
curl http://localhost:8080/
```

3. Press `Ctrl+C` in the server terminal

You'll see:
```
Starting server addr=:8080
Press Ctrl+C to shutdown gracefully
^C
Server shutdown complete
```

The server waits for in-flight requests to complete before shutting down.

## WebServer Benefits

### Before (Manual http.Server)

```go
// ❌ 30+ lines of boilerplate
server := &http.Server{
    Addr:         ":8080",
    Handler:      router,
    ReadTimeout:  10 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  120 * time.Second,
}

go func() {
    if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatal(err)
    }
}()

quit := make(chan os.Signal, 1)
signal.Notify(quit, os.Interrupt)
<-quit

ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

if err := server.Shutdown(ctx); err != nil {
    log.Fatal(err)
}
```

### After (WebServer)

```go
// ✅ 7 lines - production ready
server := mux.NewServer(":8080", router)

ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
defer cancel()

if err := server.Listen(ctx); err != nil { panic(err) }
```

**Lines of code**: 30+ → **7**

## HTTPS/TLS Example

Add TLS with one line:

```go
server := mux.NewServer(":8443", router,
    mux.WithTLS("server.crt", "server.key"),
)
```

Or use automatic certificate discovery:

```go
server := mux.NewServer(":8443", router,
    mux.WithTLSDiscovery("certs", "server.crt", "server.key"),
)
```

This searches up to 10 parent directories for a `certs/` folder.

## Kubernetes Deployment

Perfect for Kubernetes:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  template:
    spec:
      containers:
      - name: myapp
        image: myapp:latest
        ports:
        - containerPort: 8080
        
        livenessProbe:
          httpGet:
            path: /livez
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        
        readinessProbe:
          httpGet:
            path: /readyz
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
      
      terminationGracePeriodSeconds: 30
```

**WebServer automatically**:
- Handles SIGTERM from Kubernetes
- Drains connections gracefully
- Completes in-flight requests

## What You Learned

- ✅ Using `mux.NewServer()` for production deployments
- ✅ Graceful shutdown with `signal.NotifyContext`
- ✅ Context-based lifecycle management
- ✅ Health probe endpoints
- ✅ Production-ready timeouts

## Next Steps

- Add middleware: [Middleware Guide](../../docs/middleware.md)
- Add authentication: [Authentication](../../docs/authentication-middleware.md)
- Deploy to production: [Best Practices](../../docs/best-practices.md)
- Complete WebServer docs: [WebServer Guide](../../docs/webserver.md)

---

**Built with ❤️ using Mux**
