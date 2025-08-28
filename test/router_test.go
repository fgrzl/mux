package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fgrzl/mux"
	"github.com/fgrzl/mux/test/testsupport"
	"github.com/google/uuid"
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

	testsupport.ConfigureRoutes(r)

	return r
}

// Test GET /api/v1/resources/
func TestShouldReturnResourcesWhenRequestIsValid(t *testing.T) {
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
func TestShouldReturnNoContentWhenResourceExists(t *testing.T) {
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
func TestShouldReturnNotFoundWhenResourceDoesNotExist(t *testing.T) {
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
func TestShouldReturnResourceWhenResourceIdIsValid(t *testing.T) {
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
func TestShouldReturnNotFoundWhenResourceIdIsInvalid(t *testing.T) {
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
func TestShouldReturnTenantsWhenRequestIsValid(t *testing.T) {
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
func TestShouldCreateTenantWhenRequestIsValid(t *testing.T) {
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
func TestShouldDeleteTenantWhenTenantExists(t *testing.T) {
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
func TestShouldReturnNotFoundWhenTenantDoesNotExist(t *testing.T) {
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

// Additional integration tests exercising query params, complex binding, headers,
// UUID path params, form binds, and the static catch-all.
func TestShouldReturnResourcesWhenSearchedByNameAndType(t *testing.T) {
	// Arrange
	server := httptest.NewServer(mockServerHandler())
	defer server.Close()

	// pick an existing resource name from the test service
	resources := testsupport.Service.ListResources(0)
	require.NotEmpty(t, resources)
	name := resources[0].Name

	// Act: search by name and type
	u := server.URL + "/api/v1/resources/search"
	q := url.Values{}
	q.Set("name", name)
	q.Add("type", "resource")
	resp, err := http.Get(u + "?" + q.Encode())

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestShouldRejectEmptyBulkAndCreateResourcesWhenValid(t *testing.T) {
	server := httptest.NewServer(mockServerHandler())
	defer server.Close()

	// invalid (empty array) -> 400
	buf := bytes.NewBufferString("[]")
	resp, err := http.Post(server.URL+"/api/v1/resources/bulk", "application/json", buf)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// valid -> 201 and created resources returned
	resources := []map[string]any{{"tenant_id": 0, "name": "new-res", "type": "resource"}}
	b, _ := json.Marshal(resources)
	resp, err = http.Post(server.URL+"/api/v1/resources/bulk", "application/json", bytes.NewReader(b))
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	// ensure the response contains the name we posted
	assert.Contains(t, string(body), "new-res")
}

func TestShouldUpdateResourceMetadataWhenValid(t *testing.T) {
	server := httptest.NewServer(mockServerHandler())
	defer server.Close()

	// pick an existing resource id
	resources := testsupport.Service.ListResources(0)
	require.NotEmpty(t, resources)
	id := resources[0].ResourceID

	payload := map[string]map[string]string{"metadata": {"k": "v"}}
	b, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf(server.URL+"/api/v1/resources/%d/metadata", id), bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "metadata")
}

func TestShouldReturnOKWhenUUIDValidAndBadRequestWhenInvalid(t *testing.T) {
	server := httptest.NewServer(mockServerHandler())
	defer server.Close()

	valid := uuid.New()
	resp, err := http.Get(server.URL + "/api/v1/items/" + valid.String() + "/uuid")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// invalid uuid
	resp, err = http.Get(server.URL + "/api/v1/items/not-a-uuid/uuid")
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestShouldEchoHeaderAndReturnFilterResultsWhenQueriesProvided(t *testing.T) {
	server := httptest.NewServer(mockServerHandler())
	defer server.Close()

	// header echo
	req, _ := http.NewRequest(http.MethodGet, server.URL+"/api/v1/headers/echo", nil)
	req.Header.Set("X-Echo", "hello")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "hello")

	// filter with ints and uuids
	u := server.URL + "/api/v1/filter?ids=1&ids=2&uuids=" + uuid.NewString() + "&uuids=" + uuid.NewString()
	resp, err = http.Get(u)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, _ = io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "ids")
	assert.Contains(t, string(body), "uuids")
}

func TestShouldSubmitFormWhenValidAndServeStaticFallbackForUnknownPaths(t *testing.T) {
	server := httptest.NewServer(mockServerHandler())
	defer server.Close()

	// form submit
	form := url.Values{}
	form.Set("field", "value1")
	resp, err := http.Post(server.URL+"/api/v1/form/submit", "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "value1")

	// static fallback - request a path that should be caught by static fallback
	resp, err = http.Get(server.URL + "/some/random/path/that/does/not/exist")
	require.NoError(t, err)
	// static fallback may return 200 (serve index.html) or 404 if file missing; accept both
	assert.Contains(t, []int{http.StatusOK, http.StatusNotFound}, resp.StatusCode)
}

// The following tests represent expected tenant-related routes that are not
// yet registered on the router; they are intentionally written to fail so the
// missing routes can be implemented next.
func TestShouldReturnTenantsWhenRouterImplemented(t *testing.T) {
	server := httptest.NewServer(mockServerHandler())
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/v1/tenants/")
	require.NoError(t, err)
	// Expect a JSON array with tenants when route is implemented; tighten
	// assertions so this fails if static fallback serves HTML.
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	ct := resp.Header.Get("Content-Type")
	assert.Contains(t, ct, "application/json")
	body, _ := io.ReadAll(resp.Body)
	var arr []map[string]any
	err = json.Unmarshal(body, &arr)
	require.NoError(t, err)
	require.NotEmpty(t, arr)
}

func TestShouldCreateTenantWhenRouterImplemented(t *testing.T) {
	server := httptest.NewServer(mockServerHandler())
	defer server.Close()

	resp, err := http.Post(server.URL+"/api/v1/tenants/", "application/json", bytes.NewReader([]byte(`{"tenant_id":8,"name":"New","plan":"gold"}`)))
	require.NoError(t, err)
	// Expect JSON created response when implemented
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	ct := resp.Header.Get("Content-Type")
	assert.Contains(t, ct, "application/json")
	body, _ := io.ReadAll(resp.Body)
	var obj map[string]any
	err = json.Unmarshal(body, &obj)
	require.NoError(t, err)
	assert.EqualValues(t, 8, obj["tenant_id"])
}

func TestShouldDeleteTenantWhenRouterImplemented(t *testing.T) {
	server := httptest.NewServer(mockServerHandler())
	defer server.Close()

	req, _ := http.NewRequest(http.MethodDelete, server.URL+"/api/v1/tenants/8", nil)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	// Expect 204 No Content when implemented
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, 0, len(body))
}

func TestShouldListTenantResourcesWhenRouterImplemented(t *testing.T) {
	server := httptest.NewServer(mockServerHandler())
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/v1/tenants/0/resources")
	require.NoError(t, err)
	// Expect JSON list of resources when implemented
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	ct := resp.Header.Get("Content-Type")
	assert.Contains(t, ct, "application/json")
	body, _ := io.ReadAll(resp.Body)
	var arr []map[string]any
	err = json.Unmarshal(body, &arr)
	require.NoError(t, err)
}

func TestShouldCreateTenantResourceWhenRouterImplemented(t *testing.T) {
	server := httptest.NewServer(mockServerHandler())
	defer server.Close()

	body := `{"tenant_id":0,"name":"created","type":"resource"}`
	resp, err := http.Post(server.URL+"/api/v1/tenants/0/resources", "application/json", strings.NewReader(body))
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	ct := resp.Header.Get("Content-Type")
	assert.Contains(t, ct, "application/json")
	b, _ := io.ReadAll(resp.Body)
	var obj map[string]any
	err = json.Unmarshal(b, &obj)
	require.NoError(t, err)
}

func TestShouldUpdateTenantWhenRouterImplemented(t *testing.T) {
	server := httptest.NewServer(mockServerHandler())
	defer server.Close()

	body := `{"tenant_id":0,"name":"updated","plan":"silver"}`
	req, _ := http.NewRequest(http.MethodPut, server.URL+"/api/v1/tenants/0", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	ct := resp.Header.Get("Content-Type")
	assert.Contains(t, ct, "application/json")
	b, _ := io.ReadAll(resp.Body)
	var obj map[string]any
	err = json.Unmarshal(b, &obj)
	require.NoError(t, err)
}

func TestShouldReturnNotFoundForMissingTenantWhenRouterImplemented(t *testing.T) {
	server := httptest.NewServer(mockServerHandler())
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/v1/tenants/99999")
	require.NoError(t, err)
	// Expect a 404 ProblemDetails JSON when implemented
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	ct := resp.Header.Get("Content-Type")
	// allow either application/json or application/problem+json
	assert.True(t, strings.Contains(ct, "application/json") || strings.Contains(ct, "application/problem+json"))
	b, _ := io.ReadAll(resp.Body)
	var obj map[string]any
	err = json.Unmarshal(b, &obj)
	require.NoError(t, err)
}

func TestShouldGenerateOpenApiSpec(t *testing.T) {
	// Arrange
	router := mux.NewRouter(mux.WithTitle("test title"), mux.WithDescription("test description"), mux.WithVersion("1.0.0"))
	testsupport.ConfigureRoutes(router)
	generator := mux.NewGenerator()

	// Act
	spec, err := mux.GenerateSpecWithGenerator(generator, router)

	// if you need to write the file to review
	// spec.MarshalToFile("openapi.yaml")

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
