package test

// import (
// 	"reflect"
// 	"testing"

// 	openapi "github.com/fgrzl/mux/pkg/openapi"
// 	routing "github.com/fgrzl/mux/internal/routing"
// 	"github.com/google/uuid"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )

// // TestShouldGenerateOpenAPISpecWithMapSupportWhenUsingAnonymousStructWithMaps demonstrates the exact use case from the user request
// func TestShouldGenerateOpenAPISpecWithMapSupportWhenUsingAnonymousStructWithMaps(t *testing.T) {
// 	// Create a router with OpenAPI info
// 	router := NewRouter(
// 		WithTitle("Test API"),
// 		WithVersion("1.0.0"),
// 	)

// 	// Create a route group
// 	rg := router.NewRouteGroup("/api/v1")

// 	// This is the exact code from the user's request, adapted for testing
// 	rg.PUT("/{secret_id}", func(c routing.RouteContext) {
// 		// Mock implementation - in real usage this would call dispatcher.RequestAccepted
// 		c.OK(map[string]string{"status": "accepted"})
// 	}).
// 		WithOperationID("reconcileSecret").
// 		WithSummary("Reconcile a secret").
// 		WithPathParam("secret_id", uuid.Nil).
// 		WithJsonBody(&struct {
// 			Name        string            `json:"name"`
// 			Description string            `json:"description"`
// 			Values      map[string]string `json:"values"`
// 		}{}).
// 		WithResponse(202, map[string]string{"status": "accepted"}). // Simulating messaging.Accepted
// 		WithBadRequestResponse()

// 	// Generate the OpenAPI spec
// 	gen := openapi.NewGenerator()
// 	spec, err := GenerateSpecWithGenerator(gen, router)
// 	require.NoError(t, err)
// 	require.NotNil(t, spec)

// 	// Verify that the path was registered
// 	pathItem := spec.Paths["/api/v1/{secret_id}"]
// 	require.NotNil(t, pathItem, "Expected path /api/v1/{secret_id} to be registered")
// 	require.NotNil(t, pathItem.Put)

// 	// Verify the operation details
// 	operation := pathItem.Put
// 	assert.Equal(t, "reconcileSecret", operation.OperationID)
// 	assert.Equal(t, "Reconcile a secret", operation.Summary)

// 	// Verify the path parameter
// 	require.Len(t, operation.Parameters, 1)
// 	param := operation.Parameters[0]
// 	assert.Equal(t, "secret_id", param.Name)
// 	assert.Equal(t, "path", param.In)

// 	// Verify the request body
// 	require.NotNil(t, operation.RequestBody)
// 	jsonContent := operation.RequestBody.Content["application/json"]
// 	require.NotNil(t, jsonContent)
// 	require.NotNil(t, jsonContent.Schema)

// 	// The schema should be inline for anonymous struct
// 	schema := jsonContent.Schema
// 	assert.Equal(t, "object", schema.Type)
// 	require.NotNil(t, schema.Properties)

// 	// Verify the "name" field
// 	nameSchema := schema.Properties["name"]
// 	require.NotNil(t, nameSchema)
// 	assert.Equal(t, "string", nameSchema.Type)

// 	// Verify the "description" field
// 	descSchema := schema.Properties["description"]
// 	require.NotNil(t, descSchema)
// 	assert.Equal(t, "string", descSchema.Type)

// 	// Verify the "values" field (map[string]string)
// 	valuesSchema := schema.Properties["values"]
// 	require.NotNil(t, valuesSchema)
// 	assert.Equal(t, "object", valuesSchema.Type)
// 	require.NotNil(t, valuesSchema.AdditionalProperties)
// 	assert.Equal(t, "string", valuesSchema.AdditionalProperties.Type)

// 	// Verify responses
// 	require.Contains(t, operation.Responses, "202")
// 	require.Contains(t, operation.Responses, "400")

// 	t.Logf("Successfully generated OpenAPI spec with map[string]string support!")
// }

// func TestShouldGenerateInlineSchemaWhenUsingJsonBodyWithAnonymousStructContainingMaps(t *testing.T) {
// 	// Test anonymous struct with map[string]string
// 	rb := Route("POST", "/test").WithJsonBody(struct {
// 		Name   string            `json:"name"`
// 		Values map[string]string `json:"values"`
// 	}{})

// 	require.NotNil(t, rb.Options.RequestBody)
// 	jsonContent := rb.Options.RequestBody.Content["application/json"]
// 	require.NotNil(t, jsonContent.Schema)

// 	// Should be inline object
// 	assert.Equal(t, "object", jsonContent.Schema.Type)
// 	assert.Contains(t, jsonContent.Schema.Properties, "name")
// 	assert.Contains(t, jsonContent.Schema.Properties, "values")

