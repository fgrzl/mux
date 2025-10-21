# HTTP Redirect Examples

This example demonstrates all the HTTP redirect methods available in the mux router.

## Redirect Types

### 301 - Moved Permanently
Used when a resource has permanently moved to a new location.

```go
router.GET("/old-api", func(c mux.RouteContext) {
    c.MovedPermanently("/api/v2")
})
```

**Use case**: Old URLs that should never be used again, SEO-friendly permanent redirects.

### 302 - Found (Temporary Redirect)
The most common redirect. Temporarily redirects to another URL.

```go
router.GET("/login", func(c mux.RouteContext) {
    c.Found("/auth/login")
})
```

**Use case**: Login redirects, temporary maintenance pages, A/B testing.

### 303 - See Other
Used after POST requests to redirect to a GET page (POST-Redirect-GET pattern).

```go
router.POST("/submit", func(c mux.RouteContext) {
    // Process form submission...
    c.SeeOther("/result?id=123")
})
```

**Use case**: Form submissions, preventing duplicate form submissions on refresh.

### 307 - Temporary Redirect (Method Preserved)
Like 302 but guarantees the HTTP method is preserved.

```go
router.POST("/api/v1/users", func(c mux.RouteContext) {
    c.TemporaryRedirect("/api/v2/users")  // POST method preserved
})
```

**Use case**: API versioning, load balancing, maintaining POST/PUT/DELETE methods.

### 308 - Permanent Redirect (Method Preserved)
Like 301 but guarantees the HTTP method is preserved.

```go
router.POST("/old-webhook", func(c mux.RouteContext) {
    c.PermanentRedirect("/webhooks/v2")  // POST method preserved
})
```

**Use case**: API migration where you need to preserve HTTP methods permanently.

## Running the Example

```bash
cd examples/redirects
go run main.go
```

The server will start on `http://localhost:8080`.

## Testing the Redirects

### Test 301 Redirect
```bash
curl -i http://localhost:8080/old-api
```

### Test 302 Redirect
```bash
curl -i http://localhost:8080/login
```

### Test 303 Redirect (POST->GET)
```bash
curl -i -X POST http://localhost:8080/submit \
  -H "Content-Type: application/json" \
  -d '{"data":"test"}'
```

### Test 307 Redirect (POST preserved)
```bash
curl -i -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"test"}'
```

### Test 308 Redirect (POST preserved)
```bash
curl -i -X POST http://localhost:8080/old-webhook \
  -H "Content-Type: application/json" \
  -d '{"event":"test"}'
```

## OpenAPI Documentation

You can also document redirects in your route builders:

```go
import "github.com/fgrzl/mux/pkg/builder"

router.Handle(
    builder.Route(http.MethodGet, "/old-page").
        WithSummary("Old page endpoint").
        With302Response().  // Document the redirect
        Options(),
    func(c mux.RouteContext) {
        c.Found("/new-page")
    },
)
```

Available builder methods:
- `With301Response()` - Moved Permanently
- `With302Response()` - Found
- `With303Response()` - See Other
- `With307Response()` - Temporary Redirect
- `With308Response()` - Permanent Redirect

## Quick Reference

| Status | Method | Builder | Preserves Method | Permanent |
|--------|--------|---------|------------------|-----------|
| 301 | `MovedPermanently()` | `With301Response()` | No | Yes |
| 302 | `Found()` | `With302Response()` | No | No |
| 303 | `SeeOther()` | `With303Response()` | No (→GET) | No |
| 307 | `TemporaryRedirect()` | `With307Response()` | Yes | No |
| 308 | `PermanentRedirect()` | `With308Response()` | Yes | Yes |

## Best Practices

1. **Use 302 for most cases** - It's the most widely supported
2. **Use 303 after POST** - Prevents form resubmission
3. **Use 307/308 for APIs** - When you need to preserve the HTTP method
4. **Use 301 for SEO** - When content has permanently moved
5. **Always use absolute URLs** - The router will handle relative paths automatically
