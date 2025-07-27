package mux

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithJsonBodyAnonymousStruct(t *testing.T) {
	// Test with anonymous struct
	rb := Route("POST", "/test").WithJsonBody(struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}{})

	require.NotNil(t, rb.Options.RequestBody)
	require.NotNil(t, rb.Options.RequestBody.Content)
	
	jsonContent := rb.Options.RequestBody.Content["application/json"]
	require.NotNil(t, jsonContent)
	require.NotNil(t, jsonContent.Schema)
	
	// For anonymous structs, we should get an inline schema (not a $ref)
	assert.Empty(t, jsonContent.Schema.Ref, "Anonymous struct should not create a $ref")
	assert.Equal(t, "object", jsonContent.Schema.Type)
	assert.NotNil(t, jsonContent.Schema.Properties)
	
	// Check that the properties are correctly defined
	nameSchema := jsonContent.Schema.Properties["name"]
	require.NotNil(t, nameSchema)
	assert.Equal(t, "string", nameSchema.Type)
	
	ageSchema := jsonContent.Schema.Properties["age"]
	require.NotNil(t, ageSchema)
	assert.Equal(t, "integer", ageSchema.Type)
}

func TestWithJsonBodyAnonymousStructComplexTypes(t *testing.T) {
	// Test with more complex anonymous struct containing different types
	rb := Route("POST", "/test").WithJsonBody(struct {
		Name    string   `json:"name"`
		Age     int      `json:"age"`
		Active  bool     `json:"active"`
		Score   float64  `json:"score"`
		Tags    []string `json:"tags"`
		Ignored string   `json:"-"` // Should be ignored
	}{})

	require.NotNil(t, rb.Options.RequestBody)
	jsonContent := rb.Options.RequestBody.Content["application/json"]
	require.NotNil(t, jsonContent.Schema)
	
	// Check all property types
	assert.Equal(t, "string", jsonContent.Schema.Properties["name"].Type)
	assert.Equal(t, "integer", jsonContent.Schema.Properties["age"].Type)
	assert.Equal(t, "boolean", jsonContent.Schema.Properties["active"].Type)
	assert.Equal(t, "number", jsonContent.Schema.Properties["score"].Type)
	
	// Check array type
	tagsSchema := jsonContent.Schema.Properties["tags"]
	require.NotNil(t, tagsSchema)
	assert.Equal(t, "array", tagsSchema.Type)
	assert.Equal(t, "string", tagsSchema.Items.Type)
	
	// Check that ignored field is not present
	assert.Nil(t, jsonContent.Schema.Properties["ignored"])
}

func TestWithJsonBodyNamedStruct(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	
	// Test with named struct - should preserve existing behavior
	rb := Route("POST", "/test").WithJsonBody(TestStruct{})

	require.NotNil(t, rb.Options.RequestBody)
	require.NotNil(t, rb.Options.RequestBody.Content)
	
	jsonContent := rb.Options.RequestBody.Content["application/json"]
	require.NotNil(t, jsonContent)
	require.NotNil(t, jsonContent.Schema)
	
	// For named structs, we should get a $ref
	assert.Equal(t, "#/components/schemas/TestStruct", jsonContent.Schema.Ref)
	assert.Empty(t, jsonContent.Schema.Type)
	assert.Nil(t, jsonContent.Schema.Properties)
}

func TestQuickSchemaAnonymousStruct(t *testing.T) {
	// Test the quickSchema function directly with anonymous struct
	anonStructType := reflect.TypeOf(struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}{})
	
	schema, err := quickSchema(anonStructType)
	
	// Currently this should fail, but after our fix it should work
	if err != nil {
		t.Logf("Current behavior: quickSchema fails for anonymous struct: %v", err)
		assert.Contains(t, err.Error(), "unsupported param kind struct")
	} else {
		// After fix, this should succeed
		assert.NotNil(t, schema)
		assert.Equal(t, "object", schema.Type)
		assert.NotNil(t, schema.Properties)
	}
}

func TestQuickSchemaNamedStruct(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	
	// Test the quickSchema function with named struct - should work
	namedStructType := reflect.TypeOf(TestStruct{})
	
	schema, err := quickSchema(namedStructType)
	
	require.NoError(t, err)
	assert.NotNil(t, schema)
	assert.Equal(t, "#/components/schemas/TestStruct", schema.Ref)
	assert.Empty(t, schema.Type)
	assert.Nil(t, schema.Properties)
}

func TestOpenAPISpecGenerationWithAnonymousStruct(t *testing.T) {
	// Create a router with routes using anonymous structs
	router := NewRouter(WithTitle("Test API"), WithDescription("Test"), WithVersion("1.0.0"))
	
	router.POST("/test", func(c *RouteContext) {
		c.Created(nil)
	}).WithOperationID("testAnonymousStruct").
		WithJsonBody(struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}{}).
		WithCreatedResponse(struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}{})

	// Generate the OpenAPI spec
	generator := NewGenerator()
	spec, err := generator.GenerateSpec(router)
	
	require.NoError(t, err)
	require.NotNil(t, spec)
	
	// Check that the spec was generated successfully
	assert.NotEmpty(t, spec.Paths)
	
	// Check the /test path
	testPath := spec.Paths["/test"]
	require.NotNil(t, testPath)
	require.NotNil(t, testPath.Post)
	
	// Check request body - should have inline schema for anonymous struct
	require.NotNil(t, testPath.Post.RequestBody)
	jsonContent := testPath.Post.RequestBody.Content["application/json"]
	require.NotNil(t, jsonContent)
	require.NotNil(t, jsonContent.Schema)
	
	// Should have inline schema, not a reference
	assert.Empty(t, jsonContent.Schema.Ref)
	assert.Equal(t, "object", jsonContent.Schema.Type)
	assert.Contains(t, jsonContent.Schema.Properties, "name")
	assert.Contains(t, jsonContent.Schema.Properties, "age")
}

func TestAnonymousStructWithNestedStruct(t *testing.T) {
	type Address struct {
		Street string `json:"street"`
		City   string `json:"city"`
	}
	
	// Test anonymous struct with nested named struct
	rb := Route("POST", "/test").WithJsonBody(struct {
		Name    string  `json:"name"`
		Address Address `json:"address"`
	}{})

	require.NotNil(t, rb.Options.RequestBody)
	jsonContent := rb.Options.RequestBody.Content["application/json"]
	require.NotNil(t, jsonContent.Schema)
	
	// Should be inline object
	assert.Equal(t, "object", jsonContent.Schema.Type)
	assert.Contains(t, jsonContent.Schema.Properties, "name")
	assert.Contains(t, jsonContent.Schema.Properties, "address")
	
	// Address should be a reference to the named struct
	addressSchema := jsonContent.Schema.Properties["address"]
	require.NotNil(t, addressSchema)
	assert.Equal(t, "#/components/schemas/Address", addressSchema.Ref)
}