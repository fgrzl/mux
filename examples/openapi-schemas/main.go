package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/fgrzl/mux"
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
	router := mux.NewRouter(
		mux.WithTitle("Example API with Schema Descriptions"),
		mux.WithVersion("1.0.0"),
		mux.WithDescription("Demonstrates struct-tag-driven descriptions in generated OpenAPI schemas"),
	)

	registerRoutes(router)

	gen := mux.NewGenerator(mux.WithOpenAPIExamples())
	spec, err := mux.GenerateSpecWithGenerator(gen, router)
	if err != nil {
		log.Fatalf("Failed to generate spec: %v", err)
	}

	if err := spec.MarshalToFile("openapi.json"); err != nil {
		log.Fatalf("Failed to write JSON spec: %v", err)
	}
	fmt.Println("Generated openapi.json with schema descriptions")

	if err := spec.MarshalToFile("openapi.yaml"); err != nil {
		log.Fatalf("Failed to write YAML spec: %v", err)
	}
	fmt.Println("Generated openapi.yaml with schema descriptions")

	displaySchemaExample(spec)
}

func registerRoutes(router *mux.Router) {
	createUserRoute := router.POST("/users", createUser)
	createUserRoute.
		WithOperationID("createUser").
		WithSummary("Create a new user").
		WithDescription("Creates a new user account with the provided information").
		WithTags("Users").
		WithJsonBody(User{}).
		WithCreatedResponse(User{})

	getUserRoute := router.GET("/users/{id}", getUser)
	getUserRoute.
		WithOperationID("getUser").
		WithSummary("Get user by ID").
		WithTags("Users").
		WithPathParam("id", "The unique identifier of the user", "user-123").
		WithOKResponse(User{}).
		WithResponse(404, mux.ProblemDetails{})

	createProductRoute := router.POST("/products", createProduct)
	createProductRoute.
		WithOperationID("createProduct").
		WithSummary("Create a new product").
		WithTags("Products").
		WithJsonBody(Product{}).
		WithCreatedResponse(Product{})
}

func displaySchemaExample(spec *mux.OpenAPISpec) {
	fmt.Println("\n== User Schema with Property Descriptions ==")

	data, err := json.Marshal(spec)
	if err != nil {
		log.Printf("failed to inspect generated spec: %v", err)
		return
	}

	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		log.Printf("failed to decode generated spec: %v", err)
		return
	}

	components, ok := doc["components"].(map[string]any)
	if !ok {
		return
	}
	schemas, ok := components["schemas"].(map[string]any)
	if !ok {
		return
	}
	userSchema, ok := schemas["User"].(map[string]any)
	if !ok {
		return
	}
	propertiesMap, ok := userSchema["properties"].(map[string]any)
	if !ok {
		return
	}

	fmt.Println("\nProperties:")

	// Display properties with their descriptions
	properties := []string{"id", "username", "email", "firstName", "lastName", "age", "isActive", "roles"}
	for _, propName := range properties {
		prop, ok := propertiesMap[propName].(map[string]any)
		if ok {
			fmt.Printf("  - %-12s (%-8v): %v\n", propName, prop["type"], prop["description"])
		}
	}

	fmt.Println("\n============================================")
}

// Handler functions (not implemented in this example)
func createUser(c mux.RouteContext)    { c.Created(User{}) }
func getUser(c mux.RouteContext)       { c.OK(User{}) }
func createProduct(c mux.RouteContext) { c.Created(Product{}) }
