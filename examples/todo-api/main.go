package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/fgrzl/mux"
	"github.com/fgrzl/mux/pkg/openapi"
	"github.com/google/uuid"
)

// Task represents a single task item
type Todo struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// CreateTodoRequest is the input for creating a task
type CreateTodoRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// UpdateTodoRequest is the input for updating a task
type UpdateTodoRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Completed   *bool   `json:"completed,omitempty"`
}

// In-memory storage (for demo purposes)
var (
	todos   = make(map[string]*Todo)
	todosMu sync.RWMutex // Protect concurrent access
)

const (
	todoIDParam     = "todo-id"
	errMissingParam = "Missing parameter"
	errIDRequired   = "id parameter is required"
)

func main() {
	router := mux.NewRouter()

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
		WithPathParam("id", "The unique identifier of the todo item", todoIDParam).
		WithOKResponse(Todo{}).
		WithNotFoundResponse()

	api.PUT("/{id}", updateTodo).
		WithOperationID("updateTodo").
		WithSummary("Update a todo").
		WithPathParam("id", "The unique identifier of the todo item", todoIDParam).
		WithJsonBody(UpdateTodoRequest{}).
		WithOKResponse(Todo{})

	api.DELETE("/{id}", deleteTodo).
		WithOperationID("deleteTodo").
		WithSummary("Delete a todo").
		WithPathParam("id", "The unique identifier of the todo item", todoIDParam).
		WithNoContentResponse()

	// Serve OpenAPI spec
	router.GET("/openapi.json", func(c mux.RouteContext) {
		info, _ := router.InfoObject()
		routes, _ := router.Routes()

		gen := openapi.NewGenerator()
		spec, _ := gen.GenerateSpecFromRoutes(info, routes)
		c.OK(spec)
	})

	// Root endpoint
	router.GET("/", func(c mux.RouteContext) {
		c.OK(map[string]string{
			"message": "Todo API",
			"version": "1.0.0",
			"docs":    "/openapi.json",
		})
	})

	// Start server with graceful shutdown
	server := mux.NewServer(":8080", router)

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	server.Listen(ctx)
}

func listTodos(c mux.RouteContext) {
	completed, hasCompleted := c.Params()["completed"]

	todosMu.RLock()
	defer todosMu.RUnlock()

	// Convert map to slice
	result := make([]*Todo, 0, len(todos))
	for _, todo := range todos {
		// Filter by completion status if provided
		if hasCompleted {
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

func createTodo(c mux.RouteContext) {
	var req CreateTodoRequest

	// Parse JSON body
	if err := c.Bind(&req); err != nil {
		c.BadRequest("Invalid JSON", err.Error())
		return
	}

	// Validate input
	if req.Title == "" {
		c.BadRequest("Validation Error", "Title is required")
		return
	}

	// Construct the new task item
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

	// Return the item that was created
	c.Created(todo)
}

func getTodo(c mux.RouteContext) {
	id, ok := c.Param("id")
	if !ok {
		c.BadRequest(errMissingParam, errIDRequired)
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
	id, ok := c.Param("id")
	if !ok {
		c.BadRequest(errMissingParam, errIDRequired)
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

func deleteTodo(c mux.RouteContext) {
	id, ok := c.Param("id")
	if !ok {
		c.BadRequest(errMissingParam, errIDRequired)
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
