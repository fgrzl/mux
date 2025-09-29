package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fgrzl/mux"
	"github.com/stretchr/testify/assert"
)

// Reproduce calling server root with no path (http://host:port)
func TestRootRequest(t *testing.T) {
	r := mux.NewRouter()
	r.GET("/", func(rc mux.RouteContext) {
		rc.OK("ok")
	})
	server := httptest.NewServer(r)
	defer server.Close()

	// Request without explicit path
	resp, err := http.Get(server.URL)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var s string
	err = json.NewDecoder(resp.Body).Decode(&s)
	assert.NoError(t, err)
	assert.Equal(t, "ok", s)
}
