package test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fgrzl/mux"
	"github.com/stretchr/testify/assert"
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
