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
	router := mux.NewRouter("/api/v1")

	router.GET("/hello", func(c *mux.RouteContext) {
		c.OK("Hello, world!")
	})

	http.ListenAndServe(":8080", router)
}
```

## Features

- Route grouping with prefixes
- Middleware chaining
- Request binding
- Structured response helpers
- Error and redirect handling
- Auth & role-based access control
- OpenAPI 3.1 generation

## Defining Routes

```go
router := mux.NewRouter("/api")

router.GET("/users", func(c *mux.RouteContext) {
	c.OK(map[string]string{"message": "GET users"})
})

router.POST("/users", func(c *mux.RouteContext) {
	c.Created(map[string]string{"message": "User created"})
})
```

## Middleware

### Built-in

```go
router.UseLogging(&mux.LoggingOptions{})
router.UseCompression(&mux.CompressionOptions{})
router.UseAuthentication(&mux.AuthenticationOptions{})
router.UseAuthorization(&mux.AuthorizationOptions{})
```

### Custom

```go
type LoggingMiddleware struct{}

func (m *LoggingMiddleware) Invoke(c *mux.RouteContext, next mux.HandlerFunc) {
	fmt.Println("Request:", c.Request.URL.Path)
	next(c)
}

router.Use(&LoggingMiddleware{})
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

---

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
