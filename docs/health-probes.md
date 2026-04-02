# Health Probes

Mux provides built-in Kubernetes-style health probe endpoints for monitoring application health and readiness.

## Overview

Health probes help orchestration platforms like Kubernetes determine:
- **Liveness**: Is the application alive (not deadlocked)?
- **Readiness**: Is the application ready to serve traffic?
- **Health**: Overall application health status

## Quick Start

```go
router := mux.NewRouter()

if err := router.Configure(func(router *mux.Router) {
    // Simple health probes (always return "ok")
    router.Healthz()  // GET /healthz
    router.Livez()    // GET /livez
    router.Readyz()   // GET /readyz
}); err != nil {
    panic(err)
}

ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer cancel()

server := mux.NewServer(":8080", router)
if err := server.Listen(ctx); err != nil {
    panic(err)
}
```

Test:
```bash
curl http://localhost:8080/healthz
# Returns: ok (200 OK)
```

## Built-in Methods

### Healthz()

General health check endpoint that always returns healthy by default.

```go
router.Healthz()
```

**Response**:
- `200 OK` with body `"ok"` when healthy
- `503 Service Unavailable` with body `"not ready"` when unhealthy

**Kubernetes usage**:
```yaml
livenessProbe:
  httpGet:
    path: /healthz
    port: 8080
  initialDelaySeconds: 3
  periodSeconds: 10
```

### Livez()

Liveness probe to check if the application is alive (not deadlocked or crashed).

```go
router.Livez()
```

**When to use**: Check if application should be restarted.

**Kubernetes usage**:
```yaml
livenessProbe:
  httpGet:
    path: /livez
    port: 8080
  initialDelaySeconds: 15
  periodSeconds: 20
```

### Readyz()

Readiness probe to check if the application is ready to serve traffic.

```go
router.Readyz()
```

**When to use**: Check if application should receive traffic from load balancer.

**Kubernetes usage**:
```yaml
readinessProbe:
  httpGet:
    path: /readyz
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
```

## Custom Health Checks

### HealthzWithReady()

Add custom logic to determine health status:

```go
router.HealthzWithReady(func(c mux.RouteContext) bool {
    // Check database connection
    if err := db.Ping(); err != nil {
        return false
    }
    
    // Check cache connection
    if err := cache.Ping(); err != nil {
        return false
    }
    
    return true
})
```

### LivezWithCheck()

Add custom liveness check:

```go
router.LivezWithCheck(func(c mux.RouteContext) bool {
    // Check if application is not deadlocked
    if runtime.NumGoroutine() > 10000 {
        return false // Too many goroutines, possible leak
    }
    
    // Check if critical goroutines are running
    if !workerPool.IsAlive() {
        return false
    }
    
    return true
})
```

### ReadyzWithCheck()

Add custom readiness check:

```go
router.ReadyzWithCheck(func(c mux.RouteContext) bool {
    // Check database is ready
    if !db.Ready() {
        return false
    }
    
    // Check cache is ready
    if !cache.Ready() {
        return false
    }
    
    // Check migrations completed
    if !migration.IsComplete() {
        return false
    }
    
    // Check all dependencies are ready
    return true
})
```

## Complete Example

```go
package main

import (
	"database/sql"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/fgrzl/mux"
)

var (
	db              *sql.DB
	cache           *Cache
	isReady         atomic.Bool
	startupComplete atomic.Bool
)

func main() {
	router := mux.NewRouter()

	if err := router.Configure(func(router *mux.Router) {
		// Liveness: Check if app is alive (not deadlocked)
		router.LivezWithCheck(func(c mux.RouteContext) bool {
			// Check goroutine count
			if runtime.NumGoroutine() > 10000 {
				return false
			}
			return true
		})

		// Readiness: Check if app is ready to serve traffic
		router.ReadyzWithCheck(func(c mux.RouteContext) bool {
			// Don't serve traffic until startup is complete
			if !startupComplete.Load() {
				return false
			}

			// Check database
			if err := db.Ping(); err != nil {
				return false
			}

			// Check cache
			if !cache.IsHealthy() {
				return false
			}

			return isReady.Load()
		})

		// General health check
		router.HealthzWithReady(func(c mux.RouteContext) bool {
			return isReady.Load()
		})

		// API routes
		api := router.Group("/api/v1")
		setupRoutes(api)
	}); err != nil {
		panic(err)
	}

	// Initialize dependencies in background
	go func() {
		initializeDatabase()
		initializeCache()
		runMigrations()

		startupComplete.Store(true)
		isReady.Store(true)
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	server := mux.NewServer(":8080", router)
	if err := server.Listen(ctx); err != nil {
		panic(err)
	}
}
```

## Best Practices

### 1. Liveness vs Readiness

**Liveness checks should be simple**:
- Check if app is responsive
- Check for deadlocks
- Don't check external dependencies (DB, cache)

**Readiness checks can be complex**:
- Check database connections
- Check cache availability  
- Check migrations completed
- Check external service dependencies

### 2. Fast Health Checks

Keep health checks fast (< 100ms):

