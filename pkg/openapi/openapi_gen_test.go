package openapi

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper functions for debugging
func getPropertyNames(props map[string]*Schema) []string {
	if props == nil {
		return nil
	}
	names := make([]string, 0, len(props))
	for name := range props {
		names = append(names, name)
	}
	return names
}

func getComponentNames(schemas map[string]*Schema) []string {
	if schemas == nil {
		return nil
	}
	names := make([]string, 0, len(schemas))
	for name := range schemas {
		names = append(names, name)
	}
	return names
}

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Page[T any] struct {
	Items []T `json:"items"`
	Total int `json:"total"`
}

func TestShouldGenerateSchemaForConcreteGenericType(t *testing.T) {
	// Arrange
	gen := NewGenerator()
	gen.ensureComponentInit() // Initialize components before calling GenerateSchemaForType
	page := Page[User]{
		Items: []User{{ID: 1, Name: "Alice"}},
		Total: 1,
	}

	typeOfPage := reflect.TypeOf(page)

	// Act
	schema, err := gen.GenerateSchemaForType(typeOfPage)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, schema)

	// Debug: Show the generated schema structure
	t.Logf("Generated Schema for Page[User]:")
	t.Logf("  Type: %s", schema.Type)
	t.Logf("  Properties: %v", getPropertyNames(schema.Properties))

	if itemsSchema, ok := schema.Properties["items"]; ok {
		t.Logf("  Items Schema:")
		t.Logf("    Type: %s", itemsSchema.Type)
		if itemsSchema.Items != nil {
			if itemsSchema.Items.Ref != "" {
				t.Logf("    Items Reference: %s", itemsSchema.Items.Ref)
			} else {
				t.Logf("    Items Type: %s", itemsSchema.Items.Type)
				t.Logf("    Items Properties: %v", getPropertyNames(itemsSchema.Items.Properties))
			}
		}
	}

	t.Logf("Component Schemas registered: %v", getComponentNames(gen.spec.Components.Schemas))

	// Output the actual JSON schemas
	if schemaJSON, err := json.MarshalIndent(schema, "", "  "); err == nil {
		t.Logf("Main Schema JSON:\n%s", string(schemaJSON))
	}

	if userSchema, exists := gen.spec.Components.Schemas["User"]; exists {
		if userJSON, err := json.MarshalIndent(userSchema, "", "  "); err == nil {
			t.Logf("User Component Schema JSON:\n%s", string(userJSON))
		}
	}

	// Check the Page schema structure
	assert.Equal(t, "object", schema.Type)
	assert.Contains(t, schema.Properties, "items")
	assert.Contains(t, schema.Properties, "total")

	// Check the items array property
	itemsSchema, ok := schema.Properties["items"]
	assert.True(t, ok)
	assert.Equal(t, "array", itemsSchema.Type)
	assert.NotNil(t, itemsSchema.Items)

	// The items schema should be a reference to User or contain User properties
	if itemsSchema.Items.Ref != "" {
		// It's a reference to User component
		assert.Contains(t, itemsSchema.Items.Ref, "User")
		assert.Contains(t, gen.spec.Components.Schemas, "User")
		userSchema := gen.spec.Components.Schemas["User"]
		assert.Equal(t, "object", userSchema.Type)
		assert.Contains(t, userSchema.Properties, "id")
		assert.Contains(t, userSchema.Properties, "name")
	} else {
		// It's an inline User schema
		assert.Equal(t, "object", itemsSchema.Items.Type)
		assert.Contains(t, itemsSchema.Items.Properties, "id")
		assert.Contains(t, itemsSchema.Items.Properties, "name")
	}

	// Check the total property
	totalSchema, ok := schema.Properties["total"]
	assert.True(t, ok)
	assert.Equal(t, "integer", totalSchema.Type)
}

