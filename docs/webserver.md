# WebServer

`WebServer` is a production-ready HTTP server wrapper that provides graceful shutdown, TLS support, and sensible timeouts out of the box.

## Why Use WebServer?

Instead of manually configuring `http.Server` with timeouts and shutdown logic, `WebServer` gives you:

- ✅ **Production-ready defaults**: Sensible timeouts for read/write/idle
- ✅ **Graceful shutdown**: Automatic cleanup when context is canceled
- ✅ **TLS support**: Easy HTTPS with certificate discovery
- ✅ **Context-aware**: Start/Listen methods respect context cancellation
- ✅ **Error handling**: Proper error propagation and logging

## Quick Start

### Basic HTTP Server

```go
package main

import (
    "context"
    "os"
    "os/signal"
    
    "github.com/fgrzl/mux"
)

func main() {
    router := mux.NewRouter()
    
    router.GET("/", func(c mux.RouteContext) {
        c.OK("Hello, World!")
    })
    
    // Create server with production defaults
    server := mux.NewServer(":8080", router)
    
    // Run with graceful shutdown
    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
    defer cancel()
    
    if err := server.Listen(ctx); err != nil {
        panic(err)
    }
}
```

**That's it!** The server will:
- Start listening on port 8080
- Handle requests with proper timeouts
- Gracefully shutdown when you press Ctrl+C

## API Reference

### NewServer

Creates a new `WebServer` with production-ready defaults.

```go
func NewServer(addr string, rtr *router.Router, opts ...WebServerOption) *WebServer
```

**Parameters**:
- `addr`: Server address (e.g., `:8080`, `localhost:3000`)
- `rtr`: Configured Mux router
- `opts`: Optional configuration (TLS, custom settings)

**Default Timeouts**:
- `ReadTimeout`: 10 seconds
- `WriteTimeout`: 10 seconds
- `IdleTimeout`: 120 seconds

**Example**:
```go
server := mux.NewServer(":8080", router)
```

### Listen

Blocks until the server exits or context is canceled. **Most common method**.

```go
func (ws *WebServer) Listen(ctx context.Context) error
```

**Usage**:
```go
ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
defer cancel()

if err := server.Listen(ctx); err != nil {
    log.Fatal(err)
}
```

**Behavior**:
- Binds to the address and starts accepting connections
- Blocks until context is canceled or error occurs
- Gracefully shuts down when context is canceled
- Returns `nil` on clean shutdown, error otherwise

### Start

Starts the server in the background and returns immediately.

```go
func (ws *WebServer) Start(ctx context.Context) error
```

**Usage**:
```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

if err := server.Start(ctx); err != nil {
    log.Fatal(err)
}

// Server is now running in background
// Do other work...

// Cancel context to trigger graceful shutdown
cancel()
```

**Behavior**:
- Starts server in a goroutine
- Returns immediately (non-blocking)
- Server shuts down when context is canceled
- Use when you need to run other code after starting the server

### Stop

Manually triggers graceful shutdown.

```go
func (ws *WebServer) Stop(ctx context.Context) error
```

**Usage**:
```go
// Give server 30 seconds to shutdown gracefully
shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

if err := server.Stop(shutdownCtx); err != nil {
    log.Println("Forced shutdown:", err)
}
```

## TLS/HTTPS Support

### Basic HTTPS

```go
server := mux.NewServer(":8443", router,
    mux.WithTLS("server.crt", "server.key"),
)

ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
defer cancel()

server.Listen(ctx)
```

### TLS Certificate Discovery

Automatically searches for certificates in parent directories:

```go
server := mux.NewServer(":8443", router,
    mux.WithTLSDiscovery("certs", "server.crt", "server.key"),
)
```

**How it works**:
1. Starts from current directory
2. Looks for a `certs/` directory
3. Checks up to 10 parent directories
4. Uses `certs/server.crt` and `certs/server.key` when found

**Directory structure example**:
```
project/
├── certs/
│   ├── server.crt
│   └── server.key
├── cmd/
│   └── api/
│       └── main.go  ← WithTLSDiscovery finds ../../certs/
└── pkg/
```

## Complete Examples

### Production Server with Graceful Shutdown

