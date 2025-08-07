# Quick Start Tutorial

This tutorial will walk you through building your first API with Mux in just a few minutes. You'll create a simple user management API with authentication, validation, and OpenAPI documentation.

## Prerequisites

- Go 1.24.4 or later installed
- Basic familiarity with Go and HTTP APIs
- Text editor or IDE

## Step 1: Project Setup

Create a new project directory and initialize a Go module:

```bash
mkdir my-first-api
cd my-first-api
go mod init my-first-api
```

Install Mux:

```bash
go get github.com/fgrzl/mux
go get github.com/google/uuid  # For UUID support
```

## Step 2: Hello World

Create `main.go` with a simple endpoint:

```go
package main

import (
    "net/http"
    "github.com/fgrzl/mux"
)

func main() {
    // Create a new router
    router := mux.NewRouter()
    
    // Add a simple endpoint
    router.GET("/hello", func(c *mux.RouteContext) {
        c.OK("Hello, World!")
    })
    
    // Start the server
    http.ListenAndServe(":8080", router)
}
```

Run your application:

```bash
go run main.go
```

Test it:

```bash
curl http://localhost:8080/hello
# Output: "Hello, World!"
```

🎉 Congratulations! Your first Mux API is running.

## Step 3: Add Basic Middleware

Let's add logging and compression middleware for better development experience:

```go
package main

import (
    "net/http"
    "github.com/fgrzl/mux"
)

func main() {
    router := mux.NewRouter()
    
    // Add middleware
    router.UseLogging()      // Log all requests
    router.UseCompression()  // Compress responses
    
    router.GET("/hello", func(c *mux.RouteContext) {
        c.OK("Hello, World!")
    })
    
    http.ListenAndServe(":8080", router)
}
```

Now when you make requests, you'll see structured logs in your console.

## Step 4: Create a User Model

Let's build something more interesting. Create a user management API:

```go
package main

import (
    "net/http"
    "time"
    "github.com/fgrzl/mux"
    "github.com/google/uuid"
)

// User represents a user in our system
type User struct {
    ID       uuid.UUID `json:"id"`
    Name     string    `json:"name"`
    Email    string    `json:"email"`
    Created  time.Time `json:"created"`
}

// In-memory storage (don't use in production!)
var users = []User{
    {
        ID:      uuid.New(),
        Name:    "John Doe",
        Email:   "john@example.com",
        Created: time.Now().Add(-24 * time.Hour),
    },
    {
        ID:      uuid.New(),
        Name:    "Jane Smith", 
        Email:   "jane@example.com",
        Created: time.Now().Add(-12 * time.Hour),
    },
}

func main() {
    router := mux.NewRouter()
    router.UseLogging()
    router.UseCompression()
    
    // User endpoints
    router.GET("/users", listUsers)
    router.POST("/users", createUser)
    router.GET("/users/{id}", getUser)
    
    http.ListenAndServe(":8080", router)
}

// Handler functions
func listUsers(c *mux.RouteContext) {
    c.OK(users)
}

func createUser(c *mux.RouteContext) {
    var user User
    if err := c.Bind(&user); err != nil {
        c.BadRequest("Invalid request", err.Error())
        return
    }
    
    // Validate required fields
    if user.Name == "" || user.Email == "" {
        c.BadRequest("Missing required fields", "name and email are required")
        return
    }
    
    // Set server-side fields
    user.ID = uuid.New()
    user.Created = time.Now()
    
    // Add to storage
    users = append(users, user)
    
    c.Created(user)
}

func getUser(c *mux.RouteContext) {
    // Extract UUID from path parameter
    userID, ok := c.ParamUUID("id")
    if !ok {
        c.BadRequest("Invalid user ID", "user ID must be a valid UUID")
        return
    }
    
    // Find user
    for _, user := range users {
        if user.ID == userID {
            c.OK(user)
            return
        }
    }
    
    c.NotFound()
}
```

## Step 5: Test Your API

Now you can test all the endpoints:

### List all users
```bash
curl http://localhost:8080/users
```

### Create a new user
```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice Johnson", "email": "alice@example.com"}'
```

### Get a specific user
```bash
# Use an ID from the list users response
curl http://localhost:8080/users/123e4567-e89b-12d3-a456-426614174000
```

## Step 6: Add Route Groups and OpenAPI Documentation

Organize your routes and add API documentation:

