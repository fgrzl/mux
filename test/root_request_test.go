package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fgrzl/mux"
	"github.com/stretchr/testify/assert"
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
	resp, err := http.Get(server.URL)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Assert
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var s string
	err = json.NewDecoder(resp.Body).Decode(&s)
	assert.NoError(t, err)
	assert.Equal(t, "ok", s)
}
