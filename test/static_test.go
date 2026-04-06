package test

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/fgrzl/mux"
	"github.com/fgrzl/mux/test/testsupport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldReturnStaticFallback200(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{name: "root path", path: "/"},
		{name: "one level deep", path: "/foo"},
		{name: "two levels deep", path: "/foo/bar"},
		{name: "three levels deep", path: "/foo/bar/baz"},
	}

	r := mux.NewRouter()
	testsupport.ConfigureRoutes(r)
	r.StaticFallback("/**", "static", "static/index.html").AllowAnonymous()
	server := newTestServerWithHandler(t, r)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqPath := server.URL + tt.path
			resp, err := testClient.Get(reqPath)
			require.NoError(t, err)
			require.NotNil(t, resp)
			defer resp.Body.Close()
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestShouldReturnStaticFallback404(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{name: "root path", path: "/"},
		{name: "one level deep", path: "/foo"},
		{name: "two levels deep", path: "/foo/bar"},
		{name: "three levels deep", path: "/foo/bar/baz"},
	}

	r := mux.NewRouter()
	testsupport.ConfigureRoutes(r)
	r.StaticFallback("/**", "assets", "assets/index.html").AllowAnonymous()
	server := newTestServerWithHandler(t, r)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := testClient.Get(server.URL + tt.path)
			require.NoError(t, err)
			require.NotNil(t, resp)
			defer resp.Body.Close()
			assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		})
	}
}

func TestShouldServeNestedStaticFallbackFile(t *testing.T) {
	rootDir := t.TempDir()
	staticDir := filepath.Join(rootDir, "static")
	fallbackPath := filepath.Join(staticDir, "spa", "index.html")

	require.NoError(t, os.MkdirAll(filepath.Dir(fallbackPath), 0o755))
	require.NoError(t, os.WriteFile(fallbackPath, []byte("nested fallback"), 0o644))

	r := mux.NewRouter()
	r.StaticFallback("/app/**", staticDir, fallbackPath).AllowAnonymous()
	server := newTestServerWithHandler(t, r)

	resp, err := testClient.Get(server.URL + "/app/unknown")
	require.NoError(t, err)
	require.NotNil(t, resp)
	body := mustReadBody(t, resp)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(body), "nested fallback")
}