```go
package main

import (
    "net/http"
    "time"
    "github.com/fgrzl/mux"
    "github.com/google/uuid"
)

type User struct {
    ID       uuid.UUID `json:"id"`
    Name     string    `json:"name"`
    Email    string    `json:"email"`
    Created  time.Time `json:"created"`
}

var users = []User{
    {ID: uuid.New(), Name: "John Doe", Email: "john@example.com", Created: time.Now()},
}

func main() {
    // Create router with API information
    router := mux.NewRouter(
        mux.WithTitle("My First API"),
        mux.WithVersion("1.0.0"),
        mux.WithDescription("A simple user management API built with Mux"),
    )
    
    router.UseLogging()
    router.UseCompression()
    
    // Create API v1 route group
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
        WithCreatedResponse(User{}).
        WithBadRequestResponse()
        
    users.GET("/{id}", getUser).
        WithSummary("Get a user by ID").
        WithParam("id", "path", uuid.Nil, true).
        WithOKResponse(User{}).
        WithNotFoundResponse()
    
    // Add health check endpoint
    router.GET("/health", func(c *mux.RouteContext) {
        c.OK(map[string]string{
            "status":    "healthy",
            "timestamp": time.Now().Format(time.RFC3339),
        })
    }).WithSummary("Health check").WithTags("Health")
    
    http.ListenAndServe(":8080", router)
}

// ... (handler functions remain the same)
func listUsers(c *mux.RouteContext) {
    c.OK(users)
}

func createUser(c *mux.RouteContext) {
    var user User
    if err := c.Bind(&user); err != nil {
        c.BadRequest("Invalid request", err.Error())
        return
    }
    
    if user.Name == "" || user.Email == "" {
        c.BadRequest("Missing required fields", "name and email are required")
        return
    }
    
    user.ID = uuid.New()
    user.Created = time.Now()
    users = append(users, user)
    
    c.Created(user)
}

func getUser(c *mux.RouteContext) {
    userID, ok := c.ParamUUID("id")
    if !ok {
        c.BadRequest("Invalid user ID", "user ID must be a valid UUID")
        return
    }
    
    for _, user := range users {
        if user.ID == userID {
            c.OK(user)
            return
        }
    }
    
    c.NotFound()
}
```

## Step 7: Generate OpenAPI Specification

Add OpenAPI spec generation to see your documented API:

```go
// Add this to your main function, before starting the server
func main() {
    // ... (router setup code)
    
    // Generate and save OpenAPI spec
    generator := mux.NewGenerator()
    spec := generator.GenerateSpec(router)
    spec.MarshalToFile("openapi.yaml")
    
    // Add endpoint to serve the spec
    router.GET("/openapi.yaml", func(c *mux.RouteContext) {
        c.Response.Header().Set("Content-Type", "application/yaml")
        spec.MarshalToWriter(c.Response)
    })
    
    http.ListenAndServe(":8080", router)
}
```

Now you can view your API specification:

```bash
curl http://localhost:8080/openapi.yaml
```

## Step 8: Add Error Handling and Validation

Improve your API with better error handling:

```go
func createUser(c *mux.RouteContext) {
    var user User
    if err := c.Bind(&user); err != nil {
        c.BadRequest("Invalid JSON", err.Error())
        return
    }
    
    // Validate input
    if user.Name == "" {
        c.BadRequest("Validation failed", "name is required")
        return
    }
    
    if user.Email == "" {
        c.BadRequest("Validation failed", "email is required")
        return
    }
    
    // Simple email validation
    if !strings.Contains(user.Email, "@") {
        c.BadRequest("Validation failed", "email must be a valid email address")
        return
    }
    
    // Check for duplicate email
    for _, existingUser := range users {
        if existingUser.Email == user.Email {
            c.Conflict("User already exists", "a user with this email already exists")
            return
        }
    }
    
    user.ID = uuid.New()
    user.Created = time.Now()
    users = append(users, user)
    
    c.Created(user)
}
```

## Step 9: Add Query Parameters

Let's add pagination and filtering to the list endpoint:

```go
func listUsers(c *mux.RouteContext) {
    // Get query parameters
    page, _ := c.QueryInt("page")      // Default: 0
    limit, _ := c.QueryInt("limit")    // Default: 0
    search, _ := c.QueryValue("search") // Default: ""
    
    // Set defaults
    if page < 1 {
        page = 1
    }
    if limit < 1 || limit > 100 {
        limit = 10
    }
    
    // Filter users
    filteredUsers := users
    if search != "" {
        filteredUsers = []User{}
        for _, user := range users {
            if strings.Contains(strings.ToLower(user.Name), strings.ToLower(search)) ||
               strings.Contains(strings.ToLower(user.Email), strings.ToLower(search)) {
                filteredUsers = append(filteredUsers, user)
            }
        }
    }
    
    // Paginate
    start := (page - 1) * limit
    end := start + limit
    
    if start >= len(filteredUsers) {
        c.OK([]User{})
        return
    }
    
    if end > len(filteredUsers) {
        end = len(filteredUsers)
    }
    
    result := filteredUsers[start:end]
    
    // Return paginated results with metadata
    response := map[string]interface{}{
        "users": result,
        "page":  page,
        "limit": limit,
        "total": len(filteredUsers),
    }
    
    c.OK(response)
}
```

