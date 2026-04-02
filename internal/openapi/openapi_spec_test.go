package openapi

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

const (
	testAPITitle   = "Test API"
	testAPIVersion = "1.0.0"
)

func TestShouldCreateNewOpenAPISpecWithDefaults(t *testing.T) {
	// Arrange & Act
	spec := NewOpenAPISpec()

	// Assert
	assert.Equal(t, "3.1.0", spec.OpenAPI)
	assert.Equal(t, "https://json-schema.org/draft/2020-12/schema", spec.JsonSchemaDialect)
	assert.NotNil(t, spec.Info)
	assert.Equal(t, "1.0.0", spec.Info.Version)
	assert.NotNil(t, spec.Paths)
	assert.Len(t, spec.Servers, 1)
	assert.Equal(t, "/", spec.Servers[0].URL)
}

func TestShouldValidateRequiredFields(t *testing.T) {
	// Arrange
	tests := []struct {
		name        string
		spec        *OpenAPISpec
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid spec",
			spec: &OpenAPISpec{
				OpenAPI: "3.1.0",
				Info:    &InfoObject{Title: testAPITitle, Version: testAPIVersion},
				Paths:   map[string]*PathItem{"/test": {}},
			},
			expectError: false,
		},
		{
			name: "invalid OpenAPI version",
			spec: &OpenAPISpec{
				OpenAPI: "3.0.0",
				Info:    &InfoObject{Title: testAPITitle, Version: testAPIVersion},
				Paths:   map[string]*PathItem{"/test": {}},
			},
			expectError: true,
			errorMsg:    "openapi must be '3.1.0'",
		},
		{
			name: "missing title",
			spec: &OpenAPISpec{
				OpenAPI: "3.1.0",
				Info:    &InfoObject{Version: "1.0.0"},
				Paths:   map[string]*PathItem{"/test": {}},
			},
			expectError: true,
			errorMsg:    "info.title is required",
		},
		{
			name: "missing version",
			spec: &OpenAPISpec{
				OpenAPI: "3.1.0",
				Info:    &InfoObject{Title: testAPITitle},
				Paths:   map[string]*PathItem{"/test": {}},
			},
			expectError: true,
			errorMsg:    "info.version is required",
		},
		{
			name: "missing paths",
			spec: &OpenAPISpec{
				OpenAPI: "3.1.0",
				Info:    &InfoObject{Title: testAPITitle, Version: testAPIVersion},
				Paths:   map[string]*PathItem{},
			},
			expectError: true,
			errorMsg:    "at least one path is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := tt.spec.Validate()

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestShouldNormalizeEmptyComponents(t *testing.T) {
	// Arrange
	spec := &OpenAPISpec{
		Components: &ComponentsObject{
			Responses:       map[string]*ResponseObject{},
			Parameters:      map[string]*ParameterObject{},
			Examples:        map[string]*ExampleObject{},
			RequestBodies:   map[string]*RequestBodyObject{},
			Headers:         map[string]*HeaderObject{},
			SecuritySchemes: map[string]*SecurityScheme{},
			Links:           map[string]*LinkObject{},
		},
	}

	// Act
	normalized := spec.Normalize()

	// Assert
	assert.Nil(t, normalized.Components.Responses)
	assert.Nil(t, normalized.Components.Parameters)
	assert.Nil(t, normalized.Components.Examples)
	assert.Nil(t, normalized.Components.RequestBodies)
	assert.Nil(t, normalized.Components.Headers)
	assert.Nil(t, normalized.Components.SecuritySchemes)
	assert.Nil(t, normalized.Components.Links)
}

func TestShouldNotNormalizeNonEmptyComponents(t *testing.T) {
	// Arrange
	spec := &OpenAPISpec{
		Components: &ComponentsObject{
			Responses: map[string]*ResponseObject{
				"200": {Description: "OK"},
			},
			Parameters: map[string]*ParameterObject{
				"limit": {Name: "limit", In: "query"},
			},
		},
	}

	// Act
	normalized := spec.Normalize()

	// Assert
	assert.NotNil(t, normalized.Components.Responses)
	assert.NotNil(t, normalized.Components.Parameters)
	assert.Len(t, normalized.Components.Responses, 1)
	assert.Len(t, normalized.Components.Parameters, 1)
}

func TestShouldMarshalToJSONFile(t *testing.T) {
	// Arrange
	spec := &OpenAPISpec{
		OpenAPI: "3.1.0",
		Info:    &InfoObject{Title: testAPITitle, Version: testAPIVersion},
		Paths: map[string]*PathItem{
			"/test": {
				Get: &Operation{
					Summary: "Test endpoint",
					Responses: map[string]*ResponseObject{
						"200": {Description: "Success"},
					},
				},
			},
		},
	}

	tempFile := filepath.Join(t.TempDir(), "test-spec.json")

	// Act
	err := spec.MarshalToFile(tempFile)

	// Assert
	require.NoError(t, err)

	// Verify file exists and is valid JSON
	data, err := os.ReadFile(tempFile)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, "3.1.0", result["openapi"])
	assert.Contains(t, result, "info")
	assert.Contains(t, result, "paths")
}

func TestShouldMarshalToYAMLFile(t *testing.T) {
	// Arrange
	spec := &OpenAPISpec{
		OpenAPI: "3.1.0",
		Info:    &InfoObject{Title: testAPITitle, Version: testAPIVersion},
		Paths: map[string]*PathItem{
			"/test": {
				Post: &Operation{
					Summary: "Create test",
					Responses: map[string]*ResponseObject{
						"201": {Description: "Created"},
					},
				},
			},
		},
	}

	tempFile := filepath.Join(t.TempDir(), "test-spec.yaml")

	// Act
	err := spec.MarshalToFile(tempFile)

	// Assert
	require.NoError(t, err)

	// Verify file exists and is valid YAML
	data, err := os.ReadFile(tempFile)
	require.NoError(t, err)

	var result map[string]interface{}
	err = yaml.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, "3.1.0", result["openapi"])
	assert.Contains(t, result, "info")
	assert.Contains(t, result, "paths")
}

