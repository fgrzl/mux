package openapi

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	baruser "github.com/fgrzl/mux/internal/openapi/testtypes/bar"
	foouser "github.com/fgrzl/mux/internal/openapi/testtypes/foo"
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

func TestShouldReturnErrorWhenRouteMissingOperationID(t *testing.T) {
	// Arrange
	gen := NewGenerator()
	info := &InfoObject{Title: "API", Version: "1.0"}
	routes := []RouteData{{
		Path:   "/users",
		Method: "GET",
		Options: &Operation{
			Responses: map[string]*ResponseObject{
				"200": {Description: "ok"},
			},
		},
	}}

	// Act
	spec, err := gen.GenerateSpecFromRoutes(info, routes)

	// Assert
	require.Error(t, err)
	assert.Nil(t, spec)
	assert.Contains(t, err.Error(), "missing OperationID")
}

func TestShouldReturnErrorWhenPathParameterDoesNotDeclareInPath(t *testing.T) {
	// Arrange
	gen := NewGenerator()
	info := &InfoObject{Title: "API", Version: "1.0"}
	routes := []RouteData{{
		Path:   "/users/{id}",
		Method: "GET",
		Options: &Operation{
			OperationID: "getUser",
			Parameters: []*ParameterObject{{
				Name:     "id",
				Schema:   &Schema{Type: "string"},
				Required: true,
			}},
			Responses: map[string]*ResponseObject{
				"200": {Description: "ok"},
			},
		},
	}}

	// Act
	spec, err := gen.GenerateSpecFromRoutes(info, routes)

	// Assert
	require.Error(t, err)
	assert.Nil(t, spec)
	assert.Contains(t, err.Error(), "must declare in=\"path\"")
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

func TestShouldRegisterComponentForArrayResponse(t *testing.T) {
	// Arrange
	gen := NewGenerator()
	gen.ensureComponentInit()

	info := &InfoObject{Title: "API", Version: "1.0"}

	// Create an operation with an array response of User
	op := &Operation{
		OperationID: "getUsers",
		Responses: map[string]*ResponseObject{
			"200": {
				Description: "Success",
				Content: map[string]*MediaType{
					"application/json": {
						Schema:  &Schema{Type: "array", Items: &Schema{Ref: "#/components/schemas/User"}},
						Example: []User{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}},
					},
				},
			},
		},
	}

	routes := []RouteData{{Path: "/users", Method: "GET", Options: op}}

	// Act
	spec, err := gen.GenerateSpecFromRoutes(info, routes)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, spec)

	// Verify the User component was registered
	assert.Contains(t, gen.spec.Components.Schemas, "User", "User component should be registered from array items")
	userSchema := gen.spec.Components.Schemas["User"]
	assert.Equal(t, "object", userSchema.Type)
	assert.Contains(t, userSchema.Properties, "id")
	assert.Contains(t, userSchema.Properties, "name")

	// Verify the response schema is an array with a ref to User
	pathItem := spec.Paths["/users"]
	require.NotNil(t, pathItem)
	getOp := pathItem.Get
	require.NotNil(t, getOp)
	resp := getOp.Responses["200"]
	require.NotNil(t, resp)
	mediaType := resp.Content["application/json"]
	require.NotNil(t, mediaType)
	assert.Equal(t, "array", mediaType.Schema.Type)
	assert.NotNil(t, mediaType.Schema.Items)
	assert.Equal(t, "#/components/schemas/User", mediaType.Schema.Items.Ref)
}

func TestShouldNotLeakSpecStateWhenReusingGenerator(t *testing.T) {
	// Arrange
	gen := NewGenerator()

	routesOne := []RouteData{{
		Path:   "/users",
		Method: "GET",
		Options: &Operation{
			OperationID: "getUsers",
			Responses: map[string]*ResponseObject{
				"200": {Description: "Success"},
			},
		},
	}}
	routesTwo := []RouteData{{
		Path:   "/orders",
		Method: "GET",
		Options: &Operation{
			OperationID: "getOrders",
			Responses: map[string]*ResponseObject{
				"200": {Description: "Success"},
			},
		},
	}}

	// Act
	firstSpec, err := gen.GenerateSpecFromRoutes(&InfoObject{Title: "Users API", Version: "1.0.0"}, routesOne)
	require.NoError(t, err)
	secondSpec, err := gen.GenerateSpecFromRoutes(&InfoObject{Title: "Orders API", Version: "2.0.0"}, routesTwo)
	require.NoError(t, err)

	// Assert
	assert.Contains(t, firstSpec.Paths, "/users")
	assert.NotContains(t, firstSpec.Paths, "/orders")
	assert.Equal(t, "Users API", firstSpec.Info.Title)

	assert.Contains(t, secondSpec.Paths, "/orders")
	assert.NotContains(t, secondSpec.Paths, "/users")
	assert.Equal(t, "Orders API", secondSpec.Info.Title)
}