Update the OpenAPI documentation:

```go
users.GET("/", listUsers).
    WithSummary("List all users").
    WithParam("page", "query", 1, false).
    WithParam("limit", "query", 10, false).
    WithParam("search", "query", "", false).
    WithOKResponse(map[string]interface{}{
        "users": []User{},
        "page":  1,
        "limit": 10,
        "total": 0,
    })
```

## Step 10: Test Your Complete API

Now you can test your enhanced API:

### List users with pagination
```bash
curl "http://localhost:8080/api/v1/users?page=1&limit=5"
```

### Search users
```bash
curl "http://localhost:8080/api/v1/users?search=john"
```

### Health check
```bash
curl http://localhost:8080/health
```

### View OpenAPI spec
```bash
curl http://localhost:8080/openapi.yaml
```

## What's Next?

You've built a complete API with Mux! Here are some next steps to explore:

### 1. Add Authentication
```go
router.UseAuthentication(
    mux.WithValidator(validateToken),
    mux.WithTokenCreator(createToken),
)
```

### 2. Add Rate Limiting
```go
router.POST("/api/v1/users", createUser).
    WithRateLimit(10, time.Minute)  // 10 requests per minute
```

### 3. Add Database Integration
Replace the in-memory storage with a real database using your preferred Go database library.

### 4. Add More Middleware
```go
router.UseEnforceHTTPS()     // Force HTTPS in production
router.UseOpenTelemetry()    // Add distributed tracing
```

### 5. Add Tests
Create tests for your handlers:

```go
func TestCreateUser(t *testing.T) {
    router := mux.NewRouter()
    router.POST("/users", createUser)
    
    // Test user creation
    // ...
}
```

## Complete Example

Here's the complete `main.go` for reference:

```go
package main

import (
    "net/http"
    "strings"
    "time"
    "github.com/fgrzl/mux"
    "github.com/google/uuid"
)

type User struct {
    ID       uuid.UUID `json:"id"`
    Name     string    `json:"name"`
    Email    string    `json:"email"`
    Created  time.Time `json:"created"`
}

var users = []User{
    {ID: uuid.New(), Name: "John Doe", Email: "john@example.com", Created: time.Now()},
    {ID: uuid.New(), Name: "Jane Smith", Email: "jane@example.com", Created: time.Now()},
}

func main() {
    router := mux.NewRouter(
        mux.WithTitle("My First API"),
        mux.WithVersion("1.0.0"),
        mux.WithDescription("A simple user management API built with Mux"),
    )
    
    router.UseLogging()
    router.UseCompression()
    
    api := router.NewRouteGroup("/api/v1")
    api.WithTags("API v1")
    
    users := api.NewRouteGroup("/users")
    users.WithTags("Users")
    
    users.GET("/", listUsers).
        WithSummary("List all users").
        WithParam("page", "query", 1, false).
        WithParam("limit", "query", 10, false).
        WithParam("search", "query", "", false).
        WithOKResponse(map[string]interface{}{})
        
    users.POST("/", createUser).
        WithSummary("Create a new user").
        WithJsonBody(User{}).
        WithCreatedResponse(User{}).
        WithBadRequestResponse()
        
    users.GET("/{id}", getUser).
        WithSummary("Get a user by ID").
        WithParam("id", "path", uuid.Nil, true).
        WithOKResponse(User{}).
        WithNotFoundResponse()
    
    router.GET("/health", func(c *mux.RouteContext) {
        c.OK(map[string]string{
            "status":    "healthy", 
            "timestamp": time.Now().Format(time.RFC3339),
        })
    }).WithSummary("Health check").WithTags("Health")
    
    generator := mux.NewGenerator()
    spec := generator.GenerateSpec(router)
    spec.MarshalToFile("openapi.yaml")
    
    router.GET("/openapi.yaml", func(c *mux.RouteContext) {
        c.Response.Header().Set("Content-Type", "application/yaml")
        spec.MarshalToWriter(c.Response)
    })
    
    http.ListenAndServe(":8080", router)
}

// Handler functions would go here...
```

🎉 Congratulations! You've built a complete, documented API with Mux in just 10 steps. You now have:

- RESTful endpoints with proper HTTP methods
- Request validation and error handling  
- Query parameter support with pagination and search
- OpenAPI 3.1 specification generation
- Structured logging and response compression
- Route grouping and organization

Check out the other documentation files to learn about advanced features like authentication, custom middleware, and production deployment patterns.