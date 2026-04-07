# Code Guidelines

This document provides comprehensive coding guidelines for the mux project. All code contributions must follow these standards to ensure consistency, maintainability, and idiomatic Go practices.

## Philosophy: The Pit of Success

We design our APIs and code to make **correct usage easy and incorrect usage difficult**. This "pit of success" approach means:

- **Safe by default** - The simplest way to use an API should be the safest
- **Hard to misuse** - Type safety and clear contracts prevent common mistakes
- **Fail fast** - Errors should be caught at compile time when possible
- **Clear intent** - Code should be self-documenting through good naming and structure
- **Guided path** - Developers naturally fall into correct patterns

**Example: Context as first parameter**
```go
// Pit of Success - Context first makes it hard to forget
func GetUser(ctx context.Context, id string) (*User, error)

// Easy to misuse - Context might be forgotten
func GetUser(id string, ctx context.Context) (*User, error)
```

**Example: Functional options pattern**
```go
// Pit of Success - Safe defaults, optional configuration
rtr := NewRouter() // Works with sensible defaults
rtr := NewRouter(WithTimeout(30 * time.Second)) // Easy to customize

// Easy to misuse - Must specify everything
rtr := NewRouter(30, true, false, 100, nil) // What do these mean?
```

## Table of Contents

- [Philosophy: The Pit of Success](#philosophy-the-pit-of-success)
- [Code Style](#code-style)
- [Package Design](#package-design)
- [Naming Conventions](#naming-conventions)
- [Function Design](#function-design)
- [Error Handling](#error-handling)
- [Context Usage](#context-usage)
- [Logging](#logging)
- [Comments and Documentation](#comments-and-documentation)
- [Best Practices](#best-practices)

## Code Style

### Idiomatic Go

Always write **idiomatic Go** code:

- Use `gofmt` (formatting happens automatically)
- Follow Go community conventions
- Keep it simple and readable
- Prefer clarity over cleverness
- Avoid unnecessary abstractions
- Don't fight the language

### Formatting

Code is automatically formatted by `gofmt`. Key points:

- **Tabs** for indentation (not spaces)
- **Line length** - No hard limit, but keep it reasonable (~120 chars)
- **Imports** - Grouped: standard library, external, internal
- **Braces** - K&R style (opening brace on same line)

```go
// Good - Idiomatic formatting
func Process(input string) (string, error) {
    if input == "" {
        return "", errors.New("input required")
    }
    return strings.ToUpper(input), nil
}

// Bad - Non-idiomatic
func Process(input string) (string, error) 
{  // Opening brace on new line
    if input == ""  // Missing braces
        return "", errors.New("input required")
    return strings.ToUpper(input), nil
}
```

### Import Organization

```go
import (
    // Standard library
    "context"
    "fmt"
    "net/http"
    
    // External packages
    "github.com/stretchr/testify/assert"
    
    // Internal packages
    "github.com/fgrzl/mux/internal/router"
    "github.com/fgrzl/mux/internal/routing"
)
```

## Package Design

### Minimal Dependencies

Favor small, focused packages with minimal dependencies:

```go
// Good - Minimal imports
package tokenizer

import "strings"

// Bad - Too many dependencies
package tokenizer

import (
    "strings"
    "net/http"
    "github.com/fgrzl/mux/internal/router"
    "github.com/fgrzl/mux/internal/middleware/cors"
)
```

### Package Cohesion

Each package should have a single, clear purpose:

```go
// Good - Focused packages
internal/router/     // HTTP routing
internal/binder/     // Request binding
internal/middleware/ // HTTP middleware

// Bad - Mixed responsibilities
internal/utils/      // Too generic
internal/helpers/    // Unclear purpose
```

### Avoid Circular Dependencies

Structure packages to prevent import cycles:

```go
// Good - Clear hierarchy
internal/routing/       // Core types (no dependencies)
internal/router/        // Router (imports routing)
internal/middleware/    // Middleware (imports routing)

// Bad - Circular dependencies
internal/router/ imports internal/middleware/
internal/middleware/ imports internal/router/
```

### Pit of Success: Package Design

**Make it easy to:**
- Keep packages small and focused
- Minimize dependencies
- Avoid circular imports

**Make it hard to:**
- Create "god packages" with everything
- Mix unrelated concerns
- Build circular dependencies

```go
// Pit of Success - Clear boundaries
internal/tokenizer/     // No dependencies, pure logic
internal/routing/       // Core types only
internal/router/        // Imports routing, not middleware
internal/middleware/    // Imports routing, not router

// Easy to misuse - Everything depends on everything
internal/core/          // "Core" utilities used everywhere
internal/common/        // Shared code creates coupling
```

## Naming Conventions

### Variables

Use short, clear names:

```go
// Good - Clear and concise
func Process(ctx context.Context, req *http.Request) error {
    user := getUser(ctx)
    if user == nil {
        return errors.New("user not found")
    }
    return nil
}

// Bad - Too verbose or unclear
func Process(applicationContext context.Context, httpRequest *http.Request) error {
    u := getUser(applicationContext)
    if u == nil {
        return errors.New("user not found")
    }
    return nil
}
```

### Common Short Names

```go
ctx     // context.Context
req     // *http.Request
res/rec // http.ResponseWriter
err     // error
buf     // *bytes.Buffer
i, j, k // loop indices
n       // count/length
r       // io.Reader
w       // io.Writer
```

### Functions and Methods

Use **verb** or **verb-noun** format:

```go
// Good - Clear action
func GetUser(id string) (*User, error)
func ValidateInput(input string) error
func ProcessRequest(req *http.Request) error

// Bad - Unclear or noun-only
func User(id string) (*User, error)
func Input(input string) error
func Request(req *http.Request) error
```

### Types

Use **noun** format, singular for single items:

```go
// Good - Clear nouns
type Router struct { ... }
type Middleware interface { ... }
type User struct { ... }

// Bad - Unclear or plural
type RouterThing struct { ... }
type Middlewares interface { ... }
type Users struct { ... } // Unless it's a collection
```

### Constants

Use **PascalCase** for exported, **camelCase** for unexported:

```go
// Good
const (
    DefaultTimeout = 30 * time.Second
    MaxRetries     = 3
)

const (
    defaultBufferSize = 4096
    maxCacheEntries   = 1000
)

// Bad - ALL_CAPS (not idiomatic Go)
const (
    DEFAULT_TIMEOUT = 30
    MAX_RETRIES     = 3
)
```

### Interfaces

Use **-er** suffix when possible:

```go
// Good - Standard Go convention
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Middleware interface {
    Invoke(c MutableRouteContext, next HandlerFunc)
}

// Also acceptable - Clear purpose
type Authenticator interface {
    Authenticate(token string) (*User, error)
}
```

### Pit of Success: Naming

**Make it easy to:**
- Write self-documenting code
- Follow Go conventions
- Understand intent at a glance

**Make it hard to:**
- Use vague or misleading names
- Create confusion about purpose
- Ignore context

```go
// Pit of Success - Names tell a story
func ValidateEmail(email string) error
func GetUserByID(ctx context.Context, id string) (*User, error)
const DefaultTimeout = 30 * time.Second

// Easy to misuse - Unclear intent
func Check(s string) error
func Get(c context.Context, s string) (*User, error)
const TIMEOUT = 30
```

## Function Design

### Small, Focused Functions

Keep functions small and focused on one task:

```go
// Good - Single responsibility
func ValidateEmail(email string) error {
    if email == "" {
        return errors.New("email required")
    }
    if !strings.Contains(email, "@") {
        return errors.New("invalid email format")
    }
    return nil
}

func ValidateAge(age int) error {
    if age < 0 {
        return errors.New("age must be positive")
    }
    if age > 150 {
        return errors.New("invalid age")
    }
    return nil
}

// Bad - Does too much
func ValidateUser(email string, age int, password string) error {
    // 50+ lines of validation logic
}
```

### Return Early

Avoid deep nesting by returning early:

```go
// Good - Early returns
func Process(input string) error {
    if input == "" {
        return errors.New("input required")
    }
    
    data, err := parse(input)
    if err != nil {
        return err
    }
    
    if !validate(data) {
        return errors.New("validation failed")
    }
    
    return save(data)
}

// Bad - Nested if statements
func Process(input string) error {
    if input != "" {
        data, err := parse(input)
        if err == nil {
            if validate(data) {
                return save(data)
            } else {
                return errors.New("validation failed")
            }
        } else {
            return err
        }
    } else {
        return errors.New("input required")
    }
}
```

### Function Parameters

- Keep parameter lists short (ideally <= 3)
- Use options pattern for many parameters
- Put context first (if used)

```go
// Good - Few parameters
func Get(ctx context.Context, id string) (*User, error)

// Good - Options pattern for many params
type Options struct {
    Timeout time.Duration
    Retries int
    Cache   bool
}

func New(opts Options) *Client

// Bad - Too many parameters
func Get(ctx context.Context, id string, timeout int, retries int, cache bool, validate bool) (*User, error)
```

### Functional Options Pattern

For optional configuration, use functional options:

```go
// Good - Functional options
type Option func(*Router)

func WithTimeout(d time.Duration) Option {
    return func(r *Router) {
        r.timeout = d
    }
}

func WithLogging() Option {
    return func(r *Router) {
        r.logging = true
    }
}

func NewRouter(opts ...Option) *Router {
    r := &Router{
        timeout: DefaultTimeout,
    }
    for _, opt := range opts {
        opt(r)
    }
    return r
}

// Usage:
rtr := NewRouter(
    WithTimeout(30 * time.Second),
    WithLogging(),
)
```

### Pit of Success: Function Design

**Make it easy to:**
- Write small, focused functions
- Return early and avoid nesting
- Use safe defaults with optional overrides
- Understand function parameters

**Make it hard to:**
- Create large, multi-purpose functions
- Write deeply nested code
- Forget required parameters
- Misuse optional configuration

```go
// Pit of Success - Context first, clear signature
func GetUser(ctx context.Context, id string) (*User, error)

// Pit of Success - Safe defaults, extensible
func NewRouter(opts ...Option) *Router {
    r := &Router{
        timeout: DefaultTimeout,  // Safe default
        maxBodySize: 10 * 1024 * 1024, // Safe default
    }
    for _, opt := range opts {
        opt(r)
    }
    return r
}

// Easy to misuse - Too many parameters, no defaults
func NewRouter(timeout int, maxBody int, logging bool, cache bool, validate bool) *Router

// Easy to misuse - Context last or missing
func GetUser(id string, ctx context.Context) (*User, error)
func GetUser(id string) (*User, error)
```

**Example: Making errors hard to ignore**
```go
// Pit of Success - Must handle error
user, err := GetUser(ctx, id)
if err != nil {
    return err
}

// Easy to misuse - Error can be ignored
user := GetUser(ctx, id) // Compiles but loses error
```

## Error Handling

### Prefer errors.Is / errors.As

Use `errors.Is` and `errors.As` for error checking:

```go
// Good - errors.Is
if errors.Is(err, ErrNotFound) {
    return nil // Handle not found
}

// Good - errors.As
var validationErr *ValidationError
if errors.As(err, &validationErr) {
    return fmt.Errorf("validation failed: %w", validationErr)
}

// Bad - Direct comparison
if err == ErrNotFound {
    return nil
}

// Bad - Type assertion
if _, ok := err.(*ValidationError); ok {
    // Handle validation error
}
```

### Wrap Errors

Add context when returning errors:

```go
// Good - Wrapped with context
func LoadUser(id string) (*User, error) {
    user, err := db.Get(id)
    if err != nil {
        return nil, fmt.Errorf("failed to load user %s: %w", id, err)
    }
    return user, nil
}

// Bad - Lost context
func LoadUser(id string) (*User, error) {
    user, err := db.Get(id)
    if err != nil {
        return nil, err
    }
    return user, nil
}
```

### Sentinel Errors

Define package-level sentinel errors:

```go
// Good - Exported sentinel errors
var (
    ErrNotFound      = errors.New("resource not found")
    ErrInvalidInput  = errors.New("invalid input")
    ErrUnauthorized  = errors.New("unauthorized")
)

// Usage:
if user == nil {
    return nil, ErrNotFound
}
```

### Custom Error Types

For structured errors, create custom types:

```go
// Good - Custom error type
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}

// Usage:
if email == "" {
    return &ValidationError{
        Field:   "email",
        Message: "required",
    }
}
```

### Pit of Success: Error Handling

**Make it easy to:**
- Check errors with `errors.Is` and `errors.As`
- Add context when wrapping errors
- Define clear sentinel errors
- Create type-safe error handling

**Make it hard to:**
- Ignore errors silently
- Lose error context
- Do brittle error string matching
- Forget to wrap errors

```go
// Pit of Success - Error must be checked
user, err := GetUser(ctx, id)
if err != nil {
    return fmt.Errorf("failed to get user: %w", err) // Context added
}

// Pit of Success - Type-safe error checking
if errors.Is(err, ErrNotFound) {
    // Handle not found
}

// Pit of Success - Structured errors
type ValidationError struct {
    Field   string
    Message string
    Value   interface{}
}

// Easy to misuse - Error ignored
user, _ := GetUser(ctx, id) // Silent failure

// Easy to misuse - Brittle string matching
if strings.Contains(err.Error(), "not found") {
    // Breaks if error message changes
}

// Easy to misuse - Lost context
if err != nil {
    return err // Where did this fail?
}
```

**Design APIs to prevent error misuse:**
```go
// Pit of Success - Can't forget to close
func WithResource(ctx context.Context, fn func(*Resource) error) error {
    resource, err := openResource(ctx)
    if err != nil {
        return err
    }
    defer resource.Close() // Automatic cleanup
    
    return fn(resource)
}

// Easy to misuse - Caller must remember to close
func OpenResource(ctx context.Context) (*Resource, error) {
    // Caller must remember: defer resource.Close()
}
```

## Context Usage

### Always Use context.Context

Use `context.Context` in public APIs where cancellation or timeouts may apply:

```go
// Good - Context first parameter
func GetUser(ctx context.Context, id string) (*User, error)
func ProcessRequest(ctx context.Context, req *Request) error

// Bad - No context for long-running operations
func GetUser(id string) (*User, error)
func ProcessRequest(req *Request) error
```

### Context Best Practices

```go
// Good - Pass context through call chain
func Handler(ctx context.Context) error {
    user, err := getUserFromDB(ctx)
    if err != nil {
        return err
    }
    return processUser(ctx, user)
}

// Good - Check context cancellation
func LongOperation(ctx context.Context) error {
    for i := 0; i < 1000; i++ {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            // Continue processing
        }
        
        if err := processItem(ctx, i); err != nil {
            return err
        }
    }
    return nil
}

// Bad - Ignoring context
func Handler(ctx context.Context) error {
    user, err := getUserFromDB(context.Background()) // Lost cancellation
    return processUser(context.TODO(), user) // Lost deadline
}
```

### Don't Store Context

Never store context in structs:

```go
// Bad - Stored context
type Handler struct {
    ctx context.Context
}

// Good - Pass context to methods
type Handler struct {
    // Other fields
}

func (h *Handler) Process(ctx context.Context) error {
    // Use ctx parameter
}
```

### Pit of Success: Context Usage

**Make it easy to:**
- Pass context through call chains
- Support cancellation and timeouts
- Follow the context-first convention

**Make it hard to:**
- Forget context in long-running operations
- Store context in structs
- Break the cancellation chain

```go
// Pit of Success - Context first parameter (convention)
func ProcessRequest(ctx context.Context, req *Request) error
func FetchData(ctx context.Context, url string) ([]byte, error)

// Pit of Success - Methods that need context take it as parameter
type Service struct {
    db *Database
}

func (s *Service) GetUser(ctx context.Context, id string) (*User, error) {
    return s.db.Query(ctx, "SELECT * FROM users WHERE id = ?", id)
}

// Easy to misuse - Context not first
func ProcessRequest(req *Request, ctx context.Context) error

// Easy to misuse - Stored context becomes stale
type Service struct {
    ctx context.Context // Don't do this
    db  *Database
}

// Easy to misuse - Easy to forget context
func ProcessRequest(req *Request) error // No way to cancel
```

**API design that enforces context:**
```go
// Pit of Success - Can't call without context
type Database interface {
    Query(ctx context.Context, query string, args ...interface{}) (*Rows, error)
    Exec(ctx context.Context, query string, args ...interface{}) error
}

// Easy to misuse - Context is optional/forgotten
type Database interface {
    Query(query string, args ...interface{}) (*Rows, error)
    QueryContext(ctx context.Context, query string, args ...interface{}) (*Rows, error)
}
```

## Logging

### Use slog for Structured Logging

Always use `log/slog` for structured logging:

```go
import "log/slog"

// Good - Structured logging
func ProcessRequest(ctx context.Context, req *Request) error {
    slog.InfoContext(ctx, "processing request",
        "method", req.Method,
        "path", req.Path,
    )
    
    if err := validate(req); err != nil {
        slog.ErrorContext(ctx, "validation failed",
            "error", err,
            "path", req.Path,
        )
        return err
    }
    
    return nil
}

// Bad - Unstructured logging
func ProcessRequest(ctx context.Context, req *Request) error {
    log.Printf("Processing request: %s %s", req.Method, req.Path)
    
    if err := validate(req); err != nil {
        log.Printf("Validation failed: %v", err)
        return err
    }
    
    return nil
}
```

### Always Pass Context

Pass `context.Context` to log calls:

```go
// Good - Context passed
slog.InfoContext(ctx, "operation completed", "duration", elapsed)
slog.ErrorContext(ctx, "operation failed", "error", err)

// Bad - No context
slog.Info("operation completed", "duration", elapsed)
slog.Error("operation failed", "error", err)
```

### Log Levels

Use appropriate log levels:

- **Debug** - Detailed diagnostic information
- **Info** - General informational messages
- **Warn** - Warning messages, not errors
- **Error** - Error messages that need attention

```go
// Good - Appropriate levels
slog.DebugContext(ctx, "cache hit", "key", key)
slog.InfoContext(ctx, "user logged in", "user_id", userID)
slog.WarnContext(ctx, "deprecated API used", "endpoint", path)
slog.ErrorContext(ctx, "database connection failed", "error", err)
```

### Pit of Success: Logging

**Make it easy to:**
- Use structured logging
- Include context in logs
- Use appropriate log levels

**Make it hard to:**
- Log unstructured messages
- Forget context
- Mix concerns with print statements

```go
// Pit of Success - Structured, with context
slog.InfoContext(ctx, "user action",
    "action", "login",
    "user_id", userID,
    "ip", req.RemoteAddr,
)

// Pit of Success - Errors with context
slog.ErrorContext(ctx, "database query failed",
    "error", err,
    "query", query,
    "duration", elapsed,
)

// Easy to misuse - Unstructured, no context
log.Printf("User %s logged in from %s", userID, req.RemoteAddr)

// Easy to misuse - Lost context
slog.Info("processing request") // Which request? Which user?
```

**Design logging helpers:**
```go
// Pit of Success - Helpers enforce structure
func LogRequest(ctx context.Context, req *http.Request, duration time.Duration) {
    slog.InfoContext(ctx, "request completed",
        "method", req.Method,
        "path", req.URL.Path,
        "duration_ms", duration.Milliseconds(),
        "status", getStatus(ctx),
    )
}

// Usage is simple and consistent
LogRequest(ctx, req, time.Since(start))
```

## Comments and Documentation

### GoDoc Style

All exported items must have GoDoc comments:

```go
// Good - Starts with item name
// Router handles HTTP routing and middleware composition.
// It provides a fluent API for defining routes and applying middleware.
type Router struct {
    // ...
}

// NewRouter creates a new Router instance with the given options.
// Options can be used to configure timeouts, middleware, and other behavior.
func NewRouter(opts ...Option) *Router {
    // ...
}

// Bad - Doesn't start with item name
// This is a router that handles HTTP routing
type Router struct {
    // ...
}

// Creates a new router
func NewRouter(opts ...Option) *Router {
    // ...
}
```

### Package Documentation

Every package must have package-level documentation:

```go
// Package router provides HTTP routing with middleware support.
//
// The router uses a tree-based matching algorithm for efficient
// route lookup and supports path parameters, wildcards, and catch-all routes.
//
// Example usage:
//
// rtr := router.NewRouter()
// rtr.GET("/users/{id}", getUserHandler)
// http.ListenAndServe(":8080", rtr)
package router
```

### Comment Style

```go
// Good - Clear, concise comments
// validateEmail checks if the email format is valid.
func validateEmail(email string) error {
    // ...
}

// Process handles the request by validating input,
// calling the handler, and formatting the response.
func Process(req *Request) (*Response, error) {
    // Validate request before processing
    if err := validate(req); err != nil {
        return nil, err
    }
    // ...
}

// Bad - Obvious or redundant comments
// This function validates email
func validateEmail(email string) error {
    // Check if email is empty
    if email == "" { // If the email is an empty string
        return errors.New("email required")
    }
    // ...
}
```

### Pit of Success: Documentation

**Make it easy to:**
- Write clear, helpful documentation
- Follow GoDoc conventions
- Provide usage examples

**Make it hard to:**
- Skip documentation
- Write misleading comments
- Forget to document exports

```go
// Pit of Success - Clear, helpful, with example
// GetUser retrieves a user by ID from the database.
// Returns ErrNotFound if the user doesn't exist.
//
// Example:
// user, err := GetUser(ctx, "user-123")
// if errors.Is(err, ErrNotFound) {
// // Handle not found
// }
func GetUser(ctx context.Context, id string) (*User, error)

// Pit of Success - Package doc with examples
// Package router provides HTTP routing with middleware support.
//
// Basic usage:
// rtr := router.NewRouter()
// rtr.GET("/users/{id}", handleUser)
// http.ListenAndServe(":8080", rtr)
package router

// Easy to misuse - Missing or unclear documentation
// GetUser gets a user
func GetUser(ctx context.Context, id string) (*User, error)

// Easy to misuse - No package documentation
package router
```

## Best Practices

### Do's

1. **Write idiomatic Go** - Follow community conventions
2. **Keep functions small** - Single responsibility
3. **Return early** - Avoid deep nesting
4. **Use context.Context** - For cancellation and timeouts
5. **Handle all errors** - Never ignore errors
6. **Use errors.Is/As** - For error checking
7. **Add GoDoc comments** - Document all exports
8. **Use slog** - For structured logging
9. **Minimize dependencies** - Keep packages focused
10. **Write tests** - Test all public APIs

### Don'ts

1. **Don't ignore errors** - Always check `err != nil`
2. **Don't use panic** - Except for unrecoverable errors
3. **Don't store context** - Pass it as parameter
4. **Don't use ALL_CAPS** - For constants (use PascalCase)
5. **Don't nest deeply** - Return early instead
6. **Don't write clever code** - Write clear code
7. **Don't over-abstract** - Keep it simple
8. **Don't create circular deps** - Structure packages properly
9. **Don't use `interface{}`** - Use `any` (Go 1.18+)
10. **Don't forget documentation** - Comment all exports

### Pit of Success Checklist

When designing APIs, ask yourself:

**Safety:**
- Are safe defaults provided?
- Can the API be misused? How can we prevent that?
- Will the compiler catch common mistakes?
- Is the happy path the easiest path?

**Clarity:**
- Is the API self-documenting?
- Are function signatures clear about what they do?
- Do types prevent invalid states?
- Are error cases obvious?

**Consistency:**
- Does this follow Go conventions?
- Does this match patterns used elsewhere in the codebase?
- Is context handled consistently?
- Are naming patterns consistent?

**Examples of Pit of Success Design:**

```go
// Type safety prevents invalid states
type Status string

const (
    StatusPending  Status = "pending"
    StatusApproved Status = "approved"
    StatusRejected Status = "rejected"
)

func SetStatus(status Status) error // Can't pass invalid status

// Builder pattern with compile-time safety
type RequestBuilder struct {
    method string
    path   string
    body   io.Reader
}

func NewRequest(method, path string) *RequestBuilder {
    return &RequestBuilder{method: method, path: path}
}

func (b *RequestBuilder) WithBody(body io.Reader) *RequestBuilder {
    b.body = body
    return b
}

func (b *RequestBuilder) Build() *http.Request {
    // Can't forget method or path - required in constructor
}

// Options pattern prevents misconfiguration
func NewServer(addr string, opts ...Option) *Server {
    s := &Server{
        addr:           addr,
        readTimeout:    30 * time.Second,    // Safe default
        writeTimeout:   30 * time.Second,    // Safe default
        maxHeaderBytes: 1 << 20,             // Safe default
    }
    for _, opt := range opts {
        opt(s)
    }
    return s
}
```

### Code Organization

```go
package mypackage

import (
    // Imports
)

// Constants
const (
    DefaultTimeout = 30 * time.Second
)

// Variables
var (
    ErrNotFound = errors.New("not found")
)

// Types
type Config struct {
    // ...
}

// Constructors
func NewConfig() *Config {
    // ...
}

// Methods
func (c *Config) Validate() error {
    // ...
}

// Functions
func Process(input string) error {
    // ...
}
```

## Tools

### Required Tools

```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run

# Run tests
go test ./...

# Check for race conditions
go test ./... -race

# Generate coverage
go test ./... -cover
```

### Recommended IDE Setup

- **VS Code** with Go extension
- **GoLand** / **IntelliJ IDEA** with Go plugin
- Enable format on save
- Enable organize imports on save

## Resources

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [Go Proverbs](https://go-proverbs.github.io/)

---

**Remember:** Good code is simple, clear, and idiomatic. When in doubt, favor readability over cleverness.

