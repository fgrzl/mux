# Hello World Example

The simplest possible Mux application demonstrating basic routing and response handling.

## Features

- Basic GET endpoints
- Path parameters
- JSON responses
- Error handling

## Running the Example

```bash
# From the hello-world directory
go mod init hello-world-example
go get github.com/fgrzl/mux
go run main.go
```

The server will start on `http://localhost:8080`

## Testing the Endpoints

### Simple Hello World
```bash
curl http://localhost:8080/
```
**Expected Response:**
```json
"Hello, World!"
```

### Personalized Greeting
```bash
curl http://localhost:8080/hello/John
```
**Expected Response:**
```json
{
  "message": "Hello, John!",
  "status": "success"
}
```

### Test Error Handling
```bash
curl http://localhost:8080/hello/
```
This will return a 404 since the route requires a name parameter.

## Code Explanation

### Router Creation
```go
router := mux.NewRouter().Safe()
```
Creates a new Mux router instance and enables startup validation collection.

### Basic Route
```go
router.GET("/", func(c mux.RouteContext) {
    c.OK("Hello, World!")
})
```
- Defines a GET endpoint at the root path
- Returns a simple string response with 200 OK status

### Parameterized Route
```go
router.GET("/hello/{name}", func(c mux.RouteContext) {
    name, ok := c.Param("name")
    if !ok {
        c.BadRequest("Missing name", "name parameter is required")
        return
    }

    c.OK(map[string]string{
        "message": "Hello, " + name + "!",
        "status":  "success",
    })
})
```
- Returns a structured JSON response

### Startup Validation
```go
if err := router.Err(); err != nil {
    panic(err)
}
```
Checks route configuration before the server starts accepting traffic.

### Server Startup
```go
server := mux.NewServer(":8080", router)

ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer cancel()

if err := server.Listen(ctx); err != nil {
    panic(err)
}
```
Uses `WebServer` for production-ready server with:
- Automatic graceful shutdown (press Ctrl+C)
- Production timeouts (10s read/write, 120s idle)
- Context-based lifecycle management

## Key Concepts Demonstrated

1. **Router Creation**: How to create and configure a Mux router
2. **Route Definition**: Defining HTTP endpoints with handlers
3. **Path Parameters**: Capturing dynamic values from URLs
4. **Response Handling**: Using Mux response helpers (`c.OK`, `c.BadRequest`)
5. **Parameter Extraction**: Safely extracting path parameters
6. **Error Handling**: Basic error responses for invalid requests

## Next Steps

After understanding this example, try:
- [Todo API Example](../todo-api/) - Full CRUD operations with OpenAPI
- [CORS Wildcard Example](../cors-wildcard/) - Middleware configuration
- [WebServer Example](../webserver/) - Server lifecycle and timeouts