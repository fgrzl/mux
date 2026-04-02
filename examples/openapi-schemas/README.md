# OpenAPI Schema Descriptions Example

This example shows how struct tags become property descriptions in the generated OpenAPI schema.

## Overview

Mux reads description and constraint tags from your Go types and carries them into the generated OpenAPI document. That keeps the happy path simple: define your routes, define your types, generate the spec.

## Features

- schema generation from Go structs
- property descriptions from struct tags
- numeric constraints from struct tags
- OpenAPI generation from the router
- JSON and YAML output

## Run It

```bash
go run .
```

The example:
1. generates `openapi.json`
2. generates `openapi.yaml`
3. prints the generated `User` schema in the console

## Key Pieces

### Model tags

```go
type User struct {
    ID        string   `json:"id" description:"The unique identifier for the user (UUID format)"`
    Username  string   `json:"username" description:"The user's unique username for login"`
    Email     string   `json:"email" description:"The user's email address (must be valid email format)"`
    FirstName string   `json:"firstName" description:"The user's first name"`
    LastName  string   `json:"lastName" description:"The user's last name"`
    Age       int      `json:"age" description:"The user's age in years (must be 13 or older)" minimum:"13"`
    IsActive  bool     `json:"isActive" description:"Whether the user account is currently active"`
    Roles     []string `json:"roles" description:"List of roles assigned to the user"`
}
```

### Spec generation

```go
router := mux.NewRouter(
    mux.WithTitle("Example API with Schema Descriptions"),
    mux.WithVersion("1.0.0"),
    mux.WithDescription("Demonstrates struct-tag-driven descriptions in generated OpenAPI schemas"),
)

gen := mux.NewGenerator(mux.WithOpenAPIExamples())
spec, err := mux.GenerateSpecWithGenerator(gen, router)
if err != nil {
    log.Fatalf("Failed to generate spec: %v", err)
}
```

### Output

```go
spec.MarshalToFile("openapi.json")
spec.MarshalToFile("openapi.yaml")
```

## Console Output

```text
Generated openapi.json with schema descriptions
Generated openapi.yaml with schema descriptions

== User Schema with Property Descriptions ==

Properties:
  - id          (string  ): The unique identifier for the user (UUID format)
  - username    (string  ): The user's unique username for login
  - email       (string  ): The user's email address (must be valid email format)
  - firstName   (string  ): The user's first name
  - lastName    (string  ): The user's last name
  - age         (integer ): The user's age in years (must be 13 or older)
  - isActive    (boolean ): Whether the user account is currently active
  - roles       (array   ): List of roles assigned to the user

============================================
```

## Best Practices

- Put human-readable descriptions on exported request and response types.
- Add constraints in struct tags when they are part of the contract.
- Generate the spec from registered routes instead of maintaining a parallel OpenAPI document by hand.

## Related Docs

- [Router](../../docs/router.md#openapi-specification)
- [Overview](../../docs/overview.md)
- [Examples](../README.md)