func TestShouldGenerateSupportedNonCrudOperations(t *testing.T) {
	// Arrange
	gen := NewGenerator()
	info := &InfoObject{Title: "API", Version: "1.0"}
	routes := []RouteData{
		{
			Path:   "/system",
			Method: "HEAD",
			Options: &Operation{
				OperationID: "headSystem",
				Responses:   map[string]*ResponseObject{"200": {Description: "OK"}},
			},
		},
		{
			Path:   "/system",
			Method: "OPTIONS",
			Options: &Operation{
				OperationID: "optionsSystem",
				Responses:   map[string]*ResponseObject{"200": {Description: "OK"}},
			},
		},
		{
			Path:   "/system",
			Method: "TRACE",
			Options: &Operation{
				OperationID: "traceSystem",
				Responses:   map[string]*ResponseObject{"200": {Description: "OK"}},
			},
		},
	}

	// Act
	spec, err := gen.GenerateSpecFromRoutes(info, routes)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, spec)
	pathItem := spec.Paths["/system"]
	require.NotNil(t, pathItem)
	require.NotNil(t, pathItem.Head)
	require.NotNil(t, pathItem.Options)
	require.NotNil(t, pathItem.Trace)
	assert.Equal(t, "headSystem", pathItem.Head.OperationID)
	assert.Equal(t, "optionsSystem", pathItem.Options.OperationID)
	assert.Equal(t, "traceSystem", pathItem.Trace.OperationID)
}

func TestShouldReturnErrorForUnsupportedMethodInOpenAPIGeneration(t *testing.T) {
	// Arrange
	gen := NewGenerator()
	info := &InfoObject{Title: "API", Version: "1.0"}
	routes := []RouteData{{
		Path:   "/tunnel",
		Method: "CONNECT",
		Options: &Operation{
			OperationID: "connectTunnel",
			Responses:   map[string]*ResponseObject{"200": {Description: "OK"}},
		},
	}}

	// Act
	spec, err := gen.GenerateSpecFromRoutes(info, routes)

	// Assert
	require.Error(t, err)
	assert.Nil(t, spec)
	assert.Contains(t, err.Error(), "unsupported HTTP method")
	assert.Contains(t, err.Error(), "CONNECT")
}

func TestGenerateSpecFromRoutesShouldOwnInfoObject(t *testing.T) {
	// Arrange
	gen := NewGenerator()
	info := &InfoObject{
		Title:   "API",
		Version: "1.0",
		Contact: &ContactObject{Name: "Support"},
		License: &LicenseObject{Name: "MIT"},
		Extensions: map[string]any{
			"x-meta": map[string]any{"env": "prod"},
		},
	}
	routes := []RouteData{{
		Path:   "/health",
		Method: "GET",
		Options: &Operation{
			OperationID: "getHealth",
			Responses:   map[string]*ResponseObject{"200": {Description: "OK"}},
		},
	}}

	// Act
	spec, err := gen.GenerateSpecFromRoutes(info, routes)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, spec)
	require.NotNil(t, spec.Info)
	info.Title = "Changed caller"
	info.Contact.Name = "Changed caller contact"
	info.License.Name = "Changed caller license"
	info.Extensions["x-meta"].(map[string]any)["env"] = "dev"
	assert.Equal(t, "API", spec.Info.Title)
	assert.Equal(t, "Support", spec.Info.Contact.Name)
	assert.Equal(t, "MIT", spec.Info.License.Name)
	assert.Equal(t, "prod", spec.Info.Extensions["x-meta"].(map[string]any)["env"])

	spec.Info.Title = "Changed spec"
	spec.Info.Contact.Name = "Changed spec contact"
	spec.Info.License.Name = "Changed spec license"
	spec.Info.Extensions["x-meta"].(map[string]any)["env"] = "staging"
	assert.Equal(t, "Changed caller", info.Title)
	assert.Equal(t, "Changed caller contact", info.Contact.Name)
	assert.Equal(t, "Changed caller license", info.License.Name)
	assert.Equal(t, "dev", info.Extensions["x-meta"].(map[string]any)["env"])
	assert.NotSame(t, info, spec.Info)
	assert.NotSame(t, info.Contact, spec.Info.Contact)
	assert.NotSame(t, info.License, spec.Info.License)
}