```go
// Bad: Slow health check
router.ReadyzWithCheck(func(c mux.RouteContext) bool {
    // This might take seconds!
    rows, _ := db.Query("SELECT COUNT(*) FROM large_table")
    return rows != nil
})

// Good: Fast health check
router.ReadyzWithCheck(func(c mux.RouteContext) bool {
    // Quick ping
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()
    return db.PingContext(ctx) == nil
})
```

### 3. Separate Liveness and Readiness

Don't conflate liveness with readiness:

```go
// Bad: Database failure kills pod
router.LivezWithCheck(func(c mux.RouteContext) bool {
    return db.Ping() == nil // Restarting won't fix DB issues!
})

// Good: Database failure removes from load balancer
router.ReadyzWithCheck(func(c mux.RouteContext) bool {
    return db.Ping() == nil // Pod stays alive, just not ready
})
```

### 4. Graceful Degradation

Mark as not ready before shutdown:

```go
func main() {
	router := mux.NewRouter()

	if err := router.Configure(func(router *mux.Router) {
		router.ReadyzWithCheck(func(c mux.RouteContext) bool {
			return isReady.Load()
		})
	}); err != nil {
		panic(err)
	}

	serveCtx, stopServing := context.WithCancel(context.Background())
	defer stopServing()

	server := mux.NewServer(":8080", router)
	if err := server.Start(serveCtx); err != nil {
		panic(err)
	}

	// Wait for shutdown signal
	quitCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-quitCtx.Done()

	// Mark as not ready (stop receiving traffic)
	isReady.Store(false)
	time.Sleep(5 * time.Second) // Let readiness probes detect

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Stop(ctx); err != nil {
		panic(err)
	}
}
```

## Kubernetes Configuration

### Complete Pod Spec

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: myapp
spec:
  containers:
  - name: myapp
    image: myapp:latest
    ports:
    - containerPort: 8080
    
    # Liveness: Restart if not responding
    livenessProbe:
      httpGet:
        path: /livez
        port: 8080
      initialDelaySeconds: 30
      periodSeconds: 10
      timeoutSeconds: 5
      failureThreshold: 3
    
    # Readiness: Remove from service if not ready
    readinessProbe:
      httpGet:
        path: /readyz
        port: 8080
      initialDelaySeconds: 5
      periodSeconds: 5
      timeoutSeconds: 3
      failureThreshold: 2
    
    # Startup: Wait for app to start before liveness checks
    startupProbe:
      httpGet:
        path: /healthz
        port: 8080
      initialDelaySeconds: 0
      periodSeconds: 2
      timeoutSeconds: 3
      failureThreshold: 30  # 60 seconds total
```

## Advanced Patterns

### Circuit Breaker Integration

```go
var dbCircuitBreaker *CircuitBreaker

router.ReadyzWithCheck(func(c mux.RouteContext) bool {
    // Don't serve traffic if circuit is open
    return dbCircuitBreaker.State() != Open
})
```

### Dependency Health Aggregation

```go
type HealthChecker interface {
    IsHealthy() bool
}

var dependencies = []HealthChecker{
    db,
    cache,
    queue,
    externalAPI,
}

router.ReadyzWithCheck(func(c mux.RouteContext) bool {
    for _, dep := range dependencies {
        if !dep.IsHealthy() {
            return false
        }
    }
    return true
})
```

### Detailed Health Response

For debugging, use a custom endpoint with detailed status:

```go
type HealthStatus struct {
    Status       string            `json:"status"`
    Version      string            `json:"version"`
    Dependencies map[string]string `json:"dependencies"`
}

router.GET("/health/detailed", func(c mux.RouteContext) {
    status := HealthStatus{
        Status:       "healthy",
        Version:      "1.0.0",
        Dependencies: make(map[string]string),
    }
    
    // Check each dependency
    if err := db.Ping(); err != nil {
        status.Status = "unhealthy"
        status.Dependencies["database"] = err.Error()
    } else {
        status.Dependencies["database"] = "ok"
    }
    
    if !cache.IsHealthy() {
        status.Status = "degraded"
        status.Dependencies["cache"] = "unavailable"
    } else {
        status.Dependencies["cache"] = "ok"
    }
    
    if status.Status == "healthy" {
        c.OK(status)
    } else {
        c.JSON(http.StatusServiceUnavailable, status)
    }
})
```

## Authentication

Health probes automatically allow anonymous access (no authentication required).

```go
// Authentication required for API
mux.UseAuthentication(router, /* ... */)

// But health probes are always accessible
router.Healthz()  // No auth required
router.Livez()    // No auth required
router.Readyz()   // No auth required
```

This is by design - health probes must be accessible to orchestration platforms.

## Response Format

All probe endpoints return plain text:

**Success**: HTTP 200 OK with body `"ok"`  
**Failure**: HTTP 503 Service Unavailable with body `"not ready"`, `"not live"`, etc.

This follows Kubernetes conventions for health checks.

## See Also

- [Kubernetes Probes Documentation](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
- [Best Practices](best-practices.md) - Production deployment patterns
- [Middleware Guide](middleware.md) - Additional middleware options

---

**Built with Mux**



