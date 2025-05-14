[![CI](https://github.com/fgrzl/mux/actions/workflows/ci.yml/badge.svg)](https://github.com/fgrzl/mux/actions/workflows/ci.yml)
[![Dependabot](https://github.com/fgrzl/mux/actions/workflows/dependabot/dependabot-updates/badge.svg)](https://github.com/fgrzl/mux/actions/workflows/dependabot/dependabot-updates)

# Mux

A lightweight HTTP router for Go with middleware support, authentication, structured responses, and flexible request binding.

## Installation

```sh
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

---

## Features

* Route grouping with prefixes
* Middleware chaining
* Request binding
* Built-in helpers for common HTTP responses
* Structured error handling
* Authentication & Authorization hooks

---

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

---

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

func (m *LoggingMiddleware) Invoke(ctx *mux.RouteContext, next mux.HandlerFunc) {
	fmt.Println("Request:", ctx.Request.URL.Path)
	next(ctx)
}

router.Use(&LoggingMiddleware{})
```

---

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

---

## Authentication

```go
router.GET("/private", func(c *mux.RouteContext) {
	if !c.User.IsAuthenticated() {
		c.Unauthorized()
		return
	}
	c.OK(map[string]string{"message": "Access granted"})
})
```

---

## Redirects

```go
router.GET("/old-page", func(c *mux.RouteContext) {
	c.PermanentRedirect("/new-page")
})
```

---

## Not Found Handling

Unmatched routes automatically return a `404 Not Found` response.

---

## Route Methods

* `HEAD(pattern, handler)`
* `GET(pattern, handler)`
* `POST(pattern, handler)`
* `PUT(pattern, handler)`
* `DELETE(pattern, handler)`

Each returns a `*RouteBuilder` for optional chaining (e.g., `.AllowAnonymous()` or `.WithRateLimit(...)`).

---

## Development

```sh
go run main.go
```
