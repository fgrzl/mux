# HTTP Redirect Examples

This example demonstrates the redirect helpers available on `mux.RouteContext`.

## Redirect Types

### 301 - Moved Permanently

```go
router.GET("/old-api", func(c mux.RouteContext) {
    c.MovedPermanently("/api/v2")
})
```

Use when a URL has moved for good.

### 302 - Found

```go
router.GET("/login", func(c mux.RouteContext) {
    c.Found("/auth/login")
})
```

Use for temporary redirects like login flows or maintenance pages.

### 303 - See Other

```go
router.POST("/submit", func(c mux.RouteContext) {
    c.SeeOther("/result?id=123")
})
```

Use after a POST when the client should follow with a GET.

### 307 - Temporary Redirect

```go
router.POST("/api/v1/users", func(c mux.RouteContext) {
    c.TemporaryRedirect("/api/v2/users")
})
```

Use when the redirect is temporary and the HTTP method must be preserved.

### 308 - Permanent Redirect

```go
router.POST("/old-webhook", func(c mux.RouteContext) {
    c.PermanentRedirect("/webhooks/v2")
})
```

Use when the redirect is permanent and the HTTP method must be preserved.

## Run It

```bash
go run .
```

The server listens on `http://localhost:8080`.

## Try It

```bash
curl -i http://localhost:8080/old-api
curl -i http://localhost:8080/login
curl -i -X POST http://localhost:8080/submit -H "Content-Type: application/json" -d '{"data":"test"}'
curl -i -X POST http://localhost:8080/api/v1/users -H "Content-Type: application/json" -d '{"name":"test"}'
curl -i -X POST http://localhost:8080/old-webhook -H "Content-Type: application/json" -d '{"event":"test"}'
```

## Documenting Redirects

Redirects are documented with the normal route-builder response helpers:

```go
router.GET("/old-page", func(c mux.RouteContext) {
    c.Found("/new-page")
}).
    WithSummary("Old page endpoint").
    WithFoundResponse()
```

Most applications should register redirect routes directly inside `router.Configure(...)`.

Common redirect response declarations:
- `WithMovedPermanentlyResponse()`
- `WithFoundResponse()`
- `WithSeeOtherResponse()`
- `WithTemporaryRedirectResponse()`
- `WithPermanentRedirectResponse()`
