package main

import (
	"fmt"
	"log"

	"github.com/fgrzl/mux"
	"github.com/fgrzl/mux/pkg/openapi"
)

// User represents a user in the system with detailed field documentation
type User struct {
	ID        string   `json:"id" description:"The unique identifier for the user (UUID format)"`
	Username  string   `json:"username" description:"The user's unique username for login"`
	Email     string   `json:"email" description:"The user's email address (must be valid email format)"`
	FirstName string   `json:"firstName" description:"The user's first name"`
	LastName  string   `json:"lastName" description:"The user's last name"`
	Age       int      `json:"age" description:"The user's age in years (must be 13 or older)" minimum:"13"`
	IsActive  bool     `json:"isActive" description:"Whether the user account is currently active"`
	Roles     []string `json:"roles" description:"List of roles assigned to the user (e.g., 'admin', 'user', 'moderator')"`
}

// Product represents a product with pricing information
type Product struct {
	ID          string  `json:"id" description:"The unique product identifier (SKU or UUID)"`
	Name        string  `json:"name" description:"The display name of the product"`
	Description string  `json:"description" description:"A detailed description of the product and its features"`
	Price       float64 `json:"price" description:"The price of the product in USD" minimum:"0"`
	InStock     bool    `json:"inStock" description:"Whether the product is currently available for purchase"`
	Category    string  `json:"category" description:"The product category (e.g., 'Electronics', 'Clothing', 'Books')"`
}

func main() {
	router := mux.NewRouter()

	// Define the API routes
	router.POST("/users", createUser).
		WithOperationID("createUser").
		WithSummary("Create a new user").
		WithDescription("Creates a new user account with the provided information").
		WithJsonBody(User{}).
		WithCreatedResponse(User{}).
		WithTags("Users")

	router.GET("/users/{id}", getUser).
		WithOperationID("getUser").
		WithSummary("Get user by ID").
		WithPathParam("id", "The unique identifier of the user").
		WithOKResponse(User{}).
		WithNotFoundResponse().
		WithTags("Users")

	router.POST("/products", createProduct).
		WithOperationID("createProduct").
		WithSummary("Create a new product").
		WithJsonBody(Product{}).
		WithCreatedResponse(Product{}).
		WithTags("Products")

	// Generate OpenAPI specification
	info := &openapi.InfoObject{
		Title:       "Example API with Schema Descriptions",
		Version:     "1.0.0",
		Description: "Demonstrates how to use property-level descriptions in OpenAPI schemas",
	}

	routes, err := router.Routes()
	if err != nil {
		log.Fatalf("Failed to get routes: %v", err)
	}

	gen := openapi.NewGenerator(openapi.WithExamples())
	spec, err := gen.GenerateSpecFromRoutes(info, routes)
	if err != nil {
		log.Fatalf("Failed to generate spec: %v", err)
	}

	// Descriptions from struct tags are automatically included!
	// No need to manually enhance schemas unless you want to add
	// top-level schema descriptions or override individual properties.

	// Output the spec as JSON
	if err := spec.MarshalToFile("openapi.json"); err != nil {
		log.Fatalf("Failed to write JSON spec: %v", err)
	}
	fmt.Println("✓ Generated openapi.json with property descriptions")

	// Output the spec as YAML
	if err := spec.MarshalToFile("openapi.yaml"); err != nil {
		log.Fatalf("Failed to write YAML spec: %v", err)
	}
	fmt.Println("✓ Generated openapi.yaml with property descriptions")

	// Display a sample of the generated schema
	displaySchemaExample(spec)
}

// enhanceUserSchema adds detailed descriptions to the User schema properties
func enhanceUserSchema(spec *openapi.OpenAPISpec) {
	if spec.Components == nil || spec.Components.Schemas == nil {
		return
	}

	userSchema, exists := spec.Components.Schemas["User"]
	if !exists {
		return
	}

	// Add top-level schema description
	userSchema.Description = "Represents a user account in the system"

	// Add property-level descriptions
	if userSchema.Properties != nil {
		if id := userSchema.Properties["id"]; id != nil {
			id.Description = "The unique identifier for the user (UUID format)"
		}
		if username := userSchema.Properties["username"]; username != nil {
			username.Description = "The user's unique username for login"
		}
		if email := userSchema.Properties["email"]; email != nil {
			email.Description = "The user's email address (must be valid email format)"
		}
		if firstName := userSchema.Properties["firstName"]; firstName != nil {
			firstName.Description = "The user's first name"
		}
		if lastName := userSchema.Properties["lastName"]; lastName != nil {
			lastName.Description = "The user's last name"
		}
		if age := userSchema.Properties["age"]; age != nil {
			age.Description = "The user's age in years (must be 13 or older)"
			min := 13.0
			age.Minimum = &min
		}
		if isActive := userSchema.Properties["isActive"]; isActive != nil {
			isActive.Description = "Whether the user account is currently active"
		}
		if roles := userSchema.Properties["roles"]; roles != nil {
			roles.Description = "List of roles assigned to the user (e.g., 'admin', 'user', 'moderator')"
		}
	}
}

// displaySchemaExample shows a formatted example of the User schema with descriptions
func displaySchemaExample(spec *openapi.OpenAPISpec) {
	fmt.Println("\n" + string('=') + " Example: User Schema with Property Descriptions " + string('='))

	if spec.Components == nil || spec.Components.Schemas == nil {
		return
	}

	userSchema, exists := spec.Components.Schemas["User"]
	if !exists {
		return
	}

	fmt.Printf("\nSchema Description: %s\n\n", userSchema.Description)
	fmt.Println("Properties:")

	// Display properties with their descriptions
	properties := []string{"id", "username", "email", "firstName", "lastName", "age", "isActive", "roles"}
	for _, propName := range properties {
		if prop, ok := userSchema.Properties[propName]; ok {
			fmt.Printf("  • %-12s (%-8s): %s\n", propName, prop.Type, prop.Description)
		}
	}

	fmt.Println("\n" + string('=') + "======================================================" + string('='))
}

// Handler functions (not implemented in this example)
func createUser(c mux.RouteContext)    { c.Created(User{}) }
func getUser(c mux.RouteContext)       { c.OK(User{}) }
func createProduct(c mux.RouteContext) { c.Created(Product{}) }
