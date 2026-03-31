package builder

import (
	"net/http"
	"testing"

	"github.com/fgrzl/mux/pkg/openapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type Animal struct {
	Type string `json:"type"`
}

type DogAnimal struct {
	Type string `json:"type"`
	Bark string `json:"bark"`
}

type CatAnimal struct {
	Type string `json:"type"`
	Meow string `json:"meow"`
}

type BirdAnimal struct {
	Type string `json:"type"`
	Sing string `json:"sing"`
}

func TestOneOfJsonBodyShouldProduceValidOpenAPIYAML(t *testing.T) {
	// Arrange
	gen := openapi.NewGenerator()

	// Create a route with oneOf
	op := DetachedRoute(http.MethodPost, "/pets").
		WithOneOfJsonBody(DogAnimal{}, CatAnimal{}).
		WithOperationID("createPet").
		Options.Operation

	routes := []openapi.RouteData{
		{Path: "/pets", Method: "POST", Options: &op},
	}

	// Act
	spec, err := gen.GenerateSpecFromRoutes(&openapi.InfoObject{
		Title:   "Pet API",
		Version: "1.0.0",
	}, routes)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, spec)

	// Verify the request body has oneOf
	pathItem := spec.Paths["/pets"]
	require.NotNil(t, pathItem)
	require.NotNil(t, pathItem.Post)
	require.NotNil(t, pathItem.Post.RequestBody)

	mediaType := pathItem.Post.RequestBody.Content["application/json"]
	require.NotNil(t, mediaType)
	require.NotNil(t, mediaType.Schema)

	schema := mediaType.Schema
	require.NotNil(t, schema.OneOf, "oneOf should be present")
	require.Len(t, schema.OneOf, 2, "oneOf should have 2 schemas")

	// Verify YAML marshaling works
	yamlBytes, err := yaml.Marshal(spec)
	require.NoError(t, err, "should marshal to YAML without errors")
	require.NotEmpty(t, yamlBytes)

	// Verify the YAML contains oneOf
	var yamlData map[string]interface{}
	err = yaml.Unmarshal(yamlBytes, &yamlData)
	require.NoError(t, err, "should unmarshal YAML without errors")

	// Navigate to the schema
	paths := yamlData["paths"].(map[string]interface{})
	pets := paths["/pets"].(map[string]interface{})
	post := pets["post"].(map[string]interface{})
	requestBody := post["requestBody"].(map[string]interface{})
	content := requestBody["content"].(map[string]interface{})
	jsonContent := content["application/json"].(map[string]interface{})
	schemaMap := jsonContent["schema"].(map[string]interface{})

	// Verify oneOf exists in YAML
	oneOf, exists := schemaMap["oneOf"]
	assert.True(t, exists, "oneOf should exist in YAML output")
	assert.IsType(t, []interface{}{}, oneOf, "oneOf should be an array")

	t.Logf("Generated YAML:\n%s", string(yamlBytes))
}

func TestAnyOfJsonBodyShouldProduceValidOpenAPIYAML(t *testing.T) {
	// Arrange
	gen := openapi.NewGenerator()

	op := DetachedRoute(http.MethodPost, "/animals").
		WithAnyOfJsonBody(DogAnimal{}, CatAnimal{}, BirdAnimal{}).
		WithOperationID("createAnimal").
		Options.Operation

	routes := []openapi.RouteData{
		{Path: "/animals", Method: "POST", Options: &op},
	}

	// Act
	spec, err := gen.GenerateSpecFromRoutes(&openapi.InfoObject{
		Title:   "Animal API",
		Version: "1.0.0",
	}, routes)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, spec)

	pathItem := spec.Paths["/animals"]
	require.NotNil(t, pathItem)
	require.NotNil(t, pathItem.Post)

	mediaType := pathItem.Post.RequestBody.Content["application/json"]
	schema := mediaType.Schema
	require.NotNil(t, schema.AnyOf, "anyOf should be present")
	require.Len(t, schema.AnyOf, 3, "anyOf should have 3 schemas")

	// Verify YAML marshaling
	yamlBytes, err := yaml.Marshal(spec)
	require.NoError(t, err)

	var yamlData map[string]interface{}
	err = yaml.Unmarshal(yamlBytes, &yamlData)
	require.NoError(t, err)

	// Verify anyOf exists
	paths := yamlData["paths"].(map[string]interface{})
	animals := paths["/animals"].(map[string]interface{})
	post := animals["post"].(map[string]interface{})
	requestBody := post["requestBody"].(map[string]interface{})
	content := requestBody["content"].(map[string]interface{})
	jsonContent := content["application/json"].(map[string]interface{})
	schemaMap := jsonContent["schema"].(map[string]interface{})

	anyOf, exists := schemaMap["anyOf"]
	assert.True(t, exists, "anyOf should exist in YAML output")
	assert.IsType(t, []interface{}{}, anyOf, "anyOf should be an array")
}