func TestShouldReturnErrorForUnsupportedFileExtension(t *testing.T) {
	// Arrange
	spec := NewOpenAPISpec()
	tempFile := filepath.Join(t.TempDir(), "test-spec.txt")

	// Act
	err := spec.MarshalToFile(tempFile)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported file extension")
}

func TestShouldUnmarshalFromJSONFile(t *testing.T) {
	// Arrange
	jsonContent := `{
		"openapi": "3.1.0",
		"info": {
			"title": "Unmarshal Test API",
			"version": "2.0.0"
		},
		"paths": {
			"/unmarshal-test": {
				"get" : {
					"summary": "Test unmarshal",
					"responses": {
						"200": {
							"description": "Success"
						}
					}
				}
			}
		}
	}`

	tempFile := filepath.Join(t.TempDir(), "test-unmarshal.json")
	require.NoError(t, os.WriteFile(tempFile, []byte(jsonContent), 0644))

	var spec OpenAPISpec

	// Act
	err := spec.UnmarshalFromFile(tempFile)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "3.1.0", spec.OpenAPI)
	assert.Equal(t, "Unmarshal Test API", spec.Info.Title)
	assert.Equal(t, "2.0.0", spec.Info.Version)
	assert.Contains(t, spec.Paths, "/unmarshal-test")
}

func TestShouldUnmarshalFromYAMLFile(t *testing.T) {
	// Arrange
	yamlContent := `
openapi: 3.1.0
info:
  title: YAML Test API
  version: 3.0.0
paths:
  /yaml-test:
    put:
      summary: Test YAML unmarshal
      responses:
        '200':
          description: Success
`

	tempFile := filepath.Join(t.TempDir(), "test-unmarshal.yaml")
	require.NoError(t, os.WriteFile(tempFile, []byte(yamlContent), 0644))

	var spec OpenAPISpec

	// Act
	err := spec.UnmarshalFromFile(tempFile)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "3.1.0", spec.OpenAPI)
	assert.Equal(t, "YAML Test API", spec.Info.Title)
	assert.Equal(t, "3.0.0", spec.Info.Version)
	assert.Contains(t, spec.Paths, "/yaml-test")
}

func TestShouldReturnErrorForNonExistentFile(t *testing.T) {
	// Arrange
	var spec OpenAPISpec
	nonExistentFile := "/path/that/does/not/exist.json"

	// Act
	err := spec.UnmarshalFromFile(nonExistentFile)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reading file")
}

func TestShouldReturnErrorForUnsupportedUnmarshalExtension(t *testing.T) {
	// Arrange
	tempFile := filepath.Join(t.TempDir(), "test.txt")
	require.NoError(t, os.WriteFile(tempFile, []byte("some content"), 0644))

	var spec OpenAPISpec

	// Act
	err := spec.UnmarshalFromFile(tempFile)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported file extension")
}

func TestShouldHandleComplexOpenAPISpec(t *testing.T) {
	// Arrange
	spec := &OpenAPISpec{
		OpenAPI:           "3.1.0",
		JsonSchemaDialect: "https://json-schema.org/draft/2020-12/schema",
		Info: &InfoObject{
			Title:       "Complex API",
			Version:     "1.0.0",
			Description: "A complex API with all features",
			Contact: &ContactObject{
				Name:  "API Support",
				URL:   "https://example.com/support",
				Email: "support@example.com",
			},
			License: &LicenseObject{
				Name: "MIT",
				URL:  "https://opensource.org/licenses/MIT",
			},
		},
		Servers: []*ServerObject{
			{URL: "https://api.example.com"},
			{URL: "https://staging.api.example.com"},
		},
		Paths: map[string]*PathItem{
			"/users/{id}": {
				Parameters: []*ParameterObject{
					{Name: "id", In: "path", Required: true, Schema: &Schema{Type: "integer"}},
				},
				Get: &Operation{
					OperationID: "getUser",
					Summary:     "Get user by ID",
					Tags:        []string{"users"},
					Parameters: []*ParameterObject{
						{Name: "include", In: "query", Schema: &Schema{Type: "string"}},
					},
					Responses: map[string]*ResponseObject{
						"200": {Description: "User found"},
						"404": {Description: "User not found"},
					},
				},
			},
		},
		Components: &ComponentsObject{
			SecuritySchemes: map[string]*SecurityScheme{
				"bearerAuth": {
					Type:   "http",
					Scheme: "bearer",
				},
			},
		},
		Security: []*SecurityRequirement{
			{"bearerAuth": []string{}},
		},
	}

	// Act & Assert - Should not panic and should be valid
	err := spec.Validate()
	assert.NoError(t, err)

	normalized := spec.Normalize()
	assert.NotNil(t, normalized)

	// Test JSON marshalling
	tempFile := filepath.Join(t.TempDir(), "complex-spec.json")
	err = spec.MarshalToFile(tempFile)
	assert.NoError(t, err)

	// Test YAML marshalling
	tempFileYAML := filepath.Join(t.TempDir(), "complex-spec.yaml")
	err = spec.MarshalToFile(tempFileYAML)
	assert.NoError(t, err)
}

