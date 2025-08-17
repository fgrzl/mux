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
router := mux.NewRouter()
```
Creates a new Mux router instance.

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
    name := c.Param("name")
    // ...
})
- Returns a structured JSON response

### Server Startup
```go
http.ListenAndServe(":8080", router)
```
Starts the HTTP server on port 8080 using the Mux router.

## Key Concepts Demonstrated

1. **Router Creation**: How to create and configure a Mux router
2. **Route Definition**: Defining HTTP endpoints with handlers
3. **Path Parameters**: Capturing dynamic values from URLs
4. **Response Handling**: Using Mux response helpers (`c.OK`, `c.BadRequest`)
5. **Parameter Extraction**: Safely extracting path parameters
6. **Error Handling**: Basic error responses for invalid requests

## Next Steps

After understanding this example, try:
- [REST API Example](../rest-api/) - Full CRUD operations
- [Authentication Example](../auth-api/) - Adding security
- [Middleware Demo](../middleware-demo/) - Using middleware