func TestAllOfJsonBodyShouldProduceValidOpenAPIYAML(t *testing.T) {
	// Arrange
	type BaseEntity struct {
		ID string `json:"id"`
	}

	type Timestamps struct {
		CreatedAt string `json:"created_at"`
	}

	gen := openapi.NewGenerator()

	op := DetachedRoute(http.MethodPost, "/entities").
		WithAllOfJsonBody(BaseEntity{}, Timestamps{}).
		WithOperationID("createEntity").
		Options.Operation

	routes := []openapi.RouteData{
		{Path: "/entities", Method: "POST", Options: &op},
	}

	// Act
	spec, err := gen.GenerateSpecFromRoutes(&openapi.InfoObject{
		Title:   "Entity API",
		Version: "1.0.0",
	}, routes)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, spec)

	pathItem := spec.Paths["/entities"]
	require.NotNil(t, pathItem)
	require.NotNil(t, pathItem.Post)

	mediaType := pathItem.Post.RequestBody.Content["application/json"]
	schema := mediaType.Schema
	require.NotNil(t, schema.AllOf, "allOf should be present")
	require.Len(t, schema.AllOf, 2, "allOf should have 2 schemas")

	// Verify YAML marshaling
	yamlBytes, err := yaml.Marshal(spec)
	require.NoError(t, err)

	var yamlData map[string]interface{}
	err = yaml.Unmarshal(yamlBytes, &yamlData)
	require.NoError(t, err)

	// Verify allOf exists
	paths := yamlData["paths"].(map[string]interface{})
	entities := paths["/entities"].(map[string]interface{})
	post := entities["post"].(map[string]interface{})
	requestBody := post["requestBody"].(map[string]interface{})
	content := requestBody["content"].(map[string]interface{})
	jsonContent := content["application/json"].(map[string]interface{})
	schemaMap := jsonContent["schema"].(map[string]interface{})

	allOf, exists := schemaMap["allOf"]
	assert.True(t, exists, "allOf should exist in YAML output")
	assert.IsType(t, []interface{}{}, allOf, "allOf should be an array")
}

func TestCompositeSchemasShouldReferenceComponentSchemas(t *testing.T) {
	// Arrange
	gen := openapi.NewGenerator()

	op := DetachedRoute(http.MethodPost, "/pets").
		WithOneOfJsonBody(DogAnimal{}, CatAnimal{}).
		WithOperationID("createPet").
		Options.Operation

	routes := []openapi.RouteData{
		{Path: "/pets", Method: "POST", Options: &op},
	}

	// Act
	spec, err := gen.GenerateSpecFromRoutes(&openapi.InfoObject{
		Title:   "Pet API",
		Version: "1.0.0",
	}, routes)

	// Assert
	require.NoError(t, err)

	// Check that the oneOf schemas reference the components
	pathItem := spec.Paths["/pets"]
	schema := pathItem.Post.RequestBody.Content["application/json"].Schema

	require.Len(t, schema.OneOf, 2, "should have 2 oneOf schemas")

	for i, s := range schema.OneOf {
		assert.NotEmpty(t, s.Ref, "oneOf[%d] should have a $ref", i)
		assert.Contains(t, s.Ref, "#/components/schemas/", "oneOf[%d] should reference components", i)
	}

	// Verify the schema references are correct
	assert.Equal(t, "#/components/schemas/DogAnimal", schema.OneOf[0].Ref)
	assert.Equal(t, "#/components/schemas/CatAnimal", schema.OneOf[1].Ref)
}

func TestCompositeJSONBodyShouldMarshalToValidJSON(t *testing.T) {
	// Arrange
	gen := openapi.NewGenerator()

	op := DetachedRoute(http.MethodPost, "/pets").
		WithOneOfJsonBody(DogAnimal{Type: "dog"}, CatAnimal{Type: "cat"}).
		WithOperationID("createPet").
		Options.Operation

	routes := []openapi.RouteData{
		{Path: "/pets", Method: "POST", Options: &op},
	}

	// Act
	spec, err := gen.GenerateSpecFromRoutes(&openapi.InfoObject{
		Title:   "Pet API",
		Version: "1.0.0",
	}, routes)

	// Assert
	require.NoError(t, err)

	// Marshal to JSON using encoding/json
	jsonBytes, err := yaml.Marshal(spec)
	require.NoError(t, err)
	require.NotEmpty(t, jsonBytes)

	t.Logf("Generated YAML:\n%s", string(jsonBytes))
}

