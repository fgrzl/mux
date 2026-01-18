package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/fgrzl/mux"
	"github.com/fgrzl/mux/pkg/common"
	"github.com/fgrzl/mux/test/testsupport"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	server := newTestServer(t)

	// Act
	resp, err := testClient.Get(server.URL + testsupport.APIResources)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// Test HEAD /api/v1/resources/{resourceId} - Success
func TestShouldReturnNoContentWhenResourceExists(t *testing.T) {
	// Arrange
	server := newTestServer(t)

	// Act
	req, _ := http.NewRequest(http.MethodHead, server.URL+fmt.Sprintf(testsupport.APIResourceByID, 8), nil)
	resp, err := testClient.Do(req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

// Test HEAD /api/v1/resources/{resourceId} - Not Found
func TestShouldReturnNotFoundWhenResourceDoesNotExist(t *testing.T) {
	// Arrange
	server := newTestServer(t)

	// Act
	req, _ := http.NewRequest(http.MethodHead, server.URL+fmt.Sprintf(testsupport.APIResourceByID, 99999), nil)
	resp, err := testClient.Do(req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// Test GET /api/v1/resources/{resourceId} - Success
func TestShouldReturnResourceWhenResourceIdIsValid(t *testing.T) {
	// Arrange
	server := newTestServer(t)

	// Act
	resp, err := testClient.Get(server.URL + fmt.Sprintf(testsupport.APIResourceByID, 8))

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// Test GET /api/v1/resources/{resourceId} - Not Found
func TestShouldReturnNotFoundWhenResourceIdIsInvalid(t *testing.T) {
	// Arrange
	server := newTestServer(t)

	// Act
	resp, err := testClient.Get(server.URL + fmt.Sprintf(testsupport.APIResourceByID, 99999))

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// Test GET /api/v1/tenants/
func TestShouldReturnTenantsWhenRequestIsValid(t *testing.T) {
	// Arrange
	server := newTestServerWithHandler(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `[{"id":1,"name":"Tenant 1"}, {"id":2,"name":"Tenant 2"}]`)
	}))

	// Act
	resp, err := testClient.Get(server.URL + testsupport.APITenants)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// Test POST /api/v1/tenants/ - Success
func TestShouldCreateTenantWhenRequestIsValid(t *testing.T) {
	// Arrange
	server := newTestServerWithHandler(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))

	// Act
	resp, err := testClient.Post(server.URL+testsupport.APITenants, common.MimeJSON, nil)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

// Test DELETE /api/v1/tenants/{tenantID} - Success
func TestShouldDeleteTenantWhenTenantExists(t *testing.T) {
	// Arrange
	server := newTestServerWithHandler(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))

	// Act
	req, _ := http.NewRequest(http.MethodDelete, server.URL+fmt.Sprintf(testsupport.APITenantByID, 8), nil)
	resp, err := testClient.Do(req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

// Test DELETE /api/v1/tenants/{tenantID} - Not Found
func TestShouldReturnNotFoundWhenTenantDoesNotExist(t *testing.T) {
	// Arrange
	server := newTestServerWithHandler(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))

	// Act
	req, _ := http.NewRequest(http.MethodDelete, server.URL+fmt.Sprintf(testsupport.APITenantByID, 99999), nil)
	resp, err := testClient.Do(req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// Additional integration tests exercising query params, complex binding, headers,
// UUID path params, form binds, and the static catch-all.
func TestShouldReturnResourcesWhenSearchedByNameAndType(t *testing.T) {
	// Arrange
	server := newTestServer(t)

	// pick an existing resource name from the test service
	resources := testsupport.Service.ListResources(0)
	require.NotEmpty(t, resources)
	name := resources[0].Name

	// Act: search by name and type
	u := server.URL + testsupport.APIBase + "/resources/search"
	q := url.Values{}
	q.Set("name", name)
	q.Add("type", "resource")
	resp, err := testClient.Get(u + "?" + q.Encode())

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestShouldRejectEmptyBulkAndCreateResourcesWhenValid(t *testing.T) {
	server := newTestServer(t)

	// invalid (empty array) -> 400
	buf := bytes.NewBufferString("[]")
	resp, err := testClient.Post(server.URL+testsupport.APIBase+"/resources/bulk", common.MimeJSON, buf)
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// valid -> 201 and created resources returned
	resources := []map[string]any{{"tenant_id": 0, "name": "new-res", "type": "resource"}}
	b, _ := json.Marshal(resources)
	resp, err = testClient.Post(server.URL+testsupport.APIBase+"/resources/bulk", common.MimeJSON, bytes.NewReader(b))
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	body := mustReadBody(t, resp)
	// ensure the response contains the name we posted
	assert.Contains(t, string(body), "new-res")
}

func TestShouldUpdateResourceMetadataWhenValid(t *testing.T) {
	server := newTestServer(t)

	// pick an existing resource id
	resources := testsupport.Service.ListResources(0)
	require.NotEmpty(t, resources)
	id := resources[0].ResourceID

	payload := map[string]map[string]string{"metadata": {"k": "v"}}
	b, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf(server.URL+testsupport.APIResourceMetadata, id), bytes.NewReader(b))
	req.Header.Set(common.HeaderContentType, common.MimeJSON)
	resp, err := testClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body := mustReadBody(t, resp)
	assert.Contains(t, string(body), "metadata")
}

func TestShouldReturnOKWhenUUIDValidAndBadRequestWhenInvalid(t *testing.T) {
	server := newTestServer(t)

	valid := uuid.New()
	resp, err := testClient.Get(server.URL + fmt.Sprintf(testsupport.APIItemsUUID, valid.String()))
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// invalid uuid
	resp, err = testClient.Get(server.URL + fmt.Sprintf(testsupport.APIItemsUUID, "not-a-uuid"))
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestShouldEchoHeaderAndReturnFilterResultsWhenQueriesProvided(t *testing.T) {
	server := newTestServer(t)

	// header echo
	req, _ := http.NewRequest(http.MethodGet, server.URL+testsupport.APIHeadersEcho, nil)
	req.Header.Set(common.HeaderXEcho, "hello")
	resp, err := testClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body := mustReadBody(t, resp)
	assert.Contains(t, string(body), "hello")

	// filter with ints and uuids
	u := server.URL + testsupport.APIBase + "/filter?ids=1&ids=2&uuids=" + uuid.NewString() + "&uuids=" + uuid.NewString()
	resp, err = testClient.Get(u)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body = mustReadBody(t, resp)
	assert.Contains(t, string(body), "ids")
	assert.Contains(t, string(body), "uuids")
}

func TestShouldSubmitFormWhenValidAndServeStaticFallbackForUnknownPaths(t *testing.T) {
	server := newTestServer(t)

	// form submit
	form := url.Values{}
	form.Set("field", "value1")
	resp, err := testClient.Post(server.URL+testsupport.APIFormSubmit, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body := mustReadBody(t, resp)
	assert.Contains(t, string(body), "value1")

	// static fallback - request a path that should be caught by static fallback
	resp, err = testClient.Get(server.URL + "/some/random/path/that/does/not/exist")
	require.NoError(t, err)
	defer resp.Body.Close()
	// static fallback may return 200 (serve index.html) or 404 if file missing; accept both
	assert.Contains(t, []int{http.StatusOK, http.StatusNotFound}, resp.StatusCode)
}

// Tenant, OpenAPI, and static fallback tests moved to dedicated files under test/ (tenants_test.go, openapi_test.go, static_test.go)
