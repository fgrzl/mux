# Custom Middleware

Learn how to create and use custom middleware with Mux.

## Middleware Interface

All middleware must implement the `Middleware` interface:

```go
type Middleware interface {
    Invoke(ctx *RouteContext, next HandlerFunc)
}
```

## Simple Middleware Example

```go
type TimingMiddleware struct{}

func (m *TimingMiddleware) Invoke(c *mux.RouteContext, next mux.HandlerFunc) {
    start := time.Now()
    
    // Call next middleware/handler
    next(c)
    
    // Log execution time
    duration := time.Since(start)
    log.Printf("Request to %s took %v", c.Request.URL.Path, duration)
}

// Add to router
router.Use(&TimingMiddleware{})
```

## Middleware with Configuration

```go
type RateLimitMiddleware struct {
    RequestsPerMinute int
    visitors          map[string]*rate.Limiter
    mu                sync.RWMutex
}

func NewRateLimitMiddleware(rpm int) *RateLimitMiddleware {
    return &RateLimitMiddleware{
        RequestsPerMinute: rpm,
        visitors:          make(map[string]*rate.Limiter),
    }
}

func (m *RateLimitMiddleware) Invoke(c *mux.RouteContext, next mux.HandlerFunc) {
    ip := c.Request.RemoteAddr
    
    m.mu.Lock()
    limiter, exists := m.visitors[ip]
    if !exists {
        limiter = rate.NewLimiter(rate.Every(time.Minute/time.Duration(m.RequestsPerMinute)), 1)
        m.visitors[ip] = limiter
    }
    m.mu.Unlock()
    
    if !limiter.Allow() {
        c.Response.WriteHeader(http.StatusTooManyRequests)
        return
    }
    
    next(c)
}

// Usage
router.Use(NewRateLimitMiddleware(100)) // 100 requests per minute
```

## Request/Response Modification

```go
type SecurityHeadersMiddleware struct{}

func (m *SecurityHeadersMiddleware) Invoke(c *mux.RouteContext, next mux.HandlerFunc) {
    // Add security headers before processing
    c.Response.Header().Set("X-Content-Type-Options", "nosniff")
    c.Response.Header().Set("X-Frame-Options", "DENY")
    c.Response.Header().Set("X-XSS-Protection", "1; mode=block")
    
    next(c)
}
```

## Conditional Middleware

```go
type CORSMiddleware struct {
    AllowedOrigins []string
}

func (m *CORSMiddleware) Invoke(c *mux.RouteContext, next mux.HandlerFunc) {
    origin := c.Request.Header.Get("Origin")
    
    // Check if origin is allowed
    allowed := false
    for _, allowedOrigin := range m.AllowedOrigins {
        if origin == allowedOrigin {
            allowed = true
            break
        }
    }
    
    if allowed {
        c.Response.Header().Set("Access-Control-Allow-Origin", origin)
        c.Response.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
        c.Response.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
    }
    
    // Handle preflight requests
    if c.Request.Method == "OPTIONS" {
        c.Response.WriteHeader(http.StatusOK)
        return
    }
    
    next(c)
}
```

## Error Handling Middleware

```go
type ErrorHandlingMiddleware struct{}

func (m *ErrorHandlingMiddleware) Invoke(c *mux.RouteContext, next mux.HandlerFunc) {
    defer func() {
        if err := recover(); err != nil {
            log.Printf("Panic in handler: %v", err)
            
            // Return proper error response
            c.Response.Header().Set("Content-Type", "application/json")
            c.Response.WriteHeader(http.StatusInternalServerError)
            
            errorResponse := map[string]interface{}{
                "error": "Internal server error",
                "code":  500,
            }
            
            json.NewEncoder(c.Response).Encode(errorResponse)
        }
    }()
    
    next(c)
}
```

## Database Transaction Middleware