func TestShouldSupportPropertyLevelDescriptions(t *testing.T) {
	// Arrange - Create a schema with property-level descriptions
	userSchema := &Schema{
		Type:        "object",
		Description: "A user object representing a system user",
		Properties: map[string]*Schema{
			"id": {
				Type:        "integer",
				Description: "The unique identifier for the user",
			},
			"name": {
				Type:        "string",
				Description: "The full name of the user",
			},
			"email": {
				Type:        "string",
				Format:      "email",
				Description: "The email address of the user",
			},
			"age": {
				Type:        "integer",
				Description: "The age of the user in years",
				Minimum:     ptrFloat64(0),
				Maximum:     ptrFloat64(150),
			},
		},
		Required: []string{"id", "name", "email"},
	}

	spec := &OpenAPISpec{
		OpenAPI: "3.1.0",
		Info:    &InfoObject{Title: "User API", Version: "1.0.0"},
		Paths: map[string]*PathItem{
			"/users": {
				Post: &Operation{
					OperationID: "createUser",
					RequestBody: &RequestBodyObject{
						Required: true,
						Content: map[string]*MediaType{
							"application/json": {
								Schema: userSchema,
							},
						},
					},
					Responses: map[string]*ResponseObject{
						"201": {Description: "Created"},
					},
				},
			},
		},
	}

	// Act - Marshal to JSON and YAML to verify descriptions are preserved
	tempJSON := filepath.Join(t.TempDir(), "schema-with-descriptions.json")
	err := spec.MarshalToFile(tempJSON)
	require.NoError(t, err)

	tempYAML := filepath.Join(t.TempDir(), "schema-with-descriptions.yaml")
	err = spec.MarshalToFile(tempYAML)
	require.NoError(t, err)

	// Assert - Read back and verify descriptions are present
	jsonData, err := os.ReadFile(tempJSON)
	require.NoError(t, err)

	var jsonSpec map[string]any
	err = json.Unmarshal(jsonData, &jsonSpec)
	require.NoError(t, err)

	// Navigate to the schema properties and verify descriptions
	paths := jsonSpec["paths"].(map[string]any)
	usersPath := paths["/users"].(map[string]any)
	post := usersPath["post"].(map[string]any)
	requestBody := post["requestBody"].(map[string]any)
	content := requestBody["content"].(map[string]any)
	jsonContent := content["application/json"].(map[string]any)
	schema := jsonContent["schema"].(map[string]any)
	properties := schema["properties"].(map[string]any)

	// Verify top-level schema description
	assert.Equal(t, "A user object representing a system user", schema["description"])

	// Verify property-level descriptions
	idProp := properties["id"].(map[string]any)
	assert.Equal(t, "The unique identifier for the user", idProp["description"])

	nameProp := properties["name"].(map[string]any)
	assert.Equal(t, "The full name of the user", nameProp["description"])

	emailProp := properties["email"].(map[string]any)
	assert.Equal(t, "The email address of the user", emailProp["description"])

	ageProp := properties["age"].(map[string]any)
	assert.Equal(t, "The age of the user in years", ageProp["description"])

	// Verify YAML as well
	yamlData, err := os.ReadFile(tempYAML)
	require.NoError(t, err)

	var yamlSpec map[string]any
	err = yaml.Unmarshal(yamlData, &yamlSpec)
	require.NoError(t, err)

	// Navigate to the schema in YAML
	yamlPaths := yamlSpec["paths"].(map[string]any)
	yamlUsersPath := yamlPaths["/users"].(map[string]any)
	yamlPost := yamlUsersPath["post"].(map[string]any)
	yamlRequestBody := yamlPost["requestBody"].(map[string]any)
	yamlContent := yamlRequestBody["content"].(map[string]any)
	yamlJsonContent := yamlContent["application/json"].(map[string]any)
	yamlSchema := yamlJsonContent["schema"].(map[string]any)
	yamlProperties := yamlSchema["properties"].(map[string]any)

	// Verify descriptions in YAML
	yamlIdProp := yamlProperties["id"].(map[string]any)
	assert.Equal(t, "The unique identifier for the user", yamlIdProp["description"])
}

func ptrFloat64(v float64) *float64 {
	return &v
}
