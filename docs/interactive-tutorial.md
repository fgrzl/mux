# Interactive Tutorial

**Build a Todo API in 30 Minutes**

This hands-on tutorial will guide you through building a complete Todo API with Mux. You'll learn by doing!

---

## 🎯 What We're Building

A REST API for managing todos with:
- ✅ Create, Read, Update, Delete operations
- ✅ JSON request/response handling
- ✅ Input validation
- ✅ Error handling
- ✅ OpenAPI documentation
- ✅ Authentication

**Final API Endpoints**:
```
GET    /todos           - List all todos
POST   /todos           - Create a new todo
GET    /todos/{id}      - Get a specific todo
PUT    /todos/{id}      - Update a todo
DELETE /todos/{id}      - Delete a todo
```

---

## Step 1: Project Setup (2 minutes)

Create your project:

```bash
mkdir todo-api
cd todo-api
go mod init todo-api
go get github.com/fgrzl/mux
go get github.com/google/uuid
```

Create `main.go`:

```go
package main

import (
    "net/http"
    "github.com/fgrzl/mux"
)

func main() {
    router := mux.NewRouter()
    
    router.GET("/", func(c mux.RouteContext) {
        c.OK(map[string]string{
            "message": "Todo API",
            "version": "1.0.0",
        })
    })
    
    http.ListenAndServe(":8080", router)
}
```

**Test it**:
```bash
go run main.go
curl http://localhost:8080/
```

✅ **Checkpoint**: You should see the welcome message!

---

## Step 2: Define the Todo Model (3 minutes)

Add the Todo struct to `main.go`:

```go
package main

import (
    "sync"
    "time"
    "github.com/google/uuid"
    "net/http"
    "github.com/fgrzl/mux"
)

// Todo represents a single todo item
type Todo struct {
    ID          string    `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Completed   bool      `json:"completed"`
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
}

// In-memory storage (for demo purposes)
var (
    todos   = make(map[string]*Todo)
    todosMu sync.RWMutex // Protect concurrent access
)

func main() {
    // ... existing code
}
```

✅ **Checkpoint**: Code should compile without errors.

---

## Step 3: List All Todos (5 minutes)

Add the `listTodos` handler:

```go
func listTodos(c mux.RouteContext) {
    todosMu.RLock()
    defer todosMu.RUnlock()
    
    // Convert map to slice
    result := make([]*Todo, 0, len(todos))
    for _, todo := range todos {
        result = append(result, todo)
    }
    
    c.OK(result)
}

func main() {
    router := mux.NewRouter()
    
    // Add the endpoint
    router.GET("/todos", listTodos)
    
    http.ListenAndServe(":8080", router)
}
```

**Test it**:
```bash
curl http://localhost:8080/todos
# Should return: []
```

✅ **Checkpoint**: Empty array means it's working!

---

## Step 4: Create a Todo (7 minutes)

Add validation and the create handler:

```go
// CreateTodoRequest is the input for creating a todo
type CreateTodoRequest struct {
    Title       string `json:"title"`
    Description string `json:"description"`
}

func createTodo(c mux.RouteContext) {
    var req CreateTodoRequest
    
    // Parse JSON body
    if err := c.Bind(&req); err != nil {
        c.BadRequest("Invalid JSON", err.Error())
        return
    }
    
    // Validate input
    if req.Title == "" {
        c.BadRequest("Invalid todo", "title is required")
        return
    }
    
    // Create the todo
    todo := &Todo{
        ID:          uuid.New().String(),
        Title:       req.Title,
        Description: req.Description,
        Completed:   false,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }
    
    // Save it
    todosMu.Lock()
    todos[todo.ID] = todo
    todosMu.Unlock()
    
    // Return the created todo
    c.Created(todo)
}

func main() {
    router := mux.NewRouter()
    
    router.GET("/todos", listTodos)
    router.POST("/todos", createTodo)  // Add this line
    
    http.ListenAndServe(":8080", router)
}
```

**Test it**:
```bash
curl -X POST http://localhost:8080/todos \
  -H "Content-Type: application/json" \
  -d '{"title":"Learn Mux","description":"Complete the tutorial"}'

# Should return the created todo with an ID

# Now list todos
curl http://localhost:8080/todos
# Should return array with one todo
```

✅ **Checkpoint**: You should see your todo in the list!

---

## Step 5: Get a Single Todo (5 minutes)

Add the get handler:

```go
func getTodo(c mux.RouteContext) {
    id, ok := c.Param("id")
    if !ok {
        c.BadRequest("Missing todo ID", "id parameter is required")
        return
    }
    
    todosMu.RLock()
    todo, exists := todos[id]
    todosMu.RUnlock()
    
    if !exists {
        c.NotFound()
        return
    }
    
    c.OK(todo)
}