func TestShouldDisambiguateSanitizedComponentNameCollisions(t *testing.T) {
	// Arrange
	gen := NewGenerator()
	gen.ensureComponentInit()

	info := &InfoObject{Title: "API", Version: "1.0"}

	fooPage := Page[foouser.User]{Items: []foouser.User{{Foo: "alpha"}}, Total: 1}
	barPage := Page[baruser.User]{Items: []baruser.User{{Bar: 42}}, Total: 1}

	routes := []RouteData{
		{
			Path:   "/foo-users",
			Method: "POST",
			Options: &Operation{
				OperationID: "createFooUsers",
				RequestBody: &RequestBodyObject{
					Content: map[string]*MediaType{
						"application/json": {
							Schema:  &Schema{Ref: "#/components/schemas/FooPage"},
							Example: fooPage,
						},
					},
				},
			},
		},
		{
			Path:   "/bar-users",
			Method: "POST",
			Options: &Operation{
				OperationID: "createBarUsers",
				RequestBody: &RequestBodyObject{
					Content: map[string]*MediaType{
						"application/json": {
							Schema:  &Schema{Ref: "#/components/schemas/BarPage"},
							Example: barPage,
						},
					},
				},
			},
		},
	}

	// Act
	spec, err := gen.GenerateSpecFromRoutes(info, routes)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, spec)

	fooComponent := spec.Components.Schemas["FooPage"]
	barComponent := spec.Components.Schemas["BarPage"]
	require.NotNil(t, fooComponent)
	require.NotNil(t, barComponent)
	require.Contains(t, fooComponent.Properties, "items")
	require.Contains(t, barComponent.Properties, "items")
	require.NotNil(t, fooComponent.Properties["items"].Items)
	require.NotNil(t, barComponent.Properties["items"].Items)

	fooRef := strings.TrimPrefix(fooComponent.Properties["items"].Items.Ref, "#/components/schemas/")
	barRef := strings.TrimPrefix(barComponent.Properties["items"].Items.Ref, "#/components/schemas/")
	require.NotEmpty(t, fooRef)
	require.NotEmpty(t, barRef)
	assert.NotEqual(t, fooRef, barRef)

	fooUserSchema := spec.Components.Schemas[fooRef]
	barUserSchema := spec.Components.Schemas[barRef]
	require.NotNil(t, fooUserSchema)
	require.NotNil(t, barUserSchema)
	assert.Contains(t, fooUserSchema.Properties, "foo")
	assert.NotContains(t, fooUserSchema.Properties, "bar")
	assert.Contains(t, barUserSchema.Properties, "bar")
	assert.NotContains(t, barUserSchema.Properties, "foo")
}

type Dog struct {
	Bark string `json:"bark"`
}

type Cat struct {
	Meow string `json:"meow"`
}

type PointerUserEnvelope struct {
	Primary *User    `json:"primary"`
	Labels  []string `json:"labels"`
}

