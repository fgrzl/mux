# Mux

[![CI](https://github.com/fgrzl/mux/actions/workflows/ci.yml/badge.svg)](https://github.com/fgrzl/mux/actions/workflows/ci.yml)
[![Dependabot](https://github.com/fgrzl/mux/actions/workflows/dependabot/dependabot-updates/badge.svg)](https://github.com/fgrzl/mux/actions/workflows/dependabot/dependabot-updates)

A lightweight, modular HTTP router for Go with middleware, request binding, OpenAPI 3.1 generation, structured responses, and flexible auth support.

## Why Mux?

✨ **API-first Development** - Built-in OpenAPI support without code generation  
🧩 **Modular & Extensible** - Composable middleware and lifecycle hooks  
💡 **Ergonomic DSL** - Intuitive route definition and documentation  
🧪 **Clear & Testable** - Type-safe design with predictable behavior  

## Quick Start

### Installation

```bash
go get github.com/fgrzl/mux
```

### Hello World

```go
package main

import (
    "net/http"
    "github.com/fgrzl/mux"
)

func main() {
    router := mux.NewRouter()
    
    router.GET("/hello", func(c *mux.RouteContext) {
        c.OK("Hello, World!")
    })
    
    http.ListenAndServe(":8080", router)
}
```

### Simple API

```go
package main

import (
    "net/http"
    "github.com/fgrzl/mux"
    "github.com/google/uuid"
)

type User struct {
    ID    uuid.UUID `json:"id"`
    Name  string    `json:"name"`
    Email string    `json:"email"`
}

func main() {
    router := mux.NewRouter()
    
    // Add basic middleware
    router.UseLogging()
    router.UseCompression()
    
    // Create API routes
    api := router.NewRouteGroup("/api/v1")
    api.WithTags("API v1")
    
    // User routes with OpenAPI documentation
    users := api.NewRouteGroup("/users")
    users.WithTags("Users")
    
    users.GET("/", listUsers).
        WithSummary("List all users").
        WithOKResponse([]User{})
        
    users.POST("/", createUser).
        WithSummary("Create a new user").
        WithJsonBody(User{}).
        WithCreatedResponse(User{})
        
    users.GET("/{id}", getUser).
        WithSummary("Get a user by ID").
        WithParam("id", "path", uuid.Nil, true).
        WithOKResponse(User{})
    
    http.ListenAndServe(":8080", router)
}

func listUsers(c *mux.RouteContext) {
    users := []User{
        {ID: uuid.New(), Name: "John Doe", Email: "john@example.com"},
        {ID: uuid.New(), Name: "Jane Smith", Email: "jane@example.com"},
    }
    c.OK(users)
}

func createUser(c *mux.RouteContext) {
    var user User
    if err := c.Bind(&user); err != nil {
        c.BadRequest("Invalid request", err.Error())
        return
    }
    
    user.ID = uuid.New()
    c.Created(user)
}

func getUser(c *mux.RouteContext) {
    userID, ok := c.ParamUUID("id")
    if !ok {
        c.BadRequest("Invalid user ID", "")
        return
    }
    
    user := User{
        ID:    userID,
        Name:  "John Doe",
        Email: "john@example.com",
    }
    c.OK(user)
}
```

## Key Features

- **🛣️ Route Management**: Flexible patterns, parameter binding, route groups
- **🔧 Middleware System**: Built-in and custom middleware support
- **📝 Request Binding**: Automatic data collection from multiple sources
- **📤 Response Helpers**: Structured responses for common HTTP status codes
- **🔐 Authentication**: JWT-based auth with role-based access control
- **⚡ Rate Limiting**: Per-route token bucket rate limiting
- **📖 OpenAPI 3.1**: Automatic spec generation with inline documentation
- **🌍 Geographic Control**: Export control with GeoIP support
- **📊 Observability**: OpenTelemetry integration and structured logging

## Basic Usage

### Route Definition

```go
router := mux.NewRouter()

// HTTP methods
router.GET("/users", listUsers)
router.POST("/users", createUser)
router.PUT("/users/{id}", updateUser)
router.DELETE("/users/{id}", deleteUser)

// Route groups
api := router.NewRouteGroup("/api/v1")
api.GET("/health", healthCheck)
```

### Request Handling

```go
func createUser(c *mux.RouteContext) {
    // Bind request data to struct
    var user User
    if err := c.Bind(&user); err != nil {
        c.BadRequest("Invalid data", err.Error())
        return
    }
    
    // Access individual parameters
    orgID, _ := c.QueryUUID("org_id")
    includeDetails, _ := c.QueryBool("include_details")
    
    // Create user and respond
    createdUser := service.CreateUser(user)
    c.Created(createdUser)
}
```

### Middleware

```go
// Built-in middleware
router.UseLogging()
router.UseCompression()
router.UseEnforceHTTPS()

// Authentication
router.UseAuthentication(
    mux.WithValidator(validateToken),
    mux.WithTokenCreator(createToken),
)

// Custom middleware
router.Use(&CustomMiddleware{})
```

## Documentation

Comprehensive documentation to help you build production-ready APIs:

### 🚀 Getting Started
- [**Installation Guide**](docs/installation.md) - Setup requirements and installation
- [**Quick Start Tutorial**](docs/quick-start.md) - Build your first API in 10 steps
- [**Examples**](examples/) - Working example applications

### 📖 Core Documentation
- [**Overview**](docs/overview.md) - Architecture and core concepts
- [**Router**](docs/router.md) - Route definition and configuration
- [**Built-in Middleware Guide**](docs/middleware.md) - Complete middleware reference
- [**Authentication Middleware**](docs/authentication-middleware.md) - JWT auth setup
- [**Custom Middleware**](docs/custom-middleware.md) - Building custom middleware

### 📚 Advanced Topics
- [**Best Practices Guide**](docs/best-practices.md) - Production-ready patterns and conventions
- [**FAQ**](docs/faq.md) - Common questions and troubleshooting

### 🎯 Quick Links
- [API Reference](https://pkg.go.dev/github.com/fgrzl/mux) - Complete API documentation on pkg.go.dev
- [Examples Directory](examples/) - Working code examples
- [GitHub Issues](https://github.com/fgrzl/mux/issues) - Bug reports and feature requests

## OpenAPI Generation

Generate OpenAPI specifications from your routes:

```go
// Define API with documentation
router.POST("/users", createUser).
    WithOperationID("createUser").
    WithSummary("Create a user").
    WithJsonBody(User{}).
    WithCreatedResponse(User{}).
    WithBadRequestResponse().
    WithTags("Users")

// Generate spec
generator := mux.NewGenerator()
spec := generator.GenerateSpec(router)
spec.MarshalToFile("openapi.yaml")
```

## Testing

The library includes comprehensive test coverage and examples in the `test/` directory.

```bash
go test ./... -v
go test ./... -coverprofile=coverage.out
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Write tests for your changes
4. Ensure all tests pass (`go test ./...`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