func main() {
    router := mux.NewRouter()
    
    router.GET("/todos", listTodos)
    router.POST("/todos", createTodo)
    router.GET("/todos/{id}", getTodo)  // Add this line
    
    http.ListenAndServe(":8080", router)
}
```

**Test it**:
```bash
# First, create a todo and note its ID
curl -X POST http://localhost:8080/todos \
  -H "Content-Type: application/json" \
  -d '{"title":"Test Todo"}'

# Use the returned ID
curl http://localhost:8080/todos/{paste-id-here}
```

✅ **Checkpoint**: You should see the specific todo!

---

## Step 6: Update a Todo (6 minutes)

Add the update handler:

```go
type UpdateTodoRequest struct {
    Title       *string `json:"title,omitempty"`
    Description *string `json:"description,omitempty"`
    Completed   *bool   `json:"completed,omitempty"`
}

func updateTodo(c mux.RouteContext) {
    id, ok := c.Param("id")
    if !ok {
        c.BadRequest("Missing todo ID", "id parameter is required")
        return
    }
    
    var req UpdateTodoRequest
    if err := c.Bind(&req); err != nil {
        c.BadRequest("Invalid JSON", err.Error())
        return
    }
    
    todosMu.Lock()
    defer todosMu.Unlock()
    
    todo, exists := todos[id]
    if !exists {
        c.NotFound()
        return
    }
    
    // Update only provided fields
    if req.Title != nil {
        todo.Title = *req.Title
    }
    if req.Description != nil {
        todo.Description = *req.Description
    }
    if req.Completed != nil {
        todo.Completed = *req.Completed
    }
    
    todo.UpdatedAt = time.Now()
    
    c.OK(todo)
}

func main() {
    router := mux.NewRouter()
    
    router.GET("/todos", listTodos)
    router.POST("/todos", createTodo)
    router.GET("/todos/{id}", getTodo)
    router.PUT("/todos/{id}", updateTodo)  // Add this line
    
    http.ListenAndServe(":8080", router)
}
```

**Test it**:
```bash
# Mark a todo as completed
curl -X PUT http://localhost:8080/todos/{id} \
  -H "Content-Type: application/json" \
  -d '{"completed":true}'
```

✅ **Checkpoint**: Todo should now be marked as completed!

---

## Step 7: Delete a Todo (4 minutes)

Add the delete handler:

```go
func deleteTodo(c mux.RouteContext) {
    id, ok := c.Param("id")
    if !ok {
        c.BadRequest("Missing todo ID", "id parameter is required")
        return
    }
    
    todosMu.Lock()
    defer todosMu.Unlock()
    
    if _, exists := todos[id]; !exists {
        c.NotFound()
        return
    }
    
    delete(todos, id)
    c.NoContent() // 204 No Content
}

func main() {
    router := mux.NewRouter()
    
    router.GET("/todos", listTodos)
    router.POST("/todos", createTodo)
    router.GET("/todos/{id}", getTodo)
    router.PUT("/todos/{id}", updateTodo)
    router.DELETE("/todos/{id}", deleteTodo)  // Add this line
    
    http.ListenAndServe(":8080", router)
}
```

**Test it**:
```bash
# Delete a todo
curl -X DELETE http://localhost:8080/todos/{id}

# Verify it's gone
curl http://localhost:8080/todos/{id}
# Should return 404
```

✅ **Checkpoint**: Todo should be deleted!

---

## Step 8: Add OpenAPI Documentation (8 minutes)

Document your API:

```go
func main() {
    router := mux.NewRouter()
    
    // Add middleware
    mux.UseLogging(router)
    
    // Create API group
    api := router.NewRouteGroup("/todos")
    api.WithTags("Todos")
    
    // Document each endpoint
    api.GET("/", listTodos).
        WithOperationID("listTodos").
        WithSummary("List all todos").
        WithOKResponse([]Todo{})
    
    api.POST("/", createTodo).
        WithOperationID("createTodo").
        WithSummary("Create a new todo").
        WithJsonBody(CreateTodoRequest{}).
        WithCreatedResponse(Todo{})
    
    api.GET("/{id}", getTodo).
        WithOperationID("getTodo").
        WithSummary("Get a todo by ID").
        WithPathParam("id", "The unique identifier of the todo", "todo-123").
        WithOKResponse(Todo{}).
        WithNotFoundResponse()
    
    api.PUT("/{id}", updateTodo).
        WithOperationID("updateTodo").
        WithSummary("Update a todo").
        WithPathParam("id", "The unique identifier of the todo", "todo-123").
        WithJsonBody(UpdateTodoRequest{}).
        WithOKResponse(Todo{})
    
    api.DELETE("/{id}", deleteTodo).
        WithOperationID("deleteTodo").
        WithSummary("Delete a todo").
        WithPathParam("id", "The unique identifier of the todo", "todo-123").
        WithNoContentResponse()
    
    // Serve OpenAPI spec
    router.GET("/openapi.json", func(c mux.RouteContext) {
        spec := router.OpenAPI(&mux.OpenAPIOptions{
            Title:       "Todo API",
            Version:     "1.0.0",
            Description: "A simple todo management API built with Mux",
        })
        c.OK(spec)
    })
    
    http.ListenAndServe(":8080", router)
}
```

**Test it**:
```bash
curl http://localhost:8080/openapi.json | jq

