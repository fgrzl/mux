# OpenAPI Schema Descriptions Example

This example demonstrates how to use property-level descriptions in OpenAPI schemas when generating API documentation with Mux.

## Overview

Property-level descriptions allow you to document each field in your data models, making your API documentation more comprehensive and user-friendly. This is especially important for:

- **Developer Experience**: Clear field documentation helps API consumers understand what each property represents
- **API Documentation Tools**: Tools like Swagger UI, Redoc, and Stoplight display these descriptions
- **Code Generation**: Many OpenAPI code generators use these descriptions to generate comments in client libraries

## Features Demonstrated

- **Schema Generation**: Automatic schema generation from Go structs
- **Property Descriptions**: Adding detailed descriptions to individual properties
- **Schema-level Descriptions**: Adding descriptions to the schema itself
- **Type Constraints**: Setting minimum/maximum values for numeric fields
- **Multiple Output Formats**: Generating both JSON and YAML OpenAPI specs

## Running the Example

```bash
# From the openapi-schemas directory
go run main.go
```

This will:
1. Generate `openapi.json` with property-level descriptions
2. Generate `openapi.yaml` with property-level descriptions
3. Display a formatted example of the schema in the console

## Code Walkthrough

### 1. Define Your Data Models

```go
type User struct {
    ID        string   `json:"id"`
    Username  string   `json:"username"`
    Email     string   `json:"email"`
    FirstName string   `json:"firstName"`
    LastName  string   `json:"lastName"`
    Age       int      `json:"age"`
    IsActive  bool     `json:"isActive"`
    Roles     []string `json:"roles"`
}
```

### 2. Generate the OpenAPI Spec

```go
gen := openapi.NewGenerator(openapi.WithExamples())
spec, err := gen.GenerateSpecFromRoutes(info, routes)
if err != nil {
    log.Fatalf("Failed to generate spec: %v", err)
}
```

### 3. Enhance with Property Descriptions

```go
func enhanceUserSchema(spec *openapi.OpenAPISpec) {
    userSchema := spec.Components.Schemas["User"]
    
    // Add schema-level description
    userSchema.Description = "Represents a user account in the system"
    
    // Add property-level descriptions
    if id := userSchema.Properties["id"]; id != nil {
        id.Description = "The unique identifier for the user (UUID format)"
    }
    if username := userSchema.Properties["username"]; username != nil {
        username.Description = "The user's unique username for login"
    }
    // ... more properties
}
```

### 4. Save the Spec

```go
// Output as JSON
spec.MarshalToFile("openapi.json")

// Output as YAML
spec.MarshalToFile("openapi.yaml")
```

## Generated Schema Example

When you run the example, you'll see output like this:

```
✓ Generated openapi.json with property descriptions
✓ Generated openapi.yaml with property descriptions

= Example: User Schema with Property Descriptions =

Schema Description: Represents a user account in the system

Properties:
  • id          (string  ): The unique identifier for the user (UUID format)
  • username    (string  ): The user's unique username for login
  • email       (string  ): The user's email address (must be valid email format)
  • firstName   (string  ): The user's first name
  • lastName    (string  ): The user's last name
  • age         (integer ): The user's age in years (must be 13 or older)
  • isActive    (boolean ): Whether the user account is currently active
  • roles       (array   ): List of roles assigned to the user

========================================================
```

## Viewing the Generated Documentation

### With Swagger UI

You can view the generated OpenAPI spec with Swagger UI:

1. Visit [Swagger Editor](https://editor.swagger.io/)
2. Click "File" → "Import file"
3. Select the generated `openapi.yaml` or `openapi.json`
4. View the property descriptions in the schema section

### With Redoc

```bash
# Install redoc-cli
npm install -g redoc-cli

# Generate HTML documentation
redoc-cli bundle openapi.yaml -o docs.html

# Open in browser
open docs.html
```

## Key Concepts

### Schema vs Property Descriptions

- **Schema Description**: Describes what the entire object represents
  ```go
  userSchema.Description = "Represents a user account in the system"
  ```

- **Property Description**: Describes what an individual field means
  ```go
  id.Description = "The unique identifier for the user (UUID format)"
  ```

### Adding Constraints

You can also add validation constraints that appear in the documentation:

```go
if age := userSchema.Properties["age"]; age != nil {
    age.Description = "The user's age in years (must be 13 or older)"
    min := 13.0
    age.Minimum = &min
}
```

## Best Practices

1. **Be Specific**: Include format requirements, examples, and constraints
   - ❌ "User ID"
   - ✅ "The unique identifier for the user (UUID format)"

2. **Document Required Fields**: Clearly state which fields are mandatory

3. **Explain Relationships**: If a field references another resource, mention it

4. **Include Value Sets**: For fields with limited values, list them
   - Example: "User role (one of: 'admin', 'user', 'guest')"

5. **Mention Validation Rules**: Include any validation constraints
   - Example: "Age must be between 13 and 120"

## Related Documentation

- [Mux OpenAPI Generation](../../docs/router.md#openapi-specification)
- [OpenAPI Specification](https://spec.openapis.org/oas/v3.1.0)
- [JSON Schema Documentation](https://json-schema.org/)

## Learn More

Check out the other examples in the [examples directory](../README.md) to see more Mux features in action.
