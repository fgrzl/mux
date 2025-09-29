# Copilot Instructions for Go Libraries

Welcome, Copilot. Please follow these instructions to ensure clean, idiomatic, and maintainable Go code.

## 🧠 Code Style

* Always write **idiomatic Go**. Use `gofmt`, idiomatic naming, and avoid unnecessary abstractions.
* Favor small, focused functions and minimal dependencies.
* Always use `context.Context` in public APIs where cancellation or timeouts may apply.
* Prefer `errors.Is` / `errors.As` for error handling.
* Return early and avoid deep nesting when possible.

## 🧪 Testing Guidelines

* All code must be covered with meaningful tests.

* Use Go’s standard `testing` package with `testify/assert` for assertions.

* Write tests in **behavioral style**, using names like:

  ```go
  func TestShouldReturnErrorWhenUserIsInvalid(t *testing.T)
  func TestShouldStoreResultGivenValidInput(t *testing.T)
  ```

* Inside tests, follow this structure with clear comments:

  ```go
  // Arrange
  require.NoError(t, err)
  ...

  // Act
  ...

  // Assert
  assert.Equal(t, expected, actual)
  ```

* Prefer table-driven tests when appropriate.
* Each test should test one thing

## 🧪 Benchmarking Guidance

Mux includes a comprehensive suite of benchmarks to track performance of the router, registry, middleware, and utilities. When adding new benchmarks, please follow Go’s standard conventions to keep results consistent and easy to interpret.

### Naming Conventions

* All benchmarks must start with the `Benchmark` prefix:

  ```go
  func BenchmarkRouterExactMatch(b *testing.B) { ... }
  ```
* Use **PascalCase** (no underscores, no snake_case).
* Names should clearly describe **what is being measured**:

  * `BenchmarkRouterParamMatch`
  * `BenchmarkRouteRegistryManyRoutes`
  * `BenchmarkCompressionGzipSmall`

### Sub-Benchmarks

For variations (different sizes, modes, etc.), prefer sub-benchmarks instead of encoding details into the function name:

```go
func BenchmarkCompression(b *testing.B) {
    b.Run("GzipSmall", func(b *testing.B) { ... })
    b.Run("GzipLarge", func(b *testing.B) { ... })
}
```

## 📚 Documentation

* All exported functions, methods, and types must have GoDoc-style comments.
* Start comments with the name of the item being documented.
* Maintain accurate documentation in the /docs directory
* Ensure the main README.md provides clear, concise documentation to help new developers understand how to use the project effectively.

## 🩵 Logging

* Use `slog` for structured logging.
* Always pass `context.Context` to log calls.

Thanks for helping us write clean and idiomatic Go!





