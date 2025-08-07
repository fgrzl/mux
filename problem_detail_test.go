package mux

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldCreateProblemDetailsWithAllFields(t *testing.T) {
	// Arrange
	instance := "/users/123"
	problem := ProblemDetails{
		Type:     "https://example.com/problems/invalid-user",
		Title:    "Invalid User",
		Status:   400,
		Detail:   "The user ID provided is invalid",
		Instance: &instance,
	}

	// Act & Assert
	assert.Equal(t, "https://example.com/problems/invalid-user", problem.Type)
	assert.Equal(t, "Invalid User", problem.Title)
	assert.Equal(t, 400, problem.Status)
	assert.Equal(t, "The user ID provided is invalid", problem.Detail)
	assert.NotNil(t, problem.Instance)
	assert.Equal(t, "/users/123", *problem.Instance)
}

func TestShouldCreateProblemDetailsWithoutInstance(t *testing.T) {
	// Arrange
	problem := ProblemDetails{
		Type:   "https://example.com/problems/generic-error",
		Title:  "Generic Error",
		Status: 500,
		Detail: "An unexpected error occurred",
	}

	// Act & Assert
	assert.Equal(t, "https://example.com/problems/generic-error", problem.Type)
	assert.Equal(t, "Generic Error", problem.Title)
	assert.Equal(t, 500, problem.Status)
	assert.Equal(t, "An unexpected error occurred", problem.Detail)
	assert.Nil(t, problem.Instance)
}

func TestShouldSerializeProblemDetailsToJSON(t *testing.T) {
	// Arrange
	instance := "/api/users/123"
	problem := ProblemDetails{
		Type:     "https://example.com/problems/not-found",
		Title:    "User Not Found",
		Status:   404,
		Detail:   "The requested user could not be found",
		Instance: &instance,
	}

	// Act
	jsonData, err := json.Marshal(problem)

	// Assert
	require.NoError(t, err)
	
	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	require.NoError(t, err)
	
	assert.Equal(t, "https://example.com/problems/not-found", result["type"])
	assert.Equal(t, "User Not Found", result["title"])
	assert.Equal(t, float64(404), result["status"]) // JSON numbers are float64
	assert.Equal(t, "The requested user could not be found", result["detail"])
	assert.Equal(t, "/api/users/123", result["instance"])
}

func TestShouldSerializeProblemDetailsWithoutInstanceToJSON(t *testing.T) {
	// Arrange
	problem := ProblemDetails{
		Type:   "about:blank",
		Title:  "Bad Request",
		Status: 400,
		Detail: "The request was malformed",
	}

	// Act
	jsonData, err := json.Marshal(problem)

	// Assert
	require.NoError(t, err)
	
	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	require.NoError(t, err)
	
	assert.Equal(t, "about:blank", result["type"])
	assert.Equal(t, "Bad Request", result["title"])
	assert.Equal(t, float64(400), result["status"])
	assert.Equal(t, "The request was malformed", result["detail"])
	assert.NotContains(t, result, "instance") // Should be omitted when nil
}

func TestShouldDeserializeProblemDetailsFromJSON(t *testing.T) {
	// Arrange
	jsonData := `{
		"type": "https://example.com/problems/validation-error",
		"title": "Validation Error",
		"status": 422,
		"detail": "One or more fields failed validation",
		"instance": "/api/users"
	}`

	// Act
	var problem ProblemDetails
	err := json.Unmarshal([]byte(jsonData), &problem)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/problems/validation-error", problem.Type)
	assert.Equal(t, "Validation Error", problem.Title)
	assert.Equal(t, 422, problem.Status)
	assert.Equal(t, "One or more fields failed validation", problem.Detail)
	require.NotNil(t, problem.Instance)
	assert.Equal(t, "/api/users", *problem.Instance)
}

func TestShouldDeserializeProblemDetailsFromJSONWithoutInstance(t *testing.T) {
	// Arrange
	jsonData := `{
		"type": "about:blank",
		"title": "Internal Server Error",
		"status": 500,
		"detail": "An unexpected error occurred on the server"
	}`

	// Act
	var problem ProblemDetails
	err := json.Unmarshal([]byte(jsonData), &problem)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "about:blank", problem.Type)
	assert.Equal(t, "Internal Server Error", problem.Title)
	assert.Equal(t, 500, problem.Status)
	assert.Equal(t, "An unexpected error occurred on the server", problem.Detail)
	assert.Nil(t, problem.Instance)
}

func TestShouldCreateMinimalProblemDetails(t *testing.T) {
	// Arrange
	problem := ProblemDetails{
		Status: 500,
	}

	// Act & Assert
	assert.Empty(t, problem.Type)
	assert.Empty(t, problem.Title)
	assert.Equal(t, 500, problem.Status)
	assert.Empty(t, problem.Detail)
	assert.Nil(t, problem.Instance)
}

func TestShouldHandleEmptyInstance(t *testing.T) {
	// Arrange
	emptyInstance := ""
	problem := ProblemDetails{
		Type:     "https://example.com/problems/test",
		Title:    "Test Error",
		Status:   400,
		Detail:   "Test detail",
		Instance: &emptyInstance,
	}

	// Act
	jsonData, err := json.Marshal(problem)

	// Assert
	require.NoError(t, err)
	
	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	require.NoError(t, err)
	
	assert.Equal(t, "", result["instance"]) // Empty string should be preserved
}

func TestDefaultProblemDetailsShouldBeInitialized(t *testing.T) {
	// Arrange & Act
	// defaultProblem is initialized at package level

	// Assert
	assert.NotNil(t, defaultProblem)
	assert.IsType(t, &ProblemDetails{}, defaultProblem)
}