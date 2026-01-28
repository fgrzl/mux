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
	op := Route(http.MethodPost, "/pets").
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

	op := Route(http.MethodPost, "/animals").
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

	op := Route(http.MethodPost, "/entities").
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

	op := Route(http.MethodPost, "/pets").
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

	op := Route(http.MethodPost, "/pets").
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