// 	// Name should be a string
// 	nameSchema := jsonContent.Schema.Properties["name"]
// 	require.NotNil(t, nameSchema)
// 	assert.Equal(t, "string", nameSchema.Type)

// 	// Values should be an object with string additionalProperties
// 	valuesSchema := jsonContent.Schema.Properties["values"]
// 	require.NotNil(t, valuesSchema)
// 	assert.Equal(t, "object", valuesSchema.Type)
// 	require.NotNil(t, valuesSchema.AdditionalProperties)
// 	assert.Equal(t, "string", valuesSchema.AdditionalProperties.Type)
// }

// func TestShouldHandleComplexMapTypesWhenUsingJsonBodyWithAnonymousStruct(t *testing.T) {
// 	// Test anonymous struct with different map value types
// 	rb := Route("POST", "/test").WithJsonBody(struct {
// 		StringMap  map[string]string   `json:"string_map"`
// 		IntMap     map[string]int      `json:"int_map"`
// 		BoolMap    map[string]bool     `json:"bool_map"`
// 		FloatMap   map[string]float64  `json:"float_map"`
// 		StringsMap map[string][]string `json:"strings_map"`
// 	}{})

// 	require.NotNil(t, rb.Options.RequestBody)
// 	jsonContent := rb.Options.RequestBody.Content["application/json"]
// 	require.NotNil(t, jsonContent.Schema)

// 	// Should be inline object
// 	assert.Equal(t, "object", jsonContent.Schema.Type)

// 	// Test string map
// 	stringMapSchema := jsonContent.Schema.Properties["string_map"]
// 	require.NotNil(t, stringMapSchema)
// 	assert.Equal(t, "object", stringMapSchema.Type)
// 	require.NotNil(t, stringMapSchema.AdditionalProperties)
// 	assert.Equal(t, "string", stringMapSchema.AdditionalProperties.Type)

// 	// Test int map
// 	intMapSchema := jsonContent.Schema.Properties["int_map"]
// 	require.NotNil(t, intMapSchema)
// 	assert.Equal(t, "object", intMapSchema.Type)
// 	require.NotNil(t, intMapSchema.AdditionalProperties)
// 	assert.Equal(t, "integer", intMapSchema.AdditionalProperties.Type)

// 	// Test bool map
// 	boolMapSchema := jsonContent.Schema.Properties["bool_map"]
// 	require.NotNil(t, boolMapSchema)
// 	assert.Equal(t, "object", boolMapSchema.Type)
// 	require.NotNil(t, boolMapSchema.AdditionalProperties)
// 	assert.Equal(t, "boolean", boolMapSchema.AdditionalProperties.Type)

// 	// Test float map
// 	floatMapSchema := jsonContent.Schema.Properties["float_map"]
// 	require.NotNil(t, floatMapSchema)
// 	assert.Equal(t, "object", floatMapSchema.Type)
// 	require.NotNil(t, floatMapSchema.AdditionalProperties)
// 	assert.Equal(t, "number", floatMapSchema.AdditionalProperties.Type)

// 	// Test string slice map
// 	stringsMapSchema := jsonContent.Schema.Properties["strings_map"]
// 	require.NotNil(t, stringsMapSchema)
// 	assert.Equal(t, "object", stringsMapSchema.Type)
// 	require.NotNil(t, stringsMapSchema.AdditionalProperties)
// 	assert.Equal(t, "array", stringsMapSchema.AdditionalProperties.Type)
// 	require.NotNil(t, stringsMapSchema.AdditionalProperties.Items)
// 	assert.Equal(t, "string", stringsMapSchema.AdditionalProperties.Items.Type)
// }

// func TestShouldSupportNumericMapKeysWhenUsingJsonBodyWithAnonymousStruct(t *testing.T) {
// 	// Test anonymous struct with numeric key maps
// 	rb := Route("POST", "/test").WithJsonBody(struct {
// 		IntMap   map[int]string   `json:"int_map"`
// 		Int64Map map[int64]bool   `json:"int64_map"`
// 		UintMap  map[uint]float64 `json:"uint_map"`
// 	}{})

// 	require.NotNil(t, rb.Options.RequestBody)
// 	jsonContent := rb.Options.RequestBody.Content["application/json"]
// 	require.NotNil(t, jsonContent.Schema)

// 	// Should be inline object
// 	assert.Equal(t, "object", jsonContent.Schema.Type)

// 	// Test int map
// 	intMapSchema := jsonContent.Schema.Properties["int_map"]
// 	require.NotNil(t, intMapSchema)
// 	assert.Equal(t, "object", intMapSchema.Type)
// 	require.NotNil(t, intMapSchema.AdditionalProperties)
// 	assert.Equal(t, "string", intMapSchema.AdditionalProperties.Type)