```go
type TransactionMiddleware struct {
    DB *sql.DB
}

func (m *TransactionMiddleware) Invoke(c *mux.RouteContext, next mux.HandlerFunc) {
    tx, err := m.DB.Begin()
    if err != nil {
        c.ServerError("Database Error", err.Error())
        return
    }
    
    // Store transaction in context
    c.SetService("tx", tx)
    
    // Track if we need to commit or rollback
    shouldCommit := true
    
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
            panic(r) // Re-panic to let other middleware handle it
        } else if shouldCommit {
            tx.Commit()
        } else {
            tx.Rollback()
        }
    }()
    
    // Wrap response writer to detect errors
    originalWriter := c.Response
    c.Response = &transactionResponseWriter{
        ResponseWriter: originalWriter,
        onError: func() {
            shouldCommit = false
        },
    }
    
    next(c)
}

type transactionResponseWriter struct {
    http.ResponseWriter
    onError func()
}

func (w *transactionResponseWriter) WriteHeader(statusCode int) {
    if statusCode >= 400 {
        w.onError()
    }
    w.ResponseWriter.WriteHeader(statusCode)
}
```

## Testing Custom Middleware

```go
func TestTimingMiddleware(t *testing.T) {
    // Arrange
    middleware := &TimingMiddleware{}
    req := httptest.NewRequest(http.MethodGet, "/test", nil)
    recorder := httptest.NewRecorder()
    ctx := mux.NewRouteContext(recorder, req)
    
    nextCalled := false
    next := func(c *mux.RouteContext) {
        nextCalled = true
        time.Sleep(10 * time.Millisecond) // Simulate work
    }
    
    // Act
    middleware.Invoke(ctx, next)
    
    // Assert
    assert.True(t, nextCalled)
    // Check logs for timing information
}
```

## Middleware Order

Middleware executes in the order it's added to the router:

```go
router.Use(&SecurityHeadersMiddleware{})  // 1st
router.Use(&LoggingMiddleware{})          // 2nd
router.Use(&AuthenticationMiddleware{})   // 3rd
router.Use(&AuthorizationMiddleware{})    // 4th
router.Use(&TimingMiddleware{})           // 5th (closest to handler)
```

Request flow: 1 → 2 → 3 → 4 → 5 → Handler → 5 → 4 → 3 → 2 → 1

## Best Practices

1. **Keep middleware focused** - Each middleware should have a single responsibility
2. **Handle errors gracefully** - Always provide fallback behavior
3. **Be mindful of order** - Authentication before authorization, logging early
4. **Use context for data** - Store middleware data in RouteContext services
5. **Test thoroughly** - Unit test your middleware logic
6. **Document configuration** - Make options clear and well-documented
7. **Consider performance** - Avoid heavy operations in middleware
8. **Handle panics** - Use defer/recover for critical middleware
9. **Respect the chain** - Always call `next()` unless terminating the request
10. **Clean up resources** - Use defer for cleanup operations

## Common Middleware Patterns

### Pre-processing
Execute logic before the handler:
```go
func (m *Middleware) Invoke(c *mux.RouteContext, next mux.HandlerFunc) {
    // Pre-processing logic
    validateRequest(c)
    
    next(c)
}
```

### Post-processing
Execute logic after the handler:
```go
func (m *Middleware) Invoke(c *mux.RouteContext, next mux.HandlerFunc) {
    next(c)
    
    // Post-processing logic
    logResponse(c)
}
```

### Wrapping
Execute logic before and after:
```go
func (m *Middleware) Invoke(c *mux.RouteContext, next mux.HandlerFunc) {
    // Before
    start := time.Now()
    
    next(c)
    
    // After
    duration := time.Since(start)
    recordMetrics(duration)
}
```

### Short-circuiting
Terminate the request chain:
```go
func (m *Middleware) Invoke(c *mux.RouteContext, next mux.HandlerFunc) {
    if !isAuthorized(c) {
        c.Unauthorized()
        return // Don't call next()
    }
    
    next(c)
}
```