func TestShouldNotMutateOperationInputsWhenGeneratingSpecWithoutExamples(t *testing.T) {
	// Arrange
	gen := NewGenerator()
	info := &InfoObject{Title: "API", Version: "1.0"}
	page := Page[User]{Items: []User{{ID: 1, Name: "Alice"}}, Total: 1}
	requestRef := "#/components/schemas/Page[github.com/acme/project/pkg.User]"
	responseRef := "#/components/schemas/Page[github.com/acme/project/pkg.User]"

	op := &Operation{
		OperationID: "createUsers",
		RequestBody: &RequestBodyObject{
			Content: map[string]*MediaType{
				"application/json": {
					Schema:   &Schema{Ref: requestRef},
					Example:  page,
					Examples: map[string]*ExampleObject{"default": {Value: page}},
				},
			},
		},
		Responses: map[string]*ResponseObject{
			"200": {
				Content: map[string]*MediaType{
					"application/json": {
						Schema:   &Schema{Ref: responseRef},
						Example:  page,
						Examples: map[string]*ExampleObject{"default": {Value: page}},
					},
				},
			},
		},
	}

	// Act
	spec, err := gen.GenerateSpecFromRoutes(info, []RouteData{{Path: "/users", Method: "POST", Options: op}})

	// Assert
	require.NoError(t, err)
	require.NotNil(t, spec)
	assert.Equal(t, requestRef, op.RequestBody.Content["application/json"].Schema.Ref)
	assert.Equal(t, page, op.RequestBody.Content["application/json"].Example)
	assert.Contains(t, op.RequestBody.Content["application/json"].Examples, "default")
	assert.Equal(t, "", op.Responses["200"].Description)
	assert.Equal(t, responseRef, op.Responses["200"].Content["application/json"].Schema.Ref)
	assert.Equal(t, page, op.Responses["200"].Content["application/json"].Example)
	assert.Contains(t, op.Responses["200"].Content["application/json"].Examples, "default")

	postOp := spec.Paths["/users"].Post
	require.NotNil(t, postOp)
	assert.Equal(t, "Success", postOp.Responses["200"].Description)
	assert.Equal(t, "#/components/schemas/PageUser", postOp.RequestBody.Content["application/json"].Schema.Ref)
	assert.Nil(t, postOp.RequestBody.Content["application/json"].Example)
	assert.Nil(t, postOp.Responses["200"].Content["application/json"].Example)
}

func TestShouldNotMutateSchemaInputsWhenGeneratingSpecWithExamples(t *testing.T) {
	// Arrange
	gen := NewGenerator(WithExamples())
	info := &InfoObject{Title: "API", Version: "1.0"}
	page := Page[User]{Items: []User{{ID: 1, Name: "Alice"}}, Total: 1}
	schemaRef := "#/components/schemas/Page[github.com/acme/project/pkg.User]"
	requestSchema := &Schema{Ref: schemaRef}

	op := &Operation{
		OperationID: "createUsersWithExamples",
		RequestBody: &RequestBodyObject{
			Content: map[string]*MediaType{
				"application/json": {
					Schema:  requestSchema,
					Example: page,
				},
			},
		},
		Responses: map[string]*ResponseObject{
			"201": {Description: "Created"},
		},
	}

	// Act
	spec, err := gen.GenerateSpecFromRoutes(info, []RouteData{{Path: "/users", Method: "POST", Options: op}})

	// Assert
	require.NoError(t, err)
	require.NotNil(t, spec)
	assert.Equal(t, schemaRef, requestSchema.Ref)
	assert.Nil(t, requestSchema.Example)

	media := spec.Paths["/users"].Post.RequestBody.Content["application/json"]
	require.NotNil(t, media)
	assert.Equal(t, page, media.Schema.Example)
	assert.Equal(t, "#/components/schemas/PageUser", media.Schema.Ref)
}

func TestShouldNotMutateCompositeSubSchemasWhenGeneratingSpecWithoutExamples(t *testing.T) {
	// Arrange
	gen := NewGenerator()
	info := &InfoObject{Title: "API", Version: "1.0"}
	dogExample := Dog{Bark: "woof"}
	catExample := Cat{Meow: "meow"}
	dogSchema := &Schema{Ref: "#/components/schemas/Dog", Example: dogExample}
	catSchema := &Schema{Ref: "#/components/schemas/Cat", Example: catExample}

	op := &Operation{
		OperationID: "createPet",
		RequestBody: &RequestBodyObject{
			Content: map[string]*MediaType{
				"application/json": {
					Schema: &Schema{OneOf: []*Schema{dogSchema, catSchema}},
				},
			},
		},
		Responses: map[string]*ResponseObject{
			"201": {Description: "Created"},
		},
	}

	// Act
	spec, err := gen.GenerateSpecFromRoutes(info, []RouteData{{Path: "/pets", Method: "POST", Options: op}})

	// Assert
	require.NoError(t, err)
	require.NotNil(t, spec)
	assert.Equal(t, dogExample, dogSchema.Example)
	assert.Equal(t, catExample, catSchema.Example)
	assert.Equal(t, "#/components/schemas/Dog", dogSchema.Ref)
	assert.Equal(t, "#/components/schemas/Cat", catSchema.Ref)

	oneOf := spec.Paths["/pets"].Post.RequestBody.Content["application/json"].Schema.OneOf
	require.Len(t, oneOf, 2)
	assert.Nil(t, oneOf[0].Example)
	assert.Nil(t, oneOf[1].Example)
}