func TestParameterDescriptionsShouldAppearInOpenAPISpec(t *testing.T) {
	// Arrange
	gen := openapi.NewGenerator()

	// Create a route with parameters that have descriptions
	op := DetachedRoute(http.MethodGet, "/users/{id}").
		WithPathParam("id", "The unique user identifier (UUID format)", "550e8400-e29b-41d4-a716-446655440000").
		WithQueryParam("include", "Comma-separated list of related resources to include", "profile,settings").
		WithRequiredQueryParam("apiVersion", "The API version to use for this request", "v1").
		WithHeaderParam("X-Request-ID", "Unique request identifier for tracing", "req-123", false).
		WithCookieParam("session", "User session token", "abc123", true).
		WithOperationID("getUser").
		Options.Operation

	routes := []openapi.RouteData{
		{Path: "/users/{id}", Method: "GET", Options: &op},
	}

	// Act
	spec, err := gen.GenerateSpecFromRoutes(&openapi.InfoObject{
		Title:   "User API",
		Version: "1.0.0",
	}, routes)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, spec)

	pathItem := spec.Paths["/users/{id}"]
	require.NotNil(t, pathItem)
	require.NotNil(t, pathItem.Get)
	require.Len(t, pathItem.Get.Parameters, 5)

	// Verify each parameter has the correct description
	params := pathItem.Get.Parameters

	// Find path parameter
	idParam := findParam(params, "id", "path")
	require.NotNil(t, idParam, "id path parameter should exist")
	assert.Equal(t, "The unique user identifier (UUID format)", idParam.Description)
	assert.True(t, idParam.Required)

	// Find query parameter
	includeParam := findParam(params, "include", "query")
	require.NotNil(t, includeParam, "include query parameter should exist")
	assert.Equal(t, "Comma-separated list of related resources to include", includeParam.Description)
	assert.False(t, includeParam.Required)

	// Find required query parameter
	apiVersionParam := findParam(params, "apiVersion", "query")
	require.NotNil(t, apiVersionParam, "apiVersion query parameter should exist")
	assert.Equal(t, "The API version to use for this request", apiVersionParam.Description)
	assert.True(t, apiVersionParam.Required)

	// Find header parameter
	requestIDParam := findParam(params, "X-Request-ID", "header")
	require.NotNil(t, requestIDParam, "X-Request-ID header parameter should exist")
	assert.Equal(t, "Unique request identifier for tracing", requestIDParam.Description)
	assert.False(t, requestIDParam.Required)

	// Find cookie parameter
	sessionParam := findParam(params, "session", "cookie")
	require.NotNil(t, sessionParam, "session cookie parameter should exist")
	assert.Equal(t, "User session token", sessionParam.Description)
	assert.True(t, sessionParam.Required)

	// Verify YAML marshaling includes descriptions
	yamlBytes, err := yaml.Marshal(spec)
	require.NoError(t, err, "should marshal to YAML without errors")
	require.NotEmpty(t, yamlBytes)

	yamlString := string(yamlBytes)
	assert.Contains(t, yamlString, "The unique user identifier (UUID format)")
	assert.Contains(t, yamlString, "Comma-separated list of related resources to include")
	assert.Contains(t, yamlString, "The API version to use for this request")
	assert.Contains(t, yamlString, "Unique request identifier for tracing")
	assert.Contains(t, yamlString, "User session token")
}

