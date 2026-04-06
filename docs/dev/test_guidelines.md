# Testing Guidelines

This document provides comprehensive testing guidelines for the mux project. All code contributions must be covered with meaningful tests following these standards.

## Table of Contents

- [Testing Philosophy](#testing-philosophy)
- [Test Framework](#test-framework)
- [Test Naming](#test-naming)
- [Test Structure](#test-structure)
- [Test Coverage](#test-coverage)
- [Table-Driven Tests](#table-driven-tests)
- [Middleware Testing](#middleware-testing)
- [Helper Functions](#helper-functions)
- [Common Patterns](#common-patterns)
- [Best Practices](#best-practices)

## Testing Philosophy

### Core Principles

1. **All code must be covered with meaningful tests** - Not just for coverage metrics, but to verify behavior
2. **Each test should test one thing** - Focus on a single behavior or edge case
3. **Tests are documentation** - They should clearly communicate what the code does
4. **Tests should be fast** - Use mocks/stubs for external dependencies
5. **Tests should be deterministic** - Same input should always produce same output

### What to Test

- **Public APIs** - All exported functions and methods
- **Edge Cases** - Empty inputs, nil values, boundary conditions
- **Error Paths** - All error conditions and error handling
- **Business Logic** - Core functionality and invariants
- **Implementation Details** - Don't test private methods directly

## Test Framework

### Standard Library + Testify

We use Go's standard `testing` package with `testify` for assertions:

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)
```

**When to use `assert` vs `require`:**

- **`require`** - Test should stop immediately if this fails (e.g., setup, prerequisites)
- **`assert`** - Test can continue after failure (e.g., multiple independent assertions)

```go
func TestShouldProcessValidInput(t *testing.T) {
    // Arrange
    input := createInput()
    require.NotNil(t, input) // MUST succeed for test to be valid
    
    // Act
    result, err := Process(input)
    
    // Assert
    assert.NoError(t, err)           // Continue even if fails
    assert.Equal(t, expected, result) // Check other assertions
}
```

## Test Naming

### Behavioral Test Names

Use **behavioral style** names that describe what the code should do:

```go
// Good - Clear behavior description
func TestShouldReturnErrorGivenInvalidUserWhenValidating(t *testing.T)
func TestShouldStoreResultGivenValidInputWhenProcessing(t *testing.T)
func TestShouldIgnoreDuplicatesGivenMultipleEntriesWhenProcessing(t *testing.T)

// Also good - Shorter variations when obvious
func TestShouldReturnErrorWhenUserIsInvalid(t *testing.T)
func TestShouldStoreResultGivenValidInput(t *testing.T)

// Bad - Vague or implementation-focused
func TestValidation(t *testing.T)
func TestProcess(t *testing.T)
func TestCheckUser(t *testing.T)
```

### Naming Pattern

Follow this pattern: `TestShould{ExpectedBehavior}Given{Context}When{Action}`

**Components:**
- **Should** - Indicates expected behavior (what the code does)
- **{ExpectedBehavior}** - The expected outcome or result
- **Given** - The initial state or preconditions (optional if obvious)
- **{Context}** - What exists or is provided
- **When** - The action being performed (optional if Given is sufficient)
- **{Action}** - What triggers the behavior

**Pattern Variations:**
```go
// Full pattern with Given and When
TestShould{X}Given{Y}When{Z}

// Just Given (when action is obvious)
TestShould{X}Given{Y}

// Just When (when context is obvious)
TestShould{X}When{Y}
```

**Examples:**
```go
// Full pattern - context + action
func TestShouldReturnErrorGivenNilInputWhenProcessing(t *testing.T)
func TestShouldMatchRouteGivenExactPathWhenNoParameters(t *testing.T)
func TestShouldApplyInOrderGivenMultipleMiddlewareWhenInvoked(t *testing.T)

// Given only - action is implied
func TestShouldCacheResultGivenValidConfiguration(t *testing.T)
func TestShouldReturnTrueGivenValidEmail(t *testing.T)

// When only - context is implied
func TestShouldReturnErrorWhenInputIsNil(t *testing.T)
func TestShouldRedirectWhenSchemeIsHTTP(t *testing.T)
```

### Special Cases

For simple getters/setters or obvious behavior:
```go
func TestNewRouter(t *testing.T)          // Constructor tests
func TestGet(t *testing.T)                // Simple getter
func TestRouteRegistration(t *testing.T)  // General functionality
```

## Test Structure

### AAA Pattern (Arrange-Act-Assert)

All tests must follow the AAA pattern with clear comment sections:

```go
func TestShouldReturnErrorGivenInvalidUserWhenValidating(t *testing.T) {
    // Arrange
    user := &User{Name: ""} // Invalid user
    validator := NewValidator()
    
    // Act
    err := validator.Validate(user)
    
    // Assert
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "name is required")
}
```

### Structure Sections

**Arrange** - Set up test data, mocks, and preconditions:
```go
// Arrange
req := httptest.NewRequest(http.MethodGet, "/test", nil)
rec := httptest.NewRecorder()
rtr := router.NewRouter()
```

**Act** - Execute the code under test:
```go
// Act
rtr.ServeHTTP(rec, req)
```

**Assert** - Verify the expected outcome:
```go
// Assert
assert.Equal(t, http.StatusOK, rec.Code)
assert.Contains(t, rec.Body.String(), "success")
```

## Test Coverage

### Coverage Requirements

- **Minimum**: 80% code coverage
- **Goal**: 90%+ coverage for critical paths
- **Middleware**: 100% coverage required

### Running Coverage

```bash
# Run tests with coverage
go test ./... -cover

# Generate HTML coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### What Coverage Doesn't Mean

**High coverage != good tests**

Focus on meaningful tests, not just hitting lines:

```go
// Bad - High coverage, low value
func TestEverything(t *testing.T) {
    DoThing()
    DoOtherThing()
    DoMoreThings()
    // No assertions!
}

// Good - Tests actual behavior
func TestShouldReturnErrorGivenInvalidInput(t *testing.T) {
    err := DoThing(nil)
    assert.Error(t, err)
}
```

## Table-Driven Tests

### When to Use

Use table-driven tests when testing multiple similar scenarios:
- Multiple input/output pairs
- Different edge cases for same function
- Parameterized behavior testing

### Pattern

```go
func TestShouldValidateInput(t *testing.T) {
    tests := []struct {
        name        string
        input       string
        expected    bool
        expectError bool
    }{
        {
            name:        "valid input",
            input:       "valid",
            expected:    true,
            expectError: false,
        },
        {
            name:        "empty input",
            input:       "",
            expected:    false,
            expectError: true,
        },
        {
            name:        "nil input",
            input:       "",
            expected:    false,
            expectError: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange (uses tt fields)
            
            // Act
            result, err := Validate(tt.input)
            
            // Assert
            if tt.expectError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expected, result)
            }
        })
    }
}
```

### Table Test Best Practices

1. **Use descriptive names** - Each test case should have a clear `name` field
2. **Keep it simple** - If setup is complex, use individual tests instead
3. **Use t.Run** - Always use subtests for each case
4. **Arrange section** - Can be outside the loop if shared

## Middleware Testing

### Middleware Test Patterns

Middleware requires testing both in isolation and within the router pipeline.

#### Pattern 1: Invoke Method Testing

Test the middleware's `Invoke` method directly:

```go
func TestShouldInvokeNextHandlerGivenValidCondition(t *testing.T) {
    // Arrange
    middleware := &myMiddleware{config: config}
    req := httptest.NewRequest(http.MethodGet, "/test", nil)
    rec := httptest.NewRecorder()
    ctx := routing.NewRouteContext(rec, req)
    
    nextCalled := false
    next := func(c routing.RouteContext) {
        nextCalled = true
    }
    
    // Act
    middleware.Invoke(ctx, next)
    
    // Assert
    assert.True(t, nextCalled)
}
```

#### Pattern 2: Router Integration Testing

Test middleware within actual router:

```go
func TestShouldApplyMiddlewareGivenRouterPipeline(t *testing.T) {
    // Arrange
    rtr := router.NewRouter()
    UseMyMiddleware(rtr, options)
    
    rtr.GET("/test", func(c routing.RouteContext) {
        c.Response().WriteHeader(http.StatusOK)
    })
    
    req := httptest.NewRequest(http.MethodGet, "/test", nil)
    rec := httptest.NewRecorder()
    
    // Act
    rtr.ServeHTTP(rec, req)
    
    // Assert
    assert.Equal(t, http.StatusOK, rec.Code)
}
```

### Middleware Coverage Requirements

All middleware must test:
- Happy path (middleware allows request through)
- Rejection path (middleware blocks/redirects request)
- Header manipulation (if applicable)
- Error conditions
- Integration with router
- Configuration options

## Helper Functions

### Test Helpers

Create helpers for common test setup:

```go
// helpers_test.go
package mypackage

import (
    "net/http/httptest"
    "net/http"
)

// Helper functions for tests
func newTestRouter() *Router {
    return NewRouter()
}

func newTestRequest(method, path string) *http.Request {
    return httptest.NewRequest(method, path, nil)
}

func newRecorder() *httptest.ResponseRecorder {
    return httptest.NewRecorder()
}
```

### Benchmark Helpers

For middleware benchmarks, use the standardized helpers in `internal/middlewarebench`:

```go
import "github.com/fgrzl/mux/internal/middlewarebench"

func BenchmarkMyMiddlewareInvoke(b *testing.B) {
    middleware := &myMiddleware{}
    middlewarebench.BenchmarkMiddlewareInvoke(b, middleware.Invoke, nil)
}
```

See [Benchmark Guidelines](bench_guidelines.md) for details.

## Common Patterns

### Testing HTTP Handlers

```go
func TestShouldReturnOKGivenMatchingRoute(t *testing.T) {
    // Arrange
    rtr := router.NewRouter()
    rtr.GET("/test", func(c routing.RouteContext) {
        c.Response().WriteHeader(http.StatusOK)
    })
    
    req := httptest.NewRequest(http.MethodGet, "/test", nil)
    rec := httptest.NewRecorder()
    
    // Act
    rtr.ServeHTTP(rec, req)
    
    // Assert
    assert.Equal(t, http.StatusOK, rec.Code)
}
```

### Testing Error Conditions

```go
func TestShouldReturnErrorGivenNilInput(t *testing.T) {
    // Arrange
    var input *Input = nil
    
    // Act
    result, err := Process(input)
    
    // Assert
    assert.Error(t, err)
    assert.Nil(t, result)
    assert.Contains(t, err.Error(), "input cannot be nil")
}
```

### Testing Panics

```go
func TestShouldPanicGivenInvalidConfig(t *testing.T) {
    // Arrange & Act & Assert
    assert.Panics(t, func() {
        NewRouter(nil) // Should panic with nil config
    })
}
```

### Testing Context Cancellation

```go
func TestShouldRespectCancellationGivenCancelledContext(t *testing.T) {
    // Arrange
    ctx, cancel := context.WithCancel(context.Background())
    cancel() // Cancel immediately
    
    // Act
    err := ProcessWithContext(ctx)
    
    // Assert
    assert.Error(t, err)
    assert.True(t, errors.Is(err, context.Canceled))
}
```

## Best Practices

### Do's

1. **Write tests first** (TDD when possible)
2. **Test behavior, not implementation**
3. **Keep tests independent** - No shared state between tests
4. **Use descriptive names** - Test names should be self-documenting
5. **Follow AAA pattern** - Always structure with Arrange/Act/Assert
6. **Test edge cases** - nil, empty, boundary values
7. **Use table-driven tests** for similar scenarios
8. **Clean up resources** - Use `defer` or `t.Cleanup()`
9. **Use subtests** - `t.Run()` for grouping related tests
10. **Keep tests fast** - Mock external dependencies

### Don'ts

1. **Don't test implementation details** - Focus on public API
2. **Don't share state between tests** - Each test should be isolated
3. **Don't skip assertions** - Always verify expected behavior
4. **Don't use sleep for timing** - Use proper synchronization
5. **Don't ignore errors** - Check error conditions explicitly
6. **Don't write brittle tests** - Avoid testing exact strings when not necessary
7. **Don't mix unit and integration tests** - Keep them separate
8. **Don't test the framework** - Trust that Go/testify works
9. **Don't duplicate test logic** - Use helpers for common setup
10. **Don't write tests just for coverage** - Write meaningful tests

### Test Organization

```go
package mypackage

import "testing"

// Unit tests - test individual functions/methods
func TestShouldDoSomething(t *testing.T) { ... }

// Integration tests - test multiple components
func TestIntegrationShouldProcessEndToEnd(t *testing.T) { ... }

// Table-driven tests
func TestShouldValidateMultipleInputs(t *testing.T) { ... }

// Benchmarks
func BenchmarkProcess(b *testing.B) { ... }
```

### File Naming

- **Test files**: `*_test.go` (e.g., `router_test.go`)
- **Benchmark files**: `*_bench_test.go` (e.g., `router_bench_test.go`)
- **Helper files**: `helpers_test.go` or `testhelpers/helpers.go`

## Running Tests

### Basic Commands

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test ./... -v

# Run specific test
go test ./internal/router -run TestShouldMatchExactRoute

# Run tests with coverage
go test ./... -cover

# Run tests in short mode (skip long tests)
go test ./... -short

# Run only benchmarks
go test ./... -bench=. -run=^$

# Run with race detection
go test ./... -race
```

### Continuous Integration

Tests must pass before merging:
```bash
# CI pipeline runs
go test ./... -v -race -cover
```

## Examples

### Complete Test Example

```go
package router

import (
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestShouldMatchExactRouteGivenNoParameters(t *testing.T) {
    // Arrange
    rtr := NewRouter()
    handlerCalled := false
    
    rtr.GET("/users", func(c RouteContext) {
        handlerCalled = true
        c.Response().WriteHeader(http.StatusOK)
    })
    
    req := httptest.NewRequest(http.MethodGet, "/users", nil)
    rec := httptest.NewRecorder()
    
    // Act
    rtr.ServeHTTP(rec, req)
    
    // Assert
    assert.True(t, handlerCalled, "handler should have been called")
    assert.Equal(t, http.StatusOK, rec.Code)
}

func TestShouldReturnNotFoundGivenNonExistentRoute(t *testing.T) {
    // Arrange
    rtr := NewRouter()
    rtr.GET("/users", func(c RouteContext) {
        c.Response().WriteHeader(http.StatusOK)
    })
    
    req := httptest.NewRequest(http.MethodGet, "/invalid", nil)
    rec := httptest.NewRecorder()
    
    // Act
    rtr.ServeHTTP(rec, req)
    
    // Assert
    assert.Equal(t, http.StatusNotFound, rec.Code)
}
```

## Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Testify Documentation](https://pkg.go.dev/github.com/stretchr/testify)
- [Effective Go - Testing](https://go.dev/doc/effective_go#testing)
- [Table Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)

---

**Remember:** Good tests are an investment in code quality and maintainability. Take the time to write clear, comprehensive tests that serve as living documentation of your code's behavior.

