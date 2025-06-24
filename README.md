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

```go
type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

router.POST("/users", func(c *mux.RouteContext) {
	var user User
	if err := c.Bind(&user); err != nil {
		c.BadRequest("Invalid request", err.Error())
		return
	}
	c.Created(user)
})
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
import "github.com/google/uuid"

router.GET("/api/v1/resources", listResources).
  WithOperationID("listResources").
  WithSummary("List all resources").
  WithOKResponse([]Resource{}).
  WithNotFoundResponse().
  WithTags("Resources")

router.GET("/api/v1/resources/{resourceID}", getResource).
  WithOperationID("getResource").
  WithSummary("Get a resource").
  WithParam("resourceID", "path", uuid.Nil, true).
  WithOKResponse(Resource{}).
  WithNotFoundResponse().
  WithTags("Resources")

router.POST("/api/v1/resources", createResource).
  WithOperationID("createResource").
  WithSummary("Create a resource").
  WithJsonBody(Resource{}).
  WithCreatedResponse(Resource{}).
  WithBadRequestResponse().
  WithTags("Resources")

router.PUT("/api/v1/resources/{resourceID}", updateResource).
  WithOperationID("updateResource").
  WithSummary("Update a resource").
  WithParam("resourceID", "path", uuid.Nil, true).
  WithJsonBody(Resource{}).
  WithOKResponse(Resource{}).
  WithBadRequestResponse().
  WithTags("Resources")

router.DELETE("/api/v1/resources/{resourceID}", deleteResource).
  WithOperationID("deleteResource").
  WithSummary("Delete a resource").
  WithParam("resourceID", "path", uuid.Nil, true).
  WithNoContentResponse().
  WithNotFoundResponse().
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

```bash
go run main.go