```go
package main

import (
    "context"
    "log/slog"
    "os"
    "os/signal"
    "syscall"
    
    "github.com/fgrzl/mux"
)

func main() {
    // Configure router
    router := mux.NewRouter()
    
    // Add health probes
    router.Healthz()
    router.Readyz()
    
    // Add your routes
    setupRoutes(router)
    
    // Create server
    server := mux.NewServer(":8080", router)
    
    // Setup graceful shutdown on SIGINT or SIGTERM
    ctx, cancel := signal.NotifyContext(
        context.Background(),
        os.Interrupt,
        syscall.SIGTERM,
    )
    defer cancel()
    
    slog.Info("Starting server", "addr", ":8080")
    
    // Listen blocks until shutdown
    if err := server.Listen(ctx); err != nil {
        slog.Error("Server error", "error", err)
        os.Exit(1)
    }
    
    slog.Info("Server shutdown complete")
}

func setupRoutes(router *mux.Router) {
    api := router.NewRouteGroup("/api/v1")
    // ... your routes
}
```

### HTTPS Server with TLS

```go
package main

import (
    "context"
    "os"
    "os/signal"
    
    "github.com/fgrzl/mux"
)

func main() {
    router := mux.NewRouter()
    
    router.GET("/", func(c mux.RouteContext) {
        c.OK("Secure Hello!")
    })
    
    // HTTPS server on port 8443
    server := mux.NewServer(":8443", router,
        mux.WithTLS("certs/server.crt", "certs/server.key"),
    )
    
    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
    defer cancel()
    
    server.Listen(ctx)
}
```

### Running HTTP and HTTPS Simultaneously

```go
package main

import (
    "context"
    "os"
    "os/signal"
    "sync"
    
    "github.com/fgrzl/mux"
)

func main() {
    router := mux.NewRouter()
    router.GET("/", func(c mux.RouteContext) {
        c.OK("Hello from both HTTP and HTTPS!")
    })
    
    // Create both HTTP and HTTPS servers
    httpServer := mux.NewServer(":8080", router)
    httpsServer := mux.NewServer(":8443", router,
        mux.WithTLS("certs/server.crt", "certs/server.key"),
    )
    
    // Setup graceful shutdown
    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
    defer cancel()
    
    // Run both servers
    var wg sync.WaitGroup
    
    wg.Add(1)
    go func() {
        defer wg.Done()
        httpServer.Listen(ctx)
    }()
    
    wg.Add(1)
    go func() {
        defer wg.Done()
        httpsServer.Listen(ctx)
    }()
    
    wg.Wait()
}
```

### Background Server with Cleanup

```go
package main

import (
    "context"
    "log/slog"
    "time"
    
    "github.com/fgrzl/mux"
)

func main() {
    router := mux.NewRouter()
    router.GET("/", func(c mux.RouteContext) {
        c.OK("Background server")
    })
    
    server := mux.NewServer(":8080", router)
    
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    // Start server in background
    if err := server.Start(ctx); err != nil {
        panic(err)
    }
    
    slog.Info("Server started in background")
    
    // Do other work
    time.Sleep(5 * time.Second)
    processRequests()
    
    // Trigger graceful shutdown
    slog.Info("Shutting down...")
    cancel()
    
    // Give server time to finish
    time.Sleep(1 * time.Second)
    slog.Info("Complete")
}

func processRequests() {
    // Your application logic
}
```

## Comparison with http.Server

### Manual http.Server (Before)

```go
// ❌ Verbose, easy to forget timeouts or graceful shutdown
func main() {
    router := mux.NewRouter()
    
    server := &http.Server{
        Addr:         ":8080",
        Handler:      router,
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  120 * time.Second,
    }
    
    // Need to manually handle graceful shutdown
    go func() {
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatal(err)
        }
    }()
    
    // Setup signal handling
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt)
    <-quit
    
    // Graceful shutdown
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    if err := server.Shutdown(ctx); err != nil {
        log.Fatal(err)
    }
}
```

### Using WebServer (After)

```go
// ✅ Simple, safe, production-ready
func main() {
    router := mux.NewRouter()
    server := mux.NewServer(":8080", router)
    
    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
    defer cancel()
    
    server.Listen(ctx)
}
```

**Lines of code**: 30+ → **7**

## Best Practices

### 1. Always Use Context for Shutdown

