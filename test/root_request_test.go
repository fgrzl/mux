package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fgrzl/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestShouldHandleRootPathRequestGivenNoExplicitPath ensures the router can handle
// requests to the server root (http://host:port) without an explicit path.
func TestShouldHandleRootPathRequestGivenNoExplicitPath(t *testing.T) {
	// Arrange
	r := mux.NewRouter()
	r.GET("/", func(rc mux.RouteContext) {
		rc.OK("ok")
	})
	server := httptest.NewServer(r)
	defer server.Close()

	// Act - Request without explicit path
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL, nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	// Assert
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var s string
	err = json.NewDecoder(resp.Body).Decode(&s)
	assert.NoError(t, err)
	assert.Equal(t, "ok", s)
}