func TestShouldRegisterNestedModelFromArrayExample(t *testing.T) {
	// Arrange
	gen := NewGenerator()
	gen.ensureComponentInit()

	info := &InfoObject{Title: "API", Version: "1.0"}

	// Create an operation whose request body is an array of User (example)
	op := &Operation{
		OperationID: "createUsers",
		RequestBody: &RequestBodyObject{
			Content: map[string]*MediaType{
				"application/json": {
					Schema:  &Schema{Ref: "#/components/schemas/UserList"},
					Example: []User{{ID: 1, Name: "Alice"}},
				},
			},
		},
	}

	routes := []RouteData{{Path: "/users", Method: "POST", Options: op}}

	// Act
	spec, err := gen.GenerateSpecFromRoutes(info, routes)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, spec)

	// The element type User should be registered as a component
	assert.Contains(t, gen.spec.Components.Schemas, "User")
	userSchema := gen.spec.Components.Schemas["User"]
	assert.Equal(t, "object", userSchema.Type)
	assert.Contains(t, userSchema.Properties, "id")
	assert.Contains(t, userSchema.Properties, "name")

	// The named ref used in the request (UserList) should also be present
	assert.Contains(t, gen.spec.Components.Schemas, "UserList")
	userList := gen.spec.Components.Schemas["UserList"]
	// UserList may be the same as User or a schema that references User
	if userList.Ref != "" {
		assert.Contains(t, userList.Ref, "User")
	} else {
		assert.Equal(t, "object", userList.Type)
	}
}

func TestShouldRegisterNestedModelFromMapValueExample(t *testing.T) {
	// Arrange
	gen := NewGenerator()
	gen.ensureComponentInit()

	info := &InfoObject{Title: "API", Version: "1.0"}

	// Create an operation whose request body is a map[string]User (example)
	op := &Operation{
		OperationID: "uploadUserMap",
		RequestBody: &RequestBodyObject{
			Content: map[string]*MediaType{
				"application/json": {
					Schema:  &Schema{Ref: "#/components/schemas/UserMap"},
					Example: map[string]User{"a": {ID: 2, Name: "Bob"}},
				},
			},
		},
	}

	routes := []RouteData{{Path: "/usermap", Method: "POST", Options: op}}

	// Act
	spec, err := gen.GenerateSpecFromRoutes(info, routes)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, spec)

	// The element type User should be registered
	assert.Contains(t, gen.spec.Components.Schemas, "User")
	userSchema := gen.spec.Components.Schemas["User"]
	assert.Equal(t, "object", userSchema.Type)
	assert.Contains(t, userSchema.Properties, "id")
	assert.Contains(t, userSchema.Properties, "name")

	// The named ref used in the request (UserMap) should be present
	assert.Contains(t, gen.spec.Components.Schemas, "UserMap")
}

func TestShouldRegisterNestedModelFromPointerToSliceExample(t *testing.T) {
	// Arrange
	gen := NewGenerator()
	gen.ensureComponentInit()

	info := &InfoObject{Title: "API", Version: "1.0"}

	// Example is a pointer to slice of User
	list := &[]User{{ID: 3, Name: "Carol"}}

	op := &Operation{
		OperationID: "createUsersPtr",
		RequestBody: &RequestBodyObject{
			Content: map[string]*MediaType{
				"application/json": {
					Schema:  &Schema{Ref: "#/components/schemas/UserPtrList"},
					Example: list,
				},
			},
		},
	}

	routes := []RouteData{{Path: "/users/ptr", Method: "POST", Options: op}}

	// Act
	spec, err := gen.GenerateSpecFromRoutes(info, routes)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, spec)

	// Ensure User was registered
	assert.Contains(t, gen.spec.Components.Schemas, "User")
}

func TestShouldRegisterNestedModelFromNestedArrayExample(t *testing.T) {
	// Arrange
	gen := NewGenerator()
	gen.ensureComponentInit()

	info := &InfoObject{Title: "API", Version: "1.0"}

	// Example is a nested array [][]User
	double := [][]User{{{ID: 4, Name: "Dave"}}}

	op := &Operation{
		OperationID: "createUsersNested",
		RequestBody: &RequestBodyObject{
			Content: map[string]*MediaType{
				"application/json": {
					Schema:  &Schema{Ref: "#/components/schemas/UserDoubleList"},
					Example: double,
				},
			},
		},
	}

	routes := []RouteData{{Path: "/users/nested", Method: "POST", Options: op}}

	// Act
	spec, err := gen.GenerateSpecFromRoutes(info, routes)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, spec)

	// Ensure User was registered
	assert.Contains(t, gen.spec.Components.Schemas, "User")
}
