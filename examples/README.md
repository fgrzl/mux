# Examples

This directory contains working example applications demonstrating various Mux features and use cases.

## Available Examples

### 1. Hello World (`hello-world/`)
The simplest possible Mux application. Perfect for getting started.

**Features:**
- Basic routing
- Simple response handling

### 2. REST API (`rest-api/`)
A complete REST API example with CRUD operations.

**Features:**
- Full CRUD operations
- Request validation
- Error handling
- OpenAPI documentation
- Structured responses

### 3. Authentication (`auth-api/`)
API with JWT authentication and role-based authorization.

**Features:**
- JWT token authentication
- Role-based access control
- Login/logout endpoints
- Protected routes

### 4. Middleware Demo (`middleware-demo/`)
Demonstrates all built-in middleware features.

**Features:**
- All built-in middleware
- Custom middleware
- Middleware ordering
- Configuration examples

### 5. File Upload (`file-upload/`)
File upload service with validation and processing.

**Features:**
- File upload handling
- File type validation
- Size limits
- Storage management

## Running Examples

Each example is a standalone Go application. To run any example:

```bash
cd examples/[example-name]
go mod init example
go get github.com/fgrzl/mux
go run main.go
```

For examples with additional dependencies, check the individual README files.

## Example Structure

Each example follows this structure:

```
example-name/
├── README.md          # Example-specific documentation
├── main.go           # Main application file
├── handlers.go       # HTTP handlers (if needed)
├── models.go         # Data models (if needed)
├── middleware.go     # Custom middleware (if needed)
└── openapi.yaml      # Generated OpenAPI spec (if applicable)
```

## Learning Path

We recommend exploring the examples in this order:

1. **Hello World** - Understand basic routing
2. **REST API** - Learn CRUD operations and validation
3. **Authentication** - Add security to your APIs
4. **Middleware Demo** - Master middleware usage
5. **File Upload** - Handle file operations

## Testing Examples

All examples include curl commands in their README files for easy testing. You can also use tools like:

- [Postman](https://postman.com)
- [Insomnia](https://insomnia.rest)  
- [HTTPie](https://httpie.org)

## Contributing Examples

If you have a useful example that demonstrates Mux features:

1. Create a new directory in `examples/`
2. Follow the standard structure
3. Include a comprehensive README.md
4. Add curl commands for testing
5. Submit a pull request

## Need Help?

- Check the main [documentation](../docs/)
- Review existing examples for patterns
- Open an [issue](https://github.com/fgrzl/mux/issues) if you find problems