func TestShouldRewriteRefsWithinCompositeSchemas(t *testing.T) {
	// Arrange
	schema := &Schema{
		OneOf: []*Schema{
			{Ref: "#/components/schemas/Page[github.com/acme/project/pkg.User]"},
		},
		AnyOf: []*Schema{
			{
				Properties: map[string]*Schema{
					"payload": {Ref: "#/components/schemas/Foo[github.com/acme/project/pkg.User]"},
				},
			},
		},
		AllOf: []*Schema{
			{
				AnyOf: []*Schema{
					{Items: &Schema{Ref: "#/components/schemas/Bar[github.com/acme/project/pkg.User]"}},
				},
			},
		},
	}
	nameMap := map[string]string{
		"Page[github.com/acme/project/pkg.User]": "PageUser",
		"Foo[github.com/acme/project/pkg.User]":  "FooUser",
		"Bar[github.com/acme/project/pkg.User]":  "BarUser",
	}

	// Act
	rewriteSchemaRefs(schema, nameMap)

	// Assert
	require.Len(t, schema.OneOf, 1)
	require.Len(t, schema.AnyOf, 1)
	require.Len(t, schema.AllOf, 1)
	require.Len(t, schema.AllOf[0].AnyOf, 1)
	assert.Equal(t, "#/components/schemas/PageUser", schema.OneOf[0].Ref)
	assert.Equal(t, "#/components/schemas/FooUser", schema.AnyOf[0].Properties["payload"].Ref)
	assert.Equal(t, "#/components/schemas/BarUser", schema.AllOf[0].AnyOf[0].Items.Ref)
}

func TestShouldNotAliasPointerBackedExamplesWhenGeneratingSpecWithExamples(t *testing.T) {
	// Arrange
	gen := NewGenerator(WithExamples())
	info := &InfoObject{Title: "API", Version: "1.0"}
	payload := &PointerUserEnvelope{
		Primary: &User{ID: 1, Name: "Alice"},
		Labels:  []string{"alpha"},
	}
	requestSchema := &Schema{Ref: "#/components/schemas/PointerUserEnvelope"}

	op := &Operation{
		OperationID: "createPointerEnvelope",
		RequestBody: &RequestBodyObject{
			Content: map[string]*MediaType{
				"application/json": {
					Schema:  requestSchema,
					Example: payload,
				},
			},
		},
		Responses: map[string]*ResponseObject{
			"201": {Description: "Created"},
		},
	}

	// Act
	spec, err := gen.GenerateSpecFromRoutes(info, []RouteData{{Path: "/envelopes", Method: "POST", Options: op}})

	// Assert
	require.NoError(t, err)
	require.NotNil(t, spec)
	media := spec.Paths["/envelopes"].Post.RequestBody.Content["application/json"]
	require.NotNil(t, media)
	clonedPayload, ok := media.Schema.Example.(*PointerUserEnvelope)
	require.True(t, ok)
	require.NotNil(t, clonedPayload.Primary)
	assert.NotSame(t, payload, clonedPayload)
	assert.NotSame(t, payload.Primary, clonedPayload.Primary)
	clonedPayload.Primary.Name = "Bob"
	clonedPayload.Labels[0] = "beta"
	assert.Equal(t, "Alice", payload.Primary.Name)
	assert.Equal(t, []string{"alpha"}, payload.Labels)
	assert.Equal(t, "#/components/schemas/PointerUserEnvelope", requestSchema.Ref)
}
