# Interactive Tutorial

Build a Todo API in about 30 minutes using the current safe public APIs.

## What You Will Build

By the end of this tutorial you will have:

- `GET /todos`
- `POST /todos`
- `GET /todos/{id}`
- `PUT /todos/{id}`
- `DELETE /todos/{id}`
- `GET /openapi.json`

The finished version in this repository lives at [examples/todo-api](../examples/todo-api/).

## Step 1: Create the Project

```bash
mkdir todo-api
cd todo-api
go mod init todo-api
go get github.com/fgrzl/mux
go get github.com/google/uuid
```

## Step 2: Define Models and Storage

Create `main.go` and start with the data model and in-memory storage:

```go
package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/fgrzl/mux"
	"github.com/google/uuid"
)

type Todo struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type CreateTodoRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type UpdateTodoRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Completed   *bool   `json:"completed,omitempty"`
}

var (
	todos   = make(map[string]*Todo)
	todosMu sync.RWMutex
)
```

## Step 3: Add the CRUD Handlers

Add the handlers below the model definitions.

### List Todos

```go
func listTodos(c mux.RouteContext) {
	completed, hasCompleted := c.Query().Bool("completed")

	todosMu.RLock()
	defer todosMu.RUnlock()

	result := make([]*Todo, 0, len(todos))
	for _, todo := range todos {
		if hasCompleted && todo.Completed != completed {
			continue
		}
		result = append(result, todo)
	}

	c.OK(result)
}
```

### Create Todo

```go
func createTodo(c mux.RouteContext) {
	var req CreateTodoRequest
	if err := c.Bind(&req); err != nil {
		c.BadRequest("Invalid JSON", err.Error())
		return
	}

	if req.Title == "" {
		c.BadRequest("Validation Error", "Title is required")
		return
	}

	todo := &Todo{
		ID:          uuid.New().String(),
		Title:       req.Title,
		Description: req.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	todosMu.Lock()
	todos[todo.ID] = todo
	todosMu.Unlock()

	c.Created(todo)
}
```

### Get, Update, and Delete

```go
func getTodo(c mux.RouteContext) {
	id, ok := c.Params().String("id")
	if !ok {
		c.BadRequest("Missing parameter", "id parameter is required")
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

func updateTodo(c mux.RouteContext) {
	id, ok := c.Params().String("id")
	if !ok {
		c.BadRequest("Missing parameter", "id parameter is required")
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

func deleteTodo(c mux.RouteContext) {
	id, ok := c.Params().String("id")
	if !ok {
		c.BadRequest("Missing parameter", "id parameter is required")
		return
	}

	todosMu.Lock()
	defer todosMu.Unlock()

	if _, exists := todos[id]; !exists {
		c.NotFound()
		return
	}

	delete(todos, id)
	c.NoContent()
}
```

## Step 4: Wire the Router and OpenAPI Metadata

Now add `main()` and register the routes inside `Configure(...)`.

```go
func main() {
	router := mux.NewRouter()

	if err := router.Configure(func(router *mux.Router) {
		api := router.Group("/todos")
		api.Tags("Todos")

		api.GET("/", listTodos).
			OperationID("listTodos").
			Summary("List all todos").
			WithQueryParam("completed", "Filter todos by completion state", true).
			OK([]Todo{})

		api.POST("/", createTodo).
			OperationID("createTodo").
			Summary("Create a new todo").
			AcceptJSON(CreateTodoRequest{}).
			Created(Todo{})

		api.GET("/{id}", getTodo).
			OperationID("getTodo").
			Summary("Get a todo by ID").
			WithPathParam("id", "The unique identifier of the todo", "todo-123").
			OK(Todo{}).
			Responds(404, mux.ProblemDetails{})

		api.PUT("/{id}", updateTodo).
			OperationID("updateTodo").
			Summary("Update a todo").
			WithPathParam("id", "The unique identifier of the todo", "todo-123").
			AcceptJSON(UpdateTodoRequest{}).
			OK(Todo{})

		api.DELETE("/{id}", deleteTodo).
			OperationID("deleteTodo").
			Summary("Delete a todo").
			WithPathParam("id", "The unique identifier of the todo", "todo-123").
			NoContent()

		router.GET("/openapi.json", func(c mux.RouteContext) {
			spec, err := mux.GenerateSpecWithGenerator(mux.NewGenerator(), router)
			if err != nil {
				c.ServerError("OpenAPI generation failed", err.Error())
				return
			}
			c.OK(spec)
		})

		router.GET("/", func(c mux.RouteContext) {
			c.OK(map[string]string{
				"message": "Todo API",
				"docs":    "/openapi.json",
			})
		})
	}); err != nil {
		panic(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	server := mux.NewServer(":8080", router)
	if err := server.Listen(ctx); err != nil {
		panic(err)
	}
}
```

## Step 5: Run the API

```bash
go run .
```

## Step 6: Test the Endpoints

Create a todo:

```bash
curl -X POST http://localhost:8080/todos \
  -H "Content-Type: application/json" \
  -d '{"title":"Learn Mux","description":"Finish the tutorial"}'
```

List todos:

```bash
curl http://localhost:8080/todos
```

Filter completed todos:

```bash
curl "http://localhost:8080/todos?completed=true"
```

Update a todo:

```bash
curl -X PUT http://localhost:8080/todos/{id} \
  -H "Content-Type: application/json" \
  -d '{"completed":true}'
```

Delete a todo:

```bash
curl -X DELETE http://localhost:8080/todos/{id}
```

Inspect the generated OpenAPI document:

```bash
curl http://localhost:8080/openapi.json
```

## Step 7: Compare with the Repository Example

Once you have your own version working, compare it with the maintained example in [examples/todo-api/main.go](../examples/todo-api/main.go). That version includes the same flow with repository-style naming and comments.

## Next Improvements

After the in-memory version works, the next practical upgrades are:

1. Replace the map with a real database.
2. Add authentication middleware.
3. Add pagination and sorting.
4. Add integration tests.
5. Serve the OpenAPI document with Swagger UI or another viewer.



