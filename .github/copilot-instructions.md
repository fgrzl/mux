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


## 📚 Documentation

* All exported functions, methods, and types must have GoDoc-style comments.
* Start comments with the name of the item being documented.

## 🩵 Logging

* Use `slog` for structured logging.
* Always pass `context.Context` to log calls.

Thanks for helping us write clean and idiomatic Go!
