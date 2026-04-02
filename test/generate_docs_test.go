package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fgrzl/mux"
	"github.com/fgrzl/mux/test/testsupport"
	"github.com/stretchr/testify/require"
)

func TestShouldWriteSpecFileGivenGenerateSpecWithGenerator(t *testing.T) {
	r := mux.NewRouter(mux.WithTitle("gen-test"), mux.WithVersion("0.0.1"))
	r.GET("/", func(rc mux.RouteContext) { rc.NoContent() })
	testsupport.ConfigureRoutes(r)

	spec, err := mux.GenerateSpecWithGenerator(mux.NewGenerator(), r)
	require.NoError(t, err)

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "test.yml")

	require.NoError(t, spec.MarshalToFile(outPath))

	_, err = os.Stat(outPath)
	require.NoError(t, err)

	var reloaded mux.OpenAPISpec
	require.NoError(t, reloaded.UnmarshalFromFile(outPath))
	require.NoError(t, reloaded.Validate())

	data, err := os.ReadFile(outPath)
	require.NoError(t, err)
	require.True(t, strings.Contains(string(data), "/api/v1/resources"))
}
