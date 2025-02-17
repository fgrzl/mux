[![ci](https://github.com/fgrzl/mux/actions/workflows/ci.yml/badge.svg)](https://github.com/fgrzl/mux/actions/workflows/ci.yml)
[![Dependabot Updates](https://github.com/fgrzl/mux/actions/workflows/dependabot/dependabot-updates/badge.svg)](https://github.com/fgrzl/mux/actions/workflows/dependabot/dependabot-updates)

# Mux Package

The `mux` package provides a lightweight HTTP router for handling requests with middleware support. It includes features for route handling, parameter binding, authentication, and error handling.

## Installation

```sh
go get github.com/yourusername/mux
```

## Usage

### Creating a Router

```go
package main

import (
	"fmt"
	"net/http"
	"github.com/yourusername/mux"
)

func main() {
	router := mux.NewRouter("/api/v1")

	router.GET("/hello", func(c *mux.RouteContext) {
		c.OK("Hello, world!")
	})

	http.ListenAndServe(":8080", router)
}
```

### Defining Routes

```go
router := mux.NewRouter("/api")

router.GET("/users", func(c *mux.RouteContext) {
	c.OK(map[string]string{"message": "GET users"})
})

router.POST("/users", func(c *mux.RouteContext) {
	c.Created(map[string]string{"message": "User created"})
})
```

### Using Middleware

Middleware functions are used to process requests before they reach the actual route handler. They allow for reusable components that can handle tasks such as logging, authentication, and request modification.

```go
type LoggingMiddleware struct{}

func (m *LoggingMiddleware) Invoke(ctx *mux.RouteContext, next mux.HandlerFunc) {
	fmt.Println("Request received for", ctx.Request.URL.Path)
	next(ctx)
}

router := mux.NewRouter("/api")
router.Use(&LoggingMiddleware{})
```

Middleware executes in the order they are added to the router. Each middleware function can call `next(ctx)` to pass execution to the next middleware or the final handler.

### Handling Requests and Responses

The `RouteContext` provides helper methods for handling responses and error states:

```go
router.GET("/data", func(c *mux.RouteContext) {
	if data, err := fetchData(); err != nil {
		c.ServerError("Fetch Error", "Unable to retrieve data")
	} else {
		c.OK(data)
	}
})
```

### Binding Request Data

The `Bind` method allows automatic deserialization of request bodies:

```go
type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

router.POST("/users", func(c *mux.RouteContext) {
	var user User
	if err := c.Bind(&user); err != nil {
		c.BadRequest("Invalid Data", err.Error())
		return
	}
	c.Created(user)
})
```

### Handling Authentication

```go
router.GET("/private", func(c *mux.RouteContext) {
	if !c.User.IsAuthenticated() {
		c.Unauthorized()
		return
	}
	c.OK(map[string]string{"message": "Welcome!"})
})
```

### Handling Redirects

```go
router.GET("/old-page", func(c *mux.RouteContext) {
	c.PermanentRedirect("/new-page")
})
```

### Handling 404 Not Found

If a request does not match any registered route, the router automatically responds with a 404 status.

## Route Registration Methods

- `HEAD(pattern string, handler HandlerFunc) *RouteBuilder`
- `GET(pattern string, handler HandlerFunc) *RouteBuilder`
- `POST(pattern string, handler HandlerFunc) *RouteBuilder`
- `PUT(pattern string, handler HandlerFunc) *RouteBuilder`
- `DELETE(pattern string, handler HandlerFunc) *RouteBuilder`

## Running the Server

```sh
go run main.go
```