# Or view in Swagger UI:
# 1. Go to https://editor.swagger.io/
# 2. Paste your OpenAPI JSON
```

✅ **Checkpoint**: You should see a complete OpenAPI 3.1 spec!

---

## Bonus Step: Add Health Checks (2 minutes)

Add Kubernetes-style health probe endpoints:

```go
func main() {
    router := mux.NewRouter()
    
    // Add built-in health probes
    router.Healthz()  // GET /healthz
    router.Livez()    // GET /livez
    router.Readyz()   // GET /readyz
    
    // Rest of your routes...
    api := router.NewRouteGroup("/todos")
    // ...
}
```

**Test it**:
```bash
curl http://localhost:8080/healthz
# Returns: ok

curl http://localhost:8080/readyz
# Returns: ok
```

These endpoints are perfect for Kubernetes liveness and readiness probes!

✅ **Checkpoint**: Health probes working!

---

## 🎉 Congratulations!

You've built a complete REST API with Mux! You learned:

- ✅ Setting up routes
- ✅ Handling JSON requests/responses
- ✅ Input validation
- ✅ CRUD operations
- ✅ Error handling (404, 400, etc.)
- ✅ OpenAPI documentation
- ✅ Route parameters

---

## 🚀 Next Steps

### Challenge 1: Add Filtering
Add a query parameter to filter by completion status:

```go
func listTodos(c mux.RouteContext) {
    completed, ok := c.QueryValue("completed")
    
    todosMu.RLock()
    defer todosMu.RUnlock()
    
    result := make([]*Todo, 0)
    for _, todo := range todos {
        if ok {
            if completed == "true" && !todo.Completed {
                continue
            }
            if completed == "false" && todo.Completed {
                continue
            }
        }
        result = append(result, todo)
    }
    
    c.OK(result)
}
```

### Challenge 2: Add Pagination
Limit results to 10 per page with page numbers.

### Challenge 3: Add Authentication
Protect endpoints with Bearer token authentication.

### Challenge 4: Connect a Real Database
Replace the in-memory map with PostgreSQL or MongoDB.

---

## 📚 Complete Code

The full working code for this tutorial is available at:
`examples/todo-api/main.go`

---

## 💡 What You Learned

**Routing**: How to set up GET, POST, PUT, DELETE endpoints  
**Request Handling**: Parsing JSON with `c.Bind()`  
**Response Handling**: Using `c.OK()`, `c.Created()`, `c.NotFound()`  
**Path Parameters**: Accessing URL params with `c.Param()`  
**Validation**: Checking required fields  
**Error Handling**: Returning appropriate HTTP status codes  
**Documentation**: Adding OpenAPI metadata  

**Time to production**: Just add a database and deploy! 🚀

---

## 🆘 Troubleshooting

**"Cannot bind to pointer"**  
- Make sure your struct fields are exported (capitalized)

**"404 Not Found" for valid endpoints**  
- Check that you registered the route with `router.GET()` etc.
- Verify the path matches exactly (case-sensitive)

**"Empty array" when expecting todos**  
- Remember: todos are stored in memory and reset when you restart the server

**JSON not parsing**  
- Ensure `Content-Type: application/json` header is set
- Check JSON syntax with a validator

---

**Ready for more? Check out the [Learning Path](learning-path.md) for advanced topics!**

## See Also

- [Learning Path](learning-path.md) - Structured progression from beginner to advanced
- [Getting Started](getting-started.md) - Comprehensive introduction
- [Middleware](middleware.md) - Built-in middleware guide
- [Best Practices](best-practices.md) - Patterns and conventions
- [Cheat Sheet](cheat-sheet.md) - Quick reference