```go
// ✅ Good: Context-based shutdown
ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
defer cancel()
server.Listen(ctx)

// ❌ Bad: No shutdown mechanism
server.Listen(context.Background())  // Runs forever, no way to stop
```

### 2. Handle Errors from Listen

```go
// ✅ Good: Check for errors
if err := server.Listen(ctx); err != nil {
    log.Fatal(err)
}

// ❌ Bad: Ignore errors
server.Listen(ctx)  // Silent failures
```

### 3. Use Listen for Main Servers, Start for Background

```go
// ✅ Good: Listen for primary server
func main() {
    server := mux.NewServer(":8080", router)
    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
    defer cancel()
    server.Listen(ctx)  // Blocks until shutdown
}

// ✅ Good: Start for background servers
func startMetricsServer() {
    metricsRouter := mux.NewRouter()
    server := mux.NewServer(":9090", metricsRouter)
    server.Start(context.Background())  // Returns immediately
}
```

### 4. Configure Appropriate Timeouts

For long-running requests (uploads, streaming):

```go
// Default WebServer timeouts might be too short
// Access the underlying http.Server if needed:

server := mux.NewServer(":8080", router)
// Timeouts are already set to production defaults:
// ReadTimeout:  10s
// WriteTimeout: 10s  
// IdleTimeout:  120s
```

For custom timeouts, use standard `http.Server`:

```go
srv := &http.Server{
    Addr:         ":8080",
    Handler:      router,
    ReadTimeout:  30 * time.Second,  // Longer for uploads
    WriteTimeout: 30 * time.Second,
}
// Then implement your own shutdown logic
```

## Kubernetes Deployment

`WebServer` works perfectly with Kubernetes:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: myapp
        image: myapp:latest
        ports:
        - containerPort: 8080
        
        # Liveness probe
        livenessProbe:
          httpGet:
            path: /livez
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        
        # Readiness probe
        readinessProbe:
          httpGet:
            path: /readyz
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        
        # Graceful shutdown
        lifecycle:
          preStop:
            exec:
              command: ["/bin/sh", "-c", "sleep 5"]
        
      # Give pods 30 seconds to shutdown gracefully
      terminationGracePeriodSeconds: 30
```

**The WebServer handles**:
- Graceful shutdown when pod receives SIGTERM
- Proper connection draining
- Health probe endpoints (if you add them to router)

## Docker Example

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]
```

```go
// main.go
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
    router.Healthz()
    router.GET("/", func(c mux.RouteContext) {
        c.OK("Hello from Docker!")
    })
    
    server := mux.NewServer(":8080", router)
    
    ctx, cancel := signal.NotifyContext(
        context.Background(),
        os.Interrupt,
        syscall.SIGTERM,  // Docker sends SIGTERM
    )
    defer cancel()
    
    server.Listen(ctx)
}
```

## Troubleshooting

### "Address already in use"

**Problem**: Port is already bound by another process

**Solution**:
```bash
# Linux/Mac: Find process using port
lsof -i :8080

# Windows: Find process using port
netstat -ano | findstr :8080

# Kill the process or use a different port
server := mux.NewServer(":8081", router)
```

### TLS Certificate Errors

**Problem**: "invalid TLS cert/key files"

**Solutions**:
```go
// 1. Check file paths
server := mux.NewServer(":8443", router,
    mux.WithTLS("./certs/server.crt", "./certs/server.key"),  // Use absolute paths
)

// 2. Verify files exist
if _, err := os.Stat("server.crt"); err != nil {
    log.Fatal("cert file not found:", err)
}

// 3. Use TLS discovery
server := mux.NewServer(":8443", router,
    mux.WithTLSDiscovery("certs", "server.crt", "server.key"),
)
```

### Graceful Shutdown Not Working

**Problem**: Server doesn't shutdown cleanly

**Solution**:
```go
// ✅ Use signal.NotifyContext for clean shutdown
ctx, cancel := signal.NotifyContext(
    context.Background(),
    os.Interrupt,
    syscall.SIGTERM,
)
defer cancel()

// Give server time to finish requests
server.Listen(ctx)
```

## See Also

- [Health Probes](health-probes.md) - Add /healthz, /livez, /readyz endpoints
- [Best Practices](best-practices.md) - Production deployment patterns
- [Examples](../examples/) - Working code samples

---

**Built with ❤️ using Mux**
