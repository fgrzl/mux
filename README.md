[![CI](https://github.com/fgrzl/mux/actions/workflows/ci.yml/badge.svg)](https://github.com/fgrzl/mux/actions/workflows/ci.yml)
[![Dependabot](https://github.com/fgrzl/mux/actions/workflows/dependabot/dependabot-updates/badge.svg)](https://github.com/fgrzl/mux/actions/workflows/dependabot/dependabot-updates)

# Mux

A lightweight, modular HTTP router for Go with middleware, request binding, OpenAPI 3.1 generation, structured responses, and flexible auth support.

## Installation

```bash
go get github.com/fgrzl/mux
```

## Quick Start

```go
package main

import (
	"net/http"
	"github.com/fgrzl/mux"
)

func main() {
	router := mux.NewRouter()
	
	// Create a route group with prefix
	api := router.NewRouteGroup("/api/v1")

	api.GET("/hello", func(c *mux.RouteContext) {
		c.OK("Hello, world!")
	})

	http.ListenAndServe(":8080", router)
}
```

## Features

- **Route Management**: Route grouping with prefixes, parameter binding, and flexible patterns
- **Middleware System**: Modular middleware with built-in options for logging, compression, auth, rate limiting, and more
- **Request Binding**: Automatic data collection from query parameters, body, headers, and path parameters
- **Response Helpers**: Structured response helpers for common HTTP status codes
- **Authentication & Authorization**: JWT-based auth with role-based access control and permissions
- **Rate Limiting**: Per-route token bucket rate limiting
- **OpenAPI 3.1**: Automatic spec generation with inline documentation
- **Geographic Control**: Export control middleware with GeoIP database support
- **Observability**: OpenTelemetry integration and structured logging

## Project Structure

This project follows a flat file structure for simplicity and clarity:

```
mux/
├── README.md                          # Project documentation
├── go.mod                            # Go module definition
├── LICENSE                           # MIT license
├── .github/                          # GitHub Actions workflows
│   ├── workflows/
│   │   ├── ci.yml                   # Build and test pipeline
│   │   └── pre-release.yml          # Release automation
│   └── dependabot.yml               # Dependency updates
├── test/                             # Test files and examples
│   ├── router_test.go               # Integration tests  
│   ├── test_*.go                    # Test utilities and models
│   └── openapi.yaml                 # Generated OpenAPI spec
├── router*.go                        # Core router and options
├── route_*.go                        # Route handling and context
├── middleware_*.go                   # Built-in middleware components
├── openapi_*.go                      # OpenAPI generation
├── auth_provider.go                  # Authentication provider
├── cookie_jar.go                     # Cookie utilities
├── problem_detail.go                 # RFC 7807 problem details
└── server.go                         # HTTP server integration
```

**Key Design Principles:**
- **Flat structure**: All core functionality in the root package for simplicity
- **Middleware modularity**: Each middleware in its own file with options pattern
- **Test separation**: Tests isolated in `/test` directory with example implementations
- **Zero dependencies**: Minimal external dependencies for core functionality

## Defining Routes

```go
router := mux.NewRouter()
api := router.NewRouteGroup("/api")

api.GET("/users", func(c *mux.RouteContext) {
	c.OK(map[string]string{"message": "GET users"})
})

api.POST("/users", func(c *mux.RouteContext) {
	c.Created(map[string]string{"message": "User created"})
})
```

## Middleware

### Built-in Middleware

```go
// Logging - structured request/response logging
router.UseLogging() // optionally: router.UseLogging( /* LoggingOption... */ )

// Compression - gzip/deflate response compression
router.UseCompression() // optionally: router.UseCompression( /* CompressionOption... */ )

// Authentication - JWT token validation and creation
router.UseAuthentication(
  mux.WithValidator(validateToken),
  mux.WithTokenCreator(createToken),
  mux.WithTokenTTL(30 * time.Minute),
)

// Authorization - role-based access control
router.UseAuthorization(
  mux.WithRoles("admin", "user"),
  mux.WithPermissions("tenant:{tenantID}:read"),
  mux.WithPermissionChecker(checkPermissions),
)

// HTTPS enforcement - redirect HTTP to HTTPS
router.UseEnforceHTTPS()

// Forwarded headers - parse X-Forwarded-* headers  
router.UseForwardedHeaders()

// Export control - geographic access restrictions
router.UseExportControl(mux.WithGeoIPDatabase(geoipDB))

// OpenTelemetry - distributed tracing and metrics
router.UseOpenTelemetry()
```

### Rate Limiting

Rate limiting is configured per-route using a token bucket algorithm:

```go
router.GET("/api/data", handler).
  WithRateLimit(100, time.Minute) // 100 requests per minute
```

### Custom

```go
type SimpleLogger struct{}

func (m *SimpleLogger) Invoke(c *mux.RouteContext, next mux.HandlerFunc) {
	start := time.Now()
	rec := &statusRecorder{ResponseWriter: c.Response}
	c.Response = rec

	next(c)

	log.Printf("[%s] %s %d (%s)",
		c.Request.Method,
		c.Request.URL.Path,
		rec.Status,
		time.Since(start),
	)
}

type statusRecorder struct {
	http.ResponseWriter
	Status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.Status = code
	r.ResponseWriter.WriteHeader(code)
}

// Register custom middleware
router.Use(&SimpleLogger{})
```

## Request Binding

It collects input data from:

- Query Parameters (GET, HEAD, DELETE)
- Request Body (POST, PUT):
  - JSON (application/json)
  - Form (application/x-www-form-urlencoded)
- Headers
- Path Parameters (RouteParams)

Then it:

- Normalizes all values into a map[string]any
- Marshals it into JSON
- Unmarshals it into the user-defined model

This allows a struct to receive values from multiple sources in a single pass.

```go
type User struct {
  ID    uuid.UUID `json:"id"`
  Name  string    `json:"name"`
  Email string    `json:"email"`
}

router.PUT("/users/{id}", func(c *mux.RouteContext) {
  var user User
  if err := c.Bind(&user); err != nil {
    c.BadRequest("Invalid request", err.Error())
    return
  }
  c.OK(user)
}).
  WithOperationID("updateUser").
  WithSummary("Update a user").
  WithParam("id", "path", uuid.Nil, true).
  WithJsonBody(User{}).
  WithOKResponse(User{}).
  WithBadRequestResponse().
  WithTags("Users")
```

## Authentication Example

```go
router.GET("/private", func(c *mux.RouteContext) {
	if !c.User.IsAuthenticated() {
		c.Unauthorized()
		return
	}
	c.OK(map[string]string{"message": "Access granted"})
})
```

## Redirects

```go
router.GET("/old-page", func(c *mux.RouteContext) {
	c.PermanentRedirect("/new-page")
})
```

## Not Found Handling

Unmatched routes automatically return `404 Not Found`.

## OpenAPI 3.1 DSL Example

```go
router.GET("/api/v1/resources", listResources).
  WithOperationID("listResources").
  WithSummary("List all resources").
  WithOKResponse([]*Resource{}).
  WithStandardErrors().
  WithTags("Resources")

router.GET("/api/v1/resources/{resourceID}", getResource).
  WithOperationID("getResource").
  WithSummary("Get a resource").
  WithParam("resourceID", "path", uuid.Nil, true).
  WithOKResponse(*Resource{}).
  WithStandardErrors().
  WithTags("Resources")

router.POST("/api/v1/resources", createResource).
  WithOperationID("createResource").
  WithSummary("Create a resource").
  WithJsonBody(*Resource{}).
  WithCreatedResponse(*Resource{}).
  WithBadRequestResponse().
  WithTags("Resources")

router.PUT("/api/v1/resources/{resourceID}", updateResource).
  WithOperationID("updateResource").
  WithSummary("Update a resource").
  WithParam("resourceID", "path", uuid.Nil, true).
  WithJsonBody(*Resource{}).
  WithOKResponse(*Resource{}).
  WithStandardErrors().
  WithTags("Resources")

router.DELETE("/api/v1/resources/{resourceID}", deleteResource).
  WithOperationID("deleteResource").
  WithSummary("Delete a resource").
  WithParam("resourceID", "path", uuid.Nil, true).
  WithNoContentResponse().
  WithStandardErrors().
  WithTags("Resources")
```

## Route Methods

Each method returns a `*RouteBuilder` to allow chaining:

- `HEAD(pattern, handler)`
- `GET(pattern, handler)`
- `POST(pattern, handler)`
- `PUT(pattern, handler)`
- `DELETE(pattern, handler)`

## Development

### Setup

```bash
# Clone the repository
git clone https://github.com/fgrzl/mux.git
cd mux

# Install dependencies  
go mod tidy

# Build the project
go build ./...

# Run tests
go test ./... -v

# Run tests with coverage
go test ./... -v -coverprofile=coverage.out
```

### Testing

The project includes comprehensive tests in the `test/` directory:

- **Integration tests**: Full HTTP request/response testing in `router_test.go`
- **Example implementations**: Reference implementations in `test_*.go` files  
- **OpenAPI validation**: Automatic spec generation testing

Test coverage includes all middleware, routing, authentication, and OpenAPI generation.

### Code Conventions

- **Go standards**: Follow standard Go conventions (gofmt, go vet)
- **Options pattern**: Use functional options for configurable components
- **Middleware interface**: Implement `Invoke(c *RouteContext, next HandlerFunc)`
- **Builder pattern**: Chainable methods for route configuration
- **Error handling**: Use structured error responses with problem details (RFC 7807)

### Contributing

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Write** tests for your changes
4. **Ensure** all tests pass (`go test ./...`)
5. **Format** code (`go fmt ./...`)
6. **Commit** changes (`git commit -m 'Add amazing feature'`)
7. **Push** to branch (`git push origin feature/amazing-feature`)
8. **Open** a Pull Request

### CI/CD

The project uses GitHub Actions for:

- **Continuous Integration**: Build and test on every push/PR
- **Dependency Updates**: Automated Dependabot updates
- **Pre-release**: Automated versioning and releases

Build requirements:
- **Go**: 1.24.x or later
- **Platform**: Linux, macOS, Windows supported

## Why Use `fgrzl/mux`?

`fgrzl/mux` is a developer-focused HTTP router designed to be small, explicit, and modern. It aims to simplify the most common patterns for building HTTP APIs in Go — while keeping you in full control.

### ✨ Built for API-first Development

- **First-class OpenAPI support**: Define routes and documentation in one step using the DSL.
- **No codegen required**: All specs are derived directly from your actual routes and handlers.
- **Schema-aware parameter typing**: Define `uuid.UUID`, `int`, `bool`, etc. directly in `.WithParam(...)`.
- **Inline request/response modeling**: Document your APIs as you build them.

### 🧩 Modular and Extensible

- Composable middleware and lifecycle hooks
- Custom auth and permission logic with fallback support
- Small, readable core built for real-world APIs

### 💡 Ergonomic DSL

```go
router.POST("/users", createUser).
  WithJsonBody(User{}).
  WithCreatedResponse(User{}).
  WithBadRequestResponse().
  WithTags("Users")
```

### 🧪 Clear and Testable

- Type-safe design avoids reflection pitfalls
- Handlers remain concise and decoupled
- Predictable runtime behavior with minimal magic

Whether you're building internal tools or public APIs, `mux` provides the building blocks for well-structured, maintainable, and self-documenting Go services.
