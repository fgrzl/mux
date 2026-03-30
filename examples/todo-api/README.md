# Todo API Example

A complete REST API demonstrating CRUD operations with Mux.

## Features

✅ Full CRUD operations (Create, Read, Update, Delete)  
✅ JSON request/response handling  
✅ Input validation  
✅ Proper HTTP status codes  
✅ OpenAPI 3.1 documentation  
✅ Thread-safe in-memory storage  
✅ Startup validation before the server begins serving  
✅ Production-ready server with graceful shutdown  

## Quick Start

```bash
# Run the API
go run main.go

# Server starts on http://localhost:8080
# Press Ctrl+C for graceful shutdown
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/todos` | List all todos (supports `?completed=true/false` filter) |
| POST | `/todos` | Create a new todo |
| GET | `/todos/{id}` | Get a specific todo |
| PUT | `/todos/{id}` | Update a todo |
| DELETE | `/todos/{id}` | Delete a todo |
| GET | `/openapi.json` | Get OpenAPI specification |

## Usage Examples

### Create a Todo

```bash
curl -X POST http://localhost:8080/todos \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Learn Mux",
    "description": "Complete the interactive tutorial"
  }'
```

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Learn Mux",
  "description": "Complete the interactive tutorial",
  "completed": false,
  "createdAt": "2024-01-15T10:30:00Z",
  "updatedAt": "2024-01-15T10:30:00Z"
}
```

### List All Todos

```bash
curl http://localhost:8080/todos
```

**Filter by completion status:**
```bash
# Only completed todos
curl "http://localhost:8080/todos?completed=true"

# Only incomplete todos
curl "http://localhost:8080/todos?completed=false"
```

### Get a Specific Todo

```bash
curl http://localhost:8080/todos/{id}
```

### Update a Todo

```bash
curl -X PUT http://localhost:8080/todos/{id} \
  -H "Content-Type: application/json" \
  -d '{
    "completed": true
  }'
```

You can update any combination of fields:
```bash
curl -X PUT http://localhost:8080/todos/{id} \
  -H "Content-Type: application/json" \
  -d '{
    "title": "New Title",
    "description": "Updated description",
    "completed": true
  }'
```

### Delete a Todo

```bash
curl -X DELETE http://localhost:8080/todos/{id}
```

## OpenAPI Documentation

View the generated OpenAPI specification:

```bash
curl http://localhost:8080/openapi.json | jq
```

Or paste the JSON into [Swagger Editor](https://editor.swagger.io/) for an interactive UI.

## Code Structure

```go
router := mux.NewRouter().Safe()

// Data model
type Todo struct {
    ID          string
    Title       string
    Description string
    Completed   bool
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// Route group with OpenAPI metadata
api := router.NewRouteGroup("/todos")
api.WithTags("Todos")

api.GET("/", listTodos).
    WithOperationID("listTodos").
    WithSummary("List all todos").
    WithQueryParam("completed", "Filter todos by completion state", true).
    WithOKResponse([]Todo{})

api.POST("/", createTodo).
    WithOperationID("createTodo").
    WithSummary("Create a new todo").
    WithJsonBody(CreateTodoRequest{}).
    WithCreatedResponse(Todo{})

if err := router.Err(); err != nil {
    panic(err)
}

// Production-ready server
server := mux.NewServer(":8080", router)

ctx, cancel := signal.NotifyContext(
    context.Background(),
    os.Interrupt,
    syscall.SIGTERM,
)
defer cancel()

if err := server.Listen(ctx); err != nil {
    panic(err)
}
```

### Production Features

The example uses `WebServer` which provides:
- **Graceful shutdown**: Completes in-flight requests before stopping
- **Production timeouts**: 10s read/write, 120s idle
- **Signal handling**: Responds to Ctrl+C and SIGTERM (Kubernetes)
- **Startup validation**: Fails before serving traffic when route configuration is invalid
- **Context-based lifecycle**: Clean shutdown management

**Why WebServer vs http.ListenAndServe?**
- Reduces boilerplate from 30+ lines → 7 lines
- Automatic graceful shutdown (critical for production)
- Proper signal handling for containers
- Production-ready defaults

## What You'll Learn

This example demonstrates:

1. **RESTful API Design** - Proper HTTP methods and status codes
2. **JSON Handling** - Using `c.Bind()` and `c.OK()`
3. **Path Parameters** - Extracting `{id}` from URLs
4. **Query Parameters** - Filtering with `?completed=true`
5. **Validation** - Checking required fields
6. **Error Handling** - Returning appropriate errors
7. **Documentation** - Auto-generating OpenAPI specs
8. **Concurrency** - Thread-safe access with `sync.RWMutex`
9. **Production Deployment** - Using `WebServer` for graceful shutdown

## Next Steps

This example uses in-memory storage. For production:

1. **Add a Database** - Replace the map with PostgreSQL, MongoDB, etc.
2. **Add Authentication** - Protect endpoints with JWT or API keys
3. **Add Pagination** - Limit results for large datasets
4. **Add Sorting** - Allow sorting by date, title, etc.
5. **Add Search** - Full-text search on title/description

## Related Examples

- [Hello World](../hello-world/) - Minimal example to get started
- [Interactive Tutorial](../../docs/interactive-tutorial.md) - Step-by-step guide building this API
- [Learning Path](../../docs/learning-path.md) - Progressive learning path

## Testing

Test all endpoints with this script:

```bash
#!/bin/bash

# Create a todo
ID=$(curl -s -X POST http://localhost:8080/todos \
  -H "Content-Type: application/json" \
  -d '{"title":"Test","description":"Testing the API"}' \
  | jq -r '.id')

echo "Created todo: $ID"

# List todos
curl -s http://localhost:8080/todos | jq

# Get specific todo
curl -s http://localhost:8080/todos/$ID | jq

# Update todo
curl -s -X PUT http://localhost:8080/todos/$ID \
  -H "Content-Type: application/json" \
  -d '{"completed":true}' | jq

# Delete todo
curl -X DELETE http://localhost:8080/todos/$ID

echo "Test complete!"
```

---

**Built with ❤️ using Mux**