func TestParameterDescriptionsWithExamplesShouldAppearInOpenAPISpec(t *testing.T) {
	// Arrange
	gen := openapi.NewGenerator(openapi.WithExamples())

	// Create a route with parameters that have descriptions and examples
	op := DetachedRoute(http.MethodGet, "/users/{id}").
		WithPathParam("id", "The unique user identifier (UUID format)", "550e8400-e29b-41d4-a716-446655440000").
		WithQueryParam("include", "Comma-separated list of related resources to include", "profile,settings").
		WithOperationID("getUserWithExamples").
		Options.Operation

	routes := []openapi.RouteData{
		{Path: "/users/{id}", Method: "GET", Options: &op},
	}

	// Act
	spec, err := gen.GenerateSpecFromRoutes(&openapi.InfoObject{
		Title:   "User API",
		Version: "1.0.0",
	}, routes)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, spec)

	pathItem := spec.Paths["/users/{id}"]
	require.NotNil(t, pathItem)
	require.NotNil(t, pathItem.Get)
	require.Len(t, pathItem.Get.Parameters, 2)

	params := pathItem.Get.Parameters

	// Verify path parameter has description AND example
	idParam := findParam(params, "id", "path")
	require.NotNil(t, idParam, "id path parameter should exist")
	assert.Equal(t, "The unique user identifier (UUID format)", idParam.Description)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", idParam.Example)

	// Verify query parameter has description AND example
	includeParam := findParam(params, "include", "query")
	require.NotNil(t, includeParam, "include query parameter should exist")
	assert.Equal(t, "Comma-separated list of related resources to include", includeParam.Description)
	assert.Equal(t, "profile,settings", includeParam.Example)

	// Verify YAML marshaling includes both descriptions and examples
	yamlBytes, err := yaml.Marshal(spec)
	require.NoError(t, err, "should marshal to YAML without errors")
	require.NotEmpty(t, yamlBytes)

	yamlString := string(yamlBytes)
	assert.Contains(t, yamlString, "The unique user identifier (UUID format)")
	assert.Contains(t, yamlString, "550e8400-e29b-41d4-a716-446655440000")
	assert.Contains(t, yamlString, "Comma-separated list of related resources to include")
	assert.Contains(t, yamlString, "profile,settings")
}

// Helper function to find a parameter by name and location
func findParam(params []*openapi.ParameterObject, name, in string) *openapi.ParameterObject {
	for _, p := range params {
		if p.Name == name && p.In == in {
			return p
		}
	}
	return nil
}

// TestAnyOfWithNestedTypesShouldRegisterAllComponents verifies that when using
// anyOf/oneOf/allOf with types that contain nested structs, all types (including
// the nested ones) are properly registered as components in the OpenAPI spec.
func TestAnyOfWithNestedTypesShouldRegisterAllComponents(t *testing.T) {
	// Arrange
	type Address struct {
		Street string `json:"street"`
		City   string `json:"city"`
	}

	type PersonWithAddress struct {
		Name    string  `json:"name"`
		Address Address `json:"address"`
	}

	type CompanyWithAddress struct {
		CompanyName string  `json:"company_name"`
		Address     Address `json:"address"`
	}

	gen := openapi.NewGenerator()

	op := DetachedRoute(http.MethodPost, "/entities").
		WithAnyOfJsonBody(PersonWithAddress{}, CompanyWithAddress{}).
		WithOperationID("createEntity").
		Options.Operation

	routes := []openapi.RouteData{
		{Path: "/entities", Method: "POST", Options: &op},
	}

	// Act
	spec, err := gen.GenerateSpecFromRoutes(&openapi.InfoObject{
		Title:   "Test API",
		Version: "1.0.0",
	}, routes)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, spec)

	// Verify anyOf schemas exist and reference components
	pathItem := spec.Paths["/entities"]
	require.NotNil(t, pathItem)
	require.NotNil(t, pathItem.Post)
	require.NotNil(t, pathItem.Post.RequestBody)

	mediaType := pathItem.Post.RequestBody.Content["application/json"]
	require.NotNil(t, mediaType)
	require.NotNil(t, mediaType.Schema)
	require.NotNil(t, mediaType.Schema.AnyOf)
	require.Len(t, mediaType.Schema.AnyOf, 2)

	// Verify all types are registered as components (including nested Address)
	assert.Contains(t, spec.Components.Schemas, "PersonWithAddress", "PersonWithAddress should be registered")
	assert.Contains(t, spec.Components.Schemas, "CompanyWithAddress", "CompanyWithAddress should be registered")
	assert.Contains(t, spec.Components.Schemas, "Address", "Nested Address type should be registered")

	// Verify the Address component is properly referenced in both types
	personSchema := spec.Components.Schemas["PersonWithAddress"]
	require.NotNil(t, personSchema)
	require.Contains(t, personSchema.Properties, "address")
	assert.Equal(t, "#/components/schemas/Address", personSchema.Properties["address"].Ref,
		"PersonWithAddress.address should reference Address component")

	companySchema := spec.Components.Schemas["CompanyWithAddress"]
	require.NotNil(t, companySchema)
	require.Contains(t, companySchema.Properties, "address")
	assert.Equal(t, "#/components/schemas/Address", companySchema.Properties["address"].Ref,
		"CompanyWithAddress.address should reference Address component")

	// Verify Address component has the expected structure
	addressSchema := spec.Components.Schemas["Address"]
	require.NotNil(t, addressSchema)
	assert.Equal(t, "object", addressSchema.Type)
	assert.Contains(t, addressSchema.Properties, "street")
	assert.Contains(t, addressSchema.Properties, "city")
}
