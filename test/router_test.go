package test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/fgrzl/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func mockServerHandler() *mux.Router {
	r := mux.NewRouter()

	// Add middleware
	// r.UseLogging(&mux.LoggingOptions{})
	// r.UseCompression(&mux.CompressionOptions{})
	// r.UseAuthentication(&mux.AuthenticationOptions{})
	// r.UseAuthorization(&mux.AuthorizationOptions{})

	// break up your routes

	ConfigureRoutes(r)

	return r
}

// Test GET /api/v1/resources/
func TestGetResourcesSuccess(t *testing.T) {
	// Arrange: Start the mock server
	server := httptest.NewServer(mockServerHandler())
	defer server.Close()

	// Act
	resp, err := http.Get(server.URL + "/api/v1/resources/")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// Test HEAD /api/v1/resources/{resourceId} - Success
func TestHeadResourceSuccess(t *testing.T) {
	// Arrange
	server := httptest.NewServer(mockServerHandler())
	defer server.Close()

	// Act
	req, _ := http.NewRequest(http.MethodHead, server.URL+"/api/v1/resources/8", nil)
	resp, err := http.DefaultClient.Do(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

// Test HEAD /api/v1/resources/{resourceId} - Not Found
func TestHeadResourceNotFound(t *testing.T) {
	// Arrange
	server := httptest.NewServer(mockServerHandler())
	defer server.Close()

	// Act
	req, _ := http.NewRequest(http.MethodHead, server.URL+"/api/v1/resources/99999", nil)
	resp, err := http.DefaultClient.Do(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// Test GET /api/v1/resources/{resourceId} - Success
func TestGetResourceSuccess(t *testing.T) {
	// Arrange
	server := httptest.NewServer(mockServerHandler())
	defer server.Close()

	// Act
	resp, err := http.Get(server.URL + "/api/v1/resources/8")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// Test GET /api/v1/resources/{resourceId} - Not Found
func TestGetResourceNotFound(t *testing.T) {
	// Arrange
	server := httptest.NewServer(mockServerHandler())
	defer server.Close()

	// Act
	resp, err := http.Get(server.URL + "/api/v1/resources/99999")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// Test GET /api/v1/tenants/
func TestGetTenantsSuccess(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `[{"id":1,"name":"Tenant 1"}, {"id":2,"name":"Tenant 2"}]`)
	}))
	defer server.Close()

	// Act
	resp, err := http.Get(server.URL + "/api/v1/tenants/")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// Test POST /api/v1/tenants/ - Success
func TestCreateTenantSuccess(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
	defer server.Close()

	// Act
	resp, err := http.Post(server.URL+"/api/v1/tenants/", "application/json", nil)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

// Test DELETE /api/v1/tenants/{tenantID} - Success
func TestDeleteTenantSuccess(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
	defer server.Close()

	// Act
	req, _ := http.NewRequest(http.MethodDelete, server.URL+"/api/v1/tenants/8", nil)
	resp, err := http.DefaultClient.Do(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

// Test DELETE /api/v1/tenants/{tenantID} - Not Found
func TestDeleteTenantNotFound(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
	defer server.Close()

	// Act
	req, _ := http.NewRequest(http.MethodDelete, server.URL+"/api/v1/tenants/99999", nil)
	resp, err := http.DefaultClient.Do(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestShouldGenerateOpenApiSpec(t *testing.T) {
	// Arrange
	router := mux.NewRouter(mux.WithTitle("test title"), mux.WithDescription("test description"), mux.WithVersion("1.0.0"))
	ConfigureRoutes(router)
	generator := mux.NewGenerator()

	// Act
	spec, err := generator.GenerateSpec(router)
	//spec.MarshalToFile("openapi.yaml")

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, spec)

	expected := loadExpected(t)
	assert.Equal(t, expected.Normalize(), spec.Normalize())
}

func loadExpected(t *testing.T) mux.OpenAPISpec {
	expectedPath := filepath.Join(".", "openapi.yaml")
	data, err := os.ReadFile(expectedPath)
	require.NoError(t, err)

	var expected mux.OpenAPISpec
	err = yaml.Unmarshal(data, &expected)
	require.NoError(t, err)
	return expected
}