// 	// Test int64 map
// 	int64MapSchema := jsonContent.Schema.Properties["int64_map"]
// 	require.NotNil(t, int64MapSchema)
// 	assert.Equal(t, "object", int64MapSchema.Type)
// 	require.NotNil(t, int64MapSchema.AdditionalProperties)
// 	assert.Equal(t, "boolean", int64MapSchema.AdditionalProperties.Type)

// 	// Test uint map
// 	uintMapSchema := jsonContent.Schema.Properties["uint_map"]
// 	require.NotNil(t, uintMapSchema)
// 	assert.Equal(t, "object", uintMapSchema.Type)
// 	require.NotNil(t, uintMapSchema.AdditionalProperties)
// 	assert.Equal(t, "number", uintMapSchema.AdditionalProperties.Type)
// }

// func TestShouldGenerateCorrectSchemaWhenQuickSchemaProcessesMapTypes(t *testing.T) {
// 	// Test QuickSchema function directly with map types
// 	tests := []struct {
// 		name     string
// 		mapType  reflect.Type
// 		expected *openapi.Schema
// 	}{
// 		{
// 			name:    "map[string]string",
// 			mapType: reflect.TypeOf(map[string]string{}),
// 			expected: &openapi.Schema{
// 				Type:                 "object",
// 				AdditionalProperties: &openapi.Schema{Type: "string"},
// 			},
// 		},
// 		{
// 			name:    "map[string]int",
// 			mapType: reflect.TypeOf(map[string]int{}),
// 			expected: &openapi.Schema{
// 				Type:                 "object",
// 				AdditionalProperties: &openapi.Schema{Type: "integer"},
// 			},
// 		},
// 		{
// 			name:    "map[string][]string",
// 			mapType: reflect.TypeOf(map[string][]string{}),
// 			expected: &openapi.Schema{
// 				Type: "object",
// 				AdditionalProperties: &openapi.Schema{
// 					Type:  "array",
// 					Items: &openapi.Schema{Type: "string"},
// 				},
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			schema, err := QuickSchema(tt.mapType)
// 			require.NoError(t, err)
// 			require.NotNil(t, schema)

// 			assert.Equal(t, tt.expected.Type, schema.Type)
// 			require.NotNil(t, schema.AdditionalProperties)
// 			assert.Equal(t, tt.expected.AdditionalProperties.Type, schema.AdditionalProperties.Type)

// 			if tt.expected.AdditionalProperties.Items != nil {
// 				require.NotNil(t, schema.AdditionalProperties.Items)
// 				assert.Equal(t, tt.expected.AdditionalProperties.Items.Type, schema.AdditionalProperties.Items.Type)
// 			}
// 		})
// 	}
// }

// func TestShouldReturnErrorWhenQuickSchemaProcessesUnsupportedMapKeyType(t *testing.T) {
// 	// Test that unsupported map key types are rejected (e.g., complex types)
// 	complexKeyMapType := reflect.TypeOf(map[struct{ Name string }]string{})

// 	schema, err := QuickSchema(complexKeyMapType)
// 	assert.Error(t, err)
// 	assert.Nil(t, schema)
// 	assert.Contains(t, err.Error(), "unsupported map key type")
// 	assert.Contains(t, err.Error(), "only string and numeric keys are supported")
// }

// func TestShouldSupportNumericMapKeyTypesWhenUsingQuickSchema(t *testing.T) {
// 	// Test that numeric key types are now supported
// 	tests := []struct {
// 		name     string
// 		mapType  reflect.Type
// 		expected *openapi.Schema
// 	}{
// 		{
// 			name:    "map[int]string",
// 			mapType: reflect.TypeOf(map[int]string{}),
// 			expected: &openapi.Schema{
// 				Type:                 "object",
// 				AdditionalProperties: &openapi.Schema{Type: "string"},
// 			},
// 		},
// 		{
// 			name:    "map[int64]bool",
// 			mapType: reflect.TypeOf(map[int64]bool{}),
// 			expected: &openapi.Schema{
// 				Type:                 "object",
// 				AdditionalProperties: &openapi.Schema{Type: "boolean"},
// 			},
// 		},
// 		{
// 			name:    "map[uint32]int",
// 			mapType: reflect.TypeOf(map[uint32]int{}),
// 			expected: &openapi.Schema{
// 				Type:                 "object",
// 				AdditionalProperties: &openapi.Schema{Type: "integer"},
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			schema, err := QuickSchema(tt.mapType)
// 			require.NoError(t, err)
// 			require.NotNil(t, schema)

// 			assert.Equal(t, tt.expected.Type, schema.Type)
// 			require.NotNil(t, schema.AdditionalProperties)
// 			assert.Equal(t, tt.expected.AdditionalProperties.Type, schema.AdditionalProperties.Type)
// 		})
// 	}
// }
