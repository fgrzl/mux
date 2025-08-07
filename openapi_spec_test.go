package mux

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
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
				Info:    &InfoObject{Title: "Test API", Version: "1.0.0"},
				Paths:   map[string]*PathItem{"/test": {}},
			},
			expectError: false,
		},
		{
			name: "invalid OpenAPI version",
			spec: &OpenAPISpec{
				OpenAPI: "3.0.0",
				Info:    &InfoObject{Title: "Test API", Version: "1.0.0"},
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
				Info:    &InfoObject{Title: "Test API"},
				Paths:   map[string]*PathItem{"/test": {}},
			},
			expectError: true,
			errorMsg:    "info.version is required",
		},
		{
			name: "missing paths",
			spec: &OpenAPISpec{
				OpenAPI: "3.1.0",
				Info:    &InfoObject{Title: "Test API", Version: "1.0.0"},
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
		Info:    &InfoObject{Title: "Test API", Version: "1.0.0"},
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
		Info:    &InfoObject{Title: "Test API", Version: "1.0.0"},
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

	tempFile := filepath.Join(t.TempDir(), "test-unmarshal.yml")
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
