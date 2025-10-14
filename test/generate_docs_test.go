package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fgrzl/mux"
	openapi "github.com/fgrzl/mux/pkg/openapi"
	"github.com/fgrzl/mux/test/testsupport"
	"github.com/stretchr/testify/require"
)

func TestShouldWriteSpecFileGivenGenerateAndSave(t *testing.T) {
	// Arrange: create a router and configure the canonical test routes
	r := mux.NewRouter(mux.WithTitle("gen-test"), mux.WithVersion("0.0.1"))
	r.GET("/", func(rc mux.RouteContext) { rc.NoContent() })
	testsupport.ConfigureRoutes(r)

	// Build generator and collect info/routes from router
	gen := mux.NewGenerator()

	info, err := r.InfoObject()
	require.NoError(t, err)

	routes, err := r.Routes()
	require.NoError(t, err)

	// Use a temp file to write the spec (YAML)
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "test.yml")

	// Act: generate and save
	require.NoError(t, gen.GenerateAndSave(info, routes, outPath))

	// Assert: file exists and can be unmarshaled
	_, err = os.Stat(outPath)
	require.NoError(t, err)

	var spec openapi.OpenAPISpec
	require.NoError(t, spec.UnmarshalFromFile(outPath))

	// Basic content checks: Verify top-level spec and one known path from test routes
	require.Equal(t, "3.1.0", spec.OpenAPI)
	require.NotNil(t, spec.Info)
	// The test routes register /api/v1/resources/ (generator normalizes to no trailing slash)
	require.Contains(t, spec.Paths, "/api/v1/resources")
}
