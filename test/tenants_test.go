package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/fgrzl/mux/pkg/common"
	"github.com/fgrzl/mux/test/testsupport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldReturnTenantsWhenRouterImplemented(t *testing.T) {
	// Ensure fresh service state
	testsupport.Service = testsupport.NewFakeService()

	server := newTestServer(t)

	resp, err := testClient.Get(server.URL + testsupport.APITenants)
	require.NoError(t, err)
	body := mustReadBody(t, resp)
	// Expect a JSON array with tenants when route is implemented; tighten
	// assertions so this fails if static fallback serves HTML.
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	ct := resp.Header.Get(common.HeaderContentType)
	assert.Contains(t, ct, common.MimeJSON)
	var arr []map[string]any
	err = json.Unmarshal(body, &arr)
	require.NoError(t, err)
	require.NotEmpty(t, arr)
}

func TestShouldCreateTenantWhenRouterImplemented(t *testing.T) {
	testsupport.Service = testsupport.NewFakeService()

	server := newTestServer(t)

	resp, err := testClient.Post(server.URL+testsupport.APITenants, common.MimeJSON, bytes.NewReader([]byte(`{"tenant_id":8,"name":"New","plan":"gold"}`)))
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	ct := resp.Header.Get(common.HeaderContentType)
	assert.Contains(t, ct, common.MimeJSON)
	body := mustReadBody(t, resp)
	var obj map[string]any
	err = json.Unmarshal(body, &obj)
	require.NoError(t, err)
	assert.EqualValues(t, 8, obj["tenant_id"])
}

func TestShouldDeleteTenantWhenRouterImplemented(t *testing.T) {
	testsupport.Service = testsupport.NewFakeService()

	server := newTestServer(t)

	req, _ := http.NewRequest(http.MethodDelete, server.URL+fmt.Sprintf(testsupport.APITenantByID, 8), nil)
	resp, err := testClient.Do(req)
	require.NoError(t, err)
	// Expect 204 No Content when implemented
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	body := mustReadBody(t, resp)
	assert.Equal(t, 0, len(body))
}

func TestShouldListTenantResourcesWhenRouterImplemented(t *testing.T) {
	testsupport.Service = testsupport.NewFakeService()

	server := newTestServer(t)

	resp, err := testClient.Get(server.URL + fmt.Sprintf(testsupport.APITenantResources, 0))
	require.NoError(t, err)
	// Expect JSON list of resources when implemented
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	ct := resp.Header.Get(common.HeaderContentType)
	assert.Contains(t, ct, common.MimeJSON)
	body := mustReadBody(t, resp)
	var arr []map[string]any
	err = json.Unmarshal(body, &arr)
	require.NoError(t, err)
}

func TestShouldCreateTenantResourceWhenRouterImplemented(t *testing.T) {
	testsupport.Service = testsupport.NewFakeService()

	server := newTestServer(t)

	body := `{"tenant_id":0,"name":"created","type":"resource"}`
	resp, err := testClient.Post(server.URL+fmt.Sprintf(testsupport.APITenantResources, 0), common.MimeJSON, strings.NewReader(body))
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	ct := resp.Header.Get(common.HeaderContentType)
	assert.Contains(t, ct, common.MimeJSON)
	b := mustReadBody(t, resp)
	var obj map[string]any
	err = json.Unmarshal(b, &obj)
	require.NoError(t, err)
}

func TestShouldUpdateTenantWhenRouterImplemented(t *testing.T) {
	testsupport.Service = testsupport.NewFakeService()

	server := newTestServer(t)

	body := `{"tenant_id":0,"name":"updated","plan":"silver"}`
	req, _ := http.NewRequest(http.MethodPut, server.URL+fmt.Sprintf(testsupport.APITenantByID, 0), strings.NewReader(body))
	req.Header.Set(common.HeaderContentType, common.MimeJSON)
	resp, err := testClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	ct := resp.Header.Get(common.HeaderContentType)
	assert.Contains(t, ct, common.MimeJSON)
	b := mustReadBody(t, resp)
	var obj map[string]any
	err = json.Unmarshal(b, &obj)
	require.NoError(t, err)
}

func TestShouldReturnNotFoundForMissingTenantWhenRouterImplemented(t *testing.T) {
	testsupport.Service = testsupport.NewFakeService()

	server := newTestServer(t)

	resp, err := testClient.Get(server.URL + fmt.Sprintf(testsupport.APITenantByID, 99999))
	require.NoError(t, err)
	// Expect a 404 ProblemDetails JSON when implemented
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	ct := resp.Header.Get(common.HeaderContentType)
	// allow either application/json or application/problem+json
	assert.True(t, strings.Contains(ct, common.MimeJSON) || strings.Contains(ct, common.MimeProblemJSON))
	b := mustReadBody(t, resp)
	var obj map[string]any
	err = json.Unmarshal(b, &obj)
	require.NoError(t, err)
}
